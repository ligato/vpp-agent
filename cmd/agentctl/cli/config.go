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
	"fmt"
	"os"
	"path/filepath"

	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/logging"
)

const (
	configFileDir  = ".agentctl"
	configFileName = "config.yml"
)

// TLSConfig represents configuration for TLS.
type TLSConfig struct {
	Disabled   bool   `json:"disabled"`
	SkipVerify bool   `json:"skip-verify"`
	Certfile   string `json:"cert-file"`
	Keyfile    string `json:"key-file"`
	CAfile     string `json:"ca-file"`
}

// ConfigFile represents info from ~/.agentctl/config.yml.
type ConfigFile struct {
	GrpcTLS TLSConfig `json:"grpc-tls"`
	KvdbTLS TLSConfig `json:"kvdb-tls"`
}

// DefaultConfigDir returns default path to agentctl's config.
func DefaultConfigDir() string {
	uhd, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get current user's home directory: %v\n", err)
		os.Exit(1)
	}

	p := filepath.Join(uhd, configFileDir)
	logging.Debugf("default path to directory with agentctl's config is '%s'", p)
	return p
}

// ReadConfig parses a config file in `dirPath` directory.
func ReadConfig(dirPath string) (*ConfigFile, error) {
	filename := filepath.Join(dirPath, configFileName)
	logging.Debugf("reading config file from %s", filename)

	cf := &ConfigFile{}

	err := config.ParseConfigFromYamlFile(filename, cf)
	if err != nil {
		return cf, fmt.Errorf("error parsing config file: %v", err)
	}

	logging.Debugf("config file data: %+v", cf)
	return cf, nil
}
