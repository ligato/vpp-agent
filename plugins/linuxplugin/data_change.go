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
	intf "github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
)

// DataChangeIface propagates data change to the ifConfigurator
func (plugin *Plugin) dataChangeIface(diff bool, value *intf.LinuxInterfaces_Interface, prevValue *intf.LinuxInterfaces_Interface,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeIface ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.ifConfigurator.DeleteLinuxInterface(prevValue)
	} else if diff {
		return plugin.ifConfigurator.ModifyLinuxInterface(value, prevValue)
	}
	return plugin.ifConfigurator.ConfigureLinuxInterface(value)
}

// DataChangeArp propagates data change to the arpConfigurator
func (plugin *Plugin) dataChangeArp(diff bool, value *l3.LinuxStaticArpEntries_ArpEntry, prevValue *l3.LinuxStaticArpEntries_ArpEntry,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeArp ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.arpConfigurator.DeleteLinuxStaticArpEntry(prevValue)
	} else if diff {
		return plugin.arpConfigurator.ModifyLinuxStaticArpEntry(value, prevValue)
	}
	return plugin.arpConfigurator.ConfigureLinuxStaticArpEntry(value)
}

// DataChangeRoute propagates data change to the routeConfigurator
func (plugin *Plugin) dataChangeRoute(diff bool, value *l3.LinuxStaticRoutes_Route, prevValue *l3.LinuxStaticRoutes_Route,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeRoute ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.routeConfigurator.DeleteLinuxStaticRoute(prevValue)
	} else if diff {
		return plugin.routeConfigurator.ModifyLinuxStaticRoute(value, prevValue)
	}
	return plugin.routeConfigurator.ConfigureLinuxStaticRoute(value)
}
