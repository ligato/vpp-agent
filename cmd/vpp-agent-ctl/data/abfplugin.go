// Copyright (c) 2019 Cisco and/or its affiliates.
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

package data

import (
	abf "github.com/ligato/vpp-agent/api/models/vpp/abf"
)

// ABFCtl provides access list related methods for vpp-agent-ctl
type ABFCtl interface {
	// PutABF puts ACL-based forwarding to the ETCD
	PutABF() error
	// DeleteABF removes ACL-based forwarding from the ETCD
	DeleteABF() error
}

// PutABF puts ACL-based forwarding to the ETCD
func (ctl *VppAgentCtlImpl) PutABF() error {
	abfData := &abf.ABF{
		Index:   "1",
		AclName: "aclip1",
		AttachedInterfaces: []*abf.ABF_AttachedInterface{
			{
				InputInterface: "tap1",
				IsIpv6:         false,
				Priority:       40,
			},
			{
				InputInterface: "memif1",
				IsIpv6:         false,
				Priority:       60,
			},
		},
		ForwardingPaths: []*abf.ABF_ForwardingPath{
			{
				NextHopIp:     "10.0.0.10",
				InterfaceName: "loop1",
				Weight:        20,
				Preference:    25,
				Dvr:           false,
			},
		},
	}

	ctl.Log.Infof("ABF put: %v", abfData)
	return ctl.broker.Put(abf.Key(abfData.Index), abfData)
}

// DeleteABF removes ACL-based forwarding from the ETCD
func (ctl *VppAgentCtlImpl) DeleteABF() error {
	abfKey := abf.Key("1")

	ctl.Log.Infof("Deleted ABF: %v", abfKey)
	_, err := ctl.broker.Delete(abfKey)
	return err
}
