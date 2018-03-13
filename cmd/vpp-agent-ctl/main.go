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
	"fmt"
	"io/ioutil"
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
	l32 "github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/l3"
)

var (
	log          logging.Logger
	serviceLabel servicelabel.Plugin
)

func main() {
	log = logrus.DefaultLogger()
	log.SetLevel(logging.InfoLevel)
	flag.CommandLine.ParseEnv(os.Environ())

	serviceLabel.Init()

	bDB, db := createEtcdClient()

	ifName1 := "tap1"
	ifName2 := "tap2"
	memif := "memif"
	vxlan := "vxlan"
	eth := "GigabitEthernet0/8/0"
	loopback := "loop1"
	bridgeDomain1 := "bd1"
	afpacket1 := "afpacket1"
	veth1 := "veth1"
	veth2 := "veth2"

	if len(os.Args) > 1 && os.Args[len(os.Args)-1] == "-list" {
		listAllAgentKeys(bDB)
	} else if len(os.Args) > 1 && os.Args[len(os.Args)-1] == "-dump" {
		etcdDump(bDB, "")
	} else if len(os.Args) > 2 && os.Args[len(os.Args)-2] == "-dump" {
		etcdDump(bDB, os.Args[len(os.Args)-1])
	} else if len(os.Args) > 3 && os.Args[len(os.Args)-3] == "-put" {
		etcdPut(bDB, os.Args[len(os.Args)-2], os.Args[len(os.Args)-1])
	} else if len(os.Args) > 2 && os.Args[len(os.Args)-2] == "-get" {
		etcdGet(bDB, os.Args[len(os.Args)-1])
	} else if len(os.Args) > 2 && os.Args[len(os.Args)-2] == "-del" {
		etcdDel(bDB, os.Args[len(os.Args)-1])
	} else if len(os.Args) > 1 {
		switch os.Args[len(os.Args)-1] {
		case "-ct":
			create(db, ifName1, "192.168.1.2/24")
		case "-mt":
			create(db, ifName1, "192.168.1.2/24")
		case "-dt":
			delete(db, interfaces.InterfaceKey(ifName1))
		case "-ce":
			createEthernet(db, eth, "192.168.1.1", "2001:db8:0:0:0:ff00:5168:2bc8/48")
		case "-me":
			createEthernet(db, eth, "192.168.2.2", "2001:db8:0:0:0:ff00:48d2:c7d9/48")
		case "-de":
			delete(db, interfaces.InterfaceKey(eth))
		case "-cl":
			createLoopback(db, loopback, "8a:f1:be:90:00:dd", "192.168.15.1/24",
				"2001:db8:0:0:0:ff00:89e3:bb42/48")
		case "-ml":
			createLoopback(db, loopback, "b3:65:f1:f5:dc:f6", "192.168.25.2/24",
				"2001:db8:0:0:0:ff00:7772:1234/48")
		case "-dl":
			delete(db, interfaces.InterfaceKey(loopback))
		case "-cmm":
			createMemif(db, memif, "192.168.42.1/24", true)
		case "-dmm":
			delete(db, interfaces.InterfaceKey(memif))
		case "-cms":
			createMemif(db, memif, "192.168.42.2/24", false)
		case "-dms":
			delete(db, interfaces.InterfaceKey(memif))
		case "-cvx":
			createVxlan(db, vxlan, 13, "192.168.42.1", "192.168.42.2")
		case "-dvx":
			delete(db, interfaces.InterfaceKey(vxlan))
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
			createBfdSession(db, ifName1)
		case "-dbfds":
			delete(db, bfd.SessionKey(ifName1))
		case "-bfdk":
			createBfdKey(db, uint32(1))
		case "-dbfdk":
			delete(db, bfd.AuthKeysKey(string(1)))
		case "-bfde":
			createBfdEcho(db, ifName1)
		case "-dbfde":
			delete(db, bfd.EchoFunctionKey(ifName1))
		case "-daf":
			delete(db, interfaces.InterfaceKey(afpacket1))
		case "-cvth1":
			createVeth(db, veth1, veth2, "ns1", "d2:74:8c:12:67:d2", "192.168.22.1/24",
				"2001:db8:0:0:0:ff00:89e3:bb42/48")
		case "-cvth2":
			createVeth(db, veth2, veth1, "ns3", "92:c7:42:67:ab:cd", "192.168.22.5/24",
				"2001:842:0:0:0:ff00:13c7:1245/48")
		case "-cltap":
			createLinuxTap(db)
		case "-dltap":
			deleteLinuxTap(db)
		case "-dvth1":
			delete(db, linuxIntf.InterfaceKey(veth1))
		case "-dvth2":
			delete(db, linuxIntf.InterfaceKey(veth2))
		case "-ps":
			printState(db)
		case "-ierr":
			reportIfaceErrorState(db)
		case "-bderr":
			reportBdErrorState(db)
		case "-clarp":
			createLinuxArp(db)
		case "-dlarp":
			delete(db, l32.StaticArpKey("arp3"))
		case "-clrt":
			createLinuxRoute(db)
		case "-clrtdef":
			createDefaultLinuxRoute(db)
		case "-appns":
			createAppNamespace(db)
		case "-ef":
			enableL4Features(db)
		case "-df":
			disableL4Features(db)
		case "-dlrt":
			delete(db, l32.StaticRouteKey("route1"))
		case "-stna":
			createStnRule(db, ifName1, "10.1.1.3")
		case "-stnd":
			delete(db, stn.Key("rule1"))
		case "-natg":
			setNatGlobalConfig(db)
		case "-natdg":
			deleteNatGlobalConfig(db)
		case "-snat":
			createSNat(db)
		case "-dnat":
			createDNat(db)
		case "-ddnat":
			deleteDNat(db)
		default:
			usage()
		}
	} else {
		usage()
	}
}

func usage() {
	fmt.Println(os.Args[0], ": [etcd-config-file] <command>")
	fmt.Println("\tcommands: -ct -mt -dt -ce -me -cl -ml -dl -cmm -dmm -cms -dms -cvx -dvx -cr -dr -stna -stnd")
	fmt.Println(os.Args[0], ": [etcd-config-file] -put <etc_key> <json-file>")
	fmt.Println(os.Args[0], ": [etcd-config-file] -get <etc_key>")
}

