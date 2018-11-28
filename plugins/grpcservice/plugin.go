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

package grpcservice

import (
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/grpc"

	"github.com/ligato/vpp-agent/api"
)

// Registry is used for propagating transactions.
var Registry = local.DefaultRegistry

// Plugin implements sync service for GRPC.
type Plugin struct {
	Deps

	syncSvc *syncService
}

// Deps represents dependencies for the plugin.
type Deps struct {
	infra.PluginDeps
	GRPC grpc.Server
}

// Init registers the service to GRPC server.
func (p *Plugin) Init() error {
	p.syncSvc = &syncService{p.Log}

	api.RegisterSyncServiceServer(p.GRPC.GetServer(), p.syncSvc)

	return nil
}
