//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ligato/cn-infra/agent"

	"go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"
	linux_ifplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	linux_l3plugin "go.ligato.io/vpp-agent/v3/plugins/linux/l3plugin"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	vpp_ifplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	vpp_l2plugin "go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_ns "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

/*
	This example demonstrates L2 Plugin v.2.
*/

func main() {
	// Set inter-dependency between VPP & Linux plugins
	vpp_ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &vpp_ifplugin.DefaultPlugin

	ep := &ExamplePlugin{
		Orchestrator:  &orchestrator.DefaultPlugin,
		LinuxIfPlugin: &linux_ifplugin.DefaultPlugin,
		LinuxL3Plugin: &linux_l3plugin.DefaultPlugin,
		VPPIfPlugin:   &vpp_ifplugin.DefaultPlugin,
		VPPL2Plugin:   &vpp_l2plugin.DefaultPlugin,
	}

	a := agent.NewAgent(
		agent.AllPlugins(ep),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// ExamplePlugin is the main plugin which
// handles resync and changes in this example.
type ExamplePlugin struct {
	LinuxIfPlugin *linux_ifplugin.IfPlugin
	LinuxL3Plugin *linux_l3plugin.L3Plugin
	VPPIfPlugin   *vpp_ifplugin.IfPlugin
	VPPL2Plugin   *vpp_l2plugin.L2Plugin
	Orchestrator  *orchestrator.Plugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "l2-example"
}

// Init handles initialization phase.
func (p *ExamplePlugin) Init() error {
	return nil
}

// AfterInit handles phase after initialization.
func (p *ExamplePlugin) AfterInit() error {
	go testLocalClientWithScheduler()
	return nil
}

// Close cleans up the resources.
func (p *ExamplePlugin) Close() error {
	return nil
}

func testLocalClientWithScheduler() {
	// initial resync
	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC (using bridge domain) ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		LinuxInterface(veth2).
		LinuxInterface(veth1).
		LinuxInterface(linuxTap).
		VppInterface(afpacket).
		VppInterface(vppTap).
		VppInterface(bviLoop).
		VppInterface(loop2).
		BD(bd).
		BDFIB(fibForLoop).
		BDFIB(fibForTAP).
		BDFIB(fibForVETH).
		BDFIB(dropFIB).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data changes

	time.Sleep(time.Second * 10)
	fmt.Printf("=== CHANGE (switching to XConnect) ===\n")

	txn3 := localclient.DataChangeRequest("example")
	err = txn3.Delete().
		BD(bd.Name). // FIBs will be pending
		Put().
		XConnect(xConnectMs1ToMs2).
		XConnect(xConnectMs2ToMs1).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
}

const (
	bdNetPrefix = "10.11.1."
	bdNetMask   = "/24"

	veth1LogicalName = "myVETH1"
	veth1HostName    = "veth1"
	veth1IPAddr      = bdNetPrefix + "1"
	veth1HwAddr      = "66:66:66:66:66:66"

	veth2LogicalName = "myVETH2"
	veth2HostName    = "veth2"

	afPacketLogicalName = "myAFPacket"
	afPacketHwAddr      = "a7:35:45:55:65:75"

	vppTapLogicalName = "myVPPTap"
	vppTapHwAddr      = "b3:12:12:45:A7:B7"
	vppTapVersion     = 2

	linuxTapLogicalName = "myLinuxTAP"
	linuxTapHostName    = "tap_to_vpp"
	linuxTapIPAddr      = bdNetPrefix + "2"
	linuxTapHwAddr      = "88:88:88:88:88:88"

	mycroservice1 = "microservice1"
	mycroservice2 = "microservice2"

	bviLoopName   = "myLoopback1"
	bviLoopIP     = bdNetPrefix + "3"
	bviLoopHwAddr = "cd:cd:cd:cd:cd:cd"

	loop2Name   = "myLoopback2"
	loop2HwAddr = "ef:ef:ef:ef:ef:ef"

	bdName                = "myBridgeDomain"
	bdFlood               = true
	bdUnknownUnicastFlood = true
	bdForward             = true
	bdLearn               = false /* Learning turned off, FIBs are needed for connectivity */
	bdArpTermination      = true
	bdMacAge              = 0
)

var (
	/* microservice1 <-> VPP */

	veth1 = &linux_interfaces.Interface{
		Name:        veth1LogicalName,
		Type:        linux_interfaces.Interface_VETH,
		Enabled:     true,
		PhysAddress: veth1HwAddr,
		IpAddresses: []string{
			veth1IPAddr + bdNetMask,
		},
		HostIfName: veth1HostName,
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{PeerIfName: veth2LogicalName},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: mycroservice1,
		},
	}
	veth2 = &linux_interfaces.Interface{
		Name:       veth2LogicalName,
		Type:       linux_interfaces.Interface_VETH,
		Enabled:    true,
		HostIfName: veth2HostName,
		Link: &linux_interfaces.Interface_Veth{
			Veth: &linux_interfaces.VethLink{PeerIfName: veth1LogicalName},
		},
	}
	afpacket = &vpp_interfaces.Interface{
		Name:        afPacketLogicalName,
		Type:        vpp_interfaces.Interface_AF_PACKET,
		Enabled:     true,
		PhysAddress: afPacketHwAddr,
		Link: &vpp_interfaces.Interface_Afpacket{
			Afpacket: &vpp_interfaces.AfpacketLink{
				HostIfName: veth2HostName,
			},
		},
	}

	/* microservice2 <-> VPP */

	linuxTap = &linux_interfaces.Interface{
		Name:        linuxTapLogicalName,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		PhysAddress: linuxTapHwAddr,
		IpAddresses: []string{
			linuxTapIPAddr + bdNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapLogicalName,
			},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: mycroservice2,
		},
	}
	vppTap = &vpp_interfaces.Interface{
		Name:        vppTapLogicalName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		PhysAddress: vppTapHwAddr,
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        vppTapVersion,
				ToMicroservice: mycroservice2,
			},
		},
	}

	/* Bridge domain */

	bd = &vpp_l2.BridgeDomain{
		Name:                bdName,
		Flood:               bdFlood,
		UnknownUnicastFlood: bdUnknownUnicastFlood,
		Forward:             bdForward,
		Learn:               bdLearn,
		ArpTermination:      bdArpTermination,
		MacAge:              bdMacAge,
		Interfaces: []*vpp_l2.BridgeDomain_Interface{
			{
				Name: vppTapLogicalName,
			},
			{
				Name: afPacketLogicalName,
			},
			{
				Name:                    bviLoopName,
				BridgedVirtualInterface: true,
			},
		},
	}
	bviLoop = &vpp_interfaces.Interface{
		Name:        bviLoopName,
		Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		PhysAddress: bviLoopHwAddr,
		IpAddresses: []string{
			bviLoopIP + bdNetMask,
		},
	}
	loop2 = &vpp_interfaces.Interface{
		Name:        loop2Name,
		Type:        vpp_interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		PhysAddress: loop2HwAddr,
	}

	/* FIB entries */

	fibForLoop = &vpp_l2.FIBEntry{
		PhysAddress:             bviLoopHwAddr,
		BridgeDomain:            bdName,
		Action:                  vpp_l2.FIBEntry_FORWARD,
		OutgoingInterface:       bviLoopName,
		BridgedVirtualInterface: true,
		StaticConfig:            true,
	}
	fibForVETH = &vpp_l2.FIBEntry{
		PhysAddress:       veth1HwAddr,
		BridgeDomain:      bdName,
		Action:            vpp_l2.FIBEntry_FORWARD,
		OutgoingInterface: afPacketLogicalName,
	}
	fibForTAP = &vpp_l2.FIBEntry{
		PhysAddress:       linuxTapHwAddr,
		BridgeDomain:      bdName,
		Action:            vpp_l2.FIBEntry_FORWARD,
		OutgoingInterface: vppTapLogicalName,
	}
	dropFIB = &vpp_l2.FIBEntry{
		PhysAddress:  loop2HwAddr,
		BridgeDomain: bdName,
		Action:       vpp_l2.FIBEntry_DROP,
	}

	/* XConnect */

	xConnectMs1ToMs2 = &vpp_l2.XConnectPair{
		ReceiveInterface:  afPacketLogicalName,
		TransmitInterface: vppTapLogicalName,
	}
	xConnectMs2ToMs1 = &vpp_l2.XConnectPair{
		ReceiveInterface:  vppTapLogicalName,
		TransmitInterface: afPacketLogicalName,
	}
)
