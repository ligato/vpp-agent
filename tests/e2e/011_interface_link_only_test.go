//  Copyright (c) 2019 Cisco and/or its affiliates.
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
	. "net"
	"testing"

	"github.com/vishvananda/netlink"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/plugins/netalloc/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

// configure only link on the Linux side of the interface and leave addresses
// untouched during resync.
func TestInterfaceLinkOnlyTap(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const (
		vppTapName         = "vpp-tap"
		linuxTapName       = "linux-tap"
		linuxTapHostname   = "tap"
		vppTapIP           = "192.168.1.1"
		linuxTapIPIgnored  = "192.168.1.2"
		linuxTapIPExternal = "192.168.1.3"
		linuxTapHwIgnored  = "22:22:22:33:33:33"
		linuxTapHwExternal = "44:44:44:55:55:55"
		netMask            = "/24"
		msName             = "microservice1"
	)

	vppTap := &vpp_interfaces.Interface{
		Name:        vppTapName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTapIP + netMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: MsNamePrefix + msName,
			},
		},
	}
	linuxTap := &linux_interfaces.Interface{
		Name:        linuxTapName,
		LinkOnly:    true, // <--- link only
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		IpAddresses: []string{linuxTapIPIgnored + netMask},
		HostIfName:  linuxTapHostname,
		PhysAddress: linuxTapHwIgnored,
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapName,
			},
		},
		Namespace: &linux_namespace.NetNamespace{
			Type:      linux_namespace.NetNamespace_MICROSERVICE,
			Reference: MsNamePrefix + msName,
		},
	}

	ms := ctx.StartMicroservice(msName)
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		vppTap,
		linuxTap,
	).Send(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())

	ctx.Eventually(ctx.GetValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.GetValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	ctx.Expect(ctx.PingFromVPP(linuxTapIPIgnored)).NotTo(Succeed()) // IP address was not set

	hasIP := func(tapLinkName netlink.Link, ipAddr string) bool {
		addrs, err := netlink.AddrList(tapLinkName, netlink.FAMILY_ALL)
		ctx.Expect(err).ToNot(HaveOccurred())
		for _, addr := range addrs {
			if addr.IP.String() == ipAddr {
				return true
			}
		}
		return false
	}

	leaveMs := ms.enterNetNs()
	tapLinkName, err := netlink.LinkByName(linuxTapHostname)
	ctx.Expect(err).ToNot(HaveOccurred())

	// agent didn't set IP address
	ctx.Expect(hasIP(tapLinkName, linuxTapIPIgnored)).To(BeFalse())

	// set IP and MAC addresses from outside of the agent
	ipAddr, _, err := utils.ParseIPAddr(linuxTapIPExternal+netMask, nil)
	ctx.Expect(err).ToNot(HaveOccurred())
	err = netlink.AddrAdd(tapLinkName, &netlink.Addr{IPNet: ipAddr})
	ctx.Expect(err).ToNot(HaveOccurred())
	hwAddr, err := ParseMAC(linuxTapHwExternal)
	ctx.Expect(err).ToNot(HaveOccurred())
	err = netlink.LinkSetHardwareAddr(tapLinkName, hwAddr)
	ctx.Expect(err).ToNot(HaveOccurred())
	leaveMs()

	// run downstream resync
	ctx.Expect(ctx.AgentInSync()).To(BeTrue()) // everything in-sync even though the IP addr was added
	leaveMs = ms.enterNetNs()
	ctx.Expect(hasIP(tapLinkName, linuxTapIPIgnored)).To(BeFalse())
	ctx.Expect(hasIP(tapLinkName, linuxTapIPExternal)).To(BeTrue())
	link, err := netlink.LinkByName(linuxTapHostname)
	ctx.Expect(err).ToNot(HaveOccurred())
	ctx.Expect(link).ToNot(BeNil())
	ctx.Expect(link.Attrs().HardwareAddr.String()).To(Equal(linuxTapHwExternal))
	leaveMs()

	// test with ping
	ctx.Expect(ctx.PingFromVPP(linuxTapIPExternal)).To(Succeed())
	ctx.Expect(ctx.PingFromMs(msName, vppTapIP)).To(Succeed())
}
