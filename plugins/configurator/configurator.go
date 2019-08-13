package configurator

import (
	"runtime/trace"

	"github.com/gogo/status"
	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"

	rpc "github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/api/models/linux"
	"github.com/ligato/vpp-agent/api/models/vpp"
	"github.com/ligato/vpp-agent/pkg/models"
	"github.com/ligato/vpp-agent/pkg/util"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/orchestrator"
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

	util.PlaceProtos(svc.dispatch.ListData(), config.LinuxConfig, config.VppConfig)

	return &rpc.GetResponse{Config: config}, nil
}

// Update adds configuration data present in data request to the VPP/Linux
func (svc *configuratorServer) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	ctx, task := trace.NewTask(ctx, "grpc.Update")
	defer task.End()
	trace.Logf(ctx, "updateData", "%+v", req)

	defer trackOperation("Update")()

	protos := util.ExtractProtos(req.Update.VppConfig, req.Update.LinuxConfig)

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

	protos := util.ExtractProtos(req.Delete.VppConfig, req.Delete.LinuxConfig)

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
		LinuxConfig: &linux.ConfigData{},
		VppConfig:   &vpp.ConfigData{},
	}
}
