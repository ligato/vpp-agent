// Copyright (c) 2018 Cisco and/or its affiliates.
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

package vppcalls

import (
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
)

// InterfaceDetails is the wrapper structure for the interface northbound API structure.
type InterfaceDetails struct {
	Interface *interfaces.Interface `json:"interface"`
	Meta      *InterfaceMeta        `json:"interface_meta"`
}

// InterfaceMeta is combination of proto-modelled Interface data and VPP provided metadata
type InterfaceMeta struct {
	SwIfIndex    uint32           `json:"sw_if_index"`
	SupSwIfIndex uint32           `json:"sub_sw_if_index"`
	L2Address    net.HardwareAddr `json:"l2_address"`
	InternalName string           `json:"internal_name"`
	AdminState   uint8            `json:"admin_state"`
	LinkState    uint8            `json:"link_state"`
	LinkDuplex   uint8            `json:"link_duplex"`
	LinkMTU      uint16           `json:"link_mtu"`
	LinkSpeed    uint32           `json:"link_speed"`
	SubID        uint32           `json:"sub_id"`
	Tag          string           `json:"tag"`
	// dhcp
	Dhcp *Dhcp `json:"dhcp"`
	// vrf
	VrfIPv4 uint32 `json:"vrf_ipv4"`
	VrfIPv6 uint32 `json:"vrf_ipv6"`
	// wmxnet3
	Pci uint32 `json:"pci"`
}

// InterfaceEvent represents interface event from VPP.
type InterfaceEvent struct {
	SwIfIndex  uint32
	AdminState uint8
	LinkState  uint8
	Deleted    bool
}

// Dhcp is helper struct for DHCP metadata, split to client and lease (similar to VPP binary API)
type Dhcp struct {
	Client *Client `json:"dhcp_client"`
	Lease  *Lease  `json:"dhcp_lease"`
}

// Client is helper struct grouping DHCP client data
type Client struct {
	SwIfIndex        uint32
	Hostname         string
	ID               string
	WantDhcpEvent    bool
	SetBroadcastFlag bool
	PID              uint32
}

// Lease is helper struct grouping DHCP lease data
type Lease struct {
	SwIfIndex     uint32
	State         uint8
	Hostname      string
	IsIPv6        bool
	MaskWidth     uint8
	HostAddress   string
	RouterAddress string
	HostMac       string
}

// InterfaceState is a helper function grouping interface state data.
type InterfaceState struct {
	AdminState interfaces.InterfaceState_Status
	LinkState  interfaces.InterfaceState_Status
}

