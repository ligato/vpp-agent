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

package linux_iptables

import (
	"github.com/ligato/vpp-agent/pkg/models"
)

// ModuleName is the module name used for models.
const ModuleName = "linux.iptables"

var (
	ModelRuleChain = models.Register(&RuleChain{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "rulechain",
	}, models.WithNameTemplate("{{.Name}}"))
)

// RuleChainKey returns the key used in KV database to store configuration of a particular Linux iptables rule chain.
func RuleChainKey(name string) string {
	return models.Key(&RuleChain{
		Name: name,
	})
}
