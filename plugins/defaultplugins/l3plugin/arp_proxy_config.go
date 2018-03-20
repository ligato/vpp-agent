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

package l3plugin

import (
	"fmt"
	"net"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

var msgCompatibilityProxyARP = []govppapi.Message{
	&ip.ProxyArpIntfcEnableDisable{},
	&ip.ProxyArpIntfcEnableDisableReply{},
	&ip.ProxyArpAddDel{},
	&ip.ProxyArpAddDelReply{},
}

// ProxyArpConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 proxy arp entries as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/proxyarp". Configuration uses separate keys
// for proxy arp range and interfaces. Updates received from the northbound API are compared with the VPP
// run-time configuration and differences are applied through the VPP binary API.
type ProxyArpConfigurator struct {
	Log logging.Logger

	GoVppmux govppmux.API

	// ProxyArpIndices is a list of proxy ARP interface entries which are successfully configured on the VPP
	ProxyArpIfIndices idxvpp.NameToIdxRW
	// ProxyArpRngIndices is a list of proxy ARP range entries which are successfully configured on the VPP
	ProxyArpRngIndices idxvpp.NameToIdxRW
	// Cached interfaces
	ProxyARPIfCache []string

	ProxyARPIndexSeq uint32
	SwIfIndexes      ifaceidx.SwIfIndex
	vppChan          *govppapi.Channel

	Stopwatch *measure.Stopwatch
}

// Init VPP channel and check message compatibility
func (plugin *ProxyArpConfigurator) Init() error {
	plugin.Log.Debugf("Initializing proxy ARP configurator")

	// Init VPP API channel
	var err error
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	if err := plugin.vppChan.CheckMessageCompatibility(msgCompatibilityProxyARP...); err != nil {
		plugin.Log.Error(err)
		return err
	}

	return nil
}

// Close VPP channel
func (plugin *ProxyArpConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

func (plugin *ProxyArpConfigurator) AddInterface(pArpIf *l3.ProxyArpInterfaces_InterfaceList) error {
	plugin.Log.Infof("Enabling interfaces from proxy ARP config %s", pArpIf.Label)

	var wasErr error
	for _, proxyArpIf := range pArpIf.Interfaces {
		ifName := proxyArpIf.Name
		if ifName == "" {
			err := fmt.Errorf("proxy ARP interface not set")
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		// Check interface, cache if does not exist
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(ifName)
		if !found {
			plugin.Log.Debugf("Interface %s does not exist, moving to cache", ifName)
			plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache, ifName)
			continue
		}

		// Call VPP API to enable interface for proxy ARP
		if err := vppcalls.EnableProxyArpInterface(ifIdx, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
			plugin.Log.Debugf("Interface %s enabled for proxy ARP", ifName)
		} else {
			err := fmt.Errorf("enabling interface %s for proxy ARP failed: %v", ifName, err)
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
	}
	// Register
	plugin.ProxyArpIfIndices.RegisterName(pArpIf.Label, plugin.ProxyARPIndexSeq, nil)
	plugin.ProxyARPIndexSeq++
	plugin.Log.Debugf("Proxy ARP interface configuration %s registered", pArpIf.Label)

	return wasErr
}

// ModifyInterface does nothing
func (plugin *ProxyArpConfigurator) ModifyInterface(newPArpIf, oldPArpIf *l3.ProxyArpInterfaces_InterfaceList) error {
	plugin.Log.Infof("Modifying proxy ARP interface configuration %s", newPArpIf.Label)

	toEnable, toDisable := plugin.calculateIfDiff(newPArpIf.Interfaces, oldPArpIf.Interfaces)
	var wasErr error
	// Disable obsolete interfaces
	for _, ifName := range toDisable {
		// Check cache
		for idx, cachedIf := range plugin.ProxyARPIfCache {
			if cachedIf == ifName {
				plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache[:idx], plugin.ProxyARPIfCache[idx+1:]...)
				plugin.Log.Debugf("Proxy ARP interface %s removed from cache", ifName)
				continue
			}
		}
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(ifName)
		// If interface is not found, there is nothing to do
		if found {
			if err := vppcalls.DisableProxyArpInterface(ifIdx, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
				plugin.Log.Debugf("Interface %s disabled for proxy ARP", ifName)
			} else {
				err = fmt.Errorf("disabling interface %s for proxy ARP failed: %v", ifName, err)
				plugin.Log.Error(err)
				wasErr = err
			}
		}
	}
	// Enable new interfaces
	for _, ifName := range toEnable {
		// Put to cache if interface is missing
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(ifName)
		if !found {
			plugin.Log.Debugf("Interface %s does not exist, moving to cache", ifName)
			plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache, ifName)
			continue
		}
		// Configure
		if err := vppcalls.EnableProxyArpInterface(ifIdx, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
			plugin.Log.Debugf("Interface %s enabled for proxy ARP", ifName)
		} else {
			err := fmt.Errorf("enabling interface %s for proxy ARP failed: %v", ifName, err)
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
	}

	plugin.Log.Debugf("Proxy ARP interface config %s modification done", newPArpIf.Label)

	return wasErr
}

// DeleteInterface disables proxy ARP interface or removes it from cache
func (plugin *ProxyArpConfigurator) DeleteInterface(pArpIf *l3.ProxyArpInterfaces_InterfaceList) error {
	plugin.Log.Infof("Disabling interfaces from proxy ARP config %s", pArpIf.Label)

	var wasErr error
ProxyArpIfLoop:
	for _, proxyArpIf := range pArpIf.Interfaces {
		ifName := proxyArpIf.Name
		// Check if interface is cached
		for idx, cachedIf := range plugin.ProxyARPIfCache {
			if cachedIf == ifName {
				plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache[:idx], plugin.ProxyARPIfCache[idx+1:]...)
				plugin.Log.Debugf("Proxy ARP interface %s removed from cache", ifName)
				continue ProxyArpIfLoop
			}
		}
		// Look for interface
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(ifName)
		if !found {
			// Interface does not exist, nothing more to do
			continue
		}
		// Call VPP API to disable interface for proxy ARP
		if err := vppcalls.DisableProxyArpInterface(ifIdx, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
			plugin.Log.Debugf("Interface %s disabled for proxy ARP", ifName)
		} else {
			err = fmt.Errorf("disabling interface %s for proxy ARP failed: %v", ifName, err)
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
	}

	// Un-register
	plugin.ProxyArpIfIndices.UnregisterName(pArpIf.Label)
	plugin.Log.Debugf("Proxy ARP interface config %s un-registered", pArpIf.Label)

	return wasErr
}

// AddRange configures new IP range for proxy ARP
func (plugin *ProxyArpConfigurator) AddRange(pArpRng *l3.ProxyArpRanges_RangeList) error {
	plugin.Log.Infof("Setting up proxy ARP IP range config %s", pArpRng.Label)

	var wasErr error
	for _, proxyArpRange := range pArpRng.Ranges {
		// Prune addresses
		firstIP, err := plugin.pruneIP(proxyArpRange.FirstIp)
		if err != nil {
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		lastIP, err := plugin.pruneIP(proxyArpRange.LastIp)
		if err != nil {
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		// Convert to byte representation
		bFirstIP := net.ParseIP(firstIP).To4()
		bLastIP := net.ParseIP(lastIP).To4()
		// Call VPP API to configure IP range for proxy ARP
		if err := vppcalls.AddProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
			plugin.Log.Debugf("Address range %s - %s configured for proxy ARP", firstIP, lastIP)
		} else {
			err := fmt.Errorf("failed to configure proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
	}

	// Register
	plugin.ProxyArpRngIndices.RegisterName(pArpRng.Label, plugin.ProxyARPIndexSeq, nil)
	plugin.ProxyARPIndexSeq++
	plugin.Log.Debugf("Proxy ARP range config %s registered", pArpRng.Label)

	return wasErr
}

// ModifyRange does nothing
func (plugin *ProxyArpConfigurator) ModifyRange(newPArpRng, oldPArpRng *l3.ProxyArpRanges_RangeList) error {
	plugin.Log.Infof("Modifying proxy ARP range config %s", oldPArpRng.Label)

	toAdd, toDelete := plugin.calculateRngDiff(newPArpRng.Ranges, oldPArpRng.Ranges)
	var wasErr error
	// Remove old ranges
	for _, rng := range toDelete {
		// Prune
		firstIP, err := plugin.pruneIP(rng.FirstIp)
		if err != nil {
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		lastIP, err := plugin.pruneIP(rng.LastIp)
		if err != nil {
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		// Convert to byte representation
		bFirstIP := net.ParseIP(firstIP).To4()
		bLastIP := net.ParseIP(lastIP).To4()
		// Call VPP API to configure IP range for proxy ARP
		if err := vppcalls.DeleteProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
			plugin.Log.Debugf("Address range %s - %s removed from proxy ARP setup", firstIP, lastIP)
		} else {
			err = fmt.Errorf("failed to remove proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
	}
	// Add new ranges
	for _, rng := range toAdd {
		// Prune addresses
		firstIP, err := plugin.pruneIP(rng.FirstIp)
		if err != nil {
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		lastIP, err := plugin.pruneIP(rng.LastIp)
		if err != nil {
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		// Convert to byte representation
		bFirstIP := net.ParseIP(firstIP).To4()
		bLastIP := net.ParseIP(lastIP).To4()
		// Call VPP API to configure IP range for proxy ARP
		if err := vppcalls.AddProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
			plugin.Log.Debugf("Address range %s - %s configured for proxy ARP", firstIP, lastIP)
		} else {
			err := fmt.Errorf("failed to configure proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
	}

	plugin.Log.Debugf("Proxy ARP range config %s modification done", newPArpRng.Label)

	return wasErr
}

func (plugin *ProxyArpConfigurator) DeleteRange(pArpRng *l3.ProxyArpRanges_RangeList) error {
	plugin.Log.Infof("Removing proxy ARP IP range config %s", pArpRng.Label)

	var wasErr error
	for _, proxyArpRange := range pArpRng.Ranges {
		// Prune addresses
		firstIP, err := plugin.pruneIP(proxyArpRange.FirstIp)
		if err != nil {
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		lastIP, err := plugin.pruneIP(proxyArpRange.LastIp)
		if err != nil {
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
		// Convert to byte representation
		bFirstIP := net.ParseIP(firstIP).To4()
		bLastIP := net.ParseIP(lastIP).To4()
		// Call VPP API to configure IP range for proxy ARP
		if err := vppcalls.DeleteProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
			plugin.Log.Debugf("Address range %s - %s removed from proxy ARP setup", firstIP, lastIP)
		} else {
			err = fmt.Errorf("failed to remove proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
			plugin.Log.Error(err)
			wasErr = err
			continue
		}
	}

	// Un-register
	plugin.ProxyArpIfIndices.UnregisterName(pArpRng.Label)
	plugin.Log.Debugf("Proxy ARP range config %s un-registered", pArpRng.Label)

	return wasErr
}

// ResolveCreatedInterface handles new registered interface for proxy ARP
func (plugin *ProxyArpConfigurator) ResolveCreatedInterface(ifName string, ifIdx uint32) error {
	plugin.Log.Debugf("Proxy ARP: handling new interface %s", ifName)

	// Look for interface in cache
	var wasErr error
	for idx, cachedIf := range plugin.ProxyARPIfCache {
		if cachedIf == ifName {
			// Configure cached interface
			if err := vppcalls.EnableProxyArpInterface(ifIdx, plugin.vppChan, plugin.Log, plugin.Stopwatch); err != nil {
				plugin.Log.Error(err)
				wasErr = err
			}
			// Remove from cache
			plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache[:idx], plugin.ProxyARPIfCache[idx+1:]...)
			plugin.Log.Debugf("Proxy ARP interface %s configured and removed from cache", ifName)
			return wasErr
		}
	}

	return wasErr
}

// ResolveDeletedInterface handles new registered interface for proxy ARP
func (plugin *ProxyArpConfigurator) ResolveDeletedInterface(ifName string) error {
	plugin.Log.Debugf("Proxy ARP: handling removed interface %s", ifName)

	// Check if interface was enabled for proxy ARP
	_, _, found := plugin.ProxyArpIfIndices.LookupIdx(ifName)
	if found {
		// Put interface to cache (no need to call delete)
		plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache, ifName)
		plugin.Log.Debugf("Removed interface %s was configured for proxy ARP, added to cache", ifName)
	}

	return nil
}

// Remove IP mask if set
func (plugin *ProxyArpConfigurator) pruneIP(ip string) (string, error) {
	ipParts := strings.Split(ip, "/")
	if len(ipParts) == 1 {
		return ipParts[0], nil
	}
	if len(ipParts) == 2 {
		plugin.Log.Warnf("Proxy ARP range: removing unnecessary mask from IP address %s", ip)
		return ipParts[0], nil
	}
	return ip, fmt.Errorf("proxy ARP range: invalid IP address format: %s", ip)
}

// Calculate difference between old and new interfaces
func (plugin *ProxyArpConfigurator) calculateIfDiff(newIfs, oldIfs []*l3.ProxyArpInterfaces_InterfaceList_Interface) (toEnable, toDisable []string) {
	// Find missing new interfaces
	for _, newIf := range newIfs {
		var found bool
		for _, oldIf := range oldIfs {
			if newIf.Name == oldIf.Name {
				found = true
			}
		}
		if !found {
			toEnable = append(toEnable, newIf.Name)
		}
	}
	// Find obsolete interfaces
	for _, oldIf := range oldIfs {
		var found bool
		for _, newIf := range newIfs {
			if oldIf.Name == newIf.Name {
				found = true
			}
		}
		if !found {
			toDisable = append(toDisable, oldIf.Name)
		}
	}
	return
}

// Calculate difference between old and new ranges
func (plugin *ProxyArpConfigurator) calculateRngDiff(newRngs, oldRngs []*l3.ProxyArpRanges_RangeList_Range) (toAdd, toDelete []*l3.ProxyArpRanges_RangeList_Range) {
	// Find missing ranges
	for _, newRng := range newRngs {
		var found bool
		for _, oldRng := range oldRngs {
			if newRng.FirstIp == oldRng.FirstIp && newRng.LastIp == oldRng.LastIp {
				found = true
			}
		}
		if !found {
			toAdd = append(toAdd, newRng)
		}
	}
	// Find obsolete interfaces
	for _, oldRng := range oldRngs {
		var found bool
		for _, newRng := range newRngs {
			if oldRng.FirstIp == newRng.FirstIp && oldRng.LastIp == newRng.LastIp {
				found = true
			}
		}
		if !found {
			toDelete = append(toDelete, oldRng)
		}
	}
	return

}
