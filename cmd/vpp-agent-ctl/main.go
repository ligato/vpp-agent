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

// package vpp-agent-ctl implements the vpp-agent-ctl test tool for testing
// VPP Agent plugins. In addition to testing, the vpp-agent-ctl tool can
// be used to demonstrate the usage of VPP Agent plugins and their APIs.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/namsral/flag"

	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l4"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/stn"
	linuxIntf "github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/interfaces"
	linuxL3 "github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/l3"
)

// VppAgentCtl is ctl context
type VppAgentCtl struct {
	log             logging.Logger
	serviceLabel    servicelabel.Plugin
	bytesConnection *etcdv3.BytesConnectionEtcd
	broker          keyval.ProtoBroker
}

func main() {
	var ctl VppAgentCtl
	// Setup logger
	ctl.log = logrus.DefaultLogger()
	ctl.log.SetLevel(logging.InfoLevel)
	// Parse service label
	flag.CommandLine.ParseEnv(os.Environ())
	ctl.serviceLabel.Init()
	// Establish ETCD connection
	ctl.bytesConnection, ctl.broker = ctl.createEtcdClient()

	ctl.do()
}

func (ctl *VppAgentCtl) do() {
	args := os.Args
	argsLen := len(args)
	if argsLen <= 1 {
		// No commands
		ctl.usage()
		return
	}
	switch args[1] {
	case "-list":
		// List all keys
		ctl.listAllAgentKeys()
	case "-dump":
		if argsLen > 2 {
			// Dump specific key
			ctl.etcdDump(args[2])
		} else {
			// Dump all keys
			ctl.etcdDump("")
		}
	case "-get":
		if argsLen > 2 {
			// Get key
			ctl.etcdGet(args[2])
		}
	case "-del":
		if argsLen > 2 {
			// Del key
			ctl.etcdDel(args[2])
		}
	case "-put":
		if argsLen > 3 {
			ctl.etcdPut(args[2], args[3])
		}
	default:
		switch args[1] {
		// ACL
		case "-acl":
			createACL(db)
		case "-dacl":
			delete(db, acl.Key("acl1"))
		case "-cr":
			createRoute(db)
		case "-dr":
			deleteRoute(db, "10.1.1.3/32", "")
		case "-txn":
			txn(db)
		case "-dtxn":
			deleteTxn(db)
		case "-cbd":
			createBridgeDomain(db, bridgeDomain1)
		case "-dbd":
			delete(db, l2.BridgeDomainKey(bridgeDomain1))
		case "-aft":
			addStaticFibTableEntry(db, bridgeDomain1, ifName1)
		case "-dft":
			deleteStaticFibTableEntry(db, bridgeDomain1)
		case "-aae":
			addArpEntry(db, ifName1)
		case "-dae":
			deleteArpEntry(db, ifName1)
		case "-aat":
			addArpTableEntry(db, bridgeDomain1)
		case "-cxc":
			createL2xConnect(db, ifName1, ifName2)
		case "-dxc":
			delete(db, l2.XConnectKey(ifName1))
		case "-caf":
			createAfPacket(db, afpacket1, "lo", "b4:e6:1c:a1:0d:31", "",
				"fdcd:f7fb:995c::/48")
		case "-maf":
			createAfPacket(db, afpacket1, "lo", "41:69:e3:1d:82:81", "192.168.12.1/24",
				"fd21:7408:186f::/48")
		case "-bfds":
			ctl.createBfdSession()
		case "-bfdsd":
			ctl.deleteBfdSession()
		case "-bfdk":
			ctl.createBfdKey()
		case "-bfdkd":
			ctl.deleteBfdKey()
		case "-bfde":
			ctl.createBfdEcho()
		case "-bfded":
			ctl.deleteBfdEcho()
			// VPP interfaces
		case "-eth":
			ctl.createEthernet()
		case "-ethd":
			ctl.deleteEthernet()
		case "-tap":
			ctl.createTap()
		case "-tapd":
			ctl.deleteTap()
		case "-loop":
			ctl.createLoopback()
		case "-loopd":
			ctl.deleteLoopback()
		case "-memif":
			ctl.createMemif()
		case "-memifd":
			ctl.deleteMemif()
		case "-vxlan":
			ctl.createVxlan()
		case "-vxland":
			ctl.deleteVxlan()
		case "-afpkt":
			ctl.createAfPacket()
		case "-afpktd":
			ctl.deleteAfPacket()
			// Linux interfaces
		case "-veth":
			ctl.createVethPair()
		case "-vethd":
			ctl.deleteVethPair()
		case "-ltap":
			ctl.createLinuxTap()
		case "-ltapd":
			ctl.deleteLinuxTap()
			// STN
		case "-stn":
			ctl.createStn()
		case "-stnd":
			ctl.deleteStn()
			// NAT
		case "-gnat":
			ctl.createGlobalNat()
		case "-gnatd":
			ctl.deleteGlobalNat()
		case "-snat":
			ctl.createSNat()
		case "-snatd":
			ctl.deleteSNat()
		case "-dnat":
			ctl.createDNat()
		case "-dnatd":
			ctl.deleteDNat()
			// Bridge domains
		case "-bd":
			ctl.createBridgeDomain()
		case "-bdd":
			ctl.deleteBridgeDomain()
			// FIB
		case "-fib":
			ctl.createFib()
		case "-fibd":
			ctl.deleteFib()
			// L2 xConnect
		case "-xconn":
			ctl.createXConn()
		case "-xconnd":
			ctl.deleteXConn()
			// VPP routes
		case "-route":
			ctl.createRoute()
		case "-routed":
			ctl.deleteRoute()
			// Linux routes
		case "-lrte":
			ctl.createLinuxRoute()
		case "-lrted":
			ctl.deleteLinuxRoute()
			// VPP ARP
		case "-arp":
			ctl.createArp()
		case "-arpd":
			ctl.deleteArp()
			// Linux ARP
		case "-larp":
			ctl.createLinuxArp()
		case "-larpd":
			ctl.deleteLinuxArp()
			// L4 plugin
		case "-el4":
			ctl.enableL4Features()
		case "-dl4":
			ctl.disableL4Features()
		case "-appns":
			ctl.createAppNamespace()
		case "-appnsd":
			ctl.deleteAppNamespace()
			// TXN (transaction)
		case "-txn":
			ctl.createTxn()
		case "-txnd":
			ctl.deleteTxn()
			// Error reporting
		case "-errIf":
			ctl.reportIfaceErrorState()
		case "-errBd":
			ctl.reportBdErrorState()
		default:
			ctl.usage()
		}
	}
}

