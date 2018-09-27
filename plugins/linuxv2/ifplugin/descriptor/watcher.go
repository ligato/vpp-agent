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
	"strings"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/vishvananda/netlink"
	"github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"

	scheduler "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/linuxcalls"
	ifmodel "github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
)

const (
	// InterfaceWatcherName is the name of the descriptor watching Linux interfaces
	// in the default namespace.
	InterfaceWatcherName = "linux-interface-watcher"
)

// InterfaceWatcher watches default namespace for newly added/removed Linux interfaces.
type InterfaceWatcher struct {
	// input arguments
	log       logging.Logger
	scheduler scheduler.KVScheduler
	ifHandler linuxcalls.NetlinkAPI

	// go routine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// a set of interfaces present in the default namespace
	intfsLock sync.Mutex
	intfs     map[string]struct{}

	// interface changes delayed to give Linux time to "finalize" them
	pendingIntfs map[string]bool // interface name -> exists?

	// conditional variable to check if the list of interfaces is in-sync with
	// Linux network stack
	intfsInSync     bool
	intfsInSyncCond *sync.Cond

	// Linux notifications
	notifCh chan netlink.LinkUpdate
	doneCh  chan struct{}
}

// NewInterfaceWatcher creates a new instance of the Interface Watcher.
func NewInterfaceWatcher(scheduler scheduler.KVScheduler, ifHandler linuxcalls.NetlinkAPI, log logging.PluginLogger) *InterfaceWatcher {
	descriptor := &InterfaceWatcher{
		log:          log.NewLogger("-watcher"),
		scheduler:    scheduler,
		ifHandler:    ifHandler,
		intfs:        make(map[string]struct{}),
		pendingIntfs: make(map[string]bool),
		notifCh:      make(chan netlink.LinkUpdate),
		doneCh:       make(chan struct{}),
	}
	descriptor.intfsInSyncCond = sync.NewCond(&descriptor.intfsLock)
	descriptor.ctx, descriptor.cancel = context.WithCancel(context.Background())

	return descriptor
}

// GetDescriptor returns descriptor suitable for registration with the KVScheduler.
func (intfw *InterfaceWatcher) GetDescriptor() *scheduler.KVDescriptor {
	return &scheduler.KVDescriptor{
		Name:        InterfaceWatcherName,
		KeySelector: intfw.IsLinuxInterfaceNotification,
		Dump:        intfw.Dump,
	}
}

// IsLinuxInterfaceNotification returns <true> for keys representing
// notifications about Linux interfaces in the default network namespace.
func (intfw *InterfaceWatcher) IsLinuxInterfaceNotification(key string) bool {
	return strings.HasPrefix(key, ifmodel.InterfaceHostNameKeyPrefix)
}

// Dump returns key with empty value for every currently existing Linux interface
// in the default network namespace.
func (intfw *InterfaceWatcher) Dump(correlate []scheduler.KVWithMetadata) (dump []scheduler.KVWithMetadata, err error) {
	// wait until the set of interfaces is in-sync with the Linux network stack
	intfw.intfsLock.Lock()
	if !intfw.intfsInSync {
		intfw.intfsInSyncCond.Wait()
	}
	defer intfw.intfsLock.Unlock()

	for ifName := range intfw.intfs {
		dump = append(dump, scheduler.KVWithMetadata{
			Key:    ifmodel.InterfaceHostNameKey(ifName),
			Value:  &prototypes.Empty{},
			Origin: scheduler.FromSB,
		})
	}
	intfw.log.WithField("dump", dump).Debug("Dumping Linux interface host names in default namespace")
	return dump, nil
}

// StartWatching starts interface watching.
func (intfw *InterfaceWatcher) StartWatching() error {
	// watch default namespace to be aware of interfaces not created by this plugin
	err := intfw.ifHandler.LinkSubscribe(intfw.notifCh, intfw.doneCh)
	if err != nil {
		err = errors.Errorf("failed to subscribe link: %v", err)
		intfw.log.Error(err)
		return err
	}
	go intfw.watchDefaultNamespace()
	return nil
}

