// Copyright (c) 2018 Cisco and/or its affiliates.
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

package ifplugin

import (
	"net"
	"testing"
	"time"

	"git.fd.io/govpp.git/adapter/mock"
	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/af_packet"
	dhcp_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/dhcp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/memif"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/tap"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/tapv2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vxlan"
	if_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var ifNames = []string{"if1", "if2", "if3"}
var mtu uint32 = 1500
var netAddresses = []string{"10.0.0.1/24", "10.0.0.2/24", "192.168.50.1/24"}
var ipv6Addresses = []string{"fd21:7408:186f::/48", "2001:db8:a0b:12f0::1/48"}
var macs = []string{"46:06:18:DB:05:3A", "BC:FE:E9:5E:07:04"}

/* Interface configurator init and close */

// Test init function
func TestInterfaceConfiguratorInit(t *testing.T) {
	var err error
	// Setup
	RegisterTestingT(t)
	ctx := &vppcallmock.TestCtx{
		MockVpp: &mock.VppAdapter{},
	}
	connection, _ := core.Connect(ctx.MockVpp)
	defer connection.Disconnect()
	plugin := &InterfaceConfigurator{
		Log:      logrus.DefaultLogger(),
		GoVppmux: connection,
	}
	swIfIndices := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(plugin.Log, "swIf-test", "swIf", nil))
	dhcpIndices := ifaceidx.NewDHCPIndex(nametoidx.NewNameToIdx(plugin.Log, "dhcp-test", "dhcp", nil))
	ifVppNotifChan := make(chan govppapi.Message, 100)
	// Reply set
	ctx.MockVpp.MockReply(&memif.MemifSocketFilenameDetails{
		SocketID:       1,
		SocketFilename: []byte("test-socket-filename"),
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	// Register
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	// Test init
	err = plugin.Init(swIfIndices, dhcpIndices, mtu, ifVppNotifChan)
	Expect(err).To(BeNil())
	Expect(plugin.swIfIndexes).ToNot(BeNil())
	Expect(plugin.dhcpIndices).ToNot(BeNil())
	Expect(plugin.notifChan).ToNot(BeNil())
	Expect(plugin.mtu).To(BeEquivalentTo(mtu))
	Expect(plugin.afPacketConfigurator).ToNot(BeNil())
	Expect(plugin.memifScCache["test-socket-filename"]).To(BeEquivalentTo(1))
	// Test DHCP notifications
	dhcpIpv4 := &dhcp_api.DhcpComplEvent{
		HostAddress:   net.ParseIP(ipAddresses[0]),
		RouterAddress: net.ParseIP(ipAddresses[1]),
		HostMac: func(mac string) []byte {
			parsed, _ := net.ParseMAC(mac)
			return parsed
		}("7C:4E:E7:8A:63:68"),
		Hostname: []byte(ifNames[0]),
		IsIpv6:   0,
	}
	dhcpIpv6 := &dhcp_api.DhcpComplEvent{
		HostAddress:   net.ParseIP(ipv6Addresses[0]),
		RouterAddress: net.ParseIP(ipv6Addresses[1]),
		HostMac: func(mac string) []byte {
			parsed, err := net.ParseMAC(mac)
			Expect(err).To(BeNil())
			return parsed
		}("7C:4E:E7:8A:63:68"),
		Hostname: []byte(ifNames[1]),
		IsIpv6:   1,
	}
	plugin.dhcpChan <- dhcpIpv4
	time.Sleep(1 * time.Second)
	Eventually(func() bool {
		_, _, found := plugin.dhcpIndices.LookupIdx(ifNames[0])
		return found
	}, 2).Should(BeTrue())
	plugin.dhcpChan <- dhcpIpv6
	Eventually(func() bool {
		_, _, found := plugin.dhcpIndices.LookupIdx(ifNames[1])
		return found
	}, 2).Should(BeTrue())
	// Test close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

/* Interface configurator test cases */

// Get interface details and propagate it to status
func TestInterfaceConfiguratorPropagateIfDetailsToStatus(t *testing.T) {
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceDetails{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceDetails{
		SwIfIndex: 2,
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	// Register
	plugin.swIfIndexes.RegisterName(ifNames[0], 1, nil)
	// Do not register second interface
	// Process notifications
	done := make(chan int)
	go func() {
		var counter int
		for {
			select {
			case notification := <-plugin.notifChan:
				Expect(notification).ShouldNot(BeNil())
				counter++
			case <-time.NewTimer(1 * time.Second).C:
				done <- counter
				break
			}
		}
	}()
	// Test notifications
	err := plugin.PropagateIfDetailsToStatus()
	Expect(err).To(BeNil())
	// This blocks until the result is sent
	Expect(<-done).To(BeEquivalentTo(1))
	close(done)
}

// Configure new TAPv1 interface with IP address
func TestInterfacesConfigureTapV1(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&tap.TapConnectReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetRxModeReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_TAP_INTERFACE, append(addresses, netAddresses[0]), false, macs[0], 1500)
	data.Tap = getTestTapInterface(1, ifNames[0])
	data.RxModeSettings = getTestRxModeSettings(if_api.RxModeType_DEFAULT)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_TAP_INTERFACE))
	Expect(meta.IpAddresses).To(HaveLen(1))
	Expect(meta.IpAddresses[0]).To(BeEquivalentTo(netAddresses[0]))
	Expect(meta.PhysAddress).To(BeEquivalentTo(macs[0]))
	Expect(meta.Mtu).To(BeEquivalentTo(1500))
	Expect(meta.RxModeSettings).ToNot(BeNil())
	Expect(meta.RxModeSettings.RxMode).To(BeEquivalentTo(if_api.RxModeType_DEFAULT))
	Expect(meta.Tap).ToNot(BeNil())
	Expect(meta.Tap.Version).To(BeEquivalentTo(1))
	Expect(meta.Tap.HostIfName).To(BeEquivalentTo(ifNames[0]))
}

// Configure new TAPv2 interface without IP set as dhcp
func TestInterfacesConfigureTapV2(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&tapv2.TapCreateV2Reply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetRxModeReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&dhcp_api.DhcpClientConfigReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_TAP_INTERFACE, addresses, true, macs[0], 1500)
	data.Tap = getTestTapInterface(2, ifNames[0])
	data.RxModeSettings = getTestRxModeSettings(if_api.RxModeType_DEFAULT)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_TAP_INTERFACE))
	Expect(meta.SetDhcpClient).To(BeTrue())
	Expect(meta.PhysAddress).To(BeEquivalentTo(macs[0]))
	Expect(meta.Mtu).To(BeEquivalentTo(1500))
	Expect(meta.RxModeSettings).ToNot(BeNil())
	Expect(meta.RxModeSettings.RxMode).To(BeEquivalentTo(if_api.RxModeType_DEFAULT))
	Expect(meta.Tap).ToNot(BeNil())
	Expect(meta.Tap.Version).To(BeEquivalentTo(2))
	Expect(meta.Tap.HostIfName).To(BeEquivalentTo(ifNames[0]))
}

