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

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	nat_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var label = "test-dnat"
var ifNames = []string{"if1", "if2", "if3"}
var ipAddresses = []string{"10.0.0.1", "10.0.0.2", "125.0.0.1", "172.125.0.1", "124.10.0.1"}
var invalidIP = "invalid-ip"
var ports = []uint32{8000, 8500, 8989, 9000}

/* Global NAT Test Cases */

// Enable NAT forwarding
func TestNatConfiguratorEnableForwarding(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	// Data
	data := getTestNatForwardingConfig(true)
	// Test
	err = plugin.SetNatGlobalConfig(data)
	Expect(err).To(BeNil())
}

// Disable NAT forwarding
func TestNatConfiguratorDisableForwarding(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	// Data
	data := getTestNatForwardingConfig(false)
	// Test
	err = plugin.SetNatGlobalConfig(data)
	Expect(err).To(BeNil())
}

// Modify NAT forwarding
func TestNatConfiguratorModifyForwarding(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // Create
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // Modify
	// Data
	oldData := getTestNatForwardingConfig(true)
	newData := getTestNatForwardingConfig(false)
	// Test create
	err = plugin.SetNatGlobalConfig(oldData)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.Forwarding).To(BeTrue())
	// Test modify
	err = plugin.ModifyNatGlobalConfig(oldData, newData)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.Forwarding).To(BeFalse())
}

// Enable two interfaces for NAT, then remove one
func TestNatConfiguratorEnableDisableInterfaces(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // First case
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // Second case
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	// Registration
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	// Data
	var ifs1, ifs2 []*nat.Nat44Global_NatInterfaces
	firstData := &nat.Nat44Global{NatInterfaces: append(ifs1,
		getTestNatInterfaceConfig(ifNames[0], true, false),
		getTestNatInterfaceConfig(ifNames[1], true, false))}
	secondData := &nat.Nat44Global{NatInterfaces: append(ifs2,
		getTestNatInterfaceConfig(ifNames[0], true, false))}
	// Test set interfaces
	err = plugin.SetNatGlobalConfig(firstData)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(2))
	// Test disable one interface
	err = plugin.SetNatGlobalConfig(secondData)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(1))
}

// Enable two output interfaces for NAT, then remove one
func TestNatConfiguratorEnableDisableOutputInterfaces(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // First case
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // Second case
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	// Registration
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	// Data
	var ifs1, ifs2 []*nat.Nat44Global_NatInterfaces
	firstData := &nat.Nat44Global{NatInterfaces: append(ifs1,
		getTestNatInterfaceConfig(ifNames[0], true, true),
		getTestNatInterfaceConfig(ifNames[1], true, true))}
	secondData := &nat.Nat44Global{NatInterfaces: append(ifs2,
		getTestNatInterfaceConfig(ifNames[0], true, true))}
	// Test set output interfaces
	err = plugin.SetNatGlobalConfig(firstData)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(2))
	// Test disable one output interface
	err = plugin.SetNatGlobalConfig(secondData)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(1))
}

// Create and modify NAT interfaces and output interfaces
func TestNatConfiguratorModifyInterfaces(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // Create
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{}) // Modify
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	// Registration
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	swIfIndices.RegisterName(ifNames[2], 3, nil)
	// Data
	var ifs1, ifs2 []*nat.Nat44Global_NatInterfaces
	oldData := &nat.Nat44Global{NatInterfaces: append(ifs1,
		getTestNatInterfaceConfig(ifNames[0], true, false),
		getTestNatInterfaceConfig(ifNames[1], false, true),
		getTestNatInterfaceConfig(ifNames[2], true, true))}
	newData := &nat.Nat44Global{NatInterfaces: append(ifs2,
		getTestNatInterfaceConfig(ifNames[0], false, true),
		getTestNatInterfaceConfig(ifNames[1], true, false),
		getTestNatInterfaceConfig(ifNames[2], true, true))}
	// Test create
	err = plugin.SetNatGlobalConfig(oldData)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(3))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(3))
	// Test modify
	err = plugin.ModifyNatGlobalConfig(oldData, newData)
	Expect(err).To(BeNil())
}