func createACL(db keyval.ProtoBroker) {
	accessList := acl.AccessLists{}
	accessList.Acl = make([]*acl.AccessLists_Acl, 1)
	accessList.Acl[0] = new(acl.AccessLists_Acl)
	accessList.Acl[0].AclName = "acl1"
	accessList.Acl[0].Rules = make([]*acl.AccessLists_Acl_Rule, 1)
	accessList.Acl[0].Rules[0] = new(acl.AccessLists_Acl_Rule)
	accessList.Acl[0].Rules[0].Actions = new(acl.AccessLists_Acl_Rule_Actions)
	accessList.Acl[0].Rules[0].Actions.AclAction = acl.AclAction_PERMIT
	accessList.Acl[0].Rules[0].Matches = new(acl.AccessLists_Acl_Rule_Matches)

	//// Actions
	//accessList.Acl[0].Rules[1] = new(acl.AccessLists_Acl_Rule)
	//accessList.Acl[0].Rules[1].Actions = new(acl.AccessLists_Acl_Rule_Actions)
	//accessList.Acl[0].Rules[1].Actions.AclAction = acl.AclAction_PERMIT
	//accessList.Acl[0].Rules[1].Matches = new(acl.AccessLists_Acl_Rule_Matches)
	//
	//accessList.Acl[0].Rules[2] = new(acl.AccessLists_Acl_Rule)
	//accessList.Acl[0].Rules[2].Actions = new(acl.AccessLists_Acl_Rule_Actions)
	//accessList.Acl[0].Rules[2].Actions.AclAction = acl.AclAction_PERMIT
	//accessList.Acl[0].Rules[2].Matches = new(acl.AccessLists_Acl_Rule_Matches)

	// Ipv4Rule
	accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	accessList.Acl[0].Rules[0].Matches.IpRule.Ip = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Ip)
	accessList.Acl[0].Rules[0].Matches.IpRule.Ip.SourceNetwork = "192.168.1.1/32"
	accessList.Acl[0].Rules[0].Matches.IpRule.Ip.DestinationNetwork = "10.20.0.1/24"

	//// Ipv6Rule
	//accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Ip)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip.SourceNetwork = "1201:0db8:0a0b:12f0:0000:0000:0000:0000"
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip.DestinationNetwork = "5064:ff9b:0000:0000:0000:0000:0000:0000"
	//
	//// Ipv4Rule with empty destination address
	//accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Ip)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip.SourceNetwork = "192.168.1.2"
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip.DestinationNetwork = ""
	//
	//// Ipv6Rule with empty source address
	//accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Ip)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip.SourceNetwork = ""
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip.DestinationNetwork = "0064:ff9b:0000:0000:0000:0000:0000:0000"
	//
	//// Ip Rule with empty source and destination addresses
	//accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Ip)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip.SourceNetwork = ""
	//accessList.Acl[0].Rules[0].Matches.IpRule.Ip.DestinationNetwork = ""
	//
	//// Icmpv4 Rule (comment out "...IpRule = new(...)" to include the IP layer rule definition from above)
	//accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Icmp = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Icmp.IcmpTypeRange = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp_IcmpTypeRange)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Icmp.IcmpTypeRange.First = 150
	//accessList.Acl[0].Rules[0].Matches.IpRule.Icmp.IcmpTypeRange.Last = 250
	//accessList.Acl[0].Rules[0].Matches.IpRule.Icmp.IcmpCodeRange = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp_IcmpCodeRange)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Icmp.IcmpCodeRange.First = 1150
	//accessList.Acl[0].Rules[0].Matches.IpRule.Icmp.IcmpCodeRange.Last = 1250
	//
	//// Icmpv6 Rule (comment out "...IpRule = new(...)" to include the IP layer rule definition from above)
	//accessList.Acl[0].Rules[2].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	//accessList.Acl[0].Rules[2].Matches.IpRule.Icmp = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp)
	//accessList.Acl[0].Rules[2].Matches.IpRule.Icmp.IcmpTypeRange = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp_IcmpTypeRange)
	//accessList.Acl[0].Rules[2].Matches.IpRule.Icmp.IcmpTypeRange.First = 150
	//accessList.Acl[0].Rules[2].Matches.IpRule.Icmp.IcmpTypeRange.Last = 250
	//accessList.Acl[0].Rules[2].Matches.IpRule.Icmp.IcmpCodeRange = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Icmp_IcmpCodeRange)
	//accessList.Acl[0].Rules[2].Matches.IpRule.Icmp.IcmpCodeRange.First = 1150
	//accessList.Acl[0].Rules[2].Matches.IpRule.Icmp.IcmpCodeRange.Last = 1250
	//// Tcp Rule (comment out "...IpRule = new(...)" to include the IP layer rule definition from above)
	//accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp)
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp.SourcePortRange = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_SourcePortRange)
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp.SourcePortRange.LowerPort = 150
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp.SourcePortRange.UpperPort = 250
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp.DestinationPortRange = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Tcp_DestinationPortRange)
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp.DestinationPortRange.LowerPort = 1150
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp.DestinationPortRange.UpperPort = 1250
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp.TcpFlagsValue = 10
	accessList.Acl[0].Rules[0].Matches.IpRule.Tcp.TcpFlagsMask = 20
	//
	//// Udp Rule (comment out "...IpRule = new(...)" to include the IP layer rule definition from above)
	//accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Udp = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Udp)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Udp.SourcePortRange = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Udp_SourcePortRange)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Udp.SourcePortRange.LowerPort = 150
	//accessList.Acl[0].Rules[0].Matches.IpRule.Udp.SourcePortRange.UpperPort = 250
	//accessList.Acl[0].Rules[0].Matches.IpRule.Udp.DestinationPortRange = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Udp_DestinationPortRange)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Udp.DestinationPortRange.LowerPort = 1150
	//accessList.Acl[0].Rules[0].Matches.IpRule.Udp.DestinationPortRange.UpperPort = 1250
	//
	//// Other (comment out "...IpRule = new(...)" to include the IP layer rule definition from above)
	//accessList.Acl[0].Rules[0].Matches.IpRule = new(acl.AccessLists_Acl_Rule_Matches_IpRule)
	//accessList.Acl[0].Rules[0].Matches.IpRule.Other = new(acl.AccessLists_Acl_Rule_Matches_IpRule_Other)
	//
	//// Macip v4
	//accessList.Acl[0].Rules[0].Matches.MacipRule = new(acl.AccessLists_Acl_Rule_Matches_MacIpRule)
	//accessList.Acl[0].Rules[0].Matches.MacipRule.SourceAddress = "192.168.0.1"
	//accessList.Acl[0].Rules[0].Matches.MacipRule.SourceAddressPrefix = uint32(16)
	//accessList.Acl[0].Rules[0].Matches.MacipRule.SourceMacAddress = "b2:74:8c:12:67:d2"
	//accessList.Acl[0].Rules[0].Matches.MacipRule.SourceMacAddressMask = "ff:ff:ff:ff:00:00"
	//
	//// Macip v6
	//accessList.Acl[0].Rules[0].Matches.MacipRule = new(acl.AccessLists_Acl_Rule_Matches_MacIpRule)
	//accessList.Acl[0].Rules[0].Matches.MacipRule.SourceAddress = "12001:0db8:0a0b:12f0:0000:0000:0000:0001"
	//accessList.Acl[0].Rules[0].Matches.MacipRule.SourceAddressPrefix = uint32(64)
	//accessList.Acl[0].Rules[0].Matches.MacipRule.SourceMacAddress = "d2:74:8c:12:67:d2"
	//accessList.Acl[0].Rules[0].Matches.MacipRule.SourceMacAddressMask = "ff:ff:ff:ff:00:00"

	accessList.Acl[0].Interfaces = new(acl.AccessLists_Acl_Interfaces)
	accessList.Acl[0].Interfaces.Egress = make([]string, 2)
	accessList.Acl[0].Interfaces.Egress[0] = "tap1"
	accessList.Acl[0].Interfaces.Egress[1] = "tap2"
	accessList.Acl[0].Interfaces.Ingress = make([]string, 2)
	accessList.Acl[0].Interfaces.Ingress[0] = "tap3"
	accessList.Acl[0].Interfaces.Ingress[1] = "tap4"

	log.Print(accessList.Acl[0])

	db.Put(acl.Key(accessList.Acl[0].AclName), accessList.Acl[0])
	//db.Delete(acl.Key(accessList.Acl[0].AclName))
}

