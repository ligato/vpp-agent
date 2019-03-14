// Copyright (c) 2019 Cisco and/or its affiliates.
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

package descriptor

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/ligato/cn-infra/logging"
	ifmodel "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/api/models/linux/iptables"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	ifdescriptor "github.com/ligato/vpp-agent/plugins/linux/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/linux/iptablesplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/linux/iptablesplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linux/nsplugin/linuxcalls"
)

const (
	// RuleChainDescriptorName is the name of the descriptor for Linux iptables rule chains.
	RuleChainDescriptorName = "linux-ipt-rulechain-descriptor"

	// dependency labels
	ruleChainInterfaceDep = "interface-exists"

	// minimum number of namespaces to be given to a single Go routine for processing
	// in the Retrieve operation
	minWorkForGoRoutine = 3
)

// A list of non-retriable errors:
var (
	// ErrCustomChainWithoutName is returned when the chain name is not provided for the custom iptables chain.
	ErrCustomChainWithoutName = errors.New("iptables chain of type CUSTOM defined without chain name")

	// ErrInvalidChainForTable is returned when the chain is not valid for the provided table.
	ErrInvalidChainForTable = errors.New("provided chain is not valid for the provided table")

	// ErrDefaultPolicyOnNonFilterRule is returned when a default policy is applied on a table different to FILTER.
	ErrDefaultPolicyOnNonFilterRule = errors.New("iptables default policy can be only applied on FILTER tables")

	// ErrDefaultPolicyOnCustomChain is returned when a default policy is applied on a custom chain, which is not allowed in iptables.
	ErrDefaultPolicyOnCustomChain = errors.New("iptables default policy cannot be applied on custom chains")
)

// RuleChainDescriptor teaches KVScheduler how to configure Linux iptables rule chains.
type RuleChainDescriptor struct {
	log             logging.Logger
	nsPlugin        nsplugin.API
	scheduler       kvs.KVScheduler
	ipTablesHandler linuxcalls.IPTablesAPI

	// parallelization of the Retrieve operation
	goRoutinesCnt int
}

