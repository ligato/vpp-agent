package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/flavors/etcdkafka"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/utils/safeclose"
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

	flavor := etcdkafka.Flavor{}

	// Example plugin (Kafka)
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{Kafka: &flavor.Kafka}}

	// Create new agent
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(flavor.Plugins(), examplePlugin)...)

	// End when the kafka example is finished
	go closeExample("kafka example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(10 * time.Second)
	log.StandardLogger().Info(message)
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
	Kafka               messaging.Mux
	subscription        chan (messaging.ProtoMessage)
	kafkaSyncPublisher  messaging.ProtoPublisher
	kafkaAsyncPublisher messaging.ProtoPublisher
	kafkaWatcher        messaging.ProtoWatcher
	// Successfully published kafka message is sent through the message channel, error channel otherwise
	asyncMessageChannel chan (messaging.ProtoMessage)
	asyncErrorChannel   chan (messaging.ProtoMessageErr)
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	topic := "example-topic"
	// Init channels required for async handler
	plugin.asyncMessageChannel = make(chan messaging.ProtoMessage, 0)
	plugin.asyncErrorChannel = make(chan messaging.ProtoMessageErr, 0)

	// Create a synchronous publisher for the selected topic.
	plugin.kafkaSyncPublisher = plugin.Kafka.NewSyncPublisher(topic)

	// Create an asynchronous publisher for the selected topic.
	plugin.kafkaAsyncPublisher = plugin.Kafka.NewAsyncPublisher(topic, messaging.ToProtoMsgChan(plugin.asyncMessageChannel), messaging.ToProtoMsgErrChan(plugin.asyncErrorChannel))

	plugin.kafkaWatcher = plugin.Kafka.NewWatcher("example-plugin")

	// ConsumePartition is called to start consuming a topic/partition.
	plugin.subscription = make(chan messaging.ProtoMessage)
	err = plugin.kafkaWatcher.Watch(messaging.ToProtoMsgChan(plugin.subscription), topic)
	if err != nil {
		log.StandardLogger().Error(err)
	}

	log.StandardLogger().Info("Initialization of the custom plugin for the Kafka example is completed")

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

	// Synchronous message with protobuf-encoded message
	enc := &etcd_example.EtcdExample{
		StringVal: "value",
		Uint32Val: uint32(0),
		BoolVal:   true,
	}
	log.StandardLogger().Info("Sending Kafka notification (protobuf)")
	err := plugin.kafkaSyncPublisher.Put("proto-key", enc)
	if err != nil {
		log.StandardLogger().Errorf("Failed to sync-send a proto message, error %v", err)
	} else {
		log.StandardLogger().Debugf("Sent sync proto message.")
	}

	// Asynchronous message with protobuf encoded message. A success event is sent to the app asynchronously
	// on an event channel when the message has been successfully sent to Kafka. An error message is sent to
	// the app asynchronously if the message could not be sent.
	log.StandardLogger().Info("Sending async Kafka notification (protobuf)")
	plugin.kafkaAsyncPublisher.Put("async-proto-key", enc)
}

/************
 * Consumer *
 ************/

// Kafka consumer is subscribed to channel with specific topic. If producer sends a message with the topic, consumer will
// receive it
func (plugin *ExamplePlugin) syncEventHandler() {
	log.StandardLogger().Info("Started Kafka event handler...")

	// Watch on message channel for sync kafka events
	for message := range plugin.subscription {
		log.StandardLogger().Infof("Received Kafka Message, topic '%s', key: '%s', ", message.GetTopic(), message.GetKey())
	}
}

// asyncEventHandler shows handling of asynchronous events coming from the Kafka client
func (plugin *ExamplePlugin) asyncEventHandler() {
	log.StandardLogger().Info("Started Kafka async event handler...")
	for {
		select {
		case message := <-plugin.asyncMessageChannel:
			log.StandardLogger().Infof("Received async Kafka Message, topic '%s', key: '%s', ", message.GetTopic(), message.GetKey())
		case err := <-plugin.asyncErrorChannel:
			log.StandardLogger().Errorf("Failed to publish async message, %v", err)
		}
	}
}