// Show command info
func (ctl *VppAgentCtl) usage() {
	var buffer bytes.Buffer
	// Crud operation
	buffer.WriteString("\nCrud operations with .json\n\t-put <etc_key> <json-file>\n\t-get <etc_key>\n\t-del <etc_key>\n\t-dump\n\t-list\n\n")
	// Prearranged flags
	buffer.WriteString("Prearranged flags (create, delete):\n")
	// ACL
	buffer.WriteString("\t-acl,\t-acld\t- Access List\n")
	// BFD
	buffer.WriteString("\t-bfds,\t-bfdsd\t- BFD session\n\t-bfdk,\t-bfdkd\t- BFD authentication key\n\t-bfde,\t-bfded\t- BFD echo function\n")
	// Interfaces
	buffer.WriteString("\t-eth,\t-ethd\t- Physical interface\n\t-tap,\t-tapd\t- TAP type interface\n\t-loop,\t-loopd\t- Loop type interface\n")
	buffer.WriteString("\t-memif,\t-memifd\t- Memif type interface\n\t-vxlan,\t-vxland\t- VxLAN type interface\n\t-afpkt,\t-afpktd\t- af_packet type interface\n")
	// Linux interfaces
	buffer.WriteString("\t-veth,\t-vethd\t- Linux VETH interface pair\n\t-ltap,\t-ltapd\t- Linux TAP interface\n")
	// STN
	buffer.WriteString("\t-stn,\t-stnd\t- STN rule\n")
	// NAT
	buffer.WriteString("\t-gnat,\t-gnatd\t- Global NAT configuration\n\t-snat,\t-snatd\t- SNAT configuration\n\t-dnat,\t-dnatd\t- DNAT configuration\n")
	// L2
	buffer.WriteString("\t-bd,\t-bdd\t- Bridge doamin\n\t-fib,\t-fibd\t- L2 FIB\n\t-xconn,\t-xconnd\t- L2 X-Connect\n")
	// L3
	buffer.WriteString("\t-route,\t-routed\t- L3 route\n\t-arp,\t-arpd\t- ARP entry\n")
	// Linux L3
	buffer.WriteString("\t-lrte,\t-lrted\t- Linux route\n\t-larp,\t-larpd\t- Linux ARP entry\n")
	// L4
	buffer.WriteString("\t-el4,\t-dl4\t- L4 features\n")
	buffer.WriteString("\t-appns,\t-appnsd\t- Application namespace\n\n")
	// Other
	buffer.WriteString("Other:\n\t-txn,\t-txnd\t- Transaction\n\t-errIf\t\t- Interface error state report\n\t-errBd\t\t- Bridge domain error state report\n")
	ctl.log.Print(buffer.String())
}

// Access lists

func (ctl *VppAgentCtl) createACL() {
	accessList := acl.AccessLists{
		Acl: []*acl.AccessLists_Acl{
			// Single ACL entry
			{
				AclName: "acl1",
				// ACL rules
				Rules: []*acl.AccessLists_Acl_Rule{
					// ACL IP rule
					{
						Actions: &acl.AccessLists_Acl_Rule_Actions{
							AclAction: acl.AclAction_PERMIT,
						},
						Matches: &acl.AccessLists_Acl_Rule_Matches{
							IpRule: &acl.AccessLists_Acl_Rule_Matches_IpRule{
								Ip: &acl.AccessLists_Acl_Rule_Matches_IpRule_Ip{
									SourceNetwork:      "192.168.1.1/32",
									DestinationNetwork: "10.20.0.1/24",
								},
							},
						},
					},
					// ACL ICMP rule
					{
						Actions: &acl.AccessLists_Acl_Rule_Actions{
							AclAction: acl.AclAction_PERMIT,
						},
						Matches: &acl.AccessLists_Acl_Rule_Matches{
							IpRule: &acl.AccessLists_Acl_Rule_Matches_IpRule{
								Icmp: &acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp{
									Icmpv6: false,
									IcmpCodeRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp_IcmpCodeRange{
										First: 150,
										Last:  250,
									},
									IcmpTypeRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp_IcmpTypeRange{
										First: 1150,
										Last:  1250,
									},
								},
							},
						},
					},
					// ACL TCP rule
					{
						Actions: &acl.AccessLists_Acl_Rule_Actions{
							AclAction: acl.AclAction_PERMIT,
						},
						Matches: &acl.AccessLists_Acl_Rule_Matches{
							IpRule: &acl.AccessLists_Acl_Rule_Matches_IpRule{
								Tcp: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp{
									TcpFlagsMask:  20,
									TcpFlagsValue: 10,
									SourcePortRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_SourcePortRange{
										LowerPort: 150,
										UpperPort: 250,
									},
									DestinationPortRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_DestinationPortRange{
										LowerPort: 1150,
										UpperPort: 1250,
									},
								},
							},
						},
					},
					// ACL UDP rule
					{
						Actions: &acl.AccessLists_Acl_Rule_Actions{
							AclAction: acl.AclAction_PERMIT,
						},
						Matches: &acl.AccessLists_Acl_Rule_Matches{
							IpRule: &acl.AccessLists_Acl_Rule_Matches_IpRule{
								Udp: &acl.AccessLists_Acl_Rule_Matches_IpRule_Udp{
									SourcePortRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Udp_SourcePortRange{
										LowerPort: 150,
										UpperPort: 250,
									},
									DestinationPortRange: &acl.AccessLists_Acl_Rule_Matches_IpRule_Udp_DestinationPortRange{
										LowerPort: 1150,
										UpperPort: 1250,
									},
								},
							},
						},
					},
					// ACL other rule
					{
						Actions: &acl.AccessLists_Acl_Rule_Actions{
							AclAction: acl.AclAction_PERMIT,
						},
						Matches: &acl.AccessLists_Acl_Rule_Matches{
							IpRule: &acl.AccessLists_Acl_Rule_Matches_IpRule{
								Other: &acl.AccessLists_Acl_Rule_Matches_IpRule_Other{
									Protocol: 0,
								},
							},
						},
					},
					// ACL MAC IP rule. Note: do not combine ACL ip and mac ip rules in single acl
					//{
					//	Actions: &acl.AccessLists_Acl_Rule_Actions{
					//		AclAction: acl.AclAction_PERMIT,
					//	},
					//	Matches: &acl.AccessLists_Acl_Rule_Matches{
					//		MacipRule: &acl.AccessLists_Acl_Rule_Matches_MacIpRule{
					//			SourceAddress: "192.168.0.1",
					//			SourceAddressPrefix: uint32(16),
					//			SourceMacAddress: "11:44:0A:B8:4A:35",
					//			SourceMacAddressMask: "ff:ff:ff:ff:00:00",
					//		},
					//	},
					//},
				},
				// Interfaces
				Interfaces: &acl.AccessLists_Acl_Interfaces{
					Ingress: []string{"tap1", "tap2"},
					Egress:  []string{"tap1", "tap2"},
				},
			},
		},
	}

	ctl.log.Print(accessList.Acl[0])
	ctl.broker.Put(acl.Key(accessList.Acl[0].AclName), accessList.Acl[0])
}

