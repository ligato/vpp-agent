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
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/servicelabel"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "RedisClient"

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	LogFactory   logging.LogFactory
	ServiceLabel *servicelabel.Plugin
	*plugin.Skeleton
}

// Init is called on plugin startup. It establishes the connection to redis.
func (p *Plugin) Init() error {

	// FIXME: properly retrieve config
	pool, err := CreateNodeClientConnPool(NodeClientConfig{})
	if err != nil {
		return err
	}

	skeleton := plugin.NewSkeleton(string(PluginID), p.LogFactory, p.ServiceLabel,
		func(log logging.Logger) (plugin.Connection, error) {
			return NewBytesConnectionRedis(pool, log)
		},
	)
	p.Skeleton = skeleton
	return p.Skeleton.Init()
}
