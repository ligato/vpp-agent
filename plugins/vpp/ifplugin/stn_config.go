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

//go:generate protoc --proto_path=../model/stn --gogo_out=../model/stn ../model/stn/stn.proto

package ifplugin

import (
	"fmt"
	"net"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	modelStn "github.com/ligato/vpp-agent/plugins/vpp/model/stn"
)

// StnConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/stn/stn.proto"
// and stored in ETCD under the key "vpp/config/v1/stn/rules/".
type StnConfigurator struct {
	log logging.Logger
	// Indexes
	ifIndexes        ifaceidx.SwIfIndex
	allIndexes       idxvpp.NameToIdxRW
	allIndexesSeq    uint32
	unstoredIndexes  idxvpp.NameToIdxRW
	unstoredIndexSeq uint32
	// VPP
	vppChan govppapi.Channel
	// VPP API handler
	stnHandler vppcalls.StnVppAPI
	// Stopwatch
	stopwatch *measure.Stopwatch
}

// IndexExistsFor returns true if there is and mapping entry for provided name
func (plugin *StnConfigurator) IndexExistsFor(name string) bool {
	_, _, found := plugin.allIndexes.LookupIdx(name)
	return found
}

// UnstoredIndexExistsFor returns true if there is and mapping entry for provided name
func (plugin *StnConfigurator) UnstoredIndexExistsFor(name string) bool {
	_, _, found := plugin.unstoredIndexes.LookupIdx(name)
	return found
}

// Init initializes STN configurator
func (plugin *StnConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, ifIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Init logger
	plugin.log = logger.NewLogger("-stn-conf")
	plugin.log.Debug("Initializing STN configurator")

	// Configurator-wide stopwatch instance
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("STN-configurator", plugin.log)
	}

	// Init VPP API channel
	plugin.vppChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Init indexes
	plugin.ifIndexes = ifIndexes
	plugin.allIndexes = nametoidx.NewNameToIdx(plugin.log, "stn-all-indexes", nil)
	plugin.unstoredIndexes = nametoidx.NewNameToIdx(plugin.log, "stn-unstored-indexes", nil)
	plugin.allIndexesSeq, plugin.unstoredIndexSeq = 1, 1

	// VPP API handler
	plugin.stnHandler = vppcalls.NewStnVppHandler(plugin.vppChan, plugin.ifIndexes, plugin.log, plugin.stopwatch)

	return nil
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (plugin *StnConfigurator) clearMapping() {
	plugin.allIndexes.Clear()
	plugin.unstoredIndexes.Clear()
}

// ResolveDeletedInterface resolves when interface is deleted. If there exist a rule for this interface
// the rule will be deleted also.
func (plugin *StnConfigurator) ResolveDeletedInterface(interfaceName string) {
	plugin.log.Debugf("STN plugin: resolving deleted interface: %v", interfaceName)
	if rule := plugin.ruleFromIndex(interfaceName, true); rule != nil {
		plugin.Delete(rule)
	}
}

// ResolveCreatedInterface will check rules and if there is one waiting for interfaces it will be written
// into VPP.
func (plugin *StnConfigurator) ResolveCreatedInterface(interfaceName string) {
	plugin.log.Debugf("STN plugin: resolving created interface: %v", interfaceName)
	if rule := plugin.ruleFromIndex(interfaceName, false); rule != nil {
		if err := plugin.Add(rule); err == nil {
			plugin.unstoredIndexes.UnregisterName(StnIdentifier(interfaceName))
		}
	}
}

// Add create a new STN rule.
func (plugin *StnConfigurator) Add(rule *modelStn.STN_Rule) error {
	plugin.log.Infof("Configuring new STN rule %v", rule)

	// Check stn data
	stnRule, doVPPCall, err := plugin.checkStn(rule, plugin.ifIndexes)
	if err != nil {
		return err
	}
	if !doVPPCall {
		plugin.log.Debugf("There is no interface for rule: %+v. Waiting for interface.", rule.Interface)
		plugin.indexSTNRule(rule, true)
	} else {
		plugin.log.Debugf("adding STN rule: %+v", rule)
		// Create and register new stn
		if err := plugin.stnHandler.AddStnRule(stnRule.IfaceIdx, &stnRule.IPAddress); err != nil {
			return err
		}
		plugin.indexSTNRule(rule, false)

		plugin.log.Infof("STN rule %v configured", rule)
	}

	return nil
}

// Delete removes STN rule.
func (plugin *StnConfigurator) Delete(rule *modelStn.STN_Rule) error {
	plugin.log.Infof("Removing STN rule on if: %v with IP: %v", rule.Interface, rule.IpAddress)
	// Check stn data
	stnRule, _, err := plugin.checkStn(rule, plugin.ifIndexes)

	if err != nil {
		return err
	}

	if withoutIf, _ := plugin.removeRuleFromIndex(rule.Interface); withoutIf {
		plugin.log.Debug("STN rule was not stored into VPP, removed only from indexes.")
		return nil
	}
	plugin.log.Debugf("STN rule: %+v was stored in VPP, trying to delete it. %+v", stnRule)

	// Remove rule
	if err := plugin.stnHandler.DelStnRule(stnRule.IfaceIdx, &stnRule.IPAddress); err != nil {
		return err
	}

	plugin.log.Infof("STN rule %v removed", rule)

	return nil
}

