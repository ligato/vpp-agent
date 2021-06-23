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

package vpp2106_test

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/dhcp"
	interfaces "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
)

func TestWatchInterfaceEvents(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&interfaces.WantInterfaceEventsReply{})
	eventsChan := make(chan *vppcalls.InterfaceEvent)
	err := ifHandler.WatchInterfaceEvents(ctx.Context, eventsChan)
	notifChan := ctx.MockChannel.GetChannel()
	Expect(notifChan).ToNot(BeNil())
	Expect(err).To(BeNil())

	notifChan <- &interfaces.SwInterfaceEvent{
		SwIfIndex: 1,
		Flags:     3,
		Deleted:   true,
	}
	var result *vppcalls.InterfaceEvent
	Eventually(eventsChan, 2).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.InterfaceEvent{
		SwIfIndex:  1,
		AdminState: 1,
		LinkState:  1,
		Deleted:    true,
	}))

	notifChan <- &interfaces.SwInterfaceEvent{
		SwIfIndex: 2,
		Flags:     1,
		Deleted:   true,
	}
	result = &vppcalls.InterfaceEvent{}
	Eventually(eventsChan, 2).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.InterfaceEvent{SwIfIndex: 2, AdminState: 1, LinkState: 0, Deleted: true}))

	notifChan <- &interfaces.SwInterfaceEvent{
		SwIfIndex: 3,
		Flags:     3,
		Deleted:   false,
	}
	result = &vppcalls.InterfaceEvent{}
	Eventually(eventsChan, 2).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.InterfaceEvent{
		SwIfIndex:  3,
		AdminState: 1,
		LinkState:  1,
		Deleted:    false,
	}))

	close(notifChan)
}

func TestWatchDHCPLeases(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	leasesChChan := make(chan *vppcalls.Lease)
	err := ifHandler.WatchDHCPLeases(ctx.Context, leasesChChan)
	notifChan := ctx.MockChannel.GetChannel()
	Expect(notifChan).ToNot(BeNil())
	Expect(err).To(BeNil())

	var hostAddr, routerAddr [16]byte
	copy(hostAddr[:], net.ParseIP("10.10.10.5").To4())
	copy(routerAddr[:], net.ParseIP("10.10.10.1").To4())

	notifChan <- &dhcp.DHCPComplEvent{
		PID: 50,
		Lease: dhcp.DHCPLease{
			SwIfIndex:     1,
			State:         1,
			Hostname:      "host1",
			IsIPv6:        false,
			MaskWidth:     24,
			HostAddress:   ip_types.Address{Un: ip_types.AddressUnion{XXX_UnionData: hostAddr}},
			RouterAddress: ip_types.Address{Un: ip_types.AddressUnion{XXX_UnionData: routerAddr}},
			HostMac:       [6]byte{16, 16, 32, 32, 48, 48},
		},
	}
	var result *vppcalls.Lease
	Eventually(leasesChChan, 50).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.Lease{
		SwIfIndex:     1,
		State:         1,
		Hostname:      "host1",
		HostAddress:   "10.10.10.5/24",
		RouterAddress: "10.10.10.1/24",
		HostMac:       "10:10:20:20:30:30",
	}))

	copy(hostAddr[:], net.ParseIP("1234::").To16())
	copy(routerAddr[:], net.ParseIP("abcd::").To16())

	notifChan <- &dhcp.DHCPComplEvent{
		PID: 50,
		Lease: dhcp.DHCPLease{
			SwIfIndex:     2,
			State:         0,
			Hostname:      "host2",
			IsIPv6:        true,
			MaskWidth:     64,
			HostAddress:   ip_types.Address{Un: ip_types.AddressUnion{XXX_UnionData: hostAddr}},
			RouterAddress: ip_types.Address{Un: ip_types.AddressUnion{XXX_UnionData: routerAddr}},
			HostMac:       [6]byte{16, 16, 32, 32, 64, 64},
		},
	}
	Eventually(leasesChChan, 2).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.Lease{
		SwIfIndex:     2,
		Hostname:      "host2",
		IsIPv6:        true,
		HostAddress:   "1234::/64",
		RouterAddress: "abcd::/64",
		HostMac:       "10:10:20:20:40:40",
	}))

	close(leasesChChan)
}
