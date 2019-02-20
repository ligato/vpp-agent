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

package vpp_srv6

import (
	"fmt"
	"net"
	"strings"

	"github.com/ligato/vpp-agent/pkg/models"
)

// ModuleName is the module name used for models.
const (
	ModuleName = "vpp.srv6"
)

var (
	// ModelLocalSID is registered NB model of LocalSID
	ModelLocalSID = models.Register(&LocalSID{}, models.Spec{
		Module:  ModuleName,
		Type:    "localsid",
		Version: "v2",
	}, models.WithNameTemplate("{{.Sid}}"))

	// ModelPolicy is registered NB model of Policy
	ModelPolicy = models.Register(&Policy{}, models.Spec{
		Module:  ModuleName,
		Type:    "policy",
		Version: "v2",
	}, models.WithNameTemplate("{{.Bsid}}"))

	// ModelPolicySegmentList is registered NB model of PolicySegmentList
	ModelPolicySegmentList = models.Register(&PolicySegmentList{}, models.Spec{
		Module:  ModuleName,
		Type:    "policysegmentlist",
		Version: "v2",
	}, models.WithNameTemplate(fmt.Sprintf("{{.Name}}/%s/{{.PolicyBsid}}", ModelPolicy.Type)))

	// ModelSteering is registered NB model of Steering
	ModelSteering = models.Register(&Steering{}, models.Spec{
		Module:  ModuleName,
		Type:    "steering",
		Version: "v2",
	}, models.WithNameTemplate("{{.Name}}"))
)

// SID (in srv6 package) is SRv6's segment id. It is always represented as IPv6 address
type SID = net.IP

// PolicyKey returns the key used in ETCD to store vpp sr policy for vpp instance.
func PolicyKey(bsid string) string {
	return models.Key(&Policy{
		Bsid: bsid,
	})
}

// ParsePolicySegmentList parses key representing policy segment list
func ParsePolicySegmentList(key string) (policyBSID string, policySegmentListName string, isPolicySegmentListKey bool) {
	keyComps := strings.Split(key, "/")
	if len(keyComps) == 8 && isCorrectModule(keyComps) && keyComps[4] == ModelPolicySegmentList.Type && keyComps[6] == ModelPolicy.Type {
		return keyComps[7], keyComps[5], true
	}
	return "", "", false
}

func isCorrectModule(keyComps []string) bool {
	expectedKeys := strings.Split(ModuleName, ".")
	return keyComps[1] == expectedKeys[0] && keyComps[2] == expectedKeys[1]
}
