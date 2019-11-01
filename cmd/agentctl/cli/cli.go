//  Copyright (c) 2019 Cisco and/or its affiliates.
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
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/pkg/term"
	"github.com/spf13/viper"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v2/cmd/agentctl/api"
	"go.ligato.io/vpp-agent/v2/cmd/agentctl/client"
	"go.ligato.io/vpp-agent/v2/pkg/debug"
)

// Cli represents the agent command line client.
type Cli interface {
	Client() client.APIClient
	KVProtoBroker() (keyval.ProtoBroker, error)

	Out() *streams.Out
	Err() io.Writer
	In() *streams.In
	SetIn(in *streams.In)
	Apply(ops ...AgentCliOption) error
	ServerInfo() ServerInfo
	ClientInfo() ClientInfo
	DefaultVersion() string
}

type AgentCli struct {
	client     client.APIClient
	in         *streams.In
	out        *streams.Out
	err        io.Writer
	serverInfo ServerInfo
	clientInfo ClientInfo
}

// NewAgentCli returns a AgentCli instance with all operators applied on it.
// It applies by default the standard streams.
func NewAgentCli(ops ...AgentCliOption) (*AgentCli, error) {
	cli := new(AgentCli)
	var defaultOps []AgentCliOption
	ops = append(defaultOps, ops...)
	if err := cli.Apply(ops...); err != nil {
		return nil, err
	}
	if cli.out == nil || cli.in == nil || cli.err == nil {
		stdin, stdout, stderr := term.StdStreams()
		if cli.in == nil {
			cli.in = streams.NewIn(stdin)
		}
		if cli.out == nil {
			cli.out = streams.NewOut(stdout)
		}
		if cli.err == nil {
			cli.err = stderr
		}
	}
	return cli, nil
}

// Client returns the APIClient
func (cli *AgentCli) Client() client.APIClient {
	return cli.client
}

// Apply all the operation on the cli
func (cli *AgentCli) Apply(ops ...AgentCliOption) error {
	for _, op := range ops {
		if err := op(cli); err != nil {
			return err
		}
	}
	return nil
}

func (cli *AgentCli) Out() *streams.Out {
	return cli.out
}

func (cli *AgentCli) Err() io.Writer {
	return cli.err
}

func (cli *AgentCli) In() *streams.In {
	return cli.in
}

func (cli *AgentCli) SetIn(in *streams.In) {
	cli.in = in
}

func (cli *AgentCli) ServerInfo() ServerInfo {
	return cli.serverInfo
}

func (cli *AgentCli) ClientInfo() ClientInfo {
	return cli.clientInfo
}

func (cli *AgentCli) DefaultVersion() string {
	return cli.clientInfo.DefaultVersion
}

func (cli *AgentCli) KVProtoBroker() (keyval.ProtoBroker, error) {
	kvdb, err := cli.Client().KVDBClient()
	if err != nil {
		return nil, fmt.Errorf("connecting to KVBDB failed: %v", err)
	}
	return jsonProtoBroker(kvdb), nil
}

func jsonProtoBroker(broker keyval.CoreBrokerWatcher) keyval.ProtoBroker {
	return kvproto.NewProtoWrapper(broker, &keyval.SerializerJSON{})
}

// ServerInfo stores details about the supported features and platform of the
// server
type ServerInfo struct {
	OSType string
}

// ClientInfo stores details about the supported features of the client
type ClientInfo struct {
	DefaultVersion string
}

// UserAgent returns the user agent string used for making API requests
func UserAgent() string {
	return "Ligato-Client/" + api.DefaultVersion + " (" + runtime.GOOS + ")"
}

// InitializeOpt is the type of the functional options passed to AgentCli.Initialize
type InitializeOpt func(agentCli *AgentCli) error

// Initialize the agentCli runs initialization that must happen after command
// line flags are parsed.
func (cli *AgentCli) Initialize(opts *ClientOptions, ops ...InitializeOpt) error {
	var err error
	for _, o := range ops {
		if err := o(cli); err != nil {
			return err
		}
	}

	if opts.Debug {
		debug.Enable()
		SetLogLevel("debug")
	} else {
		SetLogLevel(opts.LogLevel)
	}

	ReadConfig() // TODO: maybe move it elsewhere

	if cli.client == nil {
		clientOptions := buildClientOptions()
		cli.client, err = client.NewClientWithOpts(clientOptions...)
		if err != nil {
			return err
		}
	}
	cli.clientInfo = ClientInfo{
		DefaultVersion: cli.client.Version(),
	}
	cli.initializeFromClient()
	return nil
}

