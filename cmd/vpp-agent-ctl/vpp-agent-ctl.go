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
	"bytes"
	"os"
	"strings"

	"github.com/ligato/cn-infra/logging/logrus"
	ctlImpl "github.com/ligato/vpp-agent/cmd/vpp-agent-ctl/impl"
)

func main() {
	// Read args
	args := os.Args
	argsLen := len(args)

	// First argument is not a command
	if argsLen == 1 {
		usage()
		return
	}
	// Check if second argument is a command or path to the ETCD config file
	var etcdCfg string
	var cmdSet []string
	if argsLen >= 2 && !strings.HasPrefix(args[1], "-") {
		etcdCfg = args[1]
		// Remove first two arguments
		cmdSet = args[2:]
	} else {
		// Remove first argument
		cmdSet = args[1:]
	}
	ctl, err := ctlImpl.Init(etcdCfg, cmdSet)
	if err != nil {
		// Error is already printed in 'bytes_broker_impl.go'
		usage()
		return
	}

	do(ctl)
}

func do(ctl *ctlImpl.VppAgentCtl) {
	switch ctl.Commands[0] {
	case "-list":
		// List all keys
		ctl.ListAllAgentKeys()
	case "-dump":
		if len(ctl.Commands) > 2 {
			// Dump specific key
			ctl.EtcdDump(ctl.Commands[1])
		} else {
			// Dump all keys
			ctl.EtcdDump("")
		}
	case "-get":
		if len(ctl.Commands) > 2 {
			// Get key
			ctl.EtcdGet(ctl.Commands[1])
		}
	case "-del":
		if len(ctl.Commands) > 2 {
			// Del key
			ctl.EtcdDel(ctl.Commands[1])
		}
	case "-put":
		if len(ctl.Commands) > 3 {
			ctl.EtcdPut(ctl.Commands[1], ctl.Commands[2])
		}
	default:
		switch ctl.Commands[0] {
		// ACL
		case "-acl":
			ctl.CreateACL()
		case "-acld":
			ctl.DeleteACL()
			// BFD
		case "-bfds":
			ctl.CreateBfdSession()
		case "-bfdsd":
			ctl.DeleteBfdSession()
		case "-bfdk":
			ctl.CreateBfdKey()
		case "-bfdkd":
			ctl.DeleteBfdKey()
		case "-bfde":
			ctl.CreateBfdEcho()
		case "-bfded":
			ctl.DeleteBfdEcho()
			// VPP interfaces
		case "-eth":
			ctl.CreateEthernet()
		case "-ethd":
			ctl.DeleteEthernet()
		case "-tap":
			ctl.CreateTap()
		case "-tapd":
			ctl.DeleteTap()
		case "-loop":
			ctl.CreateLoopback()
		case "-loopd":
			ctl.DeleteLoopback()
		case "-memif":
			ctl.CreateMemif()
		case "-memifd":
			ctl.DeleteMemif()
		case "-vxlan":
			ctl.CreateVxlan()
		case "-vxland":
			ctl.DeleteVxlan()
		case "-afpkt":
			ctl.CreateAfPacket()
		case "-afpktd":
			ctl.DeleteAfPacket()
			// Linux interfaces
		case "-veth":
			ctl.CreateVethPair()
		case "-vethd":
			ctl.DeleteVethPair()
		case "-ltap":
			ctl.CreateLinuxTap()
		case "-ltapd":
			ctl.DeleteLinuxTap()
			// STN
		case "-stn":
			ctl.CreateStn()
		case "-stnd":
			ctl.DeleteStn()
			// NAT
		case "-gnat":
			ctl.CreateGlobalNat()
		case "-gnatd":
			ctl.DeleteGlobalNat()
		case "-snat":
			ctl.CreateSNat()
		case "-snatd":
			ctl.DeleteSNat()
		case "-dnat":
			ctl.CreateDNat()
		case "-dnatd":
			ctl.DeleteDNat()
			// Bridge domains
		case "-bd":
			ctl.CreateBridgeDomain()
		case "-bdd":
			ctl.DeleteBridgeDomain()
			// FIB
		case "-fib":
			ctl.CreateFib()
		case "-fibd":
			ctl.DeleteFib()
			// L2 xConnect
		case "-xconn":
			ctl.CreateXConn()
		case "-xconnd":
			ctl.DeleteXConn()
			// VPP routes
		case "-route":
			ctl.CreateRoute()
		case "-routed":
			ctl.DeleteRoute()
			// Linux routes
		case "-lrte":
			ctl.CreateLinuxRoute()
		case "-lrted":
			ctl.DeleteLinuxRoute()
			// VPP ARP
		case "-arp":
			ctl.CreateArp()
		case "-arpd":
			ctl.DeleteArp()
		case "-prxi":
			ctl.AddProxyArpInterfaces()
		case "-prxid":
			ctl.DeleteProxyArpInterfaces()
		case "-prxr":
			ctl.AddProxyArpRanges()
		case "-prxrd":
			ctl.DeleteProxyArpRanges()
			// Linux ARP
		case "-larp":
			ctl.CreateLinuxArp()
		case "-larpd":
			ctl.DeleteLinuxArp()
			// L4 plugin
		case "-el4":
			ctl.EnableL4Features()
		case "-dl4":
			ctl.DisableL4Features()
		case "-appns":
			ctl.CreateAppNamespace()
		case "-appnsd":
			ctl.DeleteAppNamespace()
			// TXN (transaction)
		case "-txn":
			ctl.CreateTxn()
		case "-txnd":
			ctl.DeleteTxn()
			// Error reporting
		case "-errIf":
			ctl.ReportIfaceErrorState()
		case "-errBd":
			ctl.ReportBdErrorState()
		default:
			usage()
		}
	}
}

// Show command info
func usage() {
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
	buffer.WriteString("\t-route,\t-routed\t- L3 route\n\t-arp,\t-arpd\t- ARP entry\n\t-prxi,\t-prxid\t- Proxy ARP interfaces\n\t-prxr,\t-prxrd\t- Proxy ARP ranges\n")
	// Linux L3
	buffer.WriteString("\t-lrte,\t-lrted\t- Linux route\n\t-larp,\t-larpd\t- Linux ARP entry\n")
	// L4
	buffer.WriteString("\t-el4,\t-dl4\t- L4 features\n")
	buffer.WriteString("\t-appns,\t-appnsd\t- Application namespace\n\n")
	// Other
	buffer.WriteString("Other:\n\t-txn,\t-txnd\t- Transaction\n\t-errIf\t\t- Interface error state report\n\t-errBd\t\t- Bridge domain error state report\n")
	logrus.DefaultLogger().Print(buffer.String())
}
