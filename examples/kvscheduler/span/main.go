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
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"

	linux_ifplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	linux_nsplugin "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	vpp_ifplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_ns "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

/*
	This example demonstrates usage of SPAN feature from VPPIfPlugin.
*/

func main() {
	// Set inter-dependency between VPP & Linux plugins
	vpp_ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	vpp_ifplugin.DefaultPlugin.NsPlugin = &linux_nsplugin.DefaultPlugin
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &vpp_ifplugin.DefaultPlugin

	ep := &ExamplePlugin{
		Orchestrator:  &orchestrator.DefaultPlugin,
		LinuxIfPlugin: &linux_ifplugin.DefaultPlugin,
		VPPIfPlugin:   &vpp_ifplugin.DefaultPlugin,
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
	VPPIfPlugin   *vpp_ifplugin.IfPlugin
	Orchestrator  *orchestrator.Plugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "span-example"
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
	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC ===")

	txn := localclient.DataResyncRequest("span-example")
	err := txn.
		LinuxInterface(hostLinuxTap).
		LinuxInterface(clientLinuxTap).
		VppInterface(hostVPPTap).
		VppInterface(clientVPPTap).
		Span(spanRx).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second * 20)
	fmt.Println("=== CHANGE ===")
	txn2 := localclient.DataChangeRequest("span-change")
	err = txn2.Delete().Span(spanRx).Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second * 20)
	err = txn2.Put().Span(spanBoth).Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(time.Second * 20)
	err = txn2.Put().Span(spanRx).Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

}

var (
	/* host <-> VPP */

	hostLinuxTap = &linux_interfaces.Interface{
		Name:    "linux_span_tap1",
		Type:    linux_interfaces.Interface_TAP_TO_VPP,
		Enabled: true,
		IpAddresses: []string{
			"10.10.1.1/24",
		},
		HostIfName: "linux_span_tap1",
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: "vpp_span_tap1",
			},
		},
	}
	hostVPPTap = &vpp_interfaces.Interface{
		Name:    "vpp_span_tap1",
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		IpAddresses: []string{
			"10.10.1.2/24",
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version: 2,
			},
		},
	}

	/* host <-> VPP */

	clientLinuxTap = &linux_interfaces.Interface{
		Name:    "linux_span_tap2",
		Type:    linux_interfaces.Interface_TAP_TO_VPP,
		Enabled: true,
		IpAddresses: []string{
			"10.20.1.1/24",
		},
		HostIfName: "linux_span_tap2",
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: "vpp_span_tap2",
			},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: "microservice-client",
		},
	}
	clientVPPTap = &vpp_interfaces.Interface{
		Name:    "vpp_span_tap2",
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		IpAddresses: []string{
			"10.20.1.2/24",
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: "microservice-client",
			},
		},
	}

	spanRx = &vpp_interfaces.Span{
		InterfaceFrom: "vpp_span_tap1",
		InterfaceTo:   "vpp_span_tap2",
		Direction:     vpp_interfaces.Span_RX,
	}

	spanBoth = &vpp_interfaces.Span{
		InterfaceFrom: "vpp_span_tap1",
		InterfaceTo:   "vpp_span_tap2",
		Direction:     vpp_interfaces.Span_BOTH,
	}
)