// Configure new memory interface without IP set unnumbered, master and without socket filename registered
func TestInterfacesConfigureMemif(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&memif.MemifSocketFilenameAddDelReply{}) // Memif socket filename registration
	ctx.MockVpp.MockReply(&memif.MemifCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetUnnumberedReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_MEMORY_INTERFACE, addresses, false, macs[0], 1500)
	data.Memif = getTestMemifInterface(true, 1)
	data.Unnumbered = getTestUnnumberedSettings(ifNames[1])
	// Register unnumbered interface
	plugin.swIfIndexes.RegisterName(ifNames[1], 2, nil)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_MEMORY_INTERFACE))
	Expect(meta.Unnumbered).ToNot(BeNil())
	Expect(meta.Unnumbered.InterfaceWithIP).To(BeEquivalentTo(ifNames[1]))
	Expect(meta.PhysAddress).To(BeEquivalentTo(macs[0]))
	Expect(meta.Mtu).To(BeEquivalentTo(1500))
	Expect(meta.Memif).ToNot(BeNil())
	Expect(meta.Memif.Master).To(BeTrue())
	Expect(meta.Memif.Id).To(BeEquivalentTo(1))
	Expect(plugin.memifScCache["socket-filename"]).To(BeEquivalentTo(0)) // Socket ID registration starts with 0
}

