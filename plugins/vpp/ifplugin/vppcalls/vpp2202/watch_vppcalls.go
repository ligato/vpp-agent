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

package vpp2202

import (
	"context"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	govppapi "go.fd.io/govpp/api"

	vpp_dhcp "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/dhcp"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
)

var (
	// EventDeliverTimeout defines maximum time to deliver event upstream.
	EventDeliverTimeout = time.Second
	// NotifChanBufferSize defines size of notification channel buffer.
	NotifChanBufferSize = 10
)

func (h *InterfaceVppHandler) WatchInterfaceEvents(ctx context.Context, eventsCh chan<- *vppcalls.InterfaceEvent) error {
	notifChan := make(chan govppapi.Message, NotifChanBufferSize)

	// subscribe to SwInterfaceEvent notifications
	sub, err := h.callsChannel.SubscribeNotification(notifChan, &vpp_ifs.SwInterfaceEvent{})
	if err != nil {
		return errors.Errorf("subscribing to VPP notification (sw_interface_event) failed: %v", err)
	}
	unsub := func() {
		if err := sub.Unsubscribe(); err != nil {
			h.log.Warnf("unsubscribing VPP notification (sw_interface_event) failed: %v", err)
		}
	}

	go func() {
		h.log.Debugf("start watching interface events")
		defer h.log.Debugf("done watching interface events (%v)", ctx.Err())

		for {
			select {
			case e, open := <-notifChan:
				if !open {
					h.log.Debugf("interface events channel was closed")
					unsub()
					return
				}

				ifEvent, ok := e.(*vpp_ifs.SwInterfaceEvent)
				if !ok {
					h.log.Debugf("unexpected notification type: %#v", ifEvent)
					continue
				}

				// try to send event
				select {
				case eventsCh <- toInterfaceEvent(ifEvent):
					// sent ok
				case <-ctx.Done():
					unsub()
					return
				default:
					// channel full send event in goroutine for later processing
					go func() {
						select {
						case eventsCh <- toInterfaceEvent(ifEvent):
							// sent ok
						case <-time.After(EventDeliverTimeout):
							h.log.Warnf("unable to deliver interface event, dropping it: %+v", ifEvent)
						}
					}()
				}
			case <-ctx.Done():
				unsub()
				return
			}
		}
	}()

	// enable interface events from VPP
	if _, err := h.interfaces.WantInterfaceEvents(ctx, &vpp_ifs.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}); err != nil {
		if errors.Is(err, govppapi.VPPApiError(govppapi.INVALID_REGISTRATION)) {
			h.log.Warnf("already subscribed to interface events: %v", err)
			return nil
		}
		return errors.Errorf("failed to watch interface events: %v", err)
	}

	return nil
}

func (h *InterfaceVppHandler) WatchDHCPLeases(ctx context.Context, leasesCh chan<- *vppcalls.Lease) error {
	notifChan := make(chan govppapi.Message, NotifChanBufferSize)

	// subscribe for receiving DHCPComplEvent notifications
	sub, err := h.callsChannel.SubscribeNotification(notifChan, &vpp_dhcp.DHCPComplEvent{})
	if err != nil {
		return errors.Errorf("subscribing to VPP notification (dhcp_compl_event) failed: %v", err)
	}
	unsub := func() {
		if err := sub.Unsubscribe(); err != nil {
			h.log.Warnf("unsubscribing VPP notification (dhcp_compl_event) failed: %v", err)
		}
	}

	go func() {
		h.log.Debugf("start watching DHCP leases")
		defer h.log.Debugf("done watching DHCP lease (%v)", ctx.Err())

		for {
			select {
			case e, open := <-notifChan:
				if !open {
					h.log.Debugf("interface notification channel was closed")
					unsub()
					return
				}

				dhcpEvent, ok := e.(*vpp_dhcp.DHCPComplEvent)
				if !ok {
					h.log.Debugf("unexpected notification type: %#v", dhcpEvent)
					continue
				}

				// try to send event
				select {
				case leasesCh <- toDHCPLease(dhcpEvent):
					// sent ok
				case <-ctx.Done():
					unsub()
					return
				default:
					// channel full send event in goroutine for later processing
					go func() {
						select {
						case leasesCh <- toDHCPLease(dhcpEvent):
							// sent ok
						case <-time.After(EventDeliverTimeout):
							h.log.Warnf("unable to deliver DHCP lease event, dropping it: %+v", dhcpEvent)
						}
					}()
				}
			case <-ctx.Done():
				unsub()
				return
			}
		}
	}()

	return nil
}

func toInterfaceEvent(ifEvent *vpp_ifs.SwInterfaceEvent) *vppcalls.InterfaceEvent {
	event := &vppcalls.InterfaceEvent{
		SwIfIndex: uint32(ifEvent.SwIfIndex),
		Deleted:   ifEvent.Deleted,
	}
	if ifEvent.Flags&interface_types.IF_STATUS_API_FLAG_ADMIN_UP == interface_types.IF_STATUS_API_FLAG_ADMIN_UP {
		event.AdminState = 1
	}
	if ifEvent.Flags&interface_types.IF_STATUS_API_FLAG_LINK_UP == interface_types.IF_STATUS_API_FLAG_LINK_UP {
		event.LinkState = 1
	}
	return event
}

func toDHCPLease(dhcpEvent *vpp_dhcp.DHCPComplEvent) *vppcalls.Lease {
	lease := dhcpEvent.Lease
	return &vppcalls.Lease{
		SwIfIndex:     uint32(lease.SwIfIndex),
		State:         uint8(lease.State),
		Hostname:      strings.TrimRight(lease.Hostname, "\x00"),
		IsIPv6:        lease.IsIPv6,
		HostAddress:   dhcpAddressToString(lease.HostAddress, uint32(lease.MaskWidth), lease.IsIPv6),
		RouterAddress: dhcpAddressToString(lease.RouterAddress, uint32(lease.MaskWidth), lease.IsIPv6),
		HostMac:       net.HardwareAddr(lease.HostMac[:]).String(),
	}
}
