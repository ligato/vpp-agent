// Copyright (c) 2017 Cisco and/or its affiliates.
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

package etcd

import (
	"fmt"
	"sync"
	"time"

	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/utils/safeclose"
)

const (
	// healthCheckProbeKey is a key used to probe Etcd state
	healthCheckProbeKey = "/probe-etcd-connection"
	// ETCD reconnect interval
	defaultReconnectInterval = 2 * time.Second
)

// Plugin implements etcd plugin.
type Plugin struct {
	Deps
	sync.Mutex

	// Plugin is disabled if there is no config file available
	disabled bool
	// Set if connected to ETCD db
	connected bool
	// ETCD connection encapsulation
	connection *BytesConnectionEtcd
	// Read/Write proto modelled data
	protoWrapper *kvproto.ProtoWrapper

	// plugin config
	config *Config

	// List of callback functions, used in case ETCD is not connected immediately. All plugins using
	// ETCD as dependency add their own function if cluster is not reachable. After connection, all
	// functions are executed.
	onConnection []func() error

	autoCompactDone chan struct{}
	lastConnErr     error
}

// Deps lists dependencies of the etcd plugin.
// If injected, etcd plugin will use StatusCheck to signal the connection status.
type Deps struct {
	infra.Deps
	StatusCheck statuscheck.PluginStatusWriter // inject
	Resync      *resync.Plugin
}

// Init retrieves ETCD configuration and establishes a new connection
// with the etcd data store.
// If the configuration file doesn't exist or cannot be read, the returned error
// will be of os.PathError type. An untyped error is returned in case the file
// doesn't contain a valid YAML configuration.
// The function may also return error if TLS connection is selected and the
// CA or client certificate is not accessible(os.PathError)/valid(untyped).
// Check clientv3.New from coreos/etcd for possible errors returned in case
// the connection cannot be established.
func (plugin *Plugin) Init() (err error) {
	// Read ETCD configuration file. Returns error if does not exists.
	plugin.config, err = plugin.getEtcdConfig()
	if err != nil || plugin.disabled {
		return err
	}
	// Transforms .yaml config to ETCD client configuration
	etcdClientCfg, err := ConfigToClient(plugin.config)
	if err != nil {
		return err
	}
	// Uses config file to establish connection with the database
	plugin.connection, err = NewEtcdConnectionWithBytes(*etcdClientCfg, plugin.Log)
	// Register for providing status reports (polling mode).
	if plugin.StatusCheck != nil {
		plugin.StatusCheck.Register(plugin.PluginName, plugin.statusCheckProbe)
	} else {
		plugin.Log.Warnf("Unable to start status check for etcd")
	}
	if err != nil && plugin.config.AllowDelayedStart {
		// If the connection cannot be established during init, keep trying in another goroutine (if allowed) and
		// end the init
		go plugin.etcdReconnectionLoop(etcdClientCfg)
		return nil
	} else if err != nil {
		// If delayed start is not allowed, return error
		return fmt.Errorf("error connecting to ETCD: %v", err)
	}

	// If successful, configure and return
	plugin.configureConnection()

	// Mark plugin as connected at this point
	plugin.connected = true

	return nil
}

// Close shutdowns the connection.
func (plugin *Plugin) Close() error {
	err := safeclose.Close(plugin.autoCompactDone)
	return err
}

// NewBroker creates new instance of prefixed broker that provides API with arguments of type proto.Message.
func (plugin *Plugin) NewBroker(keyPrefix string) keyval.ProtoBroker {
	return plugin.protoWrapper.NewBroker(keyPrefix)
}

// NewWatcher creates new instance of prefixed broker that provides API with arguments of type proto.Message.
func (plugin *Plugin) NewWatcher(keyPrefix string) keyval.ProtoWatcher {
	return plugin.protoWrapper.NewWatcher(keyPrefix)
}

// Disabled returns *true* if the plugin is not in use due to missing
// etcd configuration.
func (plugin *Plugin) Disabled() (disabled bool) {
	return plugin.disabled
}

// OnConnect executes callback if plugin is connected, or gathers functions from all plugin with ETCD as dependency
func (plugin *Plugin) OnConnect(callback func() error) {
	plugin.Lock()
	defer plugin.Unlock()

	if plugin.connected {
		if err := callback(); err != nil {
			plugin.Log.Error(err)
		}
	} else {
		plugin.onConnection = append(plugin.onConnection, callback)
	}
}

// GetPluginName returns name of the plugin
func (plugin *Plugin) GetPluginName() infra.PluginName {
	return plugin.PluginName
}

