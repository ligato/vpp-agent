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

package ifplugin

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/nsplugin"
	"github.com/vishvananda/netlink"
)

// A List of known linux interface types which can be processed.
const (
	tap  = "tun"
	veth = "veth"
)

// LinuxDataPair stores linux interface with matching NB configuration
type LinuxDataPair struct {
	linuxIfData netlink.Link
	nbIfData    *interfaces.LinuxInterfaces_Interface
}

// Resync writes interfaces to Linux. Interface host name corresponds with Linux host interface name (but name can
// be different). Resync consists of following steps:
// 1. Iterate over all NB interfaces. Try to find interface with the same name in required namespace for every NB interface.
// 2. If interface does not exist, will be created anew
// 3. If interface exists, it is correlated and modified if needed.
// Resync configures an initial set of interfaces. Existing Linux interfaces are registered and potentially re-configured.
func (plugin *LinuxInterfaceConfigurator) Resync(nbIfs []*interfaces.LinuxInterfaces_Interface) (errs []error) {
	plugin.Log.Debugf("RESYNC Linux interface begin.")

	start := time.Now()
	defer func() {
		if plugin.Stopwatch != nil {
			timeLog := measure.GetTimeLog("linux-interface resync", plugin.Stopwatch)
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()

	// Cache for interfaces modified later (interface name/link data)
	linkMap := make(map[string]*LinuxDataPair)

	// Iterate over NB configuration. Look for interfaces with the same host name
	for _, nbIf := range nbIfs {
		plugin.handleOptionalHostIfName(nbIf)
		plugin.addInterfaceToCache(nbIf)

		// Find linux equivalent for every NB interface and register it
		linkIf, err := plugin.findLinuxInterface(nbIf, nsMgmtCtx)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if linkIf != nil {
			// If interface was found, it will be compared and modified in the next step
			plugin.Log.Debugf("RESYNC Linux interface %v: interface found in namespace", nbIf.Name)
			linkMap[nbIf.Name] = &LinuxDataPair{
				linuxIfData: linkIf,
				nbIfData:    nbIf,
			}
		} else {
			// If not, configure it
			plugin.Log.Debugf("RESYNC Linux interface %v: interface not found and will be configured", nbIf.Name)
			if err := plugin.ConfigureLinuxInterface(nbIf); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// Process all interfaces waiting for modification. All NB interfaces are already registered at this point.
	for linkName, linkDataPair := range linkMap {
		linuxIf := plugin.reconstructIfConfig(linkDataPair.linuxIfData, linkDataPair.nbIfData.Namespace, linkName)

		// For VETH, resolve peer
		if linkDataPair.nbIfData.Type == interfaces.LinuxInterfaces_VETH {
			// Search registered config for peer
			var found bool
			for _, cachedIfCfg := range plugin.intfByName {
				if cachedIfCfg.config != nil && cachedIfCfg.config.Type == interfaces.LinuxInterfaces_VETH {
					if cachedIfCfg.config.Veth != nil && cachedIfCfg.config.Veth.PeerIfName == linuxIf.HostIfName {
						found = true
						linuxIf.Veth = &interfaces.LinuxInterfaces_Interface_Veth{
							PeerIfName: cachedIfCfg.config.HostIfName,
						}
					}
				}
			}
			if found {
				plugin.Log.Debugf("RESYNC Linux interface %v: found peer %v", linkName, linuxIf.Veth.PeerIfName)
			} else {
				err := fmt.Errorf("RESYNC Linux interface %v: failed to obtain peer for veth", linkName)
				errs = append(errs, err)
			}
		}
		// Check if interface needs to be modified
		if plugin.isLinuxIfModified(linkDataPair.nbIfData, linuxIf) {
			plugin.Log.Debugf("RESYNC Linux interface %s: configuration changed, interface will be modified", linkName)
			if err := plugin.ModifyLinuxInterface(linkDataPair.nbIfData, linuxIf); err != nil {
				errs = append(errs, err)
			}
		} else {
			plugin.Log.Debugf("RESYNC Linux interface %s: data was not changed", linkName)
		}
	}

	// Register all interfaces in default namespace which were not already registered
	linkList, err := netlink.LinkList()
	if err != nil {
		plugin.Log.Errorf("Failed to read linux interfaces: %v", err)
		errs = append(errs, err)
	}
	for _, linuxIf := range linkList {
		if linuxIf.Attrs() == nil {
			continue
		}
		attrs := linuxIf.Attrs()
		_, _, found := plugin.IfIndexes.LookupIdx(attrs.Name)
		if !found {
			// Register interface with name (other parameters can be read if needed)
			plugin.IfIndexes.RegisterName(attrs.Name, plugin.IfIdxSeq, &ifaceidx.IndexedLinuxInterface{
				Index: uint32(attrs.Index),
				Data: &interfaces.LinuxInterfaces_Interface{
					Name:       attrs.Name,
					HostIfName: attrs.Name,
				},
			})
			plugin.IfIdxSeq++
		}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface end. ", errs)

	return
}

// Reconstruct common interface configuration from netlink.Link data
func (plugin *LinuxInterfaceConfigurator) reconstructIfConfig(linuxIf netlink.Link, ns *interfaces.LinuxInterfaces_Interface_Namespace, name string) *interfaces.LinuxInterfaces_Interface {
	linuxIfAttr := linuxIf.Attrs()
	return &interfaces.LinuxInterfaces_Interface{
		Name: name,
		Type: func(ifType string) interfaces.LinuxInterfaces_InterfaceType {
			if ifType == veth {
				return interfaces.LinuxInterfaces_VETH
			}
			return interfaces.LinuxInterfaces_AUTO_TAP
		}(linuxIf.Type()),
		Enabled: func(state netlink.LinkOperState) bool {
			if state == netlink.OperDown {
				return false
			}
			return true
		}(linuxIfAttr.OperState),
		IpAddresses: plugin.getLinuxInterfaces(linuxIf, ns),
		PhysAddress: func(hwAddr net.HardwareAddr) (mac string) {
			if hwAddr != nil {
				mac = hwAddr.String()
			}
			return
		}(linuxIfAttr.HardwareAddr),
		Mtu:        uint32(linuxIfAttr.MTU),
		HostIfName: linuxIfAttr.Name,
	}
}

// Reads linux interface IP addresses
func (plugin *LinuxInterfaceConfigurator) getLinuxInterfaces(linuxIf netlink.Link, ns *interfaces.LinuxInterfaces_Interface_Namespace) (addresses []string) {
	// Move to proper namespace
	if ns != nil {
		if !plugin.NsHandler.IsNamespaceAvailable(ns) {
			plugin.Log.Errorf("RESYNC Linux interface %s: namespace is not available", linuxIf.Attrs().Name)
			return
		}
		// Switch to namespace
		revertNs, err := plugin.NsHandler.SwitchToNamespace(nsplugin.NewNamespaceMgmtCtx(), ns)
		if err != nil {
			plugin.Log.Errorf("RESYNC Linux interface %s: failed to switch to namespace %s: %v",
				linuxIf.Attrs().Name, ns.Name, err)
			return
		}
		defer revertNs()
	}

	addressList, err := netlink.AddrList(linuxIf, netlink.FAMILY_ALL)
	if err != nil {
		plugin.Log.Errorf("failed to read linux interface %s address list: %v", linuxIf.Attrs().Name, err)
		return
	}

	for _, address := range addressList {
		mask, _ := address.Mask.Size()
		addrStr := address.IP.String() + "/" + strconv.Itoa(mask)
		addresses = append(addresses, addrStr)
	}

	return addresses
}

// Compare interface fields in order to find differences.
func (plugin *LinuxInterfaceConfigurator) isLinuxIfModified(nbIf, linuxIf *interfaces.LinuxInterfaces_Interface) bool {
	plugin.Log.Debugf("Linux interface RESYNC comparison started for interface %s", nbIf.Name)

	// Type
	if nbIf.Type != linuxIf.Type {
		plugin.Log.Debugf("Linux interface RESYNC comparison: type changed (NB: %v, Linux: %v)",
			nbIf.Type, linuxIf.Type)
		return true
	}
	// Enabled
	if nbIf.Enabled != linuxIf.Enabled {
		plugin.Log.Debugf("Linux interface RESYNC comparison: enabled value changed (NB: %t, Linux: %t)",
			nbIf.Enabled, linuxIf.Enabled)
		return true
	}
	// Remove IPv6 link local addresses (default values)
	for ipIdx, ipAddress := range linuxIf.IpAddresses {
		if strings.HasPrefix(ipAddress, "fe80") {
			linuxIf.IpAddresses = append(linuxIf.IpAddresses[:ipIdx], linuxIf.IpAddresses[ipIdx+1:]...)
		}
	}
	// IP address count
	if len(nbIf.IpAddresses) != len(linuxIf.IpAddresses) {
		plugin.Log.Debugf("Linux interface RESYNC comparison: ip address count does not match (NB: %d, Linux: %d)",
			len(nbIf.IpAddresses), len(linuxIf.IpAddresses))
		return true
	}
	// IP address comparison
	for _, nbIP := range nbIf.IpAddresses {
		var ipFound bool
		for _, linuxIP := range linuxIf.IpAddresses {
			pNbIP, nbIPNet, err := net.ParseCIDR(nbIP)
			if err != nil {
				plugin.Log.Error(err)
				continue
			}
			pVppIP, vppIPNet, err := net.ParseCIDR(linuxIP)
			if err != nil {
				plugin.Log.Error(err)
				continue
			}
			if nbIPNet.Mask.String() == vppIPNet.Mask.String() && bytes.Compare(pNbIP, pVppIP) == 0 {
				ipFound = true
				break
			}
		}
		if !ipFound {
			plugin.Log.Debugf("Interface RESYNC comparison: linux interface %v does not contain IP %s", nbIf.Name, nbIP)
			return true
		}
	}
	// Physical address
	if nbIf.PhysAddress != linuxIf.PhysAddress {
		plugin.Log.Debugf("Interface RESYNC comparison: MAC address changed (NB: %s, Linux: %s)",
			nbIf.PhysAddress, linuxIf.PhysAddress)
		return true
	}
	// MTU (if NB value is set)
	if nbIf.Mtu != 0 && nbIf.Mtu != linuxIf.Mtu {
		plugin.Log.Debugf("Interface RESYNC comparison: MTU changed (NB: %d, Linux: %d)",
			nbIf.Mtu, linuxIf.Mtu)
		return true
	}
	switch nbIf.Type {
	case interfaces.LinuxInterfaces_VETH:
		if nbIf.Veth == nil && linuxIf.Veth != nil || nbIf.Veth != nil && linuxIf.Veth == nil {
			plugin.Log.Debugf("Interface RESYNC comparison: VETH setup changed (NB: %v, VPP: %v)",
				nbIf.Veth, linuxIf.Veth)
			return true
		}
		if nbIf.Veth != nil && linuxIf.Veth != nil {
			// VETH peer name
			if nbIf.Veth.PeerIfName != linuxIf.Veth.PeerIfName {
				plugin.Log.Debugf("Interface RESYNC comparison: VETH peer name changed (NB: %s, VPP: %s)",
					nbIf.Veth.PeerIfName, linuxIf.Veth.PeerIfName)
				return true
			}
		}
	case interfaces.LinuxInterfaces_AUTO_TAP:
		// Host name for TAP
		if nbIf.HostIfName != linuxIf.HostIfName {
			plugin.Log.Debugf("Interface RESYNC comparison: TAP host name changed (NB: %d, Linux: %d)",
				nbIf.HostIfName, linuxIf.HostIfName)
			return true
		}
		// Note: do not compare TAP temporary name. It is local-only parameter which cannot be leveraged externally.
	}

	return false
}

// Looks for linux interface. Returns net.Link object if found
func (plugin *LinuxInterfaceConfigurator) findLinuxInterface(nbIf *interfaces.LinuxInterfaces_Interface, nsMgmtCtx *nsplugin.NamespaceMgmtCtx) (netlink.Link, error) {
	plugin.Log.Debugf("Looking for Linux interface %v", nbIf.HostIfName)

	// Move to proper namespace
	if nbIf.Namespace != nil {
		if !plugin.NsHandler.IsNamespaceAvailable(nbIf.Namespace) {
			// Not and error
			plugin.Log.Debugf("Interface %s is not ready to be configured, namespace %s is not available",
				nbIf.Name, nbIf.Namespace.Name)
			return nil, nil
		}
		// Switch to namespace
		revertNs, err := plugin.NsHandler.SwitchToNamespace(nsMgmtCtx, nbIf.Namespace)
		if err != nil {
			return nil, fmt.Errorf("RESYNC Linux interface %s: failed to switch to namespace %s: %v",
				nbIf.HostIfName, nbIf.Namespace.Name, err)
		}
		defer revertNs()
	}
	// Look for interface
	linkIf, err := netlink.LinkByName(nbIf.HostIfName)
	if err != nil {
		// Link not found is not an error in this case
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			// Interface was not found
			return nil, nil
		} else {
			return nil, fmt.Errorf("RESYNC Linux interface %s: %v", nbIf.HostIfName, err)
		}
	}
	if linkIf == nil || linkIf.Attrs() == nil {
		return nil, fmt.Errorf("RESYNC Linux interface %v: link is nil", nbIf.HostIfName)
	}

	// Add interface to cache
	plugin.registerLinuxInterface(uint32(linkIf.Attrs().Index), nbIf)

	return linkIf, nil
}

// Register linux interface
func (plugin *LinuxInterfaceConfigurator) registerLinuxInterface(linuxIfIdx uint32, nbIf *interfaces.LinuxInterfaces_Interface) {
	// Register interface with its name
	plugin.IfIndexes.RegisterName(nbIf.Name, plugin.IfIdxSeq, &ifaceidx.IndexedLinuxInterface{
		Index: linuxIfIdx,
		Data:  nbIf,
	})
	plugin.IfIdxSeq++
}

// Add interface to cache
func (plugin *LinuxInterfaceConfigurator) addInterfaceToCache(nbIf *interfaces.LinuxInterfaces_Interface) *LinuxInterfaceConfig {
	switch nbIf.Type {
	case interfaces.LinuxInterfaces_AUTO_TAP:
		return plugin.addToCache(nbIf, nil)
	case interfaces.LinuxInterfaces_VETH:
		peerConfig := plugin.getInterfaceConfig(nbIf.Veth.PeerIfName)
		return plugin.addToCache(nbIf, peerConfig)
	}
	return nil
}
