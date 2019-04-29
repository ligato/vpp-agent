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

package vpp1901_test

import (
	"testing"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/interfaces"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/dhcp"
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
	checkIfState(1, result, eventsChan)
	notifChan <- &interfaces.SwInterfaceEvent{
		SwIfIndex:   2,
		AdminUpDown: 1,
		LinkUpDown:  0,
		Deleted:     1,
	}
	checkIfState(2, result, eventsChan)
	notifChan <- &interfaces.SwInterfaceEvent{
		SwIfIndex:   3,
		AdminUpDown: 0,
		LinkUpDown:  1,
		Deleted:     1,
	}
	checkIfState(3, result, eventsChan)
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
	checkDHCPLeases(1, result, leasesChChan)
	notifChan <- &dhcp.DHCPComplEvent{
		PID: 50,
		Lease: dhcp.DHCPLease{
			SwIfIndex:     2,
			State:         1,
			Hostname:      []byte("host2"),
			IsIPv6:        0,
			MaskWidth:     24,
			HostAddress:   []byte{10, 10, 10, 6},
			RouterAddress: []byte{10, 10, 10, 1},
			HostMac:       []byte{16, 16, 32, 32, 64, 64},
		},
	}
	checkDHCPLeases(2, result, leasesChChan)
	close(leasesChChan)
}

func checkDHCPLeases(SwIfIndex uint32, result *vppcalls.Lease, leasesChChan <-chan *vppcalls.Lease) {
	if SwIfIndex == 1 {
		Eventually(leasesChChan, 2).Should(Receive(&result))
		Expect(result.SwIfIndex).To(Equal(uint32(1)))
		Expect(result.State).To(Equal(uint8(1)))
		Expect(result.Hostname).To(BeEquivalentTo([]byte("host1")))
		Expect(result.IsIPv6).ToNot(BeTrue())
		Expect(result.MaskWidth).To(Equal(uint8(0)))
		Expect(result.HostAddress).To(BeEquivalentTo("10.10.10.5/24"))
		Expect(result.RouterAddress).To(BeEquivalentTo("10.10.10.1/24"))
		Expect(result.HostMac).To(BeEquivalentTo("10:10:20:20:30:30"))
	}
	if SwIfIndex == 2 {
		Eventually(leasesChChan, 2).Should(Receive(&result))
		Expect(result.SwIfIndex).To(Equal(uint32(2)))
		Expect(result.State).To(Equal(uint8(1)))
		Expect(result.Hostname).To(BeEquivalentTo([]byte("host2")))
		Expect(result.IsIPv6).ToNot(BeTrue())
		Expect(result.MaskWidth).To(Equal(uint8(0)))
		Expect(result.HostAddress).To(BeEquivalentTo("10.10.10.6/24"))
		Expect(result.RouterAddress).To(BeEquivalentTo("10.10.10.1/24"))
		Expect(result.HostMac).To(BeEquivalentTo("10:10:20:20:40:40"))
	}
}

func checkIfState(SwIfIndex uint32, result *vppcalls.InterfaceEvent, eventsChan <-chan *vppcalls.InterfaceEvent) {
	if SwIfIndex == 1 {
		Eventually(eventsChan, 2).Should(Receive(&result))
		Expect(result.SwIfIndex).To(Equal(uint32(1)))
		Expect(result.AdminState).To(Equal(uint8(1)))
		Expect(result.LinkState).To(Equal(uint8(1)))
		Expect(result.Deleted).To(BeTrue())
	}
	if SwIfIndex == 2 {
		Eventually(eventsChan, 2).Should(Receive(&result))
		Expect(result.SwIfIndex).To(Equal(uint32(2)))
		Expect(result.AdminState).To(Equal(uint8(1)))
		Expect(result.LinkState).To(Equal(uint8(0)))
		Expect(result.Deleted).To(BeTrue())
	}
	if SwIfIndex == 3 {
		Eventually(eventsChan, 2).Should(Receive(&result))
		Expect(result.SwIfIndex).To(Equal(uint32(3)))
		Expect(result.AdminState).To(Equal(uint8(0)))
		Expect(result.LinkState).To(Equal(uint8(1)))
		Expect(result.Deleted).To(BeTrue())
	}
}