// Test interface cache registering and un-registering interfaces after configuration
func TestNatConfiguratorInterfaceCache(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	// Data
	var ifs []*nat.Nat44Global_NatInterfaces
	data := &nat.Nat44Global{NatInterfaces: append(ifs,
		getTestNatInterfaceConfig(ifNames[0], true, false),
		getTestNatInterfaceConfig(ifNames[1], true, true))}
	// Test create
	err = plugin.SetNatGlobalConfig(data)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(0))
	Expect(plugin.notEnabledIfs).To(HaveLen(2))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	// Test register first interface
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	err = plugin.ResolveCreatedInterface(ifNames[0], 1)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(1))
	Expect(plugin.notEnabledIfs).To(HaveLen(1))
	// Test register second interface
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	err = plugin.ResolveCreatedInterface(ifNames[1], 2)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	// Test un-register second interface
	_, _, found := swIfIndices.UnregisterName(ifNames[1])
	Expect(found).To(BeTrue())
	err = plugin.ResolveDeletedInterface(ifNames[1], 1)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(1))
	Expect(plugin.notEnabledIfs).To(HaveLen(1))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	// Test re-enable second interface
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	err = plugin.ResolveCreatedInterface(ifNames[1], 2)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
}

// Set NAT address pools
func TestNatConfiguratorCreateAddressPool(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	// Data
	var aps []*nat.Nat44Global_AddressPools
	data := &nat.Nat44Global{AddressPools: append(aps,
		getTestNatAddressPoolConfig(ipAddresses[0], ipAddresses[1], 0, true),
		getTestNatAddressPoolConfig(ipAddresses[2], "", 1, false),
		getTestNatAddressPoolConfig("", ipAddresses[3], 1, false))}
	// Test set address pool
	err = plugin.SetNatGlobalConfig(data)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.AddressPools).To(HaveLen(3))
}

// Set and modify NAT address pools
func TestNatConfiguratorModifyAddressPool(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // Configure
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{}) // Modify
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	// Data
	var aps1, aps2 []*nat.Nat44Global_AddressPools
	oldData := &nat.Nat44Global{AddressPools: append(aps1,
		getTestNatAddressPoolConfig(ipAddresses[0], "", 0, true),
		getTestNatAddressPoolConfig("", ipAddresses[3], 1, false))}
	newData := &nat.Nat44Global{AddressPools: append(aps2,
		getTestNatAddressPoolConfig(ipAddresses[0], "", 0, true),
		getTestNatAddressPoolConfig("", ipAddresses[2], 1, false))}
	// Test set address pool
	err = plugin.SetNatGlobalConfig(oldData)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.AddressPools).To(HaveLen(2))
	// Test modify address pool
	err = plugin.ModifyNatGlobalConfig(oldData, newData)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.AddressPools).To(HaveLen(2))
}

// Test various errors which may occur during address pool configuration
func TestNatConfiguratorAddressPoolErrors(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	// Data
	var aps1, aps2, aps3 []*nat.Nat44Global_AddressPools
	data1 := &nat.Nat44Global{AddressPools: append(aps1, getTestNatAddressPoolConfig("", "", 0, true))}
	data2 := &nat.Nat44Global{AddressPools: append(aps2, getTestNatAddressPoolConfig(invalidIP, "", 0, true))}
	data3 := &nat.Nat44Global{AddressPools: append(aps3, getTestNatAddressPoolConfig("", invalidIP, 0, true))}
	// Test no IP address provided
	err = plugin.SetNatGlobalConfig(data1)
	Expect(err).ToNot(BeNil())
	// Test invalid first IP
	err = plugin.SetNatGlobalConfig(data2)
	Expect(err).ToNot(BeNil())
	// Test invalid last IP
	err = plugin.SetNatGlobalConfig(data3)
	Expect(err).ToNot(BeNil())
}

