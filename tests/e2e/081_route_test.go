//  Copyright (c) 2022 Cisco and/or its affiliates.
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

package e2e

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// TestIPv4Routes tests L3 routes in the default VRF and for various scopes
func TestIPv4Routes(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		// first subnet
		msName1     = "microservice1"
		subnet1     = "10.0.0.0/24"
		tap1IP      = "10.0.0.1"
		linuxTap1IP = "10.0.0.2"
		tap1Label   = "tap-1"

		// second subnet
		msName2     = "microservice2"
		subnet2     = "20.0.0.0/24"
		tap2IP      = "20.0.0.1"
		linuxTap2IP = "20.0.0.2"
		tap2Label   = "tap-2"

		suffix = "/24"
	)

	// TAP interface for the first subnet
	vppTap1 := &vpp_interfaces.Interface{
		Name:        tap1Label,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{tap1IP + suffix},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: MsNamePrefix + msName1,
			},
		},
	}
	linuxTap1 := &linux_interfaces.Interface{
		Name:        tap1Label,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTap1IP + suffix},
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: tap1Label,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName1,
		},
	}

	// TAP interfaces for the second subnet
	vppTap2 := &vpp_interfaces.Interface{
		Name:        tap2Label,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{tap2IP + suffix},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: MsNamePrefix + msName2,
			},
		},
	}
	linuxTap2 := &linux_interfaces.Interface{
		Name:        tap2Label,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTap2IP + suffix},
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: tap2Label,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName2,
		},
	}

	// Routes
	subnet1LinuxRoute := &linux_l3.Route{
		OutgoingInterface: tap1Label,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        subnet2,
		GwAddr:            tap1IP,
	}
	subnet2LinuxRoute := &linux_l3.Route{
		OutgoingInterface: tap2Label,
		Scope:             linux_l3.Route_GLOBAL,
		DstNetwork:        subnet1,
		GwAddr:            tap2IP,
	}
	subnet2LinuxLinkRoute := &linux_l3.Route{
		OutgoingInterface: tap2Label,
		Scope:             linux_l3.Route_LINK,
		DstNetwork:        subnet1,
	}

	ctx.StartMicroservice(msName1)
	ctx.StartMicroservice(msName2)

	// configure everything in one resync
	err := ctx.GenericClient().ResyncConfig(
		vppTap1, linuxTap1,
		vppTap2, linuxTap2,
		subnet1LinuxRoute, subnet2LinuxRoute,
	)
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(linuxTap1)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(vppTap2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(linuxTap2)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(subnet1LinuxRoute)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(subnet2LinuxRoute)).To(Equal(kvscheduler.ValueState_CONFIGURED))

	ctx.Expect(ctx.GetRunningMicroservice(msName1).Ping("20.0.0.2")).To(Succeed())
	ctx.Expect(ctx.GetRunningMicroservice(msName2).Ping("10.0.0.2")).To(Succeed())

	// keep the current number of routes before the update
	numLinuxRoutes := ctx.NumValues(&linux_l3.Route{}, kvs.SBView)

	// reconfigure subnet 1 route as link local
	err = ctx.GenericClient().ChangeRequest().Update(
		subnet2LinuxLinkRoute,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.GetRunningMicroservice(msName1).Ping("20.0.0.2")).NotTo(Succeed())
	ctx.Expect(ctx.GetRunningMicroservice(msName2).Ping("10.0.0.2")).NotTo(Succeed())

	// route count should be unchanged
	ctx.Expect(ctx.NumValues(&linux_l3.Route{}, kvs.SBView)).To(Equal(numLinuxRoutes))
}
