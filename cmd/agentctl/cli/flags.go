package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/go-connections/tlsconfig"
	"github.com/ligato/cn-infra/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

const (
	defaultAgentHost    = "127.0.0.1"
	defaultPortGRPC     = "9111"
	defaultPortHTTP     = "9191"
	defaultEtcdEndpoint = "127.0.0.1:2379"
)

var (
	serviceLabel  = os.Getenv("MICROSERVICE_LABEL")
	agentHost     = os.Getenv("AGENT_HOST")
	etcdEndpoints = strings.Split(os.Getenv("ETCD_ENDPOINTS"), ",")
)

func init() {
	if agentHost == "" {
		agentHost = defaultAgentHost
	}
	if len(etcdEndpoints) == 0 || etcdEndpoints[0] == "" {
		etcdEndpoints = []string{defaultEtcdEndpoint}
	}
}

// ClientOptions define options for the client.
type ClientOptions struct {
	Debug    bool
	LogLevel string

	AgentHost    string
	PortGRPC     string
	PortHTTP     string
	ServiceLabel string
	Endpoints    []string

	// TODO: support TLS
	TLS        bool
	TLSVerify  bool
	TLSOptions *tlsconfig.Options
}

// NewClientOptions returns a new ClientOptions
func NewClientOptions() *ClientOptions {
	return &ClientOptions{}
}

// InstallFlags adds flags for the common options on the FlagSet
func (opts *ClientOptions) InstallFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&opts.Debug, "debug", "D", false, "Enable debug mode")
	flags.StringVarP(&opts.LogLevel, "log-level", "l", "", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)

	flags.StringVarP(&opts.AgentHost, "host", "H", agentHost, "Address on which agent is reachable, default from AGENT_HOST env var")
	flags.StringVar(&opts.PortGRPC, "grpc-port", defaultPortGRPC, "gRPC server port")
	flags.StringVar(&opts.PortHTTP, "http-port", defaultPortHTTP, "HTTP server port")
	flags.StringVar(&opts.ServiceLabel, "service-label", serviceLabel, "Service label for specific agent instance, default from MICROSERVICE_LABEL env var")
	flags.StringSliceVarP(&opts.Endpoints, "etcd-endpoints", "e", etcdEndpoints, "Etcd endpoints to connect to, default from ETCD_ENDPOINTS env var")
}

// SetDefaultOptions sets default values for options after flag parsing is
// complete
func (opts *ClientOptions) SetDefaultOptions(flags *pflag.FlagSet) {
	// no-op
}

// SetLogLevel sets the logrus logging level
func SetLogLevel(logLevel string) {
	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse logging level: %s\n", logLevel)
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
		logging.DefaultLogger.SetLevel(logging.ParseLogLevel(logLevel))
	} else {
		logrus.SetLevel(logrus.WarnLevel)
		logging.DefaultLogger.SetLevel(logging.WarnLevel)
	}
}
