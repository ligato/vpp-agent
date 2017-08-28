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

package allcon

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/cassandra"
	"github.com/ligato/cn-infra/flavors/etcd"
	"github.com/ligato/cn-infra/flavors/kafka"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/flavors/redis"
	"github.com/ligato/cn-infra/flavors/rpc"
)

// AllConnectorsFlavor is combination of RPC, ETCD, Kafka, Redis, Cassandra flavors
// User can enable those connectors by providing configs for them.
type AllConnectorsFlavor struct {
	*local.FlavorLocal
	*etcd.FlavorEtcd
	*kafka.FlavorKafka
	*redis.FlavorRedis
	*cassandra.FlavorCassandra
	*rpc.FlavorRPC

	injected bool
}

// Inject sets object references
func (f *AllConnectorsFlavor) Inject() bool {
	if f.injected {
		return false
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	if f.FlavorEtcd == nil {
		f.FlavorEtcd = &etcd.FlavorEtcd{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorEtcd.Inject(nil)

	if f.FlavorKafka == nil {
		f.FlavorKafka = &kafka.FlavorKafka{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorKafka.Inject()

	if f.FlavorRedis == nil {
		f.FlavorRedis = &redis.FlavorRedis{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorRedis.Inject(nil)

	if f.FlavorCassandra == nil {
		f.FlavorCassandra = &cassandra.FlavorCassandra{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorCassandra.Inject()

	if f.FlavorRPC == nil {
		f.FlavorRPC = &rpc.FlavorRPC{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorRPC.Inject()

	return true
}

// Plugins combines all Plugins in flavor to the list
func (f *AllConnectorsFlavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
