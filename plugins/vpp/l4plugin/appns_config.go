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

//go:generate protoc --proto_path=../model/l4 --gogo_out=../model/l4 ../model/l4/l4.proto

package l4plugin

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l4plugin/nsidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l4plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l4"
)

// AppNsConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/l4/l4.proto"
// and stored in ETCD under the keys "/vnf-agent/{vnf-agent}/vpp/config/v1/l4/l4ftEnabled"
// and "/vnf-agent/{vnf-agent}/vpp/config/v1/l4/namespaces/{namespace_id}".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type AppNsConfigurator struct {
	log logging.Logger

	// In-memory mappings
	ifIndexes    ifaceidx.SwIfIndex
	appNsIndexes nsidx.AppNsIndexRW
	appNsCached  nsidx.AppNsIndexRW // the mapping stores not-configurable app namespaces with metadata
	appNsIdxSeq  uint32

	// VPP channel
	vppChan govppapi.Channel
	// VPP API handler
	l4Handler vppcalls.L4VppAPI

	stopwatch *measure.Stopwatch

	// Feature flag - internal state whether the L4 features are enabled or disabled
	l4ftEnabled bool
}

// Init members (channels...) and start go routines
func (plugin *AppNsConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-l4-plugin")
	plugin.log.Debugf("Initializing L4 configurator")

	// Mappings
	plugin.ifIndexes = swIfIndexes
	plugin.appNsIndexes = nsidx.NewAppNsIndex(nametoidx.NewNameToIdx(plugin.log, "namespace_indexes", nil))
	plugin.appNsCached = nsidx.NewAppNsIndex(nametoidx.NewNameToIdx(plugin.log, "not_configured_namespace_indexes", nil))
	plugin.appNsIdxSeq = 1

	// Stopwatch
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("AppNsConfigurator", plugin.log)
	}

	// VPP channels
	if plugin.vppChan, err = goVppMux.NewAPIChannel(); err != nil {
		return err
	}

	// VPP API handler
	if plugin.l4Handler, err = vppcalls.NewL4VppHandler(plugin.vppChan, plugin.stopwatch); err != nil {
		return err
	}

	// Message compatibility
	if err = plugin.vppChan.CheckMessageCompatibility(vppcalls.AppNsMessages...); err != nil {
		plugin.log.Error(err)
		return err
	}

	return nil
}

