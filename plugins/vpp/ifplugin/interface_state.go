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
	"os"
	"sync"
	"time"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	intf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

var (
	debugIfStates = os.Getenv("DEBUG_IFSTATES") != ""
)

// InterfaceStateUpdater holds state data of all VPP interfaces.
type InterfaceStateUpdater struct {
	log logging.Logger

	kvScheduler    kvs.KVScheduler
	swIfIndexes    ifaceidx.IfaceMetadataIndex
	publishIfState func(notification *intf.InterfaceNotification)

	// access guards access to ifState map
	access  sync.Mutex
	ifState map[uint32]*intf.InterfaceState // swIfIndex

	vppClient vpp.Client

	ifMetaChan chan ifaceidx.IfaceMetadataDto

	ifHandler      vppcalls.InterfaceVppAPI
	ifEvents       chan *vppcalls.InterfaceEvent
	cancelIfEvents func()

	ifsForUpdate   map[uint32]struct{}
	lastIfCounters map[uint32]govppapi.InterfaceCounters
	ifStats        govppapi.InterfaceStats

	lastIfNotif time.Time
	lastIfMeta  time.Time

	ctx    context.Context
	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Init members (channels, maps...) and start go routines
func (c *InterfaceStateUpdater) Init(
	ctx context.Context,
	logger logging.PluginLogger,
	kvScheduler kvs.KVScheduler,
	vppClient vpp.Client,
	swIfIndexes ifaceidx.IfaceMetadataIndex,
	publishIfState func(*intf.InterfaceNotification),
	readCounters bool,
) error {
	c.log = logger.NewLogger("if-state")

	// Mappings
	c.swIfIndexes = swIfIndexes

	c.vppClient = vppClient
	c.kvScheduler = kvScheduler
	c.publishIfState = publishIfState
	c.ifState = make(map[uint32]*intf.InterfaceState)

	c.ifsForUpdate = make(map[uint32]struct{})
	c.lastIfCounters = make(map[uint32]govppapi.InterfaceCounters)

	// Init handlers
	c.ifHandler = vppcalls.CompatibleInterfaceVppHandler(c.vppClient, logger.NewLogger("if-handler"))

	c.ifMetaChan = make(chan ifaceidx.IfaceMetadataDto, 1000)
	swIfIndexes.WatchInterfaces("ifplugin_ifstate", c.ifMetaChan)

	c.ifEvents = make(chan *vppcalls.InterfaceEvent, 1000)

	// Create child context
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Watch for incoming notifications
	c.wg.Add(1)
	go c.watchVPPNotifications(c.ctx)

	// Periodically read VPP counters and combined counters for VPP statistics
	if disableInterfaceStats {
		c.log.Warnf("reading interface stats is DISABLED!")
	} else if readCounters {
		c.wg.Add(1)
		go c.startReadingCounters(c.ctx)
	}

	if disableStatusPublishing {
		c.log.Warnf("publishing interface status is DISABLED!")
	} else {
		c.wg.Add(1)
		go c.startUpdatingIfStateDetails(c.ctx)
	}

	return nil
}

// AfterInit subscribes for watching VPP notifications on previously initialized channel
func (c *InterfaceStateUpdater) AfterInit() error {
	if err := c.subscribeVPPNotifications(c.ctx); err != nil {
		return err
	}
	c.vppClient.OnReconnect(func() {
		c.cancelIfEvents()
		if err := c.subscribeVPPNotifications(c.ctx); err != nil {
			c.log.Warnf("WatchInterfaceEvents failed: %v", err)
		}
	})
	return nil
}

// Close unsubscribes from interface state notifications from VPP & GOVPP channel
func (c *InterfaceStateUpdater) Close() error {
	c.cancel()
	c.wg.Wait()
	return nil
}

// subscribeVPPNotifications subscribes for interface state notifications from VPP.
func (c *InterfaceStateUpdater) subscribeVPPNotifications(ctx context.Context) error {
	ctx, c.cancelIfEvents = context.WithCancel(ctx)
	if err := c.ifHandler.WatchInterfaceEvents(ctx, c.ifEvents); err != nil {
		return err
	}
	return nil
}

// watchVPPNotifications watches for delivery of notifications from VPP.
func (c *InterfaceStateUpdater) watchVPPNotifications(ctx context.Context) {
	defer c.wg.Done()

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
				c.processIfMetaCreate(ifMetaDto.Metadata.SwIfIndex)
			}

		case <-ctx.Done():
			// stop watching for notifications and periodic statistics reader
			c.log.Debug("Interface state VPP notification watcher stopped")
			return
		}
	}
}

func (c *InterfaceStateUpdater) startUpdatingIfStateDetails(ctx context.Context) {
	defer c.wg.Done()

	/*timer := time.NewTimer(PeriodicPollingPeriod)
	  if !ifUpdateTimer.Stop() {
	  	<-ifUpdateTimer.C
	  }
	  ifUpdateTimer.Reset(PeriodicPollingPeriod)*/

	tick := time.NewTicker(StateUpdateDelay)
	for {
		select {
		case <-tick.C:
			c.doUpdatesIfStateDetails()

		case <-ctx.Done():
			c.log.Debug("update if state details polling stopped")
			return
		}
	}
}