// Configure new memory interface without IP set unnumbered, slave and with socket filename registered
func TestInterfacesConfigureMemifAsSlave(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&memif.MemifCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})                     // Break status propagation
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetUnnumberedReply{}) // After unnumbered registration
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_MEMORY_INTERFACE, addresses, false, macs[0], 1500)
	data.Memif = getTestMemifInterface(true, 1)
	data.Unnumbered = getTestUnnumberedSettings(ifNames[1])
	// Register socket filename
	plugin.memifScCache["socket-filename"] = 0
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_MEMORY_INTERFACE))
	Expect(meta.Unnumbered).ToNot(BeNil())
	Expect(meta.Unnumbered.InterfaceWithIP).To(BeEquivalentTo(ifNames[1]))
	Expect(meta.PhysAddress).To(BeEquivalentTo(macs[0]))
	Expect(meta.Mtu).To(BeEquivalentTo(1500))
	Expect(meta.Memif).ToNot(BeNil())
	Expect(meta.Memif.Master).To(BeTrue())
	Expect(meta.Memif.Id).To(BeEquivalentTo(1))
	Expect(plugin.memifScCache["socket-filename"]).To(BeEquivalentTo(0))  // Socket ID registration starts with 0
	Expect(plugin.uIfaceCache[ifNames[0]]).To(BeEquivalentTo(ifNames[1])) // Unnumbered interface is cached
	// Register Unnumbered interface
	plugin.swIfIndexes.RegisterName(ifNames[1], 2, nil)
	plugin.resolveDependentUnnumberedInterfaces(ifNames[1], 2)
	Expect(plugin.uIfaceCache[ifNames[0]]).To(BeEmpty())
}

// Configure new VxLAN interface
func TestInterfacesConfigureVxLAN(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&vxlan.VxlanAddDelTunnelReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_VXLAN_TUNNEL, append(addresses, netAddresses[0]), false, "", 0)
	data.Vxlan = getTestVxLanInterface(ipAddresses[1], ipAddresses[2], 1)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_VXLAN_TUNNEL))
	Expect(meta.IpAddresses).To(HaveLen(1))
	Expect(meta.IpAddresses[0]).To(BeEquivalentTo(netAddresses[0]))
	Expect(meta.Vxlan).ToNot(BeNil())
	Expect(meta.Vxlan.SrcAddress).To(BeEquivalentTo(ipAddresses[1]))
	Expect(meta.Vxlan.DstAddress).To(BeEquivalentTo(ipAddresses[2]))
	Expect(meta.Vxlan.Vni).To(BeEquivalentTo(1))
}

// Configure new VxLAN interface with default MTU
func TestInterfacesConfigureLoopback(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	plugin.mtu = 2000
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.CreateLoopbackReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_SOFTWARE_LOOPBACK, append(addresses, netAddresses[0]), false, macs[0], 0)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_SOFTWARE_LOOPBACK))
	Expect(meta.IpAddresses).To(HaveLen(1))
	Expect(meta.IpAddresses[0]).To(BeEquivalentTo(netAddresses[0]))
	Expect(meta.Mtu).To(BeEquivalentTo(2000))
}

// Configure existing Ethernet interface
func TestInterfacesConfigureEthernet(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_ETHERNET_CSMACD, append(addresses, netAddresses[0]), false, macs[0], 1500)
	// Register ethernet
	plugin.swIfIndexes.RegisterName(ifNames[0], 1, nil)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_ETHERNET_CSMACD))
	Expect(meta.IpAddresses).To(HaveLen(1))
	Expect(meta.IpAddresses[0]).To(BeEquivalentTo(netAddresses[0]))
	Expect(meta.PhysAddress).To(BeEquivalentTo(macs[0]))
	Expect(meta.Mtu).To(BeEquivalentTo(1500))
}

// Configure non-existing Ethernet interface
func TestInterfacesConfigureEthernetNonExisting(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_ETHERNET_CSMACD, append(addresses, netAddresses[0]), false, macs[0], 1500)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeFalse())
	Expect(meta).To(BeNil())
}