// Close members, channels
func (plugin *AppNsConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (plugin *AppNsConfigurator) clearMapping() {
	plugin.appNsIndexes.Clear()
	plugin.appNsCached.Clear()
}

// GetAppNsIndexes returns application namespace memory indexes
func (plugin *AppNsConfigurator) GetAppNsIndexes() nsidx.AppNsIndexRW {
	return plugin.appNsIndexes
}

// ConfigureL4FeatureFlag process the NB Features config and propagates it to bin api calls
func (plugin *AppNsConfigurator) ConfigureL4FeatureFlag(features *l4.L4Features) error {
	plugin.log.Info("Setting up L4 features")

	if features.Enabled {
		if err := plugin.configureL4FeatureFlag(); err != nil {
			return err
		}
		return plugin.resolveCachedNamespaces()
	} else {
		return plugin.DeleteL4FeatureFlag()
	}

	return nil
}

// configureL4FeatureFlag process the NB Features config and propagates it to bin api calls
func (plugin *AppNsConfigurator) configureL4FeatureFlag() error {
	plugin.log.Info("Configuring L4 features")

	if err := plugin.l4Handler.EnableL4Features(); err != nil {
		plugin.log.Errorf("Enabling L4 features failed: %v", err)
		return err
	}
	plugin.l4ftEnabled = true
	plugin.log.Infof("L4 features enabled")

	return nil
}

// DeleteL4FeatureFlag process the NB Features config and propagates it to bin api calls
func (plugin *AppNsConfigurator) DeleteL4FeatureFlag() error {
	plugin.log.Info("Removing L4 features")

	if err := plugin.l4Handler.DisableL4Features(); err != nil {
		plugin.log.Errorf("Disabling L4 features failed: %v", err)
		return err
	}

	plugin.l4ftEnabled = false
	plugin.log.Infof("L4 features disabled")

	return nil
}

// ConfigureAppNamespace process the NB AppNamespace config and propagates it to bin api calls
func (plugin *AppNsConfigurator) ConfigureAppNamespace(ns *l4.AppNamespaces_AppNamespace) error {
	plugin.log.Infof("Configuring new AppNamespace with ID %v", ns.NamespaceId)

	// Validate data
	if ns.Interface == "" {
		return fmt.Errorf("application namespace %v does not contain interface", ns.NamespaceId)
	}

	// Check whether L4 l4ftEnabled are enabled. If not, all namespaces created earlier are added to cache
	if !plugin.l4ftEnabled {
		plugin.appNsCached.RegisterName(ns.NamespaceId, plugin.appNsIdxSeq, ns)
		plugin.appNsIdxSeq++
		plugin.log.Infof("Unable to configure application namespace %v due to disabled L4 features, moving to cache", ns.NamespaceId)
		return nil
	}

	// Find interface. If not found, add to cache for not configured namespaces
	ifIdx, _, found := plugin.ifIndexes.LookupIdx(ns.Interface)
	if !found {
		plugin.appNsCached.RegisterName(ns.NamespaceId, plugin.appNsIdxSeq, ns)
		plugin.appNsIdxSeq++
		plugin.log.Infof("Unable to configure application namespace %v due to missing interface, moving to cache", ns.NamespaceId)
		return nil
	}

	return plugin.configureAppNamespace(ns, ifIdx)
}

// ModifyAppNamespace process the NB AppNamespace config and propagates it to bin api calls
func (plugin *AppNsConfigurator) ModifyAppNamespace(newNs *l4.AppNamespaces_AppNamespace, oldNs *l4.AppNamespaces_AppNamespace) error {
	plugin.log.Infof("Modifying AppNamespace with ID %v", newNs.NamespaceId)

	// Validate data
	if newNs.Interface == "" {
		return fmt.Errorf("modified application namespace %v does not contain interface", newNs.NamespaceId)
	}

	// At first, unregister the old configuration from both mappings (if exists)
	plugin.appNsIndexes.UnregisterName(oldNs.NamespaceId)
	plugin.appNsCached.UnregisterName(oldNs.NamespaceId)

	// Check whether L4 l4ftEnabled are enabled. If not, all namespaces created earlier are added to cache
	if !plugin.l4ftEnabled {
		plugin.appNsCached.RegisterName(newNs.NamespaceId, plugin.appNsIdxSeq, newNs)
		plugin.log.Infof("Unable to modify application namespace %v due to disabled L4 features, moving to cache", newNs.NamespaceId)
		plugin.appNsIdxSeq++
		return nil
	}

	// Check interface
	ifIdx, _, found := plugin.ifIndexes.LookupIdx(newNs.Interface)
	if !found {
		plugin.appNsCached.RegisterName(newNs.NamespaceId, plugin.appNsIdxSeq, newNs)
		plugin.log.Infof("Unable to modify application namespace %v due to missing interface, moving to cache", newNs.NamespaceId)
		plugin.appNsIdxSeq++
		return nil
	}

	// TODO: remove namespace
	return plugin.configureAppNamespace(newNs, ifIdx)
}

// DeleteAppNamespace process the NB AppNamespace config and propagates it to bin api calls. This case is not currently
// supported by VPP
func (plugin *AppNsConfigurator) DeleteAppNamespace(ns *l4.AppNamespaces_AppNamespace) error {
	// TODO: implement
	plugin.log.Warn("AppNamespace removal not supported by the VPP")
	return nil
}

// ResolveCreatedInterface looks for application namespace this interface is assigned to and configures them
func (plugin *AppNsConfigurator) ResolveCreatedInterface(interfaceName string, interfaceIndex uint32) error {
	// If L4 features are not enabled, skip (and keep all in cache)
	if !plugin.l4ftEnabled {
		return nil
	}

	// Search mapping for unregistered application namespaces using the new interface
	cachedAppNs := plugin.appNsCached.LookupNamesByInterface(interfaceName)
	if len(cachedAppNs) == 0 {
		return nil
	}

	var wasErr error
	plugin.log.Infof("L4 configurator: resolving new interface %v for %d app namespaces", interfaceName, len(cachedAppNs))
	for _, appNamespace := range cachedAppNs {
		if err := plugin.configureAppNamespace(appNamespace, interfaceIndex); err != nil {
			plugin.log.Errorf("configuring app namespace %v failed: %v", appNamespace, err)
			wasErr = err
		}
		// Remove from cache
		plugin.appNsCached.UnregisterName(appNamespace.NamespaceId)
	}
	return wasErr
}

// ResolveDeletedInterface looks for application namespace this interface is assigned to and removes
func (plugin *AppNsConfigurator) ResolveDeletedInterface(interfaceName string, interfaceIndex uint32) error {

	// Search mapping for configured application namespaces using the new interface
	cachedAppNs := plugin.appNsIndexes.LookupNamesByInterface(interfaceName)
	if len(cachedAppNs) == 0 {
		return nil
	}
	plugin.log.Infof("L4 configurator: resolving deleted interface %v for %d app namespaces", interfaceName, len(cachedAppNs))
	for _, appNamespace := range cachedAppNs {
		// TODO: remove namespace. Also check whether it can be done while L4Features are disabled
		// Unregister from configured namespaces mapping
		plugin.appNsIndexes.UnregisterName(appNamespace.NamespaceId)
		// Add to un-configured. If the interface will be recreated, all namespaces are configured back
		plugin.appNsCached.RegisterName(appNamespace.NamespaceId, plugin.appNsIdxSeq, appNamespace)
		plugin.appNsIdxSeq++
	}

	return nil
}

func (plugin *AppNsConfigurator) configureAppNamespace(ns *l4.AppNamespaces_AppNamespace, ifIdx uint32) error {
	// Namespace ID
	nsID := []byte(ns.NamespaceId)

	plugin.log.Debugf("Adding App Namespace %v to interface %v", ns.NamespaceId, ifIdx)

	appNsIdx, err := plugin.l4Handler.AddAppNamespace(ns.Secret, ifIdx, ns.Ipv4FibId, ns.Ipv6FibId, nsID)
	if err != nil {
		return err
	}

	// register namespace
	plugin.appNsIndexes.RegisterName(ns.NamespaceId, appNsIdx, ns)
	plugin.log.Debugf("Application namespace %v registered", ns.NamespaceId)

	plugin.log.WithFields(logging.Fields{"appNsIdx": appNsIdx}).
		Debugf("AppNamespace %v configured", ns.NamespaceId)

	return nil
}

// An application namespace can be cached from two reasons:
// 		- the required interface was missing
//      - the L4 features were disabled
// Namespaces skipped due to the second case are configured here
func (plugin *AppNsConfigurator) resolveCachedNamespaces() error {
	cachedAppNs := plugin.appNsCached.ListNames()
	if len(cachedAppNs) == 0 {
		return nil
	}

	plugin.log.Infof("Configuring %d cached namespaces after L4 features were enabled", len(cachedAppNs))

	// Scan all registered indexes in mapping for un-configured application namespaces
	var wasErr error
	for _, name := range cachedAppNs {
		_, ns, found := plugin.appNsCached.LookupIdx(name)
		if !found {
			continue
		}

		// Check interface. If still missing, continue (keep namespace in cache)
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(ns.Interface)
		if !found {
			plugin.log.Infof("Unable to configure application namespace %v due to missing interface, keeping in cache", ns.NamespaceId)
			continue
		}

		if err := plugin.configureAppNamespace(ns, ifIdx); err != nil {
			plugin.log.Errorf("configuring app namespace %v failed: %v", ns, err)
			wasErr = err
		} else {
			// AppNamespace was configured, remove from cache
			plugin.appNsCached.UnregisterName(ns.NamespaceId)
		}
	}
	return wasErr
}
