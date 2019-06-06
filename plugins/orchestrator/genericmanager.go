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

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/status"
	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"

	api "github.com/ligato/vpp-agent/api/genericmanager"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

type genericManagerSvc struct {
	log      logging.Logger
	dispatch Dispatcher
}

func (s *genericManagerSvc) Capabilities(ctx context.Context, req *api.CapabilitiesRequest) (*api.CapabilitiesResponse, error) {
	resp := &api.CapabilitiesResponse{
		KnownModels: models.RegisteredModels(),
	}
	return resp, nil
}

func (s *genericManagerSvc) SetConfig(ctx context.Context, req *api.SetConfigRequest) (*api.SetConfigResponse, error) {
	s.log.Debug("------------------------------")
	s.log.Debugf("=> GenericMgr.SetConfig: %d items", len(req.Updates))
	s.log.Debug("------------------------------")
	for _, item := range req.Updates {
		s.log.Debugf(" - %v", item)
	}
	s.log.Debug("------------------------------")

	var ops = make(map[string]api.UpdateResult_Operation)
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
			ops[key] = api.UpdateResult_UPDATE
		} else if item.Id != nil {
			model, err := models.ModelForItem(item)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			key = model.KeyPrefix() + item.Id.Name
			ops[key] = api.UpdateResult_DELETE
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
	results, err := s.dispatch.PushData(ctx, kvPairs)
	if err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	updateResults := []*api.UpdateResult{}
	for _, res := range results {
		var msg string
		if details := res.Status.GetDetails(); len(details) > 0 {
			msg = strings.Join(res.Status.GetDetails(), ", ")
		} else {
			msg = res.Status.GetError()
		}
		updateResults = append(updateResults, &api.UpdateResult{
			Key: res.Key,
			Status: &api.ItemStatus{
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

	return &api.SetConfigResponse{Results: updateResults}, nil
}

func (s *genericManagerSvc) GetConfig(context.Context, *api.GetConfigRequest) (*api.GetConfigResponse, error) {
	var items []*api.ConfigItem

	for key, data := range s.dispatch.ListData() {
		item, err := models.MarshalItem(data)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		var itemStatus *api.ItemStatus
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
			itemStatus = &api.ItemStatus{
				Status:  st.GetState().String(),
				Message: msg,
			}
		}
		items = append(items, &api.ConfigItem{
			Item:   item,
			Status: itemStatus,
		})
	}

	return &api.GetConfigResponse{Items: items}, nil
}

func (s *genericManagerSvc) DumpState(context.Context, *api.DumpStateRequest) (*api.DumpStateResponse, error) {
	pairs, err := s.dispatch.ListState()
	if err != nil {
		return nil, err
	}

	fmt.Printf("dispatch.ListState: %d pairs", len(pairs))
	for key, val := range pairs {
		fmt.Printf(" - [%s] %+v\n", key, proto.CompactTextString(val))
	}

	var states []*api.StateItem
	for _, kv := range pairs {
		item, err := models.MarshalItem(kv)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		meta := map[string]string{}
		states = append(states, &api.StateItem{
			Item:     item,
			Metadata: meta,
		})
	}

	return &api.DumpStateResponse{Items: states}, nil
}

func (s *genericManagerSvc) Subscribe(req *api.SubscribeRequest, server api.GenericManager_SubscribeServer) error {
	return status.Error(codes.Unimplemented, "Not implemented yet")
}