// Configure AfPacket interface
func TestInterfacesConfigureAfPacket(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	plugin.afPacketConfigurator = &AFPacketConfigurator{
		Logger:           plugin.Log,
		SwIfIndexes:      plugin.swIfIndexes,
		Linux:            1, // Flag
		vppCh:            ctx.MockChannel,
		afPacketByHostIf: make(map[string]*AfPacketConfig),
		afPacketByName:   make(map[string]*AfPacketConfig),
		hostInterfaces:   make(map[string]struct{}),
	}
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&af_packet.AfPacketCreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var addresses []string
	data := getTestAfPacketData(ifNames[0], append(addresses, netAddresses[0]), afPacketHosts[0])
	// Register host
	plugin.afPacketConfigurator.hostInterfaces[afPacketHosts[0]] = struct{}{}
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_AF_PACKET_INTERFACE))
	Expect(meta.IpAddresses).To(HaveLen(1))
	Expect(meta.IpAddresses[0]).To(BeEquivalentTo(netAddresses[0]))
}

// Configure AfPacket interface
func TestInterfacesConfigureAfPacketPending(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	plugin.afPacketConfigurator = &AFPacketConfigurator{
		Logger:           plugin.Log,
		SwIfIndexes:      plugin.swIfIndexes,
		Linux:            1, // Flag
		vppCh:            ctx.MockChannel,
		afPacketByHostIf: make(map[string]*AfPacketConfig),
		afPacketByName:   make(map[string]*AfPacketConfig),
		hostInterfaces:   make(map[string]struct{}),
	}
	defer ifTestTeardown(ctx, plugin)
	// Data
	var addresses []string
	data := getTestAfPacketData(ifNames[0], append(addresses, netAddresses[0]), afPacketHosts[0])
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeFalse())
}

// Configure new interface and tests error propagation during configuration
func TestInterfacesConfigureInterfaceErrors(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.CreateLoopbackReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetRxModeReply{
		Retval: 1, // Simulate Rx mode error
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{
		Retval: 1, // Simulate MAC error
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{
		Retval: 1, // Interface VRF error
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{
		Retval: 1, // IP address error
	})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{
		Retval: 1, // Container IP error
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{
		Retval: 1, // MTU error
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_SOFTWARE_LOOPBACK, append(addresses, netAddresses[0]), false, macs[0], 1500)
	data.RxModeSettings = getTestRxModeSettings(if_api.RxModeType_POLLING)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).ToNot(BeNil())
	Expect(err.Error()).To(ContainSubstring("found 6 errors"))
	_, meta, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Name).To(BeEquivalentTo(ifNames[0]))
	Expect(meta.Type).To(BeEquivalentTo(if_api.InterfaceType_SOFTWARE_LOOPBACK))
	Expect(meta.IpAddresses).To(HaveLen(1))
	Expect(meta.IpAddresses[0]).To(BeEquivalentTo(netAddresses[0]))
	Expect(meta.Mtu).To(BeEquivalentTo(1500))
}

// Configure new interface and tests admin up error
func TestInterfacesConfigureInterfaceAdminUpError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.CreateLoopbackReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{
		Retval: 1,
	})
	// Data
	var addresses []string
	data := getTestInterface(ifNames[0], if_api.InterfaceType_SOFTWARE_LOOPBACK, append(addresses, netAddresses[0]), false, macs[0], 0)
	// Test configure TAP
	err = plugin.ConfigureVPPInterface(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.swIfIndexes.LookupIdx(data.Name)
	Expect(found).To(BeTrue())
}

