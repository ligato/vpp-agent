// Copyright (c) 2018 Bell Canada, Pantheon Technologies and/or its affiliates.
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

//go:generate protoc --proto_path=../model/srv6 --gogo_out=../model/srv6 ../model/srv6/srv6.proto

package srplugin

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/srv6"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/cache"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/vppcalls"
)

// TODO check all SID usages for comparisons that can fail due to upper/lower case mismatch (i.e. strings "A::E" and "a::e" are not equal but for our purposes it is the same SID and should be considered equal)

// SRv6Configurator runs in the background where it watches for any changes in the configuration of interfaces as
// modelled by the proto file "../model/srv6/srv6.proto" and stored in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/srv6".
type SRv6Configurator struct {
	// injectable/public fields
	Log         logging.Logger
	GoVppmux    govppmux.API
	SwIfIndexes ifaceidx.SwIfIndex // SwIfIndexes from default plugins
	VppCalls    vppcalls.SRv6Calls

	// channels
	vppChannel vppcalls.VPPChannel // channel to communicate with VPP

	// caches
	policyCache         *cache.PolicyCache        // Cache for SRv6 policies
	policySegmentsCache *cache.PolicySegmentCache // Cache for SRv6 policy segments
	steeringCache       *cache.SteeringCache      // Cache for SRv6 steering
	createdPolicies     map[string]struct{}       // Marker for created policies (key = bsid in string form)

	// indexes
	policyIndexSeq        *gaplessSequence
	policyIndexes         idxvpp.NameToIdxRW // Mapping between policy bsid and index inside VPP
	policySegmentIndexSeq *gaplessSequence
	policySegmentIndexes  idxvpp.NameToIdxRW // Mapping between policy segment name as defined in ETCD key and index inside VPP
}

// Init members
func (plugin *SRv6Configurator) Init() (err error) {
	// NewAPIChannel returns a new API channel for communication with VPP via govpp core.
	// It uses default buffer sizes for the request and reply Go channels.
	plugin.vppChannel, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return
	}

	// Create caches
	plugin.policyCache = cache.NewPolicyCache(plugin.Log)
	plugin.policySegmentsCache = cache.NewPolicySegmentCache(plugin.Log)
	plugin.steeringCache = cache.NewSteeringCache(plugin.Log)
	plugin.createdPolicies = make(map[string]struct{})

	// Create indexes
	plugin.policySegmentIndexSeq = newSequence()
	plugin.policySegmentIndexes = nametoidx.NewNameToIdx(plugin.Log, "policy-segment-indexes", nil)
	plugin.policyIndexSeq = newSequence()
	plugin.policyIndexes = nametoidx.NewNameToIdx(plugin.Log, "policy-indexes", nil)

	return
}

// Close closes GOVPP channel
func (plugin *SRv6Configurator) Close() error {
	_, err := safeclose.CloseAll(plugin.vppChannel)
	return err
}

// AddLocalSID adds new Local SID into VPP using VPP's binary api
func (plugin *SRv6Configurator) AddLocalSID(value *srv6.LocalSID) error {
	sid, err := ParseIPv6(value.GetSid())
	if err != nil {
		return fmt.Errorf("sid should be valid ipv6 address: %v", err)
	}
	return plugin.VppCalls.AddLocalSid(sid, value, plugin.SwIfIndexes, plugin.vppChannel)
}

// DeleteLocalSID removes Local SID from VPP using VPP's binary api
func (plugin *SRv6Configurator) DeleteLocalSID(value *srv6.LocalSID) error {
	sid, err := ParseIPv6(value.GetSid())
	if err != nil {
		return fmt.Errorf("sid should be valid ipv6 address: %v", err)
	}
	return plugin.VppCalls.DeleteLocalSid(sid, plugin.vppChannel)
}

// ModifyLocalSID modifies Local SID from <prevValue> to <value> in VPP using VPP's binary api
func (plugin *SRv6Configurator) ModifyLocalSID(value *srv6.LocalSID, prevValue *srv6.LocalSID) error {
	err := plugin.DeleteLocalSID(prevValue)
	if err != nil {
		return fmt.Errorf("can't delete old version of Local SID: %v", err)
	}
	err = plugin.AddLocalSID(value)
	if err != nil {
		return fmt.Errorf("can't apply new version of Local SID: %v", err)
	}
	return nil
}

