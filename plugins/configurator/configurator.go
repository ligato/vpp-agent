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
	"context"
	"fmt"
	"reflect"
	"runtime/trace"
	"strconv"
	"time"

	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/util"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator/contextdecorator"
	pb "go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	"go.ligato.io/vpp-agent/v3/proto/ligato/linux"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
)

const (
	waitDoneCheckPendingPeriod = time.Millisecond * 10
)

// configuratorServer implements DataSyncer service.
type configuratorServer struct {
	pb.UnimplementedConfiguratorServiceServer

	dumpService
	notifyService

	log      logging.Logger
	dispatch orchestrator.Dispatcher
}

func (svc *configuratorServer) Dump(ctx context.Context, req *pb.DumpRequest) (*pb.DumpResponse, error) {
	return svc.dumpService.Dump(ctx, req)
}

func (svc *configuratorServer) Notify(from *pb.NotifyRequest, server pb.ConfiguratorService_NotifyServer) error {
	return svc.notifyService.Notify(from, server)
}

// Get retrieves actual configuration data.
func (svc *configuratorServer) Get(context.Context, *pb.GetRequest) (*pb.GetResponse, error) {
	defer trackOperation("Get")()

	config := newConfig()

	util.PlaceProtos(svc.dispatch.ListData(),
		config.LinuxConfig,
		config.VppConfig,
		config.NetallocConfig,
	)

	return &pb.GetResponse{Config: config}, nil
}

// Update adds configuration data present in data request to the VPP/Linux
func (svc *configuratorServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
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
			svc.log.WithFields(map[string]interface{}{
				"message": proto.MessageName(p),
				"type":    reflect.TypeOf(p).Elem().Name(),
			}).Debug("models.GetKey error: %s", err)
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

	md, hasMeta := metadata.FromIncomingContext(ctx)
	if hasMeta && len(md["datasrc"]) == 1 {
		ctx = contextdecorator.DataSrcContext(ctx, md["datasrc"][0])
	} else {
		ctx = contextdecorator.DataSrcContext(ctx, "grpc")
	}
	results, err := svc.dispatch.PushData(ctx, kvPairs, nil)

	header := map[string]string{}
	if seqNum := svc.extractTxnSeqNum(results); seqNum >= 0 {
		header["seqnum"] = fmt.Sprint(seqNum)
	}
	if err := grpc.SetHeader(ctx, metadata.New(header)); err != nil {
		logging.Warnf("sending grpc header failed: %v", err)
	}
	if err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	if req.WaitDone {
		waitStart := time.Now()
		var pendingKeys []string
		for _, res := range results {
			if res.Status.GetState() == kvscheduler.ValueState_PENDING {
				pendingKeys = append(pendingKeys, res.Key)
			}
		}
		if len(pendingKeys) > 0 {
			svc.log.Infof("waiting for %d pending keys", len(pendingKeys))
			for len(pendingKeys) > 0 {
				select {
				case <-time.After(waitDoneCheckPendingPeriod):
					pendingKeys = svc.listPending(pendingKeys)
				case <-ctx.Done():
					svc.log.Warnf("update returning before %d pending keys are done: %v", len(pendingKeys), ctx.Err())
					return nil, ctx.Err()
				}
			}
		} else {
			svc.log.Debugf("no pendings keys to wait for")
		}
		svc.log.Infof("finished waiting for done (took %v)", time.Since(waitStart))
	}

	svc.log.Debugf("config update finished with %d results", len(results))

	return &pb.UpdateResponse{}, nil
}

// Delete removes configuration data present in data request from the VPP/linux
func (svc *configuratorServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
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
			svc.log.WithFields(map[string]interface{}{
				"message": proto.MessageName(p),
				"type":    reflect.TypeOf(p).Elem().Name(),
			}).Debug("models.GetKey error: %s", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		kvPairs = append(kvPairs, orchestrator.KeyVal{
			Key: key,
			Val: nil, // delete
		})
	}

	md, hasMeta := metadata.FromIncomingContext(ctx)
	if hasMeta && len(md["datasrc"]) == 1 {
		ctx = contextdecorator.DataSrcContext(ctx, md["datasrc"][0])
	} else {
		ctx = contextdecorator.DataSrcContext(ctx, "grpc")
	}
	results, err := svc.dispatch.PushData(ctx, kvPairs, nil)

	header := map[string]string{}
	if seqNum := svc.extractTxnSeqNum(results); seqNum >= 0 {
		header["seqnum"] = fmt.Sprint(seqNum)
	}
	if err := grpc.SendHeader(ctx, metadata.New(header)); err != nil {
		logging.Warnf("sending grpc header failed: %v", err)
	}
	if err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	if req.WaitDone {
		waitStart := time.Now()
		var pendingKeys []string
		for _, res := range results {
			if res.Status.GetState() == kvscheduler.ValueState_PENDING {
				pendingKeys = append(pendingKeys, res.Key)
			}
		}
		if len(pendingKeys) > 0 {
			svc.log.Infof("waiting for %d pending keys", len(pendingKeys))
			for len(pendingKeys) > 0 {
				select {
				case <-time.After(waitDoneCheckPendingPeriod):
					pendingKeys = svc.listPending(pendingKeys)
				case <-ctx.Done():
					svc.log.Warnf("update returning before %d pending keys are done: %v", len(pendingKeys), ctx.Err())
					return nil, ctx.Err()
				}
			}
		} else {
			svc.log.Debugf("no pendings keys to wait for")
		}
		svc.log.Infof("finished waiting for done (took %v)", time.Since(waitStart))
	}

	svc.log.Debugf("config delete finished with %d results", len(results))

	return &pb.DeleteResponse{}, nil
}

func (svc *configuratorServer) listPending(keys []string) []string {
	var pending []string
	for _, key := range keys {
		st, err := svc.dispatch.GetStatus(key)
		if err != nil {
			svc.log.Debugf("dispatch.GetStatus for key %q error: %v", key, err)
			continue
		}
		if st.GetState() == kvscheduler.ValueState_PENDING {
			pending = append(pending, key)
		}
	}
	return pending
}

func (svc *configuratorServer) extractTxnSeqNum(results []orchestrator.Result) int {
	seqNum := -1
	for _, result := range results {
		if result.Key == "seqnum" {
			str := result.Status.Details[0]
			if n, err := strconv.Atoi(str); err == nil {
				seqNum = n
			} else {
				svc.log.Debugf("invalid seqnum in result: %q", str)
			}
		}
	}
	return seqNum
}

func newConfig() *pb.Config {
	return &pb.Config{
		LinuxConfig:    &linux.ConfigData{},
		VppConfig:      &vpp.ConfigData{},
		NetallocConfig: &netalloc.ConfigData{},
	}
}
