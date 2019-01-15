package dataconfigurator

import (
	"reflect"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/status"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"

	rpc "github.com/ligato/vpp-agent/api/dataconfigurator"
	"github.com/ligato/vpp-agent/api/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/orchestrator"
)

// configService implements DataSyncer service.
type configService struct {
	log  logging.Logger
	orch *orchestrator.Plugin
}

func (svc *configService) Get(context.Context, *rpc.GetRequest) (*rpc.GetResponse, error) {
	config := newData()

	placeProtos(svc.orch.ListData(), config.Linux, config.Vpp)

	return &rpc.GetResponse{Config: config}, nil
}

// Update adds configuration data present in data request to the VPP/Linux
func (svc *configService) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	protos := extractProtos(req.Update.Vpp, req.Update.Linux)

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

	if req.FullResync {
		ctx = kvs.WithResync(ctx, kvs.FullResync, true)
	}

	if err, _ := svc.orch.PushData(ctx, kvPairs); err != nil {
		st := status.New(codes.FailedPrecondition, err.Error())
		return nil, st.Err()
	}

	return &rpc.UpdateResponse{}, nil
}

// Delete removes configuration data present in data request from the VPP/linux
func (svc *configService) Delete(ctx context.Context, req *rpc.DeleteRequest) (*rpc.DeleteResponse, error) {
	protos := extractProtos(req.Delete.Vpp, req.Delete.Linux)

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

func placeProtos(protos map[string]proto.Message, dsts ...interface{}) {
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
