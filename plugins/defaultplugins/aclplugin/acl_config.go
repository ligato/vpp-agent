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

//go:generate protoc --proto_path=../common/model/acl --gogo_out=../common/model/acl ../common/model/acl/acl.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/acl.api.json --output-dir=../common/bin_api

// Package aclplugin implements the ACL Plugin that handles management of VPP
// Access lists.
package aclplugin

import (
	"fmt"

	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/aclidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/vppdump"
	acl_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// Interface attribute according to the configuration
const (
	INGRESS = "ingress"
	EGRESS  = "egress"
	L2      = "l2"
)

// ACLIfCacheEntry contains info about interface, aclID and whether it is MAC IP address. Used as a cache for missing
// interfaces while configuring ACL
type ACLIfCacheEntry struct {
	ifName string
	aclID  uint32
	ifAttr string
}

// ACLConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of ACLs as modelled by the proto file "../model/acl/acl.proto" and stored
// in ETCD under the key "/vnf-agent/{agent-label}/vpp/config/v1/acl/". Updates received from the northbound API
// are compared with the VPP run-time configuration and differences are applied through the VPP binary API.
type ACLConfigurator struct {
	Log            logging.Logger
	GoVppmux       govppmux.API
	ACLL3L4Indexes aclidx.AclIndexRW
	ACLL2Indexes   aclidx.AclIndexRW // mapping for L2 ACLs
	SwIfIndexes    ifaceidx.SwIfIndex
	Stopwatch      *measure.Stopwatch // timer used to measure and store time

	ACLIfCache []*ACLIfCacheEntry // cache for ACL un-configured interfaces

	vppcalls   *vppcalls.ACLInterfacesVppCalls
	vppChannel *api.Channel
}

// Init goroutines, channels and mappings.
func (plugin *ACLConfigurator) Init() (err error) {
	plugin.Log.Infof("Initializing ACL configurator")

	// Init VPP API channel.
	plugin.vppChannel, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	if err := vppcalls.CheckMsgCompatibilityForACL(plugin.Log, plugin.vppChannel); err != nil {
		return err
	}

	// TODO: possibly check acl plugin version on vpp using bin api acl_plugin_get_version

	plugin.vppcalls = vppcalls.NewACLInterfacesVppCalls(plugin.vppChannel, plugin.SwIfIndexes, plugin.Stopwatch)

	return nil
}

// Close GOVPP channel.
func (plugin *ACLConfigurator) Close() {
	safeclose.Close(plugin.vppChannel)
}

// ConfigureACL creates access list with provided rules and sets this list to every relevant interface.
func (plugin *ACLConfigurator) ConfigureACL(acl *acl.AccessLists_Acl) error {
	plugin.Log.Infof("Configuring new ACL %v", acl.AclName)

	if acl.Rules != nil && len(acl.Rules) > 0 {
		// Validate rules.
		rules, isL2MacIP := plugin.validateRules(acl.AclName, acl.Rules)
		// Configure ACL rules.
		var vppACLIndex uint32
		var err error
		if isL2MacIP {
			vppACLIndex, err = vppcalls.AddMacIPAcl(rules, acl.AclName, plugin.Log, plugin.vppChannel, plugin.Stopwatch)
			if err != nil {
				return err
			}
			// Index used for L2 registration is ACLIndex + 1 (ACL indexes start from 0).
			agentACLIndex := vppACLIndex + 1
			plugin.ACLL2Indexes.RegisterName(acl.AclName, agentACLIndex, acl)
			plugin.Log.Debugf("ACL %v registered with index %v", acl.AclName, agentACLIndex)
		} else {
			vppACLIndex, err = vppcalls.AddIPAcl(rules, acl.AclName, plugin.Log, plugin.vppChannel, plugin.Stopwatch)
			if err != nil {
				return err
			}
			// Index used for L3L4 registration is aclIndex + 1 (ACL indexes start from 0).
			agentACLIndex := vppACLIndex + 1
			plugin.ACLL3L4Indexes.RegisterName(acl.AclName, agentACLIndex, acl)
			plugin.Log.Debugf("ACL %v registered with index %v", acl.AclName, agentACLIndex)
		}

		// Set ACL to interfaces.
		if acl.Interfaces != nil {
			if isL2MacIP {
				aclIfIndices := plugin.getOrCacheInterfaces(acl.Interfaces.Ingress, vppACLIndex, L2)
				err := plugin.vppcalls.SetMacIPAclToInterface(vppACLIndex, aclIfIndices, plugin.Log)
				if err != nil {
					return err
				}
			} else {
				aclIfInIndices := plugin.getOrCacheInterfaces(acl.Interfaces.Ingress, vppACLIndex, INGRESS)
				err = plugin.vppcalls.SetACLToInterfacesAsIngress(vppACLIndex, aclIfInIndices, plugin.Log)
				if err != nil {
					return err
				}
				aclIfEgIndices := plugin.getOrCacheInterfaces(acl.Interfaces.Egress, vppACLIndex, EGRESS)
				err = plugin.vppcalls.SetACLToInterfacesAsEgress(vppACLIndex, aclIfEgIndices, plugin.Log)
				if err != nil {
					return err
				}
			}
		} else {
			plugin.Log.Infof("No interface configured for ACL %v", acl.AclName)
		}
	}

	return nil
}

// ModifyACL modifies previously created access list. L2 access list is removed and recreated,
// L3/L4 access list is modified directly. List of interfaces is refreshed as well.
func (plugin *ACLConfigurator) ModifyACL(oldACL, newACL *acl.AccessLists_Acl) (err error) {
	plugin.Log.Infof("Modifying ACL %v", oldACL.AclName)

	if newACL.Rules != nil {
		// Validate rules.
		rules, isL2MacIP := plugin.validateRules(newACL.AclName, newACL.Rules)
		var vppACLIndex uint32
		if isL2MacIP {
			agentACLIndex, _, found := plugin.ACLL2Indexes.LookupIdx(newACL.AclName)
			if !found {
				plugin.Log.Infof("Acl %v index not found", newACL.AclName)
				return nil
			}
			// Index used in VPP = index used in mapping - 1
			vppACLIndex = agentACLIndex - 1
		} else {
			agentACLIndex, _, found := plugin.ACLL3L4Indexes.LookupIdx(newACL.AclName)
			if !found {
				plugin.Log.Infof("Acl %v index not found", newACL.AclName)
				return nil
			}
			vppACLIndex = agentACLIndex - 1
		}
		if isL2MacIP {
			// L2 ACL
			err := vppcalls.DeleteMacIPAcl(vppACLIndex, plugin.Log, plugin.vppChannel, plugin.Stopwatch)
			if err != nil {
				return err
			}
			plugin.ACLL2Indexes.UnregisterName(newACL.AclName)
			newVppACLIndex, err := vppcalls.AddMacIPAcl(rules, newACL.AclName, plugin.Log, plugin.vppChannel, plugin.Stopwatch)
			if err != nil {
				return err
			}
			// Create agent index by incrementing the vpp one.
			newAgentACLIndex := newVppACLIndex + 1
			plugin.ACLL2Indexes.RegisterName(newACL.AclName, newAgentACLIndex, nil)
		} else {
			// L3/L4 ACL can be modified directly.
			err := vppcalls.ModifyIPAcl(vppACLIndex, rules, newACL.AclName, plugin.Log, plugin.vppChannel, plugin.Stopwatch)
			if err != nil {
				return err
			}
			// There is no need to update index because modified ACL keeps the old one.
		}

		// Update interfaces.
		if isL2MacIP {
			// Remove L2 ACL from old interfaces.
			if oldACL.Interfaces != nil {

				err := plugin.vppcalls.RemoveMacIPIngressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(oldACL.Interfaces.Ingress), plugin.Log)
				if err != nil {
					return err
				}
			}
			// Put L2 ACL to new interfaces.
			if newACL.Interfaces != nil {
				aclMacInterfaces := plugin.getOrCacheInterfaces(newACL.Interfaces.Ingress, vppACLIndex, L2)
				err := plugin.vppcalls.SetMacIPAclToInterface(vppACLIndex, aclMacInterfaces, plugin.Log)
				if err != nil {
					return err
				}
			}

		} else {
			// Remove L3/L4 ACL from old interfaces.
			if oldACL.Interfaces != nil {
				err = plugin.vppcalls.RemoveIPIngressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(oldACL.Interfaces.Ingress), plugin.Log)
				if err != nil {
					return err
				}
				err = plugin.vppcalls.RemoveIPEgressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(oldACL.Interfaces.Egress), plugin.Log)
				if err != nil {
					return err
				}
			}
			// Put L3/L4 ACL to new interfaces.
			if newACL.Interfaces != nil {
				aclInInterfaces := plugin.getOrCacheInterfaces(newACL.Interfaces.Ingress, vppACLIndex, INGRESS)
				err = plugin.vppcalls.SetACLToInterfacesAsIngress(vppACLIndex, aclInInterfaces, plugin.Log)
				if err != nil {
					return err
				}
				aclEgInterfaces := plugin.getOrCacheInterfaces(newACL.Interfaces.Egress, vppACLIndex, EGRESS)
				err = plugin.vppcalls.SetACLToInterfacesAsEgress(vppACLIndex, aclEgInterfaces, plugin.Log)
				if err != nil {
					return err
				}
			}
		}
	}

	return err
}

