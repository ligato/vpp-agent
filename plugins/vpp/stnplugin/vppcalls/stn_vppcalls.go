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
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	stn "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/stn"
)

// StnDetails contains a proto-modelled STN data and VPP specific metadata
type StnDetails struct {
	Rule *stn.Rule
	Meta *StnMeta
}

// StnMeta contains an index of the interface defined by name in the STN rule
type StnMeta struct {
	IfIdx uint32
}

// StnVppAPI provides methods for managing STN rules
type StnVppAPI interface {
	StnVppRead

	// AddSTNRule calls StnAddDelRule bin API with IsAdd=1
	AddSTNRule(stnRule *stn.Rule) error
	// DeleteSTNRule calls StnAddDelRule bin API with IsAdd=0
	DeleteSTNRule(stnRule *stn.Rule) error
}

// StnVppRead provides read methods for STN rules
type StnVppRead interface {
	// DumpSTNRules returns a list of all STN rules configured on the VPP
	DumpSTNRules() ([]*StnDetails, error)
}

var handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "stn",
	HandlerAPI: (*StnVppAPI)(nil),
})

type NewHandlerFunc func(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) StnVppAPI

func AddStnHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	handler.AddVersion(vpp.HandlerVersion{
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

func CompatibleStnVppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) StnVppAPI {
	if v := handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, log).(StnVppAPI)
	}
	return nil
}
