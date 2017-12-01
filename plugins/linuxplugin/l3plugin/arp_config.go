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

//go:generate protoc --proto_path=model --gogo_out=model model/l3/l3.proto

package l3plugin

import (
	"fmt"
	"net"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	common "github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
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
	Log logging.Logger

	LinuxIfIdx ifaceidx.LinuxIfIndexRW
	ArpIdxSeq  uint32
	arpIndexes l3idx.LinuxARPIndexRW

	// Time measurement
	Stopwatch *measure.Stopwatch // timer used to measure and store time

}

// Init initializes ARP configurator and starts goroutines
func (plugin *LinuxArpConfigurator) Init(arpIndexes l3idx.LinuxARPIndexRW) error {
	plugin.Log.Debug("Initializing Linux ARP configurator")
	plugin.arpIndexes = arpIndexes

	return nil
}

// Close closes all goroutines started during Init
func (plugin *LinuxArpConfigurator) Close() error {
	return nil
}

// ConfigureLinuxStaticArpEntry reacts to a new northbound Linux ARP entry config by creating and configuring
// the entry in the host network stack through Netlink API.
func (plugin *LinuxArpConfigurator) ConfigureLinuxStaticArpEntry(arpEntry *l3.LinuxStaticArpEntries_ArpEntry) error {
	plugin.Log.Infof("Configuring Linux ARP entry %v", arpEntry.Name)
	var err error

	// Prepare ARP entry object
	neigh := &netlink.Neigh{}

	// Find interface
	idx, _, found := plugin.LinuxIfIdx.LookupIdx(arpEntry.Interface)
	if !found {
		return fmt.Errorf("cannot create ARP entry %v, interface %v not found", arpEntry.Name, arpEntry.Interface)
	}
	neigh.LinkIndex = int(idx)

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
	neigh.Family = int(arpEntry.Family)

	// Prepare namespace of related interface
	nsMgmtCtx := common.NewNamespaceMgmtCtx()
	arpNs := linuxcalls.ToGenericArpNs(arpEntry.Namespace)

	// ARP entry has to be created in the same namespace as the interface
	revertNs, err := arpNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		return err
	}
	defer revertNs()

	// Create a new ARP entry in interface namespace
	err = linuxcalls.AddArpEntry(arpEntry.Name, neigh, plugin.Log, measure.GetTimeLog("add-arp-entry", plugin.Stopwatch))

	// Register created ARP entry
	plugin.arpIndexes.RegisterName(arpIdentifier(neigh), plugin.ArpIdxSeq, nil)
	plugin.ArpIdxSeq++
	plugin.Log.Debugf("ARP entry %v registered as %v", arpEntry.Name, arpIdentifier(neigh))

	plugin.Log.Infof("Linux ARP entry %v configured", arpEntry.Name)

	return err
}

// ModifyLinuxStaticArpEntry applies changes in the NB configuration of a Linux ARP through Netlink API.
func (plugin *LinuxArpConfigurator) ModifyLinuxStaticArpEntry(newArpEntry *l3.LinuxStaticArpEntries_ArpEntry, oldArpEntry *l3.LinuxStaticArpEntries_ArpEntry) error {
	plugin.Log.Infof("Modifying Linux ARP entry %v", newArpEntry.Name)
	var err error

	// If the namespace of the new ARP entry was changed, the old entry needs to be removed and the new one created
	// in the new namespace
	// If interface or IP address was changed, the old entry needs to be removed and recreated as well. In such a case,
	// ModifyArpEntry (analogy to 'ip neigh replace') would create a new entry instead of modifying the existing one
	callReplace := true

	oldArpNs := linuxcalls.ToGenericArpNs(oldArpEntry.Namespace)
	newArpNs := linuxcalls.ToGenericArpNs(newArpEntry.Namespace)
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
	idx, _, found := plugin.LinuxIfIdx.LookupIdx(newArpEntry.Interface)
	if !found {
		return fmt.Errorf("cannot create ARP entry %v, interface %v not found", newArpEntry.Name, newArpEntry.Interface)
	}
	neigh.LinkIndex = int(idx)

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
	neigh.Family = int(newArpEntry.Family)

	// Prepare namespace of related interface
	nsMgmtCtx := common.NewNamespaceMgmtCtx()
	arpNs := linuxcalls.ToGenericArpNs(newArpEntry.Namespace)

	// ARP entry has to be created in the same namespace as the interface
	revertNs, err := arpNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		return err
	}
	defer revertNs()

	err = linuxcalls.ModifyArpEntry(newArpEntry.Name, neigh, plugin.Log, measure.GetTimeLog("modify-arp-entry", plugin.Stopwatch))

	plugin.Log.Infof("Linux ARP entry %v modified", newArpEntry.Name)

	return err
}