func (ctl *VppAgentCtl) deleteACL() {
	aclKey := acl.Key("acl1")

	ctl.log.Println("Deleting", aclKey)
	ctl.broker.Delete(aclKey)
}

// Bidirectional forwarding detection

func (ctl *VppAgentCtl) createBfdSession() {
	session := bfd.SingleHopBFD{
		Sessions: []*bfd.SingleHopBFD_Session{
			{
				Interface:             "memif1",
				Enabled:               true,
				SourceAddress:         "192.168.1.2",
				DestinationAddress:    "20.10.0.5",
				RequiredMinRxInterval: 8,
				DesiredMinTxInterval:  3,
				DetectMultiplier:      9,
				Authentication: &bfd.SingleHopBFD_Session_Authentication{
					KeyId:           1,
					AdvertisedKeyId: 1,
				},
			},
		},
	}

	ctl.log.Println(session)
	ctl.broker.Put(bfd.SessionKey(session.Sessions[0].Interface), session.Sessions[0])
}

func (ctl *VppAgentCtl) deleteBfdSession() {
	sessionKey := bfd.SessionKey("memif1")

	ctl.log.Println("Deleting", sessionKey)
	ctl.broker.Delete(sessionKey)
}

func (ctl *VppAgentCtl) createBfdKey() {
	authKey := bfd.SingleHopBFD{
		Keys: []*bfd.SingleHopBFD_Key{
			{
				Id:                 1,
				AuthenticationType: bfd.SingleHopBFD_Key_METICULOUS_KEYED_SHA1, // or Keyed sha1
				Secret:             "1981491891941891",
			},
		},
	}

	ctl.log.Println(authKey)
	ctl.broker.Put(bfd.AuthKeysKey(string(authKey.Keys[0].Id)), authKey.Keys[0])
}

func (ctl *VppAgentCtl) deleteBfdKey() {
	bfdAuthKeyKey := bfd.AuthKeysKey(string(1))

	ctl.log.Println("Deleting", bfdAuthKeyKey)
	ctl.broker.Delete(bfdAuthKeyKey)
}

func (ctl *VppAgentCtl) createBfdEcho() {
	echoFunction := bfd.SingleHopBFD{
		EchoFunction: &bfd.SingleHopBFD_EchoFunction{
			EchoSourceInterface: "memif1",
		},
	}

	ctl.log.Println(echoFunction)
	ctl.broker.Put(bfd.EchoFunctionKey("memif1"), echoFunction.EchoFunction)
}

func (ctl *VppAgentCtl) deleteBfdEcho() {
	echoFunctionKey := bfd.EchoFunctionKey("memif1")

	ctl.log.Println("Deleting", echoFunctionKey)
	ctl.broker.Delete(echoFunctionKey)
}

// VPP interfaces

func (ctl *VppAgentCtl) createEthernet() {
	ethernet := &interfaces.Interfaces{
		Interface: []*interfaces.Interfaces_Interface{
			{
				Name:    "GigabitEthernet0/8/0",
				Type:    interfaces.InterfaceType_ETHERNET_CSMACD,
				Enabled: true,
				IpAddresses: []string{
					"192.168.1.1",
					"2001:db8:0:0:0:ff00:5168:2bc8/48",
				},
				//Unnumbered: &interfaces.Interfaces_Interface_Unnumbered{
				//	IsUnnumbered: true,
				//	InterfaceWithIP: "memif1",
				//},
			},
		},
	}

	ctl.log.Println(ethernet)
	ctl.broker.Put(interfaces.InterfaceKey(ethernet.Interface[0].Name), ethernet.Interface[0])
}

func (ctl *VppAgentCtl) deleteEthernet() {
	ethernetKey := interfaces.InterfaceKey("GigabitEthernet0/8/0")

	ctl.log.Println("Deleting", ethernetKey)
	ctl.broker.Delete(ethernetKey)
}

func (ctl *VppAgentCtl) createTap() {
	tap := &interfaces.Interfaces{
		Interface: []*interfaces.Interfaces_Interface{
			{
				Name:        "tap1",
				Type:        interfaces.InterfaceType_TAP_INTERFACE,
				Enabled:     true,
				PhysAddress: "12:E4:0E:D5:BC:DC",
				IpAddresses: []string{
					"192.168.20.3/24",
				},
				//Unnumbered: &interfaces.Interfaces_Interface_Unnumbered{
				//	IsUnnumbered: true,
				//	InterfaceWithIP: "memif1",
				//},
				Tap: &interfaces.Interfaces_Interface_Tap{
					HostIfName: "tap1",
				},
			},
		},
	}

	ctl.log.Println(tap)
	ctl.broker.Put(interfaces.InterfaceKey(tap.Interface[0].Name), tap.Interface[0])
}

