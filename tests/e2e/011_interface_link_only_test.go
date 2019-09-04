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
	"testing"

	. "github.com/onsi/gomega"

	"github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/namespace"
	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/netalloc/utils"
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
	Expect(err).To(BeNil())

	Eventually(ctx.getValueStateClb(vppTap), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED))
	Expect(ctx.getValueState(linuxTap)).To(Equal(kvs.ValueState_CONFIGURED))
	Expect(ctx.pingFromVPP(linuxTapIPIgnored)).ToNot(BeNil()) // IP address was not set

	ifHandler := linuxcalls.NewNetLinkHandler()
	hasIP := func(ipAddr string) bool {
		addrs, err := ifHandler.GetAddressList(linuxTapHostname)
		Expect(err).To(BeNil())
		for _, addr := range addrs {
			if addr.IP.String() == ipAddr {
				return true
			}
		}
		return false
	}

	leaveMs := ms.enterNetNs()
	// agent didn't set IP address
	Expect(hasIP(linuxTapIPIgnored)).To(BeFalse())

	// set IP and MAC addresses from outside of the agent
	ipAddr, _, err := utils.ParseIPAddr(linuxTapIPExternal+netMask, nil)
	Expect(err).To(BeNil())
	err = ifHandler.AddInterfaceIP(linuxTapHostname, ipAddr)
	Expect(err).To(BeNil())
	err = ifHandler.SetInterfaceMac(linuxTapHostname, linuxTapHwExternal)
	Expect(err).To(BeNil())
	leaveMs()

	// run downstream resync
	Expect(ctx.agentInSync()).To(BeTrue()) // everything in-sync even though the IP addr was added
	leaveMs = ms.enterNetNs()
	Expect(hasIP(linuxTapIPIgnored)).To(BeFalse())
	Expect(hasIP(linuxTapIPExternal)).To(BeTrue())
	link, err := ifHandler.GetLinkByName(linuxTapHostname)
	Expect(err).To(BeNil())
	Expect(link).ToNot(BeNil())
	Expect(link.Attrs().HardwareAddr.String()).To(Equal(linuxTapHwExternal))
	leaveMs()

	// test with ping
	Expect(ctx.pingFromVPP(linuxTapIPExternal)).To(BeNil())
	Expect(ctx.pingFromMs(msName, vppTapIP)).To(BeNil())
}
