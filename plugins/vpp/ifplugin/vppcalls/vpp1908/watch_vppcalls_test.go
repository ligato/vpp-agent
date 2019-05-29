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

package vpp1908_test

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/interfaces"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/dhcp"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"

	. "github.com/onsi/gomega"
)

func TestWatchInterfaceEvents(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&interfaces.WantInterfaceEventsReply{})
	eventsChan := make(chan *vppcalls.InterfaceEvent)
	err := ifHandler.WatchInterfaceEvents(eventsChan)
	notifChan := ctx.MockChannel.GetChannel()
	Expect(notifChan).ToNot(BeNil())
	Expect(err).To(BeNil())

	notifChan <- &interfaces.SwInterfaceEvent{
		SwIfIndex:   1,
		AdminUpDown: 1,
		LinkUpDown:  1,
		Deleted:     1,
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
		SwIfIndex:   2,
		AdminUpDown: 1,
		LinkUpDown:  0,
		Deleted:     1,
	}
	result = &vppcalls.InterfaceEvent{}
	Eventually(eventsChan, 2).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.InterfaceEvent{SwIfIndex: 2, AdminState: 1, LinkState: 0, Deleted: true}))

	notifChan <- &interfaces.SwInterfaceEvent{
		SwIfIndex:   3,
		AdminUpDown: 0,
		LinkUpDown:  1,
		Deleted:     0,
	}
	result = &vppcalls.InterfaceEvent{}
	Eventually(eventsChan, 2).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.InterfaceEvent{
		SwIfIndex:  3,
		AdminState: 0,
		LinkState:  1,
		Deleted:    false,
	}))

	close(notifChan)
}

func TestWatchDHCPLeases(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	leasesChChan := make(chan *vppcalls.Lease)
	err := ifHandler.WatchDHCPLeases(leasesChChan)
	notifChan := ctx.MockChannel.GetChannel()
	Expect(notifChan).ToNot(BeNil())
	Expect(err).To(BeNil())

	notifChan <- &dhcp.DHCPComplEvent{
		PID: 50,
		Lease: dhcp.DHCPLease{
			SwIfIndex:     1,
			State:         1,
			Hostname:      []byte("host1"),
			IsIPv6:        0,
			MaskWidth:     24,
			HostAddress:   []byte{10, 10, 10, 5},
			RouterAddress: []byte{10, 10, 10, 1},
			HostMac:       []byte{16, 16, 32, 32, 48, 48},
		},
	}
	var result *vppcalls.Lease
	Eventually(leasesChChan, 2).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.Lease{
		SwIfIndex:     1,
		State:         1,
		Hostname:      "host1",
		HostAddress:   "10.10.10.5/24",
		RouterAddress: "10.10.10.1/24",
		HostMac:       "10:10:20:20:30:30",
	}))

	notifChan <- &dhcp.DHCPComplEvent{
		PID: 50,
		Lease: dhcp.DHCPLease{
			SwIfIndex:     2,
			State:         0,
			Hostname:      []byte("host2"),
			IsIPv6:        1,
			MaskWidth:     24,
			HostAddress:   []byte{10, 10, 10, 6},
			RouterAddress: []byte{10, 10, 10, 1},
			HostMac:       []byte{16, 16, 32, 32, 64, 64},
		},
	}
	Eventually(leasesChChan, 2).Should(Receive(&result))
	Expect(result).To(Equal(&vppcalls.Lease{
		SwIfIndex:     2,
		Hostname:      "host2",
		IsIPv6:        true,
		HostAddress:   "10.10.10.6/24",
		RouterAddress: "10.10.10.1/24",
		HostMac:       "10:10:20:20:40:40",
	}))

	close(leasesChChan)
}