func (ctl *VppAgentCtl) deleteTap() {
	tapKey := interfaces.InterfaceKey("tap1")

	ctl.log.Println("Deleting", tapKey)
	ctl.broker.Delete(tapKey)
}

func (ctl *VppAgentCtl) createLoopback() {
	loopback := &interfaces.Interfaces{
		Interface: []*interfaces.Interfaces_Interface{
			{
				Name:        "loop1",
				Type:        interfaces.InterfaceType_SOFTWARE_LOOPBACK,
				Enabled:     true,
				PhysAddress: "7C:4E:E7:8A:63:68",
				Mtu:         1478,
				IpAddresses: []string{
					"192.168.20.3/24",
					"172.125.40.1/24",
				},
				//Unnumbered: &interfaces.Interfaces_Interface_Unnumbered{
				//	IsUnnumbered: true,
				//	InterfaceWithIP: "memif1",
				//},
			},
		},
	}

	ctl.log.Println(loopback)
	ctl.broker.Put(interfaces.InterfaceKey(loopback.Interface[0].Name), loopback.Interface[0])
}

func (ctl *VppAgentCtl) deleteLoopback() {
	loopbackKey := interfaces.InterfaceKey("loop1")

	ctl.log.Println("Deleting", loopbackKey)
	ctl.broker.Delete(loopbackKey)
}

func (ctl *VppAgentCtl) createMemif() {
	memif := &interfaces.Interfaces{
		Interface: []*interfaces.Interfaces_Interface{
			{
				Name:        "memif1",
				Type:        interfaces.InterfaceType_MEMORY_INTERFACE,
				Enabled:     true,
				PhysAddress: "4E:93:2A:38:A7:77",
				Mtu:         1478,
				IpAddresses: []string{
					"172.125.40.1/24",
				},
				//Unnumbered: &interfaces.Interfaces_Interface_Unnumbered{
				//	IsUnnumbered: true,
				//	InterfaceWithIP: "memif1",
				//},
				Memif: &interfaces.Interfaces_Interface_Memif{
					Id:             1,
					Secret:         "secret",
					Master:         true,
					SocketFilename: "/tmp/memif1.sock",
				},
			},
		},
	}

	ctl.log.Println(memif)
	ctl.broker.Put(interfaces.InterfaceKey(memif.Interface[0].Name), memif.Interface[0])
}

func (ctl *VppAgentCtl) deleteMemif() {
	memifKey := interfaces.InterfaceKey("memif1")

	ctl.log.Println("Deleting", memifKey)
	ctl.broker.Delete(memifKey)
}

func (ctl *VppAgentCtl) createVxlan() {
	vxlan := &interfaces.Interfaces{
		Interface: []*interfaces.Interfaces_Interface{
			{
				Name:        "vxlan1",
				Type:        interfaces.InterfaceType_VXLAN_TUNNEL,
				Enabled:     true,
				PhysAddress: "09:8E:3A:47:DD:F9",
				Mtu:         1478,
				IpAddresses: []string{
					"172.125.40.1/24",
				},
				//Unnumbered: &interfaces.Interfaces_Interface_Unnumbered{
				//	IsUnnumbered: true,
				//	InterfaceWithIP: "memif1",
				//},
				Vxlan: &interfaces.Interfaces_Interface_Vxlan{
					SrcAddress: "192.168.42.1",
					DstAddress: "192.168.42.2",
					Vni:        13,
				},
			},
		},
	}

	ctl.log.Println(vxlan)
	ctl.broker.Put(interfaces.InterfaceKey(vxlan.Interface[0].Name), vxlan.Interface[0])
}

func (ctl *VppAgentCtl) deleteVxlan() {
	vxlanKey := interfaces.InterfaceKey("vxlan1")

	ctl.log.Println("Deleting", vxlanKey)
	ctl.broker.Delete(vxlanKey)
}

func (ctl *VppAgentCtl) createAfPacket() {
	ifs := interfaces.Interfaces{
		Interface: []*interfaces.Interfaces_Interface{
			{
				Name:    "afpacket1",
				Type:    interfaces.InterfaceType_AF_PACKET_INTERFACE,
				Enabled: true,
				Mtu:     1500,
				IpAddresses: []string{
					"172.125.40.1/24",
					"192.168.12.1/24",
					"fd21:7408:186f::/48",
				},
				//Unnumbered: &interfaces.Interfaces_Interface_Unnumbered{
				//	IsUnnumbered: true,
				//	InterfaceWithIP: "memif1",
				//},
				Afpacket: &interfaces.Interfaces_Interface_Afpacket{
					HostIfName: "lo",
				},
			},
		},
	}

	ctl.log.Println(ifs)
	ctl.broker.Put(interfaces.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])
}

func (ctl *VppAgentCtl) deleteAfPacket() {
	afPacketKey := interfaces.InterfaceKey("afpacket1")

	ctl.log.Println("Deleting", afPacketKey)
	ctl.broker.Delete(afPacketKey)
}

// Linux interfaces

