package client

import (
	"context"

	"github.com/ligato/vpp-agent/api/models"
)

// SyncClient
type SyncClient interface {
	ResyncRequest() ResyncRequest
	ChangeRequest() ChangeRequest
}

type SyncRequest interface {
	Send(ctx context.Context) error
}

// ResyncRequest
type ResyncRequest interface {
	SyncRequest
	Put(models ...models.ProtoModel)
}

// ChangeRequest
type ChangeRequest interface {
	SyncRequest
	Update(models ...models.ProtoModel)
	Delete(models ...models.ProtoModel)
}
