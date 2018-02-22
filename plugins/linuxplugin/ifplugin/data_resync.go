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
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/linuxcalls"
	"github.com/vishvananda/netlink"
)

// A List of known linux interface types which can be processed.
const (
	tap             = "tun"
	veth            = "veth"
	linkNotFoundErr = "Link not found"
)

// Resync writes interfaces to Linux. Interface host name corresponds with Linux host interface name (but name can
// be different). Resync consists of following steps:
// 1. Iterate over all NB interfaces. Try to find interface with the same name in required namespace for every NB interface.
// 2. If interface does not exist, will be created anew
// 3. If interface exists, it is correlated and modified if needed.
// Resync configures an initial set of interfaces. Existing Linux interfaces are registered and potentially re-configured.
func (plugin *LinuxInterfaceConfigurator) Resync(nbIfs []*interfaces.LinuxInterfaces_Interface) (errs []error) {
	plugin.Log.Warn("RESYNC Linux interface begin.")

	start := time.Now()
	defer func() {
		if plugin.Stopwatch != nil {
			timeLog := measure.GetTimeLog("linux-interface resync", plugin.Stopwatch)
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	nsMgmtCtx := linuxcalls.NewNamespaceMgmtCtx()

	// Iterate over NB configuration. Look for similar interfaces in the namespace
	for _, nbIf := range nbIfs {
		plugin.handleOptionalHostIfName(nbIf)

		switch nbIf.Type {
		case interfaces.LinuxInterfaces_AUTO_TAP:
			recreate, err := plugin.handleTapResync(nbIf, nsMgmtCtx)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if recreate {
				// If attempt to handle interface was not successful, create is as a new one
				if err := plugin.ConfigureLinuxInterface(nbIf); err != nil {
					errs = append(errs, err)
				}
			}
		case interfaces.LinuxInterfaces_VETH:
			// todo
		}

	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface end. ", errs)

	return
}

// Resolve linux TAP interface type resync.
func (plugin *LinuxInterfaceConfigurator) handleTapResync(nbTap *interfaces.LinuxInterfaces_Interface, nsMgmtCtx *linuxcalls.NamespaceMgmtCtx) (bool, error) {
	plugin.Log.Debugf("RESYNC of TAP linux interface: %v", nbTap.HostIfName)

	var err error
	var linkIf netlink.Link

	// Look for interface in default namespace. Use first temporary name, then host name (temporary name is optional)
	if nbTap.Tap == nil {
		nbTap.Tap = &interfaces.LinuxInterfaces_Interface_Tap{
			TempIfName: nbTap.HostIfName,
		}
	}
	linkIf, err = netlink.LinkByName(nbTap.Tap.TempIfName)
	if err != nil {
		if err.Error() != linkNotFoundErr { // Link not found is not an error in this case
			return false, fmt.Errorf("RESYNC Linux interface: failed to read interface %s", nbTap.HostIfName)
		}
		// Look for interface using host name
		linkIf, err = netlink.LinkByName(nbTap.HostIfName)
		if err != nil {
			if err.Error() != linkNotFoundErr { // Link not found is not an error in this case
				return false, fmt.Errorf("RESYNC Linux interface: failed to read interface %s", nbTap.HostIfName)
			}
		}
	}
	// If link exists, process it
	if linkIf != nil {
		// Reconstruct and register linux interface
		linuxIf := plugin.reconstructIfConfig(linkIf, nbTap.Name)
		plugin.registerLinuxInterface(uint32(linkIf.Attrs().Index), nbTap)
		// Calculate whether modification is needed
		if plugin.isLinuxIfModified(nbTap, linuxIf) {
			plugin.Log.Debugf("RESYNC linux interfaces: modifying interface %v", nbTap.Name)
			if err = plugin.ModifyLinuxInterface(nbTap, linuxIf); err != nil {
				return false, fmt.Errorf("error while modifying linux interface: %v", err)
			}
		} else {
			plugin.Log.Debugf("RESYNC linux interfaces: %v registered without additional changes", nbTap.Name)
		}
		return false, nil
	}

	// If interface was not found in default namespace, it can be already configured in its namespace
	var revertNs func()
	// Move to appropriate namespace
	//TODO: interface could be moved to different namespace so all of them should be searched
	if nbTap.Namespace != nil {
		if nbTap.Namespace != nil {
			if !plugin.isNamespaceAvailable(nbTap.Namespace) {
				return false, fmt.Errorf("RESYNC Linux interface: cannot correlate %s, namespace is not available",
					nbTap.HostIfName)
			}
			// Switch to namespace
			var err error
			revertNs, err = plugin.switchToNamespace(nsMgmtCtx, nbTap.Namespace)
			if err != nil {
				return false, fmt.Errorf("RESYNC Linux interface: cannot correlate %s, failed to switch to namespace %s",
					nbTap.HostIfName, nbTap.Namespace.Name)
			}
		}
	} else {
		// Without namespace to look in, interface cannot be found
		return true, nil
	}

	// Define defer func to revert namespace if needed
	defer func(revertNs func()) {
		if revertNs != nil {
			revertNs()
		}
	}(revertNs)

	// Look for interface in it's namespace using host name
	linkIf, err = netlink.LinkByName(nbTap.HostIfName)
	if err != nil {
		// Link not found is not an error in this case
		if err.Error() != "Link not found" {
			return false, fmt.Errorf("RESYNC Linux interface: failed to read interface %s", nbTap.HostIfName)
		}
	}

	if linkIf != nil {
		// Reconstruct and register linux interface
		linuxIf := plugin.reconstructIfConfig(linkIf, nbTap.Name)
		plugin.registerLinuxInterface(uint32(linkIf.Attrs().Index), nbTap)
		// Calculate whether modification is needed
		if plugin.isLinuxIfModified(nbTap, linuxIf) {
			plugin.Log.Debugf("RESYNC linux interfaces: modifying interface %v", nbTap.Name)
			if err = plugin.ModifyLinuxInterface(nbTap, linuxIf); err != nil {
				return false, fmt.Errorf("error while modifying linux interface: %v", err)

			}
		} else {
			plugin.Log.Debugf("RESYNC linux interfaces: %v registered without additional changes", nbTap.Name)
		}
		return false, nil
	} else {
		// Interface was not found
		return true, nil
	}
}

// Register linux interface and add it to cache
func (plugin *LinuxInterfaceConfigurator) registerLinuxInterface(linuxIfIdx uint32, nbIf *interfaces.LinuxInterfaces_Interface) *LinuxInterfaceConfig {
	// Register interface with its name
	plugin.ifIndexes.RegisterName(nbIf.Name, linuxIfIdx, nbIf)
	// Add interface to cache
	switch nbIf.Type {
	case interfaces.LinuxInterfaces_AUTO_TAP:
		return plugin.addToCache(nbIf, nil)
	case interfaces.LinuxInterfaces_VETH:
		peerConfig := plugin.getInterfaceConfig(nbIf.Veth.PeerIfName)
		return plugin.addToCache(nbIf, peerConfig)
	}
	return nil
}

// Reconstruct interface configuration from netlink.Link data
func (plugin *LinuxInterfaceConfigurator) reconstructIfConfig(linuxIf netlink.Link, name string) *interfaces.LinuxInterfaces_Interface {
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
		IpAddresses: func(linuxIf netlink.Link) []string {
			addressList, err := netlink.AddrList(linuxIf, netlink.FAMILY_ALL)
			if err != nil {
				plugin.Log.Errorf("Failed to read linux interface %s address list: %v", linuxIf.Attrs().Name, err)
				return nil
			}
			var addresses []string
			for _, address := range addressList {
				mask, _ := address.Mask.Size()
				addrStr := address.IP.String() + "/" + strconv.Itoa(mask)
				addresses = append(addresses, addrStr)
			}
			return addresses
		}(linuxIf),
		PhysAddress: func(hwAddr net.HardwareAddr) (mac string) {
			if hwAddr != nil {
				mac = hwAddr.String()
			}
			return
		}(linuxIfAttr.HardwareAddr),
		Mtu:        uint32(linuxIfAttr.MTU),
		HostIfName: linuxIfAttr.Name,
		Veth: func(linuxIf netlink.Link) (veth *interfaces.LinuxInterfaces_Interface_Veth) {
			vethIf, ok := linuxIf.(*netlink.Veth)
			if ok && vethIf != nil {
				//index, _ := netlink.VethPeerIndex(vethIf)
				// todo try to find index according to nb
			} else {
				plugin.Log.Warnf("Veth %v has no peer interface", linuxIf.Attrs().Name)
			}
			return
		}(linuxIf),
	}
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
	// MTU
	if nbIf.Mtu != linuxIf.Mtu {
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
