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

//go:generate protoc --proto_path=../model/l3 --gogo_out=../model/l3 ../model/l3/l3.proto

package l3plugin

import (
	"fmt"
	"net"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linux/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linux/model/l3"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/vishvananda/netlink"
)

const (
	noIfaceIdxFilter = 0
	noFamilyFilter   = 0
)

// LinuxArpConfigurator watches for any changes in the configuration of static ARPs as modelled by the proto file
// "model/l3/l3.proto" and stored in ETCD under the key "/vnf-agent/{vnf-agent}/linux/config/v1/arp".
// Updates received from the northbound API are compared with the Linux network configuration and differences
// are applied through the Netlink AP
type LinuxArpConfigurator struct {
	log logging.Logger

	// Mappings
	ifIndexes  ifaceidx.LinuxIfIndexRW
	arpIndexes l3idx.LinuxARPIndexRW
	arpIfCache map[string]*ArpToInterface // Cache for non-configurable ARPs due to missing interface
	arpIdxSeq  uint32

	// Linux namespace/calls handler
	l3Handler linuxcalls.NetlinkAPI
	nsHandler nsplugin.NamespaceAPI

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// ArpToInterface is an object which stores ARP-to-interface pairs used in cache.
// Field 'isAdd' marks whether the entry should be added or removed
type ArpToInterface struct {
	arp    *l3.LinuxStaticArpEntries_ArpEntry
	ifName string
	isAdd  bool
}

// GetArpIndexes returns arp in-memory indexes
func (plugin *LinuxArpConfigurator) GetArpIndexes() l3idx.LinuxARPIndexRW {
	return plugin.arpIndexes
}

// GetArpInterfaceCache returns internal non-configurable interface cache, mainly for testing purpose
func (plugin *LinuxArpConfigurator) GetArpInterfaceCache() map[string]*ArpToInterface {
	return plugin.arpIfCache
}

// Init initializes ARP configurator and starts goroutines
func (plugin *LinuxArpConfigurator) Init(logger logging.PluginLogger, l3Handler linuxcalls.NetlinkAPI, nsHandler nsplugin.NamespaceAPI,
	ifIndexes ifaceidx.LinuxIfIndexRW, stopwatch *measure.Stopwatch) error {
	// Logger
	plugin.log = logger.NewLogger("-arp-conf")
	plugin.log.Debug("Initializing Linux ARP configurator")

	// In-memory mappings
	plugin.ifIndexes = ifIndexes
	plugin.arpIndexes = l3idx.NewLinuxARPIndex(nametoidx.NewNameToIdx(plugin.log, "linux_arp_indexes", nil))
	plugin.arpIfCache = make(map[string]*ArpToInterface)
	plugin.arpIdxSeq = 1

	// L3 and namespace handler
	plugin.l3Handler = l3Handler
	plugin.nsHandler = nsHandler

	// Configurator-wide stopwatch instance
	plugin.stopwatch = stopwatch

	return nil
}

// Close closes all goroutines started during Init
func (plugin *LinuxArpConfigurator) Close() error {
	return nil
}

// ConfigureLinuxStaticArpEntry reacts to a new northbound Linux ARP entry config by creating and configuring
// the entry in the host network stack through Netlink API.
func (plugin *LinuxArpConfigurator) ConfigureLinuxStaticArpEntry(arpEntry *l3.LinuxStaticArpEntries_ArpEntry) error {
	plugin.log.Infof("Configuring Linux ARP entry %v", arpEntry.Name)
	var err error

	// Prepare ARP entry object
	neigh := &netlink.Neigh{}

	// Find interface
	_, ifData, found := plugin.ifIndexes.LookupIdx(arpEntry.Interface)
	if !found || ifData == nil {
		plugin.log.Debugf("cannot create ARP entry %s due to missing interface %s (found: %v, data: %v), cached",
			arpEntry.Name, arpEntry.Interface, found, ifData)
		plugin.arpIfCache[arpEntry.Name] = &ArpToInterface{
			arp:    arpEntry,
			ifName: arpEntry.Interface,
			isAdd:  true,
		}
		return nil
	}

	neigh.LinkIndex = int(ifData.Index)

	// Set IP address
	ipAddr := net.ParseIP(arpEntry.IpAddr)
	if ipAddr == nil {
		return fmt.Errorf("cannot create ARP entry %v, unable to parse IP address %v", arpEntry.Name, arpEntry.IpAddr)
	}
	neigh.IP = ipAddr

	// Set MAC address
	var mac net.HardwareAddr
	if mac, err = net.ParseMAC(arpEntry.HwAddress); err != nil {
		return fmt.Errorf("cannot create ARP entry %v, unable to parse MAC address %v, error: %v", arpEntry.Name,
			arpEntry.HwAddress, err)
	}
	neigh.HardwareAddr = mac

	// Set ARP entry state
	neigh.State = arpStateParser(arpEntry.State)

	// Set ip family
	neigh.Family = getIPFamily(arpEntry.IpFamily)

	// Prepare namespace of related interface
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	arpNs := plugin.nsHandler.ArpNsToGeneric(arpEntry.Namespace)

	// ARP entry has to be created in the same namespace as the interface
	revertNs, err := plugin.nsHandler.SwitchNamespace(arpNs, nsMgmtCtx)
	if err != nil {
		return err
	}
	defer revertNs()

	// Create a new ARP entry in interface namespace
	err = plugin.l3Handler.AddArpEntry(arpEntry.Name, neigh)
	if err != nil {
		plugin.log.Errorf("adding arp entry %q failed: %v (%+v)", arpEntry.Name, err, neigh)
		return err
	}

	// Register created ARP entry
	plugin.arpIndexes.RegisterName(ArpIdentifier(neigh), plugin.arpIdxSeq, arpEntry)
	plugin.arpIdxSeq++
	plugin.log.Debugf("ARP entry %v registered as %v", arpEntry.Name, ArpIdentifier(neigh))

	plugin.log.Infof("Linux ARP entry %v configured", arpEntry.Name)

	return nil
}

// ModifyLinuxStaticArpEntry applies changes in the NB configuration of a Linux ARP through Netlink API.
func (plugin *LinuxArpConfigurator) ModifyLinuxStaticArpEntry(newArpEntry *l3.LinuxStaticArpEntries_ArpEntry, oldArpEntry *l3.LinuxStaticArpEntries_ArpEntry) (err error) {
	plugin.log.Infof("Modifying Linux ARP entry %v", newArpEntry.Name)

	// If the namespace of the new ARP entry was changed, the old entry needs to be removed and the new one created
	// in the new namespace
	// If interface or IP address was changed, the old entry needs to be removed and recreated as well. In such a case,
	// ModifyArpEntry (analogy to 'ip neigh replace') would create a new entry instead of modifying the existing one
	callReplace := true

	oldArpNs := plugin.nsHandler.ArpNsToGeneric(oldArpEntry.Namespace)
	newArpNs := plugin.nsHandler.ArpNsToGeneric(newArpEntry.Namespace)
	result := oldArpNs.CompareNamespaces(newArpNs)
	if result != 0 || oldArpEntry.Interface != newArpEntry.Interface || oldArpEntry.IpAddr != newArpEntry.IpAddr {
		callReplace = false
	}

	// Remove old entry and configure a new one, then return
	if !callReplace {
		if err := plugin.DeleteLinuxStaticArpEntry(oldArpEntry); err != nil {
			return nil
		}
		return plugin.ConfigureLinuxStaticArpEntry(newArpEntry)
	}

	// Create modified ARP entry object
	neigh := &netlink.Neigh{}

	// Find interface
	_, ifData, found := plugin.ifIndexes.LookupIdx(newArpEntry.Interface)
	if !found || ifData == nil {
		return fmt.Errorf("cannot create ARP entry %s due to missing interface %s (found: %v, data: %v), cached",
			newArpEntry.Name, newArpEntry.Interface, found, ifData)
	}
	neigh.LinkIndex = int(ifData.Index)

	// Set IP address
	ipAddr := net.ParseIP(newArpEntry.IpAddr)
	if ipAddr == nil {
		return fmt.Errorf("cannot create ARP entry %v, unable to parse IP address %v", newArpEntry.Name, newArpEntry.IpAddr)
	}
	neigh.IP = ipAddr

	// Set MAC address
	var mac net.HardwareAddr
	if mac, err = net.ParseMAC(newArpEntry.HwAddress); err != nil {
		return fmt.Errorf("cannot create ARP entry %v, unable to parse MAC address %v, error: %v", newArpEntry.Name,
			newArpEntry.HwAddress, err)
	}
	neigh.HardwareAddr = mac

	// Set ARP entry state
	neigh.State = arpStateParser(newArpEntry.State)

	// Set ip family
	neigh.Family = getIPFamily(newArpEntry.IpFamily)

	// Prepare namespace of related interface
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	arpNs := plugin.nsHandler.ArpNsToGeneric(newArpEntry.Namespace)

	// ARP entry has to be created in the same namespace as the interface
	revertNs, err := plugin.nsHandler.SwitchNamespace(arpNs, nsMgmtCtx)
	if err != nil {
		return err
	}
	defer revertNs()

	err = plugin.l3Handler.SetArpEntry(newArpEntry.Name, neigh)
	if err != nil {
		plugin.log.Errorf("modifying arp entry %q failed: %v (%+v)", newArpEntry.Name, err, neigh)
		return err
	}

	plugin.log.Infof("Linux ARP entry %v modified", newArpEntry.Name)

	return nil
}

// DeleteLinuxStaticArpEntry reacts to a removed NB configuration of a Linux ARP entry.
func (plugin *LinuxArpConfigurator) DeleteLinuxStaticArpEntry(arpEntry *l3.LinuxStaticArpEntries_ArpEntry) (err error) {
	plugin.log.Infof("Deleting Linux ARP entry %v", arpEntry.Name)

	// Prepare ARP entry object
	neigh := &netlink.Neigh{}

	// Find interface
	_, ifData, foundIface := plugin.ifIndexes.LookupIdx(arpEntry.Interface)
	if !foundIface || ifData == nil {
		plugin.log.Debugf("cannot remove ARP entry %v due to missing interface %v, cached", arpEntry.Name, arpEntry.Interface)
		plugin.arpIfCache[arpEntry.Name] = &ArpToInterface{
			arp:    arpEntry,
			ifName: arpEntry.Interface,
		}
		return nil
	}
	neigh.LinkIndex = int(ifData.Index)

	// Set IP address
	ipAddr := net.ParseIP(arpEntry.IpAddr)
	if ipAddr == nil {
		return fmt.Errorf("cannot create ARP entry %v, unable to parse IP address %v", arpEntry.Name, arpEntry.IpAddr)
	}
	neigh.IP = ipAddr

	// Prepare namespace of related interface
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	arpNs := plugin.nsHandler.ArpNsToGeneric(arpEntry.Namespace)

	// ARP entry has to be removed from the same namespace as the interface
	revertNs, err := plugin.nsHandler.SwitchNamespace(arpNs, nsMgmtCtx)
	if err != nil {
		return err
	}
	defer revertNs()

	// Read all ARP entries configured for interface
	entries, err := plugin.l3Handler.GetArpEntries(int(ifData.Index), noFamilyFilter)
	if err != nil {
		return err
	}

	// Look for ARP to remove. If it already does not exist, return
	var found bool
	for _, entry := range entries {
		if compareARPLinkIdxAndIP(&entry, neigh) {
			found = true
			break
		}
	}
	if !found {
		plugin.log.Infof("ARP entry with IP %v and link index %v not configured, skipped", neigh.IP.String(), neigh.LinkIndex)
		return nil
	}

	// Remove the ARP entry from the interface namespace
	err = plugin.l3Handler.DelArpEntry(arpEntry.Name, neigh)
	if err != nil {
		plugin.log.Errorf("deleting arp entry %q failed: %v (%+v)", arpEntry.Name, err, neigh)
		return err
	}

	_, _, found = plugin.arpIndexes.UnregisterName(ArpIdentifier(neigh))
	if !found {
		plugin.log.Warnf("Attempt to unregister non-existing ARP entry %v", arpEntry.Name)
	} else {
		plugin.log.Debugf("ARP entry unregistered %v", arpEntry.Name)
	}

	plugin.log.Infof("Linux ARP entry %v removed", arpEntry.Name)

	return nil
}

// LookupLinuxArpEntries reads all ARP entries from all interfaces and registers them if needed
func (plugin *LinuxArpConfigurator) LookupLinuxArpEntries() error {
	plugin.log.Infof("Browsing Linux ARP entries")

	// Set interface index and family to 0 reads all arp entries from all of the interfaces
	entries, err := plugin.l3Handler.GetArpEntries(noIfaceIdxFilter, noFamilyFilter)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		plugin.log.WithField("interface", entry.LinkIndex).Debugf("Found new static linux ARP entry")
		_, arp, found := plugin.arpIndexes.LookupIdx(ArpIdentifier(&entry))
		if !found {
			var ifName string
			if arp == nil || arp.Namespace == nil {
				ifName, _, _ = plugin.ifIndexes.LookupNameByNamespace(uint32(entry.LinkIndex), ifaceidx.DefNs)
			} else {
				ifName, _, _ = plugin.ifIndexes.LookupNameByNamespace(uint32(entry.LinkIndex), arp.Namespace.Name)
			}
			plugin.arpIndexes.RegisterName(ArpIdentifier(&entry), plugin.arpIdxSeq, &l3.LinuxStaticArpEntries_ArpEntry{
				// Register fields required to reconstruct ARP identifier
				Interface: ifName,
				IpAddr:    entry.IP.String(),
				HwAddress: entry.HardwareAddr.String(),
			})
			plugin.arpIdxSeq++
			plugin.log.Debugf("ARP entry registered as %v", ArpIdentifier(&entry))
		}
	}

	return nil
}

