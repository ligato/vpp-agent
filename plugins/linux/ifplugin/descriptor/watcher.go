// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package descriptor

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	prototypes "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	ifmodel "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
)

const (
	// InterfaceWatcherName is the name of the descriptor watching Linux interfaces
	// in the default namespace.
	InterfaceWatcherName = "linux-interface-watcher"
)

// InterfaceWatcher watches default namespace for newly added/removed Linux interfaces.
type InterfaceWatcher struct {
	// input arguments
	log         logging.Logger
	kvscheduler kvs.KVScheduler
	ifHandler   linuxcalls.NetlinkAPIRead

	// go routine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// a set of interfaces present in the default namespace
	ifacesMu sync.Mutex
	ifaces   map[string]hostInterface

	// conditional variable to check if the list of interfaces is in-sync with
	// Linux network stack
	intfsInSync     bool
	intfsInSyncCond *sync.Cond

	// Linux notifications
	linkNotifCh chan netlink.LinkUpdate
	addrNotifCh chan netlink.AddrUpdate
	doneCh      chan struct{}
	notify      func(notification *ifmodel.InterfaceNotification)
}

type hostInterface struct {
	name    string
	enabled bool
	ipAddrs []string
}

// NewInterfaceWatcher creates a new instance of the Interface Watcher.
func NewInterfaceWatcher(kvscheduler kvs.KVScheduler, ifHandler linuxcalls.NetlinkAPI, notifyInterface func(*ifmodel.InterfaceNotification), log logging.PluginLogger) *InterfaceWatcher {
	descriptor := &InterfaceWatcher{
		log:         log.NewLogger("if-watcher"),
		kvscheduler: kvscheduler,
		ifHandler:   ifHandler,
		notify:      notifyInterface,
		ifaces:      make(map[string]hostInterface),
		linkNotifCh: make(chan netlink.LinkUpdate),
		addrNotifCh: make(chan netlink.AddrUpdate),
		doneCh:      make(chan struct{}),
	}
	descriptor.intfsInSyncCond = sync.NewCond(&descriptor.ifacesMu)
	descriptor.ctx, descriptor.cancel = context.WithCancel(context.Background())

	return descriptor
}

// GetDescriptor returns descriptor suitable for registration with the KVScheduler.
func (w *InterfaceWatcher) GetDescriptor() *kvs.KVDescriptor {
	return &kvs.KVDescriptor{
		Name:        InterfaceWatcherName,
		KeySelector: w.IsLinuxInterfaceNotification,
		Retrieve:    w.Retrieve,
	}
}

// IsLinuxInterfaceNotification returns <true> for keys representing
// notifications about Linux interfaces in the default network namespace.
func (w *InterfaceWatcher) IsLinuxInterfaceNotification(key string) bool {
	return strings.HasPrefix(key, ifmodel.InterfaceHostNameKeyPrefix)
}

// Retrieve returns key with empty value for every currently existing Linux interface
// in the default network namespace.
func (w *InterfaceWatcher) Retrieve(correlate []kvs.KVWithMetadata) (values []kvs.KVWithMetadata, err error) {
	// wait until the set of interfaces is in-sync with the Linux network stack
	w.ifacesMu.Lock()
	if !w.intfsInSync {
		w.intfsInSyncCond.Wait()
	}
	defer w.ifacesMu.Unlock()

	for _, hostIface := range w.ifaces {
		if !hostIface.enabled {
			continue
		}
		values = append(values, kvs.KVWithMetadata{
			Key:    ifmodel.InterfaceHostNameKey(hostIface.name),
			Value:  &prototypes.Empty{},
			Origin: kvs.FromSB,
		})
		for _, ipAddr := range hostIface.ipAddrs {
			values = append(values, kvs.KVWithMetadata{
				Key:    ifmodel.InterfaceHostNameWithAddrKey(hostIface.name, ipAddr),
				Value:  &prototypes.Empty{},
				Origin: kvs.FromSB,
			})
		}
	}

	return values, nil
}

