//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package descriptor

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/aclidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	ifdescriptor "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/acl"
)

const (
	AclDescriptorName = "vpp-acl"
)

type AclDescriptor struct {
	// dependencies
	log        logging.Logger
	aclHandler vppcalls.ACLVppAPI

	// runtime
	ifPlugin ifplugin.API
}

func NewAclDescriptor(aclHandler vppcalls.ACLVppAPI, ifPlugin ifplugin.API,
	logger logging.PluginLogger) *AclDescriptor {
	return &AclDescriptor{
		log:        logger.NewLogger("acl-descriptor"),
		ifPlugin:   ifPlugin,
		aclHandler: aclHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *AclDescriptor) GetDescriptor() *adapter.AclDescriptor {
	return &adapter.AclDescriptor{
		Name:        AclDescriptorName,
		NBKeyPrefix: acl.Prefix,
		KeySelector: func(key string) bool {
			return strings.HasPrefix(key, acl.Prefix)
		},
		ValueTypeName: proto.MessageName((*acl.Acl)(nil)),
		KeyLabel: func(key string) string {
			name, _ := acl.ParseNameFromKey(key)
			return name
		},
		ValueComparator: d.EquivalentACLs,
		WithMetadata:    true,
		MetadataMapFactory: func() idxmap.NamedMappingRW {
			return aclidx.NewAclIndex(d.log, "vpp-acl-index")
		},
		Add:                d.Add,
		Delete:             d.Delete,
		Modify:             d.Modify,
		ModifyWithRecreate: d.ModifyWithRecreate,
		IsRetriableFailure: d.IsRetriableFailure,
		Dependencies:       d.Dependencies,
		DerivedValues:      d.DerivedValues,
		Dump:               d.Dump,
		DumpDependencies:   []string{ifdescriptor.InterfaceDescriptorName},
	}
}

func (d *AclDescriptor) EquivalentACLs(key string, oldACL, newACL *acl.Acl) bool {

	return proto.Equal(oldACL, newACL)
}

var nonRetriableErrs = []error{}

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (d *AclDescriptor) IsRetriableFailure(err error) bool {
	for _, e := range nonRetriableErrs {
		if err == e {
			return false
		}
	}
	return true
}

func (d *AclDescriptor) Add(key string, acl *acl.Acl) (metadata *aclidx.AclMetadata, err error) {
	if len(acl.Rules) == 0 {
		return nil, errors.Errorf("failed to configure ACL %s, no rules to set", acl.Name)
	}

	rules, isL2MacIP := d.validateRules(acl.Name, acl.Rules)

	// Configure ACL rules.
	var vppACLIndex uint32
	if isL2MacIP {
		vppACLIndex, err = d.aclHandler.AddMacIPACL(rules, acl.Name)
		if err != nil {
			return nil, errors.Errorf("failed to add MAC IP ACL %s: %v", acl.Name, err)
		}
	} else {
		vppACLIndex, err = d.aclHandler.AddIPACL(rules, acl.Name)
		if err != nil {
			return nil, errors.Errorf("failed to add IP ACL %s: %v", acl.Name, err)
		}
	}

	// Set ACL to interfaces.
	/*if ifaces := acl.GetInterfaces(); ifaces != nil {
		if isL2MacIP {
			aclIfIndices := d.getOrCacheInterfaces(acl.Interfaces.Ingress, vppACLIndex, L2)
			err := d.aclHandler.SetMacIPACLToInterface(vppACLIndex, aclIfIndices)
			if err != nil {
				return nil, errors.Errorf("failed to set MAC IP ACL %s to interface(s) %v: %v",
					acl.Name, acl.Interfaces.Ingress, err)
			}
		} else {
			aclIfInIndices := d.getOrCacheInterfaces(acl.Interfaces.Ingress, vppACLIndex, INGRESS)
			err = d.aclHandler.SetACLToInterfacesAsIngress(vppACLIndex, aclIfInIndices)
			if err != nil {
				return nil, errors.Errorf("failed to set IP ACL %s to interface(s) %v as ingress: %v",
					acl.Name, acl.Interfaces.Ingress, err)
			}
			aclIfEgIndices := d.getOrCacheInterfaces(acl.Interfaces.Egress, vppACLIndex, EGRESS)
			err = d.aclHandler.SetACLToInterfacesAsEgress(vppACLIndex, aclIfEgIndices)
			if err != nil {
				return nil, errors.Errorf("failed to set IP ACL %s to interface(s) %v as egress: %v",
					acl.Name, acl.Interfaces.Ingress, err)
			}
		}
	}*/

	metadata = &aclidx.AclMetadata{
		Index: vppACLIndex,
		L2:    isL2MacIP,
	}
	return metadata, nil
}

// validateRules provided in ACL. Every rule has to contain actions and matches.
// Current limitation: L2 and L3/4 have to be split to different ACLs and
// there cannot be L2 rules and L3/4 rules in the same ACL.
func (c *AclDescriptor) validateRules(aclName string, rules []*acl.Acl_Rule) ([]*acl.Acl_Rule, bool) {
	var validL3L4Rules []*acl.Acl_Rule
	var validL2Rules []*acl.Acl_Rule

	for index, rule := range rules {
		if rule.GetMatch() == nil {
			c.log.Warnf("invalid ACL %s: rule %d does not contain match", aclName, index)
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
		c.log.Warnf("ACL %s contains L2 rules and L3/L4 rules as well. This case is not supported, only L3/L4 rules will be resolved",
			aclName)
		return validL3L4Rules, false
	} else if len(validL3L4Rules) > 0 {
		return validL3L4Rules, false
	} else {
		return validL2Rules, true
	}
}

func (d *AclDescriptor) Delete(key string, acl *acl.Acl, metadata *aclidx.AclMetadata) error {
	if metadata.L2 {
		/*if acl.GetInterfaces() != nil {
			err := c.aclHandler.RemoveMacIPIngressACLFromInterfaces(vppACLIndex, c.getInterfaces(acl.Interfaces.Ingress))
			if err != nil {
				return errors.Errorf("failed to remove MAC IP interfaces from ACL %s: %v",
					acl.AclName, err)
			}
		}*/
		// Remove ACL L2.
		err := d.aclHandler.DeleteMacIPACL(metadata.Index)
		if err != nil {
			return errors.Errorf("failed to delete MACIP ACL %s: %v", acl.Name, err)
		}
		// Unregister.
		//d.l2AclIndexes.UnregisterName(acl.AclName)
		d.log.Debugf("MACIP ACL %s deleted", acl.Name)
	} else {
		// Remove interfaces.
		//vppACLIndex := agentL3L4AclIndex - 1
		/*if acl.Interfaces != nil {
			err = c.aclHandler.RemoveIPIngressACLFromInterfaces(vppACLIndex, c.getInterfaces(acl.Interfaces.Ingress))
			if err != nil {
				return errors.Errorf("failed to remove IP ingress interfaces from ACL %s: %v",
					acl.AclName, err)
			}

			err = c.aclHandler.RemoveIPEgressACLFromInterfaces(vppACLIndex, c.getInterfaces(acl.Interfaces.Egress))
			if err != nil {
				return errors.Errorf("failed to remove IP egress interfaces from ACL %s: %v",
					acl.AclName, err)
			}
		}*/
		// Remove ACL L3/L4.
		err := d.aclHandler.DeleteIPACL(metadata.Index)
		if err != nil {
			return errors.Errorf("failed to delete IP ACL %s: %v", acl.Name, err)
		}
		// Unregister.
		//c.l3l4AclIndexes.UnregisterName(acl.AclName)
		d.log.Debugf("IP ACL %s deleted", acl.Name)
	}
	return nil
}

func (d *AclDescriptor) Modify(key string, oldACL, newACL *acl.Acl, oldMetadata *aclidx.AclMetadata) (newMetadata *aclidx.AclMetadata, err error) {
	// Validate rules.
	rules, isL2MacIP := d.validateRules(newACL.Name, newACL.Rules)
	/*var vppACLIndex uint32
	if isL2MacIP {
		agentACLIndex, _, found := c.l2AclIndexes.LookupIdx(oldACL.AclName)
		if !found {
			return errors.Errorf("cannot modify IP MAC ACL %s, index not found in the mapping", oldACL.AclName)
		}
		// Index used in VPP = index used in mapping - 1
		vppACLIndex = agentACLIndex - 1
	} else {
		agentACLIndex, _, found := c.l3l4AclIndexes.LookupIdx(oldACL.AclName)
		if !found {
			return errors.Errorf("cannot modify IP ACL %s, index not found in the mapping", oldACL.AclName)
		}
		vppACLIndex = agentACLIndex - 1
	}*/
	if isL2MacIP {
		// L2 ACL
		err := d.aclHandler.ModifyMACIPACL(oldMetadata.Index, rules, newACL.Name)
		if err != nil {
			return nil, errors.Errorf("failed to modify MACIP ACL %s: %v", newACL.Name, err)
		}
	} else {
		// L3/L4 ACL can be modified directly.
		err := d.aclHandler.ModifyIPACL(oldMetadata.Index, rules, newACL.Name)
		if err != nil {
			return nil, errors.Errorf("failed to modify IP ACL %s: %v", newACL.Name, err)
		}
	}

	// Update interfaces.
	/*if isL2MacIP {
		// Remove L2 ACL from old interfaces.
		if oldACL.Interfaces != nil {
			err := c.aclHandler.RemoveMacIPIngressACLFromInterfaces(vppACLIndex, c.getInterfaces(oldACL.Interfaces.Ingress))
			if err != nil {
				return errors.Errorf("failed to remove MAC IP ingress interfaces from ACL %s: %v",
					oldACL.AclName, err)
			}
		}
		// Put L2 ACL to new interfaces.
		if newACL.Interfaces != nil {
			aclMacInterfaces := c.getOrCacheInterfaces(newACL.Interfaces.Ingress, vppACLIndex, L2)
			err := c.aclHandler.SetMacIPACLToInterface(vppACLIndex, aclMacInterfaces)
			if err != nil {
				return errors.Errorf("failed to set MAC IP ingress interfaces to ACL %s: %v",
					newACL.AclName, err)
			}
		}
	} else {
		aclOldInInterfaces := c.getInterfaces(oldACL.Interfaces.Ingress)
		aclOldEgInterfaces := c.getInterfaces(oldACL.Interfaces.Egress)
		aclNewInInterfaces := c.getOrCacheInterfaces(newACL.Interfaces.Ingress, vppACLIndex, INGRESS)
		aclNewEgInterfaces := c.getOrCacheInterfaces(newACL.Interfaces.Egress, vppACLIndex, EGRESS)
		addedInInterfaces, removedInInterfaces := diffInterfaces(aclOldInInterfaces, aclNewInInterfaces)
		addedEgInterfaces, removedEgInterfaces := diffInterfaces(aclOldEgInterfaces, aclNewEgInterfaces)

		if len(removedInInterfaces) > 0 {
			err := c.aclHandler.RemoveIPIngressACLFromInterfaces(vppACLIndex, removedInInterfaces)
			if err != nil {
				return errors.Errorf("failed to remove IP ingress interfaces from ACL %s: %v",
					oldACL.AclName, err)
			}
		}
		if len(removedEgInterfaces) > 0 {
			err := c.aclHandler.RemoveIPEgressACLFromInterfaces(vppACLIndex, removedEgInterfaces)
			if err != nil {
				return errors.Errorf("failed to remove IP egress interfaces from ACL %s: %v",
					oldACL.AclName, err)
			}
		}
		if len(addedInInterfaces) > 0 {
			err := c.aclHandler.SetACLToInterfacesAsIngress(vppACLIndex, addedInInterfaces)
			if err != nil {
				return errors.Errorf("failed to set IP ingress interfaces to ACL %s: %v",
					newACL.AclName, err)
			}
		}
		if len(addedEgInterfaces) > 0 {
			err := c.aclHandler.SetACLToInterfacesAsEgress(vppACLIndex, addedEgInterfaces)
			if err != nil {
				return errors.Errorf("failed to add IP egress interfaces to ACL %s: %v",
					oldACL.AclName, err)
			}
		}
	}*/

	newMetadata = &aclidx.AclMetadata{
		Index: oldMetadata.Index,
		L2:    isL2MacIP,
	}
	return newMetadata, nil
}

func (d *AclDescriptor) ModifyWithRecreate(key string, oldACL, newACL *acl.Acl, metadata *aclidx.AclMetadata) bool {
	var hasL2 bool
	for _, rule := range oldACL.Rules {
		if rule.GetMatch().GetMacipRule() != nil {
			hasL2 = true
		} else if rule.GetMatch().GetIpRule() != nil && hasL2 {
			return true
		}
	}
	return false
}

/*func (d *AclDescriptor) Update(key string, value *acl.Acl, metadata *aclidx.AclMetadata) error {

}*/

func (d *AclDescriptor) Dependencies(key string, value *acl.Acl) []api.Dependency {
	return nil
}

func (d *AclDescriptor) DerivedValues(key string, value *acl.Acl) []api.KeyValuePair {
	return nil
}

func (d *AclDescriptor) Dump(correlate []adapter.AclKVWithMetadata) (dump []adapter.AclKVWithMetadata, err error) {
	ipACLs, err := d.aclHandler.DumpIPACL(d.ifPlugin.GetInterfaceIndex())
	if err != nil {
		return nil, errors.Errorf("failed to dump IP ACLs: %v", err)
	}
	macipACLs, err := d.aclHandler.DumpMACIPACL(d.ifPlugin.GetInterfaceIndex())
	if err != nil {
		return nil, errors.Errorf("failed to dump MAC IP ACLs: %v", err)
	}

	for _, ipACL := range ipACLs {
		dump = append(dump, adapter.AclKVWithMetadata{
			Key:   acl.Key(ipACL.ACL.Name),
			Value: ipACL.ACL,
			Metadata: &aclidx.AclMetadata{
				Index: ipACL.Meta.Index,
			},
			Origin: api.FromNB,
		})
	}
	for _, macipACL := range macipACLs {
		dump = append(dump, adapter.AclKVWithMetadata{
			Key:   acl.Key(macipACL.ACL.Name),
			Value: macipACL.ACL,
			Metadata: &aclidx.AclMetadata{
				Index: macipACL.Meta.Index,
				L2:    true,
			},
			Origin: api.FromNB,
		})
	}
	return
}
