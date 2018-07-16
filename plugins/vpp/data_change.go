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

package vpp

import (
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/model/bfd"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l4"
	"github.com/ligato/vpp-agent/plugins/vpp/model/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/model/srv6"
	"github.com/ligato/vpp-agent/plugins/vpp/model/stn"
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
			if err := plugin.dataChangeACL(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
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
			var value, prevValue l2.FibTable_FibEntry
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
		var value, prevValue l3.ArpTable_ArpEntry
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
	} else if strings.HasPrefix(key, l3.ProxyArpInterfacePrefix()) {
		var value, prevValue l3.ProxyArpInterfaces_InterfaceList
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeProxyARPInterface(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, l3.ProxyArpRangePrefix()) {
		var value, prevValue l3.ProxyArpRanges_RangeList
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeProxyARPRange(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
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
		var value, prevValue stn.STN_Rule
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
	} else if strings.HasPrefix(key, nat.GlobalConfigPrefix()) {
		// Global NAT config
		var value, prevValue nat.Nat44Global
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeNatGlobal(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, nat.SNatPrefix()) {
		// SNAT config
		var value, prevValue nat.Nat44SNat_SNatConfig
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeSNat(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, nat.DNatPrefix()) {
		// DNAT config
		var value, prevValue nat.Nat44DNat_DNatConfig
		if err := dataChng.GetValue(&value); err != nil {
			return false, err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeDNat(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, ipsec.KeyPrefix) {
		if strings.HasPrefix(key, ipsec.KeyPrefixSPD) {
			var value, prevValue ipsec.SecurityPolicyDatabases_SPD
			if err := dataChng.GetValue(&value); err != nil {
				return false, err
			}
			if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
				if err := plugin.dataChangeIPSecSPD(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		} else if strings.HasPrefix(key, ipsec.KeyPrefixSA) {
			var value, prevValue ipsec.SecurityAssociations_SA
			if err := dataChng.GetValue(&value); err != nil {
				return false, err
			}
			if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
				if err := plugin.dataChangeIPSecSA(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		} else if strings.HasPrefix(key, ipsec.KeyPrefixTunnel) {
			var value, prevValue ipsec.TunnelInterfaces_Tunnel
			if err := dataChng.GetValue(&value); err != nil {
				return false, err
			}
			if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
				if err := plugin.dataChangeIPSecTunnel(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		}
	} else if strings.HasPrefix(key, srv6.LocalSIDPrefix()) {
		var value, prevValue srv6.LocalSID
		if diff, err := plugin.extractFrom(dataChng, &value, &prevValue); err == nil {
			if err := plugin.dataChangeLocalSID(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else if strings.HasPrefix(key, srv6.PolicyPrefix()) {
		if srv6.IsPolicySegmentPrefix(key) { //Policy segment
			var value, prevValue srv6.PolicySegment
			if diff, err := plugin.extractFrom(dataChng, &value, &prevValue); err == nil {
				if name, err := srv6.ParsePolicySegmentKey(key); err == nil {
					if err := plugin.dataChangePolicySegment(name, diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
						return false, err
					}
				} else {
					return false, err
				}
			} else {
				return false, err
			}
		} else { // Policy
			var value, prevValue srv6.Policy
			if diff, err := plugin.extractFrom(dataChng, &value, &prevValue); err == nil {
				if err := plugin.dataChangePolicy(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
					return false, err
				}
			} else {
				return false, err
			}
		}
	} else if strings.HasPrefix(key, srv6.SteeringPrefix()) {
		var value, prevValue srv6.Steering
		if diff, err := plugin.extractFrom(dataChng, &value, &prevValue); err == nil {
			if err := plugin.dataChangeSteering(strings.TrimPrefix(key, srv6.SteeringPrefix()), diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	} else {
		plugin.Log.Warnf("ignoring change %v by VPP standard plugins: %q", dataChng, key) //NOT ERROR!
	}
	return false, nil
}

// extractFrom change event <dataChng> current value into <value> and previous value into <prevValue>
func (plugin *Plugin) extractFrom(dataChng datasync.ChangeEvent, value proto.Message, prevValue proto.Message) (prevValueExist bool, err error) {
	if err := dataChng.GetValue(value); err != nil {
		return false, err
	}
	return dataChng.GetPrevValue(prevValue)
}

// dataChangeACL propagates data change to the particular aclConfigurator.
func (plugin *Plugin) dataChangeACL(diff bool, value *acl.AccessLists_Acl, prevValue *acl.AccessLists_Acl,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeAcl ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.aclConfigurator.DeleteACL(prevValue)
	} else if diff {
		return plugin.aclConfigurator.ModifyACL(prevValue, value)
	}
	return plugin.aclConfigurator.ConfigureACL(value)
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
func (plugin *Plugin) dataChangeFIB(diff bool, value *l2.FibTable_FibEntry, prevValue *l2.FibTable_FibEntry,
	changeType datasync.PutDel, callback func(error)) error {
	plugin.Log.Debug("dataChangeFIB diff=", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.fibConfigurator.Delete(prevValue, callback)
	} else if diff {
		return plugin.fibConfigurator.Modify(prevValue, value, callback)
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
func (plugin *Plugin) dataChangeARP(diff bool, value *l3.ArpTable_ArpEntry, prevValue *l3.ArpTable_ArpEntry,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeARP diff=", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.arpConfigurator.DeleteArp(prevValue)
	} else if diff {
		return plugin.arpConfigurator.ChangeArp(value, prevValue)
	}
	return plugin.arpConfigurator.AddArp(value)
}

// dataChangeProxyARPInterface propagates data change to the arpConfigurator
func (plugin *Plugin) dataChangeProxyARPInterface(diff bool, value, prevValue *l3.ProxyArpInterfaces_InterfaceList,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeProxyARPInterface diff=", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.proxyArpConfigurator.DeleteInterface(prevValue)
	} else if diff {
		return plugin.proxyArpConfigurator.ModifyInterface(value, prevValue)
	}
	return plugin.proxyArpConfigurator.AddInterface(value)
}

// dataChangeProxyARPRange propagates data change to the arpConfigurator
func (plugin *Plugin) dataChangeProxyARPRange(diff bool, value, prevValue *l3.ProxyArpRanges_RangeList,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeProxyARPRange diff=", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.proxyArpConfigurator.DeleteRange(prevValue)
	} else if diff {
		return plugin.proxyArpConfigurator.ModifyRange(value, prevValue)
	}
	return plugin.proxyArpConfigurator.AddRange(value)
}

// DataChangeStaticRoute propagates data change to the l4Configurator
func (plugin *Plugin) dataChangeAppNamespace(diff bool, value *l4.AppNamespaces_AppNamespace, prevValue *l4.AppNamespaces_AppNamespace,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeL4AppNamespace ", diff, " ", changeType, " ", value, " ", prevValue)

	if datasync.Delete == changeType {
		return plugin.appNsConfigurator.DeleteAppNamespace(prevValue)
	} else if diff {
		return plugin.appNsConfigurator.ModifyAppNamespace(value, prevValue)
	}
	return plugin.appNsConfigurator.ConfigureAppNamespace(value)
}

// DataChangeL4Features propagates data change to the l4Configurator
func (plugin *Plugin) dataChangeL4Features(value *l4.L4Features, prevValue *l4.L4Features,
	changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeL4Feature ", changeType, " ", value, " ", prevValue)

	// diff and previous value is not important, features flag can be either set or not.
	// If removed, it is always set to false
	if datasync.Delete == changeType {
		return plugin.appNsConfigurator.DeleteL4FeatureFlag()
	}
	return plugin.appNsConfigurator.ConfigureL4FeatureFlag(value)
}

// DataChangeStnRule propagates data change to the stn configurator
func (plugin *Plugin) dataChangeStnRule(diff bool, value *stn.STN_Rule, prevValue *stn.STN_Rule, changeType datasync.PutDel) error {
	plugin.Log.Debug("stnRuleChange diff->", diff, " changeType->", changeType, " value->", value, " prevValue->", prevValue)

	if datasync.Delete == changeType {
		return plugin.stnConfigurator.Delete(prevValue)
	} else if diff {
		return plugin.stnConfigurator.Modify(prevValue, value)
	}
	return plugin.stnConfigurator.Add(value)
}

// dataChangeNatGlobal propagates data change to the nat configurator
func (plugin *Plugin) dataChangeNatGlobal(diff bool, value, prevValue *nat.Nat44Global, changeType datasync.PutDel) error {
	plugin.Log.Debug("natGlobalChange diff->", diff, " changeType->", changeType, " value->", value, " prevValue->", prevValue)

	if datasync.Delete == changeType {
		return plugin.natConfigurator.DeleteNatGlobalConfig(prevValue)
	} else if diff {
		return plugin.natConfigurator.ModifyNatGlobalConfig(prevValue, value)
	}
	return plugin.natConfigurator.SetNatGlobalConfig(value)
}

// dataChangeSNat propagates data change to the nat configurator
func (plugin *Plugin) dataChangeSNat(diff bool, value, prevValue *nat.Nat44SNat_SNatConfig, changeType datasync.PutDel) error {
	plugin.Log.Debug("sNatChange diff->", diff, " changeType->", changeType, " value->", value, " prevValue->", prevValue)

	if datasync.Delete == changeType {
		return plugin.natConfigurator.DeleteSNat(prevValue)
	} else if diff {
		return plugin.natConfigurator.ModifySNat(prevValue, value)
	}
	return plugin.natConfigurator.ConfigureSNat(value)
}

// dataChangeDNat propagates data change to the nat configurator
func (plugin *Plugin) dataChangeDNat(diff bool, value, prevValue *nat.Nat44DNat_DNatConfig, changeType datasync.PutDel) error {
	plugin.Log.Debug("dNatChange diff->", diff, " changeType->", changeType, " value->", value, " prevValue->", prevValue)

	if datasync.Delete == changeType {
		return plugin.natConfigurator.DeleteDNat(prevValue)
	} else if diff {
		return plugin.natConfigurator.ModifyDNat(prevValue, value)
	}
	return plugin.natConfigurator.ConfigureDNat(value)
}

// dataChangeIPSecSPD propagates data change to the IPSec configurator
func (plugin *Plugin) dataChangeIPSecSPD(diff bool, value, prevValue *ipsec.SecurityPolicyDatabases_SPD, changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeIPSecSPD diff->", diff, " changeType->", changeType, " value->", value, " prevValue->", prevValue)

	if datasync.Delete == changeType {
		return plugin.ipSecConfigurator.DeleteSPD(prevValue)
	} else if diff {
		return plugin.ipSecConfigurator.ModifySPD(prevValue, value)
	}
	return plugin.ipSecConfigurator.ConfigureSPD(value)
}

// dataChangeIPSecSA propagates data change to the IPSec configurator
func (plugin *Plugin) dataChangeIPSecSA(diff bool, value, prevValue *ipsec.SecurityAssociations_SA, changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeIPSecSA diff->", diff, " changeType->", changeType, " value->", value, " prevValue->", prevValue)

	if datasync.Delete == changeType {
		return plugin.ipSecConfigurator.DeleteSA(prevValue)
	} else if diff {
		return plugin.ipSecConfigurator.ModifySA(prevValue, value)
	}
	return plugin.ipSecConfigurator.ConfigureSA(value)
}

// dataChangeIPSecTunnel propagates data change to the IPSec configurator
func (plugin *Plugin) dataChangeIPSecTunnel(diff bool, value, prevValue *ipsec.TunnelInterfaces_Tunnel, changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeIPSecTunnel diff->", diff, " changeType->", changeType, " value->", value, " prevValue->", prevValue)

	if datasync.Delete == changeType {
		return plugin.ipSecConfigurator.DeleteTunnel(prevValue)
	} else if diff {
		return plugin.ipSecConfigurator.ModifyTunnel(prevValue, value)
	}
	return plugin.ipSecConfigurator.ConfigureTunnel(value)
}

// DataChangeLocalSID handles change events from ETCD related to local SIDs
func (plugin *Plugin) dataChangeLocalSID(diff bool, value *srv6.LocalSID, prevValue *srv6.LocalSID, changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeLocalSIDs ", diff, " ", changeType, " ", value, " ", prevValue)
	if datasync.Delete == changeType {
		return plugin.srv6Configurator.DeleteLocalSID(prevValue)
	} else if diff {
		return plugin.srv6Configurator.ModifyLocalSID(value, prevValue)
	}
	return plugin.srv6Configurator.AddLocalSID(value)
}

// dataChangePolicy handles change events from ETCD related to policies
func (plugin *Plugin) dataChangePolicy(diff bool, value *srv6.Policy, prevValue *srv6.Policy, changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangePolicy ", diff, " ", changeType, " ", value, " ", prevValue)
	if datasync.Delete == changeType {
		return plugin.srv6Configurator.RemovePolicy(prevValue)
	} else if diff {
		return plugin.srv6Configurator.ModifyPolicy(value, prevValue)
	}
	return plugin.srv6Configurator.AddPolicy(value)
}

// dataChangePolicySegment handles change events from ETCD related to policies segments
func (plugin *Plugin) dataChangePolicySegment(segmentName string, diff bool, value *srv6.PolicySegment, prevValue *srv6.PolicySegment, changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangePolicySegment ", segmentName, " ", diff, " ", changeType, " ", value, " ", prevValue)
	if datasync.Delete == changeType {
		return plugin.srv6Configurator.RemovePolicySegment(segmentName, prevValue)
	} else if diff {
		return plugin.srv6Configurator.ModifyPolicySegment(segmentName, value, prevValue)
	}
	return plugin.srv6Configurator.AddPolicySegment(segmentName, value)
}

// dataChangeSteering handles change events from ETCD related to steering
func (plugin *Plugin) dataChangeSteering(steeringName string, diff bool, value *srv6.Steering, prevValue *srv6.Steering, changeType datasync.PutDel) error {
	plugin.Log.Debug("dataChangeSteering ", steeringName, " ", diff, " ", changeType, " ", value, " ", prevValue)
	if datasync.Delete == changeType {
		return plugin.srv6Configurator.RemoveSteering(steeringName, prevValue)
	} else if diff {
		return plugin.srv6Configurator.ModifySteering(steeringName, value, prevValue)
	}
	return plugin.srv6Configurator.AddSteering(steeringName, value)
}
