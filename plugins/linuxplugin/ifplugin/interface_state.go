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

package ifplugin

import (
	"context"
	"sync"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	"github.com/vishvananda/netlink"
)

// LinuxInterfaceStateNotification aggregates status UP/DOWN/DELETED/UNKNOWN with
// the details (state) about the interfaces including counters.
type LinuxInterfaceStateNotification struct {
	// State of the network interface
	interfaceType string
	attributes    *netlink.LinkAttrs
}

// LinuxInterfaceStateUpdater processes all linux interface state data
type LinuxInterfaceStateUpdater struct {
	Log logging.Logger

	// Linux interface indices
	ifIndexes ifaceidx.LinuxIfIndexRW

	cfgLock sync.Mutex

	// Go routine management
	wg sync.WaitGroup // Wait group allows to wait until all goroutines of the plugin have finished.

	// Linux interface state
	stateWatcherRunning bool
	ifStateChan         chan *LinuxInterfaceStateNotification
	ifWatcherNotifCh    chan netlink.LinkUpdate
	ifWatcherDoneCh     chan struct{}
}

func (plugin *LinuxInterfaceStateUpdater) Init(ctx context.Context, ifIndexes ifaceidx.LinuxIfIndexRW,
	stateChan chan *LinuxInterfaceStateNotification, notifChan chan netlink.LinkUpdate, notifDone chan struct{}) error {
	plugin.Log.Debug("Initializing Linux Interface State Updater")
	// IfIndices
	plugin.ifIndexes = ifIndexes

	// Channels
	plugin.ifStateChan = stateChan
	plugin.ifWatcherNotifCh = notifChan
	plugin.ifWatcherDoneCh = notifDone

	// Start watch on linux interfaces
	go plugin.watchLinuxInterfaces(ctx)

	return plugin.subscribeInterfaceState()
}

func (plugin *LinuxInterfaceStateUpdater) Close() error {
	return nil
}

func (plugin *LinuxInterfaceStateUpdater) subscribeInterfaceState() error {
	if !plugin.stateWatcherRunning {
		plugin.stateWatcherRunning = true
		err := netlink.LinkSubscribe(plugin.ifWatcherNotifCh, plugin.ifWatcherDoneCh)
		if err != nil {
			return err
		}
	}
	return nil
}

func (plugin *LinuxInterfaceStateUpdater) watchLinuxInterfaces(ctx context.Context) {
	plugin.Log.Warnf("Watching on linux link notifications")

	plugin.wg.Add(1)
	defer plugin.wg.Done()

	for {
		select {
		case linkNotif := <-plugin.ifWatcherNotifCh:
			plugin.Log.Warnf("Notification received: %v", linkNotif)
			plugin.processLinkNotification(linkNotif)

		case <-ctx.Done():
			close(plugin.ifWatcherDoneCh)
			return
		}
	}
}

func (plugin *LinuxInterfaceStateUpdater) processLinkNotification(link netlink.Link) {
	linkAttrs := link.Attrs()

	if linkAttrs == nil {
		return
	}

	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	plugin.Log.Warnf("Processing Linux link update: Name=%v Type=%v OperState=%v Index=%v HwAddr=%v",
		linkAttrs.Name, link.Type(), linkAttrs.OperState, linkAttrs.Index, linkAttrs.HardwareAddr)

	// Prepare linux link notification
	linkNotif := &LinuxInterfaceStateNotification{
		interfaceType: link.Type(),
		attributes:    link.Attrs(),
	}

	select {
	case plugin.ifStateChan <- linkNotif:
		// Notification sent
	default:
		plugin.Log.Warn("Unable to send to the linux if state notification channel - buffer is full.")
	}
}
