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
	asyncSubscription   chan (messaging.ProtoMessage)
	asyncSuccessChannel chan (messaging.ProtoMessage)
	asyncErrorChannel   chan (messaging.ProtoMessageErr)
	// Fields below are used to properly finish the example
	syncCaseDone  bool
	asyncCaseDone bool
	closeChannel  *chan struct{}
}

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	Kafka               *kafka.Plugin // injected
	local.PluginLogDeps               // injected
}

const (
	// Number of sync messages sent. Ensure that syncMessageCount >= syncMessageOffset
	syncMessageCount = 10
	// Partition sync messages are sent and watched on
	syncMessagePartition = 1
	// Offset for sync messages watcher
	syncMessageOffset = 5
	// Partiton async messages are sent and watched on
	asyncMessagePartition = 2
	// Offset for async messages watcher
	asyncMessageOffset = 0
)

// Topics
const (
	topic1 = "example-sync-clustered-topic"
	topic2 = "example-async-clustered-topic"
)

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() (err error) {
	// Create connection
	connection := plugin.Kafka.NewProtoManualConnection("example-proto-connection")

	// Create a synchronous and asynchronous publisher. In manual mode, every publisher has defined partition, where
	// the messages for given partition will be stored
	plugin.kafkaSyncPublisher, err = connection.NewSyncPublisherToPartition(topic1, syncMessagePartition)
	if err != nil {
		return err
	}
	// Async publisher requires two more channels to send success/error callback
	plugin.asyncSuccessChannel = make(chan messaging.ProtoMessage, 0)
	plugin.asyncErrorChannel = make(chan messaging.ProtoMessageErr, 0)
	plugin.kafkaAsyncPublisher, err = connection.NewAsyncPublisherToPartition(topic2, asyncMessagePartition,
		messaging.ToProtoMsgChan(plugin.asyncSuccessChannel), messaging.ToProtoMsgErrChan(plugin.asyncErrorChannel))
	if err != nil {
		return err
	}

	// Initialize sync watcher
	plugin.kafkaWatcher = plugin.Kafka.NewWatcher("example-watcher")

	// Prepare subscription channel. Relevant kafka messages are send to this channel so watcher can read it
	plugin.subscription = make(chan messaging.ProtoMessage)
	// The watcher is consuming messages on custom partition and offset. If there is a producer who stores message to
	// the partition and offset which is the same or newer, the message will be consumed
	err = plugin.kafkaWatcher.WatchPartition(messaging.ToProtoMsgChan(plugin.subscription), topic1,
		syncMessagePartition, syncMessageOffset)
	if err != nil {
		plugin.Log.Error(err)
	}

	// Prepare subscription channel. Relevant kafka messages are send to this channel so watcher can read it
	plugin.asyncSubscription = make(chan messaging.ProtoMessage)
	// The watcher is consuming messages on custom partition and offset. If there is a producer who stores message to
	// the partition and offset which is the same or newer, the message will be consumed
	err = plugin.kafkaWatcher.WatchPartition(messaging.ToProtoMsgChan(plugin.asyncSubscription), topic2,
		asyncMessagePartition, asyncMessageOffset)
	if err != nil {
		plugin.Log.Error(err)
	}

	plugin.Log.Info("Initialization of the custom plugin for the Kafka example is completed")

	// Run sync and async kafka consumers
	go plugin.syncEventHandler()
	go plugin.asyncEventHandler()

	// Run the producer
	go plugin.producer()

	// Verify results and close the example if successful
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
	safeclose.Close(plugin.asyncSuccessChannel)
	return nil
}

/*************
 * Producers *
 *************/

// Kafka Producer sends messages with desired topic and in manual mode also partition
func (plugin *ExamplePlugin) producer() {
	// Wait for the both event handlers to initialize
	time.Sleep(2 * time.Second)

	// Synchronous message with protobuf-encoded message
	enc := &etcd_example.EtcdExample{
		StringVal: "sync-dummy-message",
		Uint32Val: uint32(0),
		BoolVal:   true,
	}
	// Send several sync messages with offset 0,1,...
	plugin.Log.Info("Sending %v Kafka notifications (protobuf) ...", syncMessageCount)
	for i := 0; i < syncMessageCount; i++ {
		err := plugin.kafkaSyncPublisher.Put("proto-key", enc)
		if err != nil {
			plugin.Log.Errorf("Failed to sync-send a proto message, error %v", err)
		}
	}

	// Asynchronous message with protobuf encoded message. A success event is sent to the app asynchronously
	// on an event channel when the message has been successfully sent to Kafka. An error message is sent to
	// the app asynchronously if the message could not be sent (see also asyncEventHandler)
	plugin.Log.Info("Sending async Kafka notification (protobuf)")
	err := plugin.kafkaAsyncPublisher.Put("async-proto-key", enc)
	if err != nil {
		plugin.Log.Errorf("Failed to async-send a proto message, error %v", err)
	} else {
		plugin.Log.Info("Async proto message sent")
	}
}

/*************
 * Consumers *
 *************/

// Kafka consumer is subscribed to channel with specific topic, partition and offset. If producer sends a message with
// correct parameters, consumer will receive it
func (plugin *ExamplePlugin) syncEventHandler() {
	plugin.Log.Info("Started Kafka sync event handler...")

	// Producer sends several messages (set in syncMessageCount). Consumer should receive only messages from desired
	// partition and offset
	messageCounter := 0
	for message := range plugin.subscription {
		plugin.Log.Infof("Received sync Kafka Message, topic '%s', partition '%v', offset '%v', key: '%s', ",
			message.GetTopic(), message.GetPartition(), message.GetOffset(), message.GetKey())
		messageCounter++
		if message.GetPartition() != syncMessagePartition {
			plugin.Log.Errorf("Received sync message with unexpected partition: %v", message.GetOffset())
		}
		if message.GetOffset() < syncMessageOffset {
			plugin.Log.Errorf("Received sync message with unexpected offset: %v", message.GetOffset())
		}
		// For example purpose: let it know that this part of the example is done
		if messageCounter == (syncMessageCount - syncMessageOffset) {
			plugin.syncCaseDone = true
		}
	}

}

// Kafka consumer is subscribed to channel with specific topic, partition and offset. If producer sends a message with
// correct parameters, consumer will receive it
func (plugin *ExamplePlugin) asyncEventHandler() {
	plugin.Log.Info("Started Kafka async event handler...")
	for {
		select {
		// Channel subscribed with watcher
		case message := <-plugin.asyncSubscription:
			plugin.Log.Infof("Received async Kafka Message, topic '%s', partition '%v', offset '%v', key: '%s', ",
				message.GetTopic(), message.GetPartition(), message.GetOffset(), message.GetKey())
			if message.GetPartition() != asyncMessagePartition {
				plugin.Log.Errorf("Received async message with unexpected partition: %v", message.GetOffset())
			}
			if message.GetOffset() < asyncMessageOffset {
				plugin.Log.Errorf("Received async message with unexpected offset: %v", message.GetOffset())
			}
			// For example purpose: let it know that this part of the example is done
			plugin.asyncCaseDone = true
		// Success callback channel
		case message := <-plugin.asyncSuccessChannel:
			plugin.Log.Infof("Async message successfully delivered, topic '%s', partition '%v', offset '%v', key: '%s', ",
				message.GetTopic(), message.GetPartition(), message.GetOffset(), message.GetKey())
		// Error callback channel
		case err := <-plugin.asyncErrorChannel:
			plugin.Log.Errorf("Failed to publish async message, %v", err)
		}
	}
}
