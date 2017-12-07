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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
)

// Resync writes ACLs to the empty VPP.
func (plugin *ACLConfigurator) Resync(acls []*acl.AccessLists_Acl, log logging.Logger) error {
	log.Debug("Resync ACLs started")
	// Calculate and log acl resync.
	defer func() {
		if plugin.Stopwatch != nil {
			plugin.Stopwatch.PrintLog()
		}
	}()

	var wasError error

	// Create VPP ACLs.
	log.Debugf("Configuring %v new ACLs", len(acls))
	for _, aclInput := range acls {
		plugin.ConfigureACL(aclInput)
	}

	log.WithField("cfg", plugin).Debug("RESYNC ACLs end. ", wasError)

	return wasError
}