func etcdPut(bDB *etcdv3.BytesConnectionEtcd, key string, file string) {
	input, err := readData(file)

	log.Println("DB putting ", key, " ", string(input))

	err = bDB.Put(key, input)
	if err != nil {
		log.Panic("error putting the data ", key, " that to DB from ", file, ", err: ", err)
	}
	log.Println("DB put successful ", key, " ", file)
}
func readData(file string) ([]byte, error) {
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
			log.Panic("error reading the data that needs to be written to DB from ", file, ", err: ", err)
		}
	}

	// validate the JSON
	var js map[string]interface{}
	if json.Unmarshal(input, &js) != nil {
		log.Panic("Not a valid JSON: ", string(input))
	}
	return input, err
}

func etcdGet(bDB *etcdv3.BytesConnectionEtcd, key string) {
	log.Debug("GET ", key)

	data, found, _, err := bDB.GetValue(key)
	if err != nil {
		log.Error(err)
		return
	}
	if !found {
		log.Debug("No value found for the key", key)
	}
	fmt.Println(string(data))
}

func etcdDump(bDB *etcdv3.BytesConnectionEtcd, key string) {
	log.Debug("DUMP ", key)

	data, err := bDB.ListValues(key)
	if err != nil {
		log.Error(err)
		return
	}

	var found bool
	for {
		found = true
		kv, stop := data.GetNext()
		if stop {
			break
		}
		fmt.Println(kv.GetKey())
		fmt.Println(string(kv.GetValue()))
		fmt.Println()

	}

	if !found {
		log.Debug("No value found for the key", key)
	}
}

func etcdDel(bDB *etcdv3.BytesConnectionEtcd, key string) {
	found, err := bDB.Delete(key, datasync.WithPrefix())
	if err != nil {
		log.Error(err)
		return
	}
	if found {
		log.Debug("Data deleted:", key)
	} else {
		log.Debug("No value found for the key", key)
	}
}

func createRoute(db keyval.ProtoBroker) {
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

	key := l3.RouteKey(routes.Route[0].VrfId, routes.Route[0].DstIpAddr, routes.Route[0].NextHopAddr)
	db.Put(key, routes.Route[0])
	log.Printf("Adding route %v", key)
}

func deleteRoute(db keyval.ProtoBroker, routeDstIP string, routeNhIP string) {
	path := l3.RouteKey(0, routeDstIP, "192.168.1.13")
	db.Delete(path)
	log.WithField("path", path).Debug("Removing route")
}

func txn(db keyval.ProtoBroker) {
	ifs := interfaces.Interfaces{}
	ifs.Interface = make([]*interfaces.Interfaces_Interface, 2)

	ifs.Interface[0] = new(interfaces.Interfaces_Interface)
	ifs.Interface[0].Name = "tap1"
	ifs.Interface[0].Type = interfaces.InterfaceType_TAP_INTERFACE
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].Mtu = 1500
	ifs.Interface[0].IpAddresses = make([]string, 1)
	ifs.Interface[0].IpAddresses[0] = "10.4.4.1/24"
	ifs.Interface[0].Tap = &interfaces.Interfaces_Interface_Tap{HostIfName: "tap1"}

	ifs.Interface[1] = new(interfaces.Interfaces_Interface)
	ifs.Interface[1].Name = "tap2"
	ifs.Interface[1].Type = interfaces.InterfaceType_TAP_INTERFACE
	ifs.Interface[1].Enabled = true
	ifs.Interface[1].Mtu = 1500
	ifs.Interface[1].IpAddresses = make([]string, 1)
	ifs.Interface[1].IpAddresses[0] = "10.4.4.2/24"
	ifs.Interface[1].Tap = &interfaces.Interfaces_Interface_Tap{HostIfName: "tap2"}

	bd01 := l2.BridgeDomains_BridgeDomain{
		Name:                "aaa",
		Flood:               false,
		UnknownUnicastFlood: false,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /*means disable aging*/
	}

	t := db.NewTxn()
	t.Put(interfaces.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])
	t.Put(interfaces.InterfaceKey(ifs.Interface[1].Name), ifs.Interface[1])
	t.Put(l2.BridgeDomainKey("bd01"), &bd01)

	t.Commit()

}