// DeleteACL removes existing ACL. To detach ACL from interfaces, list of interfaces has to be provided.
func (plugin *ACLConfigurator) DeleteACL(acl *acl.AccessLists_Acl) (err error) {
	plugin.Log.Infof("Deleting ACL %v", acl.AclName)

	// Get ACL index. Keep in mind that all ACL Indices were incremented by 1.
	agentL2AclIndex, _, l2AclFound := plugin.ACLL2Indexes.LookupIdx(acl.AclName)
	agentL3L4AclIndex, _, l3l4AclFound := plugin.ACLL3L4Indexes.LookupIdx(acl.AclName)
	if !l2AclFound && !l3l4AclFound {
		return fmt.Errorf("ACL %v not found, cannot be removed", acl.AclName)
	}
	if l2AclFound {
		// Remove interfaces from L2 ACL.
		vppACLIndex := agentL2AclIndex - 1
		if acl.Interfaces != nil {
			err := plugin.vppcalls.RemoveMacIPIngressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(acl.Interfaces.Ingress), plugin.Log)
			if err != nil {
				return err
			}
		}
		// Remove ACL L2.
		err := vppcalls.DeleteMacIPAcl(vppACLIndex, plugin.Log, plugin.vppChannel, plugin.Stopwatch)
		if err != nil {
			return err
		}
		// Unregister.
		plugin.ACLL2Indexes.UnregisterName(acl.AclName)
	}
	if l3l4AclFound {
		// Remove interfaces.
		vppACLIndex := agentL3L4AclIndex - 1
		if acl.Interfaces != nil {
			err = plugin.vppcalls.RemoveIPIngressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(acl.Interfaces.Ingress), plugin.Log)
			if err != nil {
				return err
			}

			err = plugin.vppcalls.RemoveIPEgressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(acl.Interfaces.Egress), plugin.Log)
			if err != nil {
				return err
			}
		}
		// Remove ACL L3/L4.
		err := vppcalls.DeleteIPAcl(vppACLIndex, plugin.Log, plugin.vppChannel, plugin.Stopwatch)
		if err != nil {
			return err
		}
		// Unregister.
		plugin.ACLL3L4Indexes.UnregisterName(acl.AclName)
	}

	return err
}

