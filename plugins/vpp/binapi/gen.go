//  Copyright (c) 2018 Cisco and/or its affiliates.
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

//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/af_packet.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/bfd.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/bond.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/dhcp.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/interface.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/ip.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/ipsec.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/l2.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/memclnt.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/punt.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/session.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/sr.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/tapv2.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/vpe.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/vxlan.api.json --output-dir=default

//go:generate binapi-generator --input-file=/usr/share/vpp/api/plugins/acl.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/plugins/memif.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/plugins/nat.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/plugins/stn.api.json --output-dir=default
//go:generate binapi-generator --input-file=/usr/share/vpp/api/plugins/vmxnet3.api.json --output-dir=default

package binapi
