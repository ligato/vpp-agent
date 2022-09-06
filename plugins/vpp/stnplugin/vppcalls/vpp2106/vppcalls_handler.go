//  Copyright (c) 2021 Cisco and/or its affiliates.
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
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	vpp2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106"
	vpp_stn "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/stn"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpp_stn.AllMessages()...)

	vppcalls.AddStnHandlerVersion(vpp2106.Version, msgs, NewStnVppHandler)
}

// StnVppHandler is accessor for STN-related vppcalls methods
type StnVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// NewStnVppHandler creates new instance of STN vppcalls handler
func NewStnVppHandler(
	callsChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger,
) vppcalls.StnVppAPI {
	return &StnVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}