// DumpACL returns all configured ACLs in proto format
func (plugin *ACLConfigurator) DumpACL() (acls []*acl.AccessLists_Acl, err error) {
	aclsWithIndex, err := vppdump.DumpACLs(plugin.Log, plugin.SwIfIndexes, plugin.vppChannel, measure.GetTimeLog(acl_api.ACLDump{}, plugin.Stopwatch))
	if err != nil {
		plugin.Log.Error(err)
		return nil, err
	}
	for _, aclWithIndex := range aclsWithIndex {
		acls = append(acls, aclWithIndex.ACLDetails)
	}
	return acls, nil
}

// Returns a list of existing ACL interfaces
func (plugin *ACLConfigurator) getInterfaces(interfaces []string) (configurableIfs []uint32) {
	for _, name := range interfaces {
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(name)
		if !found {
			continue
		}
		configurableIfs = append(configurableIfs, ifIdx)
	}
	return configurableIfs
}

// ResolveCreatedInterface configures new interface for every ACL found in cache
func (plugin *ACLConfigurator) ResolveCreatedInterface(ifName string, ifIdx uint32) error {
	plugin.Log.Debugf("ACL configurator: resolving new interface %v", ifName)

	// Iterate over cache in order to find out where the interface is used
	var wasErr error
	for entryIdx, aclCacheEntry := range plugin.ACLIfCache {
		if aclCacheEntry.ifName == ifName {
			var ifIndices []uint32
			switch aclCacheEntry.ifAttr {
			case L2:
				if err := plugin.vppcalls.SetMacIPAclToInterface(aclCacheEntry.aclID, append(ifIndices, ifIdx), plugin.Log); err != nil {
					plugin.Log.Error(err)
					wasErr = err
				}
			case INGRESS:
				if err := plugin.vppcalls.SetACLToInterfacesAsIngress(aclCacheEntry.aclID, append(ifIndices, ifIdx), plugin.Log); err != nil {
					plugin.Log.Error(err)
					wasErr = err
				}
			case EGRESS:
				if err := plugin.vppcalls.SetACLToInterfacesAsEgress(aclCacheEntry.aclID, append(ifIndices, ifIdx), plugin.Log); err != nil {
					plugin.Log.Error(err)
					wasErr = err
				}
			default:
				plugin.Log.Warnf("ACL interface is not defined as L2, ingress or egress")
			}
			// Remove from cache
			plugin.Log.Debugf("New interface %s (%s) configured for ACL %d, removed from cache",
				ifName, aclCacheEntry.ifAttr, aclCacheEntry.aclID)
			plugin.ACLIfCache = append(plugin.ACLIfCache[:entryIdx], plugin.ACLIfCache[entryIdx+1:]...)
		}
	}

	plugin.Log.Debugf("ACL configurator: new interface %v resolution done", ifName)

	return wasErr
}

