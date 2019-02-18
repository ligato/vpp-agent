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
	"sort"
	"strings"

	"github.com/ligato/cn-infra/logging"
	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/descriptor/cache"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/vppcalls"
	"github.com/pkg/errors"
)

const (
	// PolicyDescriptorName is the name of the descriptor for VPP policies
	PolicyDescriptorName = "vpp-sr-policy"

	// dependency labels
	atLeastOneSegmentListExistsDep = "sr-at-least-one-segment-list-exists"
)

// PolicyDescriptor teaches KVScheduler how to configure VPP SRv6 policies.
type PolicyDescriptor struct {
	// dependencies
	log                         logging.Logger
	srHandler                   vppcalls.SRv6VppAPI
	scheduler                   scheduler.KVScheduler
	vppIndexSeq                 *gaplessSequence
	policyInfoCache             cache.PolicyInfoCache
	policyIndexCache            *cache.PolicyIndexCache
	policySegmentListIndexCache *cache.PolicySegmentListIndexCache
	slDescriptor                *PolicySegmentListDescriptor
}

// NewPolicyDescriptor creates a new instance of the Srv6 policy descriptor.
func NewPolicyDescriptor(srHandler vppcalls.SRv6VppAPI, scheduler scheduler.KVScheduler, log logging.PluginLogger,
	policyInfoCache cache.PolicyInfoCache, policyIndexCache *cache.PolicyIndexCache,
	policySegmentListIndexCache *cache.PolicySegmentListIndexCache, slDescriptor *PolicySegmentListDescriptor) *PolicyDescriptor {
	return &PolicyDescriptor{
		log:                         log.NewLogger("policy-descriptor"),
		srHandler:                   srHandler,
		scheduler:                   scheduler,
		vppIndexSeq:                 newSequence(),
		policyInfoCache:             policyInfoCache,
		policyIndexCache:            policyIndexCache,
		policySegmentListIndexCache: policySegmentListIndexCache,
		slDescriptor:                slDescriptor,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *PolicyDescriptor) GetDescriptor() *adapter.PolicyDescriptor {
	return &adapter.PolicyDescriptor{
		Name:            PolicyDescriptorName,
		NBKeyPrefix:     srv6.ModelPolicy.KeyPrefix(),
		ValueTypeName:   srv6.ModelPolicy.ProtoName(),
		KeySelector:     srv6.ModelPolicy.IsKeyValid,
		KeyLabel:        srv6.ModelPolicy.StripKeyPrefix,
		ValueComparator: d.EquivalentPolicies,
		Validate:        d.Validate,
		Create:          d.Create,
		Delete:          d.Delete,
		Update:          d.Update,
		Dependencies:    d.Dependencies,
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
	return nil
}

// Create creates new Policy into VPP using VPP's binary api
func (d *PolicyDescriptor) Create(key string, policy *srv6.Policy) (metadata interface{}, err error) {
	bsid, _ := ParseIPv6(policy.GetBsid()) // already validated
	sl, slDescChoice, err := d.segmentListForPolicyCreation(policy.GetBsid())
	if err != nil {
		return nil, errors.Errorf("can't get segment list for policy creation: %v", err)
	}
	err = d.srHandler.AddPolicy(bsid, policy, sl)
	if err != nil {
		return nil, errors.Errorf("failed to write policy %s with first segment %s: %v", bsid.String(), sl.Segments, err)
	}

	// FIXME now i'm just guessing index that will Policy get in VPP (= being dependent on internal implementation inside VPP) and that is just very fragile -> API should tell this but it doesn't!
	d.policyIndexCache.Put(policy, d.vppIndexSeq.nextID())
	if slDescChoice { // update SL index (SL create didn't created anything in VPP so no index could be retrieved, but now the SL is in VPP)
		index, err := d.srHandler.RetrievePolicySegmentIndex(sl)
		if err != nil {
			return nil, errors.Errorf("can't retrieve from VPP the index of created segment list (first SL choosen by SL descriptor): %v", err)
		}
		d.policySegmentListIndexCache.Put(sl, index)
	}
	return nil, nil
}

// segmentListForPolicyCreation retrieves policy segment list that is needed for policy creation
func (d *PolicyDescriptor) segmentListForPolicyCreation(policyBSID string) (*srv6.PolicySegmentList, bool, error) {
	policyBSID = strings.TrimSpace(strings.ToLower(policyBSID))
	pInfo, exists := d.policyInfoCache[policyBSID]
	if !exists { // policy creation is called first -> choose SL that should be used policy creation
		slList, _, err := allSegmentListsInOnePolicy(policyBSID, d.scheduler)
		if err != nil {
			return nil, false, err
		}
		if len(slList) == 0 {
			return nil, false, errors.Errorf("can't find policy segment list with parent policy %v", policyBSID)
		}
		d.policyInfoCache[policyBSID] = &cache.PolicyInfo{
			PolicyCreationSL: slList[0],
			//PolicyCreationSLIndex:   d.slDescriptor.NextSegmentListIndex(), // calling internal index generator of segment list descriptor
			PolicyCreationSLCreated: false, // this call (policy creation) is called before SL creation
			SLCounter:               1,     // normally only creation of SL can change this counter, but creation of policy also uses SL so SL count after policy creation is 1
		}
		return slList[0], false, nil
	}

	// segment list creation called first -> using it's choose of SL for policy creation (Note: it choosed itself)
	if pInfo.PolicyCreationSL == nil {
		return nil, false, errors.Errorf("policy segment list for policy creation should be choosen by previously "+
			"called SL creation method, but wasn't (parent policy %v)", policyBSID)
	}
	return pInfo.PolicyCreationSL, true, nil
}

// Delete removes Policy from VPP using VPP's binary api
func (d *PolicyDescriptor) Delete(key string, policy *srv6.Policy, metadata interface{}) error {
	bsid, _ := ParseIPv6(policy.GetBsid())                 // already validated
	if err := d.srHandler.DeletePolicy(bsid); err != nil { // expecting that policy delete will also delete policy segments in vpp
		return errors.Errorf("failed to delete policy %s: %v", bsid.String(), err)
	}
	delete(d.policyInfoCache, strings.TrimSpace(strings.ToLower(policy.GetBsid())))
	index, exists := d.policyIndexCache.Get(policy)
	if !exists {
		d.log.Warn("can't release index for policy %v", policy)
	} else {
		d.vppIndexSeq.delete(index)
		d.policyIndexCache.Remove(policy)
	}
	return nil
}

// Update updates Policy in VPP using VPP's binary api. Due to VPP binary api limitations, rearranging of other objects
// in VPP can occur, but in the end everything should be logically the same as before and with updated Policy.
func (d *PolicyDescriptor) Update(key string, oldPolicy, newPolicy *srv6.Policy, oldMetadata interface{}) (newMetadata interface{}, err error) {
	// get segment lists
	bsid := strings.TrimSpace(strings.ToLower(oldPolicy.GetBsid()))
	slList, slKeys, err := allSegmentListsInOnePolicy(bsid, d.scheduler)
	if err != nil {
		return nil, errors.Errorf("can't retrieve segment lists for policy recreation: %v", err)
	}

	// remove segment list (policy delete removes also segment lists in VPP, but some stuff on vpp-agent side won't be
	// updated properly if SL delete is not called, i.e. SL index provider(gaplessindex provider))
	for i, sl := range slList {
		err = d.slDescriptor.Delete(slKeys[i], sl, nil)
		if err != nil {
			return nil, errors.Errorf("can't recreate policy due to delete problem of policy segment list %v: %v", sl, err)
		}
	}

	// recreate policy
	err = d.Delete(key, oldPolicy, oldMetadata)
	if err != nil {
		return nil, errors.Errorf("can't recreate policy due to policy delete problems: %v", err)
	}
	newMetadata, err = d.Create(key, newPolicy)
	if err != nil {
		return nil, errors.Errorf("can't recreate policy due to policy creation problems: %v", err)
	}

	// add segments lists again because they got lost in recreation of policy
	for i, sl := range slList {
		_, err = d.slDescriptor.Create(slKeys[i], sl)
		if err != nil {
			return nil, errors.Errorf("can't apply segment list %v as part of policy modification: %v", sl, err)
		}
	}
	return newMetadata, nil
}

// Dependencies defines dependencies of Policy descriptor
func (d *PolicyDescriptor) Dependencies(key string, policy *srv6.Policy) (dependencies []scheduler.Dependency) {
	dependencies = append(dependencies, scheduler.Dependency{
		Label: atLeastOneSegmentListExistsDep,
		AnyOf: func(key string) bool { // exists at least one segment list for policy (policy with no segment list can't exists)
			policyBSID, _, isSegmentListKey := srv6.ParsePolicySegmentList(key)
			return isSegmentListKey && strings.ToLower(policyBSID) == strings.ToLower(policy.Bsid)
		},
	})
	return dependencies
}

// EquivalentPolicies determines whether 2 policies are logically equal. This comparison takes into consideration also
// semantics that couldn't be modeled into proto models (i.e. SID is IPv6 address and not only string)
func (d *PolicyDescriptor) EquivalentPolicies(key string, oldPolicy, newPolicy *srv6.Policy) bool {
	return oldPolicy.FibTableId == newPolicy.FibTableId &&
		equivalentSIDs(oldPolicy.Bsid, newPolicy.Bsid) &&
		oldPolicy.SprayBehaviour == newPolicy.SprayBehaviour &&
		oldPolicy.SrhEncapsulation == newPolicy.SrhEncapsulation
}

// gaplessSequence emulates sequence indexes grabbing for Policy segments inside VPP // FIXME this is poor VPP API, correct way is tha API should tell as choosen index at Policy segment creation
type gaplessSequence struct {
	nextfree []uint32
}

func newSequence() *gaplessSequence {
	return &gaplessSequence{
		nextfree: []uint32{0},
	}
}

func (seq *gaplessSequence) nextID() uint32 {
	if len(seq.nextfree) == 1 { // no gaps in sequence
		result := seq.nextfree[0]
		seq.nextfree[0]++
		return result
	}
	// use first gap and then remove it from free IDs list
	result := seq.nextfree[0]
	seq.nextfree = seq.nextfree[1:]
	return result
}

func (seq *gaplessSequence) delete(id uint32) {
	if id >= seq.nextfree[len(seq.nextfree)-1] {
		return // nothing to do because it is not sequenced yet
	}
	// add gap and move it to proper place (gaps with lower id should be used first by finding next ID)
	seq.nextfree = append(seq.nextfree, id)
	sort.Slice(seq.nextfree, func(i, j int) bool { return seq.nextfree[i] < seq.nextfree[j] })
}
