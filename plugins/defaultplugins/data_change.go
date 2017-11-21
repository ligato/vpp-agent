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
	"strings"

	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/stn"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/model/l4"
)

func (plugin *Plugin) changePropagateRequest(dataChng datasync.ChangeEvent, callback func(error)) (callbackCalled bool, err error) {
	key := dataChng.GetKey()

	// Skip potential changes on error keys
	if strings.HasPrefix(key, interfaces.InterfaceErrorPrefix()) || strings.HasPrefix(key, l2.BridgeDomainErrorPrefix()) {
		return false, nil
	}
	plugin.Log.Debug("Start processing change for key: ", key)
	if strings.HasPrefix(key, acl.KeyPrefix()) {
		var value, prevValue acl.AccessLists_Acl
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeACL(diff, &value, &prevValue, dataChng.GetChangeType(), callback); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
		var value, prevValue interfaces.Interfaces_Interface
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeIface(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, bfd.SessionKeyPrefix()) {
		var value, prevValue bfd.SingleHopBFD_Session
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeBfdSession(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, bfd.AuthKeysKeyPrefix()) {
		var value, prevValue bfd.SingleHopBFD_Key
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeBfdKey(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, bfd.EchoFunctionKeyPrefix()) {
		var value, prevValue bfd.SingleHopBFD_EchoFunction
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeBfdEchoFunction(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, l2.BridgeDomainKeyPrefix()) {
		fib, _, _ := l2.ParseFibKey(key)
		if fib {
			// L2 FIB entry
			var value, prevValue l2.FibTableEntries_FibTableEntry
			if err := dataChng.GetValue(&value); err != nil {
				return false, err
			}
			if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
				if err := plugin.dataChangeFIB(diff, &value, &prevValue, dataChng.GetChangeType(), callback); err != nil {
					return true, err
				}
			} else {
				return false, err
			}
		} else {
			// Bridge domain
			var value, prevValue l2.BridgeDomains_BridgeDomain
			if err := dataChng.GetValue(&value); err != nil {
				return false, err
			}
			if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
				if err := plugin.dataChangeBD(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		}
	} else if strings.HasPrefix(key, l2.XConnectKeyPrefix()) {
		var value, prevValue l2.XConnectPairs_XConnectPair
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeXCon(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, l3.VrfKeyPrefix()) {
		isRoute, vrfFromKey, _, _, _ := l3.ParseRouteKey(key)
		if isRoute {
			// Route
			var value, prevValue l3.StaticRoutes_Route
			if err := dataChng.GetValue(&value); err != nil {
				return false, err
			}
			if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
				if err := plugin.dataChangeStaticRoute(diff, &value, &prevValue, vrfFromKey, dataChng.GetChangeType()); err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		} else {
			// Vrf
			// TODO vrf not implemented yet
			plugin.Log.Warn("VRFs are not supported yet")
		}
	} else if strings.HasPrefix(key, l3.ArpKeyPrefix()) {
		_, _, err := l3.ParseArpKey(key)
		if err != nil {
			return false, err
		}
		var value, prevValue l3.ArpTable_ArpTableEntry
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeARP(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, l4.AppNamespacesKeyPrefix()) {
		var value, prevValue l4.AppNamespaces_AppNamespace
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeAppNamespace(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, l4.FeatureKeyPrefix()) {
		var value, prevValue l4.L4Features
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if _, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeL4Features(&value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, stn.KeyPrefix()) {
		var value, prevValue stn.StnRule
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeStnRule(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else {
		plugin.Log.Warn("ignoring change ", dataChng, " by VPP standard plugins") //NOT ERROR!
	}
	return false, nil
}

// dataChangeACL propagates data change to the particular aclConfigurator.
func (plugin *Plugin) dataChangeACL(diff bool, value *acl.AccessLists_Acl, prevValue *acl.AccessLists_Acl,
	changeType datasync.PutDel, callback func(error)) error {
	plugin.Log.Debug("dataChangeAcl ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.aclConfigurator.DeleteACL(prevValue, callback)
	} else if diff {
		return plugin.aclConfigurator.ModifyACL(prevValue, value, callback)
	}
	return plugin.aclConfigurator.ConfigureACL(value, callback)
}

// DataChangeIface propagates data change to the ifConfigurator.
func (plugin *Plugin) dataChangeIface(diff bool, value *interfaces.Interfaces_Interface, prevValue *interfaces.Interfaces_Interface,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeIface ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.ifConfigurator.DeleteVPPInterface(prevValue)
	} else if diff {
		return plugin.ifConfigurator.ModifyVPPInterface(value, prevValue)
	}
	return plugin.ifConfigurator.ConfigureVPPInterface(value)
}

// DataChangeBfdSession propagates data change to the bfdConfigurator.
func (plugin *Plugin) dataChangeBfdSession(diff bool, value *bfd.SingleHopBFD_Session, prevValue *bfd.SingleHopBFD_Session,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeBfdSession ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.bfdConfigurator.DeleteBfdSession(prevValue)
	} else if diff {
		return plugin.bfdConfigurator.ModifyBfdSession(prevValue, value)
	}
	return plugin.bfdConfigurator.ConfigureBfdSession(value)
}

// DataChangeBfdKey propagates data change to the bfdConfigurator.
func (plugin *Plugin) dataChangeBfdKey(diff bool, value *bfd.SingleHopBFD_Key, prevValue *bfd.SingleHopBFD_Key,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeBfdKey ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.bfdConfigurator.DeleteBfdAuthKey(prevValue)
	} else if diff {
		return plugin.bfdConfigurator.ModifyBfdAuthKey(prevValue, value)
	}
	return plugin.bfdConfigurator.ConfigureBfdAuthKey(value)
}

// DataChangeBfdEchoFunction propagates data change to the bfdConfigurator.
func (plugin *Plugin) dataChangeBfdEchoFunction(diff bool, value *bfd.SingleHopBFD_EchoFunction, prevValue *bfd.SingleHopBFD_EchoFunction,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeBfdEchoFunction ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.bfdConfigurator.DeleteBfdEchoFunction(prevValue)
	} else if diff {
		return plugin.bfdConfigurator.ModifyBfdEchoFunction(prevValue, value)
	}
	return plugin.bfdConfigurator.ConfigureBfdEchoFunction(value)
}

// dataChangeBD propagates data change to the bdConfigurator.
func (plugin *Plugin) dataChangeBD(diff bool, value *l2.BridgeDomains_BridgeDomain, prevValue *l2.BridgeDomains_BridgeDomain,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeBD ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.bdConfigurator.DeleteBridgeDomain(prevValue)
	} else if diff {
		return plugin.bdConfigurator.ModifyBridgeDomain(value, prevValue)
	}
	return plugin.bdConfigurator.ConfigureBridgeDomain(value)
}

// dataChangeFIB propagates data change to the fibConfigurator.
func (plugin *Plugin) dataChangeFIB(diff bool, value *l2.FibTableEntries_FibTableEntry, prevValue *l2.FibTableEntries_FibTableEntry,
	changeType datasync.PutDel, callback func(error)) error {
	plugin.Log.Debug("dataChangeFIB diff=", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.fibConfigurator.Delete(prevValue, callback)
	} else if diff {
		return plugin.fibConfigurator.Diff(prevValue, value, callback)
	}
	return plugin.fibConfigurator.Add(value, callback)
}

// DataChangeIface propagates data change to the xcConfugurator.
func (plugin *Plugin) dataChangeXCon(diff bool, value *l2.XConnectPairs_XConnectPair, prevValue *l2.XConnectPairs_XConnectPair,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeXCon ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.xcConfigurator.DeleteXConnectPair(prevValue)
	} else if diff {
		return plugin.xcConfigurator.ModifyXConnectPair(value, prevValue)
	}
	return plugin.xcConfigurator.ConfigureXConnectPair(value)

}

// DataChangeStaticRoute propagates data change to the routeConfigurator.
func (plugin *Plugin) dataChangeStaticRoute(diff bool, value *l3.StaticRoutes_Route, prevValue *l3.StaticRoutes_Route,
	vrfFromKey string, changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeStaticRoute ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.routeConfigurator.DeleteRoute(prevValue, vrfFromKey)
	} else if diff {
		return plugin.routeConfigurator.ModifyRoute(value, prevValue, vrfFromKey)
	}
	return plugin.routeConfigurator.ConfigureRoute(value, vrfFromKey)
}

// dataChangeARP propagates data change to the arpConfigurator
func (plugin *Plugin) dataChangeARP(diff bool, value *l3.ArpTable_ArpTableEntry, prevValue *l3.ArpTable_ArpTableEntry,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeARP diff=", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.arpConfigurator.DeleteArp(prevValue)
	} else if diff {
		return plugin.arpConfigurator.ChangeArp(value, prevValue)
	}
	return plugin.arpConfigurator.AddArp(value)
}

// DataChangeStaticRoute propagates data change to the l4Configurator
func (plugin *Plugin) dataChangeAppNamespace(diff bool, value *l4.AppNamespaces_AppNamespace, prevValue *l4.AppNamespaces_AppNamespace,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeL4AppNamespace ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.l4Configurator.DeleteAppNamespace(prevValue)
	} else if diff {
		return plugin.l4Configurator.ModifyAppNamespace(value, prevValue)
	}
	return plugin.l4Configurator.ConfigureAppNamespace(value)
}

// DataChangeL4Features propagates data change to the l4Configurator
func (plugin *Plugin) dataChangeL4Features(value *l4.L4Features, prevValue *l4.L4Features,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeL4Feature ", changeType, " ", value, " ", prevValue)

	// diff and previous value is not important, features flag can be either set or not.
	// If removed, it is always set to false
	if datasync.Delete == changeType {
		return plugin.l4Configurator.DeleteL4FeatureFlag()
	}
	return plugin.l4Configurator.ConfigureL4FeatureFlag(value)
}

func (plugin *Plugin) dataChangeStnRule(diff bool, value *stn.StnRule, prevValue *stn.StnRule, changeType datasync.PutDel) error {
	plugin.Log.Debug("stnRuleChange ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.stnConfigurator.Delete(prevValue)
	} else if diff {
		return plugin.stnConfigurator.Modify(value, prevValue)
	}
	return plugin.stnConfigurator.Add(value)
}
