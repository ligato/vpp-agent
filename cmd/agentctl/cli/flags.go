package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v2/cmd/agentctl/client"
)

const (
	defaultEtcdEndpoint = "127.0.0.1:2379"
)

var (
	etcdEndpoints = strings.Split(os.Getenv("ETCD_ENDPOINTS"), ",")
)

func init() {
	if len(etcdEndpoints) == 0 || etcdEndpoints[0] == "" {
		etcdEndpoints = []string{defaultEtcdEndpoint}
	}
}

// ClientOptions define options for the client.
type ClientOptions struct {
	Debug    bool
	LogLevel string

	PortGRPC  int
	PortHTTP  int
	Endpoints []string
}

// NewClientOptions returns a new ClientOptions
func NewClientOptions() *ClientOptions {
	return &ClientOptions{}
}

// InstallFlags adds flags for the common options on the FlagSet
func (opts *ClientOptions) InstallFlags(flags *pflag.FlagSet) {
	// TODO: consider using viper.AutomaticEnv with some prefix like `AGENTCTL`

	flags.StringP("host", "H", client.DefaultAgentHost, "Address on which agent is reachable, default from AGENT_HOST env var")
	viper.BindPFlag("host", flags.Lookup("host"))
	viper.BindEnv("host", "AGENT_HOST")

	flags.String("service-label", "", "Service label for specific agent instance, default from MICROSERVICE_LABEL env var")
	viper.BindPFlag("service-label", flags.Lookup("service-label"))
	viper.BindEnv("service-label", "MICROSERVICE_LABEL")

	flags.IntVar(&opts.PortHTTP, "http-port", client.DefaultPortHTTP, "HTTP server port")
	flags.IntVar(&opts.PortGRPC, "grpc-port", client.DefaultPortGRPC, "gRPC server port")
	flags.StringSliceVarP(&opts.Endpoints, "etcd-endpoints", "e", etcdEndpoints, "Etcd endpoints to connect to, default from ETCD_ENDPOINTS env var")

	flags.Bool("tls", false, "Use TLS for connections")
	viper.BindPFlag("use-tls", flags.Lookup("tls"))

	flags.String("config-dir", DefaultConfigDir(), "Path to directory with config file")
	viper.BindPFlag("config-dir", flags.Lookup("config-dir"))
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
