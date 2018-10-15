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
	prototypes "github.com/gogo/protobuf/types"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/aclidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	ifdescriptor "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/acl"
)

const (
	// ACLDescriptorName is descriptor name
	ACLDescriptorName = "vpp-acl"
)

// ACLDescriptor is descriptor for ACL
type ACLDescriptor struct {
	// dependencies
	log        logging.Logger
	aclHandler vppcalls.ACLVppAPI

	// runtime
	ifPlugin ifplugin.API
}

// NewACLDescriptor is constructor for ACL descriptor
func NewACLDescriptor(aclHandler vppcalls.ACLVppAPI, ifPlugin ifplugin.API,
	logger logging.PluginLogger) *ACLDescriptor {
	return &ACLDescriptor{
		log:        logger.NewLogger("acl-descriptor"),
		ifPlugin:   ifPlugin,
		aclHandler: aclHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *ACLDescriptor) GetDescriptor() *adapter.ACLDescriptor {
	return &adapter.ACLDescriptor{
		Name:        ACLDescriptorName,
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
			return aclidx.NewACLIndex(d.log, "vpp-acl-index")
		},
		Add:                d.Add,
		Delete:             d.Delete,
		Modify:             d.Modify,
		ModifyWithRecreate: d.ModifyWithRecreate,
		IsRetriableFailure: d.IsRetriableFailure,
		DerivedValues:      d.DerivedValues,
		Dump:               d.Dump,
		DumpDependencies:   []string{ifdescriptor.InterfaceDescriptorName},
	}
}

// EquivalentACLs compares two ACLs
func (d *ACLDescriptor) EquivalentACLs(key string, oldACL, newACL *acl.Acl) bool {
	// check if ACL name changed
	if oldACL.Name != newACL.Name {
		return false
	}

	// check if rules changed (order matters)
	if len(oldACL.Rules) != len(newACL.Rules) {
		return false
	}
	for i := 0; i < len(oldACL.Rules); i++ {
		if !proto.Equal(oldACL.Rules[i], newACL.Rules[i]) {
			return false
		}
	}

	return true
}

var nonRetriableErrs []error

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (d *ACLDescriptor) IsRetriableFailure(err error) bool {
	for _, e := range nonRetriableErrs {
		if err == e {
			return false
		}
	}
	return true
}

// Add configures ACL
func (d *ACLDescriptor) Add(key string, acl *acl.Acl) (metadata *aclidx.ACLMetadata, err error) {
	if len(acl.Rules) == 0 {
		return nil, errors.Errorf("failed to configure ACL %s, no rules to set", acl.Name)
	}

	rules, isL2MacIP := d.validateRules(acl.Name, acl.Rules)

	// Configure ACL rules.
	var vppACLIndex uint32
	if isL2MacIP {
		vppACLIndex, err = d.aclHandler.AddMACIPACL(rules, acl.Name)
		if err != nil {
			return nil, errors.Errorf("failed to add MACIP ACL %s: %v", acl.Name, err)
		}
	} else {
		vppACLIndex, err = d.aclHandler.AddACL(rules, acl.Name)
		if err != nil {
			return nil, errors.Errorf("failed to add IP ACL %s: %v", acl.Name, err)
		}
	}

	metadata = &aclidx.ACLMetadata{
		Index: vppACLIndex,
		L2:    isL2MacIP,
	}
	return metadata, nil
}

// validateRules provided in ACL. Every rule has to contain actions and matches.
// Current limitation: L2 and L3/4 have to be split to different ACLs and
// there cannot be L2 rules and L3/4 rules in the same ACL.
func (d *ACLDescriptor) validateRules(aclName string, rules []*acl.Acl_Rule) ([]*acl.Acl_Rule, bool) {
	var validL3L4Rules []*acl.Acl_Rule
	var validL2Rules []*acl.Acl_Rule

	for _, rule := range rules {
		if rule.GetIpRule() != nil {
			validL3L4Rules = append(validL3L4Rules, rule)
		}
		if rule.GetMacipRule() != nil {
			validL2Rules = append(validL2Rules, rule)
		}
	}
	if len(validL3L4Rules) > 0 && len(validL2Rules) > 0 {
		d.log.Warnf("ACL %s contains L2 rules and L3/L4 rules as well. This case is not supported, only L3/L4 rules will be resolved",
			aclName)
		return validL3L4Rules, false
	} else if len(validL3L4Rules) > 0 {
		return validL3L4Rules, false
	} else {
		return validL2Rules, true
	}
}

