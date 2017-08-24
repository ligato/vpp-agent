// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cassandra

import (
	"strings"
	"time"

	"github.com/gocql/gocql"
)

// Config Configuration for Cassandra clients loaded from a configuration file
type Config struct {

	// A list of host addresses of cluster nodes.
	Endpoints []string `json:"endpoints"`

	// port for Cassandra (default: 9042)
	Port int `json:"port"`

	// connection timeout (default: 600ms)
	OpTimeout time.Duration `json:"op_timeout"`

	// initial connection timeout, used during initial dial to server (default: 600ms)
	DialTimeout time.Duration `json:"dial_timeout"`

	// If not zero, gocql attempt to reconnect known DOWN nodes in every ReconnectSleep.
	RedialInterval time.Duration `json:"redial_interval"`

	// ProtoVersion sets the version of the native protocol to use, this will
	// enable features in the driver for specific protocol versions, generally this
	// should be set to a known version (2,3,4) for the cluster being connected to.
	//
	// If it is 0 or unset (the default) then the driver will attempt to discover the
	// highest supported protocol for the cluster. In clusters with nodes of different
	// versions the protocol selected is not defined (ie, it can be any of the supported in the cluster)
	ProtocolVersion int `json:"protocol_version"`
}

// ClientConfig wrapping gocql ClusterConfig
type ClientConfig struct {
	*gocql.ClusterConfig
}

const defaultOpTimeout = 600 * time.Millisecond
const defaultDialTimeout = 600 * time.Millisecond
const defaultRedialInterval = 60 * time.Second
const defaultProtocolVersion = 4

// ConfigToClientConfig transforms the yaml configuration into ClientConfig.
func ConfigToClientConfig(ymlConfig *Config) (*ClientConfig, error) {

	timeout := defaultOpTimeout
	if ymlConfig.OpTimeout > 0 {
		timeout = ymlConfig.OpTimeout
	}

	connectTimeout := defaultDialTimeout
	if ymlConfig.DialTimeout > 0 {
		connectTimeout = ymlConfig.DialTimeout
	}

	reconnectInterval := defaultRedialInterval
	if ymlConfig.RedialInterval > 0 {
		reconnectInterval = ymlConfig.RedialInterval
	}

	protoVersion := defaultProtocolVersion
	if ymlConfig.ProtocolVersion > 0 {
		protoVersion = ymlConfig.ProtocolVersion
	}

	clientConfig := &gocql.ClusterConfig{
		Hosts:             ymlConfig.Endpoints,
		Port:              ymlConfig.Port,
		Timeout:           timeout,
		ConnectTimeout:    connectTimeout,
		ReconnectInterval: reconnectInterval,
		ProtoVersion:      protoVersion,
	}

	cfg := &ClientConfig{ClusterConfig: clientConfig}

	return cfg, nil
}

// CreateSessionFromConfig Creates session from given configuration and keyspace
func CreateSessionFromConfig(config *ClientConfig) (*gocql.Session, error) {

	gocqlClusterConfig := gocql.NewCluster(HostsAsString(config.Hosts))

	session, err := gocqlClusterConfig.CreateSession()

	if err != nil {
		return nil, err
	}

	return session, nil
}

// HostsAsString converts an array of hosts addresses into a comma separated string
func HostsAsString(hostArr []string) string {
	return strings.Join(hostArr, ",")
}
