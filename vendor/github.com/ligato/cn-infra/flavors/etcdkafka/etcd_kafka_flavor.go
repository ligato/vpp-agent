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
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/etcd"
	"github.com/ligato/cn-infra/flavors/kafka"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/datasync/resync"
)

// FlavorEtcdKafka glues together FlavorLocal plugins with:
// - ETCD (useful for watching northbound config.)
// - Kafka plugins (useful for publishing events)
type FlavorEtcdKafka struct {
	*local.FlavorLocal
	*etcd.FlavorEtcd
	*kafka.FlavorKafka

	injected bool
}

// Inject sets object references
func (f *FlavorEtcdKafka) Inject(resyncOrch *resync.Plugin) bool {
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
	f.FlavorEtcd.Inject(resyncOrch)

	if f.FlavorKafka == nil {
		f.FlavorKafka = &kafka.FlavorKafka{FlavorLocal: f.FlavorLocal}
	}
	f.FlavorKafka.Inject()

	return true
}

// Plugins combines all Plugins in flavor to the list
func (f *FlavorEtcdKafka) Plugins() []*core.NamedPlugin {
	f.Inject(nil)
	return core.ListPluginsInFlavor(f)
}