// ResolveDeletedInterface puts removed interface to cache, including acl index. Note: it's not needed to remove ACL
// from interface manually, VPP handles it itself and such an behavior would cause errors (ACLs cannot be dumped
// from non-existing interface)
func (plugin *ACLConfigurator) ResolveDeletedInterface(ifName string, ifIdx uint32) error {
	plugin.Log.Debugf("ACL configurator: resolving deleted interface %v", ifName)

	var wasErr error

	// L3/L4 ingress/egress ACLs
	for _, aclName := range plugin.ACLL3L4Indexes.GetMapping().ListNames() {
		aclIdx, aclData, found := plugin.ACLL3L4Indexes.LookupIdx(aclName)
		if !found {
			plugin.Log.Warnf("ACL %v not found in the mapping", aclName)
			continue
		}
		vppAclIdx := aclIdx - 1
		if aclData != nil && aclData.Interfaces != nil {
			// Look over ingress interfaces
			for _, ingressIf := range aclData.Interfaces.Ingress {
				if ingressIf == ifName {
					plugin.ACLIfCache = append(plugin.ACLIfCache, &ACLIfCacheEntry{
						ifName: ifName,
						aclID:  vppAclIdx,
						ifAttr: INGRESS,
					})
				}
			}
			// Look over egress interfaces
			for _, ingressIf := range aclData.Interfaces.Egress {
				if ingressIf == ifName {
					plugin.ACLIfCache = append(plugin.ACLIfCache, &ACLIfCacheEntry{
						ifName: ifName,
						aclID:  vppAclIdx,
						ifAttr: EGRESS,
					})
				}
			}
		}
	}
	// L2 ACLs
	for _, aclName := range plugin.ACLL2Indexes.GetMapping().ListNames() {
		aclIdx, aclData, found := plugin.ACLL2Indexes.LookupIdx(aclName)
		if !found {
			plugin.Log.Warnf("ACL %v not found in the mapping", aclName)
			continue
		}
		vppAclIdx := aclIdx - 1
		if aclData != nil && aclData.Interfaces != nil {
			// Look over ingress interfaces
			for _, ingressIf := range aclData.Interfaces.Ingress {
				if ingressIf == ifName {
					plugin.ACLIfCache = append(plugin.ACLIfCache, &ACLIfCacheEntry{
						ifName: ifName,
						aclID:  vppAclIdx,
						ifAttr: L2,
					})
				}
			}
		}
	}

	plugin.Log.Debugf("ACL configurator: resolution done for removed interface %v", ifName)

	return wasErr
}