// PutIfNotExists puts given key-value pair into etcd if there is no value set for the key. If the put was successful
// succeeded is true. If the key already exists succeeded is false and the value for the key is untouched.
func (plugin *Plugin) PutIfNotExists(key string, value []byte) (succeeded bool, err error) {
	if plugin.connection != nil {
		return plugin.connection.PutIfNotExists(key, value)
	}
	return false, fmt.Errorf("connection is not established")
}

// Compact compatcs the ETCD database to the specific revision
func (plugin *Plugin) Compact(rev ...int64) (toRev int64, err error) {
	if plugin.connection != nil {
		return plugin.connection.Compact(rev...)
	}
	return 0, fmt.Errorf("connection is not established")
}

// Method starts loop which attempt to connect to the ETCD. If successful, send signal callback with resync,
// which will be started when datasync confirms successful registration
func (plugin *Plugin) etcdReconnectionLoop(clientCfg *ClientConfig) {
	var err error
	// Set reconnect interval
	interval := plugin.config.ReconnectInterval
	if interval == 0 {
		interval = defaultReconnectInterval
	}
	plugin.Log.Infof("ETCD server %s not reachable in init phase. Agent will continue to try to connect every %d second(s)",
		plugin.config.Endpoints, interval)
	for {
		time.Sleep(interval)

		plugin.Log.Infof("Connecting to ETCD %v ...", plugin.config.Endpoints)
		plugin.connection, err = NewEtcdConnectionWithBytes(*clientCfg, plugin.Log)
		if err != nil {
			continue
		}
		plugin.setupPostInitConnection()
		return
	}
}

func (plugin *Plugin) setupPostInitConnection() {
	plugin.Log.Infof("ETCD server %s connected", plugin.config.Endpoints)

	plugin.Lock()
	defer plugin.Unlock()

	// Configure connection and set as connected
	plugin.configureConnection()
	plugin.connected = true
	// Execute callback functions (if any)
	for _, callback := range plugin.onConnection {
		if err := callback(); err != nil {
			plugin.Log.Error(err)
		}
	}
	// Call resync if any callback was executed. Otherwise there is nothing to resync
	if plugin.Resync != nil && len(plugin.onConnection) > 0 {
		plugin.Resync.DoResync()
	}
	plugin.Log.Debugf("Etcd reconnection loop ended")
}

// If ETCD is connected, complete all other procedures
func (plugin *Plugin) configureConnection() {
	if plugin.config.AutoCompact > 0 {
		if plugin.config.AutoCompact < time.Duration(time.Minute*60) {
			plugin.Log.Warnf("Auto compact option for ETCD is set to less than 60 minutes!")
		}
		plugin.startPeriodicAutoCompact(plugin.config.AutoCompact)
	}
	plugin.protoWrapper = kvproto.NewProtoWrapperWithSerializer(plugin.connection, &keyval.SerializerJSON{})
}

// ETCD status check probe function
func (plugin *Plugin) statusCheckProbe() (statuscheck.PluginState, error) {
	if plugin.connection == nil {
		plugin.connected = false
		return statuscheck.Error, fmt.Errorf("no ETCD connection available")
	}
	if _, _, _, err := plugin.connection.GetValue(healthCheckProbeKey); err != nil {
		plugin.lastConnErr = err
		plugin.connected = false
		return statuscheck.Error, err
	}
	if plugin.config.ReconnectResync && plugin.lastConnErr != nil {
		if plugin.Resync != nil {
			plugin.Resync.DoResync()
			plugin.lastConnErr = nil
		} else {
			plugin.Log.Warn("Expected resync after ETCD reconnect could not start beacuse of missing Resync plugin")
		}
	}
	plugin.connected = true
	return statuscheck.OK, nil
}

func (plugin *Plugin) getEtcdConfig() (*Config, error) {
	var etcdCfg Config
	found, err := plugin.PluginConfig.GetValue(&etcdCfg)
	if err != nil {
		return nil, err
	}
	if !found {
		plugin.Log.Info("ETCD config not found, skip loading this plugin")
		plugin.disabled = true
	}
	return &etcdCfg, nil
}

func (plugin *Plugin) startPeriodicAutoCompact(period time.Duration) {
	plugin.autoCompactDone = make(chan struct{})
	go func() {
		plugin.Log.Infof("Starting periodic auto compacting every %v", period)
		for {
			select {
			case <-time.After(period):
				plugin.Log.Debugf("Executing auto compact")
				if toRev, err := plugin.connection.Compact(); err != nil {
					plugin.Log.Errorf("Periodic auto compacting failed: %v", err)
				} else {
					plugin.Log.Infof("Auto compacting finished (to revision %v)", toRev)
				}
			case <-plugin.autoCompactDone:
				return
			}
		}
	}()
}
