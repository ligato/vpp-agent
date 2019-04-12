package cmd_generator

import (
	"github.com/gogo/protobuf/proto"

	acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	nat "github.com/ligato/vpp-agent/api/models/vpp/nat"

	linterface "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	ll3 "github.com/ligato/vpp-agent/api/models/linux/l3"
	namespace "github.com/ligato/vpp-agent/api/models/linux/namespace"
)

type CommandType int

const (
	// VPP ACL
	ACL CommandType = iota

	// VPP Interfaces
	Interface

	// VPP l2 plugin
	Bd
	Fib
	IPScanNeighbor

	// VPP NAT
	NatGlobal
	NatDNat

	// VPP IP security
	IPSecPolicy
	IPSecAssociation

	// VPP L3 plugin
	Arps
	Routes
	PArp

	// Linux Dumps
	LinuxInterface
	LinuxARPs
	LinuxRoutes
)

func GenerateConfig(cmdType CommandType) (msg proto.Message) {

	switch cmdType {
	case ACL:
		msg = generateACLConfig()
	case Arps:
		msg = generateARPConfig()
	case Bd:
		msg = generateBridgeDomainConfig()
	case Fib:
		msg = generateFibConfig()
	case Interface:
		msg = generateInterfaceConfig()
	case Routes:
		msg = generateRouteConfig()
	case PArp:
		msg = generateProxyARPConfig()
	case IPScanNeighbor:
		msg = generateIPscanneighConfig()
	case NatGlobal:
		msg = generateDNATConfig()
	case NatDNat:
		msg = generateNATConfig()
	case IPSecPolicy:
		msg = generateIPSecPolicyConfig()
	case IPSecAssociation:
		msg = generateIPSecAssociationConfig()
	case LinuxInterface:
		msg = generateLinuxInterfaceConfig()
	case LinuxARPs:
		msg = generateARPConfig()
	case LinuxRoutes:
		msg = generateLinuxRouteConfig()
	}

	return msg
}

func generateFibConfig() proto.Message {
	fib := &l2.FIBEntry{
		PhysAddress:             "EA:FE:3C:64:A7:44",
		BridgeDomain:            "bd1",
		OutgoingInterface:       "loop1",
		StaticConfig:            true,
		BridgedVirtualInterface: true,
		Action:                  l2.FIBEntry_FORWARD, // or DROP
	}

	return fib
}

func generateInterfaceConfig() proto.Message {
	//{"name":"loop1","type":"SOFTWARE_LOOPBACK","enabled":true,
	// "phys_address":"7C:4E:E7:8A:63:68","ip_addresses":["192.168.25.3/24","172.125.45.1/24"],
	// "mtu":1478}

	intf := &interfaces.Interface{
		Name:    "loop0",
		Enabled: true,
		IpAddresses: []string{
			"192.168.25.3/24",
			"172.125.45.1/24"},
		Mtu:         1478,
		PhysAddress: "7C:4E:E7:8A:63:68",
		Type:        interfaces.Interface_SOFTWARE_LOOPBACK,
	}

	return intf
}

func generateACLConfig() proto.Message {
	accessList := &acl.ACL{
		Name: "aclip1",
		Rules: []*acl.ACL_Rule{
			// ACL IP rule
			{
				Action: acl.ACL_Rule_PERMIT,
				IpRule: &acl.ACL_Rule_IpRule{
					Ip: &acl.ACL_Rule_IpRule_Ip{
						SourceNetwork:      "192.168.1.1/32",
						DestinationNetwork: "10.20.0.1/24",
					},
				},
			},
			// ACL ICMP rule
			{
				Action: acl.ACL_Rule_PERMIT,
				IpRule: &acl.ACL_Rule_IpRule{
					Icmp: &acl.ACL_Rule_IpRule_Icmp{
						Icmpv6: false,
						IcmpCodeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
							First: 150,
							Last:  250,
						},
						IcmpTypeRange: &acl.ACL_Rule_IpRule_Icmp_Range{
							First: 1150,
							Last:  1250,
						},
					},
				},
			},
			// ACL TCP rule
			{
				Action: acl.ACL_Rule_PERMIT,
				IpRule: &acl.ACL_Rule_IpRule{
					Tcp: &acl.ACL_Rule_IpRule_Tcp{
						TcpFlagsMask:  20,
						TcpFlagsValue: 10,
						SourcePortRange: &acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 150,
							UpperPort: 250,
						},
						DestinationPortRange: &acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 1150,
							UpperPort: 1250,
						},
					},
				},
			},
			// ACL UDP rule
			{
				Action: acl.ACL_Rule_PERMIT,
				IpRule: &acl.ACL_Rule_IpRule{
					Udp: &acl.ACL_Rule_IpRule_Udp{
						SourcePortRange: &acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 150,
							UpperPort: 250,
						},
						DestinationPortRange: &acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 1150,
							UpperPort: 1250,
						},
					},
				},
			},
		},
		Interfaces: &acl.ACL_Interfaces{
			Ingress: []string{"tap1", "tap2"},
			Egress:  []string{"tap1", "tap2"},
		},
	}

	return accessList
}