// Modify TAPv1 interface
func TestInterfacesModifyTapV1WithoutTapData(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetRxModeReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var oldAddresses, newAddresses []string
	tapData := getTestTapInterface(1, ifNames[0])
	oldData := getTestInterface(ifNames[0], if_api.InterfaceType_TAP_INTERFACE, append(oldAddresses, netAddresses[0]), false, macs[0], 1500)
	oldData.Tap = tapData
	oldData.RxModeSettings = getTestRxModeSettings(if_api.RxModeType_DEFAULT)
	newData := getTestInterface(ifNames[0], if_api.InterfaceType_TAP_INTERFACE, append(newAddresses, netAddresses[1]), false, macs[1], 2000)
	newData.Tap = tapData
	newData.RxModeSettings = getTestRxModeSettings(if_api.RxModeType_INTERRUPT)
	// Register old config
	plugin.swIfIndexes.RegisterName(ifNames[0], 1, oldData)
	// Test configure TAP
	err = plugin.ModifyVPPInterface(newData, oldData)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(newData.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.IpAddresses).To(HaveLen(1))
	Expect(meta.IpAddresses[0]).To(BeEquivalentTo(netAddresses[1]))
	Expect(meta.PhysAddress).To(BeEquivalentTo(macs[1]))
	Expect(meta.Mtu).To(BeEquivalentTo(2000))
	Expect(meta.RxModeSettings).ToNot(BeNil())
	Expect(meta.RxModeSettings.RxMode).To(BeEquivalentTo(if_api.RxModeType_INTERRUPT))
}

// Modify TAPv1 interface including tap data
func TestInterfacesModifyTapV1TapData(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{}) // Delete
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&tap.TapDeleteReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&tap.TapConnectReply{ // Create
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetRxModeReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var oldAddresses, newAddresses []string
	oldData := getTestInterface(ifNames[0], if_api.InterfaceType_TAP_INTERFACE, append(oldAddresses, netAddresses[0]), false, macs[0], 1500)
	oldData.Tap = getTestTapInterface(1, ifNames[0])
	oldData.RxModeSettings = getTestRxModeSettings(if_api.RxModeType_DEFAULT)
	newData := getTestInterface(ifNames[0], if_api.InterfaceType_TAP_INTERFACE, append(newAddresses, netAddresses[0]), false, macs[0], 1500)
	newData.Tap = getTestTapInterface(1, ifNames[1])
	newData.RxModeSettings = getTestRxModeSettings(if_api.RxModeType_INTERRUPT)
	// Register old config
	plugin.swIfIndexes.RegisterName(ifNames[0], 1, oldData)
	// Test configure TAP
	err = plugin.ModifyVPPInterface(newData, oldData)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(newData.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.IpAddresses).To(HaveLen(1))
	Expect(meta.IpAddresses[0]).To(BeEquivalentTo(netAddresses[0]))
	Expect(meta.PhysAddress).To(BeEquivalentTo(macs[0]))
	Expect(meta.Mtu).To(BeEquivalentTo(1500))
	Expect(meta.RxModeSettings).ToNot(BeNil())
	Expect(meta.RxModeSettings.RxMode).To(BeEquivalentTo(if_api.RxModeType_INTERRUPT))
}

// Modify memif interface
func TestInterfacesModifyMemifWithoutMemifData(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&dhcp_api.DhcpClientConfigReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var oldAddresses, newAddresses []string
	memifData := getTestMemifInterface(true, 1)
	oldData := getTestInterface(ifNames[0], if_api.InterfaceType_MEMORY_INTERFACE, append(oldAddresses, netAddresses[0]), false, macs[0], 1500)
	oldData.Memif = memifData
	newData := getTestInterface(ifNames[0], if_api.InterfaceType_MEMORY_INTERFACE, newAddresses, true, macs[0], 1500)
	newData.Memif = memifData
	newData.Enabled = false
	newData.ContainerIpAddress = ipAddresses[3]
	// Register old config
	plugin.swIfIndexes.RegisterName(ifNames[0], 1, oldData)
	// Test configure TAP
	err = plugin.ModifyVPPInterface(newData, oldData)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(newData.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.SetDhcpClient).To(BeTrue())
	Expect(meta.Enabled).To(BeFalse())
}