// StartWatching starts interface watching.
func (w *InterfaceWatcher) StartWatching() error {
	// watch default namespace to be aware of interfaces not created by this plugin
	err := w.ifHandler.LinkSubscribe(w.linkNotifCh, w.doneCh)
	if err != nil {
		err = errors.Errorf("failed to subscribe for link notifications: %v", err)
		w.log.Error(err)
		return err
	}
	err = w.ifHandler.AddrSubscribe(w.addrNotifCh, w.doneCh)
	if err != nil {
		err = errors.Errorf("failed to subscribe for address notifications: %v", err)
		w.log.Error(err)
		return err
	}
	w.wg.Add(1)
	go w.watchDefaultNamespace()
	return nil
}

// StopWatching stops interface watching.
func (w *InterfaceWatcher) StopWatching() {
	w.cancel()
	w.wg.Wait()
}

// watchDefaultNamespace watches for notification about added/removed interfaces/addresses
// to/from the default namespace.
func (w *InterfaceWatcher) watchDefaultNamespace() {
	defer w.wg.Done()

	// get the set of interfaces already available in the default namespace
	links, err := w.ifHandler.GetLinkList()
	if err == nil {
		for _, link := range links {
			iface := hostInterface{name: link.Attrs().Name}
			enabled, err := w.ifHandler.IsInterfaceUp(iface.name)
			if err != nil {
				w.log.Warnf("IsInterfaceUp failed for interface %s: %v",
					iface.name, err)
				continue
			}
			iface.enabled = enabled
			addrs, err := w.ifHandler.GetAddressList(iface.name)
			if err != nil {
				w.log.Warnf("GetAddressList failed for interface %s: %v",
					iface.name, err)
				continue
			}
			for _, addr := range addrs {
				iface.ipAddrs = append(iface.ipAddrs, addr.IPNet.String())
			}
			w.ifaces[iface.name] = iface
		}
	} else {
		w.log.Warnf("failed to list interfaces in the default namespace: %v", err)
	}

	// mark the state in-sync with the Linux network stack
	w.ifacesMu.Lock()
	w.intfsInSync = true
	w.ifacesMu.Unlock()
	w.intfsInSyncCond.Broadcast()

	for {
		select {
		case linkNotif := <-w.linkNotifCh:
			w.processLinkNotification(linkNotif)
		case addrNotif := <-w.addrNotifCh:
			w.processAddrNotification(addrNotif)
		case <-w.ctx.Done():
			close(w.doneCh)
			return
		}
	}
}

// processLinkNotification processes link notification received from Linux.
func (w *InterfaceWatcher) processLinkNotification(linkUpdate netlink.LinkUpdate) {
	w.ifacesMu.Lock()
	defer w.ifacesMu.Unlock()

	// send notification to any interface state watcher (e.g. Configurator)
	w.sendStateNotification(linkUpdate)

	// push update to the KV Scheduler
	ifName := linkUpdate.Attrs().Name
	prevState := w.ifaces[ifName]
	newState := prevState
	newState.name = ifName
	newState.enabled = isLinkUp(linkUpdate)
	if prevState.enabled != newState.enabled {
		w.updateLinkKV(ifName, newState.enabled)
		// do not advertise IPs if interface is disabled
		for _, ipAddr := range newState.ipAddrs {
			w.updateAddrKV(ifName, ipAddr, !newState.enabled)
		}
	}
	w.ifaces[ifName] = newState
}

