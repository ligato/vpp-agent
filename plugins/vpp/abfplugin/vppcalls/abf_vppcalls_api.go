//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package vppcalls

import (
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/aclidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
)

// ABFDetails contains proto-modeled ABF data together with VPP-related metadata
type ABFDetails struct {
	ABF  *abf.ABF `json:"abf"`
	Meta *ABFMeta `json:"abf_meta"`
}

// ABFMeta contains policy ID (ABF index)
type ABFMeta struct {
	PolicyID uint32 `json:"policy_id"`
}

// ABFVppAPI provides read/write methods required to handle VPP ACL-based forwarding
type ABFVppAPI interface {
	ABFVppRead

	// AddAbfPolicy creates new ABF entry together with a list of forwarding paths
	AddAbfPolicy(policyID, aclID uint32, abfPaths []*abf.ABF_ForwardingPath) error
	// DeleteAbfPolicy removes existing ABF entry
	DeleteAbfPolicy(policyID uint32, abfPaths []*abf.ABF_ForwardingPath) error
	// AbfAttachInterfaceIPv4 attaches IPv4 interface to the ABF
	AbfAttachInterfaceIPv4(policyID, ifIdx, priority uint32) error
	// AbfDetachInterfaceIPv4 detaches IPV4 interface from the ABF
	AbfDetachInterfaceIPv4(policyID, ifIdx, priority uint32) error
	// AbfAttachInterfaceIPv6 attaches IPv6 interface to the ABF
	AbfAttachInterfaceIPv6(policyID, ifIdx, priority uint32) error
	// AbfDetachInterfaceIPv6 detaches IPv6 interface from the ABF
	AbfDetachInterfaceIPv6(policyID, ifIdx, priority uint32) error
}

// ABFVppRead provides read methods for ABF plugin
type ABFVppRead interface {
	// GetAbfVersion retrieves version of the VPP ABF plugin
	GetAbfVersion() (ver string, err error)
	// DumpABFPolicy retrieves VPP ABF configuration.
	DumpABFPolicy() ([]*ABFDetails, error)
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "abf",
	HandlerAPI: (*ABFVppAPI)(nil),
})

type NewHandlerFunc func(ch govppapi.Channel, aclIdx aclidx.ACLMetadataIndex, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) ABFVppAPI

func AddABFHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return ch.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			var aclIdx aclidx.ACLMetadataIndex
			if a[0] != nil {
				aclIdx = a[0].(aclidx.ACLMetadataIndex)
			}
			return h(ch, aclIdx, a[1].(ifaceidx.IfaceMetadataIndex), a[2].(logging.Logger))
		},
	})
}

func CompatibleABFHandler(c vpp.Client, aclIdx aclidx.ACLMetadataIndex, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) ABFVppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, aclIdx, ifIdx, log).(ABFVppAPI)
	}
	return nil
}
