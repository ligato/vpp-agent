// Copyright (c) 2018 Cisco and/or its affiliates.
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
	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/vpp/model/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/puntidx"
)

// PuntVppAPI provides methods for managing VPP punt configuration.
type PuntVppAPI interface {
	PuntVPPWrite
	PuntVPPRead
}

// PuntVPPWrite provides write methods for punt
type PuntVPPWrite interface {
	// RegisterPuntSocket registers new punt to unix domain socket entry
	RegisterPuntSocket(puntCfg *punt.Punt) ([]byte, error)
	// DeregisterPuntSocket removes existing punt to socket registration
	DeregisterPuntSocket(puntCfg *punt.Punt) error
	// RegisterPuntSocketIPv6 registers new IPv6 punt to unix domain socket entry
	RegisterPuntSocketIPv6(puntCfg *punt.Punt) ([]byte, error)
	// DeregisterPuntSocketIPv6 removes existing IPv6 punt to socket registration
	DeregisterPuntSocketIPv6(puntCfg *punt.Punt) error
}

// PuntVPPRead provides read methods for punt
type PuntVPPRead interface {
	// DumpPuntRegisteredSockets returns all punt socket registrations known to the VPP agent
	// TODO since the API to dump sockets is missing, the method works only with the entries in local cache
	DumpPuntRegisteredSockets() (punts []*PuntDetails)
}

// PuntVppHandler is accessor for punt-related vppcalls methods.
type PuntVppHandler struct {
	callsChannel api.Channel
	mapping      puntidx.PuntIndex
	log          logging.Logger
}

// NewPuntVppHandler creates new instance of punt vppcalls handler
func NewPuntVppHandler(callsChan api.Channel, mapping puntidx.PuntIndex, log logging.Logger) *PuntVppHandler {
	return &PuntVppHandler{
		callsChannel: callsChan,
		mapping:      mapping,
		log:          log,
	}
}
