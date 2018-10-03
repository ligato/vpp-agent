package descriptor

import (
	"strings"
	"github.com/gogo/protobuf/proto"
	"github.com/go-errors/errors"

	"github.com/ligato/cn-infra/utils/addrs"

	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	"time"
)

// Add creates a VPP interface.
func (d *InterfaceDescriptor) Add(key string, intf *interfaces.Interface) (metadata *ifaceidx.IfaceMetadata, err error) {
	var ifIdx uint32

	// validate configuration first
	err = d.validateInterfaceConfig(intf)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	// create interface of the given type
	switch intf.Type {
	case interfaces.Interface_TAP_INTERFACE:
		ifIdx, err = d.ifHandler.AddTapInterface(intf.Name, intf.GetTap())
		if err != nil {
			d.log.Error(err)
			return nil, err
		}

		// verify that the Linux side was created
		if d.linuxIfHandler != nil {
			var exists bool
			startTime := time.Now()

			for !exists && time.Since(startTime) < tapHostInterfaceWaitTimeout {
				exists, err := d.linuxIfHandler.InterfaceExists(intf.GetTap().GetHostIfName())
				if err != nil {
					d.log.Error(err)
					return nil, err
				}
				if !exists {
					time.Sleep(10 * time.Millisecond)
				}
			}

			if !exists {
				err = errors.Errorf("failed to create the Linux side of the TAP interface %s", intf.Name)
				d.log.Error(err)
				return nil, err
			}
		}

	case interfaces.Interface_MEMORY_INTERFACE:
		var socketID uint32
		if socketID, err = d.resolveMemifSocketFilename(intf.GetMemif()); err != nil {
			d.log.Error(err)
			return nil, err
		}
		ifIdx, err = d.ifHandler.AddMemifInterface(intf.Name, intf.GetMemif(), socketID)
		if err != nil {
			d.log.Error(err)
			return nil, err
		}

	case interfaces.Interface_VXLAN_TUNNEL:
		var multicastIfIdx uint32
		multicastIf := intf.GetVxlan().GetMulticast()
		if multicastIf != "" {
			multicastMeta, found := d.intfIndex.LookupByName(multicastIf)
			if !found {
				err = errors.Errorf("failed to find multicast interface %s referenced by VXLAN %s",
					multicastIf, intf.Name)
				d.log.Error(err)
				return nil, err
			}
			multicastIfIdx = multicastMeta.SwIfIndex
		}
		ifIdx, err = d.ifHandler.AddVxLanTunnel(intf.Name, intf.GetVrf(), multicastIfIdx, intf.GetVxlan())
		if err != nil {
			d.log.Error(err)
			return nil, err
		}

	case interfaces.Interface_SOFTWARE_LOOPBACK:
		ifIdx, err = d.ifHandler.AddLoopbackInterface(intf.Name)
		if err != nil {
			d.log.Error(err)
			return nil, err
		}

	case interfaces.Interface_ETHERNET_CSMACD:
		ifMeta, found := d.intfIndex.LookupByName(intf.Name)
		if !found {
			err = errors.Errorf("failed to find physical interface %s", intf.Name)
			return nil, err
		}
		ifIdx = ifMeta.SwIfIndex

	case interfaces.Interface_AF_PACKET_INTERFACE:
		ifIdx, err = d.ifHandler.AddAfPacketInterface(intf.Name, intf.GetPhysAddress(), intf.GetAfpacket())
		if err != nil {
			d.log.Error(err)
			return nil, err
		}
	}

	// Rx-mode
	if err = d.configRxModeForInterface(intf, ifIdx); err != nil {
		err = errors.Errorf("failed to set Rx-mode for interface %s: %v", intf.Name, err)
		d.log.Error(err)
		return nil, err
	}

	// Rx-placement
	if intf.GetRxPlacementSettings() != nil {
		if err = d.ifHandler.SetRxPlacement(ifIdx, intf.GetRxPlacementSettings()); err != nil {
			err = errors.Errorf("failed to set rx-placement for interface %s: %v", intf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	// MAC address (optional, for af-packet is configured in different way)
	if intf.GetPhysAddress() != "" && intf.GetType() != interfaces.Interface_AF_PACKET_INTERFACE {
		if err = d.ifHandler.SetInterfaceMac(ifIdx, intf.GetPhysAddress()); err != nil {
			err = errors.Errorf("failed to set MAC address %s to interface %s: %v",
				intf.GetPhysAddress(), intf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	// Get IP addresses
	ipAddrs, err := addrs.StrAddrsToStruct(intf.IpAddresses)
	if err != nil {
		err = errors.Errorf("failed to convert %s IP address list to IPNet structures: %v", intf.Name, err)
		d.log.Error(err)
		return nil, err
	}

	// VRF (optional, unavailable for VXLAN interfaces), has to be done before IP addresses are configured
	if intf.GetType() != interfaces.Interface_VXLAN_TUNNEL {
		// Configured separately for IPv4/IPv6
		isIPv4, isIPv6 := getIPAddressVersions(ipAddrs)
		if isIPv4 {
			if err = d.ifHandler.SetInterfaceVrf(ifIdx, intf.Vrf); err != nil {
				err = errors.Errorf("failed to set interface %s as IPv4 VRF %d: %v", intf.Name, intf.Vrf, err)
				d.log.Error(err)
				return nil, err
			}
		}
		if isIPv6 {
			if err := d.ifHandler.SetInterfaceVrfIPv6(ifIdx, intf.Vrf); err != nil {
				err = errors.Errorf("failed to set interface %s as IPv6 VRF %d: %v", intf.Name, intf.Vrf, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	// Configure IP addresses
	for _, address := range ipAddrs {
		if err := d.ifHandler.AddInterfaceIP(ifIdx, address); err != nil {
			err = errors.Errorf("adding IP address %v to interface %v failed: %v", address, intf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	// Configure mtu. Prefer value in the interface config, otherwise set the plugin-wide
	// default value if provided.
	if intf.Type != interfaces.Interface_VXLAN_TUNNEL {
		mtuToConfigure := intf.Mtu
		if mtuToConfigure == 0 && d.defaultMtu != 0 {
			mtuToConfigure = d.defaultMtu
		}
		if mtuToConfigure != 0 {
			if err = d.ifHandler.SetInterfaceMtu(ifIdx, mtuToConfigure); err != nil {
				err = errors.Errorf("failed to set MTU %d to interface %s: %v", mtuToConfigure, intf.Name, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	// set interface up if enabled
	// TODO: process admin up/down notification only after Add finalizes (e.g. using a "transaction barrier")
	if intf.Enabled {
		if err = d.ifHandler.InterfaceAdminUp(ifIdx); err != nil {
			err = errors.Errorf("failed to set interface %s up: %v", intf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	// fill the metadata
	metadata = &ifaceidx.IfaceMetadata{
		SwIfIndex:    ifIdx,
		IPAddresses:  intf.GetIpAddresses(),
	}
	return metadata, nil
}

// Delete removes VPP interface.
func (d *InterfaceDescriptor) Delete(key string, intf *interfaces.Interface, metadata *ifaceidx.IfaceMetadata) error {
	ifIdx := metadata.SwIfIndex

	// Skip setting interface to ADMIN_DOWN unless the type AF_PACKET_INTERFACE
	if intf.Type != interfaces.Interface_AF_PACKET_INTERFACE {
		if err := d.ifHandler.InterfaceAdminDown(ifIdx); err != nil {
			err = errors.Errorf("failed to set interface %s down: %v", intf.Name, err)
			d.log.Error(err)
			return err
		}
	}

	// unconfigure IP addresses
	var nonLocalIPs []string
	for _, ipAddr := range intf.IpAddresses {
		if strings.HasPrefix(ipAddr, "fe80") {
			nonLocalIPs = append(nonLocalIPs, ipAddr)
		}
	}
	ipAddrs, err := addrs.StrAddrsToStruct(nonLocalIPs)
	if err != nil {
		err = errors.Errorf("failed to convert %s IP address list to IPNet structures: %v", intf.Name, err)
		d.log.Error(err)
		return err
	}
	for _, ipAddr := range ipAddrs {
		if err = d.ifHandler.DelInterfaceIP(ifIdx, ipAddr); err != nil {
			err = errors.Errorf("failed to remove IP address %s from interface %s: %v",
				ipAddr, intf.Name, err)
			d.log.Error(err)
			return err
		}
	}

	// remove the interface
	switch intf.Type {
	case interfaces.Interface_TAP_INTERFACE:
		err = d.ifHandler.DeleteTapInterface(intf.Name, ifIdx, intf.GetTap().GetVersion())
	case interfaces.Interface_MEMORY_INTERFACE:
		err = d.ifHandler.DeleteMemifInterface(intf.Name, ifIdx)
	case interfaces.Interface_VXLAN_TUNNEL:
		err = d.ifHandler.DeleteVxLanTunnel(intf.Name, ifIdx, intf.Vrf, intf.GetVxlan())
	case interfaces.Interface_SOFTWARE_LOOPBACK:
		err = d.ifHandler.DeleteLoopbackInterface(intf.Name, ifIdx)
	case interfaces.Interface_ETHERNET_CSMACD:
		d.log.Debugf("Interface %s removal skipped: cannot remove (blacklist) physical interface", intf.Name) // Not an error
		return nil
	case interfaces.Interface_AF_PACKET_INTERFACE:
		err = d.ifHandler.DeleteAfPacketInterface(intf.Name, ifIdx, intf.GetAfpacket())
	}
	if err != nil {
		err = errors.Errorf("failed to remove interface %s, index %d: %v", intf.Name, ifIdx, err)
		d.log.Error(err)
		return err
	}
	return nil
}

// Modify is able to change Type-unspecific attributes.
func (d *InterfaceDescriptor) Modify(key string, oldIntf, newIntf *interfaces.Interface, oldMetadata *ifaceidx.IfaceMetadata) (newMetadata *ifaceidx.IfaceMetadata, err error) {
	// validate the new configuration first
	err = d.validateInterfaceConfig(newIntf)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	ifIdx := oldMetadata.SwIfIndex

	// Rx-mode
	if err := d.modifyRxModeForInterfaces(oldIntf, newIntf, ifIdx); err != nil {
		err = errors.Errorf("failed to modify rx-mode for interface %s: %v", newIntf.Name, err)
		d.log.Error(err)
		return oldMetadata, err
	}

	// Rx-placement
	if newIntf.RxPlacementSettings != nil && !proto.Equal(oldIntf.RxPlacementSettings, newIntf.RxPlacementSettings) {
		if err = d.ifHandler.SetRxPlacement(ifIdx, newIntf.GetRxPlacementSettings()); err != nil {
			err = errors.Errorf("failed to modify rx-placement for interface %s: %v", newIntf.Name, err)
			d.log.Error(err)
			return oldMetadata, err
		}
	}

	// Admin status
	if newIntf.Enabled != oldIntf.Enabled {
		if newIntf.Enabled {
			if err = d.ifHandler.InterfaceAdminUp(ifIdx); err != nil {
				err = errors.Errorf("failed to set interface %s up: %v", newIntf.Name, err)
				d.log.Error(err)
				return oldMetadata, err
			}
		} else {
			if err = d.ifHandler.InterfaceAdminDown(ifIdx); err != nil {
				err = errors.Errorf("failed to set interface %s down: %v", newIntf.Name, err)
				d.log.Error(err)
				return oldMetadata, err
			}
		}
	}

	// Configure new mac address if set (and only if it was changed)
	if newIntf.PhysAddress != "" && newIntf.PhysAddress != oldIntf.PhysAddress {
		if err := d.ifHandler.SetInterfaceMac(ifIdx, newIntf.PhysAddress); err != nil {
			err = errors.Errorf("setting interface %s MAC address %s failed: %v",
				newIntf.Name, newIntf.PhysAddress, err)
			d.log.Error(err)
			return oldMetadata, err
		}
	}

	// Calculate diff of IP addresses
	newIPAddrs, err := addrs.StrAddrsToStruct(newIntf.IpAddresses)
	if err != nil {
		err = errors.Errorf("failed to convert %s IP address list to IPNet structures: %v", newIntf.Name, err)
		d.log.Error(err)
		return oldMetadata, err
	}
	oldIPAddrs, err := addrs.StrAddrsToStruct(oldIntf.IpAddresses)
	if err != nil {
		err = errors.Errorf("failed to convert %s IP address list to IPNet structures: %v", oldIntf.Name, err)
		d.log.Error(err)
		return oldMetadata, err
	}
	del, add := addrs.DiffAddr(newIPAddrs, oldIPAddrs)

	// Delete obsolete IP addresses
	for _, address := range del {
		err := d.ifHandler.DelInterfaceIP(ifIdx, address)
		if nil != err {
			err = errors.Errorf("failed to remove obsolete IP address %v from interface %s: %v",
				address, newIntf.Name, err)
			d.log.Error(err)
			return oldMetadata, err
		}
	}

	// Add new IP addresses
	for _, address := range add {
		err := d.ifHandler.AddInterfaceIP(ifIdx, address)
		if nil != err {
			err = errors.Errorf("failed to add new IP addresses %v to interface %s: %v",
				address, newIntf.Name, err)
			d.log.Error(err)
			return oldMetadata, err
		}
	}

	// update IP addresses in the metadata
	oldMetadata.IPAddresses = newIntf.IpAddresses

	// update MTU
	if newIntf.Mtu != 0 && newIntf.Mtu != oldIntf.Mtu {
		if err := d.ifHandler.SetInterfaceMtu(ifIdx, newIntf.Mtu); err != nil {
			err = errors.Errorf("failed to set MTU to interface %s: %v", newIntf.Name, err)
			d.log.Error(err)
			return oldMetadata, err
		}
	} else if newIntf.Mtu == 0 && d.defaultMtu != 0 {
		if err := d.ifHandler.SetInterfaceMtu(ifIdx, d.defaultMtu); err != nil {
			err = errors.Errorf("failed to set MTU to interface %s: %v", newIntf.Name, err)
			d.log.Error(err)
			return oldMetadata, err
		}
	}

	return oldMetadata, nil
}

// Dump returns all configured VPP interfaces.
func (d *InterfaceDescriptor) Dump(correlate []adapter.InterfaceKVWithMetadata) (dump []adapter.InterfaceKVWithMetadata, err error) {
	// refresh the map of memif socket IDs
	d.memifSocketToID, err = d.ifHandler.DumpMemifSocketDetails()
	if err != nil {
		err = errors.Errorf("failed to dump memif socket details: %v", err)
		d.log.Error(err)
		return dump, err
	}

	// dump current state of VPP interfaces
	vppIfs, err := d.ifHandler.DumpInterfaces()
	if err != nil {
		err = errors.Errorf("failed to dump interfaces: %v", err)
		d.log.Error(err)
		return dump, err
	}

	for ifIdx, intf := range vppIfs {
		origin := scheduler.FromNB
		if ifIdx == 0 {
			// local0 is created automatically
			origin = scheduler.FromSB
		}
		if intf.Interface.Type == interfaces.Interface_ETHERNET_CSMACD &&
			!intf.Interface.Enabled && len(intf.Interface.IpAddresses) == 0 {
			// unconfigured physical interface
			origin = scheduler.FromSB
		}
		if intf.Interface.Name == "" {
			// untagged interface - generate a logical name for it
			// (apart from local0 it will get removed by resync)
			intf.Interface.Name = untaggedIfPreffix + intf.Meta.InternalName
		}

		// verify links between VPP and Linux side
		if d.linuxIfPlugin != nil && d.linuxIfHandler != nil {
			if intf.Interface.Type == interfaces.Interface_AF_PACKET_INTERFACE {
				hostIfName := intf.Interface.GetAfpacket().HostIfName
				exists, _ := d.linuxIfHandler.InterfaceExists(hostIfName)
				if !exists {
					// the Linux interface that the AF-Packet is attached to does not exist
					// -> clear the host name so that the AF-Packet will be re-created
					intf.Interface.GetAfpacket().HostIfName = ""
				}
			}
			if intf.Interface.Type == interfaces.Interface_TAP_INTERFACE {
				hostIfName := intf.Interface.GetTap().GetHostIfName()
				exists, _ := d.linuxIfHandler.InterfaceExists(hostIfName)
				if !exists {
					// check if it was "stolen" by the Linux plugin
					_, _, exists = d.linuxIfPlugin.GetInterfaceIndex().LookupByTapTempName(
						intf.Interface.GetTap().GetHostIfName())
				}
				if !exists {
					// the Linux side of the TAP interface side was not found
					// -> clear the TAP host name so that the TAP will be re-created
					intf.Interface.GetTap().HostIfName = ""
				}
			}
		}

		// add interface record into the dump
		metadata := &ifaceidx.IfaceMetadata{
			SwIfIndex:    ifIdx,
			IPAddresses:  intf.Interface.IpAddresses,
		}
		dump = append(dump, adapter.InterfaceKVWithMetadata{
			Key:      interfaces.InterfaceKey(intf.Interface.Name),
			Value:    intf.Interface,
			Metadata: metadata,
			Origin:   origin,
		})

	}

	d.log.WithField("dump", dump).Debug("Dumping VPP interfaces")
	return dump, nil
}

