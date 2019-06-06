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

package vpp1908

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/pkg/errors"

	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/dhcp"
	binapi_interfaces "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

var InterfaceEventTimeout = time.Second

func (h *InterfaceVppHandler) WatchInterfaceEvents(events chan<- *vppcalls.InterfaceEvent) error {
	notifChan := make(chan govppapi.Message, 10)

	// subscribe for receiving SwInterfaceEvents notifications
	vppNotifSubs, err := h.callsChannel.SubscribeNotification(notifChan, &binapi_interfaces.SwInterfaceEvent{})
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
				ifEvent, ok := e.(*binapi_interfaces.SwInterfaceEvent)
				if !ok {
					continue
				}
				event := &vppcalls.InterfaceEvent{
					SwIfIndex:  ifEvent.SwIfIndex,
					AdminState: ifEvent.AdminUpDown,
					LinkState:  ifEvent.LinkUpDown,
					Deleted:    ifEvent.Deleted != 0,
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
	wantIfEventsReply := &binapi_interfaces.WantInterfaceEventsReply{}
	err = h.callsChannel.SendRequest(&binapi_interfaces.WantInterfaceEvents{
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
	vppNotifSubs, err := h.callsChannel.SubscribeNotification(notifChan, &dhcp.DHCPComplEvent{})
	if err != nil {
		return errors.Errorf("failed to subscribe VPP notification (sw_interface_event): %v", err)
	}
	_ = vppNotifSubs

	go func() {
		for {
			select {
			case e := <-notifChan:
				dhcpEvent, ok := e.(*dhcp.DHCPComplEvent)
				if !ok {
					continue
				}
				lease := dhcpEvent.Lease
				var hostAddr, routerAddr string
				if uintToBool(lease.IsIPv6) {
					hostAddr = fmt.Sprintf("%s/%d", net.IP(lease.HostAddress).To16().String(), uint32(lease.MaskWidth))
					routerAddr = fmt.Sprintf("%s/%d", net.IP(lease.RouterAddress).To16().String(), uint32(lease.MaskWidth))
				} else {
					hostAddr = fmt.Sprintf("%s/%d", net.IP(lease.HostAddress[:4]).To4().String(), uint32(lease.MaskWidth))
					routerAddr = fmt.Sprintf("%s/%d", net.IP(lease.RouterAddress[:4]).To4().String(), uint32(lease.MaskWidth))
				}
				leasesCh <- &vppcalls.Lease{
					SwIfIndex:     lease.SwIfIndex,
					State:         lease.State,
					Hostname:      string(bytes.SplitN(lease.Hostname, []byte{0x00}, 2)[0]),
					IsIPv6:        uintToBool(lease.IsIPv6),
					HostAddress:   hostAddr,
					RouterAddress: routerAddr,
					HostMac:       net.HardwareAddr(lease.HostMac).String(),
				}
			}
		}
	}()

	return nil
}
