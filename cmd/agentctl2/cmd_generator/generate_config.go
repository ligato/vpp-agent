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
	VPPACL CommandType = iota
	VPPARP
	VPPBridgeDomain
	VPPInterface
	VPPRoute
	VPPProxyARP
	VPPIPScanNeighbor
	VPPDNat
	VPPNat
	VPPIPSecPolicy
	VPPIPSecAssociation
	LinuxInterface
	LinuxARP
	LinuxRoute
)

func GenerateConfig(cmdType CommandType) (msg proto.Message) {

	switch cmdType {
	case VPPACL:
		msg = generateACLConfig()
	case VPPARP:
		msg = generateARPConfig()
	case VPPBridgeDomain:
		msg = generateBridgeDomainConfig()
	case VPPInterface:
		msg = generateInterfaceConfig()
	case VPPRoute:
		msg = generateRouteConfig()
	case VPPProxyARP:
		msg = generateProxyARPConfig()
	case VPPIPScanNeighbor:
		msg = generateIPscanneighConfig()
	case VPPDNat:
		msg = generateDNATConfig()
	case VPPNat:
		msg = generateNATConfig()
	case VPPIPSecPolicy:
		msg = generateIPSecPolicyConfig()
	case VPPIPSecAssociation:
		msg = generateIPSecAssociationConfig()
	case LinuxInterface:
		msg = generateLinuxInterfaceConfig()
	case LinuxARP:
		msg = generateARPConfig()
	case LinuxRoute:
		msg = generateLinuxRouteConfig()
	}

	return msg
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
	///vnf-agent/vpp1/config/vpp/acls/v2/acl/acl1 '
	// {"name":"acl1","interfaces":{"egress":["tap1","tap2"],
	// "ingress":["tap3","tap4"]},
	// "rules":[{"action":1,"ip_rule":{"ip":{"destination_network":"10.20.1.0/24",
	// "source_network":"192.168.1.2/32"},
	// "tcp":{"destination_port_range":{"lower_port":1150,"upper_port":1250},
	// "source_port_range":{"lower_port":150,"upper_port":250},
	// "tcp_flags_mask":20,"tcp_flags_value":10}}}]}'

	eacl := &acl.ACL{
		Name: "acl1",
		Interfaces: &acl.ACL_Interfaces{
			Ingress: []string{"tap1", "tap2"},
			Egress:  []string{"tap1", "tap2"},
		},
		Rules: []*acl.ACL_Rule{
			{

				Action: 1,
				IpRule: &acl.ACL_Rule_IpRule{
					Ip: &acl.ACL_Rule_IpRule_Ip{
						DestinationNetwork: "10.20.1.0/24",
						SourceNetwork:      "192.168.1.2/32",
					},
					Tcp: &acl.ACL_Rule_IpRule_Tcp{
						DestinationPortRange: &acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 1150,
							UpperPort: 1250,
						},
						SourcePortRange: &acl.ACL_Rule_IpRule_PortRange{
							LowerPort: 150,
							UpperPort: 250,
						},
						TcpFlagsMask:  20,
						TcpFlagsValue: 10,
					},
				},
			},
		},
	}

	return eacl
}

