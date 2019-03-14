// Copyright (c) 2018 Cisco and/or its affiliates.
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

package data

import (
	"github.com/ligato/vpp-agent/api/models/linux/iptables"
)

const (
	chainLogicalNameV4 = "testchain-v4"
	chainLogicalNameV6 = "testchain-v6"
)

// IPTablesCtl provides Linux iptables related methods for vpp-agent-ctl
type IPTablesCtl interface {
	// PutIPTablesRule puts Linux iptables rule chain config into the ETCD
	PutIPTablesRule(ipv6 bool) error
	// DeleteIPTablesRule removes Linux iptables rule chain config from the ETCD
	DeleteIPTablesRule(ipv6 bool) error
}

// PutIPTablesRule puts Linux iptables rule chain config into the ETCD
func (ctl *VppAgentCtlImpl) PutIPTablesRule(ipv6 bool) error {
	ruleChain := &linux_iptables.RuleChain{
		Table:     linux_iptables.RuleChain_NAT,
		ChainType: linux_iptables.RuleChain_CUSTOM,
		ChainName: "TEST_CHAIN",
		Rules: []string{
			"-p tcp -m tcp --dport 80 -j REDIRECT --to-ports 9376",
			"-p tcp -m tcp --dport 443 -j REDIRECT --to-ports 9377",
		},
	}
	if ipv6 {
		ruleChain.Name = chainLogicalNameV6
		ruleChain.Protocol = linux_iptables.RuleChain_IPv6
	} else {
		ruleChain.Name = chainLogicalNameV4
		ruleChain.Protocol = linux_iptables.RuleChain_IPv4
	}

	ctl.Log.Infof("Linux iptables rulechain put: %v", ruleChain)
	return ctl.broker.Put(linux_iptables.RuleChainKey(ruleChain.Name), ruleChain)
}

// DeleteIPTablesRule removes Linux iptables rule chain config from the ETCD
func (ctl *VppAgentCtlImpl) DeleteIPTablesRule(ipv6 bool) error {
	var rchKey string

	if ipv6 {
		rchKey = linux_iptables.RuleChainKey(chainLogicalNameV6)
	} else {
		rchKey = linux_iptables.RuleChainKey(chainLogicalNameV4)
	}

	ctl.Log.Infof("Deleted Linux iptables rulechain: %v", rchKey)
	_, err := ctl.broker.Delete(rchKey)
	return err
}
