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

package l2plugin

import (
	"context"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	l2_api "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"sync"
	"time"
)

// BridgeDomainStateUpdater holds all data required to handle bridge domain state
type BridgeDomainStateUpdater struct {
	Log         logging.Logger
	GoVppmux    govppmux.API
	bdIndex     bdidx.BDIndex
	swIfIndexes ifaceidx.SwIfIndex

	publishBdState func(notification *BridgeDomainStateNotification)
	bdState        map[uint32]*l2.BridgeDomainState_BridgeDomain
	access         sync.Mutex

	vppCh                   *govppapi.Channel
	vppNotifSubs            *govppapi.NotifSubscription
	vppCountersSubs         *govppapi.NotifSubscription
	vppCombinedCountersSubs *govppapi.NotifSubscription
	notificationChan        chan BridgeDomainStateMessage
	bdIdxChan               chan bdidx.ChangeDto

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// BridgeDomainStateNotification contains bridge domain state object with all data published to ETCD
type BridgeDomainStateNotification struct {
	State *l2.BridgeDomainState_BridgeDomain
}

// Init bridge domain state updater
func (plugin *BridgeDomainStateUpdater) Init(ctx context.Context, bdIndexes bdidx.BDIndex, swIfIndexes ifaceidx.SwIfIndex,
	notificationChan chan BridgeDomainStateMessage, publishBdState func(notification *BridgeDomainStateNotification)) (err error) {

	plugin.Log.Info("Initializing BridgeDomainStateUpdater")

	plugin.bdIndex = bdIndexes
	plugin.swIfIndexes = swIfIndexes
	plugin.publishBdState = publishBdState
	plugin.bdState = make(map[uint32]*l2.BridgeDomainState_BridgeDomain)

	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	plugin.bdIdxChan = make(chan bdidx.ChangeDto, 100)
	bdIndexes.WatchNameToIdx(core.PluginName("bdplugin_bdstate"), plugin.bdIdxChan)
	plugin.notificationChan = notificationChan

	var childCtx context.Context
	childCtx, plugin.cancel = context.WithCancel(ctx)

	go plugin.watchVPPNotifications(childCtx)

	return nil
}

// watchVPPNotifications watches for delivery of notifications from VPP.
func (plugin *BridgeDomainStateUpdater) watchVPPNotifications(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	if plugin.notificationChan != nil {
		plugin.Log.Info("watchVPPNotifications for bridge domain state started")
	} else {
		plugin.Log.Error("failed to start watchVPPNotifications for bridge domain state")
		return
	}

	for {
		select {
		case notif := <-plugin.notificationChan:
			msg := notif.Message
			bdName := notif.Name
			switch msg := msg.(type) {
			case *l2_api.BridgeDomainDetails:
				bdState := plugin.processBridgeDomainDetailsNotification(msg, bdName)
				if bdState != nil {
					plugin.publishBdState(&BridgeDomainStateNotification{
						State: bdState,
					})
				}
			default:
				plugin.Log.WithFields(logging.Fields{"MessageName": msg.GetMessageName()}).Debug("L2Plugin: Ignoring unknown VPP notification")
			}
		case bdIdxDto := <-plugin.bdIdxChan:
			bdIdxDto.Done()

		case <-ctx.Done():
			// stop watching for notifications
			return
		}
	}
}

func (plugin *BridgeDomainStateUpdater) processBridgeDomainDetailsNotification(msg *l2_api.BridgeDomainDetails, name string) *l2.BridgeDomainState_BridgeDomain {
	bdState := &l2.BridgeDomainState_BridgeDomain{}
	// Delete case
	if msg.BdID == 0 && name != "" {
		// Mark index to 0 to be removed, but pass name so key can be constructed
		bdState.Index = 0
		bdState.InternalName = name
		return bdState
	}
	bdState.Index = msg.BdID
	name, _, found := plugin.bdIndex.LookupName(msg.BdID)
	if !found {
		plugin.Log.Warnf("Unable to store bridge domain state, index %v is not in the mapping", msg.BdID)
		return bdState
	}
	bdState.InternalName = name
	bdState.InterfaceCount = msg.NSwIfs
	name, _, found = plugin.swIfIndexes.LookupName(msg.BviSwIfIndex)
	if found {
		bdState.BviInterface = name
		bdState.BviInterfaceIndex = msg.BviSwIfIndex
	} else {
		bdState.BviInterface = "not_set"
	}
	bdState.L2Params = getBridgeDomainStateParams(msg)
	bdState.Interfaces = plugin.getBridgeDomainInterfaces(msg)
	bdState.LastChange = time.Now().Unix()

	return bdState
}

func (plugin *BridgeDomainStateUpdater) getBridgeDomainInterfaces(msg *l2_api.BridgeDomainDetails) []*l2.BridgeDomainState_BridgeDomain_Interfaces {
	bdStateInterfaces := []*l2.BridgeDomainState_BridgeDomain_Interfaces{}
	for _, swIfaceDetails := range msg.SwIfDetails {
		bdIfaceState := &l2.BridgeDomainState_BridgeDomain_Interfaces{}
		name, _, found := plugin.swIfIndexes.LookupName(swIfaceDetails.SwIfIndex)
		if !found {
			plugin.Log.Debugf("Interface name for index %v not found for bridge domain status", swIfaceDetails)
			bdIfaceState.Name = "unknown"
		} else {
			bdIfaceState.Name = name
		}
		bdIfaceState.SwIfIndex = swIfaceDetails.SwIfIndex
		bdIfaceState.SplitHorizonGroup = uint32(swIfaceDetails.Shg)
		bdStateInterfaces = append(bdStateInterfaces, bdIfaceState)
	}
	return bdStateInterfaces
}

func getBridgeDomainStateParams(msg *l2_api.BridgeDomainDetails) *l2.BridgeDomainState_BridgeDomain_L2Params {
	params := &l2.BridgeDomainState_BridgeDomain_L2Params{}
	params.Flood = intToBool(msg.Flood)
	params.UnknownUnicastFlood = intToBool(msg.UuFlood)
	params.Forward = intToBool(msg.Forward)
	params.Learn = intToBool(msg.Learn)
	params.ArpTermination = intToBool(msg.ArpTerm)
	params.MacAge = uint32(msg.MacAge)
	return params
}

func intToBool(num uint8) bool {
	if num == 1 {
		return true
	}
	return false
}
