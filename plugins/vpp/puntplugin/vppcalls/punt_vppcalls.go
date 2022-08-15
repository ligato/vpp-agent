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
	"errors"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	punt "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/punt"
)

var (
	ErrUnsupported = errors.New("unsupported")
)

// PuntDetails includes punt model and socket path from VPP.
type PuntDetails struct {
	PuntData   *punt.ToHost
	SocketPath string
}

// ReasonDetails includes reason model and its matching ID from VPP.
type ReasonDetails struct {
	Reason *punt.Reason
	ID     uint32
}

// ExceptionDetails include punt model and socket path from VPP.
type ExceptionDetails struct {
	Exception  *punt.Exception
	SocketPath string
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
	// AddPuntException registers new punt exception
	AddPuntException(punt *punt.Exception) (string, error)
	// DeletePuntException deregisters punt exception entry
	DeletePuntException(punt *punt.Exception) error
}

// PuntVPPRead provides read methods for punt
type PuntVPPRead interface {
	// DumpRegisteredPuntSockets returns all punt socket registrations known to the VPP agent
	DumpRegisteredPuntSockets() ([]*PuntDetails, error)
	// DumpExceptions dumps punt exceptions
	DumpExceptions() ([]*ExceptionDetails, error)
	// DumpPuntReasons returns all known punt reasons from VPP
	DumpPuntReasons() ([]*ReasonDetails, error)
	// DumpPuntRedirect dump IP redirect punts
	DumpPuntRedirect() ([]*punt.IPRedirect, error)
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "punt",
	HandlerAPI: (*PuntVppAPI)(nil),
})

type NewHandlerFunc func(ch govppapi.Channel, idx ifaceidx.IfaceMetadataIndex, log logging.Logger) PuntVppAPI

func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return ch.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return h(ch, a[0].(ifaceidx.IfaceMetadataIndex), a[1].(logging.Logger))
		},
	})
}

func CompatiblePuntVppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) PuntVppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, log).(PuntVppAPI)
	}
	return nil
}
