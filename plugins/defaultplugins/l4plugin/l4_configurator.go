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

//go:generate protoc --proto_path=../common/model/l4 --gogo_out=../common/model/l4 ../common/model/l4/l4.proto
//go:generate binapi-generator --input-file=/usr/share/vpp/api/session.api.json --output-dir=../common/bin_api

package l4plugin

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l4"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/nsidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// L4Configurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/l4/l4.proto"
// and stored in ETCD under the keys "/vnf-agent/{vnf-agent}/vpp/config/v1/l4/l4ftEnabled"
// and "/vnf-agent/{vnf-agent}/vpp/config/v1/l4/namespaces/{namespace_id}".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type L4Configurator struct {
	Log logging.Logger

	ServiceLabel servicelabel.ReaderAPI
	GoVppmux     govppmux.API

	// Indexes
	SwIfIndexes  ifaceidx.SwIfIndex
	AppNsIndexes nsidx.AppNsIndexRW

	AppNsCached nsidx.AppNsIndexRW // the mapping stores not-configurable app namespaces with metadata
	AppNsIdxSeq uint32             // used only for AppNsCached; incremented after every registration

	// timer used to measure and store time
	Stopwatch *measure.Stopwatch

	// channel to communicate with the vpp
	vppChan *govppapi.Channel

	// Features flag - internal state whether the L4 l4ftEnabled are enabled or disabled
	l4ftEnabled bool
}

// Init members (channels...) and start go routines
func (plugin *L4Configurator) Init() (err error) {
	plugin.Log.Debugf("Initializing L4 configurator")

	// init vpp channel
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}

	return nil
}

