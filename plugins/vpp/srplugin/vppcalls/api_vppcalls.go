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

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"
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
	// DeleteLocalSid deletes local sid <localSID> in VPP
	DeleteLocalSid(localSID *srv6.LocalSID) error
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

	// DumpLocalSids retrieves all localsids
	DumpLocalSids() (localsids []*srv6.LocalSID, err error)

	// RetrievePolicyIndexInfo retrieves index of policy <policy> and its segment lists
	RetrievePolicyIndexInfo(policy *srv6.Policy) (policyIndex uint32, segmentListIndexes map[*srv6.Policy_SegmentList]uint32, err error)
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "srv6",
	HandlerAPI: (*SRv6VppAPI)(nil),
})

type NewHandlerFunc func(vpp.Client, ifaceidx.IfaceMetadataIndex, logging.Logger) SRv6VppAPI

func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			return c.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			return h(c, a[0].(ifaceidx.IfaceMetadataIndex), a[1].(logging.Logger))
		},
	})
}

func CompatibleSRv6Handler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) SRv6VppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, log).(SRv6VppAPI)
	}
	return nil
}
