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

package vpp2001

import (
	"net"
	"os"
	"strings"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/pkg/errors"

	vpp_dhcp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/dhcp"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
)

var InterfaceEventTimeout = time.Second

func (h *InterfaceVppHandler) WatchInterfaceEvents(events chan<- *vppcalls.InterfaceEvent) error {
	notifChan := make(chan govppapi.Message, 10)

	// subscribe for receiving SwInterfaceEvents notifications
	vppNotifSubs, err := h.callsChannel.SubscribeNotification(notifChan, &vpp_ifs.SwInterfaceEvent{})
	if err != nil {
		return errors.Errorf("failed to subscribe VPP notification (sw_interface_event): %v", err)
	}
	_ = vppNotifSubs

	go func() {
		for {
			select {
			case e, ok := <-notifChan:
				if !ok {
					h.log.Debugf("interface notification channel was closed")
					return
				}
				ifEvent, ok := e.(*vpp_ifs.SwInterfaceEvent)
				if !ok {
					continue
				}
				event := &vppcalls.InterfaceEvent{
					SwIfIndex:  uint32(ifEvent.SwIfIndex),
					AdminState: boolToUint(ifEvent.Flags > 0),
					LinkState:  boolToUint(ifEvent.Flags > 1),
					Deleted:    ifEvent.Deleted,
				}
				// send event in goroutine for quick processing
				go func() {
					select {
					case events <- event:
						// sent ok
					case <-time.After(InterfaceEventTimeout):
						h.log.Warnf("unable to deliver interface event, dropping it")
					}
				}()
			}
		}
	}()

	// enable interface state notifications from VPP
	wantIfEventsReply := &vpp_ifs.WantInterfaceEventsReply{}
	err = h.callsChannel.SendRequest(&vpp_ifs.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(wantIfEventsReply)
	if err != nil {
		if err == govppapi.VPPApiError(govppapi.INVALID_REGISTRATION) {
			h.log.Warnf("already registered for watch interface events: %v", err)
			return nil
		}
		return errors.Errorf("failed to watch interface events: %v", err)
	}

	return nil
}

func (h *InterfaceVppHandler) WatchDHCPLeases(leasesCh chan<- *vppcalls.Lease) error {
	notifChan := make(chan govppapi.Message)

	// subscribe for receiving SwInterfaceEvents notifications
	vppNotifSubs, err := h.callsChannel.SubscribeNotification(notifChan, &vpp_dhcp.DHCPComplEvent{})
	if err != nil {
		return errors.Errorf("failed to subscribe VPP notification (sw_interface_event): %v", err)
	}
	_ = vppNotifSubs

	go func() {
		for {
			select {
			case e := <-notifChan:
				dhcpEvent, ok := e.(*vpp_dhcp.DHCPComplEvent)
				if !ok {
					continue
				}
				lease := dhcpEvent.Lease
				leasesCh <- &vppcalls.Lease{
					SwIfIndex:     uint32(lease.SwIfIndex),
					State:         uint8(lease.State),
					Hostname:      strings.TrimRight(lease.Hostname, "\x00"),
					IsIPv6:        lease.IsIPv6,
					HostAddress:   dhcpAddressToString(lease.HostAddress, uint32(lease.MaskWidth), lease.IsIPv6),
					RouterAddress: dhcpAddressToString(lease.RouterAddress, uint32(lease.MaskWidth), lease.IsIPv6),
					HostMac:       net.HardwareAddr(lease.HostMac[:]).String(),
				}
			}
		}
	}()

	return nil
}
