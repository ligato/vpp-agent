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
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/db/keyval/etcd"

	"github.com/ligato/vpp-agent/clientv2/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/kvscheduler"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	linux_l3plugin "github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin"
	linux_interfaces "github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
	linux_l3 "github.com/ligato/vpp-agent/plugins/linuxv2/model/l3"
	linux_ns "github.com/ligato/vpp-agent/plugins/linuxv2/model/namespace"
	vpp_ifplugin "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	vpp_natplugin "github.com/ligato/vpp-agent/plugins/vppv2/natplugin"
	vpp_interfaces "github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	vpp_nat "github.com/ligato/vpp-agent/plugins/vppv2/model/nat"
)

/*
	This example demonstrates natplugin v.2

	Deploy microservices with servers:

	host-term1$ docker run -it --rm  -e MICROSERVICE_LABEL=microservice-server1 lencomilan/ubuntu /bin/bash
	host-term1$ nc -l -p 8080 &
	host-term1$ nc -u -l -p 9090 &

	host-term2$ docker run -it --rm  -e MICROSERVICE_LABEL=microservice-server2 lencomilan/ubuntu /bin/bash
	host-term2$ nc -l -p 8081 &
	host-term2$ nc -u -l -p 9091 &

	Test DNATs from microservice-client:

	host-term3$ docker run -it --rm  -e MICROSERVICE_LABEL=microservice-client lencomilan/ubuntu /bin/bash
	# TCP Service:
	host-term3$ nc 10.36.10.1 80
	host-term3$ nc 10.36.10.2 80
	host-term3$ nc 10.36.10.3 80
	# UDP Service:
	host-term3$ nc -u 10.36.10.10 90
	host-term3$ nc -u 10.36.10.11 90
	host-term3$ nc -u 10.36.10.12 90

	Run server in the host:

	host-term4$ nc -l -p 8080 &

	# Accessing server 192.168.13.10:8080 running in the host should trigger
	# source-NAT in the post-routing, i.e. no need to route microservices from the host:
	host-term3$ nc 192.168.13.10 8080  # host-term3 = microservice-client
*/

