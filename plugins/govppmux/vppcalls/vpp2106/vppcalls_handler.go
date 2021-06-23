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

package vpp2106

import (
	"go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/memclnt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vpe"
)

func init() {
	msgs := vpp.Messages(
		vpe.AllMessages,
		memclnt.AllMessages,
	)
	vppcalls.AddVersion(vpp2106.Version, msgs.AllMessages(), NewVpeHandler)
}

type VpeHandler struct {
	memclnt memclnt.RPCService
	vpe     vpe.RPCService
}

func NewVpeHandler(c vpp.Client) vppcalls.VppCoreAPI {
	return &VpeHandler{
		memclnt: memclnt.NewServiceClient(c),
		vpe:     vpe.NewServiceClient(c),
	}
}
