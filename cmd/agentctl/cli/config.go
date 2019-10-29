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

	"github.com/ligato/cn-infra/logging"
	"github.com/spf13/viper"
)

const (
	configFileDir  = ".agentctl"
	configFileName = "config.yml"
	configFileType = "yaml"
)

// DefaultConfigDir returns default path to agentctl's config.
func DefaultConfigDir() string {
	uhd, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get current user's home directory: %v\n", err)
		os.Exit(1)
	}

	p := filepath.Join(uhd, configFileDir)
	logging.Debugf("default path to directory with agentctl's config is %q", p)
	return p
}

// ReadConfig loads config using Viper.
func ReadConfig() {
	cfgFile := filepath.Join(viper.GetString("config-dir"), configFileName)
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType(configFileType)

	err := viper.ReadInConfig()
	if err == nil {
		logging.Debugf("using config file: %q", viper.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logging.Debugf("Config file not found at %q", viper.GetString("config-dir"))
		} else {
			logging.Debugf("Config file was found but another error was produced: %v", err)
		}
	}
}
