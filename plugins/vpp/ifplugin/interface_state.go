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

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/pkg/errors"

	intf "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

// PeriodicPollingPeriod between statistics reads
// TODO  should be configurable
var PeriodicPollingPeriod = 1 * time.Second

// statType is the specific interface statistics type.
type statType string

// interface stats prefix as defined by the VPP
const ifPrefix = "/if/"

// interface statistics matching patterns
const (
	Drop    statType = ifPrefix + "drops"
	Punt             = ifPrefix + "punt"
	IPv4             = ifPrefix + "ip4"
	IPv6             = ifPrefix + "ip6"
	RxNoBuf          = ifPrefix + "rx-no-buf"
	RxMiss           = ifPrefix + "rx-miss"
	RxError          = ifPrefix + "rx-error"
	TxError          = ifPrefix + "tx-error"
	Rx               = ifPrefix + "rx"
	Tx               = ifPrefix + "tx"
)

const (
	megabit = 1000000 // One megabit in bytes
)

// InterfaceStateUpdater holds state data of all VPP interfaces.
type InterfaceStateUpdater struct {
	log logging.Logger

	kvScheduler    kvs.KVScheduler
	swIfIndexes    ifaceidx.IfaceMetadataIndex
	publishIfState func(notification *intf.InterfaceNotification)

	access  sync.Mutex                      // lock for the state data map
	ifState map[uint32]*intf.InterfaceState // swIfIndex to state data map

	goVppMux govppmux.StatsAPI

	vppCh      govppapi.Channel
	ifMetaChan chan ifaceidx.IfaceMetadataDto

	ifHandler vppcalls.InterfaceVppAPI
	ifEvents  chan *vppcalls.InterfaceEvent

	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Init members (channels, maps...) and start go routines
func (c *InterfaceStateUpdater) Init(ctx context.Context, logger logging.PluginLogger, kvScheduler kvs.KVScheduler,
	goVppMux govppmux.StatsAPI, swIfIndexes ifaceidx.IfaceMetadataIndex,
	publishIfState func(notification *intf.InterfaceNotification)) (err error) {

	// Logger
	c.log = logger.NewLogger("if-state")

	// Mappings
	c.swIfIndexes = swIfIndexes

	c.kvScheduler = kvScheduler
	c.publishIfState = publishIfState
	c.ifState = make(map[uint32]*intf.InterfaceState)

	// VPP channel
	c.goVppMux = goVppMux
	c.vppCh, err = c.goVppMux.NewAPIChannel()
	if err != nil {
		return errors.Errorf("failed to create API channel: %v", err)
	}

	c.ifHandler = vppcalls.CompatibleInterfaceVppHandler(c.vppCh, logger.NewLogger("if-handler"))

	c.ifMetaChan = make(chan ifaceidx.IfaceMetadataDto, 1000)
	swIfIndexes.WatchInterfaces("ifplugin_ifstate", c.ifMetaChan)

	c.ifEvents = make(chan *vppcalls.InterfaceEvent, 1000)

	// Create child context
	var childCtx context.Context
	childCtx, c.cancel = context.WithCancel(ctx)

	// Watch for incoming notifications
	c.wg.Add(1)
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

	// TODO: handle unsubscribing
	/*if c.vppNotifSubs != nil {
		if err := c.vppNotifSubs.Unsubscribe(); err != nil {
			return errors.Errorf("failed to unsubscribe interface state notification on close: %v", err)
		}
	}*/

	if err := safeclose.Close(c.vppCh); err != nil {
		return errors.Errorf("failed to safe close interface state: %v", err)
	}

	return nil
}

// watchVPPNotifications watches for delivery of notifications from VPP.
func (c *InterfaceStateUpdater) watchVPPNotifications(ctx context.Context) {
	defer c.wg.Done()

	// Periodically read VPP counters and combined counters for VPP statistics
	go c.startReadingCounters(ctx)

	for {
		select {
		case notif := <-c.ifEvents:
			// if the notification is a result of a configuration change,
			// make sure the associated transaction has already finalized
			c.kvScheduler.TransactionBarrier()

			c.processIfStateEvent(notif)

		case ifMetaDto := <-c.ifMetaChan:
			if ifMetaDto.Del {
				c.setIfStateDeleted(ifMetaDto.Metadata.SwIfIndex, ifMetaDto.Name)
			} else if !ifMetaDto.Update {
				// FIXME: dumping should be avoided here
				/*ifaces, err := c.ifHandler.DumpInterfaces()
				if err != nil {
					c.log.Warnf("dump interfaces failed: %v", err)
					continue
				}
				for _, ifaceDetails := range ifaces {
					if ifaceDetails.Meta.SwIfIndex != ifMetaDto.Metadata.SwIfIndex {
						// not the added interface
						continue
					}
					c.updateIfStateDetails(ifaceDetails)
				}*/
			}

		case <-ctx.Done():
			// stop watching for notifications and periodic statistics reader
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
	ifState.Statistics = &intf.InterfaceState_Statistics{
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
func (c *InterfaceStateUpdater) getIfStateData(swIfIndex uint32, ifName string) (*intf.InterfaceState, bool) {

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
	*intf.InterfaceState, bool) {
	ifName, _, found := c.swIfIndexes.LookupBySwIfIndex(ifIdx)
	if !found {
		return nil, found
	}

	ifState, found := c.getIfStateData(ifIdx, ifName)
	if !found {
		ifState = &intf.InterfaceState{
			IfIndex:    ifIdx,
			Name:       ifName,
			Statistics: &intf.InterfaceState_Statistics{},
		}

		c.ifState[ifIdx] = ifState
		found = true
	}

	return ifState, found
}

// updateIfStateFlags updates the interface state data in memory from provided VPP flags message and returns updated state data.
// NOTE: plugin.ifStateData needs to be locked when calling this function!
func (c *InterfaceStateUpdater) updateIfStateFlags(vppMsg *vppcalls.InterfaceEvent) (
	iface *intf.InterfaceState, found bool) {

	ifState, found := c.getIfStateDataWLookup(vppMsg.SwIfIndex)
	if !found {
		return nil, false
	}
	ifState.LastChange = time.Now().Unix()

	if vppMsg.Deleted {
		ifState.AdminStatus = intf.InterfaceState_DELETED
		ifState.OperStatus = intf.InterfaceState_DELETED
	} else {
		if vppMsg.AdminState == 1 {
			ifState.AdminStatus = intf.InterfaceState_UP
		} else {
			ifState.AdminStatus = intf.InterfaceState_DOWN
		}
		if vppMsg.LinkState == 1 {
			ifState.OperStatus = intf.InterfaceState_UP
		} else {
			ifState.OperStatus = intf.InterfaceState_DOWN
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
		ifState.AdminStatus = intf.InterfaceState_UP
	} else if ifDetails.Meta.AdminState == 0 {
		ifState.AdminStatus = intf.InterfaceState_DOWN
	} else {
		ifState.AdminStatus = intf.InterfaceState_UNKNOWN_STATUS
	}

	if ifDetails.Meta.LinkState == 1 {
		ifState.OperStatus = intf.InterfaceState_UP
	} else if ifDetails.Meta.LinkState == 0 {
		ifState.OperStatus = intf.InterfaceState_DOWN
	} else {
		ifState.OperStatus = intf.InterfaceState_UNKNOWN_STATUS
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
		ifState.Duplex = intf.InterfaceState_HALF
	case 2:
		ifState.Duplex = intf.InterfaceState_FULL
	default:
		ifState.Duplex = intf.InterfaceState_UNKNOWN_DUPLEX
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
	ifState.AdminStatus = intf.InterfaceState_DELETED
	ifState.OperStatus = intf.InterfaceState_DELETED
	ifState.LastChange = time.Now().Unix()

	// this can be post-processed by multiple plugins
	c.publishIfState(&intf.InterfaceNotification{
		Type: intf.InterfaceNotification_UNKNOWN, State: ifState})
}
