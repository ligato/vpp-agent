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

package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/versions"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.fd.io/govpp/proxy"
	"go.ligato.io/cn-infra/v2/db/keyval"
	"go.ligato.io/cn-infra/v2/db/keyval/etcd"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/client/remoteclient"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/pkg/debug"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

const (
	// DefaultAgentHost defines default host address for agent.
	DefaultAgentHost = "127.0.0.1"
	// DefaultPortGRPC defines default port for GRPC connection.
	DefaultPortGRPC = 9111
	// DefaultPortHTTP defines default port for HTTP connection.
	DefaultPortHTTP = 9191
)

// Constants for etcd connection.
const (
	// DefaultTimeout defines default timeout for client HTTP requests.
	DefaultTimeout = time.Second * 120
	// DefaultEtcdDialTimeout defines default timeout for dialing Etcd.
	DefaultEtcdDialTimeout = time.Second * 3
	// DefaultEtcdOpTimeout defines default timeout for a pending Etcd operation.
	DefaultEtcdOpTimeout = time.Second * 10
)

var _ APIClient = (*Client)(nil)

// Client is the API client that performs all operations
// against a Ligato agent.
type Client struct {
	scheme   string
	host     string
	proto    string
	addr     string
	basePath string

	grpcPort        int
	grpcAddr        string
	grpcTLS         *tls.Config
	httpPort        int
	httpAddr        string
	httpTLS         *tls.Config
	kvdbEndpoints   []string
	kvdbDialTimeout time.Duration
	kvdbTLS         *tls.Config
	serviceLabel    string

	grpcClient       *grpc.ClientConn
	httpClient       *http.Client
	govppProxyClient *proxy.Client

	customHTTPHeaders map[string]string
	version           string
	manualOverride    bool
	negotiateVersion  bool
	negotiated        bool
}

// NewClient returns client with host option.
func NewClient(host string) (*Client, error) {
	return NewClientWithOpts(WithHost(host))
}

// NewClientWithOpts returns client with ops applied.
func NewClientWithOpts(ops ...Opt) (*Client, error) {
	c := &Client{
		host:     DefaultAgentHost,
		version:  api.DefaultVersion,
		proto:    "tcp",
		scheme:   "http",
		grpcPort: DefaultPortGRPC,
		httpPort: DefaultPortHTTP,
	}
	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}
	c.grpcAddr = net.JoinHostPort(c.host, strconv.Itoa(c.grpcPort))
	c.httpAddr = net.JoinHostPort(c.host, strconv.Itoa(c.httpPort))
	return c, nil
}

func (c *Client) AgentHost() string {
	return c.host
}

func (c *Client) Version() string {
	return c.version
}

// Close the transport used by the client
func (c *Client) Close() error {
	if c.httpClient != nil {
		if t, ok := c.httpClient.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
	}
	if c.grpcClient != nil {
		if err := c.grpcClient.Close(); err != nil {
			return err
		}
	}
	return nil
}

// GRPCConn returns configured gRPC client.
func (c *Client) GRPCConn() (*grpc.ClientConn, error) {
	if c.grpcClient == nil {
		conn, err := connectGrpc(c.grpcAddr, c.grpcTLS)
		if err != nil {
			return nil, err
		}
		c.grpcClient = conn
	}
	return c.grpcClient, nil
}

func (c *Client) GenericClient() (client.GenericClient, error) {
	conn, err := c.GRPCConn()
	if err != nil {
		return nil, err
	}
	return remoteclient.NewClientGRPC(conn, remoteclient.UseRemoteRegistry("config"))
}

// ConfiguratorClient returns "config" with gRPC connection.
func (c *Client) ConfiguratorClient() (configurator.ConfiguratorServiceClient, error) {
	conn, err := c.GRPCConn()
	if err != nil {
		return nil, err
	}
	return configurator.NewConfiguratorServiceClient(conn), nil
}

// MetaServiceClient creates new client for using meta service
func (c *Client) MetaServiceClient() (generic.MetaServiceClient, error) {
	conn, err := c.GRPCConn()
	if err != nil {
		return nil, err
	}
	return generic.NewMetaServiceClient(conn), nil
}

// HTTPClient returns configured HTTP client.
func (c *Client) HTTPClient() *http.Client {
	if c.httpClient == nil {
		tr := cloneHTTPTransport()
		tr.TLSClientConfig = c.httpTLS
		c.httpClient = &http.Client{
			Transport: tr,
			Timeout:   DefaultTimeout,
		}
	}
	return c.httpClient
}