// Remove global NAT configuration
func TestNatConfiguratorDeleteGlobalConfig(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{}) // Configure
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{}) // Delete
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{}) // Re-register
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	// Registration
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	// Data
	var ifs []*nat.Nat44Global_NatInterfaces
	var aps []*nat.Nat44Global_AddressPools
	data := &nat.Nat44Global{NatInterfaces: append(ifs,
		getTestNatInterfaceConfig(ifNames[0], true, false),
		getTestNatInterfaceConfig(ifNames[1], false, true),
		getTestNatInterfaceConfig(ifNames[2], false, false)),
		AddressPools: append(aps, getTestNatAddressPoolConfig(ipAddresses[0], ipAddresses[1], 0, true))}
	// Test set config
	err = plugin.SetNatGlobalConfig(data)
	Expect(err).To(BeNil())
	Expect(plugin.notEnabledIfs).To(HaveLen(1))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	// Test un-register interface
	_, _, found := swIfIndices.UnregisterName(ifNames[1])
	Expect(found).To(BeTrue())
	err = plugin.ResolveDeletedInterface(ifNames[1], 1)
	Expect(err).To(BeNil())
	Expect(plugin.notEnabledIfs).To(HaveLen(2))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	// Test delete config
	err = plugin.DeleteNatGlobalConfig(data)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(1))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(2))
	Expect(plugin.globalNAT).To(BeNil())
	// Test re-create interfaces
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	err = plugin.ResolveCreatedInterface(ifNames[1], 2)
	Expect(err).To(BeNil())
	swIfIndices.RegisterName(ifNames[2], 3, nil)
	err = plugin.ResolveCreatedInterface(ifNames[2], 3)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(3))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
}

// Remove empty global NAT configuration
func TestNatConfiguratorDeleteGlobalConfigEmpty(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	data := &nat.Nat44Global{}
	// Test delete empty config
	err = plugin.DeleteNatGlobalConfig(data)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT).To(BeNil())
}

/* DNAT test cases */

// Configure DNAT static mapping without local IP
func TestNatConfiguratorDNatStaticMappingNoLocalIPError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], ports[0], nat.Protocol_TCP))}
	// Test configure DNAT without local IPs
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
}

// Configure DNAT static mapping using external IP
func TestNatConfiguratorDNatStaticMapping(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[1], 32)))}
	// Test configure DNAT
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.VrfID).To(BeEquivalentTo(0))
	Expect(msg.LocalPort).To(BeEquivalentTo(ports[1]))
	Expect(msg.ExternalPort).To(BeEquivalentTo(ports[0]))
	Expect(msg.TwiceNat).To(BeEquivalentTo(0))
	Expect(msg.AddrOnly).To(BeEquivalentTo(0))
	Expect(msg.LocalIPAddress).To(BeEquivalentTo(net.ParseIP(ipAddresses[1]).To4()))
	Expect(msg.ExternalIPAddress).To(BeEquivalentTo(net.ParseIP(ipAddresses[0]).To4()))
	Expect(msg.ExternalSwIfIndex).To(BeEquivalentTo(vppcalls.NoInterface))
	Expect(msg.Protocol).To(BeEquivalentTo(6))
}

// Configure DNAT static mapping using external interface
func TestNatConfiguratorDNatStaticMappingExternalInterface(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	// Registrations
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, ifNames[0], ipAddresses[0], ports[0], nat.Protocol_UDP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[1], 32)))}
	// Configure DNAT with external interface
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.ExternalSwIfIndex).To(BeEquivalentTo(1))
	Expect(msg.Protocol).To(BeEquivalentTo(17))
}

// Configure DNAT static mapping as address-only
func TestNatConfiguratorDNatStaticMappingAddressOnly(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], 0, nat.Protocol_ICMP,
			getTestNatStaticLocalIP(ipAddresses[1], 0, 32)))}
	// Test configure DNAT address only
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.AddrOnly).To(BeEquivalentTo(1))
	Expect(msg.Protocol).To(BeEquivalentTo(1))
}

// Configure DNAT with invalid local IP
func TestNatConfiguratorDNatStaticMappingInvalidLocalAddressError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], 0, 0,
			getTestNatStaticLocalIP(invalidIP, 0, 32)))}
	// Test configure DNAT with invalid local IP
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Configure DNAT with invalid external IP
func TestNatConfiguratorDNatStaticMappingInvalidExternalAddressError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", invalidIP, 0, 0,
			getTestNatStaticLocalIP(ipAddresses[1], 0, 32)))}
	// Test configure DNAT with invalid external IP
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Configure DNAT with non-existing external interface
func TestNatConfiguratorDNatStaticMappingMissingInterfaceError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, ifNames[0], ipAddresses[0], 0, 0,
			getTestNatStaticLocalIP(ipAddresses[1], 0, 32)))}
	// Test configure DNAT with missing external interface
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Configure DNAT with unknown protocol and check whether it will be set to default
func TestNatConfiguratorDNatStaticMappingUnknownProtocol(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], 0, 10,
			getTestNatStaticLocalIP(ipAddresses[1], 0, 32)))}
	// Test configure DNAT with unnown protocol
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.Protocol).To(BeEquivalentTo(6)) // Expected protocol is TCP
	err = plugin.Close()
	Expect(err).To(BeNil())
}

