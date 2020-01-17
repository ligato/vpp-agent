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
func TestLinkOnly(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

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
				ToMicroservice: msNamePrefix + msName,
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
			Reference: msNamePrefix + msName,
		},
	}

	ms := ctx.startMicroservice(msName)
	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(
		vppTap,
		linuxTap,
	).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	Eventually(ctx.getValueStateClb(vppTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(ctx.getValueState(linuxTap)).To(Equal(kvscheduler.ValueState_CONFIGURED))
	Expect(ctx.pingFromVPP(linuxTapIPIgnored)).NotTo(Succeed()) // IP address was not set

	hasIP := func(tapLinkName netlink.Link, ipAddr string) bool {
		addrs, err := netlink.AddrList(tapLinkName, netlink.FAMILY_ALL)
		Expect(err).ToNot(HaveOccurred())
		for _, addr := range addrs {
			if addr.IP.String() == ipAddr {
				return true
			}
		}
		return false
	}

	leaveMs := ms.enterNetNs()
	tapLinkName, err := netlink.LinkByName(linuxTapHostname)
	Expect(err).ToNot(HaveOccurred())

	// agent didn't set IP address
	Expect(hasIP(tapLinkName, linuxTapIPIgnored)).To(BeFalse())

	// set IP and MAC addresses from outside of the agent
	ipAddr, _, err := utils.ParseIPAddr(linuxTapIPExternal+netMask, nil)
	Expect(err).ToNot(HaveOccurred())
	err = netlink.AddrAdd(tapLinkName, &netlink.Addr{IPNet: ipAddr})
	Expect(err).ToNot(HaveOccurred())
	hwAddr, err := ParseMAC(linuxTapHwExternal)
	Expect(err).ToNot(HaveOccurred())
	err = netlink.LinkSetHardwareAddr(tapLinkName, hwAddr)
	Expect(err).ToNot(HaveOccurred())
	leaveMs()

	// run downstream resync
	Expect(ctx.agentInSync()).To(BeTrue()) // everything in-sync even though the IP addr was added
	leaveMs = ms.enterNetNs()
	Expect(hasIP(tapLinkName, linuxTapIPIgnored)).To(BeFalse())
	Expect(hasIP(tapLinkName, linuxTapIPExternal)).To(BeTrue())
	link, err := netlink.LinkByName(linuxTapHostname)
	Expect(err).ToNot(HaveOccurred())
	Expect(link).ToNot(BeNil())
	Expect(link.Attrs().HardwareAddr.String()).To(Equal(linuxTapHwExternal))
	leaveMs()

	// test with ping
	Expect(ctx.pingFromVPP(linuxTapIPExternal)).To(Succeed())
	Expect(ctx.pingFromMs(msName, vppTapIP)).To(Succeed())
}
