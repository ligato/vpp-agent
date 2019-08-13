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

//go:generate -command binapigen binapi-generator --output-dir=.

//go:generate binapigen --input-file=/usr/share/vpp/api/core/af_packet.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/bfd.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/bond.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/dhcp.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/interface.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/ip.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/ipsec.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/l2.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/memclnt.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/punt.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/session.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/sr.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/tapv2.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/vpe.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/core/vxlan.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/plugins/abf.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/plugins/acl.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/plugins/memif.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/plugins/nat.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/plugins/stn.api.json
//go:generate binapigen --input-file=/usr/share/vpp/api/plugins/vmxnet3.api.json

package vpp1904