func (ctl *VppAgentCtl) createVethPair() {
	// Note: VETH interfaces are created in pairs
	veths := linuxIntf.LinuxInterfaces{
		Interface: []*linuxIntf.LinuxInterfaces_Interface{
			{
				Name:        "veth1",
				Type:        linuxIntf.LinuxInterfaces_VETH,
				Enabled:     true,
				PhysAddress: "5D:5A:15:EE:D1:9F",
				Namespace: &linuxIntf.LinuxInterfaces_Interface_Namespace{
					Name: "ns1",
					Type: linuxIntf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
				},
				Mtu: 1500,
				IpAddresses: []string{
					"192.168.22.1/24",
				},
				Veth: &linuxIntf.LinuxInterfaces_Interface_Veth{
					PeerIfName: "veth2",
				},
			},
			{
				Name:        "veth2",
				Type:        linuxIntf.LinuxInterfaces_VETH,
				Enabled:     true,
				PhysAddress: "F1:E8:5F:62:B7:99",
				Namespace: &linuxIntf.LinuxInterfaces_Interface_Namespace{
					Name: "ns2",
					Type: linuxIntf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
				},
				Mtu: 1500,
				IpAddresses: []string{
					"192.168.22.5/24",
				},
				Veth: &linuxIntf.LinuxInterfaces_Interface_Veth{
					PeerIfName: "veth1",
				},
			},
		},
	}

	ctl.log.Println(veths)
	ctl.broker.Put(linuxIntf.InterfaceKey(veths.Interface[0].Name), veths.Interface[0])
	ctl.broker.Put(linuxIntf.InterfaceKey(veths.Interface[1].Name), veths.Interface[1])
}

func (ctl *VppAgentCtl) deleteVethPair() {
	veth1Key := linuxIntf.InterfaceKey("veth1")
	veth2Key := linuxIntf.InterfaceKey("veth2")

	ctl.log.Println("Deleting", veth1Key)
	ctl.broker.Delete(veth1Key)
	ctl.log.Println("Deleting", veth2Key)
	ctl.broker.Delete(veth2Key)
}

func (ctl *VppAgentCtl) createLinuxTap() {
	linuxTap := linuxIntf.LinuxInterfaces{
		Interface: []*linuxIntf.LinuxInterfaces_Interface{
			{
				Name:        "tap1",
				HostIfName:  "tap-host",
				Type:        linuxIntf.LinuxInterfaces_AUTO_TAP,
				Enabled:     true,
				PhysAddress: "BC:FE:E9:5E:07:04",
				Namespace: &linuxIntf.LinuxInterfaces_Interface_Namespace{
					Name: "ns1",
					Type: linuxIntf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
				},
				Mtu: 1500,
				IpAddresses: []string{
					"172.52.45.127/24",
				},
			},
		},
	}

	ctl.log.Println(linuxTap)
	ctl.broker.Put(linuxIntf.InterfaceKey(linuxTap.Interface[0].Name), linuxTap.Interface[0])
}

func (ctl *VppAgentCtl) deleteLinuxTap() {
	linuxTapKey := linuxIntf.InterfaceKey("tap1")

	ctl.log.Println("Deleting", linuxTapKey)
	ctl.broker.Delete(linuxTapKey)
}

// STN

func (ctl *VppAgentCtl) createStn() {
	stnRule := stn.StnRule{
		RuleName:  "rule1",
		IpAddress: "192.168.50.12",
		Interface: "memif1",
	}

	ctl.log.Println(stnRule)
	ctl.broker.Put(stn.Key(stnRule.RuleName), &stnRule)
}

func (ctl *VppAgentCtl) deleteStn() {
	stnRuleKey := stn.Key("rule1")

	ctl.log.Println("Deleting", stnRuleKey)
	ctl.broker.Delete(stnRuleKey)
}

// Network address translation

func (ctl *VppAgentCtl) createGlobalNat() {
	natGlobal := &nat.Nat44Global{
		Forwarding: false,
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
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
		AddressPools: []*nat.Nat44Global_AddressPools{
			{
				VrfId:           0,
				FirstSrcAddress: "192.168.0.1",
				TwiceNat:        false,
			},
			{
				VrfId:           0,
				FirstSrcAddress: "175.124.0.1",
				LastSrcAddress:  "175.124.0.3",
				TwiceNat:        false,
			},
			{
				VrfId:           0,
				FirstSrcAddress: "10.10.0.1",
				LastSrcAddress:  "10.10.0.2",
				TwiceNat:        false,
			},
		},
	}

	ctl.log.Println(natGlobal)
	ctl.broker.Put(nat.GlobalConfigKey(), natGlobal)
}

func (ctl *VppAgentCtl) deleteGlobalNat() {
	globalNat := nat.GlobalConfigKey()

	ctl.log.Println("Deleting", globalNat)
	ctl.broker.Delete(globalNat)
}

func (ctl *VppAgentCtl) createSNat() {
	// Note: SNAT not implemented
	sNat := &nat.Nat44SNat_SNatConfig{
		Label: "snat1",
	}

	ctl.log.Println(sNat)
	ctl.broker.Put(nat.SNatKey(sNat.Label), sNat)
}

func (ctl *VppAgentCtl) deleteSNat() {
	sNat := nat.SNatKey("snat1")

	ctl.log.Println("Deleting", sNat)
	ctl.broker.Delete(sNat)
}

func (ctl *VppAgentCtl) createDNat() {
	// DNat config
	dNat := &nat.Nat44DNat_DNatConfig{
		Label: "dnat1",
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:             0,
				ExternalInterface: "tap1",
				ExternalIP:        "192.168.0.1",
				ExternalPort:      8989,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     "172.124.0.2",
						LocalPort:   6500,
						Probability: 40,
					},
					{
						LocalIP:     "172.125.10.5",
						LocalPort:   2300,
						Probability: 40,
					},
				},
				Protocol: 1,
				TwiceNat: false,
			},
		},
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:     0,
				IpAddress: "10.10.0.1",
				Port:      2525,
				Protocol:  0,
			},
		},
	}

	ctl.log.Println(dNat)
	ctl.broker.Put(nat.DNatKey(dNat.Label), dNat)
}

func (ctl *VppAgentCtl) deleteDNat() {
	dNat := nat.DNatKey("dnat1")

	ctl.log.Println("Deleting", dNat)
	ctl.broker.Delete(dNat)
}

// Bridge domains

