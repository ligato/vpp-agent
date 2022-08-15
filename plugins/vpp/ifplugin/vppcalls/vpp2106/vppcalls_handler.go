//  Copyright (c) 2021 Cisco and/or its affiliates.
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

package vpp2106

import (
	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	vpp2106 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/af_packet"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/bond"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/gtpu"
	interfaces "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip6_nd"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/rd_cp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/rdma"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vmxnet3"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/wireguard"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
)

var HandlerVersion = vpp.HandlerVersion{
	Version: vpp2106.Version,
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
		if c.IsPluginLoaded(gtpu.APIFile) {
			msgs.Add(gtpu.AllMessages)
		}
		if c.IsPluginLoaded(memif.APIFile) {
			msgs.Add(memif.AllMessages)
		}
		if c.IsPluginLoaded(vmxnet3.APIFile) {
			msgs.Add(vmxnet3.AllMessages)
		}
		if c.IsPluginLoaded(wireguard.APIFile) {
			msgs.Add(wireguard.AllMessages)
		}
		if c.IsPluginLoaded(rdma.APIFile) {
			msgs.Add(rdma.AllMessages)
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
	rpcIP6nd     ip6_nd.RPCService
	rpcRdCp      rd_cp.RPCService
	wireguard    wireguard.RPCService
	rdma         rdma.RPCService
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
		interfaces:   interfaces.NewServiceClient(c),
		ipsec:        ipsec.NewServiceClient(c),
		rpcIP6nd:     ip6_nd.NewServiceClient(c),
		rpcRdCp:      rd_cp.NewServiceClient(c),
		log:          log,
	}
	if c.IsPluginLoaded(gtpu.APIFile) {
		h.gtpu = gtpu.NewServiceClient(c)
	}
	if c.IsPluginLoaded(memif.APIFile) {
		h.memif = memif.NewServiceClient(c)
	}
	if c.IsPluginLoaded(vmxnet3.APIFile) {
		h.vmxnet3 = vmxnet3.NewServiceClient(c)
	}
	if c.IsPluginLoaded(wireguard.APIFile) {
		h.wireguard = wireguard.NewServiceClient(c)
	}
	if c.IsPluginLoaded(rdma.APIFile) {
		h.rdma = rdma.NewServiceClient(c)
	}
	return h
}
