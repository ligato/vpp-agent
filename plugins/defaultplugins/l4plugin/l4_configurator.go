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

	NotConfiguredAppNs nsidx.AppNsIndexRW // the mapping stores not-configurable app namespaces with metadata
	AppNsIdxSeq        uint32             // used only for NotConfiguredAppNs; incremented after every registration

	// timer used to measure and store time
	Stopwatch *measure.Stopwatch

	// channel to communicate with the vpp
	vppCh *govppapi.Channel

	// Features flag - internal state whether the L4 l4ftEnabled are enabled or disabled
	l4ftEnabled bool
}

// Init members (channels...) and start go routines
func (plugin *L4Configurator) Init() error {
	plugin.Log.Debugf("Initializing L4 configurator")
	var err error

	// init vpp channel
	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	return nil
}

// Close members, channels
func (plugin *L4Configurator) Close() error {
	return nil
}

// ConfigureL4FeatureFlag process the NB Features config and propagates it to bin api calls
func (plugin *L4Configurator) ConfigureL4FeatureFlag(features *l4.L4Features) error {
	plugin.Log.Info("Setting up L4 features config")

	if features.Enabled {
		if err := vppcalls.EnableL4Features(plugin.Log, plugin.vppCh); err != nil {
			return err
		}
		plugin.l4ftEnabled = true
		plugin.Log.Infof("L4 features enabled")

		return plugin.resolveCachedNamespaces()

	}
	if err := vppcalls.DisableL4Features(plugin.Log, plugin.vppCh); err != nil {
		return err
	}
	plugin.l4ftEnabled = false
	plugin.Log.Infof("L4 features disabled")

	return nil
}

