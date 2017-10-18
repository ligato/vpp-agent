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

package connectors

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/db/sql/cassandra"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/messaging/kafka"
)

// AllConnectorsFlavor is a combination of all plugins that allow
// connectivity to external database/messaging...
// Effectively it is combination of ETCD, Kafka, Redis, Cassandra
// plugins.
//
// User/admin can enable those plugins/connectors by providing
// configs (at least endpoints) for them.
type AllConnectorsFlavor struct {
	*local.FlavorLocal

	ETCD         etcdv3.Plugin
	ETCDDataSync kvdbsync.Plugin

	Kafka kafka.Plugin

	Redis         redis.Plugin
	RedisDataSync kvdbsync.Plugin

	Cassandra cassandra.Plugin

	ResyncOrch resync.Plugin // the order is important because of AfterInit()

	injected bool
}

// Inject initializes flavor references/dependencies.
func (f *AllConnectorsFlavor) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	f.ETCD.Deps.PluginInfraDeps = *f.InfraDeps("etcdv3", local.WithConf())
	InjectKVDBSync(&f.ETCDDataSync, &f.ETCD, f.ETCD.PluginName, f.FlavorLocal, &f.ResyncOrch)

	f.Redis.Deps.PluginInfraDeps = *f.InfraDeps("redis", local.WithConf())
	InjectKVDBSync(&f.RedisDataSync, &f.Redis, f.Redis.PluginName, f.FlavorLocal, &f.ResyncOrch)

	f.Kafka.Deps.PluginInfraDeps = *f.InfraDeps("kafka", local.WithConf())

	f.Cassandra.Deps.PluginInfraDeps = *f.InfraDeps("cassandra", local.WithConf())

	f.ResyncOrch.PluginLogDeps = *f.LogDeps("resync-orch")

	return true
}

// Plugins combines all Plugins in flavor to the list
func (f *AllConnectorsFlavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