// processAddrNotification processes address notification received from Linux.
func (w *InterfaceWatcher) processAddrNotification(addrUpdate netlink.AddrUpdate) {
	w.ifacesMu.Lock()
	defer w.ifacesMu.Unlock()

	link, err := w.ifHandler.GetLinkByIndex(addrUpdate.LinkIndex)
	if err != nil {
		w.log.Warn(err)
		return
	}

	// push update to the KV Scheduler
	ifName := link.Attrs().Name
	addr := addrUpdate.LinkAddress.String()
	removed := !addrUpdate.NewAddr
	prevState := w.ifaces[ifName]
	addrIdx := -1
	for i, ipAddr := range prevState.ipAddrs {
		if ipAddr == addr {
			addrIdx = i
			break
		}
	}
	knownAddr := addrIdx != -1
	if knownAddr != removed {
		// removed unknown IP or added already known IP
		return
	}
	if prevState.enabled {
		w.updateAddrKV(ifName, addr, removed)
	}

	// update the internal cache
	newState := prevState
	newState.name = ifName
	if removed {
		lastIdx := len(newState.ipAddrs)-1
		newState.ipAddrs[addrIdx] = newState.ipAddrs[lastIdx]
		newState.ipAddrs[lastIdx] = ""
		newState.ipAddrs = newState.ipAddrs[:lastIdx]
	} else {
		newState.ipAddrs = append(newState.ipAddrs, addr)
	}
	w.ifaces[ifName] = newState
}

func linkToInterfaceType(link netlink.Link) ifmodel.Interface_Type {
	switch link.Type() {
	case "veth":
		return ifmodel.Interface_VETH
	case "tuntap", "tun":
		return ifmodel.Interface_TAP_TO_VPP
	case "vrf":
		return ifmodel.Interface_VRF_DEVICE
	case "dummy":
		return ifmodel.Interface_DUMMY
	default:
		if link.Attrs().Name == linuxcalls.DefaultLoopbackName {
			return ifmodel.Interface_LOOPBACK
		}
		return ifmodel.Interface_UNDEFINED
	}
}

// updateLinkKV updates key-value pair representing the interface latest link status.
func (w *InterfaceWatcher) updateLinkKV(ifName string, enabled bool) {
	var value proto.Message
	if enabled {
		// empty == enabled, nil == disabled
		value = &prototypes.Empty{}
	}
	if err := w.kvscheduler.PushSBNotification(kvs.KVWithMetadata{
		Key:      ifmodel.InterfaceHostNameKey(ifName),
		Value:    value,
		Metadata: nil,
	}); err != nil {
		w.log.Warnf("pushing SB notification failed: %v", err)
	}
}

// updateAddrKV updates key-value pair representing IP address assigned to a host interface.
func (w *InterfaceWatcher) updateAddrKV(ifName string, address string, removed bool) {
	var value proto.Message
	if !removed {
		// empty == assigned, nil == not assigned
		value = &prototypes.Empty{}
	}
	if err := w.kvscheduler.PushSBNotification(kvs.KVWithMetadata{
		Key:      ifmodel.InterfaceHostNameWithAddrKey(ifName, address),
		Value:    value,
		Metadata: nil,
	}); err != nil {
		w.log.Warnf("pushing SB notification failed: %v", err)
	}
}

func (w *InterfaceWatcher) sendStateNotification(linkUpdate netlink.LinkUpdate) {
	if w.notify != nil {
		attrs := linkUpdate.Attrs()
		adminStatus := ifmodel.InterfaceState_DOWN
		if attrs.Flags&net.FlagUp == net.FlagUp {
			adminStatus = ifmodel.InterfaceState_UP
		}
		operStatus := ifmodel.InterfaceState_DOWN
		if attrs.OperState != netlink.OperDown && attrs.OperState != netlink.OperNotPresent {
			operStatus = ifmodel.InterfaceState_UP
		}
		w.notify(&ifmodel.InterfaceNotification{
			Type: ifmodel.InterfaceNotification_UPDOWN,
			State: &ifmodel.InterfaceState{
				Name:         attrs.Alias,
				InternalName: attrs.Name,
				Type:         linkToInterfaceType(linkUpdate.Link),
				IfIndex:      int32(attrs.Index),
				AdminStatus:  adminStatus,
				OperStatus:   operStatus,
				LastChange:   time.Now().Unix(),
				PhysAddress:  attrs.HardwareAddr.String(),
				Speed:        0,
				Mtu:          uint32(attrs.MTU),
				Statistics:   nil,
			},
		})
	}
}

func isLinkUp(update netlink.LinkUpdate) bool {
	return update.Attrs().OperState != netlink.OperDown &&
		update.Attrs().OperState != netlink.OperNotPresent
}
