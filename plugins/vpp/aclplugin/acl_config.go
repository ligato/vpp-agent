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

//go:generate protoc --proto_path=../model/acl --gogo_out=../model/acl ../model/acl/acl.proto

// Package aclplugin implements the ACL Plugin that handles management of VPP
// Access lists.
package aclplugin

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin/aclidx"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
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
	log logging.Logger

	// In-memory mappings
	ifIndexes      ifaceidx.SwIfIndex
	l2AclIndexes   aclidx.AclIndexRW
	l3l4AclIndexes aclidx.AclIndexRW

	// Cache for ACL un-configured interfaces
	ifCache []*ACLIfCacheEntry

	// VPP channels
	vppChan     govppapi.Channel
	vppDumpChan govppapi.Channel

	// ACL VPP calls handler
	aclHandler vppcalls.AclVppAPI

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// Init goroutines, channels and mappings.
func (plugin *ACLConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-acl-plugin")
	plugin.log.Infof("Initializing ACL configurator")

	// Mappings
	plugin.ifIndexes = swIfIndexes
	plugin.l2AclIndexes = aclidx.NewAclIndex(nametoidx.NewNameToIdx(plugin.log, "acl_l2_indexes", nil))
	plugin.l3l4AclIndexes = aclidx.NewAclIndex(nametoidx.NewNameToIdx(plugin.log, "acl_l3_l4_indexes", nil))

	// VPP channels
	plugin.vppChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}
	plugin.vppDumpChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// Configurator-wide stopwatch instance
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("ACL-configurator", plugin.log)
	}

	// ACL binary api handler
	if plugin.aclHandler, err = vppcalls.NewAclVppHandler(plugin.vppChan, plugin.vppDumpChan, plugin.stopwatch); err != nil {
		return err
	}

	return nil
}

// Close GOVPP channel.
func (plugin *ACLConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan, plugin.vppDumpChan)
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (plugin *ACLConfigurator) clearMapping() {
	plugin.l2AclIndexes.Clear()
	plugin.l3l4AclIndexes.Clear()
}

// GetL2AclIfIndexes exposes l2 acl interface name-to-index mapping
func (plugin *ACLConfigurator) GetL2AclIfIndexes() aclidx.AclIndexRW {
	return plugin.l2AclIndexes
}

// GetL3L4AclIfIndexes exposes l3/l4 acl interface name-to-index mapping
func (plugin *ACLConfigurator) GetL3L4AclIfIndexes() aclidx.AclIndexRW {
	return plugin.l3l4AclIndexes
}

// ConfigureACL creates access list with provided rules and sets this list to every relevant interface.
func (plugin *ACLConfigurator) ConfigureACL(acl *acl.AccessLists_Acl) error {
	plugin.log.Infof("Configuring new ACL %v", acl.AclName)

	if len(acl.Rules) == 0 {
		plugin.log.Debugf("ACL %v has no rules set, skipping configuration", acl.AclName)
		return nil
	}

	rules, isL2MacIP := plugin.validateRules(acl.AclName, acl.Rules)
	// Configure ACL rules.
	var vppACLIndex uint32
	var err error
	if isL2MacIP {
		vppACLIndex, err = plugin.aclHandler.AddMacIPAcl(rules, acl.AclName)
		if err != nil {
			return err
		}
		// Index used for L2 registration is ACLIndex + 1 (ACL indexes start from 0).
		agentACLIndex := vppACLIndex + 1
		plugin.l2AclIndexes.RegisterName(acl.AclName, agentACLIndex, acl)
		plugin.log.Debugf("ACL %v registered with index %v", acl.AclName, agentACLIndex)
	} else {
		vppACLIndex, err = plugin.aclHandler.AddIPAcl(rules, acl.AclName)
		if err != nil {
			return err
		}
		// Index used for L3L4 registration is aclIndex + 1 (ACL indexes start from 0).
		agentACLIndex := vppACLIndex + 1
		plugin.l3l4AclIndexes.RegisterName(acl.AclName, agentACLIndex, acl)
		plugin.log.Debugf("ACL %v registered with index %v", acl.AclName, agentACLIndex)
	}

	// Set ACL to interfaces.
	if ifaces := acl.GetInterfaces(); ifaces != nil {
		if isL2MacIP {
			aclIfIndices := plugin.getOrCacheInterfaces(acl.Interfaces.Ingress, vppACLIndex, L2)
			err := plugin.aclHandler.SetMacIPAclToInterface(vppACLIndex, aclIfIndices)
			if err != nil {
				return err
			}
		} else {
			aclIfInIndices := plugin.getOrCacheInterfaces(acl.Interfaces.Ingress, vppACLIndex, INGRESS)
			err = plugin.aclHandler.SetACLToInterfacesAsIngress(vppACLIndex, aclIfInIndices)
			if err != nil {
				return err
			}
			aclIfEgIndices := plugin.getOrCacheInterfaces(acl.Interfaces.Egress, vppACLIndex, EGRESS)
			err = plugin.aclHandler.SetACLToInterfacesAsEgress(vppACLIndex, aclIfEgIndices)
			if err != nil {
				return err
			}
		}
	} else {
		plugin.log.Infof("No interface configured for ACL %v", acl.AclName)
	}

	return nil
}