func (ctl *VppAgentCtl) createBridgeDomain() {
	bd := l2.BridgeDomains{
		BridgeDomains: []*l2.BridgeDomains_BridgeDomain{
			{
				Name:                "bd1",
				Learn:               true,
				ArpTermination:      true,
				Flood:               true,
				UnknownUnicastFlood: true,
				Forward:             true,
				MacAge:              0,
				Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
					{
						Name: "loop1",
						BridgedVirtualInterface: true,
						SplitHorizonGroup:       0,
					},
					{
						Name: "tap1",
						BridgedVirtualInterface: false,
						SplitHorizonGroup:       0,
					},
					{
						Name: "memif1",
						BridgedVirtualInterface: false,
						SplitHorizonGroup:       0,
					},
				},
				ArpTerminationTable: []*l2.BridgeDomains_BridgeDomain_ArpTerminationTable{
					{
						IpAddress:   "192.168.50.20",
						PhysAddress: "A7:5D:44:D8:E6:51",
					},
				},
			},
		},
	}

	ctl.log.Println(bd)
	ctl.broker.Put(l2.BridgeDomainKey(bd.BridgeDomains[0].Name), bd.BridgeDomains[0])
}

func (ctl *VppAgentCtl) deleteBridgeDomain() {
	bdKey := l2.BridgeDomainKey("bd1")

	ctl.log.Println("Deleting", bdKey)
	ctl.broker.Delete(bdKey)
}

// FIB

func (ctl *VppAgentCtl) createFib() {
	fib := l2.FibTableEntries{
		FibTableEntry: []*l2.FibTableEntries_FibTableEntry{
			{
				PhysAddress:             "34:EA:FE:3C:64:A7",
				BridgeDomain:            "bd1",
				OutgoingInterface:       "loop1",
				StaticConfig:            true,
				BridgedVirtualInterface: true,
				Action:                  l2.FibTableEntries_FibTableEntry_FORWARD, // or DROP
			},
		},
	}

	ctl.log.Println(fib)
	ctl.broker.Put(l2.FibKey(fib.FibTableEntry[0].BridgeDomain, fib.FibTableEntry[0].PhysAddress), fib.FibTableEntry[0])
}

func (ctl *VppAgentCtl) deleteFib() {
	fibKey := l2.FibKey("bd1", "34:EA:FE:3C:64:A7")

	ctl.log.Println("Deleting", fibKey)
	ctl.broker.Delete(fibKey)
}

// L2 xConnect

func (ctl *VppAgentCtl) createXConn() {
	xc := l2.XConnectPairs{
		XConnectPairs: []*l2.XConnectPairs_XConnectPair{
			{
				ReceiveInterface:  "loop1",
				TransmitInterface: "tap1",
			},
		},
	}

	ctl.log.Println(xc)
	ctl.broker.Put(l2.XConnectKey(xc.XConnectPairs[0].ReceiveInterface), xc.XConnectPairs[0])
}

func (ctl *VppAgentCtl) deleteXConn() {
	xcKey := l2.XConnectKey("loop1")

	ctl.log.Println("Deleting", xcKey)
	ctl.broker.Delete(xcKey)
}

// VPP routes

func (ctl *VppAgentCtl) createRoute() {
	routes := l3.StaticRoutes{
		Route: []*l3.StaticRoutes_Route{
			{
				VrfId:             0,
				DstIpAddr:         "10.1.1.3/32",
				NextHopAddr:       "192.168.1.13",
				Weight:            6,
				OutgoingInterface: "tap1",
			},
		},
	}

	ctl.log.Print(routes.Route[0])
	ctl.broker.Put(l3.RouteKey(routes.Route[0].VrfId, routes.Route[0].DstIpAddr, routes.Route[0].NextHopAddr), routes.Route[0])
}

func (ctl *VppAgentCtl) deleteRoute() {
	routeKey := l3.RouteKey(0, "10.1.1.3/32", "192.168.1.13")

	ctl.log.Println("Deleting", routeKey)
	ctl.broker.Delete(routeKey)
}

// Linux routes

func (ctl *VppAgentCtl) createLinuxRoute() {
	linuxRoutes := linuxL3.LinuxStaticRoutes{
		Route: []*linuxL3.LinuxStaticRoutes_Route{
			// Static route
			{
				Name:      "route1",
				DstIpAddr: "10.0.2.0/24",
				Interface: "veth1",
				Metric:    100,
				Namespace: &linuxL3.LinuxStaticRoutes_Route_Namespace{
					Name: "ns1",
					Type: linuxL3.LinuxStaticRoutes_Route_Namespace_NAMED_NS,
				},
			},
			// Default route
			{
				Name:      "defRoute",
				Default:   true,
				GwAddr:    "10.0.2.2",
				Interface: "veth1",
				Metric:    100,
				Namespace: &linuxL3.LinuxStaticRoutes_Route_Namespace{
					Name: "ns1",
					Type: linuxL3.LinuxStaticRoutes_Route_Namespace_NAMED_NS,
				},
			},
		},
	}

	ctl.log.Println(linuxRoutes)
	ctl.broker.Put(linuxL3.StaticRouteKey(linuxRoutes.Route[0].Name), linuxRoutes.Route[0])
	ctl.broker.Put(linuxL3.StaticRouteKey(linuxRoutes.Route[1].Name), linuxRoutes.Route[1])
}

func (ctl *VppAgentCtl) deleteLinuxRoute() {
	linuxStaticRouteKey := linuxL3.StaticRouteKey("route1")
	linuxDefaultRouteKey := linuxL3.StaticRouteKey("defRoute")

	ctl.log.Println("Deleting", linuxStaticRouteKey)
	ctl.broker.Delete(linuxStaticRouteKey)
	ctl.log.Println("Deleting", linuxDefaultRouteKey)
	ctl.broker.Delete(linuxDefaultRouteKey)
}

// VPP ARP

func (ctl *VppAgentCtl) createArp() {
	arp := l3.ArpTable{
		ArpTableEntries: []*l3.ArpTable_ArpTableEntry{
			{
				Interface:   "tap1",
				IpAddress:   "192.168.10.21",
				PhysAddress: "59:6C:45:59:8E:BD",
				Static:      true,
			},
		},
	}

	ctl.log.Println(arp)
	ctl.broker.Put(l3.ArpEntryKey(arp.ArpTableEntries[0].Interface, arp.ArpTableEntries[0].IpAddress), arp.ArpTableEntries[0])
}

