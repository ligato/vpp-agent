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

	// Kafka plugin is using two multiplexers. The first one is using 'hash' (default) partitioner. The second mux
	// uses manual partitioner which allows to send a message to specified partition and watching to desired partition/offset
	muxHash   *mux.Multiplexer
	muxManual *mux.Multiplexer

	consumer *client.Consumer
	disabled bool
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	local.PluginInfraDeps //inject
}

// FromExistingMux is used mainly for testing purposes.
func FromExistingMux(mux *mux.Multiplexer) *Plugin {
	return &Plugin{muxHash: mux}
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
		p.disabled = true
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

	// Initialize both multiplexers to allow both, dynamic and manual mode
	if p.muxHash == nil {
		p.muxHash, err = mux.InitMultiplexerWithConfig(config, p.ServiceLabel.GetAgentLabel(), client.Hash, p.Log)
		if err != nil {
			return err
		}
		p.Log.Debug("Default multiplexer initialized")
	}
	if p.muxManual == nil {
		p.muxManual, err = mux.InitMultiplexerWithConfig(config, p.ServiceLabel.GetAgentLabel(), client.Manual, p.Log)
		if err != nil {
			return err
		}
		p.Log.Debug("Manual multiplexer initialized")
	}

	return err
}

// AfterInit is called in the second phase of the initialization. The kafka multiplexerNewWatcher
// is started, all consumers have to be subscribed until this phase.
func (p *Plugin) AfterInit() error {
	if p.muxHash == nil {
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

	return p.muxHash.Start()
}

// Close is called at plugin cleanup phase.
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.consumer, p.muxHash, p.muxManual)
	return err
}

// NewBytesConnection returns a new instance of a connection to access kafka brokers. The connection allows to create
// new kafka providers/consumers on multiplexer with hash partitioner.
func (p *Plugin) NewBytesConnection(name string) *mux.BytesConnection {
	return p.muxHash.NewBytesConnection(name)
}

// NewBytesConnectionToPartition returns a new instance of a connection to access kafka brokers. The connection allows to create
// new kafka providers/consumers on multiplexer with manual partitioner which allows to send messages to specific partition
// in kafka cluster and watch on partition/offset.
func (p *Plugin) NewBytesConnectionToPartition(name string) *mux.BytesConnection {
	return p.muxManual.NewBytesConnection(name)
}

// NewProtoConnection returns a new instance of a connection to access kafka brokers. The connection allows to create
// new kafka providers/consumers on multiplexer with hash partitioner.The connection uses proto-modelled messages.
func (p *Plugin) NewProtoConnection(name string) mux.Connection {
	return p.muxHash.NewProtoConnection(name, &keyval.SerializerJSON{})
}

// NewProtoManualConnection returns a new instance of a connection to access kafka brokers. The connection allows to create
// new kafka providers/consumers on multiplexer with manual partitioner which allows to send messages to specific partition
// in kafka cluster and watch on partition/offset. The connection uses proto-modelled messages.
func (p *Plugin) NewProtoManualConnection(name string) mux.ManualConnection {
	return p.muxManual.NewProtoManualConnection(name, &keyval.SerializerJSON{})
}

// NewSyncPublisher creates a publisher that allows to publish messages using synchronous API. The publisher creates
// new proto connection on multiplexer with default partitioner.
func (p *Plugin) NewSyncPublisher(connectionName string, topic string) (messaging.ProtoPublisher, error) {
	return p.NewProtoConnection(connectionName).NewSyncPublisher(topic)
}

// NewSyncPublisherToPartition creates a publisher that allows to publish messages to custom partition using synchronous API.
// The publisher creates new proto connection on multiplexer with manual partitioner.
func (p *Plugin) NewSyncPublisherToPartition(connectionName string, topic string, partition int32) (messaging.ProtoPublisher, error) {
	return p.NewProtoManualConnection(connectionName).NewSyncPublisherToPartition(topic, partition)
}

// NewAsyncPublisher creates a publisher that allows to publish messages using asynchronous API. The publisher creates
// new proto connection on multiplexer with default partitioner.
func (p *Plugin) NewAsyncPublisher(connectionName string, topic string, successClb func(messaging.ProtoMessage), errorClb func(messaging.ProtoMessageErr)) (messaging.ProtoPublisher, error) {
	return p.NewProtoConnection(connectionName).NewAsyncPublisher(topic, successClb, errorClb)
}

// NewAsyncPublisherToPartition creates a publisher that allows to publish messages to custom partition using asynchronous API.
// The publisher creates new proto connection on multiplexer with manual partitioner.
func (p *Plugin) NewAsyncPublisherToPartition(connectionName string,topic string, partition int32, successClb func(messaging.ProtoMessage), errorClb func(messaging.ProtoMessageErr)) (messaging.ProtoPublisher, error) {
	return p.NewProtoManualConnection(connectionName).NewAsyncPublisherToPartition(topic, partition, successClb, errorClb)
}

// NewWatcher creates a watcher that allows to start/stop consuming of messaging published to given topics.
func (p *Plugin) NewWatcher(name string) messaging.ProtoWatcher {
	return p.NewProtoConnection(name)
}

// Disabled if the plugin config was not found
func (p *Plugin) Disabled() (disabled bool) {
	return p.disabled
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
