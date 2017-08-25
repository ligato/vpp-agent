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
	"github.com/namsral/flag"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/flavors/local"
)

// defines redis flags // TODO switch to viper to avoid global configuration
func init() {
	flag.String("redis-config", "",
		"Location of Redis configuration file; Can also be set via environment variable REDIS_CONFIG")
}

// FlavorRedis glues together FlavorRPC plugins with:
// - ETCD (useful for watching config.)
// - Kafka plugins (useful for publishing events)
type FlavorRedis struct {
	*local.FlavorLocal

	Redis         redis.Plugin
	RedisDataSync kvdbsync.Plugin

	injected bool
}

// Inject sets object references
func (f *FlavorRedis) Inject() (allReadyInjected bool) {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	f.Redis.Deps.PluginInfraDeps = *f.InfraDeps("redis")
	f.RedisDataSync.Deps.PluginLogDeps = *f.LogDeps("redis-datasync")
	f.RedisDataSync.KvPlugin = &f.Redis
	f.RedisDataSync.ResyncOrch = &f.ResyncOrch
	f.RedisDataSync.ServiceLabel = &f.ServiceLabel

	if f.StatusCheck.Transport == nil {
		f.StatusCheck.Transport = &f.RedisDataSync
	}

	return true
}

// Plugins combines all Plugins in flavor to the list
func (f *FlavorRedis) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
