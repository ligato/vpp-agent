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
	"encoding/json"
	"fmt"
	"strings"
)

const (
	configFileDir  = ".agentctl"
	configFileName = "config"
)

// TLSConfig represents configuration for TLS.
type TLSConfig struct {
	Disabled   bool   `json:"disabled"`
	CertFile   string `json:"cert-file"`
	KeyFile    string `json:"key-file"`
	CAFile     string `json:"ca-file"`
	SkipVerify bool   `json:"skip-verify"`
}

// Config represents configuration for AgentCTL.
type Config struct {
	LigatoAPIVersion string     `json:"ligato-api-version"`
	Host             string     `json:"host"`
	ServiceLabel     string     `json:"service-label"`
	GRPCPort         int        `json:"grpc-port"`
	HTTPPort         int        `json:"http-port"`
	ETCDEndpoints    []string   `json:"etcd-endpoints"`
	BasicAuth        string     `json:"basic-auth"`
	UseTLS           bool       `json:"use-tls"`
	GRPCSecure       *TLSConfig `json:"grpc-tls"`
	HTTPSecure       *TLSConfig `json:"http-tls"`
	KVDBSecure       *TLSConfig `json:"kvdb-tls"`
}

// MakeConfig returns new Config with values from Viper.
func MakeConfig() (*Config, error) {
	// Prepare Viper.
	viperSetConfigFile(configFileName, configFileDir)
	viperReadInConfig()

	// Put configuration into "Config" struct.
	cfg := &Config{}
	err := viperUnmarshal(cfg)
	if err != nil {
		return nil, err
	}

	// Values adjustment.
	cfg.ETCDEndpoints = adjustETCDEndpoints(cfg.ETCDEndpoints)

	return cfg, nil
}

// DebugOutput returns Config as string to be used for debug output.
func (c *Config) DebugOutput() string {
	bConfig, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return fmt.Sprintf("error while marshaling config to json: %v", err)
	}

	return string(bConfig)
}

// ShouldUseSecureGRPC returns whether or not to use TLS for GRPC connection.
func (c *Config) ShouldUseSecureGRPC() bool {
	return c.UseTLS == true && c.GRPCSecure != nil && !c.GRPCSecure.Disabled
}

// ShouldUseSecureHTTP returns whether or not to use TLS for HTTP connection.
func (c *Config) ShouldUseSecureHTTP() bool {
	return c.UseTLS == true && c.HTTPSecure != nil && !c.HTTPSecure.Disabled
}

// ShouldUseSecureKVDB returns whether or not to use TLS for KVDB connection.
func (c *Config) ShouldUseSecureKVDB() bool {
	return c.UseTLS == true && c.KVDBSecure != nil && !c.KVDBSecure.Disabled
}

// adjustETCDEndpoints adjusts ETCD endpoints received from env variable.
func adjustETCDEndpoints(endpoints []string) []string {
	if len(endpoints) != 1 {
		return endpoints
	}

	if strings.Contains(endpoints[0], ",") {
		return strings.Split(endpoints[0], ",")
	}

	return endpoints
}
