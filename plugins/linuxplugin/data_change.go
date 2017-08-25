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

package linuxplugin

import (
	"github.com/ligato/cn-infra/datasync"
	log "github.com/ligato/cn-infra/logging/logrus"
	intf "github.com/ligato/vpp-agent/plugins/linuxplugin/model/interfaces"
)

// DataChangeIface propagates data change to the ifConfigurator
func (plugin *Plugin) dataChangeIface(diff bool, value *intf.LinuxInterfaces_Interface, prevValue *intf.LinuxInterfaces_Interface,
	changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeIface ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.ifConfigurator.DeleteLinuxInterface(prevValue)
	} else if diff {
		return plugin.ifConfigurator.ModifyLinuxInterface(value, prevValue)
	}
	return plugin.ifConfigurator.ConfigureLinuxInterface(value)
}
