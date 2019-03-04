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

package vppcalls

import (
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
)

// SRv6VppAPI is API boundary for vppcall package access, introduced to properly test code dependent on vppcalls package
type SRv6VppAPI interface {
	SRv6VPPWrite
	SRv6VPPRead
}

// SRv6VPPWrite provides write methods for segment routing
type SRv6VPPWrite interface {
	// AddLocalSid adds local sid <localSID> into VPP
	AddLocalSid(localSID *srv6.LocalSID) error
	// DeleteLocalSid delets local sid given by <sidAddr> in VPP
	DeleteLocalSid(sidAddr net.IP) error
	// SetEncapsSourceAddress sets for SRv6 in VPP the source address used for encapsulated packet
	SetEncapsSourceAddress(address string) error
	// AddPolicy adds SRv6 policy <policy> into VPP (including all policy's segment lists).
	AddPolicy(policy *srv6.Policy) error
	// DeletePolicy deletes SRv6 policy given by binding SID <bindingSid> (including all policy's segment lists).
	DeletePolicy(bindingSid net.IP) error
	// AddPolicySegmentList adds segment list <segmentList> to SRv6 policy <policy> in VPP
	AddPolicySegmentList(segmentList *srv6.Policy_SegmentList, policy *srv6.Policy) error
	// DeletePolicySegmentList removes segment list <segmentList> (with VPP-internal index <segmentVPPIndex>) from SRv6 policy <policy> in VPP
	DeletePolicySegmentList(segmentList *srv6.Policy_SegmentList, segmentVPPIndex uint32, policy *srv6.Policy) error
	// AddSteering sets in VPP steering into SRv6 policy.
	AddSteering(steering *srv6.Steering) error
	// RemoveSteering removes in VPP steering into SRv6 policy.
	RemoveSteering(steering *srv6.Steering) error
}

// SRv6VPPRead provides read methods for segment routing
type SRv6VPPRead interface {
	// TODO: implement other dump methods

	// RetrievePolicyIndexInfo retrieves index of policy <policy> and its segment lists
	RetrievePolicyIndexInfo(policy *srv6.Policy) (policyIndex uint32, segmentListIndexes map[*srv6.Policy_SegmentList]uint32, err error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, ifaceidx.IfaceMetadataIndex, logging.Logger) SRv6VppAPI
}

func CompatibleSRv6VppHandler(
	ch govppapi.Channel, idx ifaceidx.IfaceMetadataIndex, log logging.Logger,
) SRv6VppAPI {
	for ver, h := range Versions {
		log.Debugf("checking compatibility with %s", ver)
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch, idx, log)
	}
	panic("no compatible version available")
}