// NewRuleChainDescriptor creates a new instance of the iptables RuleChain descriptor.
func NewRuleChainDescriptor(
	scheduler kvs.KVScheduler, ipTablesHandler linuxcalls.IPTablesAPI, nsPlugin nsplugin.API,
	log logging.PluginLogger, goRoutinesCnt int) *kvs.KVDescriptor {

	descrCtx := &RuleChainDescriptor{
		scheduler:       scheduler,
		ipTablesHandler: ipTablesHandler,
		nsPlugin:        nsPlugin,
		goRoutinesCnt:   goRoutinesCnt,
		log:             log.NewLogger("ipt-rulechain-descriptor"),
	}

	typedDescr := &adapter.RuleChainDescriptor{
		Name:                 RuleChainDescriptorName,
		NBKeyPrefix:          linux_iptables.ModelRuleChain.KeyPrefix(),
		ValueTypeName:        linux_iptables.ModelRuleChain.ProtoName(),
		KeySelector:          linux_iptables.ModelRuleChain.IsKeyValid,
		KeyLabel:             linux_iptables.ModelRuleChain.StripKeyPrefix,
		ValueComparator:      descrCtx.EquivalentRuleChains,
		Validate:             descrCtx.Validate,
		Create:               descrCtx.Create,
		Delete:               descrCtx.Delete,
		Retrieve:             descrCtx.Retrieve,
		Dependencies:         descrCtx.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewRuleChainDescriptor(typedDescr)
}

// EquivalentRuleChains is a comparison function for two RuleChain entries.
func (d *RuleChainDescriptor) EquivalentRuleChains(key string, oldRCh, newRch *linux_iptables.RuleChain) bool {

	// first, compare everything except the rules
	oldRules := oldRCh.Rules
	newRules := newRch.Rules

	oldRCh.Rules = nil
	newRch.Rules = nil
	defer func() {
		oldRCh.Rules = oldRules
		newRch.Rules = newRules
	}()

	if !proto.Equal(oldRCh, newRch) {
		return false
	}

	// compare rule count
	if len(oldRules) != len(newRules) {
		return false
	}

	// compare individual rules one by one
	// note that the rules can have individual parts reordered, e.g. the rule
	// "-i eth0 -s 192.168.0.1 -j ACCEPT" is equivalent to
	// "-s 192.168.0.1 -i eth0 -j ACCEPT"

	for i := range oldRules {
		// tokenize the matching rules based on space separator
		oldTokens := strings.Split(oldRules[i], " ")
		newTokens := strings.Split(newRules[i], " ")
		// compare token counts first
		if len(oldTokens) != len(newTokens) {
			return false
		}
		// check if each token exists in the matching rule
		for j := range oldTokens {
			if !sliceContains(newTokens, oldTokens[j]) {
				return false
			}
		}
	}

	return true
}

// Validate validates iptables rule chain.
func (d *RuleChainDescriptor) Validate(key string, rch *linux_iptables.RuleChain) (err error) {
	if rch.ChainType == linux_iptables.RuleChain_CUSTOM && rch.ChainName == "" {
		return kvs.NewInvalidValueError(ErrCustomChainWithoutName, "chain_name")
	}
	if !isAllowedChain(rch.Table, rch.ChainType) {
		return kvs.NewInvalidValueError(ErrInvalidChainForTable, "chain_type")
	}
	if rch.Table != linux_iptables.RuleChain_FILTER && rch.DefaultPolicy != linux_iptables.RuleChain_NONE {
		return kvs.NewInvalidValueError(ErrDefaultPolicyOnNonFilterRule, "default_policy")
	}
	if rch.ChainType == linux_iptables.RuleChain_CUSTOM && rch.DefaultPolicy != linux_iptables.RuleChain_NONE {
		return kvs.NewInvalidValueError(ErrDefaultPolicyOnCustomChain, "default_policy")
	}
	return nil
}

// Create creates iptables rule chain.
func (d *RuleChainDescriptor) Create(key string, rch *linux_iptables.RuleChain) (metadata interface{}, err error) {

	d.log.Debugf("CREATE IPT rule chain %s: %v", key, rch)

	// switch network namespace
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	nsRevert, err := d.nsPlugin.SwitchToNamespace(nsCtx, rch.Namespace)
	if err != nil {
		d.log.WithFields(logging.Fields{
			"err":       err,
			"namespace": rch.Namespace,
		}).Warn("Failed to switch the namespace")
		return nil, err
	}
	// revert network namespace after returning
	defer nsRevert()

	// create custom chain if needed
	if rch.ChainType == linux_iptables.RuleChain_CUSTOM {
		err := d.ipTablesHandler.CreateChain(protocolType(rch), tableNameStr(rch), chainNameStr(rch))
		if err != nil {
			d.log.Warnf("Error by creating iptables chain: %v", err)
			// try to continue, the chain may already exist
		}
	}

	// for FILTER tables, change the default policy if it is set
	if rch.Table == linux_iptables.RuleChain_FILTER && rch.DefaultPolicy != linux_iptables.RuleChain_NONE {
		err = d.ipTablesHandler.SetChainDefaultPolicy(protocolType(rch), tableNameStr(rch), chainNameStr(rch), chainPolicyStr(rch))
		if err != nil {
			d.log.Errorf("Error by setting iptables default policy: %v", err)
			return nil, err
		}
	}

	// wipe all rules in the chain that may have existed before
	err = d.ipTablesHandler.DeleteAllRules(protocolType(rch), tableNameStr(rch), chainNameStr(rch))
	if err != nil {
		d.log.Warnf("Error by wiping iptables rules: %v", err)
	}

	// append all rules
	for _, rule := range rch.Rules {
		err := d.ipTablesHandler.AppendRule(protocolType(rch), tableNameStr(rch), chainNameStr(rch), rule)
		if err != nil {
			d.log.Errorf("Error by appending iptables rule: %v", err)
			break
		}
	}

	return nil, err
}

// Delete removes iptables rule chain.
func (d *RuleChainDescriptor) Delete(key string, rch *linux_iptables.RuleChain, metadata interface{}) error {

	d.log.Debugf("DELETE IPT rule chain %s: %v", key, rch)

	// switch network namespace
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	nsRevert, err := d.nsPlugin.SwitchToNamespace(nsCtx, rch.Namespace)
	if err != nil {
		d.log.WithFields(logging.Fields{
			"err":       err,
			"namespace": rch.Namespace,
		}).Warn("Failed to switch the namespace")
		return err
	}
	// revert network namespace after returning
	defer nsRevert()

	// delete all rules in the chain
	err = d.ipTablesHandler.DeleteAllRules(protocolType(rch), tableNameStr(rch), chainNameStr(rch))
	if err != nil {
		d.log.Errorf("Error by deleting iptables rules: %v", err)
	}

	// delete the chain if it was custom-defined
	if rch.ChainType == linux_iptables.RuleChain_CUSTOM {
		err := d.ipTablesHandler.DeleteChain(protocolType(rch), tableNameStr(rch), chainNameStr(rch))
		if err != nil {
			d.log.Errorf("Error by deleting iptables chain: %v", err)
			return err
		}
	}

	return nil
}

// Dependencies lists dependencies for a iptables rule chain.
func (d *RuleChainDescriptor) Dependencies(key string, rch *linux_iptables.RuleChain) []kvs.Dependency {
	if len(rch.Interfaces) > 0 {
		// the associated interfaces must exist
		var deps []kvs.Dependency
		for _, i := range rch.Interfaces {
			deps = append(deps, kvs.Dependency{
				Label: ruleChainInterfaceDep + "-" + i,
				Key:   ifmodel.InterfaceKey(i),
			})
		}
		return deps
	}
	return nil
}

// retrievedRuleChains is used as the return value sent via channel by retrieveRuleChains().
type retrievedRuleChains struct {
	chains []adapter.RuleChainKVWithMetadata
	err    error
}

// Retrieve returns all iptables rule chain entries managed by this agent.
func (d *RuleChainDescriptor) Retrieve(correlate []adapter.RuleChainKVWithMetadata) ([]adapter.RuleChainKVWithMetadata, error) {
	var values []adapter.RuleChainKVWithMetadata

	goRoutinesCnt := len(correlate) / minWorkForGoRoutine
	if goRoutinesCnt == 0 {
		goRoutinesCnt = 1
	}
	if goRoutinesCnt > d.goRoutinesCnt {
		goRoutinesCnt = d.goRoutinesCnt
	}
	ch := make(chan retrievedRuleChains, goRoutinesCnt)

	// invoke multiple go routines for more efficient parallel chain retrieval
	for idx := 0; idx < goRoutinesCnt; idx++ {
		if goRoutinesCnt > 1 {
			go d.retrieveRuleChains(correlate, idx, goRoutinesCnt, ch)
		} else {
			d.retrieveRuleChains(correlate, idx, goRoutinesCnt, ch)
		}
	}

	// collect results from the go routines
	for idx := 0; idx < goRoutinesCnt; idx++ {
		retrieved := <-ch
		if retrieved.err != nil {
			return values, retrieved.err
		}
		values = append(values, retrieved.chains...)
	}

	return values, nil
}

// retrieveRuleChains is run by a separate go routine to retrieve all iptables rule chains associated
// with every <goRoutineIdx>-th correlation input.
func (d *RuleChainDescriptor) retrieveRuleChains(
	correlate []adapter.RuleChainKVWithMetadata, goRoutineIdx, goRoutinesCnt int, ch chan<- retrievedRuleChains) {

	var retrieved retrievedRuleChains
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()

	for i := goRoutineIdx; i < len(correlate); i += goRoutinesCnt {
		corrrelRule := correlate[i].Value

		// switch to the namespace
		nsRevert, err := d.nsPlugin.SwitchToNamespace(nsCtx, corrrelRule.Namespace)
		if err != nil {
			d.log.WithFields(logging.Fields{
				"err":       err,
				"namespace": corrrelRule.Namespace,
			}).Warn("Failed to switch the namespace")
			continue // continue with the item
		}

		// TODO: we are not able to dump the default policy of a chain

		// list rules in provided table & chain
		rules, err := d.ipTablesHandler.ListRules(protocolType(corrrelRule), tableNameStr(corrrelRule), chainNameStr(corrrelRule))
		if err != nil {
			d.log.Warnf("Error by listing iptables rules: %v", err)
			continue // continue with the item
		}

		// build key-value pair for the retrieved rules
		val := proto.Clone(corrrelRule).(*linux_iptables.RuleChain)
		val.Rules = rules
		retrieved.chains = append(retrieved.chains, adapter.RuleChainKVWithMetadata{
			Key:    linux_iptables.RuleChainKey(val.Name),
			Value:  val,
			Origin: kvs.FromNB,
		})

		// switch back to the default namespace
		nsRevert()
	}

	ch <- retrieved
}

// sliceContains returns true if provided slice contains provided value, false otherwise.
func sliceContains(slice []string, value string) bool {
	for _, i := range slice {
		if i == value {
			return true
		}
	}
	return false
}

// isAllowedChain returns true if provided chain is valid for the provided table, false otherwise.
func isAllowedChain(table linux_iptables.RuleChain_Table, chain linux_iptables.RuleChain_ChainType) bool {
	switch table {
	case linux_iptables.RuleChain_FILTER:
		// Input / Forward / Output / Custom
		switch chain {
		case linux_iptables.RuleChain_PREROUTING:
			return false
		case linux_iptables.RuleChain_POSTROUTING:
			return false
		default:
			return true
		}
	case linux_iptables.RuleChain_NAT:
		// Prerouting / Output / Postrouting / Custom
		switch chain {
		case linux_iptables.RuleChain_INPUT:
			return false
		case linux_iptables.RuleChain_FORWARD:
			return false
		default:
			return true
		}
	case linux_iptables.RuleChain_MANGLE:
		// all chains
		return true
	case linux_iptables.RuleChain_RAW:
		// Prerouting / Output / Custom
		switch chain {
		case linux_iptables.RuleChain_INPUT:
			return false
		case linux_iptables.RuleChain_FORWARD:
			return false
		case linux_iptables.RuleChain_POSTROUTING:
			return false
		default:
			return true
		}
	case linux_iptables.RuleChain_SECURITY:
		// Input / Output / Forward
		switch chain {
		case linux_iptables.RuleChain_PREROUTING:
			return false
		case linux_iptables.RuleChain_POSTROUTING:
			return false
		default:
			return true
		}
	}
	return false
}

// protocolType returns protocol of the given rule chain in the NB API format.
func protocolType(rch *linux_iptables.RuleChain) linuxcalls.L3Protocol {
	switch rch.Protocol {
	case linux_iptables.RuleChain_IPv6:
		return linuxcalls.ProtocolIPv6
	default:
		return linuxcalls.ProtocolIPv4
	}
}

// protocolType iptables table name of the given rule chain in the NB API format.
func tableNameStr(rch *linux_iptables.RuleChain) string {
	switch rch.Table {
	case linux_iptables.RuleChain_NAT:
		return "nat"
	case linux_iptables.RuleChain_MANGLE:
		return "mangle"
	case linux_iptables.RuleChain_RAW:
		return "raw"
	case linux_iptables.RuleChain_SECURITY:
		return "security"
	default:
		return "filter"
	}
}

// protocolType iptables chain name of the given rule chain in the NB API format.
func chainNameStr(rch *linux_iptables.RuleChain) string {
	switch rch.ChainType {
	case linux_iptables.RuleChain_CUSTOM:
		return rch.ChainName
	case linux_iptables.RuleChain_OUTPUT:
		return "OUTPUT"
	case linux_iptables.RuleChain_FORWARD:
		return "FORWARD"
	case linux_iptables.RuleChain_PREROUTING:
		return "PREROUTING"
	case linux_iptables.RuleChain_POSTROUTING:
		return "POSTROUTING"
	default:
		return "INPUT"
	}
}

// protocolType iptables policy name of the given rule chain in the NB API format.
func chainPolicyStr(rch *linux_iptables.RuleChain) string {
	switch rch.DefaultPolicy {
	case linux_iptables.RuleChain_DROP:
		return "DROP"
	case linux_iptables.RuleChain_QUEUE:
		return "QUEUE"
	case linux_iptables.RuleChain_RETURN:
		return "RETURN"
	default:
		return "ACCEPT"
	}
}
