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

package configurator

import (
	"runtime/trace"

	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/util"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	rpc "go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
)

// configuratorServer implements DataSyncer service.
type configuratorServer struct {
	dumpService
	notifyService

	log      logging.Logger
	dispatch orchestrator.Dispatcher
}

// Get retrieves actual configuration data.
func (svc *configuratorServer) Get(context.Context, *rpc.GetRequest) (*rpc.GetResponse, error) {
	defer trackOperation("Get")()

	config := newConfig()

	util.PlaceProtos(svc.dispatch.ListData(),
		config.LinuxConfig,
		config.VppConfig,
		config.NetallocConfig,
	)

	return &rpc.GetResponse{Config: config}, nil
}

// Update adds configuration data present in data request to the VPP/Linux
func (svc *configuratorServer) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	ctx, task := trace.NewTask(ctx, "grpc.Update")
	defer task.End()
	trace.Logf(ctx, "updateData", "%+v", req)

	defer trackOperation("Update")()

	protos := util.ExtractProtos(
		req.GetUpdate().GetVppConfig(),
		req.GetUpdate().GetLinuxConfig(),
		req.GetUpdate().GetNetallocConfig(),
	)

	var kvPairs []orchestrator.KeyVal
	for _, p := range protos {
		key, err := models.GetKey(p)
		if err != nil {
			svc.log.Debug("models.GetKey failed: %s", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		kvPairs = append(kvPairs, orchestrator.KeyVal{
			Key: key,
			Val: p,
		})
	}

	if req.FullResync {
		ctx = kvs.WithResync(ctx, kvs.FullResync, true)
	}

	ctx = orchestrator.DataSrcContext(ctx, "grpc")
	if _, err := svc.dispatch.PushData(ctx, kvPairs); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &rpc.UpdateResponse{}, nil
}

// Delete removes configuration data present in data request from the VPP/linux
func (svc *configuratorServer) Delete(ctx context.Context, req *rpc.DeleteRequest) (*rpc.DeleteResponse, error) {
	defer trackOperation("Delete")()

	protos := util.ExtractProtos(
		req.GetDelete().GetVppConfig(),
		req.GetDelete().GetLinuxConfig(),
		req.GetDelete().GetNetallocConfig(),
	)

	var kvPairs []orchestrator.KeyVal
	for _, p := range protos {
		key, err := models.GetKey(p)
		if err != nil {
			svc.log.Debug("models.GetKey failed: %s", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		kvPairs = append(kvPairs, orchestrator.KeyVal{
			Key: key,
			Val: nil,
		})
	}

	ctx = orchestrator.DataSrcContext(ctx, "grpc")
	if _, err := svc.dispatch.PushData(ctx, kvPairs); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &rpc.DeleteResponse{}, nil
}

func newConfig() *rpc.Config {
	return &rpc.Config{
		LinuxConfig:    &linux.ConfigData{},
		VppConfig:      &vpp.ConfigData{},
		NetallocConfig: &netalloc.ConfigData{},
	}
}