// AddPolicy adds new policy into VPP using VPP's binary api
func (plugin *SRv6Configurator) AddPolicy(policy *srv6.Policy) error {
	bsid, err := ParseIPv6(policy.GetBsid())
	if err != nil {
		return fmt.Errorf("bsid should be valid ipv6 address: %v", err)
	}
	plugin.policyCache.Put(bsid, policy)
	segments, segmentNames := plugin.policySegmentsCache.LookupByPolicy(bsid)
	if len(segments) == 0 {
		plugin.Log.Debugf("addition of policy (%v) postponed until first policy segment is defined for it", bsid.String())
		return nil
	}

	plugin.addPolicyToIndexes(bsid)
	plugin.addSegmentToIndexes(bsid, segmentNames[0])
	err = plugin.VppCalls.AddPolicy(bsid, policy, segments[0], plugin.vppChannel)
	if err != nil {
		return fmt.Errorf("can't write policy (%v) with first segment (%v): %v", bsid, segments[0].Segments, err)
	}
	plugin.createdPolicies[bsid.String()] = struct{}{} // write into Set that policy was successfully created
	if len(segments) > 1 {
		for i, segment := range segments[1:] {
			err = plugin.AddPolicySegment(segmentNames[i], segment)
			if err != nil {
				return fmt.Errorf("can't apply subsequent policy segment (%v) to policy (%v): %v", segment.Segments, bsid.String(), err)
			}
		}
	}

	// adding policy dependent steerings
	idx, _, _ := plugin.policyIndexes.LookupIdx(bsid.String())
	steerings, steeringNames := plugin.lookupSteeringByPolicy(bsid, idx)
	for i, steering := range steerings {
		if err := plugin.AddSteering(steeringNames[i], steering); err != nil {
			return fmt.Errorf("can't create steering by creating policy referenced by steering: %v", err)
		}
	}
	return nil
}

func (plugin *SRv6Configurator) lookupSteeringByPolicy(bsid srv6.SID, index uint32) ([]*srv6.Steering, []string) {
	// union search by bsid and policy index
	steerings1, steeringNames1 := plugin.steeringCache.LookupByPolicyBSID(bsid)
	steerings2, steeringNames2 := plugin.steeringCache.LookupByPolicyIndex(index)
	steerings1 = append(steerings1, steerings2...)
	steeringNames1 = append(steeringNames1, steeringNames2...)
	return steerings1, steeringNames1
}

// RemovePolicy removes policy from VPP using VPP's binary api
func (plugin *SRv6Configurator) RemovePolicy(policy *srv6.Policy) error {
	bsid, err := ParseIPv6(policy.GetBsid())
	if err != nil {
		return fmt.Errorf("bsid should be valid ipv6 address: %v", err)
	}
	// adding policy dependent steerings
	idx, _, _ := plugin.policyIndexes.LookupIdx(bsid.String())
	steerings, steeringNames := plugin.lookupSteeringByPolicy(bsid, idx)
	for i, steering := range steerings {
		if err := plugin.RemoveSteering(steeringNames[i], steering); err != nil {
			return fmt.Errorf("can't remove steering in process of removing policy referenced by steering: %v", err)
		}
	}

	plugin.policyCache.Delete(bsid)
	plugin.policyIndexes.UnregisterName(bsid.String())
	delete(plugin.createdPolicies, bsid.String())
	_, segmentNames := plugin.policySegmentsCache.LookupByPolicy(bsid)
	for _, segmentName := range segmentNames {
		plugin.policySegmentsCache.Delete(bsid, segmentName)
		index, _, exists := plugin.policySegmentIndexes.UnregisterName(plugin.uniquePolicySegmentName(bsid, segmentName))
		if exists {
			plugin.policySegmentIndexSeq.delete(index)
		}
	}
	return plugin.VppCalls.DeletePolicy(bsid, plugin.vppChannel) // expecting that policy delete will also delete policy segments in vpp
}

// ModifyPolicy modifies policy in VPP using VPP's binary api
func (plugin *SRv6Configurator) ModifyPolicy(value *srv6.Policy, prevValue *srv6.Policy) error {
	bsid, err := ParseIPv6(value.GetBsid())
	if err != nil {
		return fmt.Errorf("bsid should be valid ipv6 address: %v", err)
	}
	segments, segmentNames := plugin.policySegmentsCache.LookupByPolicy(bsid)
	err = plugin.RemovePolicy(prevValue)
	if err != nil {
		return fmt.Errorf("can't delete old version of Policy: %v", err)
	}
	err = plugin.AddPolicy(value)
	if err != nil {
		return fmt.Errorf("can't apply new version of Policy: %v", err)
	}
	for i, segment := range segments {
		err = plugin.AddPolicySegment(segmentNames[i], segment)
		if err != nil {
			return fmt.Errorf("can't apply segment %v (%v) as part of policy modification: %v", segmentNames[i], segment.Segments, err)
		}
	}
	return nil
}