func buildClientOptions() []client.Opt {
	logging.Debug("----------------------------------------------------")
	logging.Debug("Building client options")
	logging.Debugf("\tHost: %q\n", viper.GetString("host"))
	logging.Debugf("\tService Label: %q\n", viper.GetString("service-label"))
	logging.Debugf("\tGRPC Port: %q\n", viper.GetString("grpc-port"))
	logging.Debugf("\tHTTP Port: %q\n", viper.GetString("http-port"))
	logging.Debugf("\tETCD endpoints: %#v\n", viper.GetStringSlice("etcd-endpoints"))
	logging.Debugf("\tLIGATO_API_VERSION env var: %q\n", viper.GetString("LIGATO_API_VERSION"))
	logging.Debugf("\tUse TLS?: %t\n", viper.GetBool("use-tls"))
	logging.Debugf("\tGRPC TLS: %#v\n", viper.GetStringMap("grpc-tls"))
	logging.Debugf("\tHTTP TLS: %#v\n", viper.GetStringMap("http-tls"))
	logging.Debugf("\tKVDB TLS: %#v\n", viper.GetStringMap("kvdb-tls"))
	logging.Debug("----------------------------------------------------")

	clientOpts := []client.Opt{
		client.WithHost(viper.GetString("host")),
		client.WithServiceLabel(viper.GetString("service-label")),
		client.WithGrpcPort(viper.GetInt("grpc-port")),
		client.WithHTTPPort(viper.GetInt("http-port")),
		client.WithVersion(viper.GetString("LIGATO_API_VERSION")),
	}

	// Handle properly case when `etcd-endpoints` returned from environment variable
	etcdEndp := viper.GetStringSlice("etcd-endpoints")
	if len(etcdEndp) == 1 && strings.Contains(etcdEndp[0], ",") {
		etcdEndp = strings.Split(etcdEndp[0], ",")
	}
	clientOpts = append(clientOpts, client.WithEtcdEndpoints(etcdEndp))

	var customHeaders = map[string]string{
		"User-Agent": UserAgent(),
	}
	basicAuth := viper.GetString("basic-auth")
	if basicAuth != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(basicAuth))
		customHeaders["Authorization"] = "Basic " + auth
	}
	clientOpts = append(clientOpts, client.WithHTTPHeaders(customHeaders))

	if viper.GetBool("use-tls") {
		if viper.InConfig("grpc-tls") && !viper.GetBool("grpc-tls.disabled") {
			clientOpts = append(clientOpts, client.WithGrpcTLS(
				viper.GetString("grpc-tls.cert-file"),
				viper.GetString("grpc-tls.key-file"),
				viper.GetString("grpc-tls.ca-file"),
				viper.GetBool("grpc-tls.skip-verify"),
			))
		}

		if viper.InConfig("http-tls") && !viper.GetBool("http-tls.disabled") {
			clientOpts = append(clientOpts, client.WithHTTPTLS(
				viper.GetString("http-tls.cert-file"),
				viper.GetString("http-tls.key-file"),
				viper.GetString("http-tls.ca-file"),
				viper.GetBool("http-tls.skip-verify"),
			))
		}

		if viper.InConfig("kvdb-tls") && !viper.GetBool("kvdb-tls.disabled") {
			clientOpts = append(clientOpts, client.WithKvdbTLS(
				viper.GetString("kvdb-tls.cert-file"),
				viper.GetString("kvdb-tls.key-file"),
				viper.GetString("kvdb-tls.ca-file"),
				viper.GetBool("kvdb-tls.skip-verify"),
			))
		}
	}

	return clientOpts
}

func (cli *AgentCli) initializeFromClient() {
	logging.Debugf("initializeFromClient (DefaultVersion: %v)", cli.DefaultVersion())

	ping, err := cli.client.Ping(context.Background())
	if err != nil {
		// Default to true if we fail to connect to daemon
		cli.serverInfo = ServerInfo{}

		if ping.APIVersion != "" {
			cli.client.NegotiateAPIVersionPing(ping)
		}
		return
	}

	cli.serverInfo = ServerInfo{
		OSType: ping.OSType,
	}
	cli.client.NegotiateAPIVersionPing(ping)
}