// ModifyACL modifies previously created access list. L2 access list is removed and recreated,
// L3/L4 access list is modified directly. List of interfaces is refreshed as well.
func (plugin *ACLConfigurator) ModifyACL(oldACL, newACL *acl.AccessLists_Acl) (err error) {
	plugin.log.Infof("Modifying ACL %v", oldACL.AclName)

	if newACL.Rules != nil {
		// Validate rules.
		rules, isL2MacIP := plugin.validateRules(newACL.AclName, newACL.Rules)
		var vppACLIndex uint32
		if isL2MacIP {
			agentACLIndex, _, found := plugin.l2AclIndexes.LookupIdx(oldACL.AclName)
			if !found {
				plugin.log.Infof("Acl %v index not found", oldACL.AclName)
				return nil
			}
			// Index used in VPP = index used in mapping - 1
			vppACLIndex = agentACLIndex - 1
		} else {
			agentACLIndex, _, found := plugin.l3l4AclIndexes.LookupIdx(oldACL.AclName)
			if !found {
				plugin.log.Infof("Acl %v index not found", oldACL.AclName)
				return nil
			}
			vppACLIndex = agentACLIndex - 1
		}
		if isL2MacIP {
			// L2 ACL
			err := plugin.aclHandler.ModifyMACIPAcl(vppACLIndex, rules, newACL.AclName)
			if err != nil {
				return err
			}
			// There is no need to update index because modified ACL keeps the old one.
		} else {
			// L3/L4 ACL can be modified directly.
			err := plugin.aclHandler.ModifyIPAcl(vppACLIndex, rules, newACL.AclName)
			if err != nil {
				return err
			}
			// There is no need to update index because modified ACL keeps the old one.
		}

		// Update interfaces.
		if isL2MacIP {
			// Remove L2 ACL from old interfaces.
			if oldACL.Interfaces != nil {

				err := plugin.aclHandler.RemoveMacIPIngressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(oldACL.Interfaces.Ingress))
				if err != nil {
					return err
				}
			}
			// Put L2 ACL to new interfaces.
			if newACL.Interfaces != nil {
				aclMacInterfaces := plugin.getOrCacheInterfaces(newACL.Interfaces.Ingress, vppACLIndex, L2)
				err := plugin.aclHandler.SetMacIPAclToInterface(vppACLIndex, aclMacInterfaces)
				if err != nil {
					return err
				}
			}

		} else {
			aclOldInInterfaces := plugin.getInterfaces(oldACL.Interfaces.Ingress)
			aclOldEgInterfaces := plugin.getInterfaces(oldACL.Interfaces.Egress)
			aclNewInInterfaces := plugin.getOrCacheInterfaces(newACL.Interfaces.Ingress, vppACLIndex, INGRESS)
			aclNewEgInterfaces := plugin.getOrCacheInterfaces(newACL.Interfaces.Egress, vppACLIndex, EGRESS)
			addedInInterfaces, removedInInterfaces := diffInterfaces(aclOldInInterfaces, aclNewInInterfaces)
			addedEgInterfaces, removedEgInterfaces := diffInterfaces(aclOldEgInterfaces, aclNewEgInterfaces)

			if len(removedInInterfaces) > 0 {
				err = plugin.aclHandler.RemoveIPIngressACLFromInterfaces(vppACLIndex, removedInInterfaces)
				if err != nil {
					return err
				}
			}
			if len(removedEgInterfaces) > 0 {
				err = plugin.aclHandler.RemoveIPEgressACLFromInterfaces(vppACLIndex, removedEgInterfaces)
				if err != nil {
					return err
				}
			}
			if len(addedInInterfaces) > 0 {
				err = plugin.aclHandler.SetACLToInterfacesAsIngress(vppACLIndex, addedInInterfaces)
				if err != nil {
					return err
				}
			}
			if len(addedEgInterfaces) > 0 {
				err = plugin.aclHandler.SetACLToInterfacesAsEgress(vppACLIndex, addedEgInterfaces)
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
	plugin.log.Infof("Deleting ACL %v", acl.AclName)

	// Get ACL index. Keep in mind that all ACL Indices were incremented by 1.
	agentL2AclIndex, _, l2AclFound := plugin.l2AclIndexes.LookupIdx(acl.AclName)
	agentL3L4AclIndex, _, l3l4AclFound := plugin.l3l4AclIndexes.LookupIdx(acl.AclName)
	if !l2AclFound && !l3l4AclFound {
		return fmt.Errorf("ACL %v not found, cannot be removed", acl.AclName)
	}
	if l2AclFound {
		// Remove interfaces from L2 ACL.
		vppACLIndex := agentL2AclIndex - 1
		if acl.Interfaces != nil {
			err := plugin.aclHandler.RemoveMacIPIngressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(acl.Interfaces.Ingress))
			if err != nil {
				return err
			}
		}
		// Remove ACL L2.
		err := plugin.aclHandler.DeleteMacIPAcl(vppACLIndex)
		if err != nil {
			return err
		}
		// Unregister.
		plugin.l2AclIndexes.UnregisterName(acl.AclName)
	}
	if l3l4AclFound {
		// Remove interfaces.
		vppACLIndex := agentL3L4AclIndex - 1
		if acl.Interfaces != nil {
			err = plugin.aclHandler.RemoveIPIngressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(acl.Interfaces.Ingress))
			if err != nil {
				return err
			}

			err = plugin.aclHandler.RemoveIPEgressACLFromInterfaces(vppACLIndex, plugin.getInterfaces(acl.Interfaces.Egress))
			if err != nil {
				return err
			}
		}
		// Remove ACL L3/L4.
		err := plugin.aclHandler.DeleteIPAcl(vppACLIndex)
		if err != nil {
			return err
		}
		// Unregister.
		plugin.l3l4AclIndexes.UnregisterName(acl.AclName)
	}

	return err
}