// InterfaceVppAPI provides methods for creating and managing interface plugin
type InterfaceVppAPI interface {
	InterfaceVppRead

	// AddAfPacketInterface calls AfPacketCreate VPP binary API.
	AddAfPacketInterface(ifName string, hwAddr string, afPacketIntf *interfaces.AfpacketLink) (swIndex uint32, err error)
	// DeleteAfPacketInterface calls AfPacketDelete VPP binary API.
	DeleteAfPacketInterface(ifName string, idx uint32, afPacketIntf *interfaces.AfpacketLink) error
	// AddLoopbackInterface calls CreateLoopback bin API.
	AddLoopbackInterface(ifName string) (swIndex uint32, err error)
	// DeleteLoopbackInterface calls DeleteLoopback bin API.
	DeleteLoopbackInterface(ifName string, idx uint32) error
	// AddMemifInterface calls MemifCreate bin API.
	AddMemifInterface(ifName string, memIface *interfaces.MemifLink, socketID uint32) (swIdx uint32, err error)
	// DeleteMemifInterface calls MemifDelete bin API.
	DeleteMemifInterface(ifName string, idx uint32) error
	// AddTapInterface calls TapConnect bin API.
	AddTapInterface(ifName string, tapIf *interfaces.TapLink) (swIfIdx uint32, err error)
	// DeleteTapInterface calls TapDelete bin API.
	DeleteTapInterface(ifName string, idx uint32, version uint32) error
	// AddVxLanTunnel calls AddDelVxLanTunnelReq with flag add=1.
	AddVxLanTunnel(ifName string, vrf, multicastIf uint32, vxLan *interfaces.VxlanLink) (swIndex uint32, err error)
	// DeleteVxLanTunnel calls AddDelVxLanTunnelReq with flag add=0.
	DeleteVxLanTunnel(ifName string, idx, vrf uint32, vxLan *interfaces.VxlanLink) error
	// AddIPSecTunnelInterface adds a new IPSec tunnel interface
	AddIPSecTunnelInterface(ifName string, ipSecLink *interfaces.IPSecLink) (uint32, error)
	// DeleteIPSecTunnelInterface removes existing IPSec tunnel interface
	DeleteIPSecTunnelInterface(ifName string, ipSecLink *interfaces.IPSecLink) error
	// AddVmxNet3 configures vmxNet3 interface. Second parameter is optional in this case.
	AddVmxNet3(ifName string, vmxNet3 *interfaces.VmxNet3Link) (uint32, error)
	// DeleteVmxNet3 removes vmxNet3 interface
	DeleteVmxNet3(ifName string, ifIdx uint32) error
	// AddBondInterface configures bond interface.
	AddBondInterface(ifName string, mac string, bondLink *interfaces.BondLink) (uint32, error)
	// DeleteBondInterface removes bond interface.
	DeleteBondInterface(ifName string, ifIdx uint32) error
	// InterfaceAdminDown calls binary API SwInterfaceSetFlagsReply with AdminUpDown=0.
	InterfaceAdminDown(ifIdx uint32) error
	// InterfaceAdminUp calls binary API SwInterfaceSetFlagsReply with AdminUpDown=1.
	InterfaceAdminUp(ifIdx uint32) error
	// SetInterfaceTag registers new interface index/tag pair
	SetInterfaceTag(tag string, ifIdx uint32) error
	// RemoveInterfaceTag un-registers new interface index/tag pair
	RemoveInterfaceTag(tag string, ifIdx uint32) error
	// SetInterfaceAsDHCPClient sets provided interface as a DHCP client
	SetInterfaceAsDHCPClient(ifIdx uint32, hostName string) error
	// UnsetInterfaceAsDHCPClient un-sets interface as DHCP client
	UnsetInterfaceAsDHCPClient(ifIdx uint32, hostName string) error
	// AddContainerIP calls IPContainerProxyAddDel VPP API with IsAdd=1
	AddContainerIP(ifIdx uint32, addr string) error
	// DelContainerIP calls IPContainerProxyAddDel VPP API with IsAdd=0
	DelContainerIP(ifIdx uint32, addr string) error
	// AddInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=1.
	AddInterfaceIP(ifIdx uint32, addr *net.IPNet) error
	// DelInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=00.
	DelInterfaceIP(ifIdx uint32, addr *net.IPNet) error
	// SetUnnumberedIP sets interface as un-numbered, linking IP address of the another interface (ifIdxWithIP)
	SetUnnumberedIP(uIfIdx uint32, ifIdxWithIP uint32) error
	// UnsetUnnumberedIP unset provided interface as un-numbered. IP address of the linked interface is removed
	UnsetUnnumberedIP(uIfIdx uint32) error
	// SetInterfaceMac calls SwInterfaceSetMacAddress bin API.
	SetInterfaceMac(ifIdx uint32, macAddress string) error
	// RegisterMemifSocketFilename registers new socket file name with provided ID.
	RegisterMemifSocketFilename(filename []byte, id uint32) error
	// SetInterfaceMtu calls HwInterfaceSetMtu bin API with desired MTU value.
	SetInterfaceMtu(ifIdx uint32, mtu uint32) error
	// SetRxMode calls SwInterfaceSetRxMode bin API
	SetRxMode(ifIdx uint32, rxMode *interfaces.Interface_RxMode) error
	// SetRxPlacement configures rx-placement for interface
	SetRxPlacement(ifIdx uint32, rxPlacement *interfaces.Interface_RxPlacement) error
	// SetInterfaceVrf sets VRF table for the interface
	SetInterfaceVrf(ifaceIndex, vrfID uint32) error
	// SetInterfaceVrfIPv6 sets IPV6 VRF table for the interface
	SetInterfaceVrfIPv6(ifaceIndex, vrfID uint32) error
	// CreateSubif creates sub interface.
	CreateSubif(ifIdx, vlanID uint32) (swIfIdx uint32, err error)
	// DeleteSubif deletes sub interface.
	DeleteSubif(ifIdx uint32) error
	// AttachInterfaceToBond adds interface as a slave to the bond interface.
	AttachInterfaceToBond(ifIdx, bondIfIdx uint32, isPassive, isLongTimeout bool) error
	// DetachInterfaceFromBond removes interface slave status from any bond interfaces.
	DetachInterfaceFromBond(ifIdx uint32) error
    // SetVLanTagRewrite sets VLan tag rewrite rule for given sub-interface
	SetVLanTagRewrite(ifIdx uint32, subIf *interfaces.SubInterface) error
}

// InterfaceVppRead provides read methods for interface plugin
type InterfaceVppRead interface {
	// DumpInterfaces dumps VPP interface data into the northbound API data structure
	// map indexed by software interface index.
	//
	// LIMITATIONS:
	// - there is no af_packet dump binary API. We relay on naming conventions of the internal VPP interface names
	DumpInterfaces() (map[uint32]*InterfaceDetails, error)
	// DumpInterfacesByType returns all VPP interfaces of the specified type
	DumpInterfacesByType(reqType interfaces.Interface_Type) (map[uint32]*InterfaceDetails, error)
	// GetInterfaceVrf reads VRF table to interface
	GetInterfaceVrf(ifIdx uint32) (vrfID uint32, err error)
	// GetInterfaceVrfIPv6 reads IPv6 VRF table to interface
	GetInterfaceVrfIPv6(ifIdx uint32) (vrfID uint32, err error)
	// DumpMemifSocketDetails dumps memif socket details from the VPP
	DumpMemifSocketDetails() (map[string]uint32, error)
	// DumpDhcpClients dumps DHCP-related information for all interfaces.
	DumpDhcpClients() (map[uint32]*Dhcp, error)
	// DumpInterfaceStates dumps link and administrative state of every interface.
	DumpInterfaceStates() (map[uint32]*InterfaceState, error)
	// WatchInterfaceEvents starts watching for interface events.
	WatchInterfaceEvents(ch chan<- *InterfaceEvent) error
	// WatchDHCPLeases starts watching for DHCP leases.
	WatchDHCPLeases(ch chan<- *Lease) error
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, logging.Logger) InterfaceVppAPI
}

func CompatibleInterfaceVppHandler(ch govppapi.Channel, log logging.Logger) InterfaceVppAPI {
	for ver, h := range Versions {
		log.Debugf("checking compatibility with %s", ver)
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch, log)
	}
	panic("no compatible version available")
}