func (ctl *VppAgentCtl) deleteArp() {
	arpKey := l3.ArpEntryKey("tap1", "192.168.10.21")

	ctl.log.Println("Deleting", arpKey)
	ctl.broker.Delete(arpKey)
}

// Linux ARP

func (ctl *VppAgentCtl) createLinuxArp() {
	linuxArp := linuxL3.LinuxStaticArpEntries{
		ArpEntry: []*linuxL3.LinuxStaticArpEntries_ArpEntry{
			{
				Name:      "arp1",
				Interface: "veth1",
				Namespace: &linuxL3.LinuxStaticArpEntries_ArpEntry_Namespace{
					Name: "ns1",
					Type: linuxL3.LinuxStaticArpEntries_ArpEntry_Namespace_NAMED_NS,
				},
				IpAddr:    "130.0.0.1",
				HwAddress: "46:06:18:DB:05:3A",
				State: &linuxL3.LinuxStaticArpEntries_ArpEntry_NudState{
					Type: linuxL3.LinuxStaticArpEntries_ArpEntry_NudState_PERMANENT, // or NOARP, REACHABLE, STALE
				},
				IpFamily: &linuxL3.LinuxStaticArpEntries_ArpEntry_IpFamily{
					Family: linuxL3.LinuxStaticArpEntries_ArpEntry_IpFamily_IPV4, // or IPv6, ALL, MPLS
				},
			},
		},
	}

	ctl.log.Println(linuxArp)
	ctl.broker.Put(linuxL3.StaticArpKey(linuxArp.ArpEntry[0].Name), linuxArp.ArpEntry[0])
}

func (ctl *VppAgentCtl) deleteLinuxArp() {
	linuxArpKey := linuxL3.StaticArpKey("arp1")

	ctl.log.Println("Deleting", linuxArpKey)
	ctl.broker.Delete(linuxArpKey)
}

// L4 plugin

func (ctl *VppAgentCtl) enableL4Features() {
	l4Features := &l4.L4Features{
		Enabled: true,
	}

	ctl.log.Println(l4Features)
	ctl.broker.Put(l4.FeatureKey(), l4Features)
}

func (ctl *VppAgentCtl) disableL4Features() {
	l4Features := &l4.L4Features{
		Enabled: false,
	}

	ctl.log.Println(l4Features)
	ctl.broker.Put(l4.FeatureKey(), l4Features)
}

func (ctl *VppAgentCtl) createAppNamespace() {
	appNs := l4.AppNamespaces{
		AppNamespaces: []*l4.AppNamespaces_AppNamespace{
			{
				NamespaceId: "appns1",
				Secret:      1,
				Interface:   "tap1",
			},
		},
	}

	ctl.log.Println(appNs)
	ctl.broker.Put(l4.AppNamespacesKey(appNs.AppNamespaces[0].NamespaceId), appNs.AppNamespaces[0])
}

func (ctl *VppAgentCtl) deleteAppNamespace() {
	// Note: application namespace cannot be removed, missing API in VPP
	ctl.log.Println("App namespace delete not supported")
}

// TXN transactions

func (ctl *VppAgentCtl) createTxn() {
	ifs := interfaces.Interfaces{
		Interface: []*interfaces.Interfaces_Interface{
			{
				Name:    "tap1",
				Type:    interfaces.InterfaceType_TAP_INTERFACE,
				Enabled: true,
				Mtu:     1500,
				IpAddresses: []string{
					"10.4.4.1/24",
				},
				Tap: &interfaces.Interfaces_Interface_Tap{
					HostIfName: "tap1",
				},
			},
			{
				Name:    "tap2",
				Type:    interfaces.InterfaceType_TAP_INTERFACE,
				Enabled: true,
				Mtu:     1500,
				IpAddresses: []string{
					"10.4.4.2/24",
				},
				Tap: &interfaces.Interfaces_Interface_Tap{
					HostIfName: "tap2",
				},
			},
		},
	}

	bd := l2.BridgeDomains{
		BridgeDomains: []*l2.BridgeDomains_BridgeDomain{
			{
				Name:                "bd1",
				Flood:               false,
				UnknownUnicastFlood: false,
				Forward:             true,
				Learn:               true,
				ArpTermination:      false,
				MacAge:              0,
				Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
					{
						Name: "tap1",
						BridgedVirtualInterface: true,
						SplitHorizonGroup:       0,
					},
					{
						Name: "tap2",
						BridgedVirtualInterface: false,
						SplitHorizonGroup:       0,
					},
				},
			},
		},
	}

	t := ctl.broker.NewTxn()
	t.Put(interfaces.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])
	t.Put(interfaces.InterfaceKey(ifs.Interface[1].Name), ifs.Interface[1])
	t.Put(l2.BridgeDomainKey(bd.BridgeDomains[0].Name), bd.BridgeDomains[0])

	t.Commit()
}

func (ctl *VppAgentCtl) deleteTxn() {
	ctl.log.Println("Deleting txn items")
	ctl.broker.Delete(interfaces.InterfaceKey("tap1"))
	ctl.broker.Delete(interfaces.InterfaceKey("tap2"))
	ctl.broker.Delete(l2.BridgeDomainKey("bd1"))
}

// Error reporting

func (ctl *VppAgentCtl) reportIfaceErrorState() {
	ifErr, err := ctl.broker.ListValues(interfaces.IfErrorPrefix)
	if err != nil {
		ctl.log.Fatal(err)
		return
	}
	for {
		kv, allReceived := ifErr.GetNext()
		if allReceived {
			break
		}
		entry := &interfaces.InterfaceErrors_Interface{}
		err := kv.GetValue(entry)
		if err != nil {
			ctl.log.Fatal(err)
			return
		}
		ctl.log.Println(entry)
	}
}

