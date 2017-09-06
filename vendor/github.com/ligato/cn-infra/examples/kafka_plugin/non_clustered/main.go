package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/flavors/local"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/messaging/kafka"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
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
	// Init close channel used to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combination of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{closeChan: &exampleFinished}
	agent := core.NewAgent(log.StandardLogger(), 15*time.Second, append(flavor.Plugins())...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

/**********
 * Flavor *
 **********/

// Kafka flag to load config
func init() {
	flag.String("kafka-config", "kafka.conf",
		"Location of the kafka configuration file")
}

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
	*local.FlavorLocal
	// Kafka plugin
	Kafka kafka.Plugin
	// Example plugin
	KafkaExample ExamplePlugin
	// For example purposes, use channel when the example is finished
	closeChan *chan struct{}
}

// Inject sets object references
func (ef *ExampleFlavor) Inject() (allReadyInjected bool) {
	// Init local flavor
	if ef.FlavorLocal == nil {
		ef.FlavorLocal = &local.FlavorLocal{}
	}
	ef.FlavorLocal.Inject()
	// Init kafka
	ef.Kafka.Deps.PluginInfraDeps = *ef.FlavorLocal.InfraDeps("kafka")
	// Inject kafka to example plugin
	ef.KafkaExample.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("kafka-example")
	ef.KafkaExample.Kafka = &ef.Kafka
	ef.KafkaExample.closeChannel = ef.closeChan

	return true
}

// Plugins combines all Plugins in flavor to the list
func (ef *ExampleFlavor) Plugins() []*core.NamedPlugin {
	ef.Inject()
	return core.ListPluginsInFlavor(ef)
}

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent. The Kafka
// ConsumerHandle is required to read messages from a topic, and PluginConnection is needed to start consuming on
// the topic
type ExamplePlugin struct {
	Deps // plugin dependencies are injected

	subscription        chan (messaging.ProtoMessage)
	kafkaSyncPublisher  messaging.ProtoPublisher
	kafkaAsyncPublisher messaging.ProtoPublisher
	kafkaWatcher        messaging.ProtoWatcher
	// Successfully published kafka message is sent through the message channel, error channel otherwise
	asyncMessageChannel chan (messaging.ProtoMessage)
	asyncErrorChannel   chan (messaging.ProtoMessageErr)
	// Fields below are used to properly finish the example
	syncCaseDone  bool
	asyncCaseDone bool
	closeChannel  *chan struct{}
}

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	Kafka               messaging.Mux // injected
	local.PluginLogDeps               // injected
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
	plugin.kafkaAsyncPublisher = plugin.Kafka.NewAsyncPublisher(topic, messaging.ToProtoMsgChan(plugin.asyncMessageChannel),
		messaging.ToProtoMsgErrChan(plugin.asyncErrorChannel))

	plugin.kafkaWatcher = plugin.Kafka.NewWatcher("example-plugin")

	// kafkaWatcher.Watch is called to start consuming a topic.
	plugin.subscription = make(chan messaging.ProtoMessage)
	err = plugin.kafkaWatcher.Watch(messaging.ToProtoMsgChan(plugin.subscription), topic)
	if err != nil {
		plugin.Log.Error(err)
	}

	plugin.Log.Info("Initialization of the custom plugin for the Kafka example is completed")

	// Run sync and async kafka consumers
	go plugin.syncEventHandler()
	go plugin.asyncEventHandler()

	// Run the producer to send notifications
	go plugin.producer()

	// Verify results and close the example
	go plugin.closeExample()

	return err
}

func (plugin *ExamplePlugin) closeExample() {
	for {
		if plugin.syncCaseDone && plugin.asyncCaseDone {
			plugin.Log.Info("kafka example finished, sending shutdown ...")
			*plugin.closeChannel <- struct{}{}
			break
		}
	}
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	safeclose.Close(plugin.subscription)
	safeclose.Close(plugin.asyncErrorChannel)
	safeclose.Close(plugin.asyncMessageChannel)
	return nil
}

/***********************
 * Kafka Example calls *
 ***********************/

// Send Kafka notifications
func (plugin *ExamplePlugin) producer() {
	// Wait for the both event handlers to initialize
	time.Sleep(2 * time.Second)

	// Synchronous message with protobuf-encoded message
	enc := &etcd_example.EtcdExample{
		StringVal: "value",
		Uint32Val: uint32(0),
		BoolVal:   true,
	}
	plugin.Log.Info("Sending Kafka notification (protobuf)")
	err := plugin.kafkaSyncPublisher.Put("proto-key", enc)
	if err != nil {
		plugin.Log.Errorf("Failed to sync-send a proto message, error %v", err)
	} else {
		plugin.Log.Debugf("Sent sync proto message.")
	}

	// Asynchronous message with protobuf encoded message. A success event is sent to the app asynchronously
	// on an event channel when the message has been successfully sent to Kafka. An error message is sent to
	// the app asynchronously if the message could not be sent.
	plugin.Log.Info("Sending async Kafka notification (protobuf)")
	plugin.kafkaAsyncPublisher.Put("async-proto-key", enc)
}

/*************
 * Consumers *
 *************/

// Kafka consumer is subscribed to channel with specific topic. If producer sends a message with the topic, consumer will
// receive it
func (plugin *ExamplePlugin) syncEventHandler() {
	plugin.Log.Info("Started Kafka event handler...")

	// Watch on message channel for sync kafka events
	for message := range plugin.subscription {
		plugin.Log.Infof("Received Kafka Message, topic '%s', partition '%v', key: '%s', ",
			message.GetTopic(), message.GetPartition(), message.GetKey())
		// Let it know that this part of the example is done
		plugin.syncCaseDone = true
	}
}

// asyncEventHandler shows handling of asynchronous events coming from the Kafka client
func (plugin *ExamplePlugin) asyncEventHandler() {
	plugin.Log.Info("Started Kafka async event handler...")
	for {
		select {
		case message := <-plugin.asyncMessageChannel:
			plugin.Log.Infof("Received async Kafka Message, topic '%s', partition '%v', key: '%s', ",
				message.GetTopic(), message.GetPartition(), message.GetKey())
			// Let it know that this part of the example is done
			plugin.asyncCaseDone = true
		case err := <-plugin.asyncErrorChannel:
			plugin.Log.Errorf("Failed to publish async message, %v", err)
		}
	}
}