// Configure DNAT static mapping with load balancer
func TestNatConfiguratorDNatStaticMappingLb(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{})
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[1], 35),
			getTestNatStaticLocalIP(ipAddresses[2], ports[2], 65)))}
	// Test configure DNAT static mapping with load balancer
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelLbStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.VrfID).To(BeEquivalentTo(0))
	Expect(msg.ExternalPort).To(BeEquivalentTo(ports[0]))
	Expect(msg.TwiceNat).To(BeEquivalentTo(0))
	Expect(msg.Protocol).To(BeEquivalentTo(6))
	Expect(msg.LocalNum).To(BeEquivalentTo(2))
}

// Configure DNAT static mapping with load balancer with invalid local IP
func TestNatConfiguratorDNatStaticMappingLbInvalidLocalError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], 0, nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], 0, 35),
			getTestNatStaticLocalIP(invalidIP, ports[1], 65)))}
	// Test configure DNAT static mapping with load balancer with invalid local IP
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Configure DNAT static mapping with load balancer with missing external port
func TestNatConfiguratorDNatStaticMappingLbMissingExternalPortError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], 0, nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[0], 35),
			getTestNatStaticLocalIP(ipAddresses[2], ports[1], 65)))}
	// Test configure static mapping with load balancer with missing external port
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Configure DNAT static mapping with load balancer with invalid external IP
func TestNatConfiguratorDNatStaticMappingLbInvalidExternalIPError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", invalidIP, ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[1], 35),
			getTestNatStaticLocalIP(ipAddresses[2], ports[2], 65)))}
	// Test DNAT static mapping invalid external IP
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Configure NAT identity mapping
func TestNatConfiguratorDNatIdentityMapping(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})
	// Data
	var idMaps []*nat.Nat44DNat_DNatConfig_IdentityMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, IdMappings: append(idMaps,
		getTestNatIdentityMappingConfig(0, "", ipAddresses[0], ports[0], nat.Protocol_TCP))}
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getIdMappingIdentifier(data.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.VrfID).To(BeEquivalentTo(0))
	Expect(msg.IPAddress).To(BeEquivalentTo(net.ParseIP(ipAddresses[0]).To4()))
	Expect(msg.Port).To(BeEquivalentTo(ports[0]))
	Expect(msg.Protocol).To(BeEquivalentTo(6))
}

// Configure NAT identity mapping with address interface
func TestNatConfiguratorDNatIdentityMappingInterface(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})
	// Register
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	// Data
	var idMaps []*nat.Nat44DNat_DNatConfig_IdentityMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, IdMappings: append(idMaps,
		getTestNatIdentityMappingConfig(0, ifNames[0], "", 0, nat.Protocol_TCP))}
	// Test identity mapping with address interface
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getIdMappingIdentifier(data.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))
}

