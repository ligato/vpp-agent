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
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"github.com/ligato/cn-infra/messaging/kafka/mux"
	"github.com/ligato/cn-infra/utils/safeclose"
)

const topic = "status-check"

// Plugin provides API for interaction with kafka brokers.
type Plugin struct {
	Deps         // inject
	subscription chan (*client.ConsumerMessage)
	mx           *mux.Multiplexer
	consumer     *client.Consumer
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	local.PluginInfraDeps //inject
}

// FromExistingMux is used mainly for testing purposes.
func FromExistingMux(mux *mux.Multiplexer) *Plugin {
	return &Plugin{mx: mux}
}

// Init is called at plugin initialization.
func (p *Plugin) Init() (err error) {
	// Prepare topic and  subscription for status check client
	p.subscription = make(chan *client.ConsumerMessage)

	// Get config data
	config := &mux.Config{}
	found, err := p.PluginConfig.GetValue(config)
	if !found {
		p.Log.Info("kafka config not found ", p.PluginConfig.GetConfigName(), " - skip loading this plugin")
		return nil //skip loading the plugin
	}
	if err != nil {
		return err
	}
	clientConfig := p.getClientConfig(config, p.Log, topic)

	// Init consumer
	p.consumer, err = client.NewConsumer(clientConfig, nil)
	if err != nil {
		return err
	}

	if p.mx == nil {
		p.mx, err = mux.InitMultiplexerWithConfig(config, p.ServiceLabel.GetAgentLabel(), p.Log)
	}

	return err
}

// AfterInit is called in the second phase of initialization. The kafka multiplexer
// is started, all consumers have to be subscribed until this phase.
func (p *Plugin) AfterInit() error {
	if p.mx == nil {
		return nil
	}

	// Register for providing status reports (polling mode)
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, func() (statuscheck.PluginState, error) {
			// Method 'RefreshMetadata()' returns error if kafka server is unavailable
			err := p.consumer.Client.RefreshMetadata(topic)
			if err == nil {
				return statuscheck.OK, nil
			}
			p.Log.Errorf("Kafka server unavailable")
			return statuscheck.Error, err
		})
	} else {
		p.Log.Warnf("Unable to start status check for kafka")
	}

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

// NewSyncPublisher creates a publisher that allows to publish messages using synchronous API.
func (p *Plugin) NewSyncPublisher(topic string) messaging.ProtoPublisher {
	return p.NewProtoConnection("").NewSyncPublisher(topic)
}

// NewSyncPublisherToPartition creates a publisher that allows to publish messages to selected topic/partition using synchronous API .
func (p *Plugin) NewSyncPublisherToPartition(topic string, partition int32) messaging.ProtoPublisher {
	p.Log.Warn("Publishing to a partition not implemented yet")
	return p.NewProtoConnection("").NewSyncPublisher(topic)
}

// NewAsyncPublisher creates a publisher that allows to publish messages using asynchronous API.
func (p *Plugin) NewAsyncPublisher(topic string, successClb func(messaging.ProtoMessage), errorClb func(messaging.ProtoMessageErr)) messaging.ProtoPublisher {
	return p.NewProtoConnection("").NewAsyncPublisher(topic, successClb, errorClb)
}

// NewAsyncPublisherToPartition creates a publisher that allows to publish messages to selected topic/partition using asynchronous API.
func (p *Plugin) NewAsyncPublisherToPartition(topic string, partition int32, successClb func(messaging.ProtoMessage), errorClb func(messaging.ProtoMessageErr)) messaging.ProtoPublisher {
	p.Log.Warn("Publishing to a partition not implemented yet")
	return p.NewProtoConnection("").NewAsyncPublisher(topic, successClb, errorClb)
}

// NewWatcher creates a watcher that allows to start/stop consuming of messaging published to given topics.
func (p *Plugin) NewWatcher(name string) messaging.ProtoWatcher {
	return p.NewProtoConnection(name)
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