func deleteTxn(db keyval.ProtoBroker) {
	db.Delete(interfaces.InterfaceKey("tap1"))
	db.Delete(interfaces.InterfaceKey("tap2"))
}

func listAllAgentKeys(db *etcdv3.BytesConnectionEtcd) {
	log.Debug("listAllAgentKeys")

	it, err := db.ListKeys(serviceLabel.GetAllAgentsPrefix())
	if err != nil {
		log.Error(err)
	}
	for {
		key, _, stop := it.GetNext()
		if stop {
			break
		}
		fmt.Println("key: ", key)
	}
}

func create(db keyval.ProtoBroker, ifname string, ipAddr string) {
	// fill in data - option 1
	ifs := interfaces.Interfaces{}
	ifs.Interface = make([]*interfaces.Interfaces_Interface, 4)

	ifs.Interface[0] = new(interfaces.Interfaces_Interface)
	ifs.Interface[0].Name = "tap2"
	ifs.Interface[0].Type = interfaces.InterfaceType_TAP_INTERFACE
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].PhysAddress = "09:9e:df:66:54:42"
	ifs.Interface[0].Mtu = 555
	ifs.Interface[0].IpAddresses = make([]string, 1)
	ifs.Interface[0].IpAddresses[0] = "192.168.20.3/24"
	//ifs.Interface[0].IpAddresses[0] = "192.168.2.9/24"
	//ifs.Interface[0].IpAddresses[2] = "10.10.1.7/24"
	//ifs.Interface[0].Unnumbered = &interfaces.Interfaces_Interface_Unnumbered{}
	//ifs.Interface[0].Unnumbered.IsUnnumbered = true
	//ifs.Interface[0].Unnumbered.InterfaceWithIP = "memif"
	//ifs.Interface[0].IpAddresses[0] = "2002:db8:0:0:0:ff00:42:8329"
	ifs.Interface[0].Tap = &interfaces.Interfaces_Interface_Tap{HostIfName: "tap2"}

	log.Println(ifs)

	db.Put(interfaces.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])

}

func createEthernet(db keyval.ProtoBroker, ifname string, ipv4Addr string, ipv6Addr string) {
	ifs := interfaces.Interfaces{}
	ifs.Interface = make([]*interfaces.Interfaces_Interface, 2)

	ifs.Interface[0] = new(interfaces.Interfaces_Interface)
	ifs.Interface[0].Name = ifname
	ifs.Interface[0].Type = interfaces.InterfaceType_ETHERNET_CSMACD
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].PhysAddress = ""

	// Ipv4
	ifs.Interface[0].SetDhcpClient = false
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].Mtu = 1500
	ifs.Interface[0].IpAddresses = make([]string, 1)
	ifs.Interface[0].IpAddresses[0] = ipv4Addr

	// Ipv6
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].Mtu = 1500
	ifs.Interface[0].IpAddresses = make([]string, 1)
	ifs.Interface[0].IpAddresses[0] = ipv6Addr

	log.Println(ifs)

	db.Put(interfaces.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])
}

func createLoopback(db keyval.ProtoBroker, ifname string, physAddr string, ipv4Addr string, ipv6Addr string) {
	ifs := interfaces.Interfaces{}
	ifs.Interface = make([]*interfaces.Interfaces_Interface, 1)

	ifs.Interface[0] = new(interfaces.Interfaces_Interface)
	ifs.Interface[0].Name = ifname
	ifs.Interface[0].Type = interfaces.InterfaceType_SOFTWARE_LOOPBACK
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].PhysAddress = physAddr

	ifs.Interface[0].Enabled = true
	ifs.Interface[0].Mtu = 1478
	ifs.Interface[0].IpAddresses = make([]string, 2)
	ifs.Interface[0].IpAddresses[0] = ipv4Addr
	ifs.Interface[0].IpAddresses[1] = ipv6Addr

	log.Println(ifs)

	db.Put(interfaces.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])
}

func createAfPacket(db keyval.ProtoBroker, ifname string, hostIfName string, physAddr string, ipv4Addr string, ipv6Addr string) {
	ifs := interfaces.Interfaces{}
	ifs.Interface = make([]*interfaces.Interfaces_Interface, 1)

	ifs.Interface[0] = new(interfaces.Interfaces_Interface)
	ifs.Interface[0].Name = ifname
	ifs.Interface[0].Type = interfaces.InterfaceType_AF_PACKET_INTERFACE
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].PhysAddress = physAddr

	ifs.Interface[0].Enabled = true
	ifs.Interface[0].Mtu = 1500
	ifs.Interface[0].IpAddresses = make([]string, 1)
	ifs.Interface[0].IpAddresses[0] = ipv4Addr

	ifs.Interface[0].Enabled = true
	ifs.Interface[0].Mtu = 1500
	ifs.Interface[0].IpAddresses = make([]string, 1)
	ifs.Interface[0].IpAddresses[0] = ipv6Addr

	ifs.Interface[0].Afpacket = new(interfaces.Interfaces_Interface_Afpacket)
	ifs.Interface[0].Afpacket.HostIfName = hostIfName

	log.Println(ifs)

	db.Put(interfaces.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])
}

