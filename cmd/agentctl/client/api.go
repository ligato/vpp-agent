package client

import (
	"context"
	"net/http"

	"github.com/ligato/cn-infra/db/keyval"

	"github.com/ligato/vpp-agent/api/types"
	"github.com/ligato/vpp-agent/client"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"

	"google.golang.org/grpc"
)

// APIClient is an interface that clients that talk with a agent server must implement.
type APIClient interface {
	InfraAPIClient
	ModelAPIClient
	SchedulerAPIClient
	VppAPIClient

	ConfigClient() (client.ConfigClient, error)

	AgentHost() string
	ClientVersion() string
	KVDBClient() (keyval.BytesBroker, error)
	GRPCConn() (*grpc.ClientConn, error)
	HTTPClient() *http.Client
	ServerVersion(ctx context.Context) (types.Version, error)
	NegotiateAPIVersion(ctx context.Context)
	NegotiateAPIVersionPing(types.Ping)
	Close() error
}

// SystemAPIClient defines API client methods for the system
type InfraAPIClient interface {
	Ping(ctx context.Context) (types.Ping, error)
	LoggerList(ctx context.Context, opts types.LoggerListOptions) ([]types.Logger, error)
	LoggerSet(ctx context.Context, logger, level string) error
}

// ModelAPIClient defines API client methods for the models
type ModelAPIClient interface {
	ModelList(ctx context.Context, opts types.ModelListOptions) ([]types.Model, error)
}

// SchedulerAPIClient defines API client methods for the scheduler
type SchedulerAPIClient interface {
	SchedulerDump(ctx context.Context, opts types.SchedulerDumpOptions) ([]api.KVWithMetadata, error)
	SchedulerStatus(ctx context.Context, opts types.SchedulerStatusOptions) ([]*api.BaseValueStatus, error)
}

// VppAPIClient
type VppAPIClient interface {
	VppRunCli(ctx context.Context, cmd string) (reply string, err error)
}
