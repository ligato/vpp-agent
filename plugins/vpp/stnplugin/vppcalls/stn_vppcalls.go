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

	stn "github.com/ligato/vpp-agent/api/models/vpp/stn"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
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
	// DelSTNRule calls StnAddDelRule bin API with IsAdd=0
	DeleteSTNRule(stnRule *stn.Rule) error
}

// StnVppRead provides read methods for STN rules
type StnVppRead interface {
	// DumpSTNRules returns a list of all STN rules configured on the VPP
	DumpSTNRules() ([]*StnDetails, error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel, ifaceidx.IfaceMetadataIndex, logging.Logger) StnVppAPI
}

func CompatibleStnVppHandler(
	ch govppapi.Channel, idx ifaceidx.IfaceMetadataIndex, log logging.Logger,
) StnVppAPI {
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
