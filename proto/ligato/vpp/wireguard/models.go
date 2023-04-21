//  Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package vpp_wg

import (
	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for models.
const ModuleName = "vpp.wg"

var ModelPeer models.KnownModel

func init() {
	// models.Register requires protoreflect capabilities, so we initialize them first
	file_ligato_vpp_wireguard_wireguard_proto_init()

	ModelPeer = models.Register(&Peer{}, models.Spec{
		Module:  ModuleName,
		Version: "v1",
		Type:    "peer",
	}, models.WithNameTemplate("{{.WgIfName}}/endpoint/{{.Endpoint}}/{{.Port}}"))
}