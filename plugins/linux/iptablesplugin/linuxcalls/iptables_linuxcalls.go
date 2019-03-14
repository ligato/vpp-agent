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

package linuxcalls

import (
	"fmt"
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

const (
	// prefix of the "append" operation on a rule
	appendRulePrefix = "-A"

	// prefix of a "new chain" rule
	newChainRulePrefix = "-N"
)

// IPTablesHandler is a handler for all operations on Linux iptables / ip6tables.
type IPTablesHandler struct {
	v4Handler *iptables.IPTables
	v6Handler *iptables.IPTables
}

// Init initializes an iptables handler.
func (h *IPTablesHandler) Init() error {
	var err error

	h.v4Handler, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		err = fmt.Errorf("errr by initializing iptables v4 handler: %v", err)
		// continue, iptables just may not be installed
	}

	h.v6Handler, err = iptables.NewWithProtocol(iptables.ProtocolIPv6)
	if err != nil {
		err = fmt.Errorf("errr by initializing iptables v6 handler: %v", err)
		// continue, ip6tables just may not be installed
	}

	return err
}

// CreateChain creates an iptables chain in the specified table.
func (h *IPTablesHandler) CreateChain(protocol L3Protocol, table, chain string) error {
	handler, err := h.getHandler(protocol)
	if err != nil {
		return err
	}
	return handler.NewChain(table, chain)
}

// DeleteChain deletes an iptables chain in the specified table.
func (h *IPTablesHandler) DeleteChain(protocol L3Protocol, table, chain string) error {
	handler, err := h.getHandler(protocol)
	if err != nil {
		return err
	}
	return handler.DeleteChain(table, chain)
}

// SetChainDefaultPolicy sets default policy in the specified chain. Should be called only on FILTER tables.
func (h *IPTablesHandler) SetChainDefaultPolicy(protocol L3Protocol, table, chain, defaultPolicy string) error {
	handler, err := h.getHandler(protocol)
	if err != nil {
		return err
	}
	return handler.ChangePolicy(table, chain, defaultPolicy)
}

// AppendRule appends a rule into the specified chain.
func (h *IPTablesHandler) AppendRule(protocol L3Protocol, table, chain string, rule string) error {
	handler, err := h.getHandler(protocol)
	if err != nil {
		return err
	}
	ruleSlice := strings.Split(rule, " ")

	return handler.Append(table, chain, ruleSlice[:]...)
}

// DeleteRule deletes a rule from the specified chain.
func (h *IPTablesHandler) DeleteRule(protocol L3Protocol, table, chain string, rule string) error {
	handler, err := h.getHandler(protocol)
	if err != nil {
		return err
	}
	ruleSlice := strings.Split(rule, " ")

	return handler.Delete(table, chain, ruleSlice[:]...)
}

// DeleteAllRules deletes all rules within the specified chain.
func (h *IPTablesHandler) DeleteAllRules(protocol L3Protocol, table, chain string) error {
	handler, err := h.getHandler(protocol)
	if err != nil {
		return err
	}
	return handler.ClearChain(table, chain)
}

// ListRules lists all rules within the specified chain.
func (h *IPTablesHandler) ListRules(protocol L3Protocol, table, chain string) (rules []string, err error) {
	handler, err := h.getHandler(protocol)
	if err != nil {
		return nil, err
	}
	dumpRules, err := handler.List(table, chain)

	// post-process & filter rules
	for _, rule := range dumpRules {
		if strings.HasPrefix(rule, newChainRulePrefix) {
			// ignore "new chain" rules
			continue
		}
		if strings.HasPrefix(rule, appendRulePrefix) {
			// trim "-A <CHAIN-NAME>" part
			rule = strings.TrimPrefix(rule, fmt.Sprintf("%s %s", appendRulePrefix, chain))
		}
		rules = append(rules, strings.TrimSpace(rule))
	}

	return
}

// getHandler returns the iptables handler for the given protocol.
// returns an error if the requested handler is not initialized.
func (h *IPTablesHandler) getHandler(protocol L3Protocol) (*iptables.IPTables, error) {
	var handler *iptables.IPTables

	if protocol == ProtocolIPv4 {
		handler = h.v4Handler
	} else {
		handler = h.v6Handler
	}

	if handler == nil {
		return nil, fmt.Errorf("iptables handler for protocol %v is not initialized", protocol)
	}
	return handler, nil
}
