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

// NAT Global test cases

func TestNatConfiguratorEnableForwarding(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	// Prepare replies
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})

	// Start
	err := plugin.SetNatGlobalConfig(getForwardingConfig(true))
	Expect(err).To(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDisableForwarding(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	// Prepare replies
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})

	// Start
	err := plugin.SetNatGlobalConfig(getForwardingConfig(false))
	Expect(err).To(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorModifyForwarding(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	// Prepare replies
	// Create
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	// Modify
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})

	// Start
	oldCfg := getForwardingConfig(true)
	newCfg := getForwardingConfig(false)

	// Set config
	err := plugin.SetNatGlobalConfig(oldCfg)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.Forwarding).To(BeTrue())

	// Modify config
	err = plugin.ModifyNatGlobalConfig(oldCfg, newCfg)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.Forwarding).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorEnableDisableInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Prepare replies
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})

	// Configurator
	plugin, swIfIndices := getNatConfigurator(ctx)

	// Prepare interface indices
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)

	// Enable two interfaces
	err := plugin.SetNatGlobalConfig(&nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
			{
				Name:          ifNames[0],
				IsInside:      true,
				OutputFeature: false,
			},
			{
				Name:          ifNames[1],
				IsInside:      true,
				OutputFeature: false,
			},
		},
	})
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(2))

	// Disable one interface
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
			{
				Name:          ifNames[0],
				IsInside:      true,
				OutputFeature: false,
			},
		},
	})
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(1))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorEnableDisableOutputInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Prepare replies
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})

	// Configurator
	plugin, swIfIndices := getNatConfigurator(ctx)

	// Prepare interface indices
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)

	// Enable two output feature interfaces
	err := plugin.SetNatGlobalConfig(&nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
			{
				Name:          ifNames[0],
				IsInside:      true,
				OutputFeature: true,
			},
			{
				Name:          ifNames[1],
				IsInside:      true,
				OutputFeature: true,
			},
		},
	})
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(2))

	// Disable one output feature interface
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
			{
				Name:          ifNames[0],
				IsInside:      true,
				OutputFeature: true,
			},
		},
	})
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(1))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorModifyInterfaces(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Prepare replies
	// Config
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})      // Called always during config
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})       // if1 configured
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{}) // if2 configured
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{}) // if3 configured
	// Modify (if3 was not changed)
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})       // if1 removed
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{}) // if2 removed
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})       // if1 configured
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{}) // if2 configured

	// Configurator
	plugin, swIfIndices := getNatConfigurator(ctx)

	// Prepare interface indices
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	swIfIndices.RegisterName(ifNames[2], 3, nil)

	oldCfg := &nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
			{
				Name:          ifNames[0],
				IsInside:      true,
				OutputFeature: false,
			},
			{
				Name:          ifNames[1],
				IsInside:      false,
				OutputFeature: true,
			},
			{
				Name:          ifNames[2],
				IsInside:      true,
				OutputFeature: true,
			},
		},
	}

	newCfg := &nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
			{
				Name:          ifNames[0],
				IsInside:      false,
				OutputFeature: true,
			},
			{
				Name:          ifNames[1],
				IsInside:      true,
				OutputFeature: false,
			},
			{
				Name:          ifNames[2],
				IsInside:      true,
				OutputFeature: true,
			},
		},
	}

	// Put old config
	err := plugin.SetNatGlobalConfig(oldCfg)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(3))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))
	Expect(plugin.globalNAT.NatInterfaces).To(HaveLen(3))

	// Modify it
	err = plugin.ModifyNatGlobalConfig(oldCfg, newCfg)
	Expect(err).To(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorInterfaceCache(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Prepare replies (one nat interface and one with output feature)
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})

	// Configurator (do not register indices)
	plugin, swIfIndices := getNatConfigurator(ctx)

	// Start
	err := plugin.SetNatGlobalConfig(&nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
			{
				Name:          ifNames[0],
				IsInside:      true,
				OutputFeature: false,
			},
			{
				Name:          ifNames[1],
				IsInside:      true,
				OutputFeature: true,
			},
		},
	})
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(0))
	// Both of them should be in cache for not enabled interfaces
	Expect(plugin.notEnabledIfs).To(HaveLen(2))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))

	// Register first interface
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	err = plugin.ResolveCreatedInterface(ifNames[0], 1)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(1))
	Expect(plugin.notEnabledIfs).To(HaveLen(1))

	// Register second interface
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	err = plugin.ResolveCreatedInterface(ifNames[1], 2)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))

	// Unregister second interface
	_, _, found := swIfIndices.UnregisterName(ifNames[1])
	Expect(found).To(BeTrue())
	err = plugin.ResolveDeletedInterface(ifNames[1], 1)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(1))
	Expect(plugin.notEnabledIfs).To(HaveLen(1))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))

	// Re-enable second interface
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	err = plugin.ResolveCreatedInterface(ifNames[1], 2)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(2))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorAddressPool(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Prepare replies (one nat interface and one with output feature)
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)

	// Start
	err := plugin.SetNatGlobalConfig(&nat.Nat44Global{
		AddressPools: []*nat.Nat44Global_AddressPools{
			{
				FirstSrcAddress: ipAddresses[0],
				LastSrcAddress:  ipAddresses[1],
				VrfId:           0,
				TwiceNat:        true,
			},
			{
				// First IP only
				FirstSrcAddress: ipAddresses[2],
				VrfId:           1,
				TwiceNat:        false,
			},
			{
				// Last IP only
				LastSrcAddress: ipAddresses[3],
				VrfId:          1,
				TwiceNat:       false,
			},
		},
	})
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.AddressPools).To(HaveLen(3))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorModifyAddressPool(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configure
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	// Modify
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)

	// Start
	oldCfg := &nat.Nat44Global{
		AddressPools: []*nat.Nat44Global_AddressPools{
			{
				FirstSrcAddress: ipAddresses[0],
				VrfId:           0,
				TwiceNat:        true,
			},
			{
				LastSrcAddress: ipAddresses[3],
				VrfId:          1,
				TwiceNat:       false,
			},
		},
	}

	newCfg := &nat.Nat44Global{
		AddressPools: []*nat.Nat44Global_AddressPools{
			{
				FirstSrcAddress: ipAddresses[0],
				VrfId:           0,
				TwiceNat:        true,
			},
			{
				FirstSrcAddress: ipAddresses[2],
				VrfId:           1,
				TwiceNat:        false,
			},
		},
	}

	err := plugin.SetNatGlobalConfig(oldCfg)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.AddressPools).To(HaveLen(2))

	err = plugin.ModifyNatGlobalConfig(oldCfg, newCfg)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT.AddressPools).To(HaveLen(2))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorAddressPoolErrors(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Prepare replies (one nat interface and one with output feature)
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)

	// No IP address provided
	err := plugin.SetNatGlobalConfig(&nat.Nat44Global{
		AddressPools: []*nat.Nat44Global_AddressPools{
			{
				VrfId:    0,
				TwiceNat: true,
			},
		},
	})
	Expect(err).ToNot(BeNil())

	// Invalid first IP
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
		AddressPools: []*nat.Nat44Global_AddressPools{
			{
				FirstSrcAddress: "not-an-ip",
				VrfId:           0,
				TwiceNat:        true,
			},
		},
	})
	Expect(err).ToNot(BeNil())

	// Invalid last IP
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
		AddressPools: []*nat.Nat44Global_AddressPools{
			{
				LastSrcAddress: "not-an-ip",
				VrfId:          0,
				TwiceNat:       true,
			},
		},
	})
	Expect(err).ToNot(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDeleteGlobalConfig(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Prepare replies (one nat interface and one with output feature)
	// Configure Global NAT
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	// Delete Global NAT
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelAddressRangeReply{})
	// Re-register interfaces
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelOutputFeatureReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44InterfaceAddDelFeatureReply{})

	// Configurator
	plugin, swIfIndices := getNatConfigurator(ctx)

	// Register required interfaces
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)

	// Config to remove
	natGlobalCfg := &nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{
			{
				// NAT interface
				Name:          ifNames[0],
				IsInside:      true,
				OutputFeature: false,
			},
			{
				// NAT output feature interface
				Name:          ifNames[1],
				IsInside:      false,
				OutputFeature: true,
			},
			{
				// NAT interface which is not registered yet
				Name:          ifNames[2],
				IsInside:      false,
				OutputFeature: false,
			},
		},
		AddressPools: []*nat.Nat44Global_AddressPools{
			{
				// Single address pool
				FirstSrcAddress: ipAddresses[0],
				LastSrcAddress:  ipAddresses[1],
				VrfId:           0,
				TwiceNat:        true,
			},
		},
	}
	err := plugin.SetNatGlobalConfig(natGlobalCfg)
	Expect(err).To(BeNil())
	Expect(plugin.notEnabledIfs).To(HaveLen(1))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))

	// Now un-register another interface
	_, _, found := swIfIndices.UnregisterName(ifNames[1])
	Expect(found).To(BeTrue())
	err = plugin.ResolveDeletedInterface(ifNames[1], 1)
	Expect(err).To(BeNil())
	Expect(plugin.notEnabledIfs).To(HaveLen(2))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))

	// Delete global config
	err = plugin.DeleteNatGlobalConfig(natGlobalCfg)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(1))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(2))
	Expect(plugin.globalNAT).To(BeNil())

	// Re-create missing interfaces
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	err = plugin.ResolveCreatedInterface(ifNames[1], 2)
	Expect(err).To(BeNil())
	swIfIndices.RegisterName(ifNames[2], 3, nil)
	err = plugin.ResolveCreatedInterface(ifNames[2], 3)
	Expect(err).To(BeNil())
	Expect(plugin.SwIfIndexes.GetMapping().ListNames()).To(HaveLen(3))
	Expect(plugin.notEnabledIfs).To(HaveLen(0))
	Expect(plugin.notDisabledIfs).To(HaveLen(0))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDeleteGlobalConfigEmpty(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)

	// Config to remove
	natGlobalCfg := &nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{},
		AddressPools:  []*nat.Nat44Global_AddressPools{},
	}

	// Delete empty global config
	err := plugin.DeleteNatGlobalConfig(natGlobalCfg)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT).To(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

// DNAT test cases

func TestNatConfiguratorDNatStaticMappingEmptyLocal(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   ipAddresses[0],
				ExternalPort: 8989,
				LocalIps:     []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMapping(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   ipAddresses[0],
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[1],
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
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

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingExternalInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})

	// Configurator
	plugin, swIfIndices := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	// Register interface
	swIfIndices.RegisterName(ifNames[0], 1, nil)

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:             0,
				ExternalInterface: ifNames[0],
				ExternalIP:        ipAddresses[0],
				ExternalPort:      ports[0],
				Protocol:          nat.Protocol_UDP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[1],
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.ExternalSwIfIndex).To(BeEquivalentTo(1))
	Expect(msg.Protocol).To(BeEquivalentTo(17))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingAddressOnly(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:      0,
				ExternalIP: ipAddresses[0],
				Protocol:   nat.Protocol_ICMP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.AddrOnly).To(BeEquivalentTo(1))
	Expect(msg.Protocol).To(BeEquivalentTo(1))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingInvalidLocalAddress(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:      0,
				ExternalIP: ipAddresses[0],
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     invalidIP,
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	// DNAT is registered
	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	// Mapping is not
	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingInvalidExternalAddress(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:      0,
				ExternalIP: invalidIP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	// DNAT is registered
	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	// Mapping is not
	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingMissingInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:             0,
				ExternalInterface: ifNames[0],
				ExternalIP:        ipAddresses[0],
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	// DNAT is registered
	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	// Mapping is not
	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingUnknownProtocol(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:      0,
				ExternalIP: ipAddresses[0],
				Protocol:   10,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	// DNAT is registered
	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	// Mapping is not
	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelStaticMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	// Expected protocol is TCP
	Expect(msg.Protocol).To(BeEquivalentTo(6))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingLb(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   ipAddresses[0],
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[1],
						Probability: 35,
					},
					{
						LocalIP:     ipAddresses[2],
						LocalPort:   ports[2],
						Probability: 65,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
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

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingLbInvalidLocals(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   ipAddresses[0],
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						// Missing port
						LocalIP:     ipAddresses[1],
						Probability: 35,
					},
					{
						// Invalid IP address
						LocalIP:     invalidIP,
						LocalPort:   ports[1],
						Probability: 65,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingLbMissingExternalPort(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:      0,
				ExternalIP: ipAddresses[0],
				Protocol:   nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[0],
						Probability: 35,
					},
					{
						LocalIP:     ipAddresses[2],
						LocalPort:   ports[1],
						Probability: 65,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatStaticMappingLbInvalidExternalIP(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   invalidIP,
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[1],
						Probability: 35,
					},
					{
						LocalIP:     ipAddresses[2],
						LocalPort:   ports[2],
						Probability: 65,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatIdentityMapping(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:     0,
				IpAddress: ipAddresses[0],
				Port:      ports[0],
				Protocol:  nat.Protocol_TCP,
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getIdMappingIdentifier(dNat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.VrfID).To(BeEquivalentTo(0))
	Expect(msg.IPAddress).To(BeEquivalentTo(net.ParseIP(ipAddresses[0]).To4()))
	Expect(msg.Port).To(BeEquivalentTo(ports[0]))
	Expect(msg.Protocol).To(BeEquivalentTo(6))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatIdentityMappingInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})

	// Configurator
	plugin, swIfIndices := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	// Register interface
	swIfIndices.RegisterName(ifNames[0], 1, nil)

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:              0,
				AddressedInterface: ifNames[0],
				Protocol:           nat.Protocol_TCP,
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getIdMappingIdentifier(dNat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	msg, ok := ctx.MockChannel.Msg.(*nat_api.Nat44AddDelIdentityMapping)
	Expect(ok).To(BeTrue())
	Expect(msg).ToNot(BeNil())
	Expect(msg.SwIfIndex).To(BeEquivalentTo(1))

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatIdentityMappingMissingInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:              0,
				AddressedInterface: ifNames[0],
				Protocol:           nat.Protocol_TCP,
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getIdMappingIdentifier(dNat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatIdentityMappingInvalidIP(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:     0,
				IpAddress: invalidIP,
				Port:      ports[0],
				Protocol:  nat.Protocol_TCP,
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getIdMappingIdentifier(dNat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatIdentityMappingNoInterfaceAndIP(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:    0,
				Port:     1000,
				Protocol: nat.Protocol_TCP,
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getIdMappingIdentifier(dNat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatModify(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT configure
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})
	// DNAT modify
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	oldDNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:      0,
				ExternalIP: ipAddresses[0],
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						Probability: 35,
					},
				},
			},
			{
				VrfId:        0,
				ExternalIP:   ipAddresses[1],
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[2],
						LocalPort:   ports[1],
						Probability: 35,
					},
					{
						LocalIP:     ipAddresses[3],
						LocalPort:   ports[2],
						Probability: 65,
					},
				},
			},
		},
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:     0,
				IpAddress: ipAddresses[4],
				Port:      ports[3],
				Protocol:  nat.Protocol_TCP,
			},
		},
	}

	err := plugin.ConfigureDNat(oldDNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(oldDNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(oldDNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	id = getStMappingIdentifier(oldDNat.StMappings[1])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	id = getIdMappingIdentifier(oldDNat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	newDnat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:      0,
				ExternalIP: ipAddresses[0],
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						Probability: 35,
					},
				},
			},
			{
				VrfId:        0,
				ExternalIP:   ipAddresses[4],
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:   ipAddresses[3],
						LocalPort: ports[1],
					},
				},
			},
		},
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:     0,
				IpAddress: ipAddresses[2],
				Port:      900,
			},
		},
	}

	err = plugin.ModifyDNat(oldDNat, newDnat)
	Expect(err).To(BeNil())

	id = getStMappingIdentifier(newDnat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
	id = getStMappingIdentifier(newDnat.StMappings[1])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	id = getIdMappingIdentifier(newDnat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())
}

