package client

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/health/probe"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

// APIClient is an interface that clients that talk with a agent server must implement.
type APIClient interface {
	InfraAPIClient
	ModelAPIClient
	SchedulerAPIClient
	VppAPIClient
	MetricsAPIClient

	ConfigClient() (client.ConfigClient, error)

	AgentHost() string
	Version() string
	KVDBClient() (KVDBAPIClient, error)
	GRPCConn() (*grpc.ClientConn, error)
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
	SchedulerValues(ctx context.Context, opts types.SchedulerValuesOptions) ([]*kvscheduler.BaseValueStatus, error)
}

// VppAPIClient defines API client methods for the VPP
type VppAPIClient interface {
	VppRunCli(ctx context.Context, cmd string) (reply string, err error)
}

type MetricsAPIClient interface {
	GetMetricData(ctx context.Context, metricName string) (map[string]interface{}, error)
}

type KVDBAPIClient interface {
	keyval.CoreBrokerWatcher
	ProtoBroker() keyval.ProtoBroker
	CompleteFullKey(key string) (string, error)
}