func createVeth(db keyval.ProtoBroker, ifname string, peerIfName string, ns string, physAddr string, ipv4Addr string, ipv6Addr string) {
	ifs := linuxIntf.LinuxInterfaces{}
	ifs.Interface = make([]*linuxIntf.LinuxInterfaces_Interface, 1)

	ifs.Interface[0] = new(linuxIntf.LinuxInterfaces_Interface)
	ifs.Interface[0].Name = ifname
	ifs.Interface[0].Type = linuxIntf.LinuxInterfaces_VETH
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].PhysAddress = physAddr

	ifs.Interface[0].Namespace = new(linuxIntf.LinuxInterfaces_Interface_Namespace)
	ifs.Interface[0].Namespace.Type = linuxIntf.LinuxInterfaces_Interface_Namespace_NAMED_NS
	ifs.Interface[0].Namespace.Name = ns

	ifs.Interface[0].Enabled = true
	ifs.Interface[0].Mtu = 1500
	ifs.Interface[0].IpAddresses = make([]string, 1)
	ifs.Interface[0].IpAddresses[0] = ipv4Addr

	//ifs.Interface[0].Enabled = true
	//ifs.Interface[0].Mtu = 1500
	//ifs.Interface[0].IpAddresses = make([]string, 1)
	//ifs.Interface[0].IpAddresses[0] = ipv6Addr

	ifs.Interface[0].Veth = &linuxIntf.LinuxInterfaces_Interface_Veth{PeerIfName: peerIfName}

	log.Println(ifs)

	db.Put(linuxIntf.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])
}
func createLinuxTap(db keyval.ProtoBroker) {
	ifs := linuxIntf.LinuxInterfaces{}
	ifs.Interface = make([]*linuxIntf.LinuxInterfaces_Interface, 1)

	ifs.Interface[0] = new(linuxIntf.LinuxInterfaces_Interface)
	ifs.Interface[0].Name = "tap1"
	ifs.Interface[0].HostIfName = "tap2"
	ifs.Interface[0].Type = linuxIntf.LinuxInterfaces_AUTO_TAP
	ifs.Interface[0].Enabled = true
	ifs.Interface[0].PhysAddress = "92:c7:42:67:ab:cc"

	ifs.Interface[0].Namespace = new(linuxIntf.LinuxInterfaces_Interface_Namespace)
	ifs.Interface[0].Namespace.Type = linuxIntf.LinuxInterfaces_Interface_Namespace_NAMED_NS
	ifs.Interface[0].Namespace.Name = "ns2"

	ifs.Interface[0].Enabled = true
	ifs.Interface[0].Mtu = 1155
	ifs.Interface[0].IpAddresses = make([]string, 1)
	ifs.Interface[0].IpAddresses[0] = "172.52.45.127/24"

	log.Println(ifs)

	db.Put(linuxIntf.InterfaceKey(ifs.Interface[0].Name), ifs.Interface[0])
}

func deleteLinuxTap(db keyval.ProtoBroker) {
	ifs := linuxIntf.LinuxInterfaces{}
	ifs.Interface = make([]*linuxIntf.LinuxInterfaces_Interface, 1)

	ifs.Interface[0] = new(linuxIntf.LinuxInterfaces_Interface)
	ifs.Interface[0].Name = "tap1"

	db.Delete(linuxIntf.InterfaceKey(ifs.Interface[0].Name))
}

func createMemif(db keyval.ProtoBroker, ifname string, ipAddr string, master bool) {
	key := interfaces.InterfaceKey(ifname)
	iface := interfaces.Interfaces_Interface{
		Name:    ifname,
		Type:    interfaces.InterfaceType_MEMORY_INTERFACE,
		Enabled: true,
		Memif: &interfaces.Interfaces_Interface_Memif{
			Id:             1,
			Secret:         "secret",
			Master:         master,
			SocketFilename: "/tmp/memif1.sock",
		},
		Mtu:         1478,
		IpAddresses: []string{ipAddr},
	}
	log.Println(key, iface)
	db.Put(key, &iface)
}

func createVxlan(db keyval.ProtoBroker, ifname string, vni uint32, src string, dst string) {

	iface := interfaces.Interfaces_Interface{
		Name:    ifname,
		Type:    interfaces.InterfaceType_VXLAN_TUNNEL,
		Enabled: true,
		Vxlan: &interfaces.Interfaces_Interface_Vxlan{
			SrcAddress: src,
			DstAddress: dst,
			Vni:        vni,
		},
	}
	log.Println(iface)
	db.Put(interfaces.InterfaceKey(iface.Name), &iface)
}

func delete(db keyval.ProtoBroker, key string) {
	db.Delete(key)
	log.Println("Deleting", key)
}

func createEtcdClient() (*etcdv3.BytesConnectionEtcd, keyval.ProtoBroker) {

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
			log.Fatal(err)
		}
	}
	etcdConfig, err := etcdv3.ConfigToClientv3(cfg)
	if err != nil {
		log.Fatal(err)
	}

	etcdLogger := logrus.NewLogger("etcdLogger")
	etcdLogger.SetLevel(logging.WarnLevel)

	bDB, err := etcdv3.NewEtcdConnectionWithBytes(*etcdConfig, etcdLogger)
	if err != nil {
		log.Fatal(err)
	}

	return bDB, kvproto.NewProtoWrapperWithSerializer(bDB, &keyval.SerializerJSON{}).
		NewBroker(serviceLabel.GetAgentPrefix())
}

func createBridgeDomain(db keyval.ProtoBroker, bdName string) {
	bd := l2.BridgeDomains{}
	bd.BridgeDomains = make([]*l2.BridgeDomains_BridgeDomain, 1)

	bd.BridgeDomains[0] = new(l2.BridgeDomains_BridgeDomain)
	bd.BridgeDomains[0].Name = "bd1"
	bd.BridgeDomains[0].Learn = true
	bd.BridgeDomains[0].ArpTermination = true
	bd.BridgeDomains[0].Flood = true
	bd.BridgeDomains[0].UnknownUnicastFlood = true
	bd.BridgeDomains[0].Forward = true

	bd.BridgeDomains[0].Interfaces = make([]*l2.BridgeDomains_BridgeDomain_Interfaces, 1)
	bd.BridgeDomains[0].Interfaces[0] = new(l2.BridgeDomains_BridgeDomain_Interfaces)
	bd.BridgeDomains[0].Interfaces[0].Name = "tap1"
	bd.BridgeDomains[0].Interfaces[0].BridgedVirtualInterface = false
	bd.BridgeDomains[0].Interfaces[0].SplitHorizonGroup = 1
	//bd.BridgeDomains[0].Interfaces[1] = new(l2.BridgeDomains_BridgeDomain_Interfaces)
	//bd.BridgeDomains[0].Interfaces[1].Name = "tap2"
	//bd.BridgeDomains[0].Interfaces[1].BridgedVirtualInterface = true

	log.Println(bd)
	db.Put(l2.BridgeDomainKey(bd.BridgeDomains[0].Name), bd.BridgeDomains[0])
}

