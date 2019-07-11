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

package vpp1901

//go:generate -command binapigen binapi-generator --output-dir=.

//go:generate binapi-generator --input-file=/usr/share/vpp/api/abf.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/acl.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/af_packet.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/bfd.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/bond.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/dhcp.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/interface.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/ip.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/ipsec.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/l2.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/memclnt.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/memif.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/nat.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/punt.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/session.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/sr.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/stn.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/tap.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/tapv2.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vmxnet3.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vpe.api.json
//go:generate binapi-generator --input-file=/usr/share/vpp/api/vxlan.api.json