// Close members, channels
func (plugin *L4Configurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// ConfigureL4FeatureFlag process the NB Features config and propagates it to bin api calls
func (plugin *L4Configurator) ConfigureL4FeatureFlag(features *l4.L4Features) error {
	plugin.Log.Info("Setting up L4 features")

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
func (plugin *L4Configurator) configureL4FeatureFlag() error {
	plugin.Log.Info("Configuring L4 features")

	if err := vppcalls.EnableL4Features(plugin.vppChan); err != nil {
		plugin.Log.Errorf("Enabling L4 features failed: %v", err)
		return err
	}
	plugin.l4ftEnabled = true
	plugin.Log.Infof("L4 features enabled")

	return nil
}

// DeleteL4FeatureFlag process the NB Features config and propagates it to bin api calls
func (plugin *L4Configurator) DeleteL4FeatureFlag() error {
	plugin.Log.Info("Removing L4 features")

	if err := vppcalls.DisableL4Features(plugin.vppChan); err != nil {
		plugin.Log.Errorf("Disabling L4 features failed: %v", err)
		return err
	}

	plugin.l4ftEnabled = false
	plugin.Log.Infof("L4 features disabled")

	return nil
}

// ConfigureAppNamespace process the NB AppNamespace config and propagates it to bin api calls
func (plugin *L4Configurator) ConfigureAppNamespace(ns *l4.AppNamespaces_AppNamespace) error {
	plugin.Log.Infof("Configuring new AppNamespace with ID %v", ns.NamespaceId)

	// Validate data
	if ns.Interface == "" {
		return fmt.Errorf("application namespace %v does not contain interface", ns.NamespaceId)
	}

	// Check whether L4 l4ftEnabled are enabled. If not, all namespaces created earlier are added to cache
	if !plugin.l4ftEnabled {
		plugin.AppNsCached.RegisterName(ns.NamespaceId, plugin.AppNsIdxSeq, ns)
		plugin.AppNsIdxSeq++
		plugin.Log.Infof("Unable to configure application namespace %v due to disabled L4 features, moving to cache", ns.NamespaceId)
		return nil
	}

	// Find interface. If not found, add to cache for not configured namespaces
	ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(ns.Interface)
	if !found {
		plugin.AppNsCached.RegisterName(ns.NamespaceId, plugin.AppNsIdxSeq, ns)
		plugin.AppNsIdxSeq++
		plugin.Log.Infof("Unable to configure application namespace %v due to missing interface, moving to cache", ns.NamespaceId)
		return nil
	}

	return plugin.configureAppNamespace(ns, ifIdx)
}

// ModifyAppNamespace process the NB AppNamespace config and propagates it to bin api calls
func (plugin *L4Configurator) ModifyAppNamespace(newNs *l4.AppNamespaces_AppNamespace, oldNs *l4.AppNamespaces_AppNamespace) error {
	plugin.Log.Infof("Modifying AppNamespace with ID %v", newNs.NamespaceId)

	// Validate data
	if newNs.Interface == "" {
		return fmt.Errorf("modified application namespace %v does not contain interface", newNs.NamespaceId)
	}

	// At first, unregister the old configuration from both mappings (if exists)
	plugin.AppNsIndexes.UnregisterName(oldNs.NamespaceId)
	plugin.AppNsCached.UnregisterName(oldNs.NamespaceId)

	// Check whether L4 l4ftEnabled are enabled. If not, all namespaces created earlier are added to cache
	if !plugin.l4ftEnabled {
		plugin.AppNsCached.RegisterName(newNs.NamespaceId, plugin.AppNsIdxSeq, newNs)
		plugin.Log.Infof("Unable to modify application namespace %v due to disabled L4 features, moving to cache", newNs.NamespaceId)
		plugin.AppNsIdxSeq++
		return nil
	}

	// Check interface
	ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(newNs.Interface)
	if !found {
		plugin.AppNsCached.RegisterName(newNs.NamespaceId, plugin.AppNsIdxSeq, newNs)
		plugin.Log.Infof("Unable to modify application namespace %v due to missing interface, moving to cache", newNs.NamespaceId)
		plugin.AppNsIdxSeq++
		return nil
	}

	// TODO: remove namespace
	return plugin.configureAppNamespace(newNs, ifIdx)
}

// DeleteAppNamespace process the NB AppNamespace config and propagates it to bin api calls. This case is not currently
// supported by VPP
func (plugin *L4Configurator) DeleteAppNamespace(ns *l4.AppNamespaces_AppNamespace) error {
	// TODO: implement
	plugin.Log.Warn("AppNamespace removal not supported by the VPP")
	return nil
}

// ResolveCreatedInterface looks for application namespace this interface is assigned to and configures them
func (plugin *L4Configurator) ResolveCreatedInterface(interfaceName string, interfaceIndex uint32) error {
	// If L4 features are not enabled, skip (and keep all in cache)
	if !plugin.l4ftEnabled {
		return nil
	}

	// Search mapping for unregistered application namespaces using the new interface
	cachedAppNs := plugin.AppNsCached.LookupNamesByInterface(interfaceName)
	if len(cachedAppNs) == 0 {
		return nil
	}

	var wasErr error
	plugin.Log.Infof("L4 configurator: resolving new interface %v for %d app namespaces", interfaceName, len(cachedAppNs))
	for _, appNamespace := range cachedAppNs {
		if err := plugin.configureAppNamespace(appNamespace, interfaceIndex); err != nil {
			plugin.Log.Errorf("configuring app namespace %v failed: %v", appNamespace, err)
			wasErr = err
		}
		// Remove from cache
		plugin.AppNsCached.UnregisterName(appNamespace.NamespaceId)
	}
	return wasErr
}

// ResolveDeletedInterface looks for application namespace this interface is assigned to and removes
func (plugin *L4Configurator) ResolveDeletedInterface(interfaceName string, interfaceIndex uint32) error {

	// Search mapping for configured application namespaces using the new interface
	cachedAppNs := plugin.AppNsIndexes.LookupNamesByInterface(interfaceName)
	if len(cachedAppNs) == 0 {
		return nil
	}
	plugin.Log.Infof("L4 configurator: resolving deleted interface %v for %d app namespaces", interfaceName, len(cachedAppNs))
	for _, appNamespace := range cachedAppNs {
		// TODO: remove namespace. Also check whether it can be done while L4Features are disabled
		// Unregister from configured namespaces mapping
		plugin.AppNsIndexes.UnregisterName(appNamespace.NamespaceId)
		// Add to un-configured. If the interface will be recreated, all namespaces are configured back
		plugin.AppNsCached.RegisterName(appNamespace.NamespaceId, plugin.AppNsIdxSeq, appNamespace)
		plugin.AppNsIdxSeq++
	}

	return nil
}

func (plugin *L4Configurator) configureAppNamespace(ns *l4.AppNamespaces_AppNamespace, ifIdx uint32) error {
	// Namespace ID
	nsID := []byte(ns.NamespaceId)

	plugin.Log.Debugf("Adding App Namespace %v to interface %v", ns.NamespaceId, ifIdx)

	appNsIdx, err := vppcalls.AddAppNamespace(ns.Secret, ifIdx, ns.Ipv4FibId, ns.Ipv6FibId, nsID, plugin.vppChan, plugin.Stopwatch)
	if err != nil {
		return err
	}

	// register namespace
	plugin.AppNsIndexes.RegisterName(ns.NamespaceId, appNsIdx, ns)
	plugin.Log.Debugf("Application namespace %v registered", ns.NamespaceId)

	plugin.Log.WithFields(logging.Fields{"appNsIdx": appNsIdx}).
		Debugf("AppNamespace %v configured", ns.NamespaceId)

	return nil
}

// An application namespace can be cached from two reasons:
// 		- the required interface was missing
//      - the L4 features were disabled
// Namespaces skipped due to the second case are configured here
func (plugin *L4Configurator) resolveCachedNamespaces() error {
	cachedAppNs := plugin.AppNsCached.ListNames()
	if len(cachedAppNs) == 0 {
		return nil
	}

	plugin.Log.Infof("Configuring %d cached namespaces after L4 features were enabled", len(cachedAppNs))

	// Scan all registered indexes in mapping for un-configured application namespaces
	var wasErr error
	for _, name := range cachedAppNs {
		_, ns, found := plugin.AppNsCached.LookupIdx(name)
		if !found {
			continue
		}

		// Check interface. If still missing, continue (keep namespace in cache)
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(ns.Interface)
		if !found {
			plugin.Log.Infof("Unable to configure application namespace %v due to missing interface, keeping in cache", ns.NamespaceId)
			continue
		}

		if err := plugin.configureAppNamespace(ns, ifIdx); err != nil {
			plugin.Log.Errorf("configuring app namespace %v failed: %v", ns, err)
			wasErr = err
		} else {
			// AppNamespace was configured, remove from cache
			plugin.AppNsCached.UnregisterName(ns.NamespaceId)
		}
	}
	return wasErr
}
