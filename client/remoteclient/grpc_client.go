package remoteclient

import (
	"context"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/util"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

type grpcClient struct {
	manager generic.ManagerServiceClient
	meta    generic.MetaServiceClient
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(conn *grpc.ClientConn) client.ConfigClient {
	manager := generic.NewManagerServiceClient(conn)
	meta := generic.NewMetaServiceClient(conn)
	return &grpcClient{
		manager: manager,
		meta:    meta,
	}
}

func (c *grpcClient) KnownModels(class string) ([]*client.ModelInfo, error) {
	ctx := context.Background()

	resp, err := c.meta.KnownModels(ctx, &generic.KnownModelsRequest{
		Class: class,
	})
	if err != nil {
		return nil, err
	}

	var modules []*client.ModelInfo
	for _, info := range resp.KnownModels {
		modules = append(modules, info)
	}

	return modules, nil
}

func (c *grpcClient) ChangeRequest() client.ChangeRequest {
	return &setConfigRequest{
		client: c.manager,
		req:    &generic.SetConfigRequest{},
	}
}

func (c *grpcClient) ResyncConfig(items ...proto.Message) error {
	req := &generic.SetConfigRequest{
		OverwriteAll: true,
	}

	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			return err
		}
		req.Updates = append(req.Updates, &generic.UpdateItem{
			Item: item,
		})
	}

	_, err := c.manager.SetConfig(context.Background(), req)
	return err
}

func (c *grpcClient) GetConfig(dsts ...interface{}) error {
	ctx := context.Background()

	resp, err := c.manager.GetConfig(ctx, &generic.GetConfigRequest{})
	if err != nil {
		return err
	}

	protos := map[string]proto.Message{}
	for _, item := range resp.Items {
		val, err := models.UnmarshalItem(item.Item)
		if err != nil {
			return err
		}
		var key string
		if data := item.Item.GetData(); data != nil {
			key, err = models.GetKey(val)
		} else {
			key, err = models.GetKeyForItem(item.Item)
		}
		if err != nil {
			return err
		}
		protos[key] = val
	}

	util.PlaceProtos(protos, dsts...)

	return nil
}

func (c *grpcClient) DumpState() ([]*client.StateItem, error) {
	ctx := context.Background()

	resp, err := c.manager.DumpState(ctx, &generic.DumpStateRequest{})
	if err != nil {
		return nil, err
	}

	return resp.GetItems(), nil
}

type setConfigRequest struct {
	client generic.ManagerServiceClient
	req    *generic.SetConfigRequest
	err    error
}

func (r *setConfigRequest) Update(items ...proto.Message) client.ChangeRequest {
	if r.err != nil {
		return r
	}
	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			r.err = err
			return r
		}
		r.req.Updates = append(r.req.Updates, &generic.UpdateItem{
			Item: item,
		})
	}
	return r
}

func (r *setConfigRequest) Delete(items ...proto.Message) client.ChangeRequest {
	if r.err != nil {
		return r
	}
	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			r.err = err
			return r
		}
		item.Data = nil // delete
		r.req.Updates = append(r.req.Updates, &generic.UpdateItem{
			Item: item,
		})
	}
	return r
}

func (r *setConfigRequest) Send(ctx context.Context) (err error) {
	if r.err != nil {
		return r.err
	}
	_, err = r.client.SetConfig(ctx, r.req)
	return err
}
