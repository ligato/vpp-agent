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
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/stn"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/stnplugin/vppcalls"
)

func init() {
	msgs := stn.AllMessages()
	vppcalls.AddStnHandlerVersion(vpp1904.Version, msgs, NewStnVppHandler)
}

// StnVppHandler is accessor for STN-related vppcalls methods
type StnVppHandler struct {
	callsChannel govppapi.Channel
	// TODO: use RPC service
	//stn          stn.RPCService
	ifIndexes ifaceidx.IfaceMetadataIndex
	log       logging.Logger
}

// NewStnVppHandler creates new instance of STN vppcalls handler
func NewStnVppHandler(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) vppcalls.StnVppAPI {
	return &StnVppHandler{
		callsChannel: ch,
		//stn:          stn.NewServiceClient(ch),
		ifIndexes: ifIdx,
		log:       log,
	}
}