func generateBridgeDomainConfig() proto.Message {
	bd := &l2.BridgeDomain{
		Name:                "bd1",
		Learn:               true,
		ArpTermination:      true,
		Flood:               true,
		UnknownUnicastFlood: true,
		Forward:             true,
		MacAge:              0,
		Interfaces: []*l2.BridgeDomain_Interface{
			{
				Name:                    "sub1",
				BridgedVirtualInterface: true,
				SplitHorizonGroup:       0,
			},
			{
				Name:                    "tap1",
				BridgedVirtualInterface: false,
				SplitHorizonGroup:       1,
			},
			{
				Name:                    "memif1",
				BridgedVirtualInterface: false,
				SplitHorizonGroup:       2,
			},
		},
		ArpTerminationTable: []*l2.BridgeDomain_ArpTerminationEntry{
			{
				IpAddress:   "192.168.50.20",
				PhysAddress: "A7:5D:44:D8:E6:51",
			},
		},
	}

	return bd
}

func generateARPConfig() proto.Message {

	///vnf-agent/vpp1/config/vpp/v2/arp/tap1/192.168.10.21
	// '{"interface":"tap1","ip_address":"192.168.10.21","phys_address":"59:6C:45:59:8E:BD",
	// "static":true}'

	arp := &l3.ARPEntry{
		Interface:   "tap1",
		IpAddress:   "192.168.10.21",
		PhysAddress: "59:6C:45:59:8E:BD",
		Static:      true,
	}

	return arp
}

func generateRouteConfig() proto.Message {
	route := &l3.Route{
		Type:        l3.Route_INTER_VRF,
		VrfId:       1,
		DstNetwork:  "10.1.1.3/32",
		NextHopAddr: "192.168.1.13",
		ViaVrfId:    0,
	}

	return route
}

func generateProxyARPConfig() proto.Message {
	///vnf-agent/vpp1/config/vpp/v2/proxyarp-global
	// '{"interfaces":[{"name":"tap1"},{"name":"tap2"}],
	// "ranges":[{"first_ip_addr":"10.0.0.1","last_ip_addr":"10.0.0.3"}]}'

	pr := &l3.ProxyARP{
		Interfaces: []*l3.ProxyARP_Interface{
			{
				Name: "tap1",
			},
			{
				Name: "tap2",
			},
		},
		Ranges: []*l3.ProxyARP_Range{
			{
				FirstIpAddr: "10.0.0.1",
				LastIpAddr:  "10.0.0.3",
			},
		},
	}

	return pr
}

func generateIPscanneighConfig() proto.Message {
	///vnf-agent/vpp1/config/vpp/v2/ipscanneigh-global
	// '{"mode":"BOTH","scan_interval":11,"max_proc_time":36,"max_update":5,
	// "scan_int_delay":16,"stale_threshold":26}'

	ips := &l3.IPScanNeighbor{
		Mode:           l3.IPScanNeighbor_BOTH,
		ScanInterval:   11,
		MaxProcTime:    36,
		MaxUpdate:      5,
		ScanIntDelay:   16,
		StaleThreshold: 26,
	}

	return ips
}

func generateDNATConfig() proto.Message {
	dNat := &nat.DNat44{
		Label: "dnat1",
		StMappings: []*nat.DNat44_StaticMapping{
			{
				ExternalInterface: "tap1",
				ExternalIp:        "192.168.0.1",
				ExternalPort:      8989,
				LocalIps: []*nat.DNat44_StaticMapping_LocalIP{
					{
						VrfId:       0,
						LocalIp:     "172.124.0.2",
						LocalPort:   6500,
						Probability: 40,
					},
					{
						VrfId:       0,
						LocalIp:     "172.125.10.5",
						LocalPort:   2300,
						Probability: 40,
					},
				},
				Protocol: 1,
				TwiceNat: nat.DNat44_StaticMapping_ENABLED,
			},
		},
		IdMappings: []*nat.DNat44_IdentityMapping{
			{
				VrfId:     0,
				IpAddress: "10.10.0.1",
				Port:      2525,
				Protocol:  0,
			},
		},
	}

	return dNat
}

func generateNATConfig() proto.Message {
	natGlobal := &nat.Nat44Global{
		Forwarding: false,
		NatInterfaces: []*nat.Nat44Global_Interface{
			{
				Name:          "tap1",
				IsInside:      false,
				OutputFeature: false,
			},
			{
				Name:          "tap2",
				IsInside:      false,
				OutputFeature: true,
			},
			{
				Name:          "tap3",
				IsInside:      true,
				OutputFeature: false,
			},
		},
		AddressPool: []*nat.Nat44Global_Address{
			{
				VrfId:    0,
				Address:  "192.168.0.1",
				TwiceNat: false,
			},
			{
				VrfId:    0,
				Address:  "175.124.0.1",
				TwiceNat: false,
			},
			{
				VrfId:    0,
				Address:  "10.10.0.1",
				TwiceNat: false,
			},
		},
		VirtualReassembly: &nat.VirtualReassembly{
			Timeout:         10,
			MaxReassemblies: 20,
			MaxFragments:    10,
			DropFragments:   true,
		},
	}

	return natGlobal
}

