package dataconfigurator

import (
	"github.com/gogo/status"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"

	rpc "github.com/ligato/vpp-agent/api/dataconfigurator"
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/plugins/dispatcher"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

// configuratorServer implements DataSyncer service.
type configuratorServer struct {
	log      logging.Logger
	dispatch *dispatcher.Plugin
	dumpService
}

func (svc *configuratorServer) Get(context.Context, *rpc.GetRequest) (*rpc.GetResponse, error) {
	config := newData()

	dispatcher.PlaceProtos(svc.dispatch.ListData(), config.Linux, config.Vpp)

	return &rpc.GetResponse{Config: config}, nil
}

// Update adds configuration data present in data request to the VPP/Linux
func (svc *configuratorServer) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	protos := dispatcher.ExtractProtos(req.Update.Vpp, req.Update.Linux)

	var kvPairs []datasync.ProtoWatchResp
	for _, p := range protos {
		key, err := models.GetKey(p)
		if err != nil {
			svc.log.Debug("models.GetKey failed: %s", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		kvPairs = append(kvPairs, &dispatcher.ProtoWatchResp{
			Key: key,
			Val: p,
		})
	}

	if req.FullResync {
		ctx = kvs.WithResync(ctx, kvs.FullResync, true)
	}

	if err, _ := svc.dispatch.PushData(ctx, kvPairs); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &rpc.UpdateResponse{}, nil
}

// Delete removes configuration data present in data request from the VPP/linux
func (svc *configuratorServer) Delete(ctx context.Context, req *rpc.DeleteRequest) (*rpc.DeleteResponse, error) {
	protos := dispatcher.ExtractProtos(req.Delete.Vpp, req.Delete.Linux)

	var kvPairs []datasync.ProtoWatchResp
	for _, p := range protos {
		key, err := models.GetKey(p)
		if err != nil {
			svc.log.Debug("models.GetKey failed: %s", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		kvPairs = append(kvPairs, &dispatcher.ProtoWatchResp{
			Key: key,
			Val: nil,
		})
	}

	if err, _ := svc.dispatch.PushData(ctx, kvPairs); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &rpc.DeleteResponse{}, nil
}
