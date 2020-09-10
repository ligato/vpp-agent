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

package vpp1908

import (
	govppapi "git.fd.io/govpp.git/api"

	vpe_vppcalls "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	vpe_vpp1908 "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp1908"
	"go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/memclnt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vpe"
)

func init() {
	msgs := vpp.Messages(
		vpe.AllMessages,
		memclnt.AllMessages,
	)
	vppcalls.AddHandlerVersion(vpp1908.Version, msgs.AllMessages(), func(c vpp.Client) vppcalls.TelemetryVppAPI {
		ch, _ := c.NewAPIChannel()
		return NewTelemetryVppHandler(ch)
	})
}

type TelemetryHandler struct {
	vpe vpe_vppcalls.VppCoreAPI
}

func NewTelemetryVppHandler(ch govppapi.Channel) vppcalls.TelemetryVppAPI {
	return &TelemetryHandler{
		vpe: vpe_vpp1908.NewVpeHandler(ch),
	}
}
