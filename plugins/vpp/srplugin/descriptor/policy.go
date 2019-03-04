// Copyright (c) 2019 Bell Canada, Pantheon Technologies and/or its affiliates.
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

package descriptor

import (
	"github.com/ligato/cn-infra/logging"
	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/vppcalls"
	"github.com/pkg/errors"
)

// PolicyDescriptorName is the name of the descriptor for VPP policies
const PolicyDescriptorName = "vpp-sr-policy"

// PolicyDescriptor teaches KVScheduler how to configure VPP SRv6 policies.
//
// Implementation note: according to https://tools.ietf.org/html/draft-ietf-spring-segment-routing-policy-02#section-2.6,
// (weight,segments) tuple is not unique identification of candidate path(segment list) in policy (VPP also allows to
// have multiple segment lists that have the same (weight,segments) but differs in index (in ietf material they call it
// "Discriminator of a Candidate Path")). However, in validation and equivalency stage of vpp-agent there is no way how
// to get from VPP the index (validation precedes creation call and VPP doesn't have it). As result, restriction must be
// made, (weight,segments) is unique within given Policy. The drawback is that you can't create multiple segment lists
// with the same (weight,segments), but that is not issue if you don't want to explicitly rely on special attributes
// used by tie-breaking rules from https://tools.ietf.org/html/draft-ietf-spring-segment-routing-policy-02#section-2.9
// (i.e. VPP-internal index). These special attributes can't be explicitly set in VPP, so can't rely on them anyway. So
// we should not miss any client use case for current VPP implementation (by removing duplicated segment list we don't
// break any use case).
type PolicyDescriptor struct {
	// dependencies
	log       logging.Logger
	srHandler vppcalls.SRv6VppAPI
}

// PolicyMetadata are Policy-related metadata that KVscheduler bundles with Policy data. They are served by KVScheduler in Create/Update descriptor methods.
type PolicyMetadata struct {
	segmentListIndexes map[*srv6.Policy_SegmentList]uint32
}

