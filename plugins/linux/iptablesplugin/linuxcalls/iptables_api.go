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

// Protocol to differentiate between IPv4 and IPv6
type Protocol byte

const (
	ProtocolIPv4 Protocol = iota
	ProtocolIPv6
)

// IPTablesAPI interface covers all methods inside linux calls package needed
// to manage linux iptables rules.
type IPTablesAPI interface {
	// Init initializes an iptables handler.
	Init() error

	IPTablesAPIWrite
	IPTablesAPIRead
}

// IPTablesAPIWrite interface covers write methods inside linux calls package
// needed to manage linux iptables rules.
type IPTablesAPIWrite interface {
	// CreateChain creates an iptables chain in the specified table.
	CreateChain(proto Protocol, table, chain string) error

	// DeleteChain deletes an iptables chain in the specified table.
	DeleteChain(proto Protocol, table, chain string) error

	// SetChainDefaultPolicy sets default policy in the specified chain. Should be called only on FILTER tables.
	SetChainDefaultPolicy(proto Protocol, table, chain, defaultPolicy string) error

	// AppendRule appends a rule into the specified chain.
	AppendRule(proto Protocol, table, chain string, rule string) error

	// DeleteRule deletes a rule from the specified chain.
	DeleteRule(proto Protocol, table, chain string, rule string) error

	// DeleteAllRules deletes all rules within the specified chain.
	DeleteAllRules(proto Protocol, table, chain string) error
}

// IPTablesAPIRead interface covers read methods inside linux calls package
// needed to manage linux iptables rules.
type IPTablesAPIRead interface {
	// ListRules lists all rules within the specified chain.
	ListRules(proto Protocol, table, chain string) (rules []string, err error)
}

// NewIPTablesHandler creates new instance of iptables handler.
func NewIPTablesHandler() *IPTablesHandler {
	return &IPTablesHandler{}
}
