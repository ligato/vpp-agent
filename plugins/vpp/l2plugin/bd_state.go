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
	"sync"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	l2_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/l2plugin/l2idx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
)

// BridgeDomainStateUpdater holds all data required to handle bridge domain state.
type BridgeDomainStateUpdater struct {
	log    logging.Logger
	mx     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// In-memory mappings
	ifIndexes ifaceidx.SwIfIndex
	bdIndexes l2idx.BDIndex

	// State publisher
	publishBdState func(notification *BridgeDomainStateNotification)
	bdState        map[uint32]*l2.BridgeDomainState_BridgeDomain

	// VPP channel
	vppCh *govppapi.Channel

	// Notification subscriptions
	vppNotifSubs            *govppapi.NotifSubscription
	vppCountersSubs         *govppapi.NotifSubscription
	vppCombinedCountersSubs *govppapi.NotifSubscription
	notificationChan        chan BridgeDomainStateMessage // Injected, do not close here
	bdIdxChan               chan l2idx.BdChangeDto
}

// BridgeDomainStateNotification contains bridge domain state object with all data published to ETCD.
type BridgeDomainStateNotification struct {
	State *l2.BridgeDomainState_BridgeDomain
}

// Init bridge domain state updater.
func (plugin *BridgeDomainStateUpdater) Init(logger logging.PluginLogger, goVppMux govppmux.API, ctx context.Context, bdIndexes l2idx.BDIndex, swIfIndexes ifaceidx.SwIfIndex,
	notificationChan chan BridgeDomainStateMessage, publishBdState func(notification *BridgeDomainStateNotification)) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-l2-bd-state")
	plugin.log.Info("Initializing BridgeDomainStateUpdater")

	// Mappings
	plugin.bdIndexes = bdIndexes
	plugin.ifIndexes = swIfIndexes

	// State publisher
	plugin.notificationChan = notificationChan
	plugin.publishBdState = publishBdState
	plugin.bdState = make(map[uint32]*l2.BridgeDomainState_BridgeDomain)

	// VPP channel
	plugin.vppCh, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Name-to-index watcher
	plugin.bdIdxChan = make(chan l2idx.BdChangeDto, 100)
	bdIndexes.WatchNameToIdx(core.PluginName("bdplugin_bdstate"), plugin.bdIdxChan)

	var childCtx context.Context
	childCtx, plugin.cancel = context.WithCancel(ctx)

	// Bridge domain notification watcher
	go plugin.watchVPPNotifications(childCtx)

	return nil
}

// watchVPPNotifications watches for delivery of notifications from VPP.
func (plugin *BridgeDomainStateUpdater) watchVPPNotifications(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	if plugin.notificationChan != nil {
		plugin.log.Info("watchVPPNotifications for bridge domain state started")
	} else {
		plugin.log.Error("failed to start watchVPPNotifications for bridge domain state")
		return
	}

	for {
		select {
		case notif, ok := <-plugin.notificationChan:
			if !ok {
				continue
			}
			bdName := notif.Name
			switch msg := notif.Message.(type) {
			case *l2_api.BridgeDomainDetails:
				bdState := plugin.processBridgeDomainDetailsNotification(msg, bdName)
				if bdState != nil {
					plugin.publishBdState(&BridgeDomainStateNotification{
						State: bdState,
					})
				}
			default:
				plugin.log.Debugf("L2Plugin: Ignoring unknown VPP notification: %v", msg)
			}

		case bdIdxDto := <-plugin.bdIdxChan:
			bdIdxDto.Done()

		case <-ctx.Done():
			// Stop watching for notifications.
			return
		}
	}
}

func (plugin *BridgeDomainStateUpdater) processBridgeDomainDetailsNotification(msg *l2_api.BridgeDomainDetails, name string) *l2.BridgeDomainState_BridgeDomain {
	bdState := &l2.BridgeDomainState_BridgeDomain{}
	// Delete case.
	if msg.BdID == 0 {
		if name == "" {
			plugin.log.Debugf("invalid bridge domain received: %+v", msg)
			return bdState
		}
		// Mark index to 0 to be removed, but pass name so that the key can be constructed.
		bdState.Index = 0
		bdState.InternalName = name
		return bdState
	}
	bdState.Index = msg.BdID
	name, _, found := plugin.bdIndexes.LookupName(msg.BdID)
	if !found {
		plugin.log.Warnf("bridge domain index not found, index %v is not in the mapping", msg.BdID)
		return bdState
	}
	bdState.InternalName = name
	bdState.InterfaceCount = msg.NSwIfs
	name, _, found = plugin.ifIndexes.LookupName(msg.BviSwIfIndex)
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
	var bdStateInterfaces []*l2.BridgeDomainState_BridgeDomain_Interfaces
	for _, swIfaceDetails := range msg.SwIfDetails {
		bdIfaceState := &l2.BridgeDomainState_BridgeDomain_Interfaces{}
		name, _, found := plugin.ifIndexes.LookupName(swIfaceDetails.SwIfIndex)
		if !found {
			plugin.log.Debugf("Interface name with index %v not found for bridge domain status", swIfaceDetails.SwIfIndex)
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