func generateBridgeDomainConfig() proto.Message {
	///vnf-agent/vpp1/config/vpp/l2/v2/bridge-domain/bd1
	// '{"name":"bd1","learn":true,"flood":true,
	// "forward":true,"unknown_unicast_flood":true,"arp_termination":true,
	// "interfaces":[{"name":"if1","split_horizon_group":0,"bridged_virtual_interface":true},
	// {"name":"if2","split_horizon_group":0},{"name":"if2","split_horizon_group":0}],
	// "arp_termination_table":[{"ip_address":"192.168.10.10","phys_address":"a7:65:f1:b5:dc:f6"},
	// {"ip_address":"10.10.0.1","phys_address":"59:6C:45:59:8E:BC"}]}'

	br := &l2.BridgeDomain{
		Name:                "bd1",
		Learn:               true,
		Flood:               true,
		Forward:             true,
		UnknownUnicastFlood: true,
		ArpTermination:      true,
		Interfaces: []*l2.BridgeDomain_Interface{
			{
				Name:              "if1",
				SplitHorizonGroup: 0,
			},
		},
		ArpTerminationTable: []*l2.BridgeDomain_ArpTerminationEntry{
			{
				IpAddress:   "192.168.10.10",
				PhysAddress: "a7:65:f1:b5:dc:f6",
			},
			{
				IpAddress:   "10.10.0.1",
				PhysAddress: "59:6C:45:59:8E:BC",
			},
		},
	}

	return br
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
	///vnf-agent/vpp1/config/vpp/v2/route/vrf/0/dst/1.2.3.4/32/gw
	// '{"type":"INTER_VRF","dst_network":"1.2.3.4/32","via_vrf_id":1}'

	route := &l3.Route{
		Type:       l3.Route_INTER_VRF,
		DstNetwork: "1.2.3.4/32",
		ViaVrfId:   1,
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
	///vnf-agent/vpp1/config/vpp/nat/v2/dnat44/dnat1
	// {"label":"dnat1","st_mappings":[{"external_interface":"tap1","external_ip":"192.168.0.1",
	// "external_port":8989,"local_ips":[{"local_ip":"172.124.0.2",
	// "local_port":6500,"probability":40},{"local_ip":"172.125.10.5",
	// "local_port":2300,"probability":40}],"protocol":"UDP","twice_nat":"ENABLED"}],
	// "id_mappings":[{"ip_address":"10.10.0.1","port":2525}]}

	dnat := &nat.DNat44{
		Label: "dnat1",
		StMappings: []*nat.DNat44_StaticMapping{
			{
				ExternalInterface: "tap1",
				ExternalIp:        "192.168.0.1",
				ExternalPort:      8989,
				LocalIps: []*nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     "172.124.0.2",
						LocalPort:   6500,
						Probability: 40,
					},
					{
						LocalPort:   2300,
						Probability: 40,
					},
				},
				Protocol: nat.DNat44_UDP,
				TwiceNat: nat.DNat44_StaticMapping_ENABLED,
			},
		},
		IdMappings: []*nat.DNat44_IdentityMapping{
			{
				IpAddress: "10.10.0.1",
				Port:      2525,
			},
		},
	}

	return dnat
}

func generateNATConfig() proto.Message {
	///vnf-agent/vpp1/config/vpp/nat/v2/nat44-global
	// {"nat_interfaces":[{"name":"tap1"},{"name":"tap2","output_feature":true},
	// {"name":"tap3","is_inside":true}],"address_pool":[{"address":"192.168.0.1"},
	// {"address":"175.124.0.1"},{"address":"10.10.0.1"}],
	// "virtual_reassembly":{"timeout":10,"max_reassemblies":20,
	// "max_fragments":10,"drop_fragments":true}}

	natg := &nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_Interface{
			{
				Name: "tap1",
			},
			{
				Name: "tap2",
			},
			{
				OutputFeature: true,
			},
			{
				Name:     "tap3",
				IsInside: true,
			},
		},
		AddressPool: []*nat.Nat44Global_Address{
			{
				Address: "192.168.0.1",
			},
			{
				Address: "175.124.0.1",
			},
			{
				Address: "10.10.0.1",
			},
		},
		VirtualReassembly: &nat.VirtualReassembly{
			Timeout:         10,
			MaxReassemblies: 20,
			MaxFragments:    10,
			DropFragments:   true,
		},
	}

	return natg
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
				SaIndex:         "20",
			},
			{
				Priority:        10,
				IsOutbound:      true,
				RemoteAddrStart: "10.0.0.1",
				RemoteAddrStop:  "10.0.0.2",
				LocalAddrStart:  "10.0.0.2",
				LocalAddrStop:   "10.0.0.2",
				Action:          3,
				SaIndex:         "10",
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
		Index:     "1",
		Spi:       1001,
		Protocol:  1,
		CryptoAlg: 1,
		CryptoKey: "4a506a794f574265564551694d653768",
		IntegAlg:  2,
		IntegKey:  "4339314b55523947594d6d3547666b45764e6a58",
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