// AddPolicySegment adds policy segment <policySegment> with name <segmentName> into referenced policy in VPP using VPP's binary api.
func (plugin *SRv6Configurator) AddPolicySegment(segmentName string, policySegment *srv6.PolicySegment) error {
	bsid, err := ParseIPv6(policySegment.GetPolicyBsid())
	if err != nil {
		return fmt.Errorf("policy bsid should be valid ipv6 address: %v", err)
	}
	plugin.policySegmentsCache.Put(bsid, segmentName, policySegment)
	policy, exists := plugin.policyCache.GetValue(bsid)
	if !exists {
		plugin.Log.Debugf("addition of policy segment (%v) postponed until policy with %v bsid is created", policySegment.GetSegments(), bsid.String())
		return nil
	}

	segments, _ := plugin.policySegmentsCache.LookupByPolicy(bsid)
	if len(segments) <= 1 {
		if _, alreadyCreated := plugin.createdPolicies[bsid.String()]; alreadyCreated {
			// last segment got deleted in etcd, but policy with last segment stays in VPP, and we want add another segment
			// -> we must remove old policy with last segment from VPP to add it again with new segment
			err := plugin.RemovePolicy(policy)
			if err != nil {
				return fmt.Errorf("can't delete Policy (with previously deleted last policy segment) to recreated it with new policy segment: %v", err)
			}
			plugin.policySegmentsCache.Put(bsid, segmentName, policySegment) // got deleted in policy removal
		}
		return plugin.AddPolicy(policy)
	}
	// FIXME there is no API contract saying what happens to VPP indexes if addition fails (also different fail code can rollback or not rollback indexes) => no way how to handle this without being dependent on internal implementation inside VPP and that is just very fragile -> API should tell this but it doesn't!
	plugin.addSegmentToIndexes(bsid, segmentName)
	return plugin.VppCalls.AddPolicySegment(bsid, policy, policySegment, plugin.vppChannel)
}

// RemovePolicySegment removes policy segment <policySegment> with name <segmentName> from referenced policy in VPP using
// VPP's binary api. In case of last policy segment in policy, policy segment is not removed, because policy can't exists
// in VPP without policy segment. Instead it is postponed until policy removal or addition of another policy segment happen.
func (plugin *SRv6Configurator) RemovePolicySegment(segmentName string, policySegment *srv6.PolicySegment) error {
	bsid, err := ParseIPv6(policySegment.GetPolicyBsid())
	if err != nil {
		return fmt.Errorf("policy bsid should be valid ipv6 address: %v", err)
	}
	plugin.policySegmentsCache.Delete(bsid, segmentName)
	index, _, exists := plugin.policySegmentIndexes.UnregisterName(plugin.uniquePolicySegmentName(bsid, segmentName))

	siblings, _ := plugin.policySegmentsCache.LookupByPolicy(bsid) // sibling segments in the same policy
	if len(siblings) == 0 {                                        // last segment for policy
		plugin.Log.Debugf("removal of policy segment (%v) postponed until policy with %v bsid is deleted", policySegment.GetSegments(), bsid.String())
		return nil
	}

	// removing not-last segment
	if !exists {
		return fmt.Errorf("can't find index of policy segment %v in policy with bsid %v", policySegment.Segments, bsid)
	}
	policy, exists := plugin.policyCache.GetValue(bsid)
	if !exists {
		return fmt.Errorf("can't find policy with bsid %v", bsid)
	}
	// FIXME there is no API contract saying what happens to VPP indexes if removal fails (also different fail code can rollback or not rollback indexes) => no way how to handle this without being dependent on internal implementation inside VPP and that is just very fragile -> API should tell this but it doesn't!
	plugin.policySegmentIndexSeq.delete(index)
	return plugin.VppCalls.DeletePolicySegment(bsid, policy, policySegment, index, plugin.vppChannel)
}

