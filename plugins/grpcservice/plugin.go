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
	"github.com/gogo/status"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"

	"github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/models"
)

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

type syncService struct {
	log logging.Logger
}

func (s *syncService) Change(ctx context.Context, req *api.SyncRequest) (*api.SyncResponse, error) {
	s.log.Debug("------------------------------")
	s.log.Debugf("=> GRPC CHANGE: %d models", len(req.Models))
	s.log.Debug("------------------------------")

	// prepare a transaction
	txn := local.NewProtoTxn(Registry.PropagateChanges)

	if err := prepareTransaction(txn, req); err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	// commit the transaction
	if err := txn.Commit(); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()

		// TODO: use the WithDetails to propagate extra info to clients.
		/*ds, err := st.WithDetails(&rpc.DebugInfo{Detail: "Local transaction failed!"})
		if err != nil {
			return nil, st.Err()
		}
		return nil, ds.Err()*/
	}

	return &api.SyncResponse{}, nil
}

func (s *syncService) Resync(ctx context.Context, req *api.SyncRequest) (*api.SyncResponse, error) {
	s.log.Debug("------------------------------")
	s.log.Debugf("=> GRPC RESYNC: %d models", len(req.Models))
	s.log.Debug("------------------------------")

	// prepare a transaction
	txn := local.NewProtoTxn(Registry.PropagateResync)

	if err := prepareTransaction(txn, req); err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	// commit the transaction
	if err := txn.Commit(); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &api.SyncResponse{}, nil
}

func prepareTransaction(txn keyval.ProtoTxn, req *api.SyncRequest) error {
	for _, m := range req.Models {
		if m.Value != nil {
			pb, err := models.Unmarshal(m)
			if err != nil {
				return err
			}
			txn.Put(m.Key, pb)
		} else {
			txn.Delete(m.Key)
		}
	}
	return nil
}
