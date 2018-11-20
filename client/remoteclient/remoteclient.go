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
func (c *remoteClient) ResyncRequest() client.ResyncRequest {
	return &request{c.serviceClient, &api.SyncRequest{}, true}
}

// ChangeRequest return new change request.
func (c *remoteClient) ChangeRequest() client.ChangeRequest {
	return &request{c.serviceClient, &api.SyncRequest{}, false}
}

type request struct {
	serviceClient api.SyncServiceClient
	req           *api.SyncRequest
	isResync      bool
}

// Update adds update for the given model data to the transaction.
func (r *request) Update(protoModels ...models.ProtoModel) {
	for _, protoModel := range protoModels {
		model, err := models.Marshal(protoModel)
		if err != nil {
			continue
		}
		r.req.Models = append(r.req.Models, model)
	}
}

// Delete adds delete for the given model keys to the transaction.
func (r *request) Delete(keys ...string) {
	for _, key := range keys {
		r.req.Models = append(r.req.Models, &models.Model{
			Key:   key,
			Value: nil, // nil value represents delete
		})
	}
}

// Send commits the transaction with all data.
func (r *request) Send() (err error) {
	ctx := context.Background()

	if r.isResync {
		_, err = r.serviceClient.Resync(ctx, r.req)
	} else {
		_, err = r.serviceClient.Change(ctx, r.req)
	}

	return err
}
