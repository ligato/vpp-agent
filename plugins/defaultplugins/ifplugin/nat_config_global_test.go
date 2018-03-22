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
	"testing"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	nat_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var ifNames = []string{"if1", "if2", "if3"}
var ipAddresses = []string{"10.0.0.1", "10.0.0.2", "125.0.0.1", "172.125.0.1"}

func TestNatConfiguratorEnableForwarding(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx.Connection)
	Expect(plugin).ToNot(BeNil())

	// Prepare replies
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})

	// Start
	err := plugin.Init()
	Expect(err).To(BeNil())
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
		Forwarding: true,
	})
	Expect(err).To(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorDisableForwarding(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx.Connection)
	Expect(plugin).ToNot(BeNil())

	// Prepare replies
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})

	// Start
	err := plugin.Init()
	Expect(err).To(BeNil())
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
		Forwarding: false,
	})
	Expect(err).To(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func TestNatConfiguratorModifyForwarding(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	// Configurator
	plugin, _ := getNatConfigurator(ctx.Connection)
	Expect(plugin).ToNot(BeNil())

	// Prepare replies
	// Create
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})
	// Modify
	ctx.MockVpp.MockReply(&nat_api.Nat44ForwardingEnableDisableReply{})

	// Start
	err := plugin.Init()
	Expect(err).To(BeNil())

	oldCfg := &nat.Nat44Global{
		Forwarding: true,
	}
	newCfg := &nat.Nat44Global{
		Forwarding: false,
	}

	// Set config
	err = plugin.SetNatGlobalConfig(oldCfg)
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
	plugin, swIfIndices := getNatConfigurator(ctx.Connection)

	// Prepare interface indices
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)

	// Enable two interfaces
	err := plugin.Init()
	Expect(err).To(BeNil())
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
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
	plugin, swIfIndices := getNatConfigurator(ctx.Connection)

	// Prepare interface indices
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)

	// Enable two output feature interfaces
	err := plugin.Init()
	Expect(err).To(BeNil())
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
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
	plugin, swIfIndices := getNatConfigurator(ctx.Connection)

	// Prepare interface indices
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)
	swIfIndices.RegisterName(ifNames[2], 3, nil)

	err := plugin.Init()
	Expect(err).To(BeNil())

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
	err = plugin.SetNatGlobalConfig(oldCfg)
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
	plugin, swIfIndices := getNatConfigurator(ctx.Connection)

	// Start
	err := plugin.Init()
	Expect(err).To(BeNil())
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
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
	plugin, _ := getNatConfigurator(ctx.Connection)

	// Start
	err := plugin.Init()
	Expect(err).To(BeNil())
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
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
	plugin, _ := getNatConfigurator(ctx.Connection)

	// Start
	err := plugin.Init()
	Expect(err).To(BeNil())

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

	err = plugin.SetNatGlobalConfig(oldCfg)
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
	plugin, _ := getNatConfigurator(ctx.Connection)

	// No IP address provided
	err := plugin.Init()
	Expect(err).To(BeNil())
	err = plugin.SetNatGlobalConfig(&nat.Nat44Global{
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
	plugin, swIfIndices := getNatConfigurator(ctx.Connection)

	// Register required interfaces
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	swIfIndices.RegisterName(ifNames[1], 2, nil)

	// Config to remove
	err := plugin.Init()
	Expect(err).To(BeNil())
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
	err = plugin.SetNatGlobalConfig(natGlobalCfg)
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
	plugin, _ := getNatConfigurator(ctx.Connection)

	// Config to remove
	err := plugin.Init()
	Expect(err).To(BeNil())
	natGlobalCfg := &nat.Nat44Global{
		NatInterfaces: []*nat.Nat44Global_NatInterfaces{},
		AddressPools:  []*nat.Nat44Global_AddressPools{},
	}

	// Delete empty global config
	err = plugin.DeleteNatGlobalConfig(natGlobalCfg)
	Expect(err).To(BeNil())
	Expect(plugin.globalNAT).To(BeNil())

	// Close
	err = plugin.Close()
	Expect(err).To(BeNil())
}

func getNatConfigurator(connection govppmux.API) (*NatConfigurator, ifaceidx.SwIfIndexRW) {
	// Logger
	log := logrus.DefaultLogger()
	log.SetLevel(logging.DebugLevel)

	// Interface indices
	swIfIndices := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(log, "nat-configurator-test", "nat", nil))

	return &NatConfigurator{
		Log:         log,
		GoVppmux:    connection,
		SwIfIndexes: swIfIndices,
	}, swIfIndices
}
