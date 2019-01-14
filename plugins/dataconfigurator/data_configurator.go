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

package dataconfigurator

import (
	"reflect"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/ligato/vpp-agent/api/models/vpp"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"

	"github.com/gogo/status"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	rpc "github.com/ligato/vpp-agent/api/dataconfigurator"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/orchestrator"
)

// Plugin registers VPP GRPC services in *grpc.Server.
type Plugin struct {
	Deps

	dataConfigurator DataConfigurator
}

// Deps - dependencies of Plugin
type Deps struct {
	infra.PluginDeps
	GRPCServer grpc.Server
	Orch       *orchestrator.Plugin
}

// Init sets plugin child loggers
func (p *Plugin) Init() error {
	p.dataConfigurator.log = p.Log.NewLogger("datasyncer")
	p.dataConfigurator.orch = p.Orch

	grpcServer := p.GRPCServer.GetServer()
	if grpcServer != nil {
		rpc.RegisterDataConfiguratorServer(grpcServer, &p.dataConfigurator)
	}

	return nil
}

// Close does nothing.
func (p *Plugin) Close() error {
	return nil
}

// DataConfigurator implements DataSyncer service.
type DataConfigurator struct {
	log  logging.Logger
	orch *orchestrator.Plugin
}

func (svc *DataConfigurator) Get(context.Context, *rpc.GetRequest) (*rpc.ConfigData, error) {
	config := &rpc.ConfigData{
		Linux: &linux.Config{},
		Vpp:   &vpp.Config{},
	}

	placeProtos(svc.orch.ListData(), config.Linux, config.Vpp)

	return config, nil
}

// Update adds configuration data present in data request to the VPP/Linux
func (svc *DataConfigurator) Update(ctx context.Context, data *rpc.ConfigData) (*rpc.UpdateResponse, error) {
	protos := extractProtos(data.Vpp, data.Linux)

	var kvPairs []datasync.ProtoWatchResp
	for _, p := range protos {
		key, err := models.GetKey(p)
		if err != nil {
			svc.log.Debug("models.GetKey failed: %s", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		kvPairs = append(kvPairs, &orchestrator.ProtoWatchResp{
			Key: key,
			Val: p,
		})
	}

	if err, _ := svc.orch.PushData(ctx, kvPairs); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &rpc.UpdateResponse{}, nil
}

// Delete removes configuration data present in data request from the VPP/linux
func (svc *DataConfigurator) Delete(ctx context.Context, data *rpc.ConfigData) (*rpc.DeleteResponse, error) {
	protos := extractProtos(data.Vpp, data.Linux)

	var kvPairs []datasync.ProtoWatchResp
	for _, p := range protos {
		key, err := models.GetKey(p)
		if err != nil {
			svc.log.Debug("models.GetKey failed: %s", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		kvPairs = append(kvPairs, &orchestrator.ProtoWatchResp{
			Key: key,
			Val: nil,
		})
	}

	if err, _ := svc.orch.PushData(ctx, kvPairs); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &rpc.DeleteResponse{}, nil
}

// Resync creates a resync request which adds data tp the VPP/linux
func (svc *DataConfigurator) Resync(ctx context.Context, data *rpc.ConfigData) (*rpc.ResyncResponse, error) {
	ctx = kvs.WithResync(ctx, kvs.FullResync, true)

	protos := extractProtos(data.Vpp, data.Linux)

	var kvPairs []datasync.ProtoWatchResp
	for _, p := range protos {
		key, err := models.GetKey(p)
		if err != nil {
			svc.log.Debug("models.GetKey failed: %s", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		kvPairs = append(kvPairs, &orchestrator.ProtoWatchResp{
			Key: key,
			Val: p,
		})
	}

	if err, _ := svc.orch.PushData(ctx, kvPairs); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &rpc.ResyncResponse{}, nil
}

func extractProtos(from ...interface{}) (protos []proto.Message) {
	for _, v := range from {
		val := reflect.ValueOf(v).Elem()
		typ := val.Type()
		if typ.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < typ.NumField(); i++ {
			field := val.Field(i)
			if field.Kind() == reflect.Slice {
				for idx := 0; idx < field.Len(); idx++ {
					elem := field.Index(idx)
					if msg, ok := elem.Interface().(proto.Message); ok {
						protos = append(protos, msg)
					}
				}
			} else {
				if msg, ok := field.Interface().(proto.Message); ok {
					protos = append(protos, msg)
				}
			}
		}
	}
	return
}

func placeProtos(protos map[string]models.ProtoItem, dsts ...interface{}) {
	for _, prot := range protos {
		protTyp := reflect.TypeOf(prot)
		for _, dst := range dsts {
			dstVal := reflect.ValueOf(dst).Elem()
			dstTyp := dstVal.Type()
			if dstTyp.Kind() != reflect.Struct {
				return
			}
			for i := 0; i < dstTyp.NumField(); i++ {
				field := dstVal.Field(i)
				if field.Kind() == reflect.Slice {
					if protTyp.AssignableTo(field.Type().Elem()) {
						field.Set(reflect.Append(field, reflect.ValueOf(prot)))
					}
				} else {
					if field.Type() == protTyp {
						field.Set(reflect.ValueOf(prot))
					}
				}
			}
		}
	}
	return
}
