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
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/abf"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/acl"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/af_packet"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/bond"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/dhcp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/gre"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/gtpu"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/interfaces"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ipip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/l2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/l3xc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/memclnt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/memif"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/nat"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/punt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/span"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/sr"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/stn"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/tapv2"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vmxnet3"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vxlan"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vxlan_gpe"
)

// Version is used to identify VPP version for binapi
const Version = "19.08.1"

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
			ipip.AllMessages,
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

//go:generate -command binapigen binapi-generator --output-dir=.

//go:generate binapigen --input-file=$VPP_API_DIR/core/af_packet.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/bond.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/dhcp.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/gre.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/interface.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/core/ip.api.json
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
//go:generate binapigen --input-file=$VPP_API_DIR/core/ipip.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/abf.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/acl.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/gtpu.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/l3xc.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/memif.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/nat.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/stn.api.json
//go:generate binapigen --input-file=$VPP_API_DIR/plugins/vmxnet3.api.json
