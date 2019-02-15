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
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/descriptor/cache"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/vppcalls"
	"github.com/pkg/errors"
)

const (
	// PolicySegmentListDescriptorName is the name of the descriptor for VPP policy segment lists
	PolicySegmentListDescriptorName = "vpp-sr-policysegmentlist"

	// dependency labels
	parentPolicyExistsDep = "sr-parent-policy-exists"
)

// PolicySegmentListDescriptor teaches KVScheduler how to configure VPP SRv6 policy segment lists.
type PolicySegmentListDescriptor struct {
	// dependencies
	log                         logging.Logger
	srHandler                   vppcalls.SRv6VppAPI
	scheduler                   scheduler.KVScheduler
	policyInfoCache             cache.PolicyInfoCache
	policyIndexCache            *cache.PolicyIndexCache
	policySegmentListIndexCache *cache.PolicySegmentListIndexCache
}

// NewPolicySegmentListDescriptor creates a new instance of the Srv6 policy segment list descriptor.
func NewPolicySegmentListDescriptor(srHandler vppcalls.SRv6VppAPI, scheduler scheduler.KVScheduler, log logging.PluginLogger,
	policyInfoCache cache.PolicyInfoCache, policyIndexCache *cache.PolicyIndexCache,
	policySegmentListIndexCache *cache.PolicySegmentListIndexCache) *PolicySegmentListDescriptor {
	return &PolicySegmentListDescriptor{
		log:                         log.NewLogger("policysegmentlist-descriptor"),
		srHandler:                   srHandler,
		scheduler:                   scheduler,
		policyInfoCache:             policyInfoCache,
		policyIndexCache:            policyIndexCache,
		policySegmentListIndexCache: policySegmentListIndexCache,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *PolicySegmentListDescriptor) GetDescriptor() *adapter.PolicySegmentListDescriptor {
	return &adapter.PolicySegmentListDescriptor{
		Name:            PolicySegmentListDescriptorName,
		NBKeyPrefix:     srv6.ModelPolicySegmentList.KeyPrefix(),
		ValueTypeName:   srv6.ModelPolicySegmentList.ProtoName(),
		KeySelector:     d.IsPolicySegmentListKey,
		KeyLabel:        srv6.ModelPolicySegmentList.StripKeyPrefix,
		ValueComparator: d.EquivalentSegmentLists,
		Validate:        d.Validate,
		Create:          d.Create,
		Delete:          d.Delete,
		Update:          d.Update,
		Dependencies:    d.Dependencies,
	}
}

// IsPolicySegmentListKey distinguishes valid PolicySegmentList key
func (d *PolicySegmentListDescriptor) IsPolicySegmentListKey(key string) bool {
	_, _, valid := srv6.ParsePolicySegmentList(key)
	return valid
}

// Validate validates VPP SRv6 PolicySegmentList.
func (d *PolicySegmentListDescriptor) Validate(key string, sl *srv6.PolicySegmentList) error {
	_, err := ParseIPv6(sl.GetPolicyBsid())
	if err != nil {
		return scheduler.NewInvalidValueError(errors.Errorf("failed to parse policy bsid %s, should be a valid ipv6 address: %v", sl.GetPolicyBsid(), err), "policybsid")
	}
	for _, segment := range sl.Segments {
		_, err := ParseIPv6(segment)
		if err != nil {
			return scheduler.NewInvalidValueError(errors.Errorf("failed to parse segment %s in segments %v, should be a valid ipv6 address: %v", segment, sl.Segments, err), "segments")
		}
	}
	return nil
}

// Create creates new PolicySegmentList into VPP using VPP's binary api
func (d *PolicySegmentListDescriptor) Create(key string, sl *srv6.PolicySegmentList) (metadata interface{}, err error) {
	bsid, _ := ParseIPv6(sl.GetPolicyBsid()) // already validated
	policy, err := d.policy(sl.GetPolicyBsid())
	if err != nil {
		return nil, errors.Errorf("can't retrieve parent policy (bsid %v) for segment list %+v: %v", sl.GetPolicyBsid(), sl, err)
	}

	bsidStr := strings.TrimSpace(strings.ToLower(sl.GetPolicyBsid()))
	pInfo, exists := d.policyInfoCache[bsidStr]
	if !exists { // segment list creation called first -> we choose this segment list for use in policy creation
		d.policyInfoCache[bsidStr] = &cache.PolicyInfo{
			PolicyCreationSL:        sl,
			PolicyCreationSLCreated: true, // we are in SL creation and the choosen SL is this one
			SLCounter:               1,    // this is first SL in Policy
		}
		// Note: policySegmentListIndexCache will be filled with proper index when policy with this SL will be created
		return nil, nil
	}
	if !pInfo.PolicyCreationSLCreated && d.EquivalentSegmentLists("", sl, pInfo.PolicyCreationSL) { // this SL was choosen by policy descriptor for creation of policy
		// SL create as part of Policy -> do nothing  (pInfo.SLCounter was update in policy)
		pInfo.PolicyCreationSLCreated = true
		index, err := d.srHandler.RetrievePolicySegmentIndex(sl)
		if err != nil {
			return nil, errors.Errorf("can't retrieve from VPP the index of created segment list (first SL choosen by Policy descriptor): %v", err)
		}
		d.policySegmentListIndexCache.Put(sl, index)
		return nil, nil
	}

	// this SL is SL not used in initial policy creation -> adding it normally
	if err := d.srHandler.AddPolicySegmentList(bsid, policy, sl); err != nil {
		return nil, errors.Errorf("failed to add policy segment %s: %v", bsid, err)
	}
	index, err := d.srHandler.RetrievePolicySegmentIndex(sl)
	if err != nil {
		return nil, errors.Errorf("can't retrieve from VPP the index of created segment list (nonfirst SL): %v", err)
	}
	pInfo.SLCounter = pInfo.SLCounter + 1
	d.policySegmentListIndexCache.Put(sl, index)
	return nil, nil
}

// Delete removes PolicySegmentList from VPP using VPP's binary api
func (d *PolicySegmentListDescriptor) Delete(key string, sl *srv6.PolicySegmentList, metadata interface{}) error {
	index, indexExists := d.policySegmentListIndexCache.Get(sl)
	bsidStr := strings.TrimSpace(strings.ToLower(sl.GetPolicyBsid()))
	pInfo, infoExists := d.policyInfoCache[bsidStr]
	if !infoExists || pInfo.SLCounter <= 1 { // last segment for policy (nonexisting global info is case when last SL delete trigger policy delete before SL delete and that removes global info)
		// doing nothing because policy removal will be triggered and it will remove policy together with this segment list (don't have to update SLcounter because policy delete will remove global info anyway)
		if indexExists {
			d.policySegmentListIndexCache.Remove(sl)
		}
		return nil
	}

	// removing not-last segment
	if !indexExists {
		return errors.Errorf("can't get index for segment list %v", sl)
	}
	policy, err := d.policy(sl.GetPolicyBsid())
	if err != nil {
		return errors.Errorf("can't delete segment list because it can't be retrieve parent policy (bsid %v) for "+
			"segment list %+v: %v", sl.GetPolicyBsid(), sl, err)
	}
	bsid, _ := ParseIPv6(sl.GetPolicyBsid()) // already validated
	if err := d.srHandler.DeletePolicySegmentList(bsid, policy, sl, index); err != nil {
		return errors.Errorf("failed to delete policy segment %s: %v", bsid, err)
	}
	pInfo.SLCounter = pInfo.SLCounter - 1
	d.policySegmentListIndexCache.Remove(sl)

	return nil
}

// Update updates PolicySegmentList in VPP using VPP's binary api.
func (d *PolicySegmentListDescriptor) Update(key string, oldSL, newSL *srv6.PolicySegmentList, oldMetadata interface{}) (newMetadata interface{}, err error) {
	// NOTE: creating newSL first and then deleting the old one solves a lot of problems in compare to delete first and create new second. It solves problem
	// with last SL that can't be removed without recreating the whole Policy. This problem is solvable, but complicated to solve.
	// Also there is no problem to create new first, because update is triggered only when weight or segment list slice differs (if policy bsid differs then
	// it is other NB key and no update call is made) and this is ok for VPP (for VPP it looks like new segment list too, so not insert duplicate errors)
	newMetadata, err = d.Create(key, newSL)
	if err != nil {
		return nil, errors.Errorf("can't update policy segment list due to policy segment list creation problems: %v", err)
	}
	err = d.Delete(key, oldSL, oldMetadata)
	if err != nil {
		return nil, errors.Errorf("can't update policy segment list due to policy segment list delete problems: %v", err)
	}

	return newMetadata, nil
}

func (d *PolicySegmentListDescriptor) policy(bsid string) (*srv6.Policy, error) {
	kvs, err := d.scheduler.DumpValuesByDescriptor(PolicyDescriptorName, scheduler.NBView)
	if err != nil {
		return nil, errors.Errorf("can't get data from value provider(scheduler) for descriptor %v and view %v: %v", PolicyDescriptorName, scheduler.NBView, err)
	}
	bsid = strings.TrimSpace(strings.ToLower(bsid))
	for _, kv := range kvs {
		policy, ok := kv.Value.(*srv6.Policy)
		if !ok {
			return nil, errors.Errorf("Unexpected proto message type. Expected %v, Got %v", proto.MessageName(&srv6.Policy{}), proto.MessageName(kv.Value))
		}
		if strings.TrimSpace(strings.ToLower(policy.Bsid)) == bsid {
			return policy, nil
		}
	}
	return nil, errors.Errorf("can't find policy with bsid %v", bsid)
}

func allSegmentListsInOnePolicy(policyBSID string, valueProvider scheduler.ValueProvider) ([]*srv6.PolicySegmentList, []string, error) {
	kvs, err := valueProvider.DumpValuesByDescriptor(PolicySegmentListDescriptorName, scheduler.NBView)
	if err != nil {
		return nil, nil, errors.Errorf("can't get data from value provider(scheduler) for descriptor %v and view %v: %v", PolicySegmentListDescriptorName, scheduler.NBView, err)
	}
	policyBSID = strings.TrimSpace(strings.ToLower(policyBSID))
	slList := make([]*srv6.PolicySegmentList, 0)
	slKeys := make([]string, 0)
	for _, kv := range kvs {
		sl, ok := kv.Value.(*srv6.PolicySegmentList)
		if !ok {
			return nil, nil, errors.Errorf("Unexpected proto message type. Expected %v, Got %v", proto.MessageName(&srv6.PolicySegmentList{}), proto.MessageName(kv.Value))
		}
		if strings.TrimSpace(strings.ToLower(sl.PolicyBsid)) == policyBSID {
			slList = append(slList, sl)
		}
		slKeys = append(slKeys, kv.Key)
	}
	return slList, slKeys, nil
}

// Dependencies defines dependencies of PolicySegmentList descriptor
func (d *PolicySegmentListDescriptor) Dependencies(key string, segmentList *srv6.PolicySegmentList) (dependencies []scheduler.Dependency) {
	dependencies = append(dependencies, scheduler.Dependency{
		Label: parentPolicyExistsDep,
		Key:   srv6.PolicyKey(segmentList.PolicyBsid),
	})
	return dependencies
}

// EquivalentSegmentLists determines whether 2 policy segment lists are logically equal. This comparison takes into
// consideration also semantics that couldn't be modeled into proto models (i.e. SID is IPv6 address and not only string)
func (d *PolicySegmentListDescriptor) EquivalentSegmentLists(key string, oldSL, newSL *srv6.PolicySegmentList) bool {
	if oldSL == nil || newSL == nil {
		return oldSL == newSL
	}
	return oldSL.Weight == newSL.Weight &&
		equivalentSIDs(oldSL.PolicyBsid, newSL.PolicyBsid) &&
		d.equivalentSegments(oldSL.Segments, newSL.Segments)
}

func (d *PolicySegmentListDescriptor) equivalentSegments(segments1, segments2 []string) bool {
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