// ResolveCreatedInterface resolves a new linux interface from ARP perspective
func (plugin *LinuxArpConfigurator) ResolveCreatedInterface(ifName string, ifIdx uint32) error {
	plugin.log.Debugf("Linux ARP configurator: resolve created interface %v", ifName)

	// Look for ARP entries where the interface is used
	var wasErr error
	for arpName, arpIfPair := range plugin.arpIfCache {
		if arpIfPair.ifName == ifName && arpIfPair.isAdd {
			plugin.log.Debugf("Cached ARP %v for interface %v created", arpName, ifName)
			if err := plugin.ConfigureLinuxStaticArpEntry(arpIfPair.arp); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
			delete(plugin.arpIfCache, arpName)
		} else if arpIfPair.ifName == ifName && !arpIfPair.isAdd {
			plugin.log.Debugf("Cached ARP %v for interface %v removed", arpName, ifName)
			if err := plugin.DeleteLinuxStaticArpEntry(arpIfPair.arp); err != nil {
				plugin.log.Error(err)
				wasErr = err
			}
			delete(plugin.arpIfCache, arpName)
		}
	}

	return wasErr
}

// ResolveDeletedInterface resolves removed linux interface from ARP perspective
func (plugin *LinuxArpConfigurator) ResolveDeletedInterface(ifName string, ifIdx uint32) error {
	plugin.log.Debugf("Linux ARP configurator: resolve deleted interface %v", ifName)

	// Read cache at first and remove obsolete entries
	for arpName, arpToIface := range plugin.arpIfCache {
		if arpToIface.ifName == ifName && !arpToIface.isAdd {
			delete(plugin.arpIfCache, arpName)
		}
	}

	// Read mapping of ARP entries and find all using removed interface
	for _, arpName := range plugin.arpIndexes.GetMapping().ListNames() {
		_, arp, found := plugin.arpIndexes.LookupIdx(arpName)
		if !found {
			// Should not happend but better to log it
			plugin.log.Warnf("ARP %v not found in the mapping", arpName)
			continue
		}
		if arp == nil {
			plugin.log.Warnf("ARP %v - no data available", arpName)
			continue
		}
		if arp.Interface == ifName {
			// Unregister
			ip := net.ParseIP(arp.IpAddr)
			if ip == nil {
				plugin.log.Errorf("ARP %v - cannot unregister, invalid IP %v", arpName, arp.IpAddr)
				continue
			}
			mac, err := net.ParseMAC(arp.HwAddress)
			if err != nil {
				plugin.log.Errorf("ARP %v - cannot unregister, invalid MAC %v: %v", arpName, arp.HwAddress, err)
				continue
			}
			plugin.arpIndexes.UnregisterName(ArpIdentifier(&netlink.Neigh{
				LinkIndex:    int(ifIdx),
				IP:           ip,
				HardwareAddr: mac,
			}))
			// Cache
			plugin.arpIfCache[arpName] = &ArpToInterface{
				arp:    arp,
				ifName: ifName,
				isAdd:  true,
			}
		}
	}

	return nil
}

