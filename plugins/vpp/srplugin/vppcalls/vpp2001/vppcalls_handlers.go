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

package vpp2001

import (
	govppapi "git.fd.io/govpp.git/api"
	"go.ligato.io/cn-infra/v2/logging"

	core_vppcalls "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	core_vpp2001 "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp2001"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/sr"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls"
)

func init() {
	msgs := vpp.Messages(
		sr.AllMessages,
		vpe.AllMessages, // using also vpe -> need to have correct vpp version also for vpe
	)
	vppcalls.AddHandlerVersion(vpp2001.Version, msgs.AllMessages(), NewSRv6VppHandler)
}

// SRv6VppHandler is accessor for SRv6-related vppcalls methods
type SRv6VppHandler struct {
	core_vppcalls.VppCoreAPI

	log          logging.Logger
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
}

// NewSRv6VppHandler creates new instance of SRv6 vppcalls handler
func NewSRv6VppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) vppcalls.SRv6VppAPI {
	vppChan, err := c.NewAPIChannel()
	if err != nil {
		logging.Warnf("failed to create API channel")
		return nil
	}
	return &SRv6VppHandler{
		callsChannel: vppChan,
		ifIndexes:    ifIdx,
		log:          log,
		VppCoreAPI:   core_vpp2001.NewVpeHandler(vppChan),
	}
}
