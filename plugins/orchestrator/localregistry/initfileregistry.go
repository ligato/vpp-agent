// Copyright (c) 2020 Pantheon.tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package localregistry

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml2 "github.com/ghodss/yaml"
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync/local"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/datasync/syncbase"
	"go.ligato.io/cn-infra/v2/db/keyval"
	"go.ligato.io/cn-infra/v2/infra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
)

const (
	registryName        = "init-file-registry"
	defaultInitFilePath = "initial-config.yaml"
)

type Option func(*InitFileRegistry)

// InitFileRegistry is local read-only NB configuration provider with exclusive data source from a file
// given by a file path (InitConfigFilePath). Its purpose is to seamlessly integrated NB configuration
// from file as another NB configuration provider (to existing providers: etcd, consul, redis) and integrate
// it's configuration into agent in the same standard way(datasync.KeyValProtoWatcher). The content of this
// registry is meant to be only the initial NB configuration for the agent and will not reflect any changes
// inside given file after initial content loading.
//
// The NB configuration provisioning process and how this registry fits into it:
// 	1. NB data sources register to default resync plugin (InitFileRegistry registers too in watchNBResync(),
//	   but only when there are some NB config data from file, otherwise it makes no sense to register because
//	   there is nothing to forward. This also means that before register to resync plugin, the NB config from
//	   file will be preloaded)
// 	2. Call to resync plugin's DoResync triggers resync to NB configuration sources (InitFileRegistry takes
//	   its preloaded NB config and stores it into another inner local registry)
// 	3. NB configuration sources are also watchable (datasync.KeyValProtoWatcher) and the resync data is
//	   collected by the watcher.Aggregator (InitFileRegistry is also watchable/forwards data to watcher.Aggregator,
//	   it relies on the watcher capabilities of its inner local registry. This is the cause why to preloaded
//	   the NB config from file([]proto.Message storage) and push it to another inner local storage later
//	   (syncbase.Registry). If we used only one storage (syncbase.Registry for its watch capabilities), we
//	   couldn't answer some questins about the storage soon enough (watcher.Aggregator in Watch(...) needs to
//	   know whether this storage will send some data or not, otherwise the retrieval can hang on waiting for
//	   data that never come))
// 	4. watcher.Aggregator merges all collected resync data and forwards them its watch clients (it also implements
//	   datasync.KeyValProtoWatcher just like the NB data sources).
//  5. Clients of Aggregator (currently orchestrator and ifplugin) handle the NB changes/resync properly.
type InitFileRegistry struct {
	infra.PluginDeps

	initialized             bool
	config                  *Config
	watchedRegistry         *syncbase.Registry
	pushedToWatchedRegistry bool
	preloadedNBConfigs      []proto.Message
}

// Config holds the InitFileRegistry configuration.
type Config struct {
	DisableInitialConfiguration  bool   `json:"disable-initial-configuration"`
	InitialConfigurationFilePath string `json:"initial-configuration-file-path"`
}

// NewInitFileRegistryPlugin creates a new InitFileRegistry Plugin with the provided Options
func NewInitFileRegistryPlugin(opts ...Option) *InitFileRegistry {
	p := &InitFileRegistry{}

	p.PluginName = "initfileregistry"
	p.watchedRegistry = syncbase.NewRegistry()

	for _, o := range opts {
		o(p)
	}
	if p.Cfg == nil {
		p.Cfg = config.ForPlugin(p.String(),
			config.WithCustomizedFlag(config.FlagName(p.String()), "initfileregistry.conf"),
		)
	}
	p.PluginDeps.SetupLog()

	return p
}

// Init initialize registry
func (r *InitFileRegistry) Init() error {
	if !r.initialized {
		return r.initialize()
	}
	return nil
}

// Empty checks whether this registry holds data or not. As result of the properties of this registry
// (readonly, will be filled only once from initial file import), this method directly indicates whether
// the watchers of this registry will receive any data (Empty() == false, receive initial resync) or
// won't receive anything at all (Empty() == true)
func (r *InitFileRegistry) Empty() bool {
	if !r.initialized { // could be called from init of other plugins -> possibly before this plugin init
		if err := r.initialize(); err != nil {
			r.Log.Errorf("cannot initialize InitFileRegistry due to: %v", err)
		}
	}
	return len(r.preloadedNBConfigs) == 0
}

// Watch functionality is forwarded to inner syncbase.Registry. For some watchers might be relevant
// whether any data will be pushed to them at all (i.e. watcher.Aggregator). They should use the
// Empty() method to find out whether there are (=ever will be do to nature of this registry) any
// data for pushing to watchers.
func (r *InitFileRegistry) Watch(resyncName string, changeChan chan datasync.ChangeEvent,
	resyncChan chan datasync.ResyncEvent, keyPrefixes ...string) (datasync.WatchRegistration, error) {
	return r.watchedRegistry.Watch(resyncName, changeChan, resyncChan, keyPrefixes...)
}