// GoVPPProxyClient returns configured GoVPP proxy client that is already connected to the exposed
// GoVPP proxy from vpp-agent
func (c *Client) GoVPPProxyClient() (*proxy.Client, error) {
	if c.govppProxyClient == nil {
		cc, err := proxy.Connect(c.httpAddr)
		if err != nil {
			return nil, fmt.Errorf("connecting to proxy failed due to: %v", err)
		}
		c.govppProxyClient = cc
	}
	return c.govppProxyClient, nil
}

// KVDBClient returns configured KVDB client.
func (c *Client) KVDBClient() (KVDBAPIClient, error) {
	kvdb, err := connectEtcd(c.kvdbEndpoints, c.kvdbDialTimeout, c.kvdbTLS)
	if err != nil {
		return nil, fmt.Errorf("connecting to Etcd failed: %v", err)
	}
	return NewKVDBClient(kvdb, c.serviceLabel), nil
}

// ParseHostURL parses a url string, validates the string is a host url, and
// returns the parsed URL
func ParseHostURL(host string) (*url.URL, error) {
	if !strings.Contains(host, "://") {
		host = "tcp://" + host
	}
	protoAddrParts := strings.SplitN(host, "://", 2)
	if len(protoAddrParts) == 1 {
		return nil, fmt.Errorf("unable to parse agent host `%s`", host)
	}
	var basePath string
	proto, addr := protoAddrParts[0], protoAddrParts[1]
	if proto == "tcp" {
		parsed, err := url.Parse("tcp://" + addr)
		if err != nil {
			return nil, err
		}
		addr = parsed.Host
		basePath = parsed.Path
	}
	return &url.URL{
		Scheme: proto,
		Host:   addr,
		Path:   basePath,
	}, nil
}

// getAPIPath returns the versioned request path to call the api.
// It appends the query parameters to the path if they are not empty.
func (c *Client) getAPIPath(ctx context.Context, p string, query url.Values) string {
	var apiPath string
	if c.negotiateVersion && !c.negotiated {
		c.NegotiateAPIVersion(ctx)
	}
	if c.version != "" {
		v := strings.TrimPrefix(c.version, "v")
		apiPath = path.Join(c.basePath, "/v"+v, p)
	} else {
		apiPath = path.Join(c.basePath, p)
	}
	return (&url.URL{Path: apiPath, RawQuery: query.Encode()}).String()
}

func (c *Client) NegotiateAPIVersion(ctx context.Context) {
	if !c.manualOverride {
		version, _ := c.AgentVersion(ctx)
		c.negotiateAPIVersionPing(version)
	}
}

func (c *Client) NegotiateAPIVersionPing(p *types.Version) {
	if !c.manualOverride {
		c.negotiateAPIVersionPing(p)
	}
}

// negotiateAPIVersionPing queries the API and updates the version to match the
// API version. Any errors are silently ignored.
func (c *Client) negotiateAPIVersionPing(p *types.Version) {
	// try the latest version before versioning headers existed
	if p.APIVersion == "" {
		p.APIVersion = "0.1"
	}
	// if the client is not initialized with a version, start with the latest supported version
	if c.version == "" {
		c.version = api.DefaultVersion
	}
	// if server version is lower than the client version, downgrade
	if versions.LessThan(p.APIVersion, c.version) {
		c.version = p.APIVersion
	}
	// Store the results, so that automatic API version negotiation (if enabled)
	// won't be performed on the next request.
	if c.negotiateVersion {
		c.negotiated = true
	}
}

func connectGrpc(addr string, tc *tls.Config) (*grpc.ClientConn, error) {
	dialOpt := grpc.WithInsecure()
	if tc != nil {
		dialOpt = grpc.WithTransportCredentials(credentials.NewTLS(tc))
	}
	logging.Debugf("dialing grpc address: %v", addr)
	return grpc.Dial(addr, dialOpt)
}

func connectEtcd(endpoints []string, dialTimeout time.Duration, tc *tls.Config) (keyval.CoreBrokerWatcher, error) {
	log := logrus.NewLogger("etcd-client")
	if debug.IsEnabledFor("kvdb") {
		log.SetLevel(logging.DebugLevel)
	} else {
		log.SetLevel(logging.WarnLevel)
	}
	dt := DefaultEtcdDialTimeout
	if dialTimeout > 0 {
		dt = dialTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), dt)
	defer cancel()
	cfg := etcd.ClientConfig{
		Config: &clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: dt,
			TLS:         tc,
			Context:     ctx,
		},
		OpTimeout: DefaultEtcdOpTimeout,
	}
	kvdb, err := etcd.NewEtcdConnectionWithBytes(cfg, log)
	if err != nil {
		return nil, err
	}
	return kvdb, nil
}
