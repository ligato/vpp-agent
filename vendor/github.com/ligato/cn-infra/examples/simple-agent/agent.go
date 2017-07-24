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

package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/http"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/messaging/kafka"
	"github.com/ligato/cn-infra/servicelabel"
	"os"
	"time"
)

// Flavour is set of common used generic plugins. This flavour can be used as a base
// for different flavours. The plugins are initialized in the same order as they appear
// in the structure.
type Flavour struct {
	injected bool

	Logrus       logrus.Plugin
	HTTP         http.Plugin
	ServiceLabel servicelabel.Plugin
	Etcd         etcdv3.Plugin
	Kafka        kafka.Plugin
}

// Inject interconnects plugins - injects the dependencies. If it has been called
// already it is no op.
func (g *Flavour) Inject() error {
	if g.injected {
		return nil
	}
	g.injected = true

	g.HTTP.LogFactory = &g.Logrus
	g.Etcd.LogFactory = &g.Logrus
	g.Etcd.ServiceLabel = &g.ServiceLabel
	g.Kafka.LogFactory = &g.Logrus
	g.Kafka.ServiceLabel = &g.ServiceLabel
	return nil
}

// Plugins returns all plugins from the flavour. The set of plugins is supposed
// to be passed to the agent constructor. The method calls inject to make sure that
// dependencies have been injected.
func (g *Flavour) Plugins() []*core.NamedPlugin {
	g.Inject()
	return core.ListPluginsInFlavor(g)
}

func main() {
	logroot.Logger().SetLevel(logging.DebugLevel)

	f := Flavour{}
	agent := core.NewAgent(logroot.Logger(), 15*time.Second, f.Plugins()...)

	err := core.EventLoopWithInterrupt(agent, nil)
	if err != nil {
		os.Exit(1)
	}
}
