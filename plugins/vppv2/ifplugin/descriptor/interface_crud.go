package descriptor

import (
	"github.com/go-errors/errors"

	"github.com/ligato/cn-infra/utils/addrs"

	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	"strings"
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
		var found bool
		ifIdx, found = d.ethernetIntfs[intf.Name]
		if !found {
			err = errors.Errorf("failed to find physical interface %s", intf.Name)
			return nil, err
		}

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

	// DHCP client
	if intf.SetDhcpClient {
		if err := d.ifHandler.SetInterfaceAsDHCPClient(ifIdx, intf.Name); err != nil {
			err = errors.Errorf("failed to set interface %s as DHCP client", intf.Name)
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

	// Configure mtu. Prefer value in the interface config, otherwise set plugin-wide
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

	// Remove DHCP if it was set
	if intf.SetDhcpClient {
		if err := d.ifHandler.UnsetInterfaceAsDHCPClient(ifIdx, intf.Name); err != nil {
			err = errors.Errorf("failed to unset interface %s as DHCP client: %v", intf.Name, err)
			d.log.Error(err)
			return err
		}
		// TODO: with DHCP descriptor, here we will send notification about unconfigured DHCP client
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

	// TODO

	return oldMetadata, nil
}

// Dump returns all configured VPP interfaces.
func (d *InterfaceDescriptor) Dump(correlate []adapter.InterfaceKVWithMetadata) ([]adapter.InterfaceKVWithMetadata, error) {
	var dump []adapter.InterfaceKVWithMetadata

	// TODO (do not forget to refill d.ethernetIntfs and d.memifSocketToID)

	d.log.WithField("dump", dump).Debug("Dumping VPP interfaces")
	return dump, nil
}