// Configure NAT identity mapping with address interface while interface is not registered
func TestNatConfiguratorDNatIdentityMappingMissingInterfaceError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var idMaps []*nat.Nat44DNat_DNatConfig_IdentityMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, IdMappings: append(idMaps,
		getTestNatIdentityMappingConfig(0, ifNames[0], "", 0, nat.Protocol_TCP))}
	// Test identity mapping with address interface	while interface is not registered
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getIdMappingIdentifier(data.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Create NAT identity mapping with invalid IP address
func TestNatConfiguratorDNatIdentityMappingInvalidIPError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var idMaps []*nat.Nat44DNat_DNatConfig_IdentityMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, IdMappings: append(idMaps,
		getTestNatIdentityMappingConfig(0, "", invalidIP, ports[0], nat.Protocol_TCP))}
	// Test identity mapping with invalid IP
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getIdMappingIdentifier(data.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Create identity mapping without IP address and interface set
func TestNatConfiguratorDNatIdentityMappingNoInterfaceAndIPError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var idMaps []*nat.Nat44DNat_DNatConfig_IdentityMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, IdMappings: append(idMaps,
		getTestNatIdentityMappingConfig(0, "", "", ports[0], nat.Protocol_TCP))}
	// Test identity mapping without interface and IP
	err = plugin.ConfigureDNat(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getIdMappingIdentifier(data.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Configure and modify static and identity mappings
func TestNatConfiguratorDNatModify(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{}) // Configure
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{}) // Modify
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	var idMaps []*nat.Nat44DNat_DNatConfig_IdentityMappings
	oldData := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], 0, 0,
			getTestNatStaticLocalIP(ipAddresses[1], 0, 35)),
		getTestNatStaticMappingConfig(0, "", ipAddresses[1], ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[2], ports[1], 35),
			getTestNatStaticLocalIP(ipAddresses[3], ports[2], 65))),
		IdMappings: append(idMaps,
			getTestNatIdentityMappingConfig(0, "", ipAddresses[4], ports[3], nat.Protocol_TCP))}
	newData := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], 0, 0,
			getTestNatStaticLocalIP(ipAddresses[1], 0, 35)),
		getTestNatStaticMappingConfig(0, "", ipAddresses[4], ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[3], ports[1], 0))),
		IdMappings: append(idMaps,
			getTestNatIdentityMappingConfig(0, "", ipAddresses[2], ports[1], 0))}
	// Test configure static and identity mappings
	err = plugin.ConfigureDNat(oldData)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(oldData.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(oldData.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	id = getStMappingIdentifier(oldData.StMappings[1])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	id = getIdMappingIdentifier(oldData.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	// Test modify static and identity mapping
	err = plugin.ModifyDNat(oldData, newData)
	Expect(err).To(BeNil())
	id = getStMappingIdentifier(newData.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	id = getStMappingIdentifier(newData.StMappings[1])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	id = getIdMappingIdentifier(newData.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
}

//  Configure and delete static mapping
func TestNatConfiguratorDeleteDNatStaticMapping(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{}) // Configure
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{}) // Delete
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[1], 35)))}
	// Test configure DNAT static mapping
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	// Test delete static mapping
	err = plugin.DeleteDNat(data)
	Expect(err).To(BeNil())
	_, _, found = plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeFalse())
	id = getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Delete static mapping with invalid IP
func TestNatConfiguratorDeleteDNatStaticMappingInvalidIPError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", invalidIP, ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[1], 35)))}
	// Test delete static mapping with invalid interface
	err = plugin.DeleteDNat(data)
	Expect(err).ToNot(BeNil())
}

// Configure and delete static mapping with load balancer
func TestNatConfiguratorDeleteDNatStaticMappingLb(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{}) // Configure
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{}) // Delete
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", ipAddresses[0], ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[1], 35),
			getTestNatStaticLocalIP(ipAddresses[2], ports[2], 65)))}
	// Test configure static mapping with load balancer
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	// Test delete static mapping with load balancer
	err = plugin.DeleteDNat(data)
	Expect(err).To(BeNil())
	_, _, found = plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeFalse())
	id = getStMappingIdentifier(data.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Delete static mapping with load ballancer with invalid IP address
func TestNatConfiguratorDeleteDNatStaticMappingLbInvalidIPError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var stMaps []*nat.Nat44DNat_DNatConfig_StaticMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, StMappings: append(stMaps,
		getTestNatStaticMappingConfig(0, "", invalidIP, ports[0], nat.Protocol_TCP,
			getTestNatStaticLocalIP(ipAddresses[1], ports[1], 35),
			getTestNatStaticLocalIP(ipAddresses[2], ports[2], 65)))}
	// Test delete static mapping with load balancer with invalid IP
	err = plugin.DeleteDNat(data)
	Expect(err).ToNot(BeNil())
}

// Configure and delete NAT identity mapping
func TestNatConfiguratorDNatDeleteIdentityMapping(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{}) // Configure
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{}) // Delete
	// Data
	var idMaps []*nat.Nat44DNat_DNatConfig_IdentityMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, IdMappings: append(idMaps,
		getTestNatIdentityMappingConfig(0, "", ipAddresses[0], ports[0], nat.Protocol_TCP))}
	// Test configure identity mapping
	err = plugin.ConfigureDNat(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeTrue())
	id := getIdMappingIdentifier(data.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	// Test delete identity mapping
	err = plugin.DeleteDNat(data)
	Expect(err).To(BeNil())
	_, _, found = plugin.DNatIndices.LookupIdx(data.Label)
	Expect(found).To(BeFalse())
	id = getIdMappingIdentifier(data.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())
}

