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

//go:generate protoc --proto_path=../common/model/stn --gogo_out=../common/model/stn ../common/model/stn/stn.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/stn.api.json --output-dir=../common/bin_api

package ifplugin

import (
	"fmt"

	"context"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/stn"
	modelStn "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// StnConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/stn/stn.proto"
// and stored in ETCD under the key "vpp/config/v1/stn/rules/".
type StnConfigurator struct {
	Log logging.Logger

	GoVppmux            govppmux.API
	SwIfIndexes         ifaceidx.SwIfIndex
	StnAllIndexes       idxvpp.NameToIdxRW
	StnAllIndexSeq      uint32
	StnUnstoredIndexes  idxvpp.NameToIdxRW
	StnUnstoredIndexSeq uint32
	vppChan             *govppapi.Channel

	cancel    context.CancelFunc
	Stopwatch *measure.Stopwatch
}

// Init initializes ARP configurator
func (plugin *StnConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing STN configurator")

	// Init VPP API channel
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	errCompatibility := plugin.checkMsgCompatibility()
	if errCompatibility != nil {
		return errCompatibility
	}

	return nil

}

// ResolveDeletedInterface resolves when interface is deleted. If there exist a rule for this interface
// the rule will be deleted also.
func (plugin *StnConfigurator) ResolveDeletedInterface(interfaceName string) {
	plugin.Log.Debugf("STN plugin: resolving deleted interface: %v", interfaceName)
	rule := plugin.ruleFromIndex(interfaceName, true)
	if rule != nil {
		plugin.Delete(rule)
	}
}

// ResolveCreatedInterface will check rules and if there is one waiting for interfaces it will be written
// into VPP.
func (plugin *StnConfigurator) ResolveCreatedInterface(interfaceName string) {
	plugin.Log.Debugf("STN plugin: resolving created interface: %v", interfaceName)
	rule := plugin.ruleFromIndex(interfaceName, false)
	if rule != nil {
		plugin.Add(rule)
	}
}

// Add create a new STN rule.
func (plugin *StnConfigurator) Add(rule *modelStn.StnRule) error {
	plugin.Log.Infof("Configuring new STN rule %v", rule)

	// Check stn data
	stnRule, doVPPCall, err := plugin.checkStn(rule, plugin.SwIfIndexes)
	if err != nil {
		return err
	}
	if !doVPPCall {
		plugin.Log.Warnf("There is no interface for rule: %+v. Waiting for interface.", rule.Interface)
		plugin.indexSTNRule(rule, true)
	} else {
		plugin.Log.Debugf("adding STN rule: %+v", rule)
		// Create and register new stn
		errVppCall := vppcalls.AddStnRule(stnRule.IfaceIdx, &stnRule.IPAddress, plugin.Log, plugin.vppChan, measure.GetTimeLog(stn.StnAddDelRule{}, plugin.Stopwatch))
		if errVppCall != nil {
			return errVppCall
		}
		plugin.indexSTNRule(rule, false)
	}

	plugin.Log.Infof("STN rule %v configured", rule)

	return nil
}

// Delete removes STN rule.
func (plugin *StnConfigurator) Delete(rule *modelStn.StnRule) error {
	plugin.Log.Infof("Removing STN rule on if: %v with IP: %v", rule.Interface, rule.IpAddress)
	// Check stn data
	stnRule, _, err := plugin.checkStn(rule, plugin.SwIfIndexes)

	if err != nil {
		return err
	}
	if stnRule == nil {
		return nil
	}

	withoutIf, _ := plugin.removeRuleFromIndex(rule.Interface)

	if withoutIf {
		plugin.Log.Debug("STN rule was not stored into VPP, removed only from indexes.")
		return nil
	}
	plugin.Log.Debugf("STN rule: %+v was stored in VPP, trying to delete it. %+v", stnRule)
	// Remove rule
	if err := vppcalls.DelStnRule(stnRule.IfaceIdx, &stnRule.IPAddress, plugin.Log, plugin.vppChan, measure.GetTimeLog(stn.StnAddDelRule{}, plugin.Stopwatch)); err != nil {
		return err
	}

	plugin.Log.Infof("STN rule %v removed", rule)

	return nil
}

// Modify configured rule.
func (plugin *StnConfigurator) Modify(ruleOld *modelStn.StnRule, ruleNew *modelStn.StnRule) error {
	plugin.Log.Infof("Modifying STN %v", ruleNew)

	if ruleOld == nil {
		return fmt.Errorf("old stn rule is null")
	}

	if ruleNew == nil {
		return fmt.Errorf("new stn rule is null")
	}

	err := plugin.Delete(ruleOld)
	if err != nil {
		return err
	}

	if err := plugin.Add(ruleNew); err != nil {
		return err
	}

	plugin.Log.Infof("STN rule %v modified", ruleNew)

	return nil
}

// Close GOVPP channel.
func (plugin *StnConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// checkStn will check the rule raw data and change it to internal data structure.
// In case the rule contains a interface that doesn't exist yet, rule is stored into index map.
func (plugin *StnConfigurator) checkStn(stnInput *modelStn.StnRule, index ifaceidx.SwIfIndex) (stnRule *vppcalls.StnRule, doVPPCall bool, err error) {

	plugin.Log.Debugf("Checking stn rule: %+v", stnInput)

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

	stnRule = &vppcalls.StnRule{
		IPAddress: *parsedIP,
		IfaceIdx:  ifIndex,
	}

	if exists {
		doVPPCall = true
	}

	return
}

func (plugin *StnConfigurator) indexSTNRule(rule *modelStn.StnRule, withoutIface bool) {
	idx := stnIdentifier(rule.Interface)
	if withoutIface {
		plugin.StnUnstoredIndexes.RegisterName(idx, plugin.StnUnstoredIndexSeq, rule)
		plugin.StnUnstoredIndexSeq++
	}
	plugin.StnAllIndexes.RegisterName(idx, plugin.StnAllIndexSeq, rule)
	plugin.StnAllIndexSeq++
}

func (plugin *StnConfigurator) removeRuleFromIndex(iface string) (withoutIface bool, rule *modelStn.StnRule) {
	idx := stnIdentifier(iface)
	rule = nil
	withoutIface = false

	//Removing rule from main index
	_, ruleIface, exists := plugin.StnAllIndexes.LookupIdx(idx)
	if exists {
		plugin.StnAllIndexes.UnregisterName(idx)
		stnRule, ok := ruleIface.(*modelStn.StnRule)
		if ok {
			rule = stnRule
		}
	}

	//Removing rule from not stored rules index
	_, _, existsWithout := plugin.StnUnstoredIndexes.LookupIdx(idx)
	if existsWithout {
		withoutIface = true
		plugin.StnUnstoredIndexes.UnregisterName(idx)
	}

	return
}

func (plugin *StnConfigurator) ruleFromIndex(iface string, fromAllRules bool) (rule *modelStn.StnRule) {
	idx := stnIdentifier(iface)
	rule = nil

	var ruleIface interface{}
	var exists bool

	if !fromAllRules {
		_, ruleIface, exists = plugin.StnUnstoredIndexes.LookupIdx(idx)
	} else {
		_, ruleIface, exists = plugin.StnAllIndexes.LookupIdx(idx)
	}
	plugin.Log.Debugf("Rule exists: %+v returned rule: %+v", exists, &ruleIface)
	if exists {
		stnRule, ok := ruleIface.(*modelStn.StnRule)
		plugin.Log.Debugf("Getting rule: %+v", stnRule)
		if ok {
			rule = stnRule
		}
	}
	return
}

// Creates unique identifier which serves as a name in name to index mapping
func stnIdentifier(iface string) string {
	id := fmt.Sprintf("stn-iface-%v", iface)
	return id
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
