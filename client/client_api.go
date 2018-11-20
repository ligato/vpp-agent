package client

import "github.com/ligato/vpp-agent/api/models"

// SyncClient
type SyncClient interface {
	ResyncRequest() ResyncRequest
	ChangeRequest() ChangeRequest
}

// ResyncRequest
type ResyncRequest interface {
	Update(models ...models.ProtoModel)
	Send() error
}

// ChangeRequest
type ChangeRequest interface {
	Update(models ...models.ProtoModel)
	Delete(keys ...string)
	Send() error
}
