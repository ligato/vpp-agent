// Package iftst provides tools and input data for unit testing of ifplugin.
// What remains to be defined are scenarios.
package iftst

import (
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
)

var (
	// Memif100011Master is an example of a memory interface configuration. (Master=true)
	Memif100011Master = interfaces.Interfaces_Interface{
		Name:    "memif1",
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Master:         true,
			SocketFilename: "/tmp/memif1.sock",
		},
		Mtu:         1500,
		IpAddresses: []string{"10.0.0.11/24"},
	}

	// Memif100011Slave is an example of a memory interface configuration. It is
	// intentionally similar to Memif100011. The only difference is that Master=false.
	Memif100011Slave = interfaces.Interfaces_Interface{
		Name:    "memif1",
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Master:         false,
			SocketFilename: "/tmp/memif1.sock",
		},
		Mtu:         1500,
		IpAddresses: []string{"10.0.0.11/24"},
	}

	// Memif100012 is an example of a memory interface configuration.
	Memif100012 = interfaces.Interfaces_Interface{
		Name:    "memif100012",
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Master:         true,
			SocketFilename: "/tmp/memif1.sock",
		},
		Mtu:         1500,
		IpAddresses: []string{"10.0.0.12/24"},
	}

	// Memif100013 is an example of a memory interface configuration.
	Memif100013 = interfaces.Interfaces_Interface{
		Name:    "memif100013",
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Master:         true,
			SocketFilename: "/tmp/memif1.sock",
		},
		Mtu:         1500,
		IpAddresses: []string{"10.0.0.13/24"},
	}

	// VxlanVni5 is an example of a memory interface configuration.
	VxlanVni5 = interfaces.Interfaces_Interface{
		Name:    "VxlanVni5",
		Type:    interfaces.InterfaceType_VXLAN_TUNNEL,
		Enabled: true,
		Vxlan: &interfaces.Interfaces_Interface_Vxlan{
			SrcAddress: "192.168.1.1",
			DstAddress: "192.168.1.2",
			Vni:        5,
		},
	}

	// AfPacketVeth1 is an example of a memory interface configuration.
	AfPacketVeth1 = interfaces.Interfaces_Interface{
		Name:    "AfPacketVeth1",
		Type:    interfaces.InterfaceType_AF_PACKET_INTERFACE,
		Enabled: true,
		Afpacket: &interfaces.Interfaces_Interface_Afpacket{
			HostIfName: "veth1",
		},
	}

	// BDMemif100011ToMemif100012 is an example of a bridge domain configuration.
	BDMemif100011ToMemif100012 = l2.BridgeDomains_BridgeDomain{
		Name:                "aaa",
		Flood:               false,
		UnknownUnicastFlood: false,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /*means disable aging*/
	}
)

// TapInterfaceBuilder serves to create a test interface.
func TapInterfaceBuilder(name string, ip string) interfaces.Interfaces_Interface {
	return interfaces.Interfaces_Interface{
		Name:        name,
		Type:        interfaces.InterfaceType_TAP_INTERFACE,
		Enabled:     true,
		Mtu:         1500,
		IpAddresses: []string{ip},
		Tap:         &interfaces.Interfaces_Interface_Tap{HostIfName: name},
	}
}

// MemifBuilder creates a new instance for testing purposes.
func MemifBuilder(ifname string, ipAddr string, master bool, id uint32) *interfaces.Interfaces_Interface {
	return &interfaces.Interfaces_Interface{
		Name:    ifname,
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Id:             id,
			Master:         master,
			SocketFilename: "/tmp/" + ifname + ".sock",
		},
		IpAddresses: []string{ipAddr},
	}

}

// LoopbackBuilder creates a new instance for testing purposes.
func LoopbackBuilder(ifname string, ipAddr string) *interfaces.Interfaces_Interface {
	return &interfaces.Interfaces_Interface{
		Name:        ifname,
		Type:        interfaces.InterfaceType_SOFTWARE_LOOPBACK,
		Enabled:     true,
		IpAddresses: []string{ipAddr},
	}

}

// BfdSessionBuilder creates BFD session without authentication.
func BfdSessionBuilder(iface string, srcAddr string, dstAddr string, desInt uint32, reqInt uint32, multiplier uint32) bfd.SingleHopBFD_Session {
	return bfd.SingleHopBFD_Session{
		Interface:             iface,
		SourceAddress:         srcAddr,
		DestinationAddress:    dstAddr,
		DesiredMinTxInterval:  desInt,
		RequiredMinRxInterval: reqInt,
		DetectMultiplier:      multiplier,
	}
}

// BfdAuthSessionBuilder creates BFD session including authentication.
func BfdAuthSessionBuilder(iface string, srcAddr string, dstAddr string, desInt uint32, reqInt uint32, multiplier uint32) bfd.SingleHopBFD_Session {
	return bfd.SingleHopBFD_Session{
		Interface:             iface,
		SourceAddress:         srcAddr,
		DestinationAddress:    dstAddr,
		DesiredMinTxInterval:  desInt,
		RequiredMinRxInterval: reqInt,
		DetectMultiplier:      multiplier,
		Authentication: &bfd.SingleHopBFD_Session_Authentication{
			KeyId:           1,
			AdvertisedKeyId: 1,
		},
	}
}

// BfdAuthKeyBuilder creates BFD authentication key.
func BfdAuthKeyBuilder(id uint32, authType bfd.SingleHopBFD_Key_AuthenticationType, secret string) bfd.SingleHopBFD_Key {
	return bfd.SingleHopBFD_Key{
		Id:                 id,
		AuthenticationType: authType,
		Secret:             secret,
	}
}

// BfdEchoFunctionBuilder builds BFD echo source.
func BfdEchoFunctionBuilder(iface string) bfd.SingleHopBFD_EchoFunction {
	return bfd.SingleHopBFD_EchoFunction{
		EchoSourceInterface: iface,
	}
}