func addArpEntry(db keyval.ProtoBroker, iface string) {
	arpTable := l3.ArpTable{}
	arpTable.ArpTableEntries = make([]*l3.ArpTable_ArpTableEntry, 1)
	arpTable.ArpTableEntries[0] = new(l3.ArpTable_ArpTableEntry)
	arpTable.ArpTableEntries[0].Interface = "tap1"
	arpTable.ArpTableEntries[0].IpAddress = "192.168.10.21"
	arpTable.ArpTableEntries[0].PhysAddress = "59:6C:45:59:8E:BD"
	arpTable.ArpTableEntries[0].Static = true

	log.Println(arpTable)
	db.Put(l3.ArpEntryKey(arpTable.ArpTableEntries[0].Interface, arpTable.ArpTableEntries[0].IpAddress), arpTable.ArpTableEntries[0])
}

func deleteArpEntry(db keyval.ProtoBroker, iface string) {
	arpTable := l3.ArpTable{}
	arpTable.ArpTableEntries = make([]*l3.ArpTable_ArpTableEntry, 1)
	arpTable.ArpTableEntries[0] = new(l3.ArpTable_ArpTableEntry)
	arpTable.ArpTableEntries[0].Interface = "tap1"
	arpTable.ArpTableEntries[0].IpAddress = "192.168.10.21"
	arpTable.ArpTableEntries[0].PhysAddress = "59:6C:45:59:8E:BD"
	arpTable.ArpTableEntries[0].Static = true

	log.Println(arpTable)
	db.Delete(l3.ArpEntryKey(arpTable.ArpTableEntries[0].Interface, arpTable.ArpTableEntries[0].IpAddress))
}

func addArpTableEntry(db keyval.ProtoBroker, bdName string) {
	bd := l2.BridgeDomains{}
	bd.BridgeDomains = make([]*l2.BridgeDomains_BridgeDomain, 1)

	bd.BridgeDomains[0] = new(l2.BridgeDomains_BridgeDomain)
	bd.BridgeDomains[0].Name = bdName
	bd.BridgeDomains[0].Learn = true
	bd.BridgeDomains[0].ArpTermination = true
	bd.BridgeDomains[0].Flood = true
	bd.BridgeDomains[0].UnknownUnicastFlood = true
	bd.BridgeDomains[0].Forward = true
	bd.BridgeDomains[0].ArpTerminationTable = make([]*l2.BridgeDomains_BridgeDomain_ArpTerminationTable, 2)
	bd.BridgeDomains[0].ArpTerminationTable[0] = new(l2.BridgeDomains_BridgeDomain_ArpTerminationTable)
	bd.BridgeDomains[0].ArpTerminationTable[1] = new(l2.BridgeDomains_BridgeDomain_ArpTerminationTable)
	bd.BridgeDomains[0].ArpTerminationTable[0].PhysAddress = "a7:65:f1:b5:dc:f6"
	bd.BridgeDomains[0].ArpTerminationTable[0].IpAddress = "192.168.10.10"
	bd.BridgeDomains[0].ArpTerminationTable[1].PhysAddress = "59:6C:45:59:8E:BC"
	bd.BridgeDomains[0].ArpTerminationTable[1].IpAddress = "10.10.0.1"

	log.Println(bd)
	db.Put(l2.BridgeDomainKey(bd.BridgeDomains[0].Name), bd.BridgeDomains[0])
}

func addStaticFibTableEntry(db keyval.ProtoBroker, bdName string, iface string) {
	fibTable := l2.FibTableEntries{}
	fibTable.FibTableEntry = make([]*l2.FibTableEntries_FibTableEntry, 1)
	fibTable.FibTableEntry[0] = new(l2.FibTableEntries_FibTableEntry)
	fibTable.FibTableEntry[0].OutgoingInterface = iface
	fibTable.FibTableEntry[0].BridgeDomain = bdName
	fibTable.FibTableEntry[0].PhysAddress = "aa:65:f1:59:8E:BC"
	fibTable.FibTableEntry[0].BridgedVirtualInterface = false

	log.Println(fibTable)

	db.Put(l2.FibKey(fibTable.FibTableEntry[0].BridgeDomain, fibTable.FibTableEntry[0].PhysAddress), fibTable.FibTableEntry[0])
}

func deleteStaticFibTableEntry(db keyval.ProtoBroker, bdName string) {
	fibTable := l2.FibTableEntries{}
	fibTable.FibTableEntry = make([]*l2.FibTableEntries_FibTableEntry, 1)
	fibTable.FibTableEntry[0] = new(l2.FibTableEntries_FibTableEntry)
	fibTable.FibTableEntry[0].PhysAddress = "aa:65:f1:59:8E:BC"
	fibTable.FibTableEntry[0].BridgeDomain = bdName

	log.Println(fibTable)

	db.Delete(l2.FibKey(fibTable.FibTableEntry[0].BridgeDomain, fibTable.FibTableEntry[0].PhysAddress))
}

func createL2xConnect(db keyval.ProtoBroker, ifnameRx string, ifnameTx string) {
	xcp := l2.XConnectPairs{}
	xcp.XConnectPairs = make([]*l2.XConnectPairs_XConnectPair, 1)
	xcp.XConnectPairs[0] = new(l2.XConnectPairs_XConnectPair)
	xcp.XConnectPairs[0].ReceiveInterface = ifnameRx
	xcp.XConnectPairs[0].TransmitInterface = ifnameTx

	log.Println(xcp)

	db.Put(l2.XConnectKey(xcp.XConnectPairs[0].ReceiveInterface), xcp.XConnectPairs[0])
}

func createBfdSession(db keyval.ProtoBroker, iface string) {
	singleHopBfd := bfd.SingleHopBFD{}
	singleHopBfd.Sessions = make([]*bfd.SingleHopBFD_Session, 1)
	singleHopBfd.Sessions[0] = new(bfd.SingleHopBFD_Session)
	singleHopBfd.Sessions[0].Interface = iface
	singleHopBfd.Sessions[0].RequiredMinRxInterval = 8
	singleHopBfd.Sessions[0].DesiredMinTxInterval = 3
	singleHopBfd.Sessions[0].SourceAddress = "192.168.1.2"
	singleHopBfd.Sessions[0].DestinationAddress = "20.10.0.5"
	//singleHopBfd.Sessions[0].SourceAddress = "2001:db8:0:0:0:ff00:42:8329"
	//singleHopBfd.Sessions[0].DestinationAddress = "2871:db18:0:0:0:ff00:42:8329"
	singleHopBfd.Sessions[0].DetectMultiplier = 9
	singleHopBfd.Sessions[0].Enabled = true
	singleHopBfd.Sessions[0].Authentication = new(bfd.SingleHopBFD_Session_Authentication)
	singleHopBfd.Sessions[0].Authentication.KeyId = 1
	singleHopBfd.Sessions[0].Authentication.AdvertisedKeyId = 1

	log.Println(singleHopBfd)

	db.Put(bfd.SessionKey(singleHopBfd.Sessions[0].Interface), singleHopBfd.Sessions[0])
}

