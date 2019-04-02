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

package vpp1810

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	ba_ip "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/ip"
	ba_punt "github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, ba_ip.Messages...)
	msgs = append(msgs, ba_punt.Messages...)

	vppcalls.Versions["vpp1810"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New:  NewPuntVppHandler,
	}
}

// PuntVppHandler is accessor for punt-related vppcalls methods.
type PuntVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// NewPuntVppHandler creates new instance of punt vppcalls handler
func NewPuntVppHandler(callsChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) vppcalls.PuntVppAPI {
	return &PuntVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}
