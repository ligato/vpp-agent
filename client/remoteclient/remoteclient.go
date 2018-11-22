package remoteclient

import (
	"context"

	"github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/models"
	"github.com/ligato/vpp-agent/client"
)

type remoteClient struct {
	serviceClient api.SyncServiceClient
}

// NewClientGRPC returns new instance that uses given service client for requests.
func NewClientGRPC(syncSvc api.SyncServiceClient) client.SyncClient {
	return &remoteClient{syncSvc}
}

// ResyncRequest returns new resync request.
func (c *remoteClient) ResyncRequest() client.SyncRequest {
	return &request{
		serviceClient: c.serviceClient,
		req: &api.SyncRequest{
			Options: &api.SyncOptions{Resync: true},
		},
	}
}

// ChangeRequest return new change request.
func (c *remoteClient) ChangeRequest() client.SyncRequest {
	return &request{
		serviceClient: c.serviceClient,
		req: &api.SyncRequest{
			Options: &api.SyncOptions{Resync: false},
		},
	}
}

/*type resyncRequest struct {
	serviceClient api.SyncServiceClient
	req           *api.SyncRequest
}

// Put adds the given model data to the transaction.
func (r *resyncRequest) Put(protoModels ...models.ProtoModel) {
	for _, protoModel := range protoModels {
		model, err := models.Marshal(protoModel)
		if err != nil {
			continue
		}
		r.req.Items = append(r.req.Items, &api.SyncItem{Model: model})
	}
}

// Send commits the transaction with all data.
func (r *resyncRequest) Send() (err error) {
	ctx := context.Background()
	_, err = r.serviceClient.Sync(ctx, r.req)
	return err
}*/

type request struct {
	serviceClient api.SyncServiceClient
	req           *api.SyncRequest
	err           error
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
