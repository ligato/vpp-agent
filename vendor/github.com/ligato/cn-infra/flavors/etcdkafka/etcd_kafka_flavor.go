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

package etcdkafka

import (
	"github.com/namsral/flag"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/cn-infra/messaging/kafka"
)

// defines etcd & kafka flags // TODO switch to viper to avoid global configuration
func init() {
	flag.String("etcdv3-config", "etcd.conf",
		"Location of the Etcd configuration file; also set via 'ETCDV3_CONFIG' env variable.")
	flag.String("kafka-config", "kafka.conf",
		"Location of the Kafka configuration file; also set via 'KAFKA_CONFIG' env variable.")
}

// FlavorEtcdKafka glues together FlavorRPC plugins with:
// - ETCD (useful for watching config.)
// - Kafka plugins (useful for publishing events)
type FlavorEtcdKafka struct {
	rpc.FlavorRPC

	ETCD         etcdv3.Plugin
	ETCDDataSync kvdbsync.Plugin

	Kafka kafka.Plugin

	injected bool
}

// Inject sets object references
func (f *FlavorEtcdKafka) Inject() (allReadyInjected bool) {
	if !f.FlavorRPC.Inject() {
		return false
	}

	f.ETCD.Deps.PluginInfraDeps = *f.InfraDeps("etcdv3")
	f.ETCDDataSync.Deps.PluginLogDeps = *f.LogDeps("etcdv3-datasync")
	f.ETCDDataSync.KvPlugin = &f.ETCD
	f.ETCDDataSync.ResyncOrch = &f.ResyncOrch
	f.ETCDDataSync.ServiceLabel = &f.ServiceLabel

	f.StatusCheck.Transport = &f.ETCDDataSync

	f.Kafka.Deps.PluginInfraDeps = *f.InfraDeps("kafka")

	return true
}

// Plugins combines all Plugins in flavor to the list
func (f *FlavorEtcdKafka) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