// ModifyPolicySegment modifies existing policy segment with name <segmentName> from <prevValue> to <value> in referenced policy.
func (plugin *SRv6Configurator) ModifyPolicySegment(segmentName string, value *srv6.PolicySegment, prevValue *srv6.PolicySegment) error {
	bsid, err := ParseIPv6(value.GetPolicyBsid())
	if err != nil {
		return fmt.Errorf("policy bsid should be valid ipv6 address: %v", err)
	}
	segments, _ := plugin.policySegmentsCache.LookupByPolicy(bsid)
	if len(segments) <= 1 { // last segment in policy can't be removed without removing policy itself
		policy, exists := plugin.policyCache.GetValue(bsid)
		if !exists {
			return fmt.Errorf("can't find Policy in cache when updating last policy segment in policy")
		}
		err := plugin.RemovePolicy(policy)
		if err != nil {
			return fmt.Errorf("can't delete Policy as part of removing old version of last policy segment in policy: %v", err)
		}
		err = plugin.AddPolicy(policy)
		if err != nil {
			return fmt.Errorf("can't add Policy as part of adding new version of last policy segment in policy: %v", err)
		}
		err = plugin.AddPolicySegment(segmentName, value)
		if err != nil {
			return fmt.Errorf("can't apply new version of last Policy segment: %v", err)
		}
		return nil
	}
	err = plugin.RemovePolicySegment(segmentName, prevValue)
	if err != nil {
		return fmt.Errorf("can't delete old version of Policy segment: %v", err)
	}
	err = plugin.AddPolicySegment(segmentName, value)
	if err != nil {
		return fmt.Errorf("can't apply new version of Policy segment: %v", err)
	}
	return nil
}

func (plugin *SRv6Configurator) addSegmentToIndexes(bsid srv6.SID, segmentName string) {
	plugin.policySegmentIndexes.RegisterName(plugin.uniquePolicySegmentName(bsid, segmentName), plugin.policySegmentIndexSeq.nextID(), nil)
}

func (plugin *SRv6Configurator) addPolicyToIndexes(bsid srv6.SID) {
	plugin.policyIndexes.RegisterName(bsid.String(), plugin.policyIndexSeq.nextID(), nil)
}

func (plugin *SRv6Configurator) uniquePolicySegmentName(bsid srv6.SID, segmentName string) string {
	return bsid.String() + srv6.EtcdKeyPathDelimiter + segmentName
}

// AddSteering adds new steering into VPP using VPP's binary api
func (plugin *SRv6Configurator) AddSteering(name string, steering *srv6.Steering) error {
	plugin.steeringCache.Put(name, steering)
	bsidStr := steering.PolicyBsid
	if len(strings.Trim(steering.PolicyBsid, " ")) == 0 { // policy defined by index
		var exists bool
		bsidStr, _, exists = plugin.policyIndexes.LookupName(steering.PolicyIndex)
		if !exists {
			plugin.Log.Debugf("addition of steering (index %v) postponed until referenced policy is defined", steering.PolicyIndex)
			return nil
		}
	}
	// policy defined by BSID (or index defined converted to BSID defined)
	bsid, err := ParseIPv6(bsidStr)
	if err != nil {
		return fmt.Errorf("can't parse policy BSID string ('%v') into IPv6 address", steering.PolicyBsid)
	}
	if _, exists := plugin.policyCache.GetValue(bsid); !exists {
		plugin.Log.Debugf("addition of steering (bsid %v) postponed until referenced policy is defined", name)
		return nil
	}

	return plugin.VppCalls.AddSteering(steering, plugin.SwIfIndexes, plugin.vppChannel)
}

// RemoveSteering removes steering from VPP using VPP's binary api
func (plugin *SRv6Configurator) RemoveSteering(name string, steering *srv6.Steering) error {
	plugin.steeringCache.Delete(name)
	return plugin.VppCalls.RemoveSteering(steering, plugin.SwIfIndexes, plugin.vppChannel)
}

// ModifySteering modifies existing steering in VPP using VPP's binary api
func (plugin *SRv6Configurator) ModifySteering(name string, value *srv6.Steering, prevValue *srv6.Steering) error {
	err := plugin.RemoveSteering(name, prevValue)
	if err != nil {
		return fmt.Errorf("can't delete old version of steering %v: %v", name, err)
	}
	err = plugin.AddSteering(name, value)
	if err != nil {
		return fmt.Errorf("can't apply new version of steering %v: %v", name, err)
	}
	return nil
}

// ParseIPv6 parses string <str> to IPv6 address (including IPv4 address converted to IPv6 address)
func ParseIPv6(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, fmt.Errorf(" %q is not ip address", str)
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, fmt.Errorf(" %q is not ipv6 address", str)
	}
	return ipv6, nil
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
