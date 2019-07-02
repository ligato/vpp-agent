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

// +build !mockvpp,!vppapiclient

package govppmux

import (
	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/adapter/statsclient"
	"github.com/ligato/cn-infra/logging"
)

// NewVppAdapter returns VPP binary API adapter, implemented as pure Go client.
func NewVppAdapter(addr string, useShm bool) adapter.VppAPI {
	if useShm {
		logging.Warnf(`Using shared memory for VPP binary API is not currently supported in pure Go client!

	To use socket client for VPP binary API:
	  - unset GOVPPMUX_NOSOCK environment variable
	  - remove these settings from govpp.conf config: shm-prefix, connect-via-shm

	If you still want to use shared memory for VPP binary API (not recommended):
	  - compile your agent with this build tag: vppapiclient
`)
		panic("No implementation for shared memory in pure Go client!")
	}
	// addr is used as socket path
	return socketclient.NewVppClient(addr)

}

// NewStatsAdapter returns VPP stats API adapter, implemented as pure Go client.
func NewStatsAdapter(socketName string) adapter.StatsAPI {
	return statsclient.NewStatsClient(socketName)
}
