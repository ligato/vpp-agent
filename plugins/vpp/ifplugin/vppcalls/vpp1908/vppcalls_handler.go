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

package vpp1908

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/af_packet"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/bond"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/gtpu"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vmxnet3"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
)

var HandlerVersion = vpp.HandlerVersion{
	Version: vpp1908.Version,
	Check: func(c vpp.Client) error {
		msgs := vpp.Messages(
			af_packet.AllMessages,
			bond.AllMessages,
			dhcp.AllMessages,
			interfaces.AllMessages,
			ip.AllMessages,
			ipsec.AllMessages,
			gre.AllMessages,
			l2.AllMessages,
			span.AllMessages,
			tapv2.AllMessages,
			vxlan.AllMessages,
		)
		if c.IsPluginLoaded(gtpu.ModuleName) {
			msgs.Add(gtpu.AllMessages)
		}
		if c.IsPluginLoaded(memif.ModuleName) {
			msgs.Add(memif.AllMessages)
		}
		if c.IsPluginLoaded(vmxnet3.ModuleName) {
			msgs.Add(vmxnet3.AllMessages)
		}
		return c.CheckCompatiblity(msgs.AllMessages()...)
	},
	NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
		return NewInterfaceVppHandler(c, a[0].(logging.Logger))
	},
}

func init() {
	vppcalls.Handler.AddVersion(HandlerVersion)
}

// InterfaceVppHandler is accessor for interface-related vppcalls methods
type InterfaceVppHandler struct {
	callsChannel govppapi.Channel
	interfaces   interfaces.RPCService
	ipsec        ipsec.RPCService
	gtpu         gtpu.RPCService
	memif        memif.RPCService
	vmxnet3      vmxnet3.RPCService
	log          logging.Logger
}

// NewInterfaceVppHandler returns new InterfaceVppHandler.
func NewInterfaceVppHandler(c vpp.Client, log logging.Logger) vppcalls.InterfaceVppAPI {
	ch, err := c.NewAPIChannel()
	if err != nil {
		return nil
	}
	h := &InterfaceVppHandler{
		callsChannel: ch,
		interfaces:   interfaces.NewServiceClient(ch),
		ipsec:        ipsec.NewServiceClient(ch),
		log:          log,
	}
	if c.IsPluginLoaded(gtpu.ModuleName) {
		h.gtpu = gtpu.NewServiceClient(ch)
	}
	if c.IsPluginLoaded(memif.ModuleName) {
		h.memif = memif.NewServiceClient(ch)
	}
	if c.IsPluginLoaded(vmxnet3.ModuleName) {
		h.vmxnet3 = vmxnet3.NewServiceClient(ch)
	}
	return h
}
