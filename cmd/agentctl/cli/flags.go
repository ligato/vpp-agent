//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package cli

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/client"
)

const (
	defaultEtcdEndpoint = "127.0.0.1:2379"
)

// ClientOptions define options for the client.
type ClientOptions struct {
	Debug    bool
	LogLevel string
}

// NewClientOptions returns a new ClientOptions
func NewClientOptions() *ClientOptions {
	return &ClientOptions{}
}

// InstallFlags adds flags for the common options on the FlagSet
func (opts *ClientOptions) InstallFlags(flags *pflag.FlagSet) {
	// TODO: consider using viper.AutomaticEnv with some prefix like `AGENTCTL`

	flags.StringP("host", "H", client.DefaultAgentHost, "Address on which agent is reachable, default from AGENT_HOST env var")
	_ = viper.BindPFlag("host", flags.Lookup("host"))
	_ = viper.BindEnv("host", "AGENT_HOST")

	flags.String("service-label", "", "Service label for specific agent instance, default from MICROSERVICE_LABEL env var")
	_ = viper.BindPFlag("service-label", flags.Lookup("service-label"))
	_ = viper.BindEnv("service-label", "MICROSERVICE_LABEL")

	flags.Int("http-port", client.DefaultPortHTTP, "HTTP server port")
	_ = viper.BindPFlag("http-port", flags.Lookup("http-port"))

	flags.Int("grpc-port", client.DefaultPortGRPC, "gRPC server port")
	_ = viper.BindPFlag("grpc-port", flags.Lookup("grpc-port"))

	flags.StringSliceP("etcd-endpoints", "e", []string{defaultEtcdEndpoint}, "Etcd endpoints to connect to, default from ETCD_ENDPOINTS env var")
	_ = viper.BindPFlag("etcd-endpoints", flags.Lookup("etcd-endpoints"))
	_ = viper.BindEnv("etcd-endpoints", "ETCD_ENDPOINTS")

	flags.String("http-basic-auth", "", "Basic auth for HTTP connection in form \"user:pass\"")
	_ = viper.BindPFlag("http-basic-auth", flags.Lookup("http-basic-auth"))
	_ = viper.BindEnv("http-basic-auth", "AGENTCTL_HTTP_BASIC_AUTH")

	flags.Bool("insecure-tls", false, "Use TLS without server's certificate validation")
	_ = viper.BindPFlag("insecure-tls", flags.Lookup("insecure-tls"))

	flags.String("config-dir", "", "Path to directory with config file.")
	_ = viper.BindPFlag("config-dir", flags.Lookup("config-dir"))

	_ = viper.BindEnv("ligato-api-version", "LIGATO_API_VERSION")
}

// SetLogLevel sets the logrus logging level (WarnLevel for empty string).
func SetLogLevel(logLevel string) {
	if logLevel == "" {
		logrus.SetLevel(logrus.WarnLevel)
		logging.DefaultLogger.SetLevel(logging.WarnLevel)
		return
	}

	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse logging level: %s\n", logLevel)
		os.Exit(1)
	}

	logrus.SetLevel(lvl)
	logging.DefaultLogger.SetLevel(logging.ParseLogLevel(logLevel))
}
