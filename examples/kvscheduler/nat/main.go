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

	"go.ligato.io/cn-infra/v2/agent"

	"go.ligato.io/vpp-agent/v3/clientv2/linux/localclient"
	linux_ifplugin "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin"
	linux_l3plugin "go.ligato.io/vpp-agent/v3/plugins/linux/l3plugin"
	linux_nsplugin "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	vpp_ifplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	vpp_natplugin "go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	linux_ns "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
)

/*
	This example demonstrates natplugin v.2

	Add this config stanza to the VPP startup configuration:

	nat {
		endpoint-dependent
	}

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
	host-term3$ nc 192.168.13.20 8080  # host-term3 = microservice-client
*/

func main() {
	// Set inter-dependency between VPP & Linux plugins
	vpp_ifplugin.DefaultPlugin.LinuxIfPlugin = &linux_ifplugin.DefaultPlugin
	vpp_ifplugin.DefaultPlugin.NsPlugin = &linux_nsplugin.DefaultPlugin
	linux_ifplugin.DefaultPlugin.VppIfPlugin = &vpp_ifplugin.DefaultPlugin

	ep := &ExamplePlugin{
		Orchestrator:  &orchestrator.DefaultPlugin,
		LinuxIfPlugin: &linux_ifplugin.DefaultPlugin,
		LinuxL3Plugin: &linux_l3plugin.DefaultPlugin,
		VPPIfPlugin:   &vpp_ifplugin.DefaultPlugin,
		VPPNATPlugin:  &vpp_natplugin.DefaultPlugin,
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
	VPPNATPlugin  *vpp_natplugin.NATPlugin
	Orchestrator  *orchestrator.Plugin
}

// String returns plugin name
func (p *ExamplePlugin) String() string {
	return "vpp-nat-example"
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
	fmt.Println("=== RESYNC ===")

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
		NAT44Interface(natInterfaceTapHost).
		NAT44Interface(natInterfaceTapClient).
		NAT44Interface(natInterfaceTapServer1).
		NAT44Interface(natInterfaceTapServer2).
		NAT44AddressPool(natPool1).
		NAT44AddressPool(natPool2).
		DNAT44(tcpServiceDNAT).
		DNAT44(udpServiceDNAT).
		DNAT44(idDNAT).
		DNAT44(externalIfaceDNAT).
		DNAT44(emptyDNAT).
		DNAT44(addrFromPoolDNAT).
		Send().ReceiveReply()
	if err != nil {
		fmt.Println(err)
		return
	}

	// data change
	/* UNCOMMENT TO TEST THE CONFIG CHANGE
	time.Sleep(time.Second * 10)
	fmt.Println("=== CHANGE ===")

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

const (
	mycroserviceClient           = "microservice-client"
	microserviceClientNetPrefix  = "10.11.1."
	mycroserviceServer1          = "microservice-server1"
	microserviceServer1NetPrefix = "10.11.2."
	mycroserviceServer2          = "microservice-server2"
	microserviceServer2NetPrefix = "10.11.3."
	microserviceNetMask          = "/30"

	hostNetPrefix = "192.168.13."
	hostNetMask   = "/24"

	vppTapHostLogicalName = "vpp-tap-host"
	vppTapHostIPAddr      = hostNetPrefix + "10"
	vppTapHostVersion     = 2

	vppTapClientLogicalName = "vpp-tap-client"
	vppTapClientIPAddr      = microserviceClientNetPrefix + "1"
	vppTapClientVersion     = 2

	vppTapServer1LogicalName = "vpp-tap-server1"
	vppTapServer1IPAddr      = microserviceServer1NetPrefix + "1"
	vppTapServer1Version     = 2

	vppTapServer2LogicalName = "vpp-tap-server2"
	vppTapServer2IPAddr      = microserviceServer2NetPrefix + "1"
	vppTapServer2Version     = 2

	linuxTapHostLogicalName = "linux-tap-host"
	linuxTapHostIPAddr      = hostNetPrefix + "20"

	linuxTapClientLogicalName = "linux-tap-client"
	linuxTapClientIPAddr      = microserviceClientNetPrefix + "2"

	linuxTapServer1LogicalName = "linux-tap-server1"
	linuxTapServer1IPAddr      = microserviceServer1NetPrefix + "2"

	linuxTapServer2LogicalName = "linux-tap-server2"
	linuxTapServer2IPAddr      = microserviceServer2NetPrefix + "2"

	linuxTapHostName = "tap_to_vpp"

	serviceNetPrefix = "10.36.10."
	serviceNetMask   = "/24"

	tcpServiceLabel            = "tcp-service"
	tcpServiceExternalIP1      = serviceNetPrefix + "1"
	tcpServiceExternalIP2      = serviceNetPrefix + "2"
	tcpServiceExternalIP3      = serviceNetPrefix + "3"
	tcpServiceExternalPort     = 80
	tcpServiceLocalPortServer1 = 8080
	tcpServiceLocalPortServer2 = 8081

	udpServiceLabel            = "udp-service"
	udpServiceExternalIP1      = serviceNetPrefix + "10"
	udpServiceExternalIP2      = serviceNetPrefix + "11"
	udpServiceExternalIP3      = serviceNetPrefix + "12"
	udpServiceExternalPort     = 90
	udpServiceLocalPortServer1 = 9090
	udpServiceLocalPortServer2 = 9091

	idDNATLabel = "id-dnat"
	idDNATPort  = 7777

	extIfaceDNATLabel        = "external-interfaces"
	extIfaceDNATExternalPort = 3333
	extIfaceDNATLocalPort    = 4444

	addrFromPoolDNATLabel = "external-address-from-pool"
	addrFromPoolDNATPort  = 6000

	emptyDNATLabel = "empty-dnat"

	natPoolAddr1 = hostNetPrefix + "100"
	natPoolAddr2 = hostNetPrefix + "101"
	natPoolAddr3 = hostNetPrefix + "250"
)

var (
	/* host <-> VPP */

	hostLinuxTap = &linux_interfaces.Interface{
		Name:    linuxTapHostLogicalName,
		Type:    linux_interfaces.Interface_TAP_TO_VPP,
		Enabled: true,
		IpAddresses: []string{
			linuxTapHostIPAddr + hostNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapHostLogicalName,
			},
		},
	}
	hostVPPTap = &vpp_interfaces.Interface{
		Name:    vppTapHostLogicalName,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		IpAddresses: []string{
			vppTapHostIPAddr + hostNetMask,
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version: vppTapHostVersion,
			},
		},
	}
	hostRouteToServices = &linux_l3.Route{
		OutgoingInterface: linuxTapHostLogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        serviceNetPrefix + "0" + serviceNetMask,
		GwAddr:            vppTapHostIPAddr,
	}

	/* microservice-client <-> VPP */

	clientLinuxTap = &linux_interfaces.Interface{
		Name:    linuxTapClientLogicalName,
		Type:    linux_interfaces.Interface_TAP_TO_VPP,
		Enabled: true,
		IpAddresses: []string{
			linuxTapClientIPAddr + microserviceNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapClientLogicalName,
			},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: mycroserviceClient,
		},
	}
	clientVPPTap = &vpp_interfaces.Interface{
		Name:    vppTapClientLogicalName,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		IpAddresses: []string{
			vppTapClientIPAddr + microserviceNetMask,
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        vppTapClientVersion,
				ToMicroservice: mycroserviceClient,
			},
		},
	}
	clientRouteToServices = &linux_l3.Route{
		OutgoingInterface: linuxTapClientLogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        serviceNetPrefix + "0" + serviceNetMask,
		GwAddr:            vppTapClientIPAddr,
	}
	clientRouteToHost = &linux_l3.Route{
		OutgoingInterface: linuxTapClientLogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        hostNetPrefix + "0" + hostNetMask,
		GwAddr:            vppTapClientIPAddr,
	}

	/* microservice-server1 <-> VPP */

	server1LinuxTap = &linux_interfaces.Interface{
		Name:    linuxTapServer1LogicalName,
		Type:    linux_interfaces.Interface_TAP_TO_VPP,
		Enabled: true,
		IpAddresses: []string{
			linuxTapServer1IPAddr + microserviceNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapServer1LogicalName,
			},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: mycroserviceServer1,
		},
	}
	server1VPPTap = &vpp_interfaces.Interface{
		Name:    vppTapServer1LogicalName,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		IpAddresses: []string{
			vppTapServer1IPAddr + microserviceNetMask,
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        vppTapServer1Version,
				ToMicroservice: mycroserviceServer1,
			},
		},
	}
	server1RouteToServices = &linux_l3.Route{
		OutgoingInterface: linuxTapServer1LogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        serviceNetPrefix + "0" + serviceNetMask,
		GwAddr:            vppTapServer1IPAddr,
	}
	server1RouteToHost = &linux_l3.Route{
		OutgoingInterface: linuxTapServer1LogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        hostNetPrefix + "0" + hostNetMask,
		GwAddr:            vppTapServer1IPAddr,
	}
	server1RouteToClient = &linux_l3.Route{
		OutgoingInterface: linuxTapServer1LogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        linuxTapClientIPAddr + "/32",
		GwAddr:            vppTapServer1IPAddr,
	}

	/* microservice-server2 <-> VPP */
	server2LinuxTap = &linux_interfaces.Interface{
		Name:    linuxTapServer2LogicalName,
		Type:    linux_interfaces.Interface_TAP_TO_VPP,
		Enabled: true,
		IpAddresses: []string{
			linuxTapServer2IPAddr + microserviceNetMask,
		},
		HostIfName: linuxTapHostName,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapServer2LogicalName,
			},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: mycroserviceServer2,
		},
	}
	server2VPPTap = &vpp_interfaces.Interface{
		Name:    vppTapServer2LogicalName,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		IpAddresses: []string{
			vppTapServer2IPAddr + microserviceNetMask,
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        vppTapServer2Version,
				ToMicroservice: mycroserviceServer2,
			},
		},
	}
	server2RouteToServices = &linux_l3.Route{
		OutgoingInterface: linuxTapServer2LogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        serviceNetPrefix + "0" + serviceNetMask,
		GwAddr:            vppTapServer2IPAddr,
	}
	server2RouteToHost = &linux_l3.Route{
		OutgoingInterface: linuxTapServer2LogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        hostNetPrefix + "0" + hostNetMask,
		GwAddr:            vppTapServer2IPAddr,
	}
	server2RouteToClient = &linux_l3.Route{
		OutgoingInterface: linuxTapServer2LogicalName,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        linuxTapClientIPAddr + "/32",
		GwAddr:            vppTapServer2IPAddr,
	}

	/* NAT44 global config */

	natGlobal = &vpp_nat.Nat44Global{
		Forwarding: true,
		VirtualReassembly: &vpp_nat.VirtualReassembly{
			Timeout:         4,
			MaxReassemblies: 2048,
			MaxFragments:    10,
			DropFragments:   true,
		},
	}

	/* NAT interfaces */

	natInterfaceTapHost = &vpp_nat.Nat44Interface{
		Name:          vppTapHostLogicalName,
		NatOutside:    true,
		OutputFeature: true,
	}
	natInterfaceTapClient = &vpp_nat.Nat44Interface{
		Name:       vppTapClientLogicalName,
		NatInside:  true, // just to test in & out together
		NatOutside: true,
	}
	natInterfaceTapServer1 = &vpp_nat.Nat44Interface{
		Name:      vppTapServer1LogicalName,
		NatInside: true,
	}
	natInterfaceTapServer2 = &vpp_nat.Nat44Interface{
		Name:      vppTapServer2LogicalName,
		NatInside: true,
	}

	/* NAT pools */

	natPool1 = &vpp_nat.Nat44AddressPool{
		FirstIp: natPoolAddr1,
		LastIp:  natPoolAddr2,
	}
	natPool2 = &vpp_nat.Nat44AddressPool{
		FirstIp:  natPoolAddr3,
		TwiceNat: true,
	}

	/* TCP service */

	tcpServiceDNAT = &vpp_nat.DNat44{
		Label: tcpServiceLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			{
				ExternalIp:   tcpServiceExternalIP1, // with LB
				ExternalPort: tcpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_TCP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
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
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTapServer1IPAddr,
						LocalPort: tcpServiceLocalPortServer1,
					},
				},
			},
			{
				ExternalIp:   tcpServiceExternalIP3, // server 2 only
				ExternalPort: tcpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_TCP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTapServer2IPAddr,
						LocalPort: tcpServiceLocalPortServer2,
					},
				},
			},
		},
	}

	/* UDP service */

	udpServiceDNAT = &vpp_nat.DNat44{
		Label: udpServiceLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			{
				ExternalIp:   udpServiceExternalIP1, // with LB
				ExternalPort: udpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_UDP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
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
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTapServer1IPAddr,
						LocalPort: udpServiceLocalPortServer1,
					},
				},
			},
			{
				ExternalIp:   udpServiceExternalIP3, // server 2 only
				ExternalPort: udpServiceExternalPort,
				Protocol:     vpp_nat.DNat44_UDP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTapServer2IPAddr,
						LocalPort: udpServiceLocalPortServer2,
					},
				},
			},
		},
	}

	/* identity mapping */

	idDNAT = &vpp_nat.DNat44{
		Label: idDNATLabel,
		IdMappings: []*vpp_nat.DNat44_IdentityMapping{
			{
				Interface: vppTapClientLogicalName,
				Port:      idDNATPort,
				Protocol:  vpp_nat.DNat44_TCP,
			},
			{
				IpAddress: natPoolAddr2,
				Port:      idDNATPort,
				Protocol:  vpp_nat.DNat44_TCP,
			},
		},
	}

	/* DNAT with external interfaces */

	externalIfaceDNAT = &vpp_nat.DNat44{
		Label: extIfaceDNATLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			{
				ExternalInterface: vppTapServer1LogicalName,
				ExternalPort:      extIfaceDNATExternalPort,
				Protocol:          vpp_nat.DNat44_TCP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTapServer1IPAddr,
						LocalPort: extIfaceDNATLocalPort,
					},
				},
			},
			{
				ExternalInterface: vppTapServer2LogicalName,
				ExternalPort:      extIfaceDNATExternalPort,
				Protocol:          vpp_nat.DNat44_TCP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTapServer2IPAddr,
						LocalPort: extIfaceDNATLocalPort,
					},
				},
			},
		},
	}

	/* empty DNAT */
	emptyDNAT = &vpp_nat.DNat44{
		Label: emptyDNATLabel,
	}

	/* DNAT with address from the pool */

	addrFromPoolDNAT = &vpp_nat.DNat44{
		Label: addrFromPoolDNATLabel,
		StMappings: []*vpp_nat.DNat44_StaticMapping{
			// Without LB
			{
				ExternalIp:   natPoolAddr1,
				ExternalPort: addrFromPoolDNATPort,
				Protocol:     vpp_nat.DNat44_TCP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTapServer1IPAddr,
						LocalPort: addrFromPoolDNATPort,
					},
				},
			},
			// With LB
			{
				ExternalIp:   natPoolAddr2,
				ExternalPort: addrFromPoolDNATPort,
				Protocol:     vpp_nat.DNat44_TCP,
				LocalIps: []*vpp_nat.DNat44_StaticMapping_LocalIP{
					{
						LocalIp:   linuxTapServer1IPAddr,
						LocalPort: addrFromPoolDNATPort,
					},
					{
						LocalIp:   linuxTapServer2IPAddr,
						LocalPort: addrFromPoolDNATPort,
					},
				},
			},
		},
	}
)
