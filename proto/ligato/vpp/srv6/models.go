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
	"net"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for models.
const ModuleName = "vpp.srv6"

var (
	ModelLocalSID   models.KnownModel
	ModelPolicy     models.KnownModel
	ModelSteering   models.KnownModel
	ModelSRv6Global models.KnownModel
)

func init() {
	// models.Register requires protoreflect capabilities, so we initialize them first
	file_ligato_vpp_srv6_srv6_proto_init()

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

	// ModelSteering is registered NB model of Steering
	ModelSteering = models.Register(&Steering{}, models.Spec{
		Module:  ModuleName,
		Type:    "steering",
		Version: "v2",
	}, models.WithNameTemplate("{{.Name}}"))

	// ModelSRv6Global is registered NB model of SRv6 global config
	ModelSRv6Global = models.Register(&SRv6Global{}, models.Spec{
		Module:  ModuleName,
		Type:    "srv6-global",
		Version: "v2",
	})
}

// SID (in srv6 package) is SRv6's segment id. It is always represented as IPv6 address
type SID = net.IP

// PolicyKey returns the key used in ETCD to store vpp sr policy for vpp instance.
func PolicyKey(bsid string) string {
	return models.Key(&Policy{
		Bsid: bsid,
	})
}
