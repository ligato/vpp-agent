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
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"bytes"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/stats"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
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
	Log            logging.Logger
	GoVppmux       govppmux.API
	swIfIndexes    ifaceidx.SwIfIndex
	publishIfState func(notification *intf.InterfaceStateNotification)

	ifState map[uint32]*intf.InterfacesState_Interface // swIfIndex to state data map
	access  sync.Mutex                                 // lock for the state data map

	vppCh                   *govppapi.Channel
	vppNotifSubs            *govppapi.NotifSubscription
	vppCountersSubs         *govppapi.NotifSubscription
	vppCombinedCountersSubs *govppapi.NotifSubscription
	notifChan               chan govppapi.Message
	swIdxChan               chan ifaceidx.SwIfIdxDto

	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Init members (channels, maps...) and start go routines
func (plugin *InterfaceStateUpdater) Init(ctx context.Context,
	swIfIndexes ifaceidx.SwIfIndex, notifChan chan govppapi.Message,
	publishIfState func(notification *intf.InterfaceStateNotification)) (err error) {

	plugin.Log.Info("Initializing InterfaceStateUpdater")

	plugin.swIfIndexes = swIfIndexes
	plugin.publishIfState = publishIfState
	plugin.ifState = make(map[uint32]*intf.InterfacesState_Interface)

	// init VPP API channel
	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	plugin.swIdxChan = make(chan ifaceidx.SwIfIdxDto, 100)
	swIfIndexes.WatchNameToIdx(core.PluginName("ifplugin_ifstate"), plugin.swIdxChan)
	plugin.notifChan = notifChan

	// create child context
	var childCtx context.Context
	childCtx, plugin.cancel = context.WithCancel(ctx)

	// watch for incoming notifications
	go plugin.watchVPPNotifications(childCtx)

	return nil
}

// AfterInit subscribes for watching VPP notifications on previously initialized channel
func (plugin *InterfaceStateUpdater) AfterInit() (err error) {
	plugin.subscribeVPPNotifications()

	return nil
}