// initialize will try to pre-load the NB initial data
// (watchers of this registry will receive it only after call to resync)
func (r *InitFileRegistry) initialize() error {
	defer func() {
		r.initialized = true
	}()

	// parse configuration file
	var err error
	r.config, err = r.retrieveConfig()
	if err != nil {
		return err
	}

	// Initial NB configuration loaded from file
	if !r.config.DisableInitialConfiguration {
		// preload NB config data from file
		if err := r.preloadNBConfigs(r.config.InitialConfigurationFilePath); err != nil {
			return fmt.Errorf("cannot preload initial NB configuration from file due to: %w", err)
		}
		if len(r.preloadedNBConfigs) != 0 {
			// watch for resync.DefaultPlugin.DoResync() that will trigger pushing of preloaded
			// NB config data from file into NB aggregator watcher
			// (see InitFileRegistry struct docs for detailed explanation)
			r.watchNBResync()
		}
	}
	return nil
}

// retrieveConfig loads InitFileRegistry plugin configuration file.
func (r *InitFileRegistry) retrieveConfig() (*Config, error) {
	cfg := &Config{
		// default configuration
		DisableInitialConfiguration:  false,
		InitialConfigurationFilePath: defaultInitFilePath,
	}
	found, err := r.Cfg.LoadValue(cfg)
	if !found {
		if err == nil {
			r.Log.Debug("InitFileRegistry plugin config not found")
		} else {
			r.Log.Debugf("InitFileRegistry plugin config cannot be loaded due to: %v", err)
		}
		return cfg, err
	}
	if err != nil {
		return nil, err
	}
	return cfg, err
}

// watchNBResync will watch to default resync plugin's resync call(resync.DefaultPlugin.DoResync()) and will
// load NB initial config from file (already preloaded from initialize()) when the first resync will be fired.
func (r *InitFileRegistry) watchNBResync() {
	registration := resync.DefaultPlugin.Register(registryName)
	go r.watchResync(registration)
}

// watchResync will listen to resync plugin resync signals and at first resync will push the preloaded
// NB initial config into internal local register (p.registry)
func (r *InitFileRegistry) watchResync(resyncReg resync.Registration) {
	for resyncStatus := range resyncReg.StatusChan() {
		// resyncReg.StatusChan == Started => resync
		if resyncStatus.ResyncStatus() == resync.Started && !r.pushedToWatchedRegistry {
			if !r.Empty() { // put preloaded NB init file data into watched p.registry
				c := client.NewClient(&txnFactory{r.watchedRegistry}, &orchestrator.DefaultPlugin)
				if err := c.ResyncConfig(r.preloadedNBConfigs...); err != nil {
					r.Log.Errorf("resyncing preloaded NB init file data "+
						"into watched registry failed: %w", err)
				}
			}
			r.pushedToWatchedRegistry = true
			resyncStatus.Ack()
			// TODO some done channel to not continue as NOP goroutine
			continue // can't unregister anymore -> need to listen to further resync signals, but it will be just NO-OPs
		}
		resyncStatus.Ack()
	}
}

// preloadNBConfigs imports NB configuration from file(filepath) into preloadedNBConfigs. If file is not found,
// it is not considered as error, but as a sign that the NB-configuration-loading-from-file feature should be
// not used (inner registry remains empty and watchers of this registry get no data).
func (r *InitFileRegistry) preloadNBConfigs(filePath string) error {
	// check existence of NB init file
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath := filePath
		if absFilePath, err := filepath.Abs(filePath); err == nil {
			filePath = absFilePath
		}
		r.Log.Debugf("Initialization configuration file(%v) not found. "+
			"Skipping its preloading.", filePath)
		return nil
	}

	// read data from file
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("problem reading file %s: %w", filePath, err)
	}

	// create dynamic config (using it instead of configurator.Config because it can hold also models defined
	// outside the VPP-Agent repo, i.e. if this code is using 3rd party code based on VPP-Agent and having its
	// additional properly registered configuration models)
	knownModels, err := client.LocalClient.KnownModels("config") // locally registered models
	if err != nil {
		return fmt.Errorf("cannot get registered models: %w", err)
	}
	cfg, err := client.NewDynamicConfig(knownModels)
	if err != nil {
		return fmt.Errorf("cannot create dynamic config due to: %w", err)
	}

	// filling dynamically created config with data from NB init file
	bj, err := yaml2.YAMLToJSON(b)
	if err != nil {
		return fmt.Errorf("cannot converting to JSON: %w", err)
	}
	err = protojson.Unmarshal(bj, cfg)
	if err != nil {
		return fmt.Errorf("cannot unmarshall init file data into dynamic config due to: %w", err)
	}

	// extracting proto messages from dynamic config structure
	// (generic client wants single proto messages and not one big hierarchical config)
	configMessages, err := client.DynamicConfigExport(cfg)
	if err != nil {
		return fmt.Errorf("cannot extract single init configuration proto messages "+
			"from one big configuration proto message due to: %w", err)
	}

	// remember extracted data for later push to watched registry
	r.preloadedNBConfigs = configMessages

	return nil
}

type txnFactory struct {
	registry *syncbase.Registry
}

func (p *txnFactory) NewTxn(resync bool) keyval.ProtoTxn {
	if resync {
		return local.NewProtoTxn(p.registry.PropagateResync)
	}
	return local.NewProtoTxn(p.registry.PropagateChanges)
}
