// Copyright (c) 2020 Cisco and/or its affiliates.
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
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

const (
	// MinPathMTU and MaxPathMTU values were copied from VPP source code.
	// If something will change, please, be kind and update values here
	// and also update error messages in the IPFIX descriptor.
	MinPathMTU = 68
	MaxPathMTU = 1450
)

// IpfixVppAPI provides methods for managing VPP IPFIX configuration.
type IpfixVppAPI interface {
	SetExporter(conf *ipfix.IPFIX) error
	DumpExporters() ([]*ipfix.IPFIX, error)

	SetFPParams(conf *ipfix.FlowProbeParams) error

	AddFPFeature(conf *ipfix.FlowProbeFeature) error
	DelFPFeature(conf *ipfix.FlowProbeFeature) error
}

var handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "ipfix",
	HandlerAPI: (*IpfixVppAPI)(nil),
})

func AddIpfixHandlerVersion(version vpp.Version, msgs []govppapi.Message,
	h func(ch govppapi.Channel, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) IpfixVppAPI,
) {
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

func CompatibleIpfixVppHandler(c vpp.Client, ifIdx ifaceidx.IfaceMetadataIndex, log logging.Logger) IpfixVppAPI {
	if v := handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, ifIdx, log).(IpfixVppAPI)
	}
	return nil
}
