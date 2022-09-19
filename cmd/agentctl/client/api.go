package client

import (
	"context"
	"net/http"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/db/keyval"
	"go.ligato.io/cn-infra/v2/health/probe"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

// APIClient is an interface that clients that talk with a agent server must implement.
type APIClient interface {
	InfraAPIClient
	ModelAPIClient
	SchedulerAPIClient
	VppAPIClient
	MetricsAPIClient

	GenericClient() (client.GenericClient, error)
	ConfiguratorClient() (configurator.ConfiguratorServiceClient, error)
	MetaServiceClient() (generic.MetaServiceClient, error)

	AgentHost() string
	Version() string
	KVDBClient() (KVDBAPIClient, error)
	GRPCConn() (*grpc.ClientConn, error)
	HTTPClient() *http.Client
	AgentVersion(ctx context.Context) (*types.Version, error)
	NegotiateAPIVersion(ctx context.Context)
	NegotiateAPIVersionPing(version *types.Version)
	Close() error
}

// InfraAPIClient defines API client methods for the system
type InfraAPIClient interface {
	Status(ctx context.Context) (*probe.ExposedStatus, error)
	LoggerList(ctx context.Context) ([]types.Logger, error)
	LoggerSet(ctx context.Context, logger, level string) error
}

// ModelAPIClient defines API client methods for the models
type ModelAPIClient interface {
	ModelList(ctx context.Context, opts types.ModelListOptions) ([]types.Model, error)
}

// SchedulerAPIClient defines API client methods for the scheduler
type SchedulerAPIClient interface {
	SchedulerDump(ctx context.Context, opts types.SchedulerDumpOptions) ([]api.RecordedKVWithMetadata, error)
	SchedulerValues(ctx context.Context, opts types.SchedulerValuesOptions) ([]*kvscheduler.BaseValueStatus, error)
	SchedulerResync(ctx context.Context, opts types.SchedulerResyncOptions) (*api.RecordedTxn, error)
	SchedulerHistory(ctx context.Context, opts types.SchedulerHistoryOptions) (api.RecordedTxns, error)
}

// VppAPIClient defines API client methods for the VPP
type VppAPIClient interface {
	VppStatsAPIClient
	VppRunCli(ctx context.Context, cmd string) (reply string, err error)
}

// VppStatsAPIClient defines stats API client methods for the VPP
type VppStatsAPIClient interface {
	VppGetStats(ctx context.Context, typ string) error
	VppGetBufferStats() (*govppapi.BufferStats, error)
	VppGetNodeStats() (*govppapi.NodeStats, error)
	VppGetSystemStats() (*govppapi.SystemStats, error)
	VppGetErrorStats() (*govppapi.ErrorStats, error)
	VppGetInterfaceStats() (*govppapi.InterfaceStats, error)
}

type MetricsAPIClient interface {
	GetMetricData(ctx context.Context, metricName string) (map[string]interface{}, error)
}

type KVDBAPIClient interface {
	keyval.CoreBrokerWatcher
	ProtoBroker() keyval.ProtoBroker
	CompleteFullKey(key string) (string, error)
}
