package remoteclient

import (
	"context"

	"github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/client"
)

type grpcClient struct {
	serviceClient api.SyncServiceClient
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(syncSvc api.SyncServiceClient) client.SyncClient {
	return &grpcClient{syncSvc}
}

func (c *grpcClient) ListSpecs() ([]models.Spec, error) {
	ctx := context.Background()

	resp, err := c.serviceClient.ListSpecs(ctx, &api.ListSpecsRequest{})
	if err != nil {
		return nil, err
	}

	var specs []models.Spec
	for _, spec := range resp.Specs {
		specs = append(specs, models.Spec{
			Version: spec.Version,
			Class:   spec.Class,
			Module:  spec.Module,
			Kind:    spec.Kind,
		})
	}

	return specs, nil
}

// ResyncRequest returns new resync request.
func (c *grpcClient) ResyncRequest() client.ResyncRequest {
	return &request{
		serviceClient: c.serviceClient,
		req: &api.SyncRequest{
			Options: &api.SyncOptions{Resync: true},
		},
	}
}

// ChangeRequest return new change request.
func (c *grpcClient) ChangeRequest() client.ChangeRequest {
	return &request{
		serviceClient: c.serviceClient,
		req: &api.SyncRequest{
			Options: &api.SyncOptions{Resync: false},
		},
	}
}

type request struct {
	serviceClient api.SyncServiceClient
	req           *api.SyncRequest
	err           error
}

// Put puts the given model data to the transaction.
func (r *request) Put(protoModels ...models.ProtoModel) {
	r.Update(protoModels...)
}

// Update adds update for the given model data to the transaction.
func (r *request) Update(protoModels ...models.ProtoModel) {
	if r.err != nil {
		return
	}
	for _, protoModel := range protoModels {
		model, err := models.Marshal(protoModel)
		if err != nil {
			r.err = err
			return
		}
		r.req.Items = append(r.req.Items, &api.SyncItem{Model: model})
	}
}

// Delete adds delete for the given model keys to the transaction.
func (r *request) Delete(protoModels ...models.ProtoModel) {
	if r.err != nil {
		return
	}
	for _, protoModel := range protoModels {
		model, err := models.Marshal(protoModel)
		if err != nil {
			if err != nil {
				r.err = err
				return
			}
		}
		r.req.Items = append(r.req.Items, &api.SyncItem{
			Model:  model,
			Delete: true,
		})
	}
}

// Send commits the transaction with all data.
func (r *request) Send(ctx context.Context) (err error) {
	if r.err != nil {
		return r.err
	}
	_, err = r.serviceClient.Sync(ctx, r.req)
	return err
}