func TestNatConfiguratorDeleteDNatStaticMapping(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelStaticMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   ipAddresses[0],
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[1],
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	err = plugin.DeleteDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found = plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeFalse())

	id = getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDeleteDNatStaticMappingInvalidInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   invalidIP,
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[1],
						Probability: 35,
					},
				},
			},
		},
	}

	err := plugin.DeleteDNat(dNat)
	Expect(err).ToNot(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDeleteDNatStaticMappingLb(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// DNAT replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelLbStaticMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   ipAddresses[0],
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[1],
						Probability: 35,
					},
					{
						LocalIP:     ipAddresses[2],
						LocalPort:   ports[2],
						Probability: 65,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	err = plugin.DeleteDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found = plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeFalse())

	id = getStMappingIdentifier(dNat.StMappings[0])
	_, _, found = plugin.DNatStMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDeleteDNatStaticMappingLbInvalidInterface(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		StMappings: []*nat.Nat44DNat_DNatConfig_StaticMappings{
			{
				VrfId:        0,
				ExternalIP:   invalidIP,
				ExternalPort: ports[0],
				Protocol:     nat.Protocol_TCP,
				LocalIps: []*nat.Nat44DNat_DNatConfig_StaticMappings_LocalIPs{
					{
						LocalIP:     ipAddresses[1],
						LocalPort:   ports[1],
						Probability: 35,
					},
					{
						LocalIP:     ipAddresses[2],
						LocalPort:   ports[2],
						Probability: 65,
					},
				},
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).ToNot(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatDeleteIdentityMapping(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Replies
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})
	ctx.MockVpp.MockReply(&nat_api.Nat44AddDelIdentityMappingReply{})

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:     0,
				IpAddress: ipAddresses[0],
				Port:      ports[0],
				Protocol:  nat.Protocol_TCP,
			},
		},
	}

	err := plugin.ConfigureDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found := plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeTrue())

	id := getIdMappingIdentifier(dNat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeTrue())

	err = plugin.DeleteDNat(dNat)
	Expect(err).To(BeNil())

	_, _, found = plugin.DNatIndices.LookupIdx(dNat.Label)
	Expect(found).To(BeFalse())

	id = getIdMappingIdentifier(dNat.IdMappings[0])
	_, _, found = plugin.DNatIdMappingIndices.LookupIdx(id)
	Expect(found).To(BeFalse())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDNatDeleteIdentityMappingNoInterfaceAndIP(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx)
	Expect(plugin).ToNot(BeNil())

	dNat := &nat.Nat44DNat_DNatConfig{
		Label: label,
		IdMappings: []*nat.Nat44DNat_DNatConfig_IdentityMappings{
			{
				VrfId:    0,
				Port:     ports[0],
				Protocol: nat.Protocol_TCP,
			},
		},
	}

	err := plugin.DeleteDNat(dNat)
	Expect(err).ToNot(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

// Auxiliary

func getNatConfigurator(ctx *vppcallmock.TestCtx) (*NatConfigurator, ifaceidx.SwIfIndexRW) {
	// Logger
	log := logrus.DefaultLogger()
	log.SetLevel(logging.DebugLevel)

	// Interface indices
	swIfIndices := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(log, "nat-configurator-test", "nat", nil))

	return &NatConfigurator{
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

func getForwardingConfig(fwd bool) *nat.Nat44Global {
	return &nat.Nat44Global{Forwarding: fwd}
}
