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

package vpp2210

import (
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/abf"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/acl"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/af_packet"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/arp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/bond"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/dns"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/flowprobe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/gtpu"
	interfaces "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/ip6_nd"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/ip_neighbor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/ipfix_export"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/ipip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/l3xc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/memclnt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/nat44_ed"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/nat44_ei"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/punt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/rd_cp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/rdma"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/sr"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/stn"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/teib"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/vlib"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/vmxnet3"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/vrrp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/vxlan_gpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2210/wireguard"
)

// Version is used to identify VPP binapi version
const Version = "22.10"

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
//go:generate binapigen --input=$VPP_API_DIR/core
//go:generate binapigen --input=$VPP_API_DIR/plugins