// Delete NAT identity mapping without ip address set
func TestNatConfiguratorDNatDeleteIdentityMappingNoInterfaceAndIP(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := natTestSetup(t)
	defer natTestTeardown(ctx, plugin)
	// Data
	var idMaps []*nat.Nat44DNat_DNatConfig_IdentityMappings
	data := &nat.Nat44DNat_DNatConfig{Label: label, IdMappings: append(idMaps,
		getTestNatIdentityMappingConfig(0, "", "", ports[0], nat.Protocol_TCP))}
	// Test delete identity mapping without interface IP
	err = plugin.DeleteDNat(data)
	Expect(err).ToNot(BeNil())
}

/* NAT Test Setup */

func natTestSetup(t *testing.T) (*vppcallmock.TestCtx, *NatConfigurator, ifaceidx.SwIfIndexRW) {
	ctx := vppcallmock.SetupTestCtx(t)
	// Logger
	log := logrus.DefaultLogger()
	log.SetLevel(logging.DebugLevel)

	// Interface indices
	swIfIndices := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(log, "nat-configurator-test", "nat", nil))

	return ctx, &NatConfigurator{
		Log:                  log,
		SwIfIndexes:          swIfIndices,
		DNatIndices:          nametoidx.NewNameToIdx(log, "dnat-test", "dnat", nil),
		DNatStMappingIndices: nametoidx.NewNameToIdx(log, "dnat-test", "dnat", nil),
		DNatIdMappingIndices: nametoidx.NewNameToIdx(log, "dnat-test", "dnat", nil),
		vppChan:              ctx.MockChannel,
		notEnabledIfs:        make(map[string]*nat.Nat44Global_NatInterfaces),
		notDisabledIfs:       make(map[string]*nat.Nat44Global_NatInterfaces),
	}, swIfIndices
}

func natTestTeardown(ctx *vppcallmock.TestCtx, plugin *NatConfigurator) {
	ctx.TeardownTestCtx()
	err := plugin.Close()
	Expect(err).To(BeNil())
}

/* NAT Test Data */

func getTestNatForwardingConfig(fwd bool) *nat.Nat44Global {
	return &nat.Nat44Global{Forwarding: fwd}
}

func getTestNatInterfaceConfig(name string, inside, output bool) *nat.Nat44Global_NatInterfaces {
	return &nat.Nat44Global_NatInterfaces{
		Name:          name,
		IsInside:      inside,
		OutputFeature: output,
	}
}

func getTestNatAddressPoolConfig(first, last string, vrf uint32, twn bool) *nat.Nat44Global_AddressPools {
	return &nat.Nat44Global_AddressPools{
		FirstSrcAddress: first,
		LastSrcAddress:  last,
		VrfId:           vrf,
		TwiceNat:        twn,
	}
}

func getTestNatStaticMappingConfig(vrf uint32, ifName, externalIP string, externalPort uint32, proto nat.Protocol, locals ...*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs) *nat.Nat44DNat_DNatConfig_StaticMappings {
	return &nat.Nat44DNat_DNatConfig_StaticMappings{
		VrfId:             vrf,
		ExternalInterface: ifName,
		ExternalIP:        externalIP,
		ExternalPort:      externalPort,
		LocalIps:          locals,
		Protocol:          proto,
	}
}

func getTestNatIdentityMappingConfig(vrf uint32, ifName, ip string, port uint32, proto nat.Protocol) *nat.Nat44DNat_DNatConfig_IdentityMappings {
	return &nat.Nat44DNat_DNatConfig_IdentityMappings{
		VrfId:              vrf,
		AddressedInterface: ifName,
		IpAddress:          ip,
		Port:               port,
		Protocol:           proto,
	}
}

func getTestNatStaticLocalIP(ip string, port, probability uint32) *nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs {
	return &nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
		LocalIP:     ip,
		LocalPort:   port,
		Probability: probability,
	}
}
