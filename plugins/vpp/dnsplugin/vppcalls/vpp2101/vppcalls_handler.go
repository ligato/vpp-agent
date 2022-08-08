// Copyright (c) 2020 Pantheon.tech
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

package vpp2101

import (
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	core_vppcalls "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	core_vpp2101 "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp2101"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/dns"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls"
)

func init() {
	msgs := vpp.Messages(
		dns.AllMessages,
	)
	vppcalls.AddHandlerVersion(vpp2101.Version, msgs.AllMessages(), NewDNSVppHandler)
}

// DNSVppHandler is accessor for DNS-related vppcalls methods
type DNSVppHandler struct {
	core_vppcalls.VppCoreAPI

	log          logging.Logger
	callsChannel govppapi.Channel
}

// NewDNSVppHandler creates new instance of DNS vppcalls handler
func NewDNSVppHandler(c vpp.Client, log logging.Logger) vppcalls.DNSVppAPI {
	vppChan, err := c.NewAPIChannel()
	if err != nil {
		logging.Warnf("failed to create API channel")
		return nil
	}
	return &DNSVppHandler{
		callsChannel: vppChan,
		log:          log,
		VppCoreAPI:   core_vpp2101.NewVpeHandler(c),
	}
}