func generateIPSecPolicyConfig() proto.Message {
	///vnf-agent/vpp1/config/vpp/ipsec/v2/spd/1
	// '{"index":"1","interfaces":[{"name":"tap1"}],
	// "policy_entries":[{"priority":10,"is_outbound":false,"remote_addr_start":"10.0.0.1",
	// "remote_addr_stop":"10.0.0.1","local_addr_start":"10.0.0.2",
	// "local_addr_stop":"10.0.0.2","action":3,"sa_index":"20"},
	// {"priority":10,"is_outbound":true,"remote_addr_start":"10.0.0.1",
	// "remote_addr_stop":"10.0.0.1","local_addr_start":"10.0.0.2",
	// "local_addr_stop":"10.0.0.2","action":3,"sa_index":"10"}]}'

	ipsd := &ipsec.SecurityPolicyDatabase{
		Index: "1",
		Interfaces: []*ipsec.SecurityPolicyDatabase_Interface{
			{
				Name: "tap1",
			},
			{
				Name: "loop1",
			},
		},
		PolicyEntries: []*ipsec.SecurityPolicyDatabase_PolicyEntry{
			{
				Priority:        10,
				IsOutbound:      false,
				RemoteAddrStart: "10.0.0.1",
				RemoteAddrStop:  "10.0.0.2",
				LocalAddrStart:  "10.0.0.2",
				LocalAddrStop:   "10.0.0.2",
				Action:          3,
				SaIndex:         "1",
			},
			{
				Priority:        10,
				IsOutbound:      true,
				RemoteAddrStart: "10.0.0.1",
				RemoteAddrStop:  "10.0.0.2",
				LocalAddrStart:  "10.0.0.2",
				LocalAddrStop:   "10.0.0.2",
				Action:          3,
				SaIndex:         "2",
			},
		},
	}

	return ipsd
}

func generateIPSecAssociationConfig() proto.Message {
	///vnf-agent/vpp1/config/vpp/ipsec/v2/sa/1
	// '{"index":"1","spi":1001,"protocol":1,"crypto_alg":1,
	// "crypto_key":"4a506a794f574265564551694d653768","integ_alg":2,
	// "integ_key":"4339314b55523947594d6d3547666b45764e6a58"}'

	ipsa := &ipsec.SecurityAssociation{
		Index:          "1",
		Spi:            1001,
		Protocol:       1,
		CryptoAlg:      1,
		CryptoKey:      "4a506a794f574265564551694d653768",
		IntegAlg:       2,
		IntegKey:       "4339314b55523947594d6d3547666b45764e6a58",
		EnableUdpEncap: true,
	}

	return ipsa
}

func generateLinuxInterfaceConfig() proto.Message {
	///vnf-agent/vpp1/config/linux/interfaces/v2/interface/veth1
	// '{"name":"veth1","type":"VETH","namespace":{"type":"NSID","reference":"ns1"},
	// "enabled":true,"ip_addresses":["192.168.22.1/24","10.0.2.2/24"],
	// "phys_address":"D2:74:8C:12:67:D2","mtu":1500}'

	lint := &linterface.Interface{
		Name: "veth1",
		Type: linterface.Interface_VETH,
		Namespace: &namespace.NetNamespace{
			Type:      namespace.NetNamespace_NSID,
			Reference: "ns1",
		},
		Enabled: true,
		IpAddresses: []string{
			"192.168.22.1/24",
			"10.0.2.2/24"},
		PhysAddress: "D2:74:8C:12:67:D2",
		Mtu:         1500,
	}

	return lint
}

func generateLinuxARPConfig() proto.Message {
	///vnf-agent/vpp1/config/linux/l3/v2/arp/veth1/130.0.0.1
	// '{"interface":"veth1","ip_address":"130.0.0.1","hw_address":"46:06:18:DB:05:3A"}'

	larp := &ll3.ARPEntry{
		Interface: "veth1",
		IpAddress: "130.0.0.1",
		HwAddress: "46:06:18:DB:05:3A",
	}

	return larp
}

func generateLinuxRouteConfig() proto.Message {
	///vnf-agent/vpp1/config/linux/l3/v2/route/10.0.2.0/24/veth1
	// '{"outgoing_interface":"veth1","dst_network":"10.0.2.0/24","metric":100}'

	lroute := &ll3.Route{
		OutgoingInterface: "veth1",
		DstNetwork:        "10.0.2.0/24",
		Metric:            100,
	}

	return lroute
}
