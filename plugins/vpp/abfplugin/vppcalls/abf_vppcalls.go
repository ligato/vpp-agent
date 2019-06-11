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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	abf "github.com/ligato/vpp-agent/api/models/vpp/abf"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin/aclidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
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

	// GetAbfVersion retrieves version of the VPP ABF plugin
	GetAbfVersion() (ver string, err error)
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
	// DumpABFPolicy retrieves VPP ABF configuration.
	DumpABFPolicy() ([]*ABFDetails, error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, aclidx.ACLMetadataIndex, ifaceidx.IfaceMetadataIndex) ABFVppAPI
}

func CompatibleABFVppHandler(ch govppapi.Channel, aclIdx aclidx.ACLMetadataIndex, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) ABFVppAPI {
	if len(Versions) == 0 {
		// abfplugin is not loaded
		return nil
	}
	for ver, h := range Versions {
		log.Debugf("checking compatibility with %s", ver)
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch, aclIdx, ifIdx)
	}
	panic("no compatible version available")
}