// subscribeVPPNotifications subscribes for interface state notifications from VPP.
func (plugin *InterfaceStateUpdater) subscribeVPPNotifications() error {
	var err error

	// subscribe for receiving SwInterfaceEvents notifications
	plugin.vppNotifSubs, err = plugin.vppCh.SubscribeNotification(plugin.notifChan, interfaces.NewSwInterfaceEvent)
	if err != nil {
		return err
	}

	// subscribe for receiving VnetInterfaceSimpleCounters notifications
	plugin.vppCountersSubs, err = plugin.vppCh.SubscribeNotification(plugin.notifChan, interfaces.NewVnetInterfaceSimpleCounters)
	if err != nil {
		return err
	}

	// subscribe for receiving VnetInterfaceCombinedCounters notifications
	plugin.vppCombinedCountersSubs, err = plugin.vppCh.SubscribeNotification(plugin.notifChan, interfaces.NewVnetInterfaceCombinedCounters)
	if err != nil {
		return err
	}

	wantInterfaceEventsReply := &interfaces.WantInterfaceEventsReply{}
	// enable interface state notifications from VPP
	err = plugin.vppCh.SendRequest(&interfaces.WantInterfaceEvents{
		Pid:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(wantInterfaceEventsReply)
	plugin.Log.Debug("wantInterfaceEventsReply: ", wantInterfaceEventsReply, " ", err)
	if err != nil {
		return err
	}
	if wantInterfaceEventsReply.Retval != 0 {
		return fmt.Errorf(fmt.Sprintf("wantStatsReply=%d", wantInterfaceEventsReply.Retval))
	}

	wantStatsReply := &stats.WantStatsReply{}
	// enable interface counters notifications from VPP
	err = plugin.vppCh.SendRequest(&stats.WantStats{
		Pid:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(wantStatsReply)
	plugin.Log.Debug("wantStatsReply: ", wantStatsReply, " ", err)
	if err != nil {
		return err
	}
	if wantStatsReply.Retval != 0 {
		return fmt.Errorf(fmt.Sprintf("wantStatsReply=%d", wantStatsReply.Retval))
	}

	return nil
}

// Close unsubscribes from interface state notifications from VPP & GOVPP channel
func (plugin *InterfaceStateUpdater) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	if plugin.vppNotifSubs != nil {
		plugin.vppCh.UnsubscribeNotification(plugin.vppNotifSubs)
	}
	if plugin.vppCountersSubs != nil {
		plugin.vppCh.UnsubscribeNotification(plugin.vppCountersSubs)
	}
	if plugin.vppCombinedCountersSubs != nil {
		plugin.vppCh.UnsubscribeNotification(plugin.vppCombinedCountersSubs)
	}

	_, err := safeclose.CloseAll(plugin.vppCh, plugin.swIdxChan)
	return err
}

// watchVPPNotifications watches for delivery of notifications from VPP.
func (plugin *InterfaceStateUpdater) watchVPPNotifications(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	if plugin.notifChan != nil {
		plugin.Log.Info("watchVPPNotifications starting")
	} else {
		plugin.Log.Error("watchVPPNotifications will not start")
		return
	}

	for {
		select {
		case msg := <-plugin.notifChan:
			switch notif := msg.(type) {
			case *interfaces.SwInterfaceEvent:
				plugin.processIfStateNotification(notif)
			case *interfaces.VnetInterfaceSimpleCounters:
				plugin.processIfCounterNotification(notif)
			case *interfaces.VnetInterfaceCombinedCounters:
				plugin.processIfCombinedCounterNotification(notif)
			case *interfaces.SwInterfaceDetails:
				plugin.updateIfStateDetails(notif)
			default:
				plugin.Log.WithFields(logging.Fields{"MessageName": msg.GetMessageName()}).Debug("Ignoring unknown VPP notification")
			}

		case swIdxDto := <-plugin.swIdxChan:
			if swIdxDto.Del {
				plugin.setIfStateDeleted(swIdxDto.Idx)
			}
			swIdxDto.Done()

		case <-ctx.Done():
			// stop watching for notifications
			return
		}
	}
}

// processIfStateNotification process a VPP state notification.
func (plugin *InterfaceStateUpdater) processIfStateNotification(notif *interfaces.SwInterfaceEvent) {
	//plugin.access.Lock() not needed because of channel synchronization
	//defer plugin.access.Unlock()

	// update and return if state data
	ifState, found, err := plugin.updateIfStateFlags(notif)
	if !found {
		plugin.Log.WithField("swIfIndex", notif.SwIfIndex).
			Debug("processIfStateNotification but the swIfIndex is not event registered")
		return
	}
	if err != nil {
		plugin.Log.Warn(err)
		return
	}

	plugin.Log.WithFields(logging.Fields{"ifName": ifState.Name, "swIfIndex": notif.SwIfIndex, "AdminUpDown": notif.AdminUpDown,
		"LinkUpDown": notif.LinkUpDown, "Deleted": notif.Deleted}).Debug("Interface state change notification.")

	// store data in ETCD
	plugin.publishIfState(&intf.InterfaceStateNotification{
		Type: intf.UPDOWN, State: ifState})
}

// getIfStateData returns interface state data structure for the specified interface index (creates it if it does not exist).
// NOTE: plugin.ifStateData needs to be locked when calling this function!
func (plugin *InterfaceStateUpdater) getIfStateData(swIfIndex uint32) (
	iface *intf.InterfacesState_Interface, found bool, err error) {

	ifState, ok := plugin.ifState[swIfIndex]
	if ok {
		return ifState, true, nil
	}
	ifName, _, found := plugin.swIfIndexes.LookupName(swIfIndex)
	if !found {
		return nil, found, nil
	}

	ifState = &intf.InterfacesState_Interface{
		IfIndex:    swIfIndex,
		Name:       ifName,
		Statistics: &intf.InterfacesState_Interface_Statistics{},
	}
	plugin.ifState[swIfIndex] = ifState
	return ifState, found, nil
}

// updateIfStateFlags updates the interface state data in memory from provided VPP flags message and returns updated state data.
// NOTE: plugin.ifStateData needs to be locked when calling this function!
func (plugin *InterfaceStateUpdater) updateIfStateFlags(vppMsg *interfaces.SwInterfaceEvent) (
	iface *intf.InterfacesState_Interface, found bool, err error) {

	ifState, found, err := plugin.getIfStateData(vppMsg.SwIfIndex)
	if !found {
		return nil, false, err
	}
	if err != nil {
		return nil, false, err
	}
	ifState.LastChange = time.Now().Unix()

	if vppMsg.Deleted == 1 {
		ifState.AdminStatus = intf.InterfacesState_Interface_DELETED
		ifState.OperStatus = intf.InterfacesState_Interface_DELETED
	} else {
		if vppMsg.AdminUpDown == 1 {
			ifState.AdminStatus = intf.InterfacesState_Interface_UP
		} else {
			ifState.AdminStatus = intf.InterfacesState_Interface_DOWN
		}
		if vppMsg.LinkUpDown == 1 {
			ifState.OperStatus = intf.InterfacesState_Interface_UP
		} else {
			ifState.OperStatus = intf.InterfacesState_Interface_DOWN
		}
	}
	return ifState, true, nil
}

// processIfCounterNotification processes a VPP (simple) counter message.
func (plugin *InterfaceStateUpdater) processIfCounterNotification(counter *interfaces.VnetInterfaceSimpleCounters) {
	plugin.access.Lock()
	defer plugin.access.Unlock()

	for i := uint32(0); i < counter.Count; i++ {
		swIfIndex := counter.FirstSwIfIndex + i
		ifState, found, err := plugin.getIfStateData(swIfIndex)
		if !found {
			plugin.Log.WithField("swIfIndex", swIfIndex).
				Debug("processIfCounterNotification but the swIfIndex is not event registered")
			continue
		}
		if err != nil {
			plugin.Log.Warn(err)
			continue
		}
		stats := ifState.Statistics
		packets := counter.Data[i]
		switch counterType(counter.VnetCounterType) {
		case Drop:
			stats.DropPackets = packets
		case Punt:
			stats.PuntPackets = packets
		case IPv4:
			stats.Ipv4Packets = packets
		case IPv6:
			stats.Ipv6Packets = packets
		case RxNoBuf:
			stats.InNobufPackets = packets
		case RxMiss:
			stats.InMissPackets = packets
		case RxError:
			stats.InErrorPackets = packets
		case TxError:
			stats.OutErrorPackets = packets
		}
	}
}

// processIfCombinedCounterNotification processes a VPP message with combined counters.
func (plugin *InterfaceStateUpdater) processIfCombinedCounterNotification(counter *interfaces.VnetInterfaceCombinedCounters) {
	plugin.access.Lock()
	defer plugin.access.Unlock()

	save := false
	for i := uint32(0); i < counter.Count; i++ {
		swIfIndex := counter.FirstSwIfIndex + i
		ifState, found, err := plugin.getIfStateData(swIfIndex)
		if !found {
			plugin.Log.WithField("swIfIndex", swIfIndex).
				Debug("processIfCombinedCounterNotification but the swIfIndex is not event registered")
			continue
		}
		if err != nil {
			plugin.Log.Warn(err)
			continue
		}
		stats := ifState.Statistics
		if combinedCounterType(counter.VnetCounterType) == Rx {
			stats.InPackets = counter.Data[i].Packets
			stats.InBytes = counter.Data[i].Bytes
		} else {
			stats.OutPackets = counter.Data[i].Packets
			stats.OutBytes = counter.Data[i].Bytes

			// this was the last counter, we can now write the stats to ETCD
			save = true
		}
	}
	if save {
		// store counters of all interfaces into ETCD
		for _, counter := range plugin.ifState {
			//plugin.deps.DB.Put(intf.InterfaceStateKey(c.Name), counter)
			plugin.publishIfState(&intf.InterfaceStateNotification{
				Type: intf.COUNTERS, State: counter})
		}
	}
}

// updateIfStateDetails updates the interface state data in memory from provided VPP details message.
func (plugin *InterfaceStateUpdater) updateIfStateDetails(ifDetails *interfaces.SwInterfaceDetails) {
	plugin.access.Lock()
	defer plugin.access.Unlock()

	ifState, found, err := plugin.getIfStateData(ifDetails.SwIfIndex)
	if !found {
		plugin.Log.WithField("swIfIndex", ifDetails.SwIfIndex).
			Debug("updateIfStateDetails but the swIfIndex is not event registered")
		return
	}
	if err != nil {
		plugin.Log.Warn(err)
		return
	}

	ifState.InternalName = string(bytes.Trim(ifDetails.InterfaceName, "\x00"))

	if ifDetails.AdminUpDown == 1 {
		ifState.AdminStatus = intf.InterfacesState_Interface_UP
	} else {
		ifState.AdminStatus = intf.InterfacesState_Interface_DOWN
	}

	if ifDetails.LinkUpDown == 1 {
		ifState.OperStatus = intf.InterfacesState_Interface_UP
	} else {
		ifState.OperStatus = intf.InterfacesState_Interface_DOWN
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
		ifState.Duplex = intf.InterfacesState_Interface_HALF
	case 2:
		ifState.Duplex = intf.InterfacesState_Interface_FULL
	default:
		ifState.Duplex = intf.InterfacesState_Interface_UNKNOWN_DUPLEX
	}

	plugin.publishIfState(&intf.InterfaceStateNotification{
		Type: intf.UNKNOWN, State: ifState})
}

// setIfStateDeleted marks the interface as deleted in the state data structure in memory.
func (plugin *InterfaceStateUpdater) setIfStateDeleted(swIfIndex uint32) {
	plugin.access.Lock()
	defer plugin.access.Unlock()

	ifState, found, err := plugin.getIfStateData(swIfIndex)
	if !found {
		plugin.Log.WithField("swIfIndex", swIfIndex).
			Debug("notification delete but the swIfIndex is not event registered")
		return
	}
	if err != nil {
		plugin.Log.Warn(err)
		return
	}
	ifState.AdminStatus = intf.InterfacesState_Interface_DELETED
	ifState.OperStatus = intf.InterfacesState_Interface_DELETED
	ifState.LastChange = time.Now().Unix()

	// this can be post-processed by multiple plugins
	plugin.publishIfState(&intf.InterfaceStateNotification{
		Type: intf.COUNTERS, State: ifState})
}
