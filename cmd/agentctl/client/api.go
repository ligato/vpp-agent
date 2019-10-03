package client

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/health/probe"

	"github.com/ligato/vpp-agent/api/types"
	"github.com/ligato/vpp-agent/client"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
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
	KVDBClient() (KVDBAPIClient, error)
	GRPCConn() (*grpc.ClientConn, error)
	GRPCAddr() (string, error)
	HTTPClient() *http.Client
	ServerVersion(ctx context.Context) (types.Version, error)
	NegotiateAPIVersion(ctx context.Context)
	NegotiateAPIVersionPing(types.Ping)
	Close() error
}

// SystemAPIClient defines API client methods for the system
type InfraAPIClient interface {
	Status(ctx context.Context) (*probe.ExposedStatus, error)
	Ping(ctx context.Context) (types.Ping, error)
	LoggerList(ctx context.Context) ([]types.Logger, error)
	LoggerSet(ctx context.Context, logger, level string) error
}

// ModelAPIClient defines API client methods for the models
type ModelAPIClient interface {
	ModelList(ctx context.Context, opts types.ModelListOptions) ([]types.Model, error)
}

// SchedulerAPIClient defines API client methods for the scheduler
type SchedulerAPIClient interface {
	SchedulerDump(ctx context.Context, opts types.SchedulerDumpOptions) ([]api.KVWithMetadata, error)
	SchedulerValues(ctx context.Context, opts types.SchedulerValuesOptions) ([]*api.BaseValueStatus, error)
}

// VppAPIClient defines API client methods for the VPP
type VppAPIClient interface {
	VppRunCli(ctx context.Context, cmd string) (reply string, err error)
}

type KVDBAPIClient interface {
	keyval.CoreBrokerWatcher
}
