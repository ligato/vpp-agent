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
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
)

// ProxyArpConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 proxy arp entries as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/proxyarp". Configuration uses separate keys
// for proxy arp range and interfaces. Updates received from the northbound API are compared with the VPP
// run-time configuration and differences are applied through the VPP binary API.
type ProxyArpConfigurator struct {
	log logging.Logger

	// Mappings
	ifIndexes ifaceidx.SwIfIndex
	// ProxyArpIndices is a list of proxy ARP interface entries which are successfully configured on the VPP
	pArpIfIndexes idxvpp.NameToIdxRW
	// ProxyArpRngIndices is a list of proxy ARP range entries which are successfully configured on the VPP
	pArpRngIndexes idxvpp.NameToIdxRW
	// Cached interfaces
	pArpIfCache  []string
	pArpIndexSeq uint32

	// VPP channel
	vppChan *govppapi.Channel

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// Init VPP channel and check message compatibility
func (plugin *ProxyArpConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-l3-proxy-arp-conf")
	plugin.log.Debugf("Initializing proxy ARP configurator")

	// Mappings
	plugin.ifIndexes = swIfIndexes
	plugin.pArpIfIndexes = nametoidx.NewNameToIdx(plugin.log, "proxyarp_if_indices", nil)
	plugin.pArpRngIndexes = nametoidx.NewNameToIdx(plugin.log, "proxyarp_rng_indices", nil)
	plugin.pArpIndexSeq = 1

	// VPP channel
	plugin.vppChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Stopwatch
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("ProxyARPConfigurator", plugin.log)
	}

	// Message compatibility
	if err := plugin.vppChan.CheckMessageCompatibility(vppcalls.ProxyArpMessages...); err != nil {
		plugin.log.Error(err)
		return err
	}

	return nil
}

