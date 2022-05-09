// Copyright (c) 2022 Pantheon.tech
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

package vpp2202

import (
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/abf"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/acl"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/af_packet"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/arp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/bond"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/dns"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/flowprobe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/gtpu"
	interfaces "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ip6_nd"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ip_neighbor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ipfix_export"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ipip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/l3xc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/memclnt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/nat44_ed"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/nat44_ei"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/punt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/rd_cp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/rdma"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/sr"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/stn"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/teib"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/vlib"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/vmxnet3"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/vrrp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/vxlan_gpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/wireguard"
)

// Version is used to identify VPP binapi version
const Version = "22.02"

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
			vlib.AllMessages,
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
			nat44_ed.AllMessages,
			nat44_ei.AllMessages,
			rdma.AllMessages,
			stn.AllMessages,
			vmxnet3.AllMessages,
			vrrp.AllMessages,
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
//go:generate binapigen --input-file=$VPP_API_DIR/core/vlib.api.json
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
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/nat44_ed.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/nat44_ei.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/rdma.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/stn.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/vmxnet3.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/wireguard.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/vrrp.api.json
