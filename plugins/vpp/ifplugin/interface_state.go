// Copyright (c) 2017 Cisco and/or its affiliates.
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
	"time"

	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/go-errors/errors"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

// PeriodicPollingPeriod between statistics reads
var PeriodicPollingPeriod = 1 * time.Second

const (
	megabit = 1000000 // One megabit in bytes
)

// InterfaceStateUpdater holds state data of all VPP interfaces.
type InterfaceStateUpdater struct {
	log logging.Logger

	goVppMux govppmux.StatsAPI

	swIfIndexes    ifaceidx.SwIfIndex
	publishIfState func(notification *intf.InterfaceNotification)

	ifHandler vppcalls.IfVppAPI
	ifEvents  chan *vppcalls.InterfaceEvent

	ifState map[uint32]*intf.InterfacesState_Interface // swIfIndex to state data map
	access  sync.Mutex                                 // lock for the state data map

	vppCh     govppapi.Channel
	notifChan chan govppapi.Message
	swIdxChan chan ifaceidx.SwIfIdxDto

	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Init members (channels, maps...) and start go routines
func (c *InterfaceStateUpdater) Init(ctx context.Context, logger logging.PluginLogger, goVppMux govppmux.StatsAPI,
	swIfIndexes ifaceidx.SwIfIndex, notifChan chan govppapi.Message,
	publishIfState func(notification *intf.InterfaceNotification)) (err error) {
	// Logger
	c.log = logger.NewLogger("if-state")

	// Mappings & handlers
	c.swIfIndexes = swIfIndexes

	c.publishIfState = publishIfState
	c.ifState = make(map[uint32]*intf.InterfacesState_Interface)

	// VPP channel
	c.goVppMux = goVppMux
	c.vppCh, err = c.goVppMux.NewAPIChannel()
	if err != nil {
		return errors.Errorf("failed to create API channel: %v", err)
	}

	// VPP API handler
	c.ifHandler = vppcalls.NewIfVppHandler(c.vppCh, c.log)

	c.swIdxChan = make(chan ifaceidx.SwIfIdxDto, 100)
	swIfIndexes.WatchNameToIdx("ifplugin_ifstate", c.swIdxChan)
	c.notifChan = notifChan

	// Create child context
	var childCtx context.Context
	childCtx, c.cancel = context.WithCancel(ctx)

	// Watch for incoming notifications
	go c.watchVPPNotifications(childCtx)

	c.log.Info("Interface state updater initialized")

	return nil
}

// AfterInit subscribes for watching VPP notifications on previously initialized channel
func (c *InterfaceStateUpdater) AfterInit() error {
	err := c.subscribeVPPNotifications()
	if err != nil {
		return err
	}
	return nil
}

// subscribeVPPNotifications subscribes for interface state notifications from VPP.
func (c *InterfaceStateUpdater) subscribeVPPNotifications() error {
	if err := c.ifHandler.WatchInterfaceEvents(c.ifEvents); err != nil {
		return err
	}

	return nil
}

// Close unsubscribes from interface state notifications from VPP & GOVPP channel
func (c *InterfaceStateUpdater) Close() error {
	c.cancel()
	c.wg.Wait()

	if err := safeclose.Close(c.vppCh); err != nil {
		return c.LogError(errors.Errorf("failed to safe close interface state: %v", err))
	}

	return nil
}

// watchVPPNotifications watches for delivery of notifications from VPP.
func (c *InterfaceStateUpdater) watchVPPNotifications(ctx context.Context) {
	c.wg.Add(1)
	defer c.wg.Done()

	if c.notifChan != nil {
		c.log.Debug("Interface state VPP notification watcher started")
	} else {
		c.log.Warn("Interface state VPP notification does not start: the channel c nil")
		return
	}

	// Periodically read VPP counters and combined counters for VPP statistics
	go c.startReadingCounters(ctx)

	for {
		select {
		case msg := <-c.ifEvents:
			c.processIfStateEvent(msg)
		case swIdxDto := <-c.swIdxChan:
			if swIdxDto.Del {
				c.setIfStateDeleted(swIdxDto.Idx, swIdxDto.Name)
			} else if !swIdxDto.Update {
				ifaces, err := c.ifHandler.DumpInterfaces()
				if err != nil {
					c.log.Warnf("dump interfaces failed: %v", err)
					continue
				}
				for _, ifaceDetails := range ifaces {
					if ifaceDetails.Meta.SwIfIndex != swIdxDto.Idx {
						// not the added interface
						continue
					}
					c.updateIfStateDetails(ifaceDetails)
				}
			}

		case <-ctx.Done():
			// stop watching for notifications
			c.log.Debug("Interface state VPP notification watcher stopped")
			return
		}
	}
}

// startReadingCounters periodically reads statistics for all interfaces
func (c *InterfaceStateUpdater) startReadingCounters(ctx context.Context) {
	for {
		select {
		case <-time.After(PeriodicPollingPeriod):
			c.doInterfaceStatsRead()
		case <-ctx.Done():
			c.log.Debug("Interface state VPP periodic polling stopped")
			return
		}
	}
}

// doInterfaceStatsRead dumps statistics using interface filter and processes them
func (c *InterfaceStateUpdater) doInterfaceStatsRead() {
	c.access.Lock()
	defer c.access.Unlock()

	ifStatsList, err := c.goVppMux.GetInterfaceStats()
	if err != nil {
		// TODO add some counter to prevent it log forever
		c.log.Errorf("failed to read statistics data: %v", err)
	}
	if ifStatsList == nil || len(ifStatsList.Interfaces) == 0 {
		return
	}
	for _, ifStats := range ifStatsList.Interfaces {
		c.processInterfaceStatEntry(ifStats)
	}
}

// processInterfaceStatEntry fills state data for every registered interface and publishes them
func (c *InterfaceStateUpdater) processInterfaceStatEntry(ifCounters govppapi.InterfaceCounters) {
	ifState, found := c.getIfStateDataWLookup(ifCounters.InterfaceIndex)
	if !found {
		return
	}
	ifState.Statistics = &intf.InterfacesState_Interface_Statistics{
		DropPackets:     ifCounters.Drops,
		PuntPackets:     ifCounters.Punts,
		Ipv4Packets:     ifCounters.IP4,
		Ipv6Packets:     ifCounters.IP6,
		InNobufPackets:  ifCounters.RxNoBuf,
		InMissPackets:   ifCounters.RxMiss,
		InErrorPackets:  ifCounters.RxErrors,
		OutErrorPackets: ifCounters.TxErrors,
		InPackets:       ifCounters.RxPackets,
		InBytes:         ifCounters.RxBytes,
		OutPackets:      ifCounters.TxPackets,
		OutBytes:        ifCounters.TxBytes,
	}

	c.publishIfState(&intf.InterfaceNotification{
		Type: intf.InterfaceNotification_COUNTERS, State: ifState})
}

// processIfStateEvent process a VPP state event notification.
func (c *InterfaceStateUpdater) processIfStateEvent(notif *vppcalls.InterfaceEvent) {
	c.access.Lock()
	defer c.access.Unlock()

	// update and return if state data
	ifState, found := c.updateIfStateFlags(notif)
	if !found {
		return
	}
	c.log.Debugf("Interface state notification for %s (idx: %d): %+v",
		ifState.Name, ifState.IfIndex, notif)

	// store data in ETCD
	c.publishIfState(&intf.InterfaceNotification{
		Type: intf.InterfaceNotification_UPDOWN, State: ifState})
}

// getIfStateData returns interface state data structure for the specified interface index and interface name.
// NOTE: plugin.ifStateData needs to be locked when calling this function!
func (c *InterfaceStateUpdater) getIfStateData(swIfIndex uint32, ifName string) (*intf.InterfacesState_Interface, bool) {

	ifState, ok := c.ifState[swIfIndex]

	// check also if the provided logical name c the same as the one associated
	// with swIfIndex, because swIfIndexes might be reused
	if ok && ifState.Name == ifName {
		return ifState, true
	}

	return nil, false
}

// getIfStateDataWLookup returns interface state data structure for the specified interface index (creates it if it does not exist).
// NOTE: plugin.ifStateData needs to be locked when calling this function!
func (c *InterfaceStateUpdater) getIfStateDataWLookup(ifIdx uint32) (
	*intf.InterfacesState_Interface, bool) {
	ifName, _, found := c.swIfIndexes.LookupName(ifIdx)
	if !found {
		c.log.Debugf("Interface state data structure lookup for %d interrupted, not registered yet", ifIdx)
		return nil, found
	}
	ifState, found := c.getIfStateData(ifIdx, ifName)
	if !found {
		ifState = &intf.InterfacesState_Interface{
			IfIndex:    ifIdx,
			Name:       ifName,
			Statistics: &intf.InterfacesState_Interface_Statistics{},
		}

		c.ifState[ifIdx] = ifState
		found = true
	}

	return ifState, found
}

// updateIfStateFlags updates the interface state data in memory from provided VPP flags message and returns updated state data.
// NOTE: plugin.ifStateData needs to be locked when calling this function!
func (c *InterfaceStateUpdater) updateIfStateFlags(vppMsg *vppcalls.InterfaceEvent) (
	iface *intf.InterfacesState_Interface, found bool) {

	ifState, found := c.getIfStateDataWLookup(vppMsg.SwIfIndex)
	if !found {
		return nil, false
	}
	ifState.LastChange = time.Now().Unix()

	if vppMsg.Deleted {
		ifState.AdminStatus = intf.InterfacesState_Interface_DELETED
		ifState.OperStatus = intf.InterfacesState_Interface_DELETED
	} else {
		if vppMsg.AdminState == 1 {
			ifState.AdminStatus = intf.InterfacesState_Interface_UP
		} else {
			ifState.AdminStatus = intf.InterfacesState_Interface_DOWN
		}
		if vppMsg.LinkState == 1 {
			ifState.OperStatus = intf.InterfacesState_Interface_UP
		} else {
			ifState.OperStatus = intf.InterfacesState_Interface_DOWN
		}
	}
	return ifState, true
}

// updateIfStateDetails updates the interface state data in memory from provided VPP details message.
func (c *InterfaceStateUpdater) updateIfStateDetails(ifDetails *vppcalls.InterfaceDetails) {
	c.access.Lock()
	defer c.access.Unlock()

	ifState, found := c.getIfStateDataWLookup(ifDetails.Meta.SwIfIndex)
	if !found {
		return
	}

	ifState.InternalName = ifDetails.Meta.InternalName

	if ifDetails.Meta.AdminState == 1 {
		ifState.AdminStatus = intf.InterfacesState_Interface_UP
	} else if ifDetails.Meta.AdminState == 0 {
		ifState.AdminStatus = intf.InterfacesState_Interface_DOWN
	} else {
		ifState.AdminStatus = intf.InterfacesState_Interface_UNKNOWN_STATUS
	}

	if ifDetails.Meta.LinkState == 1 {
		ifState.OperStatus = intf.InterfacesState_Interface_UP
	} else if ifDetails.Meta.LinkState == 0 {
		ifState.OperStatus = intf.InterfacesState_Interface_DOWN
	} else {
		ifState.OperStatus = intf.InterfacesState_Interface_UNKNOWN_STATUS
	}

	ifState.PhysAddress = ifDetails.Interface.PhysAddress

	ifState.Mtu = uint32(ifDetails.Meta.LinkMTU)

	switch ifDetails.Meta.LinkSpeed {
	case 1:
		ifState.Speed = 10 * megabit // 10M
	case 2:
		ifState.Speed = 100 * megabit // 100M
	case 4:
		ifState.Speed = 1000 * megabit // 1G
	case 8:
		ifState.Speed = 10000 * megabit // 10G
	case 16:
		ifState.Speed = 40000 * megabit // 40G
	case 32:
		ifState.Speed = 100000 * megabit // 100G
	default:
		ifState.Speed = 0
	}

	switch ifDetails.Meta.LinkSpeed {
	case 1:
		ifState.Duplex = intf.InterfacesState_Interface_HALF
	case 2:
		ifState.Duplex = intf.InterfacesState_Interface_FULL
	default:
		ifState.Duplex = intf.InterfacesState_Interface_UNKNOWN_DUPLEX
	}

	c.publishIfState(&intf.InterfaceNotification{
		Type: intf.InterfaceNotification_UNKNOWN, State: ifState})
}

// setIfStateDeleted marks the interface as deleted in the state data structure in memory.
func (c *InterfaceStateUpdater) setIfStateDeleted(swIfIndex uint32, ifName string) {
	c.access.Lock()
	defer c.access.Unlock()

	ifState, found := c.getIfStateData(swIfIndex, ifName)
	if !found {
		return
	}
	ifState.AdminStatus = intf.InterfacesState_Interface_DELETED
	ifState.OperStatus = intf.InterfacesState_Interface_DELETED
	ifState.LastChange = time.Now().Unix()

	// this can be post-processed by multiple plugins
	c.publishIfState(&intf.InterfaceNotification{
		Type: intf.InterfaceNotification_UNKNOWN, State: ifState})
}

// LogError prints error if not nil, including stack trace. The same value is also returned, so it can be easily propagated further
func (c *InterfaceStateUpdater) LogError(err error) error {
	if err == nil {
		return nil
	}
	switch err.(type) {
	case *errors.Error:
		c.log.WithField("logger", c.log).Errorf(string(err.Error() + "\n" + string(err.(*errors.Error).Stack())))
	default:
		c.log.Error(err)
	}
	return err
}
