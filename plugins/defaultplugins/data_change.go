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

package defaultplugins

import (
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"

	"github.com/ligato/cn-infra/datasync"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
)

// dataChangeACL propagates to the particular aclConfigurator
func (plugin *Plugin) dataChangeACL(diff bool, value *acl.AccessLists_Acl, prevValue *acl.AccessLists_Acl,
	changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeAcl ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.aclConfigurator.DeleteACL(prevValue)
	} else if diff {
		return plugin.aclConfigurator.ModifyACL(prevValue, value)
	}
	return plugin.aclConfigurator.ConfigureACL(value)
}

// DataChangeIface propagates data change to the ifConfigurator
func (plugin *Plugin) dataChangeIface(diff bool, value *intf.Interfaces_Interface, prevValue *intf.Interfaces_Interface,
	changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeIface ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.ifConfigurator.DeleteVPPInterface(prevValue)
	} else if diff {
		return plugin.ifConfigurator.ModifyVPPInterface(value, prevValue)
	}
	return plugin.ifConfigurator.ConfigureVPPInterface(value)
}

// DataChangeBfdSession propagates data change  to the bfdConfigurator
func (plugin *Plugin) dataChangeBfdSession(diff bool, value *bfd.SingleHopBFD_Session, prevValue *bfd.SingleHopBFD_Session,
	changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeBfdSession ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.bfdConfigurator.DeleteBfdSession(prevValue)
	} else if diff {
		return plugin.bfdConfigurator.ModifyBfdSession(prevValue, value)
	}
	return plugin.bfdConfigurator.ConfigureBfdSession(value)
}

// DataChangeBfdKey propagates data change  to the bfdConfigurator
func (plugin *Plugin) dataChangeBfdKey(diff bool, value *bfd.SingleHopBFD_Key, prevValue *bfd.SingleHopBFD_Key,
	changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeBfdKey ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.bfdConfigurator.DeleteBfdAuthKey(prevValue)
	} else if diff {
		return plugin.bfdConfigurator.ModifyBfdAuthKey(prevValue, value)
	}
	return plugin.bfdConfigurator.ConfigureBfdAuthKey(value)
}

// DataChangeBfdEchoFunction propagates data change to the bfdConfigurator
func (plugin *Plugin) dataChangeBfdEchoFunction(diff bool, value *bfd.SingleHopBFD_EchoFunction, prevValue *bfd.SingleHopBFD_EchoFunction,
	changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeBfdEchoFunction ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.bfdConfigurator.DeleteBfdEchoFunction(prevValue)
	} else if diff {
		return plugin.bfdConfigurator.ModifyBfdEchoFunction(prevValue, value)
	}
	return plugin.bfdConfigurator.ConfigureBfdEchoFunction(value)
}

// dataChangeBD propagates data change to the bdConfigurator
func (plugin *Plugin) dataChangeBD(diff bool, value *l2.BridgeDomains_BridgeDomain, prevValue *l2.BridgeDomains_BridgeDomain,
	changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeBD ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.bdConfigurator.DeleteBridgeDomain(prevValue)
	} else if diff {
		return plugin.bdConfigurator.ModifyBridgeDomain(value, prevValue)
	}
	return plugin.bdConfigurator.ConfigureBridgeDomain(value)
}

// dataChangeFIB propagates data change to the fibConfigurator
func (plugin *Plugin) dataChangeFIB(diff bool, value *l2.FibTableEntries_FibTableEntry, prevValue *l2.FibTableEntries_FibTableEntry,
	changeType datasync.PutDel, callback func(error)) error {
	log.DefaultLogger().Debug("dataChangeFIB diff=", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.fibConfigurator.Delete(prevValue, callback)
	} else if diff {
		return plugin.fibConfigurator.Diff(prevValue, value, callback)
	}
	return plugin.fibConfigurator.Add(value, callback)
}

// DataChangeIface propagates data change to the xcConfugurator
func (plugin *Plugin) dataChangeXCon(diff bool, value *l2.XConnectPairs_XConnectPair, prevValue *l2.XConnectPairs_XConnectPair,
	changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeXCon ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.xcConfigurator.DeleteXConnectPair(prevValue)
	} else if diff {
		return plugin.xcConfigurator.ModifyXConnectPair(value, prevValue)
	}
	return plugin.xcConfigurator.ConfigureXConnectPair(value)

}

// DataChangeStaticRoute propagates data change to the routeConfigurator
func (plugin *Plugin) dataChangeStaticRoute(diff bool, value *l3.StaticRoutes_Route, prevValue *l3.StaticRoutes_Route,
	vrfFromKey string, changeType datasync.PutDel) error {
	log.DefaultLogger().Debug("dataChangeStaticRoute ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.routeConfigurator.DeleteRoute(prevValue)
	} else if diff {
		return plugin.routeConfigurator.ModifyRoute(value, prevValue, vrfFromKey)
	}
	return plugin.routeConfigurator.ConfigureRoute(value, vrfFromKey)
}
