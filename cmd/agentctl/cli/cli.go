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
	"io"
	"runtime"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/pkg/term"

	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/client"
	"go.ligato.io/vpp-agent/v3/pkg/debug"
)

// Cli represents the agent command line client.
type Cli interface {
	Client() client.APIClient

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

	cfg, err := MakeConfig()
	if err != nil {
		return err
	}
	if opts.Debug {
		logging.Debug(cfg.DebugOutput())
	}

	if cli.client == nil {
		clientOptions := buildClientOptions(cfg)
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

func buildClientOptions(cfg *Config) []client.Opt {
	clientOpts := []client.Opt{
		client.WithHost(cfg.Host),
		client.WithServiceLabel(cfg.ServiceLabel),
		client.WithGrpcPort(cfg.GRPCPort),
		client.WithHTTPPort(cfg.HTTPPort),
		client.WithHTTPHeader("User-Agent", UserAgent()),
		client.WithHTTPBasicAuth(cfg.HTTPBasicAuth),
		client.WithVersion(cfg.LigatoAPIVersion),
		client.WithEtcdEndpoints(cfg.EtcdEndpoints),
		client.WithEtcdDialTimeout(cfg.EtcdDialTimeout),
	}

	if cfg.ShouldUseSecureGRPC() {
		clientOpts = append(clientOpts, client.WithGrpcTLS(
			cfg.GRPCSecure.CertFile,
			cfg.GRPCSecure.KeyFile,
			cfg.GRPCSecure.CAFile,
			cfg.GRPCSecure.SkipVerify,
		))
	}
	if cfg.ShouldUseSecureHTTP() {
		clientOpts = append(clientOpts, client.WithHTTPTLS(
			cfg.HTTPSecure.CertFile,
			cfg.HTTPSecure.KeyFile,
			cfg.HTTPSecure.CAFile,
			cfg.HTTPSecure.SkipVerify,
		))
	}
	if cfg.ShouldUseSecureKVDB() {
		clientOpts = append(clientOpts, client.WithKvdbTLS(
			cfg.KVDBSecure.CertFile,
			cfg.KVDBSecure.KeyFile,
			cfg.KVDBSecure.CAFile,
			cfg.KVDBSecure.SkipVerify,
		))
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
