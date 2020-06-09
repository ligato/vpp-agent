//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package binapi

import (
	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/binapi/syslog"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2005"
)

// This go generate directive will generate a binapi code for the VPP plugin syslog.
//go:generate binapi-generator --input-file=/usr/share/vpp/api/core/syslog.api.json

func init() {
	// This adds syslog API messages for the compatibility check of VPP 20.05.
	binapi2005 := binapi.Versions[vpp2005.Version]
	binapi2005.Core.Add(
		syslog.AllMessages,
	)
}