// DeleteLinuxStaticArpEntry reacts to a removed NB configuration of a Linux ARP entry.
func (plugin *LinuxArpConfigurator) DeleteLinuxStaticArpEntry(arpEntry *l3.LinuxStaticArpEntries_ArpEntry) error {
	plugin.Log.Infof("Deleting Linux ARP entry %v", arpEntry.Name)
	var err error

	// Prepare ARP entry object
	neigh := &netlink.Neigh{}

	// Find interface
	idx, _, foundIface := plugin.LinuxIfIdx.LookupIdx(arpEntry.Interface)
	if !foundIface {
		return fmt.Errorf("cannot remove ARP entry %v, interface %v not found", arpEntry.Name, arpEntry.Interface)
	}
	neigh.LinkIndex = int(idx)

	// Set IP address
	ipAddr := net.ParseIP(arpEntry.IpAddr)
	if ipAddr == nil {
		return fmt.Errorf("cannot create ARP entry %v, unable to parse IP address %v", arpEntry.Name, arpEntry.IpAddr)
	}
	neigh.IP = ipAddr

	// Prepare namespace of related interface
	nsMgmtCtx := common.NewNamespaceMgmtCtx()
	arpNs := linuxcalls.ToGenericArpNs(arpEntry.Namespace)

	// ARP entry has to be removed from the same namespace as the interface
	revertNs, err := arpNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		return err
	}
	defer revertNs()

	// Read all ARP entries configured for interface
	entries, err := linuxcalls.ReadArpEntries(int(idx), noFamilyFilter, plugin.Log, measure.GetTimeLog("list-arp-entries", plugin.Stopwatch))
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
		plugin.Log.Infof("ARP entry with IP %v and link index %v not configured, skipped", neigh.IP.String(), neigh.LinkIndex)
		return nil
	}

	// Remove the ARP entry from the interface namespace
	err = linuxcalls.DeleteArpEntry(arpEntry.Name, neigh, plugin.Log, measure.GetTimeLog("del-arp-entry", plugin.Stopwatch))

	_, _, found = plugin.arpIndexes.UnregisterName(arpIdentifier(neigh))
	if !found {
		plugin.Log.Warnf("Attempt to unregister non-existing ARP entry %v", arpEntry.Name)
	} else {
		plugin.Log.Debugf("ARP entry unregistered %v", arpEntry.Name)
	}

	plugin.Log.Infof("Linux ARP entry %v removed", arpEntry.Name)

	return err
}

// LookupLinuxArpEntries reads all ARP entries from all interfaces and registers them if needed
func (plugin *LinuxArpConfigurator) LookupLinuxArpEntries() error {
	plugin.Log.Infof("Browsing Linux ARP entries")

	// Set interface index and family to 0 reads all arp entries from all of the interfaces
	entries, err := linuxcalls.ReadArpEntries(noIfaceIdxFilter, noFamilyFilter, plugin.Log, nil)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		plugin.Log.WithField("interface", entry.LinkIndex).Debugf("Found new static linux ARP entry")
		_, _, found := plugin.arpIndexes.LookupIdx(arpIdentifier(&entry))
		if !found {
			plugin.arpIndexes.RegisterName(arpIdentifier(&entry), plugin.ArpIdxSeq, nil)
			plugin.ArpIdxSeq++
			plugin.Log.Debug("ARP entry registered as %v", arpIdentifier(&entry))
		}
	}

	return nil
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

func compareARPLinkIdxAndIP(arp1 *netlink.Neigh, arp2 *netlink.Neigh) bool {
	if arp1.LinkIndex != arp2.LinkIndex {
		return false
	}
	if arp1.IP.String() != arp2.IP.String() {
		return false
	}
	return true
}

func arpIdentifier(arp *netlink.Neigh) string {
	return fmt.Sprintf("iface%v-%v-%v", arp.LinkIndex, arp.IP.String(), arp.HardwareAddr)
}
