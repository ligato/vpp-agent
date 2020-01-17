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

package orchestrator

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

type genericService struct {
	log      logging.Logger
	dispatch Dispatcher
}

func (s *genericService) KnownModels(ctx context.Context, req *generic.KnownModelsRequest) (*generic.KnownModelsResponse, error) {
	var infos []*generic.ModelDetail
	for _, model := range models.RegisteredModels() {
		if req.Class == "" || model.Spec().Class == req.Class {
			infos = append(infos, model.ModelDetail())
		}
	}
	resp := &generic.KnownModelsResponse{
		KnownModels: infos,
	}
	return resp, nil
}

func (s *genericService) SetConfig(ctx context.Context, req *generic.SetConfigRequest) (*generic.SetConfigResponse, error) {
	s.log.Debug("------------------------------")
	s.log.Debugf("=> GenericMgr.SetConfig: %d items", len(req.Updates))
	s.log.Debug("------------------------------")
	for _, item := range req.Updates {
		s.log.Debugf(" - %v", item)
	}
	s.log.Debug("------------------------------")

	var ops = make(map[string]generic.UpdateResult_Operation)
	var kvPairs []KeyVal

	for _, update := range req.Updates {
		item := update.Item
		if item == nil {
			return nil, status.Error(codes.InvalidArgument, "change item is nil")
		}
		var (
			key string
			val proto.Message
		)

		var err error
		if item.Data != nil {
			val, err = models.UnmarshalItem(item)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			key, err = models.GetKey(val)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			ops[key] = generic.UpdateResult_UPDATE
		} else if item.Id != nil {
			model, err := models.GetModelForItem(item)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			key = model.KeyPrefix() + item.Id.Name
			ops[key] = generic.UpdateResult_DELETE
		} else {
			return nil, status.Error(codes.InvalidArgument, "ProtoItem has no key or val defined.")
		}
		kvPairs = append(kvPairs, KeyVal{
			Key: key,
			Val: val,
		})
	}

	ctx = DataSrcContext(ctx, "grpc")
	if req.OverwriteAll {
		ctx = kvs.WithResync(ctx, kvs.FullResync, true)
	}
	ctx = kvs.WithRetryDefault(ctx)
	results, err := s.dispatch.PushData(ctx, kvPairs)
	if err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	updateResults := []*generic.UpdateResult{}
	for _, res := range results {
		var msg string
		if details := res.Status.GetDetails(); len(details) > 0 {
			msg = strings.Join(res.Status.GetDetails(), ", ")
		} else {
			msg = res.Status.GetError()
		}
		updateResults = append(updateResults, &generic.UpdateResult{
			Key: res.Key,
			Status: &generic.ItemStatus{
				Status:  res.Status.State.String(),
				Message: msg,
			},
			//Op: res.Status.LastOperation.String(),
		})
	}

	/*
		// commit the transaction
		if err := txn.Commit(); err != nil {
			st := status.New(codes.FailedPrecondition, err.Error())
			return nil, st.Err()
			// TODO: use the WithDetails to return extra info to clients.
			//ds, err := st.WithDetails(&rpc.DebugInfo{Detail: "Local transaction failed!"})
			//if err != nil {
			//	return nil, st.Err()
			//}
			//return nil, ds.Err()
		}
	*/

	return &generic.SetConfigResponse{Results: updateResults}, nil
}

func (s *genericService) GetConfig(context.Context, *generic.GetConfigRequest) (*generic.GetConfigResponse, error) {
	var items []*generic.ConfigItem

	for key, data := range s.dispatch.ListData() {
		item, err := models.MarshalItem(data)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		var itemStatus *generic.ItemStatus
		st, err := s.dispatch.GetStatus(key)
		if err != nil {
			s.log.Warnf("GetStatus failed: %v", err)
		} else {
			var msg string
			if details := st.GetDetails(); len(details) > 0 {
				msg = strings.Join(st.GetDetails(), ", ")
			} else {
				msg = st.GetError()
			}
			itemStatus = &generic.ItemStatus{
				Status:  st.GetState().String(),
				Message: msg,
			}
		}
		items = append(items, &generic.ConfigItem{
			Item:   item,
			Status: itemStatus,
		})
	}

	return &generic.GetConfigResponse{Items: items}, nil
}

func (s *genericService) DumpState(context.Context, *generic.DumpStateRequest) (*generic.DumpStateResponse, error) {
	pairs, err := s.dispatch.ListState()
	if err != nil {
		return nil, err
	}

	fmt.Printf("dispatch.ListState: %d pairs", len(pairs))
	for key, val := range pairs {
		fmt.Printf(" - [%s] %+v\n", key, proto.CompactTextString(val))
	}

	var states []*generic.StateItem
	for _, kv := range pairs {
		item, err := models.MarshalItem(kv)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		md := map[string]string{}
		states = append(states, &generic.StateItem{
			Item:     item,
			Metadata: md,
		})
	}

	return &generic.DumpStateResponse{Items: states}, nil
}

func (s *genericService) Subscribe(req *generic.SubscribeRequest, server generic.ManagerService_SubscribeServer) error {
	return status.Error(codes.Unimplemented, "Not implemented yet")
}
