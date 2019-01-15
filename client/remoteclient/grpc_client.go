package remoteclient

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/client"
)

type grpcClient struct {
	remote api.GenericConfiguratorClient
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(client api.GenericConfiguratorClient) client.ConfigClient {
	return &grpcClient{client}
}

func (c *grpcClient) ActiveModels() (map[string][]api.Model, error) {
	ctx := context.Background()

	resp, err := c.remote.Capabilities(ctx, &api.CapabilitiesRequest{})
	if err != nil {
		return nil, err
	}

	modules := make(map[string][]api.Model)
	for _, model := range resp.ActiveModels {
		modules[model.Module] = append(modules[model.Module], *model)
	}

	return modules, nil
}

func (c *grpcClient) SetConfig(resync bool) client.SetConfigRequest {
	return &setConfigRequest{
		client: c.remote,
		req: &api.SetConfigRequest{
			Options: &api.SetConfigRequest_Options{Resync: resync},
		},
	}
}

type setConfigRequest struct {
	client api.GenericConfiguratorClient
	req    *api.SetConfigRequest
	err    error
}

func (r *setConfigRequest) Update(items ...proto.Message) {
	if r.err != nil {
		return
	}
	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			r.err = err
			return
		}
		r.req.Updates = append(r.req.Updates, &api.SetConfigRequest_UpdateItem{
			Item: item,
		})
	}
}

func (r *setConfigRequest) Delete(items ...proto.Message) {
	if r.err != nil {
		return
	}
	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			if err != nil {
				r.err = err
				return
			}
		}
		r.req.Updates = append(r.req.Updates, &api.SetConfigRequest_UpdateItem{
			Item: &api.Item{
				Key: item.Key,
			},
		})
	}
}

func (r *setConfigRequest) Send(ctx context.Context) (err error) {
	if r.err != nil {
		return r.err
	}
	_, err = r.client.SetConfig(ctx, r.req)
	return err
}
