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

package vpp2001

import (
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/abf"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/acl"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/af_packet"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/bond"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/gtpu"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/l3xc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/memclnt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/nat"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/punt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/sr"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/stn"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vmxnet3"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vxlan_gpe"
)

// Version is used to identify VPP binapi version
const Version = "20.01-rc0~379"

func init() {
	binapi.Versions[Version] = binapi.VersionMsgs{
		Core: vpp.Messages(
			af_packet.AllMessages,
			bond.AllMessages,
			dhcp.AllMessages,
			gre.AllMessages,
			interfaces.AllMessages,
			ip.AllMessages,
			ipsec.AllMessages,
			l2.AllMessages,
			memclnt.AllMessages,
			punt.AllMessages,
			span.AllMessages,
			sr.AllMessages,
			tapv2.AllMessages,
			vpe.AllMessages,
			vxlan.AllMessages,
			vxlan_gpe.AllMessages,
		),
		Plugins: vpp.Messages(
			abf.AllMessages,
			acl.AllMessages,
			gtpu.AllMessages,
			l3xc.AllMessages,
			memif.AllMessages,
			nat.AllMessages,
			stn.AllMessages,
			vmxnet3.AllMessages,
		),
	}
}

//go:generate binapi-generator --output-dir=. --input-file=$VPP_API_DIR/core/ethernet_types.api.json
//go:generate binapi-generator --output-dir=. --input-file=$VPP_API_DIR/core/interface_types.api.json
//go:generate binapi-generator --output-dir=. --input-file=$VPP_API_DIR/core/ip_types.api.json
//go:generate binapi-generator --output-dir=. --input-file=$VPP_API_DIR/core/fib_types.api.json --input-types=$VPP_API_DIR/core/ethernet_types.api.json,$VPP_API_DIR/core/interface_types.api.json,$VPP_API_DIR/core/ip_types.api.json
//go:generate binapi-generator --output-dir=. --input-file=$VPP_API_DIR/core/ipsec_types.api.json --input-types=$VPP_API_DIR/core/ethernet_types.api.json,$VPP_API_DIR/core/interface_types.api.json,$VPP_API_DIR/core/ip_types.api.json
//go:generate binapi-generator --output-dir=. --input-file=$VPP_API_DIR/core/vpe_types.api.json

//go:generate -command binapigen binapi-generator --output-dir=. --input-types=$VPP_API_DIR/core/ethernet_types.api.json,$VPP_API_DIR/core/interface_types.api.json,$VPP_API_DIR/core/ip_types.api.json,$VPP_API_DIR/core/fib_types.api.json,$VPP_API_DIR/core/ipsec_types.api.json,$VPP_API_DIR/core/vpe_types.api.json

//go:generate binapigen --input-file=$VPP_API_DIR/core/af_packet.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/arp.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/bond.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/gre.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/interface.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ip.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ip_neighbor.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ipsec.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/l2.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/memclnt.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/punt.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/span.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/sr.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/tapv2.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/vpe.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/vxlan.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/vxlan_gpe.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/abf.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/acl.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/gtpu.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/dhcp.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/l3xc.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/memif.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/nat.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/stn.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/vmxnet3.api.json
