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

package vpp1904

import (
	govppapi "git.fd.io/govpp.git/api"
	"go.ligato.io/vpp-agent/v2/plugins/vpp"

	"go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1904"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1904/memclnt"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1904/vpe"
)

/*var (
	CompatibilityCheck = vpp.MessageCheck(
		vpe.AllMessages,
		memclnt.AllMessages,
	)
)

var HandlerVersion = vpp.HandlerVersion{
	Version: vpp1904.Version,
	Check: vpp.MessageCheck(
		vpe.AllMessages,
		memclnt.AllMessages,
	),
	NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
		return NewVpeHandler(c)
	},
}*/

func init() {
	msgs := vpp.Messages(
		vpe.AllMessages,
		memclnt.AllMessages,
	)
	vppcalls.AddVersion(vpp1904.Version, msgs.AllMessages(), NewVpeHandler)
}

type VpeHandler struct {
	memclnt memclnt.RPCService
	vpe     vpe.RPCService
}

func NewVpeHandler(ch govppapi.Channel) vppcalls.VppCoreAPI {
	return &VpeHandler{
		memclnt: memclnt.NewServiceClient(ch),
		vpe:     vpe.NewServiceClient(ch),
	}
}