// Returns a list of interfaces configurable on the ACL. If interface is missing, put it to the cache. It will be
// configured when available
func (plugin *ACLConfigurator) getOrCacheInterfaces(interfaces []string, acl uint32, attr string) (configurableIfs []uint32) {
	for _, name := range interfaces {
		ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(name)
		if !found {
			// Put interface to cache
			plugin.ACLIfCache = append(plugin.ACLIfCache, &ACLIfCacheEntry{
				ifName: name,
				aclID:  acl,
				ifAttr: attr,
			})
			plugin.Log.Debugf("Interface %s (%s) not found for ACL %v, moving to cache", name, attr, acl)
			continue
		}
		configurableIfs = append(configurableIfs, ifIdx)
	}
	return configurableIfs
}

// Validate rules provided in ACL. Every rule has to contain actions and matches.
// Current limitation: L2 and L3/4 have to be split to different ACLs and
// there cannot be L2 rules and L3/4 rules in the same ACL.
func (plugin *ACLConfigurator) validateRules(aclName string, rules []*acl.AccessLists_Acl_Rule) ([]*acl.AccessLists_Acl_Rule, bool) {
	var validL3L4Rules []*acl.AccessLists_Acl_Rule
	var validL2Rules []*acl.AccessLists_Acl_Rule

	for index, rule := range rules {
		if rule.Actions == nil {
			plugin.Log.Warnf("Rule %v from acl %v does not contain actions", index, aclName)
			continue
		}
		if rule.Matches == nil {
			plugin.Log.Warnf("Rule %v from acl %v does not contain matches", index, aclName)
			continue
		}
		if rule.Matches.IpRule != nil {
			validL3L4Rules = append(validL3L4Rules, rule)
		}
		if rule.Matches.MacipRule != nil {
			validL2Rules = append(validL2Rules, rule)
		}
	}
	if len(validL3L4Rules) > 0 && len(validL2Rules) > 0 {
		plugin.Log.Errorf("Acl %v contains even L2 rules and L3/L4 rules. This case is not supported yet, only L3/L4 rules will be resolved",
			aclName)
		return validL3L4Rules, false
	} else if len(validL3L4Rules) > 0 {
		return validL3L4Rules, false
	} else {
		return validL2Rules, true
	}
}
