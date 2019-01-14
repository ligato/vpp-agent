package remoteclient

import (
	"context"

	pb "github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/client"
)

type grpcClient struct {
	remote pb.GenericConfiguratorClient
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(client pb.GenericConfiguratorClient) client.ConfiguratorClient {
	return &grpcClient{client}
}

func (c *grpcClient) ListModules() (map[string][]models.Model, error) {
	ctx := context.Background()

	resp, err := c.remote.ListCapabilities(ctx, &pb.ListCapabilitiesRequest{})
	if err != nil {
		return nil, err
	}

	modules := make(map[string][]models.Model)
	for _, model := range resp.ActiveModels {
		modules[model.Module] = append(modules[model.Module], *model)
	}

	return modules, nil
}

func (c *grpcClient) SetConfig(resync bool) client.SetConfigRequest {
	return &setConfigRequest{
		client: c.remote,
		req: &pb.SetConfigRequest{
			Options: &pb.SetConfigRequest_Options{Resync: resync},
		},
	}
}

type setConfigRequest struct {
	client pb.GenericConfiguratorClient
	req    *pb.SetConfigRequest
	err    error
}

func (r *setConfigRequest) Update(items ...models.ProtoItem) {
	if r.err != nil {
		return
	}
	for _, protoModel := range items {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			r.err = err
			return
		}
		r.req.Updates = append(r.req.Updates, &pb.SetConfigRequest_UpdateItem{
			Item: item,
		})
	}
}

func (r *setConfigRequest) Delete(items ...models.ProtoItem) {
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
		r.req.Updates = append(r.req.Updates, &pb.SetConfigRequest_UpdateItem{
			Item: &models.Item{
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
