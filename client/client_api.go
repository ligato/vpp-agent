package client

import (
	"context"

	"github.com/ligato/vpp-agent/api/models"
)

// ConfiguratorClient defines the client-side interface for sync service.
type ConfiguratorClient interface {
	// ListModules retrieves list of supported modules.
	ListModules() (map[string][]models.Model, error)

	// SetConfig
	SetConfig(resync bool) SetConfigRequest
}

// SetConfigRequest defines request interface for setting config.
type SetConfigRequest interface {
	// Update appends updates for given items to the request.
	Update(items ...models.ProtoItem)

	// Delete appends deletes for given items to the request.
	Delete(items ...models.ProtoItem)

	// Send sends the request.
	Send(ctx context.Context) error
}