// ArpIdentifier generates unique ARP ID used in mapping
func ArpIdentifier(arp *netlink.Neigh) string {
	return fmt.Sprintf("iface%v-%v-%v", arp.LinkIndex, arp.IP.String(), arp.HardwareAddr)
}

// arpStateParser returns representation of neighbor unreachability detection index as defined in netlink
func arpStateParser(stateType *l3.LinuxStaticArpEntries_ArpEntry_NudState) int {
	// if state is not set, set it to permanent as default
	if stateType == nil {
		return netlink.NUD_PERMANENT
	}
	state := stateType.Type
	switch state {
	case 0:
		return netlink.NUD_PERMANENT
	case 1:
		return netlink.NUD_NOARP
	case 2:
		return netlink.NUD_REACHABLE
	case 3:
		return netlink.NUD_STALE
	default:
		return netlink.NUD_PERMANENT
	}
}

// returns IP family netlink representation
func getIPFamily(family *l3.LinuxStaticArpEntries_ArpEntry_IpFamily) (arpIPFamily int) {
	if family == nil {
		return
	}
	if family.Family == l3.LinuxStaticArpEntries_ArpEntry_IpFamily_IPV4 {
		arpIPFamily = netlink.FAMILY_V4
	}
	if family.Family == l3.LinuxStaticArpEntries_ArpEntry_IpFamily_IPV6 {
		arpIPFamily = netlink.FAMILY_V6
	}
	if family.Family == l3.LinuxStaticArpEntries_ArpEntry_IpFamily_ALL {
		arpIPFamily = netlink.FAMILY_ALL
	}
	if family.Family == l3.LinuxStaticArpEntries_ArpEntry_IpFamily_MPLS {
		arpIPFamily = netlink.FAMILY_MPLS
	}
	return
}

func compareARPLinkIdxAndIP(arp1 *netlink.Neigh, arp2 *netlink.Neigh) bool {
	if arp1.LinkIndex != arp2.LinkIndex {
		return false
	}
	if arp1.IP.String() != arp2.IP.String() {
		return false
	}
	return true
}
