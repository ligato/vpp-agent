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
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.ligato.io/cn-infra/v2/logging"
)

const (
	configFileDir  = ".agentctl"
	configFileName = "config"
)

func init() {
	viper.SupportedExts = append(viper.SupportedExts, "conf")
}

// viperSetConfigFile setups viper to handle config file.
func viperSetConfigFile() {
	// If "config" is set then use it and skip searching for config.
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		// Set config type explicitely
		if filepath.Ext(cfgFile) == ".conf" {
			viper.SetConfigType("yaml")
		}
		viper.SetConfigFile(cfgFile)
		return
	}

	viper.SetConfigName(configFileName)

	// If "config-dir" is set use it as first config path.
	if cfgDir := viper.GetString("config-dir"); cfgDir != "" {
		viper.AddConfigPath(cfgDir)
	}

	// Add current working directory and dir in home directory as fallback.
	viper.AddConfigPath(".")
	if uhd, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(uhd, configFileDir))
	}
}

// viperReadInConfig wraps viper.ReadInConfig with more logs.
func viperReadInConfig() {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logging.Debugf("unable to find config file: %v", err)
		} else {
			logging.Errorf("config file was found but another error was produced: %v", err)
		}
		return
	}

	logging.Debugf("using config file: %q", viper.ConfigFileUsed())
}

// viperUnmarshal wraps viper.Unmarshal with providing "json" as tag name.
func viperUnmarshal(c *Config) error {
	return viper.Unmarshal(
		c, func(c *mapstructure.DecoderConfig) { c.TagName = "json" },
	)
}
