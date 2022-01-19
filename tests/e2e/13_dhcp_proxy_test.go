// Copyright (c) 2020 Pantheon.tech
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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

func TestDhcpProxy(t *testing.T) {
	if !supportsLinuxVRF() {
		t.Skip("Linux VRFs are not supported")
	}

	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		vrf1ID            = 1
		vrf2ID            = 2
		vrf1Label         = "vrf-1"
		vrf2Label         = "vrf-2"
		vppTap1Name       = "vpp-tap1"
		vppTap2Name       = "vpp-tap2"
		linuxTap1Name     = "linux-tap1"
		linuxTap2Name     = "linux-tap2"
		linuxTap1Hostname = "tap1"
		linuxTap2Hostname = "tap2"
		vppTapIP          = "192.168.1.1"
		linuxTapIP        = "192.168.1.2"
		netMask           = "/30"
		msName            = "microservice1"
	)

	vppVrf1 := &vpp_l3.VrfTable{
		Id:       vrf1ID,
		Label:    vrf1Label,
		Protocol: vpp_l3.VrfTable_IPV4,
	}
	vppVrf2 := &vpp_l3.VrfTable{
		Id:       vrf2ID,
		Label:    vrf2Label,
		Protocol: vpp_l3.VrfTable_IPV4,
	}
	linuxVrf1 := &linux_interfaces.Interface{
		Name:    vrf1Label,
		Type:    linux_interfaces.Interface_VRF_DEVICE,
		Enabled: true,
		Link: &linux_interfaces.Interface_VrfDev{
			VrfDev: &linux_interfaces.VrfDevLink{
				RoutingTable: vrf1ID,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}
	linuxVrf2 := &linux_interfaces.Interface{
		Name:    vrf2Label,
		Type:    linux_interfaces.Interface_VRF_DEVICE,
		Enabled: true,
		Link: &linux_interfaces.Interface_VrfDev{
			VrfDev: &linux_interfaces.VrfDevLink{
				RoutingTable: vrf2ID,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}
	vppTap1 := &vpp_interfaces.Interface{
		Name:        vppTap1Name,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		Vrf:         vrf1ID,
		IpAddresses: []string{vppTapIP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: MsNamePrefix + msName,
			},
		},
	}
	linuxTap1 := &linux_interfaces.Interface{
		Name:               linuxTap1Name,
		Type:               linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:            true,
		IpAddresses:        []string{linuxTapIP + netMask},
		HostIfName:         linuxTap1Hostname,
		VrfMasterInterface: vrf1Label,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTap1Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}
	vppTap2 := &vpp_interfaces.Interface{
		Name:        vppTap2Name,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		Vrf:         vrf2ID,
		IpAddresses: []string{vppTapIP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: MsNamePrefix + msName,
			},
		},
	}
	linuxTap2 := &linux_interfaces.Interface{
		Name:               linuxTap2Name,
		Type:               linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:            true,
		IpAddresses:        []string{linuxTapIP + netMask},
		HostIfName:         linuxTap2Hostname,
		VrfMasterInterface: vrf2Label,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTap2Name,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}

	dhcpProxy1 := &vpp_l3.DHCPProxy{
		SourceIpAddress: vppTapIP,
		RxVrfId:         vrf1ID,
		Servers: []*vpp_l3.DHCPProxy_DHCPServer{
			{
				VrfId:     vrf1ID,
				IpAddress: linuxTapIP,
			},
		},
	}
	dhcpProxy2 := &vpp_l3.DHCPProxy{
		SourceIpAddress: vppTapIP,
		RxVrfId:         vrf2ID,
		Servers: []*vpp_l3.DHCPProxy_DHCPServer{
			{
				VrfId:     vrf2ID,
				IpAddress: linuxTapIP,
			},
		},
	}

	ctx.StartMicroservice(msName)

	dhcpProxies := func() string {
		output, err := ctx.ExecVppctl("show", "dhcp", "proxy")
		ctx.Expect(err).ShouldNot(HaveOccurred())
		return output
	}
	dhcpProxyRegexp := func(vrf int) string {
		return fmt.Sprintf("%d[ ]+%s[ ]+%d,%s", vrf, vppTapIP, vrf, linuxTapIP)
	}

	err := ctx.GenericClient().ChangeRequest().Update(
		vppVrf1,
		vppVrf2,
		linuxVrf1,
		linuxVrf2,
		vppTap1,
		vppTap2,
		linuxTap1,
		linuxTap2,
		dhcpProxy1,
		dhcpProxy2,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(vppTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(ctx.GetValueStateClb(vppTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(ctx.GetValueStateClb(linuxTap1)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(ctx.GetValueStateClb(linuxTap2)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(ctx.GetValueStateClb(dhcpProxy1)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Eventually(ctx.GetValueStateClb(dhcpProxy2)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP, PingWithSourceInterface(linuxTap1Hostname))).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP, PingWithSourceInterface(linuxTap2Hostname))).To(Succeed())
	ctx.Expect(ctx.AgentInSync()).To(BeTrue())

	ctx.Expect(dhcpProxies()).Should(MatchRegexp(dhcpProxyRegexp(vrf1ID)))
	ctx.Expect(dhcpProxies()).Should(MatchRegexp(dhcpProxyRegexp(vrf2ID)))

	err = ctx.GenericClient().ChangeRequest().Delete(
		dhcpProxy1,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Expect(ctx.AgentInSync()).To(BeTrue())
	ctx.Expect(dhcpProxies()).ShouldNot(MatchRegexp(dhcpProxyRegexp(vrf1ID)))
	ctx.Expect(dhcpProxies()).Should(MatchRegexp(dhcpProxyRegexp(vrf2ID)))

}
