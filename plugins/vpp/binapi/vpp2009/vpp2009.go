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

package vpp2009

import (
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/abf"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/acl"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/af_packet"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/arp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/bond"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/dns"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/flowprobe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/gtpu"
	interfaces "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ip6_nd"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ip_neighbor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ipfix_export"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ipip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/l3xc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/memclnt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/nat"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/punt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/rd_cp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/rdma"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/sr"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/stn"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/teib"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/vmxnet3"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/vxlan_gpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/wireguard"
)

// Version is used to identify VPP binapi version
const Version = "20.09-rc0"

func init() {
	binapi.Versions[Version] = binapi.VersionMsgs{
		Core: vpp.Messages(
			af_packet.AllMessages,
			arp.AllMessages,
			bond.AllMessages,
			gre.AllMessages,
			interfaces.AllMessages,
			ip.AllMessages,
			ip6_nd.AllMessages,
			ip_neighbor.AllMessages,
			ipfix_export.AllMessages,
			ipip.AllMessages,
			ipsec.AllMessages,
			l2.AllMessages,
			memclnt.AllMessages,
			punt.AllMessages,
			rd_cp.AllMessages,
			span.AllMessages,
			sr.AllMessages,
			tapv2.AllMessages,
			teib.AllMessages,
			vpe.AllMessages,
			vxlan.AllMessages,
			vxlan_gpe.AllMessages,
		),
		Plugins: vpp.Messages(
			abf.AllMessages,
			acl.AllMessages,
			dhcp.AllMessages,
			dns.AllMessages,
			flowprobe.AllMessages,
			gtpu.AllMessages,
			l3xc.AllMessages,
			memif.AllMessages,
			nat.AllMessages,
			rdma.AllMessages,
			stn.AllMessages,
			vmxnet3.AllMessages,
			wireguard.AllMessages,
		),
	}
}

//go:generate -command binapigen binapi-generator --no-version-info --output-dir=.
//go:generate binapigen --input-file=$VPP_API_DIR/core/af_packet.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/arp.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/bond.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/gre.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/interface.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ip.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ip6_nd.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ip_neighbor.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ipfix_export.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ipip.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ipsec.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/l2.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/memclnt.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/punt.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/rd_cp.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/span.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/sr.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/tapv2.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/teib.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/vpe.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/vxlan.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/vxlan_gpe.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/abf.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/acl.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/dhcp.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/dns.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/flowprobe.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/gtpu.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/l3xc.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/memif.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/nat.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/rdma.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/stn.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/vmxnet3.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/wireguard.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/vrrp.api.json
