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

// Package tlsconfig provides more convenient way to create "tls.Config".
//
// Usage:
// 		package main
//
// 		import "fmt"
//		import "go.ligato.io/vpp-agent/v3/cmd/agentctl/client/tlsconfig"
//
// 		func main() {
// 			tc, err := tlsconfig.New(
// 				tlsconfig.CA("/path/to/ca.crt"),
// 				tlsconfig.CertKey("/path/to/server.crt", "/path/to/server.key"),
// 			)
//
// 			if err != nil {
// 				fmt.Printf("Error while creating TLS config: %v\n", err)
// 				return
// 			}
// 			fmt.Println("TLS config is ready to use")
//
// 			// `tc` usage
// 		}
//
package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// New returns tls.Config with all options applied.
func New(options ...Option) (*tls.Config, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	for _, op := range options {
		if err := op(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Option applies a modification on a tls.Config.
type Option func(config *tls.Config) error

// CA adds CA certificate from file to tls.Config.
// If not using this Option, then TLS will be using the host's root CA set.
func CA(path string) Option {
	return func(config *tls.Config) error {
		if config.RootCAs == nil {
			config.RootCAs = x509.NewCertPool()
		}

		cert, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		ok := config.RootCAs.AppendCertsFromPEM(cert)
		if !ok {
			return fmt.Errorf("unable to add CA from '%s' file", path)
		}

		return nil
	}
}

// CertKey adds certificate with key to tls.Config.
func CertKey(certPath, keyPath string) Option {
	return func(config *tls.Config) error {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return err
		}
		config.Certificates = append(config.Certificates, cert)
		return err
	}
}

// SkipServerVerification turns off verification of server's certificate chain and host name.
func SkipServerVerification() Option {
	return func(config *tls.Config) error {
		config.InsecureSkipVerify = true
		return nil
	}
}
