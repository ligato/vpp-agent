// Copyright (c) 2017 Cisco and/or its affiliates.
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

package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
)

// IPsecVppAPI provides methods for creating and managing of a IPsec configuration
type IPsecVppAPI interface {
	IPsecVppWrite
	IPsecVPPRead
}

// IPsecVppWrite provides write methods for IPsec
type IPsecVppWrite interface {
	// AddTunnelInterface adds tunnel interface
	AddTunnelInterface(tunnel *ipsec.TunnelInterfaces_Tunnel) (uint32, error)
	// DelTunnelInterface removes tunnel interface
	DelTunnelInterface(ifIdx uint32, tunnel *ipsec.TunnelInterfaces_Tunnel) error
	// AddSPD adds SPD to VPP via binary API
	AddSPD(spdID uint32) error
	// DelSPD deletes SPD from VPP via binary API
	DelSPD(spdID uint32) error
	// InterfaceAddSPD adds SPD interface assignment to VPP via binary API
	InterfaceAddSPD(spdID, swIfIdx uint32) error
	// InterfaceDelSPD deletes SPD interface assignment from VPP via binary API
	InterfaceDelSPD(spdID, swIfIdx uint32) error
	// AddSPDEntry adds SPD policy entry to VPP via binary API
	AddSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabases_SPD_PolicyEntry) error
	// DelSPDEntry deletes SPD policy entry from VPP via binary API
	DelSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabases_SPD_PolicyEntry) error
	// AddSAEntry adds SA to VPP via binary API
	AddSAEntry(saID uint32, sa *ipsec.SecurityAssociations_SA) error
	// DelSAEntry deletes SA from VPP via binary API
	DelSAEntry(saID uint32, sa *ipsec.SecurityAssociations_SA) error
}

// IPsecVppWrite provides read methods for IPsec
type IPsecVPPRead interface {
	// TODO define dump methods
}

// ipSecVppHandler is accessor for IPsec-related vppcalls methods
type ipSecVppHandler struct {
	stopwatch    *measure.Stopwatch
	callsChannel govppapi.Channel
	log          logging.Logger
}

// NewIPsecVppHandler creates new instance of IPsec vppcalls handler
func NewIPsecVppHandler(callsChan govppapi.Channel, log logging.Logger, stopwatch *measure.Stopwatch) (*ipSecVppHandler, error) {
	handler := &ipSecVppHandler{
		callsChannel: callsChan,
		stopwatch:    stopwatch,
		log:          log,
	}
	if err := handler.callsChannel.CheckMessageCompatibility(IPSecMessages...); err != nil {
		return nil, err
	}

	return handler, nil
}