// DumpIPACL returns all configured IP ACLs in proto format
func (plugin *ACLConfigurator) DumpIPACL() (acls []*acl.AccessLists_Acl, err error) {
	aclsWithIndex, err := plugin.aclHandler.DumpIPACL(plugin.ifIndexes)
	if err != nil {
		plugin.log.Error(err)
		return nil, err
	}
	for _, aclWithIndex := range aclsWithIndex {
		acls = append(acls, aclWithIndex.Acl)
	}
	return acls, nil
}

// DumpMACIPACL returns all configured MACIP ACLs in proto format
func (plugin *ACLConfigurator) DumpMACIPACL() (acls []*acl.AccessLists_Acl, err error) {
	aclsWithIndex, err := plugin.aclHandler.DumpMACIPACL(plugin.ifIndexes)
	if err != nil {
		plugin.log.Error(err)
		return nil, err
	}
	for _, aclWithIndex := range aclsWithIndex {
		acls = append(acls, aclWithIndex.Acl)
	}
	return acls, nil
}

// Returns a list of existing ACL interfaces
func (plugin *ACLConfigurator) getInterfaces(interfaces []string) (configurableIfs []uint32) {
	for _, name := range interfaces {
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(name)
		if !found {
			continue
		}
		configurableIfs = append(configurableIfs, ifIdx)
	}
	return configurableIfs
}

// diffInterfaces returns a difference between two lists of interfaces
func diffInterfaces(oldInterfaces, newInterfaces []uint32) (added, removed []uint32) {
	intfMap := make(map[uint32]struct{})
	for _, intf := range oldInterfaces {
		intfMap[intf] = struct{}{}
	}
	for _, intf := range newInterfaces {
		if _, has := intfMap[intf]; !has {
			added = append(added, intf)
		} else {
			delete(intfMap, intf)
		}
	}
	for intf := range intfMap {
		removed = append(removed, intf)
	}
	return added, removed
}

// ResolveCreatedInterface configures new interface for every ACL found in cache
func (plugin *ACLConfigurator) ResolveCreatedInterface(ifName string, ifIdx uint32) error {
	plugin.log.Debugf("ACL configurator: resolving new interface %v", ifName)

	// Iterate over cache in order to find out where the interface is used
	var wasErr error
	for entryIdx, aclCacheEntry := range plugin.ifCache {
		if aclCacheEntry.ifName == ifName {
			var ifIndices []uint32
			switch aclCacheEntry.ifAttr {
			case L2:
				if err := plugin.aclHandler.SetMacIPAclToInterface(aclCacheEntry.aclID, append(ifIndices, ifIdx)); err != nil {
					plugin.log.Error(err)
					wasErr = err
				}
			case INGRESS:
				if err := plugin.aclHandler.SetACLToInterfacesAsIngress(aclCacheEntry.aclID, append(ifIndices, ifIdx)); err != nil {
					plugin.log.Error(err)
					wasErr = err
				}
			case EGRESS:
				if err := plugin.aclHandler.SetACLToInterfacesAsEgress(aclCacheEntry.aclID, append(ifIndices, ifIdx)); err != nil {
					plugin.log.Error(err)
					wasErr = err
				}
			default:
				plugin.log.Warnf("ACL interface is not defined as L2, ingress or egress")
			}
			// Remove from cache
			plugin.log.Debugf("New interface %s (%s) configured for ACL %d, removed from cache",
				ifName, aclCacheEntry.ifAttr, aclCacheEntry.aclID)
			plugin.ifCache = append(plugin.ifCache[:entryIdx], plugin.ifCache[entryIdx+1:]...)
		}
	}

	plugin.log.Debugf("ACL configurator: new interface %v resolution done", ifName)

	return wasErr
}