// Close VPP channel
func (plugin *ProxyArpConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (plugin *ProxyArpConfigurator) clearMapping() {
	plugin.pArpIfIndexes.Clear()
	plugin.pArpRngIndexes.Clear()
}

// GetArpIfIndexes exposes list of proxy ARP interface indexes
func (plugin *ProxyArpConfigurator) GetArpIfIndexes() idxvpp.NameToIdxRW {
	return plugin.pArpIfIndexes
}

// GetArpRngIndexes exposes list of proxy ARP range indexes
func (plugin *ProxyArpConfigurator) GetArpRngIndexes() idxvpp.NameToIdxRW {
	return plugin.pArpRngIndexes
}

// GetArpIfCache exposes list of cached ARP interfaces
func (plugin *ProxyArpConfigurator) GetArpIfCache() []string {
	return plugin.pArpIfCache
}

func (plugin *ProxyArpConfigurator) AddInterface(pArpIf *l3.ProxyArpInterfaces_InterfaceList) error {
	plugin.log.Infof("Enabling interfaces from proxy ARP config %s", pArpIf.Label)

	for _, proxyArpIf := range pArpIf.Interfaces {
		ifName := proxyArpIf.Name
		if ifName == "" {
			err := fmt.Errorf("proxy ARP interface not set")
			plugin.log.Error(err)
			return err
		}
		// Check interface, cache if does not exist
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(ifName)
		if !found {
			plugin.log.Debugf("Interface %s does not exist, moving to cache", ifName)
			plugin.pArpIfCache = append(plugin.pArpIfCache, ifName)
			continue
		}

		// Call VPP API to enable interface for proxy ARP
		if err := vppcalls.EnableProxyArpInterface(ifIdx, plugin.vppChan, plugin.log, plugin.stopwatch); err == nil {
			plugin.log.Debugf("Interface %s enabled for proxy ARP", ifName)
		} else {
			err := fmt.Errorf("enabling interface %s for proxy ARP failed: %v", ifName, err)
			plugin.log.Error(err)
			return err
		}
	}
	// Register
	plugin.pArpIfIndexes.RegisterName(pArpIf.Label, plugin.pArpIndexSeq, nil)
	plugin.pArpIndexSeq++
	plugin.log.Debugf("Proxy ARP interface configuration %s registered", pArpIf.Label)

	return nil
}

// ModifyInterface does nothing
func (plugin *ProxyArpConfigurator) ModifyInterface(newPArpIf, oldPArpIf *l3.ProxyArpInterfaces_InterfaceList) error {
	plugin.log.Infof("Modifying proxy ARP interface configuration %s", newPArpIf.Label)

	toEnable, toDisable := plugin.calculateIfDiff(newPArpIf.Interfaces, oldPArpIf.Interfaces)
	// Disable obsolete interfaces
	for _, ifName := range toDisable {
		// Check cache
		for idx, cachedIf := range plugin.pArpIfCache {
			if cachedIf == ifName {
				plugin.pArpIfCache = append(plugin.pArpIfCache[:idx], plugin.pArpIfCache[idx+1:]...)
				plugin.log.Debugf("Proxy ARP interface %s removed from cache", ifName)
				continue
			}
		}
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(ifName)
		// If interface is not found, there is nothing to do
		if found {
			if err := vppcalls.DisableProxyArpInterface(ifIdx, plugin.vppChan, plugin.log, plugin.stopwatch); err == nil {
				plugin.log.Debugf("Interface %s disabled for proxy ARP", ifName)
			} else {
				err = fmt.Errorf("disabling interface %s for proxy ARP failed: %v", ifName, err)
				plugin.log.Error(err)
				return err
			}
		}
	}
	// Enable new interfaces
	for _, ifName := range toEnable {
		// Put to cache if interface is missing
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(ifName)
		if !found {
			plugin.log.Debugf("Interface %s does not exist, moving to cache", ifName)
			plugin.pArpIfCache = append(plugin.pArpIfCache, ifName)
			continue
		}
		// Configure
		if err := vppcalls.EnableProxyArpInterface(ifIdx, plugin.vppChan, plugin.log, plugin.stopwatch); err == nil {
			plugin.log.Debugf("Interface %s enabled for proxy ARP", ifName)
		} else {
			err := fmt.Errorf("enabling interface %s for proxy ARP failed: %v", ifName, err)
			plugin.log.Error(err)
			return err
		}
	}

	plugin.log.Debugf("Proxy ARP interface config %s modification done", newPArpIf.Label)
	return nil
}

// DeleteInterface disables proxy ARP interface or removes it from cache
func (plugin *ProxyArpConfigurator) DeleteInterface(pArpIf *l3.ProxyArpInterfaces_InterfaceList) error {
	plugin.log.Infof("Disabling interfaces from proxy ARP config %s", pArpIf.Label)

ProxyArpIfLoop:
	for _, proxyArpIf := range pArpIf.Interfaces {
		ifName := proxyArpIf.Name
		// Check if interface is cached
		for idx, cachedIf := range plugin.pArpIfCache {
			if cachedIf == ifName {
				plugin.pArpIfCache = append(plugin.pArpIfCache[:idx], plugin.pArpIfCache[idx+1:]...)
				plugin.log.Debugf("Proxy ARP interface %s removed from cache", ifName)
				continue ProxyArpIfLoop
			}
		}
		// Look for interface
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(ifName)
		if !found {
			// Interface does not exist, nothing more to do
			continue
		}
		// Call VPP API to disable interface for proxy ARP
		if err := vppcalls.DisableProxyArpInterface(ifIdx, plugin.vppChan, plugin.log, plugin.stopwatch); err == nil {
			plugin.log.Debugf("Interface %s disabled for proxy ARP", ifName)
		} else {
			err = fmt.Errorf("disabling interface %s for proxy ARP failed: %v", ifName, err)
			plugin.log.Error(err)
			return err
		}
	}

	// Un-register
	plugin.pArpIfIndexes.UnregisterName(pArpIf.Label)
	plugin.log.Debugf("Proxy ARP interface config %s un-registered", pArpIf.Label)
	return nil
}

// AddRange configures new IP range for proxy ARP
func (plugin *ProxyArpConfigurator) AddRange(pArpRng *l3.ProxyArpRanges_RangeList) error {
	plugin.log.Infof("Setting up proxy ARP IP range config %s", pArpRng.Label)

	for _, proxyArpRange := range pArpRng.Ranges {
		// Prune addresses
		firstIP, err := plugin.pruneIP(proxyArpRange.FirstIp)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
		lastIP, err := plugin.pruneIP(proxyArpRange.LastIp)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
		// Convert to byte representation
		bFirstIP := net.ParseIP(firstIP).To4()
		bLastIP := net.ParseIP(lastIP).To4()
		// Call VPP API to configure IP range for proxy ARP
		if err := vppcalls.AddProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.log, plugin.stopwatch); err == nil {
			plugin.log.Debugf("Address range %s - %s configured for proxy ARP", firstIP, lastIP)
		} else {
			err := fmt.Errorf("failed to configure proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
			plugin.log.Error(err)
			return err
		}
	}

	// Register
	plugin.pArpRngIndexes.RegisterName(pArpRng.Label, plugin.pArpIndexSeq, nil)
	plugin.pArpIndexSeq++
	plugin.log.Debugf("Proxy ARP range config %s registered", pArpRng.Label)
	return nil
}

// ModifyRange does nothing
func (plugin *ProxyArpConfigurator) ModifyRange(newPArpRng, oldPArpRng *l3.ProxyArpRanges_RangeList) error {
	plugin.log.Infof("Modifying proxy ARP range config %s", oldPArpRng.Label)

	toAdd, toDelete := plugin.calculateRngDiff(newPArpRng.Ranges, oldPArpRng.Ranges)
	// Remove old ranges
	for _, rng := range toDelete {
		// Prune
		firstIP, err := plugin.pruneIP(rng.FirstIp)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
		lastIP, err := plugin.pruneIP(rng.LastIp)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
		// Convert to byte representation
		bFirstIP := net.ParseIP(firstIP).To4()
		bLastIP := net.ParseIP(lastIP).To4()
		// Call VPP API to configure IP range for proxy ARP
		if err := vppcalls.DeleteProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.log, plugin.stopwatch); err == nil {
			plugin.log.Debugf("Address range %s - %s removed from proxy ARP setup", firstIP, lastIP)
		} else {
			err = fmt.Errorf("failed to remove proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
			plugin.log.Error(err)
			return err
		}
	}
	// Add new ranges
	for _, rng := range toAdd {
		// Prune addresses
		firstIP, err := plugin.pruneIP(rng.FirstIp)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
		lastIP, err := plugin.pruneIP(rng.LastIp)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
		// Convert to byte representation
		bFirstIP := net.ParseIP(firstIP).To4()
		bLastIP := net.ParseIP(lastIP).To4()
		// Call VPP API to configure IP range for proxy ARP
		if err := vppcalls.AddProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.log, plugin.stopwatch); err == nil {
			plugin.log.Debugf("Address range %s - %s configured for proxy ARP", firstIP, lastIP)
		} else {
			err := fmt.Errorf("failed to configure proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
			plugin.log.Error(err)
			return err
		}
	}

	plugin.log.Debugf("Proxy ARP range config %s modification done", newPArpRng.Label)
	return nil
}

func (plugin *ProxyArpConfigurator) DeleteRange(pArpRng *l3.ProxyArpRanges_RangeList) error {
	plugin.log.Infof("Removing proxy ARP IP range config %s", pArpRng.Label)

	for _, proxyArpRange := range pArpRng.Ranges {
		// Prune addresses
		firstIP, err := plugin.pruneIP(proxyArpRange.FirstIp)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
		lastIP, err := plugin.pruneIP(proxyArpRange.LastIp)
		if err != nil {
			plugin.log.Error(err)
			return err
		}
		// Convert to byte representation
		bFirstIP := net.ParseIP(firstIP).To4()
		bLastIP := net.ParseIP(lastIP).To4()
		// Call VPP API to configure IP range for proxy ARP
		if err := vppcalls.DeleteProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.log, plugin.stopwatch); err == nil {
			plugin.log.Debugf("Address range %s - %s removed from proxy ARP setup", firstIP, lastIP)
		} else {
			err = fmt.Errorf("failed to remove proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
			plugin.log.Error(err)
			return err
		}
	}

	// Un-register
	plugin.pArpRngIndexes.UnregisterName(pArpRng.Label)
	plugin.log.Debugf("Proxy ARP range config %s un-registered", pArpRng.Label)
	return nil
}

// ResolveCreatedInterface handles new registered interface for proxy ARP
func (plugin *ProxyArpConfigurator) ResolveCreatedInterface(ifName string, ifIdx uint32) error {
	plugin.log.Debugf("Proxy ARP: handling new interface %s", ifName)

	// Look for interface in cache
	for idx, cachedIf := range plugin.pArpIfCache {
		if cachedIf == ifName {
			// Configure cached interface
			if err := vppcalls.EnableProxyArpInterface(ifIdx, plugin.vppChan, plugin.log, plugin.stopwatch); err != nil {
				plugin.log.Error(err)
				return err
			}
			// Remove from cache
			plugin.pArpIfCache = append(plugin.pArpIfCache[:idx], plugin.pArpIfCache[idx+1:]...)
			plugin.log.Debugf("Proxy ARP interface %s configured and removed from cache", ifName)
			return nil
		}
	}

	return nil
}

// ResolveDeletedInterface handles new registered interface for proxy ARP
func (plugin *ProxyArpConfigurator) ResolveDeletedInterface(ifName string) error {
	plugin.log.Debugf("Proxy ARP: handling removed interface %s", ifName)

	// Check if interface was enabled for proxy ARP
	_, _, found := plugin.pArpIfIndexes.LookupIdx(ifName)
	if found {
		// Put interface to cache (no need to call delete)
		plugin.pArpIfCache = append(plugin.pArpIfCache, ifName)
		plugin.log.Debugf("Removed interface %s was configured for proxy ARP, added to cache", ifName)
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
		plugin.log.Warnf("Proxy ARP range: removing unnecessary mask from IP address %s", ip)
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
