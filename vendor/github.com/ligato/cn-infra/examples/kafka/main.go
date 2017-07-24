package main

import (
	"encoding/json"
	"fmt"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/examples/simple-agent/generic"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/messaging/kafka"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"github.com/ligato/cn-infra/messaging/kafka/mux"
	"github.com/ligato/cn-infra/utils/safeclose"
	"time"
)

//********************************************************************
// The following functions show how to use the Agent's Kafka APIs
// and perform synchronous/asynchronous call and how to watch on
// these events
//********************************************************************

/********
 * Main *
 ********/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log) and example plugin which demonstrates Kafka functionality.
func main() {
	// Init close channel to stop the example
	closeChannel := make(chan struct{}, 1)

	flavour := generic.Flavour{}

	// Example plugin (Kafka)
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{Kafka: &flavour.Kafka}}

	// Create new agent
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(flavour.Plugins(), examplePlugin)...)

	// End when the kafka example is finished
	go closeExample("kafka example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(10 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
}

/**********************
 * Example plugin API *
 **********************/

// PluginID of the custom Kafka plugin
const PluginID core.PluginName = "example-plugin"

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent. The Kafka
// ConsumerHandle is required to read messages from a topic, and PluginConnection is needed to start consuming on
// the topic
type ExamplePlugin struct {
	Kafka          kafka.Mux
	subscription   chan (*client.ConsumerMessage)
	kafkaByteConn  *mux.Connection
	kafkaProtoConn *mux.ProtoConnection
	// Successfully published kafka message is sent through the message channel, error channel otherwise
	asyncMessageChannel chan (*client.ProducerMessage)
	asyncErrorChannel   chan (*client.ProducerError)
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	// Create new kafka connection. The connection allows to consume topic/partition and to publish
	// messages in plugin
	plugin.kafkaByteConn = plugin.Kafka.NewConnection("example-plugin")

	// Create a new kafka connection that allows easily process proto-modelled messages.
	plugin.kafkaProtoConn = plugin.Kafka.NewProtoConnection("example-plugin-proto")

	// ConsumePartition is called to start consuming a topic/partition.
	topic := "example-topic"
	plugin.subscription = make(chan *client.ConsumerMessage)
	err = plugin.kafkaByteConn.ConsumeTopic(plugin.subscription, topic)
	if err != nil {
		log.Error(err)
	}
	// Init channels required for async handler
	plugin.asyncMessageChannel = make(chan *client.ProducerMessage, 0)
	plugin.asyncErrorChannel = make(chan *client.ProducerError, 0)

	log.Info("Initialization of the custom plugin for the Kafka example is completed")

	// Run sync and async kafka consumers
	go plugin.syncEventHandler()
	go plugin.asyncEventHandler()

	// Run the producer to send notifications
	go plugin.producer()

	return err
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	safeclose.Close(plugin.subscription)
	safeclose.Close(plugin.asyncErrorChannel)
	safeclose.Close(plugin.asyncMessageChannel)
	return nil
}

/**************
 * Kafka Call *
 **************/

// Send Kafka notifications
func (plugin *ExamplePlugin) producer() {
	time.Sleep(4 * time.Second)
	exampleFile, _ := json.Marshal("{}")

	log.Info("Sending Kafka notification (string)")
	// Synchronous message with string encoded-message. The SendSyncMessage() call
	// returns when the message has been successfully sent to Kafka.
	offset, err := plugin.kafkaByteConn.SendSyncString("example-topic", fmt.Sprintf("%s", "string-key"),
		string(exampleFile))
	if err != nil {
		log.Errorf("Failed to sync-send a string message, error %v", err)
	} else {
		log.Debugf("Sent sync string message, offset: %v", offset)
	}

	// Synchronous message with protobuf-encoded message
	enc := &etcd_example.EtcdExample{
		StringVal: "value",
		Uint32Val: uint32(0),
		BoolVal:   true,
	}
	log.Info("Sending Kafka notification (protobuf)")
	offset, err = plugin.kafkaProtoConn.SendSyncMessage("example-topic", "proto-key", enc)
	if err != nil {
		log.Errorf("Failed to sync-send a proto message, error %v", err)
	} else {
		log.Debugf("Sent sync proto message, offset: %v", offset)
	}

	// Asynchronous message with protobuf encoded message. A success event is sent to the app asynchronously
	// on an event channel when the message has been successfully sent to Kafka. An error message is sent to
	// the app asynchronously if the message could not be sent.
	log.Info("Sending async Kafka notification (protobuf)")
	plugin.kafkaProtoConn.SendAsyncMessage("example-topic", "async-proto-key", enc, "metadata",
		plugin.asyncMessageChannel, plugin.asyncErrorChannel)
}

/************
 * Consumer *
 ************/

// Kafka consumer is subscribed to channel with specific topic. If producer sends a message with the topic, consumer will
// receive it
func (plugin *ExamplePlugin) syncEventHandler() {
	log.Info("Started Kafka event handler...")

	// Watch on message channel for sync kafka events
	for message := range plugin.subscription {
		log.Infof("Received Kafka Message, topic '%s', key: '%s', ", message.Topic, message.Key)
	}
}

// asyncEventHandler shows handling of asynchronous events coming from the Kafka client
func (plugin *ExamplePlugin) asyncEventHandler() {
	log.Info("Started Kafka async event handler...")
	for {
		select {
		case message := <-plugin.asyncMessageChannel:
			log.Infof("Received async Kafka Message, topic '%s', key: '%s', ", message.Topic, message.Key)
		case err := <-plugin.asyncErrorChannel:
			log.Errorf("Failed to publish async message, %v", err)
		}
	}
}