func createBfdKey(db keyval.ProtoBroker, id uint32) {
	singleHopBfd := bfd.SingleHopBFD{}
	singleHopBfd.Keys = make([]*bfd.SingleHopBFD_Key, 1)
	singleHopBfd.Keys[0] = new(bfd.SingleHopBFD_Key)
	singleHopBfd.Keys[0].Id = id
	singleHopBfd.Keys[0].AuthenticationType = bfd.SingleHopBFD_Key_METICULOUS_KEYED_SHA1
	singleHopBfd.Keys[0].Secret = "1981491891941891"

	log.Println(singleHopBfd)

	db.Put(bfd.AuthKeysKey(string(singleHopBfd.Keys[0].Id)), singleHopBfd.Keys[0])
}

func createBfdEcho(db keyval.ProtoBroker, iface string) {
	singleHopBfd := bfd.SingleHopBFD{}
	singleHopBfd.EchoFunction = new(bfd.SingleHopBFD_EchoFunction)
	singleHopBfd.EchoFunction.EchoSourceInterface = iface

	log.Println(singleHopBfd)

	db.Put(bfd.EchoFunctionKey(singleHopBfd.EchoFunction.EchoSourceInterface), singleHopBfd.EchoFunction)
}

func reportIfaceErrorState(db keyval.ProtoBroker) {
	ifErr, err := db.ListValues(interfaces.IfErrorPrefix)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
			return
		}

		fmt.Println(entry)
	}
}

func reportBdErrorState(db keyval.ProtoBroker) {
	bdErr, err := db.ListValues(l2.BdErrPrefix)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
			return
		}

		fmt.Println(entry)
	}
}

func printState(db keyval.ProtoBroker) {
	respIf, errIf := db.ListValues(interfaces.InterfaceStateKeyPrefix())
	if errIf != nil {
		log.Fatal(errIf)
		return
	}

	// interfaces
	for {

		kv, allReceived := respIf.GetNext()

		if allReceived {
			break
		}
		entry := &interfaces.InterfacesState_Interface{}
		err := kv.GetValue(entry)
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Println(entry)
	}

	respBd, errBd := db.ListValues(l2.BridgeDomainKeyPrefix())
	if errBd != nil {
		log.Fatal(errBd)
		return
	}

	// bridge domains
	for {
		kv, allReceived := respBd.GetNext()
		if allReceived {
			break
		}

		entry := &l2.BridgeDomains_BridgeDomain{}
		err := kv.GetValue(entry)
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Println(entry)
	}
}

func createLinuxArp(db keyval.ProtoBroker) {
	linuxArpEntries := l32.LinuxStaticArpEntries{}
	linuxArpEntries.ArpEntry = make([]*l32.LinuxStaticArpEntries_ArpEntry, 1)
	linuxArpEntries.ArpEntry[0] = new(l32.LinuxStaticArpEntries_ArpEntry)
	linuxArpEntries.ArpEntry[0].Name = "arp1"
	linuxArpEntries.ArpEntry[0].Namespace = new(l32.LinuxStaticArpEntries_ArpEntry_Namespace)
	linuxArpEntries.ArpEntry[0].Namespace.Type = l32.LinuxStaticArpEntries_ArpEntry_Namespace_NAMED_NS
	linuxArpEntries.ArpEntry[0].Namespace.Name = "ns1"
	linuxArpEntries.ArpEntry[0].Interface = "veth1"
	linuxArpEntries.ArpEntry[0].IpAddr = "130.0.0.1"
	linuxArpEntries.ArpEntry[0].HwAddress = "ab:cd:ef:01:02:03"
	linuxArpEntries.ArpEntry[0].State = new(l32.LinuxStaticArpEntries_ArpEntry_NudState)
	linuxArpEntries.ArpEntry[0].State.Type = l32.LinuxStaticArpEntries_ArpEntry_NudState_PERMANENT
	linuxArpEntries.ArpEntry[0].IpFamily = new(l32.LinuxStaticArpEntries_ArpEntry_IpFamily)
	linuxArpEntries.ArpEntry[0].IpFamily.Family = l32.LinuxStaticArpEntries_ArpEntry_IpFamily_IPV4

	log.Println(linuxArpEntries)

	db.Put(l32.StaticArpKey(linuxArpEntries.ArpEntry[0].Name), linuxArpEntries.ArpEntry[0])
}

func createLinuxRoute(db keyval.ProtoBroker) {
	linuxRoutes := l32.LinuxStaticRoutes{}
	linuxRoutes.Route = make([]*l32.LinuxStaticRoutes_Route, 1)
	linuxRoutes.Route[0] = new(l32.LinuxStaticRoutes_Route)
	linuxRoutes.Route[0].Name = "route1"
	linuxRoutes.Route[0].Namespace = new(l32.LinuxStaticRoutes_Route_Namespace)
	linuxRoutes.Route[0].Namespace.Type = l32.LinuxStaticRoutes_Route_Namespace_NAMED_NS
	linuxRoutes.Route[0].Namespace.Name = "ns1"
	linuxRoutes.Route[0].DstIpAddr = "10.0.2.0/24"
	//linuxRoutes.Route[0].SrcIpAddr = "128.0.0.10"
	//linuxRoutes.Route[0].GwAddr = "128.0.0.1"
	linuxRoutes.Route[0].Interface = "veth1"
	linuxRoutes.Route[0].Metric = 100

	log.Println(linuxRoutes)

	db.Put(l32.StaticRouteKey(linuxRoutes.Route[0].Name), linuxRoutes.Route[0])
}