// ResolveDeletedInterface puts removed interface to cache, including acl index. Note: it's not needed to remove ACL
// from interface manually, VPP handles it itself and such an behavior would cause errors (ACLs cannot be dumped
// from non-existing interface)
func (plugin *ACLConfigurator) ResolveDeletedInterface(ifName string, ifIdx uint32) error {
	plugin.log.Debugf("ACL configurator: resolving deleted interface %v", ifName)

	var wasErr error

	// L3/L4 ingress/egress ACLs
	for _, aclName := range plugin.l3l4AclIndexes.GetMapping().ListNames() {
		aclIdx, aclData, found := plugin.l3l4AclIndexes.LookupIdx(aclName)
		if !found {
			plugin.log.Warnf("ACL %v not found in the mapping", aclName)
			continue
		}
		vppAclIdx := aclIdx - 1
		if ifaces := aclData.GetInterfaces(); ifaces != nil {
			// Look over ingress interfaces
			for _, iface := range ifaces.Ingress {
				if iface == ifName {
					plugin.ifCache = append(plugin.ifCache, &ACLIfCacheEntry{
						ifName: ifName,
						aclID:  vppAclIdx,
						ifAttr: INGRESS,
					})
				}
			}
			// Look over egress interfaces
			for _, iface := range ifaces.Egress {
				if iface == ifName {
					plugin.ifCache = append(plugin.ifCache, &ACLIfCacheEntry{
						ifName: ifName,
						aclID:  vppAclIdx,
						ifAttr: EGRESS,
					})
				}
			}
		}
	}
	// L2 ACLs
	for _, aclName := range plugin.l2AclIndexes.GetMapping().ListNames() {
		aclIdx, aclData, found := plugin.l2AclIndexes.LookupIdx(aclName)
		if !found {
			plugin.log.Warnf("ACL %v not found in the mapping", aclName)
			continue
		}
		vppAclIdx := aclIdx - 1
		if ifaces := aclData.GetInterfaces(); ifaces != nil {
			// Look over ingress interfaces
			for _, ingressIf := range ifaces.Ingress {
				if ingressIf == ifName {
					plugin.ifCache = append(plugin.ifCache, &ACLIfCacheEntry{
						ifName: ifName,
						aclID:  vppAclIdx,
						ifAttr: L2,
					})
				}
			}
		}
	}

	plugin.log.Debugf("ACL configurator: resolution done for removed interface %v", ifName)

	return wasErr
}

// Returns a list of interfaces configurable on the ACL. If interface is missing, put it to the cache. It will be
// configured when available
func (plugin *ACLConfigurator) getOrCacheInterfaces(interfaces []string, acl uint32, attr string) (configurableIfs []uint32) {
	for _, name := range interfaces {
		ifIdx, _, found := plugin.ifIndexes.LookupIdx(name)
		if !found {
			// Put interface to cache
			plugin.ifCache = append(plugin.ifCache, &ACLIfCacheEntry{
				ifName: name,
				aclID:  acl,
				ifAttr: attr,
			})
			plugin.log.Debugf("Interface %s (%s) not found for ACL %v, moving to cache", name, attr, acl)
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
		if rule.GetMatch() == nil {
			plugin.log.Warnf("Rule %v from acl %v does not contain match", index, aclName)
			continue
		}
		if rule.GetMatch().GetIpRule() != nil {
			validL3L4Rules = append(validL3L4Rules, rule)
		}
		if rule.GetMatch().GetMacipRule() != nil {
			validL2Rules = append(validL2Rules, rule)
		}
	}
	if len(validL3L4Rules) > 0 && len(validL2Rules) > 0 {
		plugin.log.Errorf("Acl %v contains even L2 rules and L3/L4 rules. This case is not supported yet, only L3/L4 rules will be resolved",
			aclName)
		return validL3L4Rules, false
	} else if len(validL3L4Rules) > 0 {
		return validL3L4Rules, false
	} else {
		return validL2Rules, true
	}
}
