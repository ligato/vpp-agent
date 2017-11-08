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

package etcdv3

import (
	"crypto/tls"
	"crypto/x509"

	"time"

	"os"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/tlsutil"
)

// Config represents a part of the etcd configuration that can be
// loaded from a file. Usually, the Config is next transformed into
// ClientConfig using ConfigToClientv3() function for use with the coreos/etcd
// package.
type Config struct {
	Endpoints             []string      `json:"endpoints"`
	DialTimeout           time.Duration `json:"dial-timeout"`
	OpTimeout             time.Duration `json:"operation-timeout"`
	InsecureTransport     bool          `json:"insecure-transport"`
	InsecureSkipTLSVerify bool          `json:"insecure-skip-tls-verify"`
	Certfile              string        `json:"cert-file"`
	Keyfile               string        `json:"key-file"`
	CAfile                string        `json:"ca-file"`
}

// ClientConfig extends clientv3.Config with configuration options introduced
// by this package.
type ClientConfig struct {
	*clientv3.Config

	// OpTimeout is the maximum amount of time the client will wait for a pending
	// operation before timing out.
	OpTimeout time.Duration
}

// defaultDialTimeout defines the default timeout for connecting to etcd.
const defaultDialTimeout = 1 * time.Second

// defaultOpTimeout defines the default timeout for any request-reply etcd
// operation.
const defaultOpTimeout = 3 * time.Second

// ConfigToClientv3 transforms yaml configuration <yc> modelled by Config
// into ClientConfig, which is ready for use with the underlying coreos/etcd
// package.
// If the etcd endpoint addresses are not specified in the configuration,
// the function will query the ETCDV3_ENDPOINTS environment variable
// for a non-empty value. If neither the config nor the environment specify the
// endpoint location, a default address "127.0.0.1:2379" is assumed.
// The function may return error only if TLS connection is selected and the
// CA or client certificate is not accessible/valid.
func ConfigToClientv3(yc *Config) (*ClientConfig, error) {

	dialTimeout := defaultDialTimeout

	if yc.DialTimeout != 0 {
		dialTimeout = yc.DialTimeout
	}

	opTimeout := defaultOpTimeout

	if yc.OpTimeout != 0 {
		opTimeout = yc.OpTimeout
	}

	clientv3Cfg := &clientv3.Config{
		Endpoints:   yc.Endpoints,
		DialTimeout: dialTimeout,
	}
	cfg := &ClientConfig{Config: clientv3Cfg, OpTimeout: opTimeout}

	if len(cfg.Endpoints) == 0 {
		if ep := os.Getenv("ETCDV3_ENDPOINTS"); ep != "" {
			cfg.Endpoints = strings.Split(ep, ",")
		} else {
			cfg.Endpoints = []string{"127.0.0.1:2379"}
		}
	}

	if yc.InsecureTransport {
		cfg.TLS = nil
		return cfg, nil
	}

	var (
		cert *tls.Certificate
		cp   *x509.CertPool
		err  error
	)

	if yc.Certfile != "" && yc.Keyfile != "" {
		cert, err = tlsutil.NewCert(yc.Certfile, yc.Keyfile, nil)
		if err != nil {
			return nil, err
		}
	}

	if yc.CAfile != "" {
		cp, err = tlsutil.NewCertPool([]string{yc.CAfile})
		if err != nil {
			return nil, err
		}
	}

	tlscfg := &tls.Config{
		MinVersion:         tls.VersionTLS10,
		InsecureSkipVerify: yc.InsecureSkipTLSVerify,
		RootCAs:            cp,
	}
	if cert != nil {
		tlscfg.Certificates = []tls.Certificate{*cert}
	}
	if yc.Certfile != "" || yc.Keyfile != "" || yc.CAfile != "" {
		cfg.TLS = tlscfg
	}

	return cfg, nil
}