func (ctl *VppAgentCtl) reportBdErrorState() {
	bdErr, err := ctl.broker.ListValues(l2.BdErrPrefix)
	if err != nil {
		ctl.log.Fatal(err)
		return
	}
	for {
		kv, allReceived := bdErr.GetNext()
		if allReceived {
			break
		}
		entry := &l2.BridgeDomainErrors_BridgeDomain{}
		err := kv.GetValue(entry)
		if err != nil {
			ctl.log.Fatal(err)
			return
		}

		ctl.log.Println(entry)
	}
}

// Auxiliary methods

func (ctl *VppAgentCtl) etcdGet(key string) {
	ctl.log.Debug("GET ", key)

	data, found, _, err := ctl.bytesConnection.GetValue(key)
	if err != nil {
		ctl.log.Error(err)
		return
	}
	if !found {
		ctl.log.Debug("No value found for the key", key)
	}
	ctl.log.Println(string(data))
}

func (ctl *VppAgentCtl) etcdPut(key string, file string) {
	input, err := ctl.readData(file)

	ctl.log.Println("DB putting ", key, " ", string(input))

	err = ctl.bytesConnection.Put(key, input)
	if err != nil {
		ctl.log.Panic("error putting the data ", key, " that to DB from ", file, ", err: ", err)
	}
	ctl.log.Println("DB put successful ", key, " ", file)
}

func (ctl *VppAgentCtl) etcdDel(key string) {
	found, err := ctl.bytesConnection.Delete(key, datasync.WithPrefix())
	if err != nil {
		ctl.log.Error(err)
		return
	}
	if found {
		ctl.log.Debug("Data deleted:", key)
	} else {
		ctl.log.Debug("No value found for the key", key)
	}
}

func (ctl *VppAgentCtl) etcdDump(key string) {
	ctl.log.Debug("DUMP ", key)

	data, err := ctl.bytesConnection.ListValues(key)
	if err != nil {
		ctl.log.Error(err)
		return
	}

	var found bool
	for {
		found = true
		kv, stop := data.GetNext()
		if stop {
			break
		}
		ctl.log.Println(kv.GetKey())
		ctl.log.Println(string(kv.GetValue()))
		ctl.log.Println()

	}
	if !found {
		ctl.log.Debug("No value found for the key", key)
	}
}

func (ctl *VppAgentCtl) readData(file string) ([]byte, error) {
	var input []byte
	var err error

	if file == "-" {
		// read JSON from STDIN
		bio := bufio.NewReader(os.Stdin)
		buf := new(bytes.Buffer)
		buf.ReadFrom(bio)
		input = buf.Bytes()
	} else {
		// read JSON from file
		input, err = ioutil.ReadFile(file)
		if err != nil {
			ctl.log.Panic("error reading the data that needs to be written to DB from ", file, ", err: ", err)
		}
	}

	// validate the JSON
	var js map[string]interface{}
	if json.Unmarshal(input, &js) != nil {
		ctl.log.Panic("Not a valid JSON: ", string(input))
	}
	return input, err
}

func (ctl *VppAgentCtl) listAllAgentKeys() {
	ctl.log.Debug("listAllAgentKeys")

	it, err := ctl.broker.ListKeys(ctl.serviceLabel.GetAllAgentsPrefix())
	if err != nil {
		ctl.log.Error(err)
	}
	for {
		key, _, stop := it.GetNext()
		if stop {
			break
		}
		ctl.log.Println("key: ", key)
	}
}

func (ctl *VppAgentCtl) createEtcdClient() (*etcdv3.BytesConnectionEtcd, keyval.ProtoBroker) {
	var err error
	var configFile string

	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		configFile = os.Args[1]
	} else {
		configFile = os.Getenv("ETCDV3_CONFIG")
	}

	cfg := &etcdv3.Config{}
	if configFile != "" {
		err := config.ParseConfigFromYamlFile(configFile, cfg)
		if err != nil {
			ctl.log.Fatal(err)
		}
	}
	etcdConfig, err := etcdv3.ConfigToClientv3(cfg)
	if err != nil {
		ctl.log.Fatal(err)
	}

	etcdLogger := logrus.NewLogger("etcdLogger")
	etcdLogger.SetLevel(logging.WarnLevel)

	bDB, err := etcdv3.NewEtcdConnectionWithBytes(*etcdConfig, etcdLogger)
	if err != nil {
		ctl.log.Fatal(err)
	}

	return bDB, kvproto.NewProtoWrapperWithSerializer(bDB, &keyval.SerializerJSON{}).
		NewBroker(ctl.serviceLabel.GetAgentPrefix())
}


func addProxyArpIf(db keyval.ProtoBroker) {
	proxyArpIf := l3.ProxyArpInterfaces{
		InterfaceList: []*l3.ProxyArpInterfaces_InterfaceList{
			{
				Label: "proxyArpIf1",
				Interfaces: []*l3.ProxyArpInterfaces_InterfaceList_Interface{
					{
						Name: "tap1",
					},
					{
						Name: "tap2",
					},
				},
			},
		},
	}

	log.Println(proxyArpIf)
	db.Put(l3.ProxyArpInterfaceKey(proxyArpIf.InterfaceList[0].Label), proxyArpIf.InterfaceList[0])
}

func delProxyArpIf(db keyval.ProtoBroker) {
	db.Delete(l3.ProxyArpInterfaceKey("proxyArpIf1"))
}

func addProxyArpRng(db keyval.ProtoBroker) {
	proxyArpRng := l3.ProxyArpRanges{
		RangeList: []*l3.ProxyArpRanges_RangeList{
			{
				Label: "proxyArpRng1",
				Ranges: []*l3.ProxyArpRanges_RangeList_Range{
					{
						FirstIp: "124.168.10.5",
						LastIp:  "124.168.10.10",
					},
					{
						FirstIp: "172.154.10.5",
						LastIp:  "172.154.10.10",
					},
				},
			},
		},
	}

	log.Println(proxyArpRng)
	db.Put(l3.ProxyArpRangeKey(proxyArpRng.RangeList[0].Label), proxyArpRng.RangeList[0])
}

func delProxyArpRng(db keyval.ProtoBroker) {
	db.Delete(l3.ProxyArpRangeKey("proxyArpRng1"))
}