func main() {
	etcdDataSync := kvdbsync.NewPlugin(kvdbsync.UseDeps(func(deps *kvdbsync.Deps) {
		deps.KvPlugin = &etcd.DefaultPlugin
	}))

	watchers := datasync.KVProtoWatchers{
		local.Get(),
		etcdDataSync,
	}

	// Set watcher for KVScheduler.
	kvscheduler.DefaultPlugin.Watcher = watchers

	vppIfPlugin := vpp_ifplugin.NewPlugin()

	linuxIfPlugin := linux_ifplugin.NewPlugin(
		linux_ifplugin.UseDeps(func(deps *linux_ifplugin.Deps) {
			deps.VppIfPlugin = vppIfPlugin
		}),
	)

	vppIfPlugin.LinuxIfPlugin = linuxIfPlugin

	linuxL3Plugin := linux_l3plugin.NewPlugin(
		linux_l3plugin.UseDeps(func(deps *linux_l3plugin.Deps) {
			deps.IfPlugin = linuxIfPlugin
		}),
	)

	vppNATPlugin := vpp_natplugin.NewPlugin(
		vpp_natplugin.UseDeps(func(deps *vpp_natplugin.Deps) {
			deps.IfPlugin = vppIfPlugin
		}),
	)

	ep := &ExamplePlugin{
		Scheduler:     &kvscheduler.DefaultPlugin,
		LinuxIfPlugin: linuxIfPlugin,
		LinuxL3Plugin: linuxL3Plugin,
		VPPIfPlugin:   vppIfPlugin,
		VPPNATPlugin:  vppNATPlugin,
		Datasync:      etcdDataSync,
	}

	a := agent.NewAgent(
		agent.AllPlugins(ep),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// PluginName is a constant with name of main plugin.
const PluginName = "example"

// ExamplePlugin is the main plugin which
// handles resync and changes in this example.
type ExamplePlugin struct {
	Scheduler     *kvscheduler.Scheduler
	LinuxIfPlugin *linux_ifplugin.IfPlugin
	LinuxL3Plugin *linux_l3plugin.L3Plugin
	VPPIfPlugin   *vpp_ifplugin.IfPlugin
	VPPNATPlugin  *vpp_natplugin.NATPlugin
	Datasync      *kvdbsync.Plugin
}

// String returns plugin name
func (plugin *ExamplePlugin) String() string {
	return PluginName
}

// Init handles initialization phase.
func (plugin *ExamplePlugin) Init() error {
	return nil
}

// AfterInit handles phase after initialization.
func (plugin *ExamplePlugin) AfterInit() error {
	go plugin.testLocalClientWithScheduler()
	return nil
}

// Close cleans up the resources.
func (plugin *ExamplePlugin) Close() error {
	return nil
}

func (plugin *ExamplePlugin) testLocalClientWithScheduler() {
	const (
		mycroserviceClient           = "microservice-client"
		microserviceClientNetPrefix  = "10.11.1."
		mycroserviceServer1          = "microservice-server1"
		microserviceServer1NetPrefix = "10.11.2."
		mycroserviceServer2          = "microservice-server2"
		microserviceServer2NetPrefix = "10.11.3."
		microserviceNetMask = "/30"

		hostNetPrefix = "192.168.13."
		hostNetMask = "/24"

		vppTapHostLogicalName = "vpp-tap-host"
		vppTapHostIPAddr      = hostNetPrefix + "10"
		vppTapHostVersion     = 2

		vppTapClientLogicalName = "vpp-tap-client"
		vppTapClientIPAddr      = microserviceClientNetPrefix + "1"
		vppTapClientVersion     = 2

		vppTapServer1LogicalName = "vpp-tap-server1"
		vppTapServer1IPAddr      = microserviceServer1NetPrefix + "1"
		vppTapServer1Version     = 1

		vppTapServer2LogicalName = "vpp-tap-server2"
		vppTapServer2IPAddr      = microserviceServer2NetPrefix + "1"
		vppTapServer2Version     = 1

		linuxTapHostLogicalName = "linux-tap-host"
		linuxTapHostIPAddr      = hostNetPrefix + "20"

		linuxTapClientLogicalName = "linux-tap-client"
		linuxTapClientIPAddr      = microserviceClientNetPrefix + "2"

		linuxTapServer1LogicalName = "linux-tap-server1"
		linuxTapServer1IPAddr      = microserviceServer1NetPrefix + "2"

		linuxTapServer2LogicalName = "linux-tap-server2"
		linuxTapServer2IPAddr      = microserviceServer2NetPrefix + "2"

		linuxTapHostName    = "tap_to_vpp"

		serviceNetPrefix = "10.36.10."
		serviceNetMask = "/24"

		tcpServiceLabel = "tcp-service"
		tcpServiceExternalIP1 = serviceNetPrefix + "1"
		tcpServiceExternalIP2 = serviceNetPrefix + "2"
		tcpServiceExternalIP3 = serviceNetPrefix + "3"
		tcpServiceExternalPort = 80
		tcpServiceLocalPortServer1 = 8080
		tcpServiceLocalPortServer2 = 8081

		udpServiceLabel = "udp-service"
		udpServiceExternalIP1 = serviceNetPrefix + "10"
		udpServiceExternalIP2 = serviceNetPrefix + "11"
		udpServiceExternalIP3 = serviceNetPrefix + "12"
		udpServiceExternalPort = 90
		udpServiceLocalPortServer1 = 9090
		udpServiceLocalPortServer2 = 9091

		idDNATLabel = "id-dnat"
		idDNATPort = 7777

		extIfaceDNATLabel = "external-interfaces"
		extIfaceDNATExternalPort = 3333
		extIfaceDNATLocalPort = 4444

		natPoolAddr1 = hostNetPrefix + "100"
		natPoolAddr2 = hostNetPrefix + "200"
	)

	/* host <-> VPP */
	hostLinuxTap := &linux_interfaces.LinuxInterface{
		Name:        linuxTapHostLogicalName,
		Type:        linux_interfaces.LinuxInterface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{
			linuxTapHostIPAddr + hostNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.LinuxInterface_Tap{
			Tap: &linux_interfaces.LinuxInterface_TapLink{
				VppTapIfName: vppTapHostLogicalName,
			},
		},
	}

	hostVPPTap := &vpp_interfaces.Interface{
		Name:        vppTapHostLogicalName,
		Type:        vpp_interfaces.Interface_TAP_INTERFACE,
		Enabled:     true,
		IpAddresses: []string{
			vppTapHostIPAddr + hostNetMask,
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.Interface_TapLink{
				Version:        vppTapHostVersion,
			},
		},
	}

	hostRouteToServices := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapHostLogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        serviceNetPrefix + "0" + serviceNetMask,
		GwAddr:            vppTapHostIPAddr,
	}

	/* microservice-client <-> VPP */
	clientLinuxTap := &linux_interfaces.LinuxInterface{
		Name:        linuxTapClientLogicalName,
		Type:        linux_interfaces.LinuxInterface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{
			linuxTapClientIPAddr + microserviceNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.LinuxInterface_Tap{
			Tap: &linux_interfaces.LinuxInterface_TapLink{
				VppTapIfName: vppTapClientLogicalName,
			},
		},
		Namespace: &linux_ns.LinuxNetNamespace{
			Type:      linux_ns.LinuxNetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroserviceClient,
		},
	}

	clientVPPTap := &vpp_interfaces.Interface{
		Name:        vppTapClientLogicalName,
		Type:        vpp_interfaces.Interface_TAP_INTERFACE,
		Enabled:     true,
		IpAddresses: []string{
			vppTapClientIPAddr + microserviceNetMask,
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.Interface_TapLink{
				Version:        vppTapClientVersion,
				ToMicroservice: mycroserviceClient,
			},
		},
	}

	clientRouteToServices := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapClientLogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        serviceNetPrefix + "0" + serviceNetMask,
		GwAddr:            vppTapClientIPAddr,
	}

	clientRouteToHost := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapClientLogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        hostNetPrefix + "0" + hostNetMask,
		GwAddr:            vppTapClientIPAddr,
	}

	/* microservice-server1 <-> VPP */
	server1LinuxTap := &linux_interfaces.LinuxInterface{
		Name:        linuxTapServer1LogicalName,
		Type:        linux_interfaces.LinuxInterface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{
			linuxTapServer1IPAddr + microserviceNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.LinuxInterface_Tap{
			Tap: &linux_interfaces.LinuxInterface_TapLink{
				VppTapIfName: vppTapServer1LogicalName,
			},
		},
		Namespace: &linux_ns.LinuxNetNamespace{
			Type:      linux_ns.LinuxNetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroserviceServer1,
		},
	}

	server1VPPTap := &vpp_interfaces.Interface{
		Name:        vppTapServer1LogicalName,
		Type:        vpp_interfaces.Interface_TAP_INTERFACE,
		Enabled:     true,
		IpAddresses: []string{
			vppTapServer1IPAddr + microserviceNetMask,
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.Interface_TapLink{
				Version:        vppTapServer1Version,
				ToMicroservice: mycroserviceServer1,
			},
		},
	}

	server1RouteToServices := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapServer1LogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        serviceNetPrefix + "0" + serviceNetMask,
		GwAddr:            vppTapServer1IPAddr,
	}

	server1RouteToHost := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapServer1LogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        hostNetPrefix + "0" + hostNetMask,
		GwAddr:            vppTapServer1IPAddr,
	}

	server1RouteToClient := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapServer1LogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        linuxTapClientIPAddr + "/32",
		GwAddr:            vppTapServer1IPAddr,
	}

	/* microservice-server2 <-> VPP */
	server2LinuxTap := &linux_interfaces.LinuxInterface{
		Name:        linuxTapServer2LogicalName,
		Type:        linux_interfaces.LinuxInterface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{
			linuxTapServer2IPAddr + microserviceNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.LinuxInterface_Tap{
			Tap: &linux_interfaces.LinuxInterface_TapLink{
				VppTapIfName: vppTapServer2LogicalName,
			},
		},
		Namespace: &linux_ns.LinuxNetNamespace{
			Type:      linux_ns.LinuxNetNamespace_NETNS_REF_MICROSERVICE,
			Reference: mycroserviceServer2,
		},
	}

	server2VPPTap := &vpp_interfaces.Interface{
		Name:        vppTapServer2LogicalName,
		Type:        vpp_interfaces.Interface_TAP_INTERFACE,
		Enabled:     true,
		IpAddresses: []string{
			vppTapServer2IPAddr + microserviceNetMask,
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.Interface_TapLink{
				Version:        vppTapServer2Version,
				ToMicroservice: mycroserviceServer2,
			},
		},
	}

	server2RouteToServices := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapServer2LogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        serviceNetPrefix + "0" + serviceNetMask,
		GwAddr:            vppTapServer2IPAddr,
	}

	server2RouteToHost := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapServer2LogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        hostNetPrefix + "0" + hostNetMask,
		GwAddr:            vppTapServer2IPAddr,
	}

	server2RouteToClient := &linux_l3.LinuxStaticRoute{
		OutgoingInterface: linuxTapServer2LogicalName,
		Scope:             linux_l3.LinuxStaticRoute_GLOBAL,
		DstNetwork:        linuxTapClientIPAddr + "/32",
		GwAddr:            vppTapServer2IPAddr,
	}

	/* NAT44 global config */

	natGlobal := &vpp_nat.Nat44Global{
		Forwarding:    true,
		VirtualReassembly: &vpp_nat.VirtualReassembly{
			Timeout:         4,
			MaxReassemblies: 2048,
			MaxFragments:    10,
			DropFragments:   true,
		},
		NatInterfaces: []*vpp_nat.Nat44Global_Interface{
			{
				Name:          vppTapHostLogicalName,
				IsInside:      false,
				OutputFeature: true,
			},
			{
				Name:          vppTapClientLogicalName,
				IsInside:      false,
				OutputFeature: false,
			},
			{
				Name:          vppTapClientLogicalName,
				IsInside:      true, // just to test in & out together
				OutputFeature: false,
			},
			{
				Name:          vppTapServer1LogicalName,
				IsInside:      true,
				OutputFeature: false,
			},
			{
				Name:          vppTapServer2LogicalName,
				IsInside:      true,
				OutputFeature: false,
			},
		},
		AddressPool: []*vpp_nat.Nat44Global_Address{
			{
				Address: natPoolAddr1,
			},
			{
				Address:  natPoolAddr2,
				TwiceNat: true,
			},
		},
	}

	/* TCP service */

	tcpServiceDNAT := &vpp_nat.DNat44{
		Label: tcpServiceLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			{
				ExternalIp:   tcpServiceExternalIP1, // with LB
				ExternalPort: tcpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_TCP,
				LocalIps:     []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     linuxTapServer1IPAddr,
						LocalPort:   tcpServiceLocalPortServer1,
						Probability: 1,
					},
					{
						LocalIp:     linuxTapServer2IPAddr,
						LocalPort:   tcpServiceLocalPortServer2,
						Probability: 2, /* twice more likely */
					},
				},
			},
			{
				ExternalIp:   tcpServiceExternalIP2, // server 1 only
				ExternalPort: tcpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_TCP,
				LocalIps:     []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     linuxTapServer1IPAddr,
						LocalPort:   tcpServiceLocalPortServer1,
					},
				},
			},
			{
				ExternalIp:   tcpServiceExternalIP3, // server 2 only
				ExternalPort: tcpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_TCP,
				LocalIps:     []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     linuxTapServer2IPAddr,
						LocalPort:   tcpServiceLocalPortServer2,
					},
				},
			},
		},
	}

	/* UDP service */

	udpServiceDNAT := &vpp_nat.DNat44{
		Label: udpServiceLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			{
				ExternalIp:   udpServiceExternalIP1, // with LB
				ExternalPort: udpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_UDP,
				LocalIps:     []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     linuxTapServer1IPAddr,
						LocalPort:   udpServiceLocalPortServer1,
						Probability: 1,
					},
					{
						LocalIp:     linuxTapServer2IPAddr,
						LocalPort:   udpServiceLocalPortServer2,
						Probability: 2, /* twice more likely */
					},
				},
			},
			{
				ExternalIp:   udpServiceExternalIP2, // server 1 only
				ExternalPort: udpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_UDP,
				LocalIps:     []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     linuxTapServer1IPAddr,
						LocalPort:   udpServiceLocalPortServer1,
					},
				},
			},
			{
				ExternalIp:   udpServiceExternalIP3, // server 2 only
				ExternalPort: udpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_UDP,
				LocalIps:     []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     linuxTapServer2IPAddr,
						LocalPort:   udpServiceLocalPortServer2,
					},
				},
			},
		},
	}

	/* identity mapping */

	idDNAT := &vpp_nat.DNat44{
		Label: idDNATLabel,
		IdMappings: []*vpp_nat.DNat44_IdentityMapping{
			{
				Interface: vppTapClientLogicalName,
				Port:      idDNATPort,
				Protocol:  vpp_nat.DNat44_TCP,
			},
			{
				IpAddress:         natPoolAddr2,
				IpAddressFromPool: true,
				Port:              idDNATPort,
				Protocol:          vpp_nat.DNat44_TCP,
			},
		},
	}

	/* DNAT with external interfaces */

	externalIfaceDNAT := &vpp_nat.DNat44{
		Label: extIfaceDNATLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			{
				ExternalInterface: vppTapServer1LogicalName,
				ExternalPort:      extIfaceDNATExternalPort,
				Protocol:          vpp_nat.DNat44_TCP,
				LocalIps:          []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     linuxTapServer1IPAddr,
						LocalPort:   extIfaceDNATLocalPort,
					},
				},
			},
			{
				ExternalInterface: vppTapServer2LogicalName,
				ExternalPort:      extIfaceDNATExternalPort,
				Protocol:          vpp_nat.DNat44_TCP,
				LocalIps:          []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:     linuxTapServer2IPAddr,
						LocalPort:   extIfaceDNATLocalPort,
					},
				},
			},
		},
	}

	// resync

	time.Sleep(time.Second * 2)
	fmt.Println("=== RESYNC 0 ===")

	txn := localclient.DataResyncRequest("example")
	err := txn.
		LinuxInterface(hostLinuxTap).
		LinuxInterface(clientLinuxTap).
		LinuxInterface(server1LinuxTap).
		LinuxInterface(server2LinuxTap).
		LinuxRoute(hostRouteToServices).
		LinuxRoute(clientRouteToServices).
		LinuxRoute(clientRouteToHost).
		LinuxRoute(server1RouteToServices).
		LinuxRoute(server1RouteToHost).
		LinuxRoute(server1RouteToClient).
		LinuxRoute(server2RouteToServices).
		LinuxRoute(server2RouteToHost).
		LinuxRoute(server2RouteToClient).
		VppInterface(hostVPPTap).
		VppInterface(clientVPPTap).
		VppInterface(server1VPPTap).
		VppInterface(server2VPPTap).
		NAT44Global(natGlobal).
		DNAT44(tcpServiceDNAT).
		DNAT44(udpServiceDNAT).
		DNAT44(idDNAT).
		DNAT44(externalIfaceDNAT).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}


	// data change

	/*
	time.Sleep(time.Second * 10)
	fmt.Println("=== CHANGE 1 ===")

	txn2 := localclient.DataChangeRequest("example")
	err = txn2.Put().
		Delete().
		NAT44Global().
		DNAT44(udpServiceDNAT.Label).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}
	*/
}
