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
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"github.com/pkg/errors"
)

const (
	// prefix of the "append" operation on a rule
	appendRulePrefix = "-A"

	// prefix of a "new chain" rule
	newChainRulePrefix = "-N"

	// command names
	IPv4SaveCmd    string = "iptables-save"
	IPv4RestoreCmd string = "iptables-restore"
	IPv6RestoreCmd string = "ip6tables-restore"
	IPv6SaveCmd    string = "ip6tables-save"
)

// IPTablesHandler is a handler for all operations on Linux iptables / ip6tables.
type IPTablesHandler struct {
	v4Handler                       *iptables.IPTables
	v6Handler                       *iptables.IPTables
	minRuleCountForPerfRuleAddition int
}

// Init initializes an iptables handler.
func (h *IPTablesHandler) Init(config *HandlerConfig) error {
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

	h.minRuleCountForPerfRuleAddition = config.MinRuleCountForPerfRuleAddition

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

// AppendRules appends rules into the specified chain.
func (h *IPTablesHandler) AppendRules(protocol L3Protocol, table, chain string, rules ...string) error {
	if len(rules) == 0 {
		return nil // nothing to do
	}

	if len(rules) < h.minRuleCountForPerfRuleAddition { // use normal method of addition
		for _, rule := range rules {
			err := h.AppendRule(protocol, table, chain, rule)
			if err != nil {
				return errors.Errorf("Error by appending iptables rule: %v", err)
			}
		}
	} else { // use performance solution (this makes performance difference with higher count of appended rules)
		// export existing iptables data
		data, err := h.saveTable(protocol, table, true)
		if err != nil {
			return errors.Errorf(": Can't export all rules due to: %v", err)
		}

		// add rules to exported data
		insertPoint := bytes.Index(data, []byte("COMMIT"))
		if insertPoint == -1 {
			return errors.Errorf("Error by adding rules: Can't find COMMIT statement in iptables-save data")
		}
		var rulesSB strings.Builder
		for _, rule := range rules {
			rulesSB.WriteString(fmt.Sprintf("[0:0] -A %s %s\n", chain, rule))
		}
		insertData := []byte(rulesSB.String())
		updatedData := make([]byte, len(data)+len(insertData))
		copy(updatedData[:insertPoint], data[:insertPoint])
		copy(updatedData[insertPoint:insertPoint+len(insertData)], insertData)
		copy(updatedData[insertPoint+len(insertData):], data[insertPoint:])

		// import modified data to linux
		err = h.restoreTable(protocol, table, updatedData, true, true)
		if err != nil {
			return errors.Errorf("Error by adding rules: Can't restore modified iptables data due to: %v", err)
		}
	}

	return nil
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

// saveTable exports all data for given table in IPTable-save output format
func (h *IPTablesHandler) saveTable(protocol L3Protocol, table string, exportCounters bool) ([]byte, error) {
	// create command with arguments
	saveCmd := IPv4SaveCmd
	if protocol == ProtocolIPv6 {
		saveCmd = IPv6SaveCmd
	}
	args := []string{"-t", table}
	if exportCounters {
		args = append(args, "-c")
	}
	cmd := exec.Command(saveCmd, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// run command and extract result
	err := cmd.Run()
	if err != nil {
		return nil, errors.Errorf("%s failed due to: %v (%s)", saveCmd, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// restoreTable import all data (in IPTable-save output format) for given table
func (h *IPTablesHandler) restoreTable(protocol L3Protocol, table string, data []byte, flush bool, importCounters bool) error {
	// create command with arguments
	restoreCmd := IPv4RestoreCmd
	if protocol == ProtocolIPv6 {
		restoreCmd = IPv6RestoreCmd
	}
	args := []string{"-T", table}
	if importCounters {
		args = append(args, "-c")
	}
	if !flush {
		args = append(args, "-n")
	}
	cmd := exec.Command(restoreCmd, args...)
	cmd.Stdin = bytes.NewReader(data)

	// run command and extract result
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Errorf("%s failed due to: %v (%s)", restoreCmd, err, string(output))
	}
	return nil
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
		return nil, fmt.Errorf("iptables handler for protocol %v is not initialized " +
			"(please check that you have installed iptables in host system)", protocol)
	}
	return handler, nil
}
