package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/tlsconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/cmd/agentctl/client"
)

const (
	defaultEtcdEndpoint = "127.0.0.1:2379"
)

var (
	serviceLabel  = os.Getenv("MICROSERVICE_LABEL")
	agentHost     = os.Getenv("AGENT_HOST")
	etcdEndpoints = strings.Split(os.Getenv("ETCD_ENDPOINTS"), ",")
)

func init() {
	if agentHost == "" {
		agentHost = client.DefaultAgentHost
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
	PortGRPC     int
	PortHTTP     int
	ServiceLabel string
	Endpoints    []string

	ConfigDir string

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
	flags.StringVarP(&opts.AgentHost, "host", "H", agentHost, "Address on which agent is reachable, default from AGENT_HOST env var")
	flags.StringVar(&opts.ServiceLabel, "service-label", serviceLabel, "Service label for specific agent instance, default from MICROSERVICE_LABEL env var")
	flags.IntVar(&opts.PortHTTP, "http-port", client.DefaultPortHTTP, "HTTP server port")
	flags.IntVar(&opts.PortGRPC, "grpc-port", client.DefaultPortGRPC, "gRPC server port")
	flags.StringSliceVarP(&opts.Endpoints, "etcd-endpoints", "e", etcdEndpoints, "Etcd endpoints to connect to, default from ETCD_ENDPOINTS env var")

	// initialize default path for config file directory
	// 🔧 Work in progress. This will be refactored later 🔧
	uhd, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get current user's home directory: %v\n", err)
		os.Exit(1)
	}
	// `.agentctl` is a default name of directory with config file
	configDir := filepath.Join(uhd, ".agentctl")

	flags.StringVar(&opts.ConfigDir, "config", configDir, "Location of client config files")
}

// SetDefaultOptions sets default values for options after flag parsing is
// complete
func (opts *ClientOptions) SetDefaultOptions(flags *pflag.FlagSet) {
	client.DefaultPortHTTP = opts.PortHTTP
	client.DefaultPortGRPC = opts.PortGRPC
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
