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
	plugin.Log.Debug("Initializing proxy ARP configurator")

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

func (plugin *ProxyArpConfigurator) AddInterface(pArpIf *l3.ProxyArpInterfaces_Interface) error {
	plugin.Log.Infof("Enabling interface %s for proxy ARP", pArpIf.Interface)

	if pArpIf.Interface == "" {
		return fmt.Errorf("proxy ARP interface not set")
	}

	// Check interface, cache if does not exist
	ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(pArpIf.Interface)
	if !found {
		plugin.Log.Debugf("Interface %s does not exist, moving to cache", pArpIf.Interface)
		plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache, pArpIf.Interface)
		return nil
	}

	// Call VPP API to enable interface for proxy ARP
	if err := vppcalls.EnableProxyArpInterface(ifIdx, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
		plugin.Log.Debugf("Interface %s enabled for proxy ARP", pArpIf.Interface)
	} else {
		return fmt.Errorf("enabling interface %s for proxy ARP failed: %v", pArpIf.Interface, err)
	}

	// Register
	plugin.ProxyArpIfIndices.RegisterName(pArpIf.Interface, plugin.ProxyARPIndexSeq, nil)
	plugin.ProxyARPIndexSeq++
	plugin.Log.Debugf("Proxy ARP interface %s registered", pArpIf.Interface)

	return nil
}

// ModifyInterface does nothing
func (plugin *ProxyArpConfigurator) ModifyInterface(newPArpIf, oldPArpIf *l3.ProxyArpInterfaces_Interface) error {
	plugin.Log.Info("There is nothing to modify for proxy ARP interface %s", oldPArpIf.Interface)
	return nil
}

// DeleteInterface disables proxy ARP interface or removes it from cache
func (plugin *ProxyArpConfigurator) DeleteInterface(pArpIf *l3.ProxyArpInterfaces_Interface) error {
	plugin.Log.Info("Disabling interface %s for proxy ARP", pArpIf.Interface)

	// Check if interface is cached
	for idx, cachedIf := range plugin.ProxyARPIfCache {
		if cachedIf == pArpIf.Interface {
			plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache[:idx], plugin.ProxyARPIfCache[idx+1:]...)
			plugin.Log.Debugf("Proxy ARP interface %s removed from cache", pArpIf.Interface)
			return nil
		}
	}

	// Look for interface
	ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(pArpIf.Interface)
	if !found {
		// Interface does not exist, nothing more to do but un-register
		plugin.ProxyArpIfIndices.UnregisterName(pArpIf.Interface)
		plugin.Log.Debugf("Proxy ARP interface %s un-registered", pArpIf.Interface)
		return nil
	}

	// Call VPP API to disable interface for proxy ARP
	if err := vppcalls.DisableProxyArpInterface(ifIdx, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
		plugin.Log.Debugf("Interface %s disabled for proxy ARP", pArpIf.Interface)
	} else {
		return fmt.Errorf("disabling interface %s for proxy ARP failed: %v", pArpIf.Interface, err)
	}

	// Un-register
	plugin.ProxyArpIfIndices.UnregisterName(pArpIf.Interface)
	plugin.Log.Debugf("Proxy ARP interface %s un-registered", pArpIf.Interface)

	return nil
}

// AddRange configures new IP range for proxy ARP
func (plugin *ProxyArpConfigurator) AddRange(pArpRng *l3.ProxyArpRanges_Range) error {
	plugin.Log.Infof("Setting up proxy ARP IP range %s - %s", pArpRng.FirstIp, pArpRng.LastIp)

	// Prune addresses
	firstIP, err := plugin.pruneIP(pArpRng.FirstIp)
	if err != nil {
		return err
	}
	lastIP, err := plugin.pruneIP(pArpRng.LastIp)
	if err != nil {
		return err
	}

	// Convert to byte representation
	bFirstIP := net.ParseIP(firstIP).To4()
	bLastIP := net.ParseIP(lastIP).To4()

	// Call VPP API to configure IP range for proxy ARP
	if err := vppcalls.AddProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
		plugin.Log.Debugf("Address range %s - %s configured for proxy ARP", firstIP, lastIP)
	} else {
		return fmt.Errorf("failed to configure proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
	}

	// Register
	id := rangeIdentifier(firstIP, lastIP)
	plugin.ProxyArpRngIndices.RegisterName(id, plugin.ProxyARPIndexSeq, nil)
	plugin.ProxyARPIndexSeq++
	plugin.Log.Debugf("Proxy ARP range %s registered", id)

	return nil
}

// ModifyRange does nothing
func (plugin *ProxyArpConfigurator) ModifyRange(newPArpRng, oldPArpRng *l3.ProxyArpRanges_Range) error {
	plugin.Log.Info("There is nothing to modify for proxy ARP range %s", oldPArpRng.FirstIp, oldPArpRng.LastIp)
	return nil
}

func (plugin *ProxyArpConfigurator) DeleteRange(pArpRng *l3.ProxyArpRanges_Range) error {
	plugin.Log.Infof("Removing proxy ARP IP range %s - %s", pArpRng.FirstIp, pArpRng.LastIp)

	// Prune addresses
	firstIP, err := plugin.pruneIP(pArpRng.FirstIp)
	if err != nil {
		return err
	}
	lastIP, err := plugin.pruneIP(pArpRng.LastIp)
	if err != nil {
		return err
	}

	// Convert to byte representation
	bFirstIP := net.ParseIP(firstIP).To4()
	bLastIP := net.ParseIP(lastIP).To4()

	// Call VPP API to configure IP range for proxy ARP
	if err := vppcalls.DeleteProxyArpRange(bFirstIP, bLastIP, plugin.vppChan, plugin.Log, plugin.Stopwatch); err == nil {
		plugin.Log.Debugf("Address range %s - %s removed from proxy ARP setup", firstIP, lastIP)
	} else {
		return fmt.Errorf("failed to remove proxy ARP address range %s - %s: %v", firstIP, lastIP, err)
	}

	// Un-register
	id := rangeIdentifier(firstIP, lastIP)
	plugin.ProxyArpIfIndices.UnregisterName(id)
	plugin.Log.Debugf("Proxy ARP range %s - %s un-registered", firstIP, lastIP)

	return nil
}

// ResolveCreatedInterface handles new registered interface for proxy ARP
func (plugin *ProxyArpConfigurator) ResolveCreatedInterface(ifName string) error {
	plugin.Log.Debugf("Proxy ARP: handling new interface %s", ifName)

	// Look for interface in cache
	var wasErr error
	for idx, cachedIf := range plugin.ProxyARPIfCache {
		if cachedIf == ifName {
			// Configure cached interface
			if err := plugin.AddInterface(&l3.ProxyArpInterfaces_Interface{
				Interface: ifName,
			}); err != nil {
				plugin.Log.Error(err)
				wasErr = err
			}
			// Remove from cache
			plugin.ProxyARPIfCache = append(plugin.ProxyARPIfCache[:idx], plugin.ProxyARPIfCache[idx+1:]...)
			plugin.Log.Debugf("Proxy ARP interface %s configured and removed from cache", ifName)
			return nil
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
		// If so, un-register it and add it to cache (no need to call delete here)
		plugin.ProxyArpIfIndices.UnregisterName(ifName)
		// Put interface to cache
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

// Generate internal range identifier (IP addresses without dots)
func rangeIdentifier(ip1, ip2 string) string {
	return strings.Trim(ip1+ip2, ".")
}
