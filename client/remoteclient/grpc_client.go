package remoteclient

import (
	"context"

	pb "github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/client"
)

type grpcClient struct {
	remote pb.ConfiguratorClient
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(syncSvc pb.ConfiguratorClient) client.SyncClient {
	return &grpcClient{syncSvc}
}

// ListCapabilities retrieves supported capabilities.
func (c *grpcClient) ListCapabilities() (map[string][]models.Model, error) {
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

// ResyncRequest returns new resync request.
func (c *grpcClient) ResyncRequest() client.ResyncRequest {
	return &request{
		serviceClient: c.remote,
		req: &pb.SetConfigRequest{
			Options: &pb.SetConfigRequest_Options{Resync: true},
		},
	}
}

// ChangeRequest return new change request.
func (c *grpcClient) ChangeRequest() client.ChangeRequest {
	return &request{
		serviceClient: c.remote,
		req: &pb.SetConfigRequest{
			Options: &pb.SetConfigRequest_Options{Resync: false},
		},
	}
}

type request struct {
	serviceClient pb.ConfiguratorClient
	req           *pb.SetConfigRequest
	err           error
}

// Put puts the given model data to the transaction.
func (r *request) Put(protoModels ...models.ProtoItem) {
	r.Update(protoModels...)
}

// Update adds update for the given model data to the transaction.
func (r *request) Update(protoModels ...models.ProtoItem) {
	if r.err != nil {
		return
	}
	for _, protoModel := range protoModels {
		model, err := models.MarshalItem(protoModel)
		if err != nil {
			r.err = err
			return
		}
		r.req.Update = append(r.req.Update, &pb.SetConfigRequest_UpdateItem{
			Item: model,
		})
	}
}

// Delete adds delete for the given model keys to the transaction.
func (r *request) Delete(protoModels ...models.ProtoItem) {
	if r.err != nil {
		return
	}
	for _, protoModel := range protoModels {
		item, err := models.MarshalItem(protoModel)
		if err != nil {
			if err != nil {
				r.err = err
				return
			}
		}
		r.req.Update = append(r.req.Update, &pb.SetConfigRequest_UpdateItem{
			Item: item,
		})
	}
}

// Send commits the transaction with all data.
func (r *request) Send(ctx context.Context) (err error) {
	if r.err != nil {
		return r.err
	}
	_, err = r.serviceClient.SetConfig(ctx, r.req)
	return err
}
