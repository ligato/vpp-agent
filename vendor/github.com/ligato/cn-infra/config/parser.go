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

package config

import (
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
)

// ParseConfigFromYamlFile parses a configuration from a file
// in yaml format specified by path into cfg structure.
func ParseConfigFromYamlFile(path string, cfg interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return err
	}
	return nil
}

// SaveConfigToYamlFile Saves the given configuration to a yaml file.
// If not empty, each line in the 'comment' parameter must be proceeded by '#'.
func SaveConfigToYamlFile(cfg interface{}, path string, perm os.FileMode, comment string) error {
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if comment != "" {
		bytes = append([]byte(comment+"\n"), bytes...)
	}

	err = ioutil.WriteFile(path, bytes, perm)
	if err != nil {
		return err
	}
	return nil
}
