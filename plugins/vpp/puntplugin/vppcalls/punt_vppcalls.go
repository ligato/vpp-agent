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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
)

// PuntDetails includes proto-modelled punt object and its socket path
type PuntDetails struct {
	PuntData   *punt.ToHost
	SocketPath string
}

type ReasonDetails struct {
	Reason *punt.Reason
	ID     uint32
}

type ExceptionDetails struct {
	Exception *punt.Exception
}

// PuntVppAPI provides methods for managing VPP punt configuration.
type PuntVppAPI interface {
	PuntVPPRead

	// AddPunt configures new punt to the host from the VPP
	AddPunt(punt *punt.ToHost) error
	// DeletePunt removes or unregisters punt entry
	DeletePunt(punt *punt.ToHost) error
	// RegisterPuntSocket registers new punt to unix domain socket entry
	RegisterPuntSocket(puntCfg *punt.ToHost) (string, error)
	// DeregisterPuntSocket removes existing punt to socket registration
	DeregisterPuntSocket(puntCfg *punt.ToHost) error
	// AddPuntRedirect adds new punt IP redirect entry
	AddPuntRedirect(punt *punt.IPRedirect) error
	// DeletePuntRedirect removes existing redirect entry
	DeletePuntRedirect(punt *punt.IPRedirect) error
	// AddPuntException configures new punt exception
	AddPuntException(punt *punt.Exception) error
	// DeletePuntException removes punt exception entry
	DeletePuntException(punt *punt.Exception) error
}

// PuntVPPRead provides read methods for punt
type PuntVPPRead interface {
	// DumpPuntRegisteredSockets returns all punt socket registrations known to the VPP agent
	DumpRegisteredPuntSockets() ([]*PuntDetails, error)
	// DumpExceptions dumps punt exceptions
	DumpExceptions() ([]*ExceptionDetails, error)
	// DumpPuntReasons returns all known punt reasons from VPP
	DumpPuntReasons() ([]*ReasonDetails, error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, ifaceidx.IfaceMetadataIndex, logging.Logger) PuntVppAPI
}

func CompatiblePuntVppHandler(
	ch govppapi.Channel, idx ifaceidx.IfaceMetadataIndex, log logging.Logger,
) PuntVppAPI {
	if len(Versions) == 0 {
		// puntplugin is not loaded
		return nil
	}
	for ver, h := range Versions {
		log.Debugf("checking compatibility with %s", ver)
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch, idx, log)
	}
	panic("no compatible version available")
}
