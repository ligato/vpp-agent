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
	"github.com/Shopify/sarama"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"github.com/ligato/cn-infra/messaging/kafka/mux"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/statuscheck"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "KafkaClient"

var configFile string

func init() {
	flag.StringVar(&configFile, "kafka-config", "", "Location of the Kafka configuration file; also set via 'KAFKA_CONFIG' env variable.")
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
	StatusCheck  *statuscheck.Plugin
	subscription chan (*client.ConsumerMessage)
	mx           *mux.Multiplexer
	consumer     *client.Consumer
}

// Init is called at plugin initialization.
func (p *Plugin) Init() error {
	logger, err := p.LogFactory.NewLogger(string(PluginID))
	if err != nil {
		return err
	}
	// Prepare topic and  subscription for status check client
	topic := "status-check"
	p.subscription = make(chan *client.ConsumerMessage)

	// Get config data
	config := &mux.Config{}
	if configFile != "" {
		config, err = mux.ConfigFromFile(configFile)
		if err != nil {
			return err
		}
	}
	clientConfig := p.getClientConfig(config, logger, topic)

	// Init consumer
	p.consumer, err = client.NewConsumer(clientConfig, nil)
	if err != nil {
		return err
	}

	// Register for providing status reports (polling mode)
	if p.StatusCheck != nil {
		p.StatusCheck.Register(PluginID, func() (statuscheck.PluginState, error) {
			// Method 'RefreshMetadata()' returns error if kafka server is unavailable
			err := p.consumer.Client.RefreshMetadata(topic)
			if err == nil {
				return statuscheck.OK, nil
			}
			logger.Errorf("Kafka server unavailable")
			return statuscheck.Error, err
		})
	} else {
		logger.Warnf("Unable to start status check for kafka")
	}

	p.mx, err = mux.InitMultiplexer(configFile, p.ServiceLabel.GetAgentLabel(), logger)

	return err
}

// AfterInit is called in the second phase of initialization. The kafka multiplexer
// is started, all consumers have to be subscribed until this phase.
func (p *Plugin) AfterInit() error {
	return p.mx.Start()
}

// Close is called at plugin cleanup phase.
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.consumer.Close(), p.mx)
	return err
}

// NewConnection returns a new instance of connection to access the kafka brokers.
func (p *Plugin) NewConnection(name string) *mux.Connection {
	return p.mx.NewConnection(name)
}

// NewProtoConnection returns a new instance of connection to access the kafka brokers. The connection
// uses proto-modelled messages.
func (p *Plugin) NewProtoConnection(name string) *mux.ProtoConnection {
	return p.mx.NewProtoConnection(name, &keyval.SerializerJSON{})
}

// Receive client config according to kafka config data
func (p *Plugin) getClientConfig(config *mux.Config, logger logging.Logger, topic string) *client.Config {
	clientConf := client.NewConfig(logger)
	if len(config.Addrs) > 0 {
		clientConf.SetBrokers(config.Addrs...)
	} else {
		clientConf.SetBrokers(mux.DefAddress)
	}
	clientConf.SetGroup(p.ServiceLabel.GetAgentLabel())
	clientConf.SetRecvMessageChan(p.subscription)
	clientConf.SetInitialOffset(sarama.OffsetNewest)
	clientConf.SetTopics(topic)
	return clientConf
}
