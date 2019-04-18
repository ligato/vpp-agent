// Copyright (c) 2017 Cisco and/or its affiliates.
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

package govppmux

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
)

// StatsAPI is extended API with ability to get VPP stats data
type StatsAPI interface {
	API

	govppapi.StatsProvider
}

// API for other plugins to get connectivity to VPP.
type API interface {
	// VPPInfo returns VPP information which is retrieved immediatelly after connecting to VPP.
	VPPInfo() (VPPInfo, error)

	govppapi.ChannelProvider
}

// VPPInfo defines retrieved information about the connected VPP instance.
type VPPInfo struct {
	Connected bool
	vppcalls.VersionInfo
	vppcalls.VpeInfo
}

// GetReleaseVersion returns VPP release version (XX.YY), which is normalized from VersionInfo.
func (vpp VPPInfo) GetReleaseVersion() string {
	if len(vpp.Version) < 5 {
		return ""
	}
	return vpp.Version[:5]
}
