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

// +build !mockvpp,vppapiclient

package govppmux

import (
	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/adapter/vppapiclient"
)

// NewVppAdapter returns VPP binary API adapter, implemented as wrapper for vppapiclient library.
func NewVppAdapter(addr string, useShm bool) adapter.VppAPI {
	if useShm {
		// addr is used as shm prefix
		return vppapiclient.NewVppClient(addr)
	}
	// addr is used as socket path
	return socketclient.NewVppClient(addr)
}

// NewStatsAdapter returns VPP stats API adapter, implemented as wrapper for vppapiclient library.
func NewStatsAdapter(socketName string) adapter.StatsAPI {
	return vppapiclient.NewStatClient(socketName)
}
