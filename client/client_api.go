package client

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/api"
)

// ConfigClient defines the client-side interface for config service.
type ConfigClient interface {
	// ActiveModels retrieves list of active modules.
	ActiveModels() (map[string][]api.Model, error)

	SetConfig(resync bool) SetConfigRequest

	GetConfig(dsts ...interface{}) error
}

// SetConfigRequest defines interface for config set request.
type SetConfigRequest interface {
	// Update appends updates for given items to the request.
	Update(items ...proto.Message)

	// Delete appends deletes for given items to the request.
	Delete(items ...proto.Message)

	// Send sends the request.
	Send(ctx context.Context) error
}
