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
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/config"
	"github.com/namsral/flag"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "EtcdClient"

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	LogFactory     logging.LogFactory
	ConfigFileName string
	*plugin.Skeleton
}

var defaultConfigFileName string

func init() {
	flag.StringVar(&defaultConfigFileName, "etcdv3-config", "", "Location of the Etcd configuration file; also set via 'ETCDV3_CONFIG' env variable.")
}

func (p *Plugin) retrieveConfig() (*Config, error) {
	cfg := &Config{}
	var configFile string
	if p.ConfigFileName != "" {
		configFile = p.ConfigFileName
	} else if defaultConfigFileName != "" {
		configFile = defaultConfigFileName
	}

	if configFile != "" {
		err := config.ParseConfigFromYamlFile(configFile, cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// Init is called at plugin startup. The connection to etcd is established.
func (p *Plugin) Init() error {
	cfg, err := p.retrieveConfig()
	if err != nil {
		return err
	}

	skeleton := plugin.NewSkeleton(string(PluginID), p.LogFactory,
		func(log logging.Logger) (plugin.Connection, error) {
			etcdConfig, err := ConfigToClientv3(cfg)
			if err != nil {
				return nil, err
			}
			return NewEtcdConnectionWithBytes(*etcdConfig, log)
		},
	)
	p.Skeleton = skeleton
	return p.Skeleton.Init()
}
