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

// Package l3plugin implements the L3 plugin that handles L3 FIBs.
package l3plugin

import (
	"fmt"
	"strconv"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// VrfConfigurator is for managing VRF tables
type VrfConfigurator struct {
	Log logging.Logger

	GoVppmux      govppmux.API
	TableIndexes  idxvpp.NameToIdxRW
	TableIndexSeq uint32
	SwIfIndexes   ifaceidx.SwIfIndex
	vppChan       *govppapi.Channel
}

// Init members (channels...) and start go routines
func (plugin *VrfConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing L3 VRF")

	// Init VPP API channel
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}

	return plugin.checkMsgCompatibility()}

// Creates unique identifier which serves as a name in name to index mapping
func tableIdentifier(vrf uint32) string {
	return fmt.Sprintf("vrftable-%v", vrf)
}

// AddTable creates new table
func (plugin *VrfConfigurator) AddTable(config *l3.VrfTable, vrfFromKey string) error {
	plugin.Log.Infof("Creating new VRF table %s (ID: %v)", config.Name, config.VrfId)
	// Validate VRF index from key and it's value in data
	if err := plugin.validateVrfFromKey(config, vrfFromKey); err != nil {
		return err
	}
	// Transform table data
	table, err := TransformVrfTable(config, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	plugin.Log.Debugf("adding table: %+v", table)
	// Create and register new table
	if table != nil {
		if err := vppcalls.VppAddIPTable(table, plugin.vppChan); err != nil {
			return err
		}
		identifier := tableIdentifier(table.TableID)
		plugin.TableIndexes.RegisterName(identifier, plugin.TableIndexSeq, nil)
		plugin.TableIndexSeq++
		plugin.Log.Infof("Table %v registered", identifier)
	}

	return nil
}

/*
// ModifyRoute process the NB config and propagates it to bin api calls
func (plugin *VrfConfigurator) ModifyRoute(newConfig *l3.VrfTable, oldConfig *l3.VrfTable, vrfFromKey string) error {
	plugin.Log.Infof("Modifying route %v -> %v", oldConfig.DstIpAddr, oldConfig.NextHopAddr)
	// Validate old route data Vrf
	if err := plugin.validateVrfFromKey(oldConfig, vrfFromKey); err != nil {
		return err
	}
	// Transform old route data
	oldRoute, err := TransformRoute(oldConfig, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	// Remove and unregister old route
	err = vppcalls.VppDelRoute(oldRoute, plugin.vppChan)
	if err != nil {
		return err
	}
	oldRouteIdentifier := routeIdentifier(oldRoute.VrfID, oldRoute.DstAddr.String(), oldRoute.NextHopAddr.String())
	_, _, found := plugin.RouteIndexes.UnregisterName(oldRouteIdentifier)
	if found {
		plugin.Log.Infof("Old route %v unregistered", oldRouteIdentifier)
	} else {
		plugin.Log.Warnf("Unregister failed, old route %v not found", oldRouteIdentifier)
	}

	// Validate new route data Vrf
	if err := plugin.validateVrfFromKey(newConfig, vrfFromKey); err != nil {
		return err
	}
	// Transform new route data
	newRoute, err := TransformRoute(newConfig, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	// Create and register new route
	err = vppcalls.VppAddRoute(newRoute, plugin.vppChan)
	if err != nil {
		return err
	}
	newRouteIdentifier := routeIdentifier(newRoute.VrfID, newRoute.DstAddr.String(), newRoute.NextHopAddr.String())
	plugin.RouteIndexes.RegisterName(newRouteIdentifier, plugin.RouteIndexSeq, nil)
	plugin.RouteIndexSeq++
	plugin.Log.Infof("New route %v registered", newRouteIdentifier)

	return nil
}

// DeleteRoute process the NB config and propagates it to bin api calls
func (plugin *VrfConfigurator) DeleteRoute(config *l3.VrfTable, vrfFromKey string) (wasError error) {
	plugin.Log.Infof("Removing route %v -> %v", config.DstIpAddr, config.NextHopAddr)
	// Validate VRF index from key and it's value in data
	if err := plugin.validateVrfFromKey(config, vrfFromKey); err != nil {
		return err
	}
	// Transform route data
	route, err := TransformRoute(config, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	if route == nil {
		return nil
	}
	plugin.Log.Debugf("deleting route: %+v", route)
	// Remove and unregister route
	err = vppcalls.VppDelRoute(route, plugin.vppChan)
	if err != nil {
		return err
	}
	routeIdentifier := routeIdentifier(route.VrfID, route.DstAddr.String(), route.NextHopAddr.String())
	_, _, found := plugin.RouteIndexes.UnregisterName(routeIdentifier)
	if found {
		plugin.Log.Infof("Route %v unregistered", routeIdentifier)
	} else {
		plugin.Log.Warnf("Unregister failed, route %v not found", routeIdentifier)
	}

	return nil
}*/

func (plugin *VrfConfigurator) validateVrfFromKey(config *l3.VrfTable, vrfFromKey string) error {
	intVrfFromKey, err := strconv.Atoi(vrfFromKey)
	if intVrfFromKey != int(config.VrfId) {
		if err != nil {
			return err
		}
		plugin.Log.Warnf("VRF index from key (%v) does not match config (%v), using value from the key",
			intVrfFromKey, config.VrfId)
		config.VrfId = uint32(intVrfFromKey)
	}
	return nil
}

func (plugin *VrfConfigurator) checkMsgCompatibility() error {
	msgs := []govppapi.Message{
		&ip.IPTableAddDel{},
		&ip.IPTableAddDelReply{},
	}
	err := plugin.vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		plugin.Log.Error(err)
	}
	return err
}

// Close GOVPP channel
func (plugin *VrfConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// TransformVrfTable transforms table data for VPP
func TransformVrfTable(input *l3.VrfTable, index ifaceidx.SwIfIndex, log logging.Logger) (*vppcalls.IPTable, error) {
	if input == nil {
		log.Infof("Table input is empty")
		return nil, nil
	}
	vrfID := input.VrfId
	//isIPv6 := input.IsIpv6
	if input.Name == "" {
		name := fmt.Sprintf("vrf_table_%03d", vrfID)
		log.Infof("Route did not contain name, will use %q", name)
		input.Name = name
		//return nil, nil
	}

	output := &vppcalls.IPTable{
		TableID: vrfID,
		//IsIPv6:  isIPv6,
		Name: []byte(input.Name),
	}
	return output, nil
}