// DeleteL4FeatureFlag process the NB Features config and propagates it to bin api calls
func (plugin *L4Configurator) DeleteL4FeatureFlag() error {
	plugin.Log.Info("Removing up L4 features config")

	if err := vppcalls.DisableL4Features(plugin.Log, plugin.vppCh); err != nil {
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
		plugin.NotConfiguredAppNs.RegisterName(ns.NamespaceId, plugin.AppNsIdxSeq, ns)
		plugin.Log.Infof("Unable to configure application namespace %v due to disabled L4 features, moving to cache", ns.NamespaceId)
		plugin.AppNsIdxSeq++
		return nil
	}

	// Find interface. If not found, add to cache for not configured namespaces
	ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(ns.Interface)
	if !found {
		plugin.NotConfiguredAppNs.RegisterName(ns.NamespaceId, plugin.AppNsIdxSeq, ns)
		plugin.Log.Infof("Unable to configure application namespace %v due to missing interface, moving to cache", ns.NamespaceId)
		plugin.AppNsIdxSeq++
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
	plugin.NotConfiguredAppNs.UnregisterName(oldNs.NamespaceId)

	// Check whether L4 l4ftEnabled are enabled. If not, all namespaces created earlier are added to cache
	if !plugin.l4ftEnabled {
		plugin.NotConfiguredAppNs.RegisterName(newNs.NamespaceId, plugin.AppNsIdxSeq, newNs)
		plugin.Log.Infof("Unable to modify application namespace %v due to disabled L4 features, moving to cache", newNs.NamespaceId)
		plugin.AppNsIdxSeq++
		return nil
	}

	// Check interface
	ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(newNs.Interface)
	if !found {
		plugin.NotConfiguredAppNs.RegisterName(newNs.NamespaceId, plugin.AppNsIdxSeq, newNs)
		plugin.Log.Infof("Unable to modify application namespace %v due to missing interface, moving to cache", newNs.NamespaceId)
		plugin.AppNsIdxSeq++
		return nil
	}

	// todo remove namespace
	return plugin.configureAppNamespace(newNs, ifIdx)
}

// DeleteAppNamespace process the NB AppNamespace config and propagates it to bin api calls. This case is not currently
// supported by VPP
func (plugin *L4Configurator) DeleteAppNamespace(ns *l4.AppNamespaces_AppNamespace) error {
	// todo implement
	plugin.Log.Warn("AppNamespace removal not supported by the VPP")
	return nil
}

// ResolveCreatedInterface looks for application namespace this interface is assigned to and configures them
func (plugin *L4Configurator) ResolveCreatedInterface(interfaceName string, interfaceIndex uint32) error {
	plugin.Log.Infof("L4 configurator: resolving new interface %v", interfaceName)

	// If L4 features are not enabled, skip (and keep all in cache)
	if !plugin.l4ftEnabled {
		return nil
	}

	// Search mapping for unregistered application namespaces using the new interface
	var wasErr error
	appNamespaces := plugin.NotConfiguredAppNs.LookupNamesByInterface(interfaceName)
	if len(appNamespaces) > 0 {
		plugin.Log.Debugf("Found %v app namespaces for interface %v", len(appNamespaces), interfaceName)
		for _, appNamespace := range appNamespaces {
			if err := plugin.configureAppNamespace(appNamespace, interfaceIndex); err != nil {
				plugin.Log.Error(err)
				wasErr = err
			}
			// Remove from cache
			plugin.NotConfiguredAppNs.UnregisterName(appNamespace.NamespaceId)
		}
	}

	return wasErr
}

// ResolveDeletedInterface looks for application namespace this interface is assigned to and removes
func (plugin *L4Configurator) ResolveDeletedInterface(interfaceName string, interfaceIndex uint32) error {
	plugin.Log.Infof("L4 configurator: resolving deleted interface %v", interfaceName)

	// Search mapping for configured application namespaces using the new interface
	confAppNs := plugin.AppNsIndexes.LookupNamesByInterface(interfaceName)
	if len(confAppNs) > 0 {
		plugin.Log.Debugf("Found %v app namespaces belonging to removed interface %v", len(confAppNs), interfaceName)
		for _, appNamespace := range confAppNs {
			// todo remove namespace. Also check whether it can be done while L4Features are disabled
			// Unregister from configured namespaces mapping
			plugin.AppNsIndexes.UnregisterName(appNamespace.NamespaceId)
			// Add to un-configured. If the interface will be recreated, all namespaces are configured back
			plugin.NotConfiguredAppNs.RegisterName(appNamespace.NamespaceId, plugin.AppNsIdxSeq, appNamespace)
			plugin.AppNsIdxSeq++
		}
	}

	return nil
}

func (plugin *L4Configurator) configureAppNamespace(ns *l4.AppNamespaces_AppNamespace, ifIdx uint32) error {
	// Namespace ID
	nsID := []byte(ns.NamespaceId)

	appnsIndex, err := vppcalls.AddAppNamespace(ns.Secret, ifIdx, ns.Ipv4FibId, ns.Ipv6FibId, nsID, plugin.Log, plugin.vppCh)
	if err != nil {
		return err
	}

	// register namespace
	plugin.AppNsIndexes.RegisterName(ns.NamespaceId, appnsIndex, ns)
	plugin.Log.Debugf("Application namespace %v registered", ns)

	return nil
}

// An application namespace can be cached from two reasons:
// 		- the required interface was missing
//      - the L4 features were disabled
// Namespaces skipped due to the second case are configured here
func (plugin *L4Configurator) resolveCachedNamespaces() error {
	plugin.Log.Info("Configuring cached namespaces after L4 features were enabled")

	// Scan all registered indexes in mapping for un-configured application namespaces
	var wasErr error
	for _, name := range plugin.NotConfiguredAppNs.ListNames() {
		_, ns, found := plugin.NotConfiguredAppNs.LookupIdx(name)
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
			plugin.Log.Error(err)
			wasErr = err
		} else {
			// AppNamespace was configured, remove from cache
			plugin.NotConfiguredAppNs.UnregisterName(ns.NamespaceId)
		}
	}

	return wasErr
}
