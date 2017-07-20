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

package redis

import (
	"io/ioutil"
	"time"

	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/coreos/etcd/pkg/tlsutil"
	"github.com/garyburd/redigo/redis"
	"github.com/ghodss/yaml"
)

// ConnPool provides abstraction of connection pool.
type ConnPool interface {
	// Get returns a vlid connection. The application must close the returned connection.
	Get() redis.Conn
	// Close releases the resources used by the pool.
	Close() error
}

// ConnPoolConfig configures connection pool
type ConnPoolConfig struct {
	// Properties mimic those in github.com/garyburd/redigo/redis/Pool

	// Maximum number of idle connections in the pool.
	MaxIdle int `json:"max-idle"`

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActive int `json:"max-active"`

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration `json:"idle-timeout"`

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
	Wait bool `json:"wait-on-max-active"`
}

// TLS configures TLS properties
type TLS struct {
	Enabled    bool   `json:"enabled"`     // enable/disable TLS
	SkipVerify bool   `json:"skip-verify"` // whether to skip verification of server name & certificate
	Certfile   string `json:"cert-file"`   // client certificate
	Keyfile    string `json:"key-file"`    // client private key
	CAfile     string `json:"ca-file"`     // certificate authority
}

// NodeClientConfig configures a node client that will connect to a single Redis server
type NodeClientConfig struct {
	Endpoint     string         `json:"endpoint"`      // like "172.17.0.1:6379"
	Db           int            `json:"db"`            // ID of the DB
	Password     string         `json:"password"`      // password, if required
	ReadTimeout  time.Duration  `json:"read-timeout"`  // timeout for read operations
	WriteTimeout time.Duration  `json:"write-timeout"` // timeout for write operations
	TLS          TLS            `json:"tls"`           // TLS configuration
	Pool         ConnPoolConfig `json:"pool"`          // connection pool configuration
}

// CreateNodeClientConnPool creates a Redis connection pool
func CreateNodeClientConnPool(config NodeClientConfig) (*redis.Pool, error) {
	options := append([]redis.DialOption{}, redis.DialDatabase(config.Db))
	options = append(options, redis.DialPassword(config.Password))
	options = append(options, redis.DialReadTimeout(config.ReadTimeout))
	options = append(options, redis.DialWriteTimeout(config.WriteTimeout))
	if config.TLS.Enabled {
		tlsConfig, err := createTLSConfig(config.TLS)
		if err != nil {
			return nil, err
		}
		options = append(options, redis.DialTLSConfig(tlsConfig))
		options = append(options, redis.DialTLSSkipVerify(config.TLS.SkipVerify))
	}
	return &redis.Pool{
		MaxIdle:     config.Pool.MaxIdle,
		MaxActive:   config.Pool.MaxActive,
		IdleTimeout: config.Pool.IdleTimeout,
		Wait:        config.Pool.Wait,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", config.Endpoint, options...) },
	}, nil
}

func createTLSConfig(config TLS) (*tls.Config, error) {
	var (
		cert *tls.Certificate
		cp   *x509.CertPool
		err  error
	)
	if config.Certfile != "" && config.Keyfile != "" {
		cert, err = tlsutil.NewCert(config.Certfile, config.Keyfile, nil)
		if err != nil {
			return nil, fmt.Errorf("tlsutil.NewCert() failed: %s", err)
		}
	}

	if config.CAfile != "" {
		cp, err = tlsutil.NewCertPool([]string{config.CAfile})
		if err != nil {
			return nil, fmt.Errorf("tlsutil.NewCertPool() failed: %s", err)
		}
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS10,
		InsecureSkipVerify: config.SkipVerify,
		RootCAs:            cp,
	}
	if cert != nil {
		tlsConfig.Certificates = []tls.Certificate{*cert}
	}

	return tlsConfig, nil
}

// GenerateConfig Generates a yaml file using the given configuration object
func GenerateConfig(config interface{}, path string) error {
	bytes, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("yaml.Marshal() failed: %s", err)
	}
	err = ioutil.WriteFile(path, bytes, 0644)
	if err != nil {
		return fmt.Errorf("ioutil.WriteFile() failed: %s", err)
	}
	return nil
}
