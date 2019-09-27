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
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"

	"github.com/ligato/vpp-agent/api"
	"github.com/ligato/vpp-agent/api/genericmanager"
	"github.com/ligato/vpp-agent/api/types"
	"github.com/ligato/vpp-agent/client"
	"github.com/ligato/vpp-agent/client/remoteclient"
	"github.com/ligato/vpp-agent/pkg/debug"
)

var (
	DefaultAgentHost = "localhost"
	DefaultPortGRPC  = 9111
	DefaultPortHTTP  = 9191
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

	grpcAddr     string
	httpAddr     string
	endpoints    []string
	serviceLabel string

	grpcClient *grpc.ClientConn
	httpClient *http.Client

	customHTTPHeaders map[string]string
	version           string
	manualOverride    bool
	negotiateVersion  bool
	negotiated        bool
}

func NewClient(host string) (*Client, error) {
	return NewClientWithOpts(WithHost(host))
}

func NewClientWithOpts(ops ...Opt) (*Client, error) {
	c := &Client{
		host:       DefaultAgentHost,
		version:    api.DefaultVersion,
		httpClient: &http.Client{},
		proto:      "tcp",
		scheme:     "http",
	}
	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Client) ConfigClient() (client.ConfigClient, error) {
	conn, err := c.GRPCConn()
	if err != nil {
		return nil, err
	}
	genericClient := genericmanager.NewGenericManagerClient(conn)
	return remoteclient.NewClientGRPC(genericClient), nil
}

func (c *Client) AgentHost() string {
	return c.host
}

func (c *Client) ClientVersion() string {
	return c.version
}

// Close the transport used by the client
func (c *Client) Close() error {
	if t, ok := c.httpClient.Transport.(*http.Transport); ok {
		t.CloseIdleConnections()
	}
	if c.grpcClient != nil {
		if err := c.grpcClient.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) GRPCConn() (*grpc.ClientConn, error) {
	if c.grpcClient == nil {
		logging.Debugf("dialing grpc address: %v", c.grpcAddr)

		conn, err := grpc.Dial(c.grpcAddr, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		c.grpcClient = conn
	}
	return c.grpcClient, nil
}

func (c *Client) HTTPClient() *http.Client {
	return &*c.httpClient
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
		ping, _ := c.Ping(ctx)
		c.negotiateAPIVersionPing(ping)
	}
}

func (c *Client) NegotiateAPIVersionPing(p types.Ping) {
	if !c.manualOverride {
		c.negotiateAPIVersionPing(p)
	}
}

// negotiateAPIVersionPing queries the API and updates the version to match the
// API version. Any errors are silently ignored.
func (c *Client) negotiateAPIVersionPing(p types.Ping) {
	// try the latest version before versioning headers existed
	if p.APIVersion == "" {
		p.APIVersion = "0.1"
	}

	// if the client is not initialized with a version, start with the latest supported version
	if c.version == "" {
		c.version = api.DefaultVersion
	}

	// if server version is lower than the client version, downgrade
	/*if versions.LessThan(p.APIVersion, cli.version) {
		cli.version = p.APIVersion
	}*/

	// Store the results, so that automatic API version negotiation (if enabled)
	// won't be performed on the next request.
	if c.negotiateVersion {
		c.negotiated = true
	}
}

func (c *Client) KVDBClient() (keyval.BytesBroker, error) {
	etcdCfg := getEtcdConfig(c.endpoints)

	log := logrus.NewLogger("kvdb-client")
	if debug.IsEnabledFor("kvdb") {
		log.SetLevel(logging.DebugLevel)
	} else {
		log.SetLevel(logging.WarnLevel)
	}
	kvdb, err := etcd.NewEtcdConnectionWithBytes(etcdCfg, log)
	if err != nil {
		return nil, fmt.Errorf("connecting to Etcd failed: %v", err)
	}

	kvdbc := &kvdbClient{
		BytesBroker:  kvdb,
		serviceLabel: c.serviceLabel,
	}
	return kvdbc, nil
}

func getEtcdConfig(endpoints []string) etcd.ClientConfig {
	cfg := etcd.ClientConfig{
		Config: &clientv3.Config{
			Endpoints:   endpoints,
			DialTimeout: time.Second * 3,
		},
		OpTimeout: time.Second * 10,
	}
	return cfg
}

// kvdbClient provides client access to the KVDB server.
type kvdbClient struct {
	keyval.BytesBroker
	serviceLabel string
}

func (k *kvdbClient) Put(key string, data []byte, opts ...datasync.PutOption) (err error) {
	key, err = k.completeFullKey(key)
	if err != nil {
		return err
	}
	logging.Debugf("kvdbClient.Put: %s", key)

	return k.BytesBroker.Put(key, data, opts...)
}

func (k *kvdbClient) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	key, err = k.completeFullKey(key)
	if err != nil {
		return nil, false, 0, err
	}
	logging.Debugf("kvdbClient.GetValue: %s", key)

	return k.BytesBroker.GetValue(key)
}

func (k *kvdbClient) ListValues(prefix string) (keyval.BytesKeyValIterator, error) {
	prefix = ensureAllAgentsPrefix(prefix)
	logging.Debugf("kvdbClient.ListValues: %s", prefix)

	return k.BytesBroker.ListValues(prefix)
}

func (k *kvdbClient) ListKeys(prefix string) (keyval.BytesKeyIterator, error) {
	prefix = ensureAllAgentsPrefix(prefix)
	logging.Debugf("kvdbClient.ListKeys: %s", prefix)

	return k.BytesBroker.ListKeys(prefix)
}

func (k *kvdbClient) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	key, err = k.completeFullKey(key)
	if err != nil {
		return false, err
	}
	logging.Debugf("kvdbClient.Delete: %s", key)

	return k.BytesBroker.Delete(key, opts...)
}

func (k *kvdbClient) completeFullKey(key string) (string, error) {
	if strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		return key, nil
	}
	if k.serviceLabel == "" {
		return "", fmt.Errorf("service label is not defined, cannot get complete key")
	}
	key = path.Join(servicelabel.GetAllAgentsPrefix(), k.serviceLabel, key)
	return key, nil
}

func ensureAllAgentsPrefix(key string) string {
	if strings.HasPrefix(key, servicelabel.GetAllAgentsPrefix()) {
		return key
	}
	return path.Join(servicelabel.GetAllAgentsPrefix(), key)
}
