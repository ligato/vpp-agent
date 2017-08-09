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
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/flavors/generic"
	"github.com/ligato/cn-infra/messaging/kafka"
)

// Flavor glues together generic.Flavor plugins with:
// - ETCD (useful for watching config.)
// - Kafka plugins (useful for publishing events)
type Flavor struct {
	Generic generic.Flavor
	Etcd    etcdv3.Plugin
	Kafka   kafka.Plugin

	injected bool
}

// Inject sets object references
func (f *Flavor) Inject() error {
	if f.injected {
		return nil
	}

	f.Generic.Inject()

	f.Etcd.LogFactory = &f.Generic.Logrus
	f.Etcd.ServiceLabel = &f.Generic.ServiceLabel
	f.Etcd.StatusCheck = &f.Generic.StatusCheck
	f.Kafka.LogFactory = &f.Generic.Logrus
	f.Kafka.ServiceLabel = &f.Generic.ServiceLabel
	f.Kafka.StatusCheck = &f.Generic.StatusCheck

	f.injected = true

	return nil
}

// Plugins combines all Plugins in flavor to the list
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
