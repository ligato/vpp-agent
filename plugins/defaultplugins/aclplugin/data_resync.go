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

package aclplugin

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppdump"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
)

// Resync writes ACLs to the empty VPP.
func (plugin *ACLConfigurator) Resync(nbACLs []*acl.AccessLists_Acl, log logging.Logger) error {
	log.Debug("Resync ACLs started")
	// Calculate and log acl resync.
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	// Retrieve existing ACL config
	vppACLs, err := vppdump.DumpACLs(plugin.Log, plugin.SwIfIndexes, plugin.vppChannel, measure.GetTimeLog(&acl_api.ACLDump{}, plugin.Stopwatch))
	if err != nil {
		return err
	}

	// Remove all configured VPP ACLs
	// Note: due to unablity to dump ACL interfaces, it is not currently possible to correctly
	// calculate difference between configs
	var wasErr error
	for _, vppACL := range vppACLs {

		// ACL with IP-type rules uses different binary call to create/remove than MACIP-type.
		// Check what type of rules is in the ACL
		ipRulesExist := checkIPRules(vppACL.ACLDetails.Rules)

		if ipRulesExist {
			if err := vppcalls.DeleteIPAcl(vppACL.Identifier.ACLIndex, plugin.Log, plugin.vppChannel,
				measure.GetTimeLog(&acl_api.ACLDel{}, plugin.Stopwatch)); err != nil {
				log.Error(err)
				wasErr = err
			}
			continue
		} else {
			if err := vppcalls.DeleteMacIPAcl(vppACL.Identifier.ACLIndex, plugin.Log, plugin.vppChannel,
				measure.GetTimeLog(&acl_api.MacipACLAdd{}, plugin.Stopwatch)); err != nil {
				log.Error(err)
				wasErr = err
			}
		}
	}

	// Configure new ACLs
	for _, nbACL := range nbACLs {
		if err := plugin.ConfigureACL(nbACL); err != nil {
			plugin.Log.Error(err)
			wasErr = err
		}
	}

	return wasErr
}

// Method checks first rule whether it is IP rule type and returns true in such a case
func checkIPRules(rules []*acl.AccessLists_Acl_Rule) bool {
	if len(rules) > 0 {
		if rules[0].Matches != nil && rules[0].Matches.IpRule != nil {
			return true
		}
	}
	return false
}
