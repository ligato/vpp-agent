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
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"
	"github.com/vishvananda/netlink"
)

// LinuxInterfaceStateNotification aggregates operational status derived from netlink with
// the details (state) about the interface.
type LinuxInterfaceStateNotification struct {
	// State of the network interface
	interfaceType  string
	interfaceState netlink.LinkOperState
	attributes     *netlink.LinkAttrs
}

// LinuxInterfaceStateUpdater processes all linux interface state data
type LinuxInterfaceStateUpdater struct {
	log     logging.Logger
	cfgLock sync.Mutex

	// Go routine management
	wg sync.WaitGroup

	// Linux interface state
	stateWatcherRunning bool
	ifStateChan         chan *LinuxInterfaceStateNotification
	ifWatcherNotifCh    chan netlink.LinkUpdate
	ifWatcherDoneCh     chan struct{}
}

// Init channels for interface state watcher, start it in separate go routine and subscribe to default namespace
func (plugin *LinuxInterfaceStateUpdater) Init(ctx context.Context, logger logging.PluginLogger, ifIndexes ifaceidx.LinuxIfIndexRW,
	stateChan chan *LinuxInterfaceStateNotification) error {
	// Logger
	plugin.log = logger.NewLogger("-if-state")
	plugin.log.Debug("Initializing Linux Interface State Updater")

	// Channels
	plugin.ifStateChan = stateChan
	plugin.ifWatcherNotifCh = make(chan netlink.LinkUpdate, 10)
	plugin.ifWatcherDoneCh = make(chan struct{})

	// Start watch on linux interfaces
	go plugin.watchLinuxInterfaces(ctx)

	// Subscribe to default linux namespace
	return plugin.subscribeInterfaceState()
}

// Close watcher channel (state chan is closed in LinuxInterfaceConfigurator)
func (plugin *LinuxInterfaceStateUpdater) Close() error {
	return safeclose.Close(plugin.ifWatcherNotifCh, plugin.ifWatcherDoneCh)
}

// Subscribe to linux default namespace
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

// Watch linux interfaces and send events to processing
func (plugin *LinuxInterfaceStateUpdater) watchLinuxInterfaces(ctx context.Context) {
	plugin.log.Debugf("Watching on linux link notifications")

	plugin.wg.Add(1)
	defer plugin.wg.Done()

	for {
		select {
		case linkNotif := <-plugin.ifWatcherNotifCh:
			plugin.processLinkNotification(linkNotif)

		case <-ctx.Done():
			close(plugin.ifWatcherDoneCh)
			return
		}
	}
}

// Prepare notification and send it to the state channel
func (plugin *LinuxInterfaceStateUpdater) processLinkNotification(link netlink.Link) {
	linkAttrs := link.Attrs()

	if linkAttrs == nil {
		return
	}

	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	plugin.log.Debugf("Processing Linux link update: Name=%v Type=%v OperState=%v Index=%v HwAddr=%v",
		linkAttrs.Name, link.Type(), linkAttrs.OperState, linkAttrs.Index, linkAttrs.HardwareAddr)

	// Prepare linux link notification
	linkNotif := &LinuxInterfaceStateNotification{
		interfaceType:  link.Type(),
		interfaceState: linkAttrs.OperState,
		attributes:     link.Attrs(),
	}

	select {
	case plugin.ifStateChan <- linkNotif:
		// Notification sent
	default:
		plugin.log.Warn("Unable to send to the linux if state notification channel - buffer is full.")
	}
}
