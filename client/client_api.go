package client

import (
	"context"

	"github.com/ligato/vpp-agent/api/models"
)

// SyncClient
type SyncClient interface {
	//SyncRequest(ctx context.Context) SyncRequest
	ResyncRequest() SyncRequest
	ChangeRequest() SyncRequest
}

type SyncRequest interface {
	Send(ctx context.Context) error

	Update(models ...models.ProtoModel)
	Delete(models ...models.ProtoModel)
}

/*
// ResyncRequest
type ResyncRequest interface {
	Request
	Put(models ...models.ProtoModel)
}

// ChangeRequest
type ChangeRequest interface {
	TxnRequest
	Update(models ...models.ProtoModel)
	Delete(models ...models.ProtoModel)
}
*/
/*
// ChangeRequest
type ChangeRequest interface {
}
*/