// NewPolicyDescriptor creates a new instance of the Srv6 policy descriptor.
func NewPolicyDescriptor(srHandler vppcalls.SRv6VppAPI, log logging.PluginLogger) *PolicyDescriptor {
	return &PolicyDescriptor{
		log:       log.NewLogger("policy-descriptor"),
		srHandler: srHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *PolicyDescriptor) GetDescriptor() *adapter.PolicyDescriptor {
	return &adapter.PolicyDescriptor{
		Name:               PolicyDescriptorName,
		NBKeyPrefix:        srv6.ModelPolicy.KeyPrefix(),
		ValueTypeName:      srv6.ModelPolicy.ProtoName(),
		KeySelector:        srv6.ModelPolicy.IsKeyValid,
		KeyLabel:           srv6.ModelPolicy.StripKeyPrefix,
		ValueComparator:    d.EquivalentPolicies,
		Validate:           d.Validate,
		Create:             d.Create,
		Delete:             d.Delete,
		Update:             d.Update,
		UpdateWithRecreate: d.UpdateWithRecreate,
		WithMetadata:       true,
	}
}

// Validate validates VPP policies.
func (d *PolicyDescriptor) Validate(key string, policy *srv6.Policy) error {
	if policy.GetFibTableId() < 0 {
		return scheduler.NewInvalidValueError(errors.Errorf("fibtableid can't be lower than zero, input value %v)", policy.GetFibTableId()), "fibtableid")
	}
	_, err := ParseIPv6(policy.GetBsid())
	if err != nil {
		return scheduler.NewInvalidValueError(errors.Errorf("failed to parse binding sid %s, should be a valid ipv6 address: %v", policy.GetBsid(), err), "bsid")
	}
	if len(policy.SegmentLists) == 0 { // includes nil check
		return scheduler.NewInvalidValueError(errors.New("there must be defined at least one segment list"), "SegmentLists")
	}

	for i, sl := range policy.SegmentLists {
		for _, segment := range sl.Segments {
			_, err := ParseIPv6(segment)
			if err != nil {
				return scheduler.NewInvalidValueError(errors.Errorf("failed to parse segment %s in segments %v, should be a valid ipv6 address: %v", segment, sl.Segments, err), "SegmentLists.segments")
			}
		}

		// checking for segment list duplicity (existence of this check is used in equivalency check too), see PolicyDescriptor's godoc implementation note for more info
		for j, previousSL := range policy.SegmentLists[:i] {
			if d.equivalentSegmentList(previousSL, sl) {
				return scheduler.NewInvalidValueError(errors.Errorf("found duplicated segment list: %+v (list index %v) and %+v (list index %v) ", sl, i, previousSL, j), "SegmentLists")
			}
		}
	}
	return nil
}

// Create creates new Policy into VPP using VPP's binary api
func (d *PolicyDescriptor) Create(key string, policy *srv6.Policy) (metadata interface{}, err error) {
	// add policy (including segment lists)
	err = d.srHandler.AddPolicy(policy) // there exist first segment (validation checked it)
	if err != nil {
		return nil, errors.Errorf("failed to write policy %s with first segment %s: %v", policy.GetBsid(), policy.SegmentLists[0], err)
	}

	// retrieve from VPP indexes of just added Policy/Segment Lists and store it as metadata
	_, slIndexes, err := d.srHandler.RetrievePolicyIndexInfo(policy)
	if err != nil {
		return nil, errors.Errorf("can't retrieve indexes of created srv6 policy with bsid %v : %v", policy.GetBsid(), err)
	}
	metadata = &PolicyMetadata{
		segmentListIndexes: slIndexes,
	}
	return metadata, nil
}

// Delete removes Policy from VPP using VPP's binary api
func (d *PolicyDescriptor) Delete(key string, policy *srv6.Policy, metadata interface{}) error {
	bsid, _ := ParseIPv6(policy.GetBsid())                 // already validated
	if err := d.srHandler.DeletePolicy(bsid); err != nil { // expecting that delete of SB defined policy will also delete policy segments in vpp
		return errors.Errorf("failed to delete policy %s: %v", bsid.String(), err)
	}
	return nil
}

// Update updates Policy in VPP using VPP's binary api. Only changes of segment list handled here. Other updates are
// handle by recreation (see function UpdateWithRecreate)
func (d *PolicyDescriptor) Update(key string, oldPolicy, newPolicy *srv6.Policy, oldMetadata interface{}) (newMetadata interface{}, err error) {
	// compute segment lists for delete (removePool) and segment lists to add (addPool)
	removePool := make([]*srv6.Policy_SegmentList, len(oldPolicy.SegmentLists))
	addPool := make([]*srv6.Policy_SegmentList, 0, len(newPolicy.SegmentLists))
	copy(removePool, oldPolicy.SegmentLists)
	for _, newSL := range newPolicy.SegmentLists {
		found := false
		for i, oldSL := range removePool {
			if d.equivalentSegmentList(oldSL, newSL) {
				removePool = append(removePool[:i], removePool[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			addPool = append(addPool, newSL)
		}
	}

	// add new segment lists not present in oldPolicy
	for _, sl := range addPool {
		if err := d.srHandler.AddPolicySegmentList(sl, newPolicy); err != nil {
			return nil, errors.Errorf("failed update policy: failed to add policy segment %s: %v", newPolicy.GetBsid(), err)
		}
	}

	// remove old segment lists not present in newPolicy
	slIndexes := oldMetadata.(*PolicyMetadata).segmentListIndexes
	for _, sl := range removePool {
		index, exists := slIndexes[sl]
		if !exists {
			return nil, errors.Errorf("failed update policy: failed to find index for segment list "+
				"%+v in policy with bsid %v (metadata segment list indexes: %+v)", sl, oldPolicy.GetBsid(), slIndexes)
		}
		if err := d.srHandler.DeletePolicySegmentList(sl, index, newPolicy); err != nil {
			return nil, errors.Errorf("failed update policy: failed to delete policy segment %s: %v", oldPolicy.GetBsid(), err)
		}
	}

	// update metadata be recreation it from scratch
	_, slIndexes, err = d.srHandler.RetrievePolicyIndexInfo(newPolicy)
	if err != nil {
		return nil, errors.Errorf("can't retrieve indexes of updated srv6 policy with bsid %v : %v", newPolicy.GetBsid(), err)
	}
	newMetadata = &PolicyMetadata{
		segmentListIndexes: slIndexes,
	}

	return newMetadata, nil
}

// UpdateWithRecreate define whether update case should be handled by complete policy recreation
func (d *PolicyDescriptor) UpdateWithRecreate(key string, oldPolicy, newPolicy *srv6.Policy, oldMetadata interface{}) bool {
	return !d.equivalentPolicyAttributes(oldPolicy, newPolicy) // update with recreate only when policy attributes changes because segment list change can be handled more efficiently
}

// EquivalentPolicies determines whether 2 policies are logically equal. This comparison takes into consideration also
// semantics that couldn't be modeled into proto models (i.e. SID is IPv6 address and not only string)
func (d *PolicyDescriptor) EquivalentPolicies(key string, oldPolicy, newPolicy *srv6.Policy) bool {
	return d.equivalentPolicyAttributes(oldPolicy, newPolicy) &&
		d.equivalentSegmentLists(oldPolicy.SegmentLists, newPolicy.SegmentLists)
}

func (d *PolicyDescriptor) equivalentPolicyAttributes(oldPolicy, newPolicy *srv6.Policy) bool {
	return oldPolicy.FibTableId == newPolicy.FibTableId &&
		equivalentSIDs(oldPolicy.Bsid, newPolicy.Bsid) &&
		oldPolicy.SprayBehaviour == newPolicy.SprayBehaviour &&
		oldPolicy.SrhEncapsulation == newPolicy.SrhEncapsulation
}

func (d *PolicyDescriptor) equivalentSegmentLists(oldSLs, newSLs []*srv6.Policy_SegmentList) bool {
	if oldSLs == nil || newSLs == nil {
		return oldSLs == nil && newSLs == nil
	}
	if len(oldSLs) != len(newSLs) {
		return false
	}

	// checking segment lists equality (segment lists with just reordered segment list items are considered equal)
	// we know that:
	// 1. lists have the same length (due to previous checks),
	// 2. there are no segment list equivalence duplicates (due to validation check)
	// => hence if we check that each segment list from newSLs has its equivalent in oldSLs then we got equivalent bijection
	// and that means that they are equal (items maybe reordered but equal). If we find one segment list from newSLs that
	// is not in oldSLs, the SLs are obviously not equal.
	for _, newSL := range newSLs {
		found := false
		for _, oldSL := range oldSLs {
			if d.equivalentSegmentList(oldSL, newSL) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (d *PolicyDescriptor) equivalentSegmentList(oldSL, newSL *srv6.Policy_SegmentList) bool {
	return oldSL.Weight == newSL.Weight && d.equivalentSegments(oldSL.Segments, newSL.Segments)
}

func (d *PolicyDescriptor) equivalentSegments(segments1, segments2 []string) bool {
	if segments1 == nil || segments2 == nil {
		return segments1 == nil && segments2 == nil
	}
	if len(segments1) != len(segments2) {
		return false
	}
	for i := range segments1 {
		if !equivalentSIDs(segments1[i], segments2[i]) {
			return false
		}
	}
	return true
}
