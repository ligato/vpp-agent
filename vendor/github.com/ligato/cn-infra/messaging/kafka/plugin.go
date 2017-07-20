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

package kafka

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/messaging/kafka/mux"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "KafkaClient"

var kafkaConfigFile string

func init() {
	flag.StringVar(&kafkaConfigFile, "kafka-config", "", "Location of the Kafka configuration file; also set via 'KAFKA_CONFIG' env variable.")
}

// Mux defines API for the plugins that use access to kafka brokers.
type Mux interface {
	NewConnection(name string) *mux.Connection
	NewProtoConnection(name string) *mux.ProtoConnection
}

// Plugin provides API for interaction with kafka brokers.
type Plugin struct {
	LogFactory   logging.LogFactory
	ServiceLabel *servicelabel.Plugin
	mx           *mux.Multiplexer
}

// Init is called at plugin initialization.
func (p *Plugin) Init() error {
	logger, err := p.LogFactory.NewLogger(string(PluginID))
	if err != nil {
		return err
	}

	p.mx, err = mux.InitMultiplexer(kafkaConfigFile, p.ServiceLabel.GetAgentLabel(), logger)
	return err
}

// AfterInit is called in the second phase of initialization. The kafka multiplexer
// is started, all consumers have to be subscribed until this phase.
func (p *Plugin) AfterInit() error {
	return p.mx.Start()
}

// Close is called at plugin cleanup phase.
func (p *Plugin) Close() error {
	return safeclose.Close(p.mx)
}