// Delete deletes ACL
func (d *ACLDescriptor) Delete(key string, acl *acl.Acl, metadata *aclidx.ACLMetadata) error {
	if metadata.L2 {
		// Remove ACL L2.
		err := d.aclHandler.DeleteMACIPACL(metadata.Index)
		if err != nil {
			return errors.Errorf("failed to delete MACIP ACL %s: %v", acl.Name, err)
		}
	} else {
		// Remove ACL L3/L4.
		err := d.aclHandler.DeleteACL(metadata.Index)
		if err != nil {
			return errors.Errorf("failed to delete IP ACL %s: %v", acl.Name, err)
		}
	}
	return nil
}

// Modify modifies ACL
func (d *ACLDescriptor) Modify(key string, oldACL, newACL *acl.Acl, oldMetadata *aclidx.ACLMetadata) (newMetadata *aclidx.ACLMetadata, err error) {
	// Validate rules.
	rules, isL2MacIP := d.validateRules(newACL.Name, newACL.Rules)

	if isL2MacIP {
		// L2 ACL
		err := d.aclHandler.ModifyMACIPACL(oldMetadata.Index, rules, newACL.Name)
		if err != nil {
			return nil, errors.Errorf("failed to modify MACIP ACL %s: %v", newACL.Name, err)
		}
	} else {
		// L3/L4 ACL can be modified directly.
		err := d.aclHandler.ModifyACL(oldMetadata.Index, rules, newACL.Name)
		if err != nil {
			return nil, errors.Errorf("failed to modify IP ACL %s: %v", newACL.Name, err)
		}
	}

	newMetadata = &aclidx.ACLMetadata{
		Index: oldMetadata.Index,
		L2:    isL2MacIP,
	}
	return newMetadata, nil
}

// ModifyWithRecreate checks if modification requires recreation
func (d *ACLDescriptor) ModifyWithRecreate(key string, oldACL, newACL *acl.Acl, metadata *aclidx.ACLMetadata) bool {
	var hasL2 bool
	for _, rule := range oldACL.Rules {
		if rule.GetMacipRule() != nil {
			hasL2 = true
		} else if rule.GetIpRule() != nil && hasL2 {
			return true
		}
	}
	return false
}

// DerivedValues returns list of derived values for ACL.
func (d *ACLDescriptor) DerivedValues(key string, value *acl.Acl) (derived []api.KeyValuePair) {
	for _, ifName := range value.GetInterfaces().GetIngress() {
		derived = append(derived, api.KeyValuePair{
			Key:   acl.ToInterfaceKey(value.Name, ifName, acl.IngressFlow),
			Value: &prototypes.Empty{},
		})
	}
	for _, ifName := range value.GetInterfaces().GetEgress() {
		derived = append(derived, api.KeyValuePair{
			Key:   acl.ToInterfaceKey(value.Name, ifName, acl.EgressFlow),
			Value: &prototypes.Empty{},
		})
	}
	return derived
}

// Dump returns list of dumped ACLs with metadata
func (d *ACLDescriptor) Dump(correlate []adapter.ACLKVWithMetadata) (dump []adapter.ACLKVWithMetadata, err error) {
	ipACLs, err := d.aclHandler.DumpACL()
	if err != nil {
		return nil, errors.Errorf("failed to dump IP ACLs: %v", err)
	}
	macipACLs, err := d.aclHandler.DumpMACIPACL()
	if err != nil {
		return nil, errors.Errorf("failed to dump MAC IP ACLs: %v", err)
	}

	for _, ipACL := range ipACLs {
		dump = append(dump, adapter.ACLKVWithMetadata{
			Key:   acl.Key(ipACL.ACL.Name),
			Value: ipACL.ACL,
			Metadata: &aclidx.ACLMetadata{
				Index: ipACL.Meta.Index,
			},
			Origin: api.FromNB,
		})
	}
	for _, macipACL := range macipACLs {
		dump = append(dump, adapter.ACLKVWithMetadata{
			Key:   acl.Key(macipACL.ACL.Name),
			Value: macipACL.ACL,
			Metadata: &aclidx.ACLMetadata{
				Index: macipACL.Meta.Index,
				L2:    true,
			},
			Origin: api.FromNB,
		})
	}
	return
}