// Modify memif interface including memif data
func TestInterfacesModifyMemifData(t *testing.T) {
	var err error
	// Setup
	ctx, plugin := ifTestSetup(t)
	defer ifTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{}) // Delete
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&memif.MemifDeleteReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&memif.MemifCreateReply{ // Create
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceTagAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMacAddressReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetTableReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceAddDelAddressReply{})
	ctx.MockVpp.MockReply(&ip.IPContainerProxyAddDelReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetMtuReply{})
	ctx.MockVpp.MockReply(&interfaces.SwInterfaceSetFlagsReply{})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{}) // Break status propagation
	// Data
	var oldAddresses, newAddresses []string
	oldData := getTestInterface(ifNames[0], if_api.InterfaceType_MEMORY_INTERFACE, append(oldAddresses, netAddresses[0]), false, macs[0], 1500)
	oldData.Memif = getTestMemifInterface(true, 1)
	newData := getTestInterface(ifNames[0], if_api.InterfaceType_MEMORY_INTERFACE, append(newAddresses, netAddresses[0]), false, macs[0], 1500)
	newData.Memif = getTestMemifInterface(false, 2)
	// Register old config and socket filename
	plugin.swIfIndexes.RegisterName(ifNames[0], 1, oldData)
	plugin.memifScCache["socket-filename"] = 0
	// Test configure TAP
	err = plugin.ModifyVPPInterface(newData, oldData)
	Expect(err).To(BeNil())
	_, meta, found := plugin.swIfIndexes.LookupIdx(newData.Name)
	Expect(found).To(BeTrue())
	Expect(meta).ToNot(BeNil())
	Expect(meta.Memif).ToNot(BeNil())
	Expect(meta.Memif.Master).To(BeFalse())
	Expect(meta.Memif.Id).To(BeEquivalentTo(2))
}

/* Interface Test Setup */

func ifTestSetup(t *testing.T) (*vppcallmock.TestCtx, *InterfaceConfigurator) {
	ctx := vppcallmock.SetupTestCtx(t)
	// Logger
	log := logrus.DefaultLogger()
	log.SetLevel(logging.DebugLevel)

	return ctx, &InterfaceConfigurator{
		Log:          log,
		swIfIndexes:  ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(log, "if-test", "if", nil)),
		dhcpIndices:  ifaceidx.NewDHCPIndex(nametoidx.NewNameToIdx(log, "dhcp-test", "dhcp", nil)),
		uIfaceCache:  make(map[string]string),
		memifScCache: make(map[string]uint32),
		vppCh:        ctx.MockChannel,
		notifChan:    make(chan govppapi.Message, 5),
	}
}

func ifTestTeardown(ctx *vppcallmock.TestCtx, plugin *InterfaceConfigurator) {
	ctx.TeardownTestCtx()
	err := plugin.Close()
	Expect(err).To(BeNil())
}

/* Interface Test Data */

func getSimpleTestInterface(name string, ip []string) *if_api.Interfaces_Interface {
	return &if_api.Interfaces_Interface{
		Name:        name,
		IpAddresses: ip,
	}
}

func getTestInterface(name string, ifType if_api.InterfaceType, ip []string, dhcp bool, mac string, mtu uint32) *if_api.Interfaces_Interface {
	return &if_api.Interfaces_Interface{
		Name:               name,
		Enabled:            true,
		Type:               ifType,
		IpAddresses:        ip,
		SetDhcpClient:      dhcp,
		PhysAddress:        mac,
		Mtu:                mtu,
		ContainerIpAddress: ipAddresses[4],
	}
}

func getTestMemifInterface(master bool, id uint32) *if_api.Interfaces_Interface_Memif {
	return &if_api.Interfaces_Interface_Memif{
		Master:         master,
		Id:             id,
		SocketFilename: "socket-filename",
	}
}

func getTestVxLanInterface(src, dst string, vni uint32) *if_api.Interfaces_Interface_Vxlan {
	return &if_api.Interfaces_Interface_Vxlan{
		SrcAddress: src,
		DstAddress: dst,
		Vni:        vni,
	}
}

func getTestTapInterface(ver uint32, host string) *if_api.Interfaces_Interface_Tap {
	return &if_api.Interfaces_Interface_Tap{
		Version:    ver,
		HostIfName: host,
	}
}

func getTestRxModeSettings(mode if_api.RxModeType) *if_api.Interfaces_Interface_RxModeSettings {
	return &if_api.Interfaces_Interface_RxModeSettings{
		RxMode: mode,
	}
}

func getTestUnnumberedSettings(ifNameWithIP string) *if_api.Interfaces_Interface_Unnumbered {
	return &if_api.Interfaces_Interface_Unnumbered{
		IsUnnumbered:    true,
		InterfaceWithIP: ifNameWithIP,
	}
}