func createDefaultLinuxRoute(db keyval.ProtoBroker) {
	linuxRoutes := l32.LinuxStaticRoutes{}
	linuxRoutes.Route = make([]*l32.LinuxStaticRoutes_Route, 1)
	linuxRoutes.Route[0] = new(l32.LinuxStaticRoutes_Route)
	linuxRoutes.Route[0].Name = "defRoute"
	linuxRoutes.Route[0].Namespace = new(l32.LinuxStaticRoutes_Route_Namespace)
	linuxRoutes.Route[0].Namespace.Type = l32.LinuxStaticRoutes_Route_Namespace_NAMED_NS
	linuxRoutes.Route[0].Namespace.Name = "ns1"
	linuxRoutes.Route[0].Default = true
	linuxRoutes.Route[0].Interface = "veth1"
	linuxRoutes.Route[0].GwAddr = "10.0.2.2"
	linuxRoutes.Route[0].Metric = 100

	log.Println(linuxRoutes)

	db.Put(l32.StaticRouteKey(linuxRoutes.Route[0].Name), linuxRoutes.Route[0])
}

func createAppNamespace(db keyval.ProtoBroker) {
	appNamespace := l4.AppNamespaces{}
	appNamespace.AppNamespaces = make([]*l4.AppNamespaces_AppNamespace, 1)
	appNamespace.AppNamespaces[0] = new(l4.AppNamespaces_AppNamespace)
	appNamespace.AppNamespaces[0].NamespaceId = "ns8"
	appNamespace.AppNamespaces[0].Secret = 1
	appNamespace.AppNamespaces[0].Interface = "tap1"

	log.Println(appNamespace)

	db.Put(l4.AppNamespacesKey(appNamespace.AppNamespaces[0].NamespaceId), appNamespace.AppNamespaces[0])
}

func enableL4Features(db keyval.ProtoBroker) {
	l4Fatures := &l4.L4Features{}
	l4Fatures.Enabled = true

	log.Println(l4Fatures)

	db.Put(l4.FeatureKey(), l4Fatures)
}

func disableL4Features(db keyval.ProtoBroker) {
	l4Fatures := &l4.L4Features{}
	l4Fatures.Enabled = false

	log.Println(l4Fatures)

	db.Put(l4.FeatureKey(), l4Fatures)
}

func createStnRule(db keyval.ProtoBroker, ifName string, ipAddress string) {
	stnRule := stn.StnRule{
		RuleName:  "rule1",
		IpAddress: ipAddress,
		Interface: ifName,
	}

	log.Println(stnRule)

	db.Put(stn.Key(stnRule.RuleName), &stnRule)
}

func setNatGlobalConfig(db keyval.ProtoBroker) {
	natGlobal := &nat.Nat44Global{}
	natGlobal.Forwarding = false
	natGlobal.NatInterfaces = make([]*nat.Nat44Global_NatInterfaces, 3)
	natGlobal.NatInterfaces[0] = &nat.Nat44Global_NatInterfaces{
		Name:          "tap1",
		IsInside:      false,
		OutputFeature: false,
	}
	natGlobal.NatInterfaces[1] = &nat.Nat44Global_NatInterfaces{
		Name:          "tap2",
		IsInside:      false,
		OutputFeature: false,
	}
	natGlobal.NatInterfaces[2] = &nat.Nat44Global_NatInterfaces{
		Name:          "tap3",
		IsInside:      false,
		OutputFeature: false,
	}
	natGlobal.AddressPools = make([]*nat.Nat44Global_AddressPools, 3)
	natGlobal.AddressPools[0] = &nat.Nat44Global_AddressPools{
		VrfId:           0,
		FirstSrcAddress: "192.168.0.1",
		TwiceNat:        false,
	}
	natGlobal.AddressPools[1] = &nat.Nat44Global_AddressPools{
		VrfId:           0,
		FirstSrcAddress: "175.124.0.1",
		LastSrcAddress:  "175.124.0.3",
		TwiceNat:        false,
	}
	natGlobal.AddressPools[2] = &nat.Nat44Global_AddressPools{
		VrfId:           0,
		FirstSrcAddress: "10.10.0.1",
		LastSrcAddress:  "10.10.0.2",
		TwiceNat:        false,
	}

	log.Println(natGlobal)

	db.Put(nat.GlobalConfigKey(), natGlobal)

	log.Println(nat.GlobalConfigKey())
}

func deleteNatGlobalConfig(db keyval.ProtoBroker) {
	db.Delete(nat.GlobalConfigKey())
}

func createSNat(db keyval.ProtoBroker) {
	sNat := &nat.Nat44SNat_SNatConfig{
		Label: "pool1",
	}

	log.Println(sNat)

	db.Put(nat.SNatKey(sNat.Label), sNat)
}

func createDNat(db keyval.ProtoBroker) {
	// Local IP list
	var localIPs []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs
	localIP := &nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
		LocalIP:     "172.124.0.2",
		LocalPort:   6500,
		Probability: 40,
	}
	localIPs = append(localIPs, localIP)
	localIP = &nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
		LocalIP:     "172.125.10.5",
		LocalPort:   2300,
		Probability: 40,
	}
	localIPs = append(localIPs, localIP)

	// Static mapping
	var mapping []*nat.Nat44DNat_DNatConfig_StaticMappings
	entry := &nat.Nat44DNat_DNatConfig_StaticMappings{
		VrfId:             0,
		ExternalInterface: "tap1",
		ExternalIP:        "192.168.0.1",
		ExternalPort:      8989,
		LocalIps:          localIPs,
		Protocol:          1,
		TwiceNat:          false,
	}
	mapping = append(mapping, entry)

	// Identity mapping
	var idMapping []*nat.Nat44DNat_DNatConfig_IdentityMappings
	idEntry := &nat.Nat44DNat_DNatConfig_IdentityMappings{
		VrfId: 0,
		//AddressedInterface: "tap1",
		IpAddress: "10.10.0.1",
		Port:      2525,
		Protocol:  0,
	}
	idMapping = append(idMapping, idEntry)

	// DNat config
	dNat := &nat.Nat44DNat_DNatConfig{
		Label:      "dnat1",
		StMappings: mapping,
		IdMappings: idMapping,
	}

	log.Println(dNat)

	db.Put(nat.DNatKey(dNat.Label), dNat)
}

func deleteDNat(db keyval.ProtoBroker) {
	dNat := &nat.Nat44DNat_DNatConfig{
		Label: "dnat1",
	}

	log.Println(dNat)

	db.Delete(nat.DNatKey(dNat.Label))
}
