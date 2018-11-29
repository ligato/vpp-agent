package remoteclient

import (
	"context"

	pb "github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/client"
)

type grpcClient struct {
	remote pb.SyncServiceClient
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(syncSvc pb.SyncServiceClient) client.SyncClient {
	return &grpcClient{syncSvc}
}

// ListModules lists all available modules and their model specs.
func (c *grpcClient) ListModules() (modules map[string]models.Module, err error) {
	ctx := context.Background()

	resp, err := c.remote.ListModules(ctx, &pb.ListModulesRequest{})
	if err != nil {
		return nil, err
	}

	modules = make(map[string]models.Module)
	for _, module := range resp.Modules {
		modules[module.Name] = models.Module{
			Name:  module.Name,
			Specs: module.Specs,
		}
	}

	return modules, nil
}

// ResyncRequest returns new resync request.
func (c *grpcClient) ResyncRequest() client.ResyncRequest {
	return &request{
		serviceClient: c.remote,
		req: &pb.SyncRequest{
			Options: &pb.SyncOptions{Resync: true},
		},
	}
}

// ChangeRequest return new change request.
func (c *grpcClient) ChangeRequest() client.ChangeRequest {
	return &request{
		serviceClient: c.remote,
		req: &pb.SyncRequest{
			Options: &pb.SyncOptions{Resync: false},
		},
	}
}

type request struct {
	serviceClient pb.SyncServiceClient
	req           *pb.SyncRequest
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
		r.req.Items = append(r.req.Items, &pb.SyncItem{Model: model})
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
		r.req.Items = append(r.req.Items, &pb.SyncItem{
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
