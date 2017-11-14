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
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	model_stn "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// StnConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/stn/stn.proto"
// and stored in ETCD under the key "vpp/config/v1/stn/rules/".
type StnConfigurator struct {
	Log logging.Logger

	GoVppmux    govppmux.API
	StnIndexes  idxvpp.NameToIdxRW
	StnIndexSeq uint32
	SwIfIndexes ifaceidx.SwIfIndex
	vppChan     *govppapi.Channel

	Stopwatch *measure.Stopwatch
}

// Init initializes ARP configurator
func (plugin *StnConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing StnConfigurator")

	// Init VPP API channel
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	return plugin.checkMsgCompatibility()
}

// CheckStn rule raw data
func CheckStn(stnInput *model_stn.StnRule, index ifaceidx.SwIfIndex, log logging.Logger) (*vppcalls.StnRule, error) {
	if stnInput == nil {
		return nil, fmt.Errorf("STN input is empty")
	}
	if stnInput.Interface == "" {
		return nil, fmt.Errorf("STN input does not contain interface")
	}
	if stnInput.IpAddress == "" {
		return nil, fmt.Errorf("STN input does not contain IP")
	}

	ifName := stnInput.Interface
	ifIndex, _, exists := index.LookupIdx(ifName)
	if !exists {
		return nil, fmt.Errorf("STN entry interface %v not found", ifName)
	}

	parsedIP, _, err := addrs.ParseIPWithPrefix(stnInput.IpAddress)
	if err != nil {
		return nil, err
	}

	stnRule := &vppcalls.StnRule{
		IPAddress: *parsedIP,
		IfaceIdx:  ifIndex,
	}
	return stnRule, nil
}

// Add create a new STN rule
func (plugin *StnConfigurator) Add(rule *model_stn.StnRule) error {
	plugin.Log.Infof("Creating new STN rule %v", rule)

	// Check stn data
	stnRule, err := CheckStn(rule, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	plugin.Log.Debugf("adding STN rule: %+v", stnRule)
	// Create and register new stn
	errVppCall := vppcalls.AddStnRule(stnRule.IfaceIdx, &stnRule.IPAddress, plugin.Log, plugin.vppChan, measure.GetTimeLog(stn.StnAddDelRule{}, plugin.Stopwatch))
	if errVppCall != nil {
		return errVppCall
	}
	stnID := stnIdentifier(stnRule.IfaceIdx, stnRule.IPAddress.String())
	plugin.StnIndexes.RegisterName(stnID, plugin.StnIndexSeq, nil)
	plugin.StnIndexSeq++
	plugin.Log.Infof("STN entry %v registered", stnID)

	return nil
}

// Delete removes STN rule
func (plugin *StnConfigurator) Delete(rule *model_stn.StnRule) error {
	plugin.Log.Infof("Removing rule on if: %v with IP: %v", rule.Interface, rule.IpAddress)
	// Check stn data
	stnRule, err := CheckStn(rule, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	if stnRule == nil {
		return nil
	}
	plugin.Log.Debugf("deleting stn rule: %+v", stnRule)
	// Remove and unregister route
	err = vppcalls.DelStnRule(stnRule.IfaceIdx, &stnRule.IPAddress, plugin.Log, plugin.vppChan, measure.GetTimeLog(stn.StnAddDelRule{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	stnID := stnIdentifier(stnRule.IfaceIdx, stnRule.IPAddress.String())
	_, _, found := plugin.StnIndexes.UnregisterName(stnID)
	if found {
		plugin.Log.Infof("STN rule %v unregistered", stnID)
	} else {
		plugin.Log.Warnf("Unregister failed, STN rule %v not found", stnID)
	}

	return nil
}

// Modify changes the stored rules
func (plugin *StnConfigurator) Modify(rule *model_stn.StnRule, rule2 *model_stn.StnRule) error {
	//TODO: Need to be implemented
	return nil
}

// Close GOVPP channel
func (plugin *StnConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// Creates unique identifier which serves as a name in name to index mapping
func stnIdentifier(iface uint32, ip string) string {
	return fmt.Sprintf("stn-iface-%v-%v", iface, ip)
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