// Modify configured rule.
func (plugin *StnConfigurator) Modify(ruleOld *modelStn.STN_Rule, ruleNew *modelStn.STN_Rule) error {
	plugin.log.Infof("Modifying STN %v", ruleNew)

	if ruleOld == nil {
		return fmt.Errorf("old stn rule is null")
	}

	if ruleNew == nil {
		return fmt.Errorf("new stn rule is null")
	}

	if err := plugin.Delete(ruleOld); err != nil {
		return err
	}

	if err := plugin.Add(ruleNew); err != nil {
		return err
	}

	plugin.log.Infof("STN rule %v modified", ruleNew)

	return nil
}

// Dump STN rules configured on the VPP
func (plugin *StnConfigurator) Dump() (*vppcalls.StnDetails, error) {
	stnDetails, err := plugin.stnHandler.DumpStnRules()
	if err != nil {
		return nil, err
	}
	plugin.log.Debugf("found %d configured STN rules", len(stnDetails.Rules))
	return stnDetails, nil
}

// Close GOVPP channel.
func (plugin *StnConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// checkStn will check the rule raw data and change it to internal data structure.
// In case the rule contains a interface that doesn't exist yet, rule is stored into index map.
func (plugin *StnConfigurator) checkStn(stnInput *modelStn.STN_Rule, index ifaceidx.SwIfIndex) (stnRule *vppcalls.StnRule, doVPPCall bool, err error) {
	plugin.log.Debugf("Checking stn rule: %+v", stnInput)

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

	ipWithMask := strings.Split(stnInput.IpAddress, "/")
	if len(ipWithMask) > 1 {
		plugin.log.Debugf("STN rule %v IP address mask is ignored", stnInput.RuleName)
		stnInput.IpAddress = ipWithMask[0]
	}
	parsedIP := net.ParseIP(stnInput.IpAddress)
	if parsedIP == nil {
		err = fmt.Errorf("unable to parse IP %v", stnInput.IpAddress)
		return
	}

	ifName := stnInput.Interface
	ifIndex, _, exists := index.LookupIdx(ifName)
	if exists {
		doVPPCall = true
	}

	stnRule = &vppcalls.StnRule{
		IPAddress: parsedIP,
		IfaceIdx:  ifIndex,
	}

	return
}

func (plugin *StnConfigurator) indexSTNRule(rule *modelStn.STN_Rule, withoutIface bool) {
	idx := StnIdentifier(rule.Interface)
	if withoutIface {
		plugin.unstoredIndexes.RegisterName(idx, plugin.unstoredIndexSeq, rule)
		plugin.unstoredIndexSeq++
	}
	plugin.allIndexes.RegisterName(idx, plugin.allIndexesSeq, rule)
	plugin.allIndexesSeq++
}

func (plugin *StnConfigurator) removeRuleFromIndex(iface string) (withoutIface bool, rule *modelStn.STN_Rule) {
	idx := StnIdentifier(iface)

	// Removing rule from main index
	_, ruleIface, exists := plugin.allIndexes.LookupIdx(idx)
	if exists {
		plugin.allIndexes.UnregisterName(idx)
		stnRule, ok := ruleIface.(*modelStn.STN_Rule)
		if ok {
			rule = stnRule
		}
	}

	// Removing rule from not stored rules index
	_, _, existsWithout := plugin.unstoredIndexes.LookupIdx(idx)
	if existsWithout {
		withoutIface = true
		plugin.unstoredIndexes.UnregisterName(idx)
	}

	return
}

func (plugin *StnConfigurator) ruleFromIndex(iface string, fromAllRules bool) (rule *modelStn.STN_Rule) {
	idx := StnIdentifier(iface)

	var ruleIface interface{}
	var exists bool

	if !fromAllRules {
		_, ruleIface, exists = plugin.unstoredIndexes.LookupIdx(idx)
	} else {
		_, ruleIface, exists = plugin.allIndexes.LookupIdx(idx)
	}
	plugin.log.Debugf("Rule exists: %+v returned rule: %+v", exists, &ruleIface)
	if exists {
		stnRule, ok := ruleIface.(*modelStn.STN_Rule)
		if ok {
			rule = stnRule
		}
		plugin.log.Debugf("Getting rule: %+v", stnRule)
	}

	return
}

// StnIdentifier creates unique identifier which serves as a name in name to index mapping
func StnIdentifier(iface string) string {
	return fmt.Sprintf("stn-iface-%v", iface)
}
