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
	"bytes"
	"context"
	"net"
	"os"
	"sync"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/pkg/errors"

	"github.com/ligato/vpp-agent/plugins/govppmux"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/stats"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	intf "github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
)

// counterType is the basic counter type - contains only packet statistics.
type counterType int

// constants as defined in the vnet_interface_counter_type_t enum in 'vnet/interface.h'
const (
	Drop    counterType = 0
	Punt                = 1
	IPv4                = 2
	IPv6                = 3
	RxNoBuf             = 4
	RxMiss              = 5
	RxError             = 6
	TxError             = 7
	MPLS                = 8
)

// combinedCounterType is the extended counter type - contains both packet and byte statistics.
type combinedCounterType int

// constants as defined in the vnet_interface_counter_type_t enum in 'vnet/interface.h'
const (
	Rx combinedCounterType = 0
	Tx                     = 1
)

const (
	megabit = 1000000 // One megabit in bytes
)

// InterfaceStateUpdater holds state data of all VPP interfaces.
type InterfaceStateUpdater struct {
	log logging.Logger

	kvScheduler    scheduler.KVScheduler
	swIfIndexes    ifaceidx.IfaceMetadataIndex
	publishIfState func(notification *intf.InterfaceNotification)

	ifState map[uint32]*intf.InterfaceState // swIfIndex to state data map
	access  sync.Mutex                      // lock for the state data map

	vppCh                   govppapi.Channel
	vppNotifSubs            govppapi.SubscriptionCtx
	vppCountersSubs         govppapi.SubscriptionCtx
	vppCombinedCountersSubs govppapi.SubscriptionCtx
	notifChan               chan govppapi.Message
	ifMetaChan              chan ifaceidx.IfaceMetadataDto

	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Init members (channels, maps...) and start go routines
func (c *InterfaceStateUpdater) Init(ctx context.Context, logger logging.PluginLogger, kvScheduler scheduler.KVScheduler,
	goVppMux govppmux.API, swIfIndexes ifaceidx.IfaceMetadataIndex,
	publishIfState func(notification *intf.InterfaceNotification)) (err error) {

	// Logger
	c.log = logger.NewLogger("if-state")

	// Mappings
	c.swIfIndexes = swIfIndexes

	c.kvScheduler = kvScheduler
	c.publishIfState = publishIfState
	c.ifState = make(map[uint32]*intf.InterfaceState)

	// VPP channel
	c.vppCh, err = goVppMux.NewAPIChannel()
	if err != nil {
		return errors.Errorf("failed to create API channel: %v", err)
	}

	c.ifMetaChan = make(chan ifaceidx.IfaceMetadataDto, 100)
	swIfIndexes.WatchInterfaces("ifplugin_ifstate", c.ifMetaChan)
	c.notifChan = make(chan govppapi.Message, 100)

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
	var err error
	// subscribe for receiving SwInterfaceEvents notifications
	if c.vppNotifSubs, err = c.vppCh.SubscribeNotification(c.notifChan, &interfaces.SwInterfaceEvent{}); err != nil {
		return errors.Errorf("failed to subscribe VPP notification (sw_interface_event): %v", err)
	}

	// subscribe for receiving VnetInterfaceSimpleCounters notifications
	if c.vppCountersSubs, err = c.vppCh.SubscribeNotification(c.notifChan, &stats.VnetInterfaceSimpleCounters{}); err != nil {
		return errors.Errorf("failed to subscribe VPP notification (vnet_interface_simple_counters): %v", err)
	}

	// subscribe for receiving VnetInterfaceCombinedCounters notifications
	if c.vppCombinedCountersSubs, err = c.vppCh.SubscribeNotification(c.notifChan, &stats.VnetInterfaceCombinedCounters{}); err != nil {
		return errors.Errorf("failed to subscribe VPP notification (vnet_interface_combined_counters): %v", err)
	}

	wantIfEventsReply := &interfaces.WantInterfaceEventsReply{}
	// enable interface state notifications from VPP
	err = c.vppCh.SendRequest(&interfaces.WantInterfaceEvents{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(wantIfEventsReply)
	if err != nil {
		return errors.Errorf("failed to get interface events: %v", err)
	}

	wantSimpleStatsReply := &stats.WantInterfaceSimpleStatsReply{}
	// enable interface counters notifications from VPP
	err = c.vppCh.SendRequest(&stats.WantInterfaceSimpleStats{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(wantSimpleStatsReply)
	if err != nil {
		return errors.Errorf("failed to subscribe for interface simple stats: %v", err)
	}

	wantCombinedStatsReply := &stats.WantInterfaceCombinedStatsReply{}
	// enable interface counters notifications from VPP
	err = c.vppCh.SendRequest(&stats.WantInterfaceCombinedStats{
		PID:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(wantCombinedStatsReply)
	if err != nil {
		return errors.Errorf("failed to subscribe for interface combined stats: %v", err)
	}

	return nil
}

// Close unsubscribes from interface state notifications from VPP & GOVPP channel
func (c *InterfaceStateUpdater) Close() error {
	c.cancel()
	c.wg.Wait()

	if c.vppNotifSubs != nil {
		if err := c.vppNotifSubs.Unsubscribe(); err != nil {
			return errors.Errorf("failed to unsubscribe interface state notification on close: %v", err)
		}
	}
	if c.vppCountersSubs != nil {
		if err := c.vppCountersSubs.Unsubscribe(); err != nil {
			return errors.Errorf("failed to unsubscribe interface state counters on close: %v", err)
		}
	}
	if c.vppCombinedCountersSubs != nil {
		if err := c.vppCombinedCountersSubs.Unsubscribe(); err != nil {
			return errors.Errorf("failed to unsubscribe interface state combined counters on close: %v", err)
		}
	}

	if err := safeclose.Close(c.vppCh, c.notifChan); err != nil {
		return errors.Errorf("failed to safe close interface state: %v", err)
	}

	return nil
}

// watchVPPNotifications watches for delivery of notifications from VPP.
func (c *InterfaceStateUpdater) watchVPPNotifications(ctx context.Context) {
	defer c.wg.Done()

	if c.notifChan != nil {
		c.log.Debug("Interface state VPP notification watcher started")
	} else {
		c.log.Warn("Interface state VPP notification does not start: the channel c nil")
		return
	}

	for {
		select {
		case msg := <-c.notifChan:
			// if the notification is a result of a configuration change,
			// make sure the associated transaction has already finalized
			c.kvScheduler.TransactionBarrier()

			switch notif := msg.(type) {
			case *interfaces.SwInterfaceEvent:
				c.processIfStateNotification(notif)
			case *stats.VnetInterfaceSimpleCounters:
				c.processIfCounterNotification(notif)
			case *stats.VnetInterfaceCombinedCounters:
				c.processIfCombinedCounterNotification(notif)
			default:
				c.log.Debugf("Ignoring unknown VPP notification: %s, %v",
					msg.GetMessageName(), msg)
			}

		case ifMetaDto := <-c.ifMetaChan:
			if ifMetaDto.Del {
				c.setIfStateDeleted(ifMetaDto.Metadata.SwIfIndex, ifMetaDto.Name)
			} else if !ifMetaDto.Update {
				// process new interface (no way to filter by swIfIndex, need to dump all of them)
				req := &interfaces.SwInterfaceDump{}
				reqCtx := c.vppCh.SendMultiRequest(req)

				for {
					msg := &interfaces.SwInterfaceDetails{}
					stop, err := reqCtx.ReceiveReply(msg)
					if stop {
						break
					}
					if err != nil {
						c.log.Warnf("failed to receive interface dump details: %v", err)
						continue
					}
					if msg.SwIfIndex != ifMetaDto.Metadata.SwIfIndex {
						// not the added interface
						continue
					}
					c.updateIfStateDetails(msg)
				}
			}

		case <-ctx.Done():
			// stop watching for notifications
			c.log.Debug("Interface state VPP notification watcher stopped")
			return
		}
	}
}

// processIfStateNotification process a VPP state notification.
func (c *InterfaceStateUpdater) processIfStateNotification(notif *interfaces.SwInterfaceEvent) {
	// update and return if state data
	ifState, found := c.updateIfStateFlags(notif)
	if !found {
		return
	}
	c.log.Debugf("Interface state notification for %s (Idx %d)", ifState.Name, ifState.IfIndex)

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
func (c *InterfaceStateUpdater) updateIfStateFlags(vppMsg *interfaces.SwInterfaceEvent) (
	iface *intf.InterfaceState, found bool) {

	ifState, found := c.getIfStateDataWLookup(vppMsg.SwIfIndex)
	if !found {
		return nil, false
	}
	ifState.LastChange = time.Now().Unix()

	if vppMsg.Deleted == 1 {
		ifState.AdminStatus = intf.InterfaceState_DELETED
		ifState.OperStatus = intf.InterfaceState_DELETED
	} else {
		if vppMsg.AdminUpDown == 1 {
			ifState.AdminStatus = intf.InterfaceState_UP
		} else {
			ifState.AdminStatus = intf.InterfaceState_DOWN
		}
		if vppMsg.LinkUpDown == 1 {
			ifState.OperStatus = intf.InterfaceState_UP
		} else {
			ifState.OperStatus = intf.InterfaceState_DOWN
		}
	}
	return ifState, true
}

// processIfCounterNotification processes a VPP (simple) counter message.
func (c *InterfaceStateUpdater) processIfCounterNotification(counter *stats.VnetInterfaceSimpleCounters) {
	c.access.Lock()
	defer c.access.Unlock()

	for i := uint32(0); i < counter.Count; i++ {
		swIfIndex := counter.FirstSwIfIndex + i
		ifState, found := c.getIfStateDataWLookup(swIfIndex)
		if !found {
			continue
		}
		ifStats := ifState.Statistics
		packets := counter.Data[i]
		switch counterType(counter.VnetCounterType) {
		case Drop:
			ifStats.DropPackets = packets
		case Punt:
			ifStats.PuntPackets = packets
		case IPv4:
			ifStats.Ipv4Packets = packets
		case IPv6:
			ifStats.Ipv6Packets = packets
		case RxNoBuf:
			ifStats.InNobufPackets = packets
		case RxMiss:
			ifStats.InMissPackets = packets
		case RxError:
			ifStats.InErrorPackets = packets
		case TxError:
			ifStats.OutErrorPackets = packets
		}
	}
}

// processIfCombinedCounterNotification processes a VPP message with combined counters.
func (c *InterfaceStateUpdater) processIfCombinedCounterNotification(counter *stats.VnetInterfaceCombinedCounters) {
	c.access.Lock()
	defer c.access.Unlock()

	if counter.VnetCounterType > Tx {
		// TODO: process other types of combined counters (RX/TX for unicast/multicast/broadcast)
		return
	}

	for i := uint32(0); i < counter.Count; i++ {
		swIfIndex := counter.FirstSwIfIndex + i
		ifState, found := c.getIfStateDataWLookup(swIfIndex)
		if !found {
			continue
		}
		ifStats := ifState.Statistics
		if combinedCounterType(counter.VnetCounterType) == Rx {
			ifStats.InPackets = counter.Data[i].Packets
			ifStats.InBytes = counter.Data[i].Bytes
		} else if combinedCounterType(counter.VnetCounterType) == Tx {
			ifStats.OutPackets = counter.Data[i].Packets
			ifStats.OutBytes = counter.Data[i].Bytes
			// publish Tx + Rx for this interface (Tx counters are received after RX counters)
			c.publishIfState(&intf.InterfaceNotification{
				Type: intf.InterfaceNotification_UPDOWN, State: ifState})
		}
	}
}

// updateIfStateDetails updates the interface state data in memory from provided VPP details message.
func (c *InterfaceStateUpdater) updateIfStateDetails(ifDetails *interfaces.SwInterfaceDetails) {
	c.access.Lock()
	defer c.access.Unlock()

	ifState, found := c.getIfStateDataWLookup(ifDetails.SwIfIndex)
	if !found {
		return
	}

	ifState.InternalName = string(bytes.SplitN(ifDetails.InterfaceName, []byte{0x00}, 2)[0])

	if ifDetails.AdminUpDown == 1 {
		ifState.AdminStatus = intf.InterfaceState_UP
	} else {
		ifState.AdminStatus = intf.InterfaceState_DOWN
	}

	if ifDetails.LinkUpDown == 1 {
		ifState.OperStatus = intf.InterfaceState_UP
	} else {
		ifState.OperStatus = intf.InterfaceState_DOWN
	}

	hwAddr := net.HardwareAddr(ifDetails.L2Address[:ifDetails.L2AddressLength])
	ifState.PhysAddress = hwAddr.String()

	ifState.Mtu = uint32(ifDetails.LinkMtu)

	switch ifDetails.LinkSpeed {
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

	switch ifDetails.LinkSpeed {
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