// startReadingCounters periodically reads statistics for all interfaces
func (c *InterfaceStateUpdater) startReadingCounters(ctx context.Context) {
	defer c.wg.Done()

	tick := time.NewTicker(PeriodicPollingPeriod)
	for {
		select {
		case <-tick.C:
			statsClient := c.vppClient.Stats()
			if statsClient == nil {
				c.log.Warnf("VPP stats client not available")
				// TODO: use retry with backoff instead of returning here
				return
			}
			c.doInterfaceStatsRead(statsClient)

		case <-ctx.Done():
			c.log.Debug("Interface state VPP periodic polling stopped")
			return
		}
	}
}

func (c *InterfaceStateUpdater) processIfMetaCreate(swIfIdx uint32) {
	c.access.Lock()
	defer c.access.Unlock()

	c.lastIfMeta = time.Now()

	c.ifsForUpdate[swIfIdx] = struct{}{}
}

func (c *InterfaceStateUpdater) doUpdatesIfStateDetails() {
	c.access.Lock()

	// prevent reading stats if last interface notification has been
	// received in less than polling period
	if time.Since(c.lastIfMeta) < StateUpdateDelay {
		c.access.Unlock()
		return
	}
	if len(c.ifsForUpdate) == 0 {
		c.access.Unlock()
		return
	}

	c.log.Debugf("updating interface states for %d interfaces", len(c.ifsForUpdate))

	var ifIdxs []uint32
	if len(c.ifsForUpdate) < 1000 {
		for ifIdx := range c.ifsForUpdate {
			ifIdxs = append(ifIdxs, ifIdx)
		}
	}
	// clear interfaces for update
	c.ifsForUpdate = make(map[uint32]struct{})

	// we dont want to lock during potentially long dump call
	c.access.Unlock()

	ifaces, err := c.ifHandler.DumpInterfaceStates(ifIdxs...)
	if err != nil {
		c.log.Warnf("dumping interface states failed: %v", err)
		return
	}

	c.access.Lock()
	for _, ifaceDetails := range ifaces {
		if ifaceDetails == nil {
			// this interface was removed and the updater hasn't yet received notification
			continue
		}
		c.updateIfStateDetails(ifaceDetails)
	}
	c.access.Unlock()
}

// doInterfaceStatsRead dumps statistics using interface filter and processes them
func (c *InterfaceStateUpdater) doInterfaceStatsRead(statsClient govppapi.StatsProvider) {
	c.access.Lock()
	defer c.access.Unlock()

	// prevent reading stats if last interface notification has been
	// received in less than polling period
	if time.Since(c.lastIfNotif) < StateUpdateDelay {
		return
	}

	err := statsClient.GetInterfaceStats(&c.ifStats)
	if err != nil {
		// TODO add some counter to prevent it log forever
		c.log.Errorf("failed to read statistics data: %v", err)
	}
	if len(c.ifStats.Interfaces) == 0 {
		return
	}

	for i, ifCounters := range c.ifStats.Interfaces {
		index := uint32(i)
		if last, ok := c.lastIfCounters[index]; ok && last == ifCounters {
			continue
		}
		c.lastIfCounters[index] = ifCounters
		c.processInterfaceStatEntry(ifCounters)
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
		InPackets:       ifCounters.Rx.Packets,
		InBytes:         ifCounters.Rx.Bytes,
		OutPackets:      ifCounters.Tx.Packets,
		OutBytes:        ifCounters.Tx.Bytes,
	}

	c.publishIfState(&intf.InterfaceNotification{
		Type:  intf.InterfaceNotification_COUNTERS,
		State: ifState,
	})
}

// processIfStateEvent process a VPP state event notification.
func (c *InterfaceStateUpdater) processIfStateEvent(notif *vppcalls.InterfaceEvent) {
	c.access.Lock()
	defer c.access.Unlock()

	c.lastIfNotif = time.Now()

	// update and return if state data
	ifState, found := c.updateIfStateFlags(notif)
	if !found {
		return
	}

	if debugIfStates {
		c.log.Debugf("Interface state notification for %s (idx: %d): %+v",
			ifState.Name, ifState.IfIndex, notif)
	}

	// store data in ETCD
	c.publishIfState(&intf.InterfaceNotification{
		Type:  intf.InterfaceNotification_UPDOWN,
		State: ifState,
	})
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
func (c *InterfaceStateUpdater) getIfStateDataWLookup(ifIdx uint32) (*intf.InterfaceState, bool) {
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
	iface *intf.InterfaceState, found bool,
) {
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
func (c *InterfaceStateUpdater) updateIfStateDetails(ifDetails *vppcalls.InterfaceState) {
	ifState, found := c.getIfStateDataWLookup(ifDetails.SwIfIndex)
	if !found {
		return
	}

	ifState.InternalName = ifDetails.InternalName
	ifState.PhysAddress = ifDetails.PhysAddress.String()
	ifState.AdminStatus = ifDetails.AdminState
	ifState.OperStatus = ifDetails.LinkState
	ifState.Speed = ifDetails.LinkSpeed
	ifState.Duplex = ifDetails.LinkDuplex
	ifState.Mtu = uint32(ifDetails.LinkMTU)

	c.publishIfState(&intf.InterfaceNotification{State: ifState})
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
	c.publishIfState(&intf.InterfaceNotification{State: ifState})
}