// StopWatching stops interface watching.
func (intfw *InterfaceWatcher) StopWatching() {
	intfw.cancel()
	intfw.wg.Wait()
}

// watchDefaultNamespace watches for notification about added/removed interfaces
// to/from the default namespace.
func (intfw *InterfaceWatcher) watchDefaultNamespace() {
	intfw.wg.Add(1)
	defer intfw.wg.Done()

	// get the set of interfaces already available in the default namespace
	links, err := intfw.ifHandler.GetLinkList()
	if err == nil {
		for _, link := range links {
			if isInterfaceEnabled(link) {
				intfw.intfs[link.Attrs().Name] = struct{}{}
			}
		}
	} else {
		intfw.log.Warnf("failed to list interfaces in the default namespace: %v", err)
	}

	// mark the state in-sync with the Linux network stack
	intfw.intfsLock.Lock()
	intfw.intfsInSync = true
	intfw.intfsLock.Unlock()
	intfw.intfsInSyncCond.Broadcast()

	for {
		select {
		case linkNotif := <-intfw.notifCh:
			intfw.processLinkNotification(linkNotif)

		case <-intfw.ctx.Done():
			close(intfw.doneCh)
			return
		}
	}
}

// processLinkNotification processes link notification received from Linux.
func (intfw *InterfaceWatcher) processLinkNotification(linkUpdate netlink.LinkUpdate) {
	intfw.intfsLock.Lock()
	defer intfw.intfsLock.Unlock()

	ifName := linkUpdate.Attrs().Name
	isEnabled := isInterfaceEnabled(linkUpdate.Link)

	_, isPendingNotif := intfw.pendingIntfs[ifName]
	if isPendingNotif {
		// notification for this interface is already scheduled, just update the state
		intfw.pendingIntfs[ifName] = isEnabled
		return
	}

	if !intfw.needsUpdate(ifName, isEnabled) {
		// ignore notification if the interface admin status remained the same
		return
	}

	if isEnabled {
		// do not notify until interface is truly finished
		intfw.pendingIntfs[ifName] = true
		go intfw.delayNotification(ifName)
		return
	}

	// notification about removed interface is propagated immediately
	intfw.notifyScheduler(ifName, false)
}

// delayNotification delays notification about enabled interface - typically
// interface is created in multiple stages and we do not want to notify scheduler
// about intermediate states.
func (intfw *InterfaceWatcher) delayNotification(ifName string) {
	intfw.wg.Add(1)
	defer intfw.wg.Done()

	select {
	case <-intfw.ctx.Done():
		return
	case <-time.After(time.Second):
		intfw.applyDelayedNotification(ifName)
	}
}

// applyDelayedNotification applies delayed interface notification.
func (intfw *InterfaceWatcher) applyDelayedNotification(ifName string) {
	intfw.intfsLock.Lock()
	defer intfw.intfsLock.Unlock()

	// in the meantime the status may have changed and may not require update anymore
	isEnabled := intfw.pendingIntfs[ifName]
	if intfw.needsUpdate(ifName, isEnabled) {
		intfw.notifyScheduler(ifName, isEnabled)
	}

	delete(intfw.pendingIntfs, ifName)
}

// notifyScheduler notifies scheduler about interface change.
func (intfw *InterfaceWatcher) notifyScheduler(ifName string, enabled bool) {
	var value proto.Message

	if enabled {
		intfw.intfs[ifName] = struct{}{}
		value = &prototypes.Empty{}
	} else {
		delete(intfw.intfs, ifName)
	}

	intfw.scheduler.PushSBNotification(
		ifmodel.InterfaceHostNameKey(ifName),
		value,
		nil)
}

func (intfw *InterfaceWatcher) needsUpdate(ifName string, isEnabled bool) bool {
	_, wasEnabled := intfw.intfs[ifName]
	return isEnabled != wasEnabled
}
