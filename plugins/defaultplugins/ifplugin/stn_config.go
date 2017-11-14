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

//go:generate protoc --proto_path=model/stn --gogo_out=model/stn model/stn/stn.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/stn.api.json --output-dir=bin_api

package ifplugin

import (
	"fmt"

	"context"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	modelStn "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// StnConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/stn/stn.proto"
// and stored in ETCD under the key "vpp/config/v1/stn/rules/".
type StnConfigurator struct {
	Log logging.Logger

	GoVppmux    govppmux.API
	SwIfIndexes ifaceidx.SwIfIndex
	stnIdx      StnIndexes
	vppChan     *govppapi.Channel
	swIdxChan   chan ifaceidx.SwIfIdxDto

	cancel    context.CancelFunc
	Stopwatch *measure.Stopwatch
}

type StnIndexes struct {
	StnAllIndexes       idxvpp.NameToIdxRW
	StnAllIndexSeq      uint32
	StnUnstoredIndexes  idxvpp.NameToIdxRW
	StnUnstoredIndexSeq uint32
}

// Init initializes ARP configurator
func (plugin *StnConfigurator) Init(ctx context.Context) (err error) {
	plugin.Log.Debug("Initializing StnConfigurator")

	// Init VPP API channel
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	errCompatibility := plugin.checkMsgCompatibility()
	if errCompatibility != nil {
		return errCompatibility
	}

	// Run consumer
	go plugin.consume(ctx)

	return nil

}

func (plugin *StnConfigurator) consume(ctx context.Context) {
	swIfIdxChan := make(chan ifaceidx.SwIfIdxDto)
	plugin.SwIfIndexes.WatchNameToIdx(core.PluginName("ifplugin_stn"), swIfIdxChan)
	// create child context
	var childCtx context.Context
	childCtx, plugin.cancel = context.WithCancel(ctx)

	for {
		select {
		case swIdxDto := <-plugin.swIdxChan:
			if swIdxDto.Del {
				withoutIface, rule := removeRuleFromIndex(swIdxDto.Idx, plugin.stnIdx)
				if !withoutIface {
					err := plugin.Delete(rule)
					plugin.Log.Debug(err)
				}
			} else {
				plugin.Add(ruleFromIndex(swIdxDto.Idx, plugin.stnIdx))
			}
			swIdxDto.Done()

		case <-childCtx.Done():
			// stop watching for notifications
			return
		}
	}
}

// checkStn will check the rule raw data and change it to internal data structure.
// In case the rule contains a interface that doesn't exist yet, rule is stored into index map.
func (plugin *StnConfigurator) checkStn(stnInput *modelStn.StnRule, index ifaceidx.SwIfIndex) (stnRule *vppcalls.StnRule, doVPPCall bool, err error) {

	stnRule = nil
	doVPPCall = false

	if stnInput == nil {
		err = fmt.Errorf("STN input is empty")
		return
	}
	if stnInput.Interface == "" {
		err = fmt.Errorf("STN input does not contain interface")
		return
	}
	if stnInput.IpAddress == "" {
		err = fmt.Errorf("STN input does not contain IP")
		return
	}

	parsedIP, _, err := addrs.ParseIPWithPrefix(stnInput.IpAddress)
	if err != nil {
		return
	}

	ifName := stnInput.Interface
	ifIndex, _, exists := index.LookupIdx(ifName)

	if !exists {
		stnRule = &vppcalls.StnRule{
			IPAddress: *parsedIP,
			IfaceIdx:  ifIndex,
		}
		return
	}

	stnRule = &vppcalls.StnRule{
		IPAddress: *parsedIP,
		IfaceIdx:  ifIndex,
	}
	doVPPCall = true
	return
}

func storeRuleToIndex(rule *modelStn.StnRule, withoutIface bool, stnIndx StnIndexes, id uint32) {
	idx := stnIdentifier(id)
	if withoutIface {
		stnIndx.StnUnstoredIndexes.RegisterName(idx, stnIndx.StnUnstoredIndexSeq, rule)
		stnIndx.StnUnstoredIndexSeq++
	}
	stnIndx.StnAllIndexes.RegisterName(idx, stnIndx.StnAllIndexSeq, rule)
	stnIndx.StnAllIndexSeq++
}

func removeRuleFromIndex(id uint32, stnIndx StnIndexes)(withoutIface bool, rule *modelStn.StnRule) {
	idx := stnIdentifier(id)
	rule = nil
	withoutIface = false

	_, ruleIface, exists := stnIndx.StnAllIndexes.LookupIdx(idx)
	if exists {
		stnIndx.StnAllIndexes.UnregisterName(idx)
		stnRule, ok := ruleIface.(*modelStn.StnRule)
		if ok {
			rule = stnRule
		}
	}

	_, _, existsWithout := stnIndx.StnUnstoredIndexes.LookupIdx(idx)
	if existsWithout {
		withoutIface = true
		stnIndx.StnUnstoredIndexes.UnregisterName(idx)
	}

	return
}

func ruleFromIndex(id uint32, stnIndx StnIndexes) (rule *modelStn.StnRule) {
	idx := stnIdentifier(id)
	rule = nil
	_, ruleIface, exists := stnIndx.StnAllIndexes.LookupIdx(idx)
	if exists {
		stnRule, ok := ruleIface.(*modelStn.StnRule)
		if ok {
			rule = stnRule
		}
	}
	return
}

// Add create a new STN rule
func (plugin *StnConfigurator) Add(rule *modelStn.StnRule) error {
	plugin.Log.Infof("Creating new STN rule %v", rule)

	// Check stn data
	stnRule, doVPPCall, err := plugin.checkStn(rule, plugin.SwIfIndexes)
	if err != nil {
		return err
	}
	if !doVPPCall {
		plugin.Log.Infof("There is no interface for rule: %+v. Rule will be stored and it will wait for appropriate interface.", stnRule)
		storeRuleToIndex(rule, true, plugin.stnIdx, stnRule.IfaceIdx)
	} else {
		plugin.Log.Debugf("adding STN rule: %+v", stnRule)
		// Create and register new stn
		errVppCall := vppcalls.AddStnRule(stnRule.IfaceIdx, &stnRule.IPAddress, plugin.Log, plugin.vppChan, measure.GetTimeLog(stn.StnAddDelRule{}, plugin.Stopwatch))
		if errVppCall != nil {
			return errVppCall
		}
		storeRuleToIndex(rule, false, plugin.stnIdx, stnRule.IfaceIdx)
	}

	return nil
}

// Delete removes STN rule
func (plugin *StnConfigurator) Delete(rule *modelStn.StnRule) error {
	plugin.Log.Infof("Removing rule on if: %v with IP: %v", rule.Interface, rule.IpAddress)
	// Check stn data
	stnRule, doVPPCall, err := plugin.checkStn(rule, plugin.SwIfIndexes)

	if err != nil {
		return err
	}
	if stnRule == nil {
		return nil
	}

	removeRuleFromIndex(stnRule.IfaceIdx, plugin.stnIdx)

	if !doVPPCall {
		return nil
	} else {
		plugin.Log.Debugf("deleting stn rule: %+v", stnRule)
		// Remove rule
		return vppcalls.DelStnRule(stnRule.IfaceIdx, &stnRule.IPAddress, plugin.Log, plugin.vppChan, measure.GetTimeLog(stn.StnAddDelRule{}, plugin.Stopwatch))
	}
}

// Modify changes the stored rules
func (plugin *StnConfigurator) Modify(ruleOld *modelStn.StnRule, ruleNew *modelStn.StnRule) error {

	if ruleOld == nil {
		return fmt.Errorf("old stn rule is null")
	}

	if ruleNew == nil {
		return fmt.Errorf("new stn rule is null")
	}

	err := plugin.Delete(ruleOld)
	if err != nil {
		return err
	} else {
		return plugin.Add(ruleNew)
	}

}

// Close GOVPP channel
func (plugin *StnConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// Creates unique identifier which serves as a name in name to index mapping
func stnIdentifier(iface uint32) string {
	return fmt.Sprintf("stn-iface-%v", iface)
}

func (plugin *StnConfigurator) checkMsgCompatibility() error {
	msgs := []govppapi.Message{
		&stn.StnAddDelRule{},
		&stn.StnAddDelRuleReply{},
	}
	err := plugin.vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		plugin.Log.Error(err)
	}
	return err
}
