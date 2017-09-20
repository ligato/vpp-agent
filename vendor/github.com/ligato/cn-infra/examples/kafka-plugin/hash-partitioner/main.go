package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/utils/safeclose"
)

//********************************************************************
// The following functions show how to use the Agent's Kafka APIs
// to perform synchronous/asynchronous calls and how to watch on
// these events.
//********************************************************************

func main() {
	// Init close channel used to stop the example.
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor
	// (combination of ExamplePlugin & reused cn-infra plugins).
	flavor := ExampleFlavor{closeChan: &exampleFinished}
	plugins := flavor.Plugins()
	agent := core.NewAgent(flavor.LogRegistry().NewLogger("core"), 15*time.Second, plugins...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExamplePlugin demonstrates the use of Kafka plugin API from another plugin.
// The Kafka ConsumerHandle is required to read messages from a topic
// and PluginConnection is needed to start consuming on that topic.
type ExamplePlugin struct {
	Deps // plugin dependencies are injected

	subscription        chan (messaging.ProtoMessage)
	kafkaSyncPublisher  messaging.ProtoPublisher
	kafkaAsyncPublisher messaging.ProtoPublisher
	kafkaWatcher        messaging.ProtoWatcher
	// Successfully published kafka message is sent through the message channel.
	// In case of a failure it sent through the error channel.
	asyncMessageChannel chan (messaging.ProtoMessage)
	asyncErrorChannel   chan (messaging.ProtoMessageErr)
	// Fields below are used to properly finish the example.
	syncCaseDone  bool
	asyncCaseDone bool
	closeChannel  *chan struct{}
}

// Init initializes and starts producers and consumers.
func (plugin *ExamplePlugin) Init() (err error) {
	conn := "example-connection"
	topic := "example-topic"
	// Init channels required for async handler.
	plugin.asyncMessageChannel = make(chan messaging.ProtoMessage, 0)
	plugin.asyncErrorChannel = make(chan messaging.ProtoMessageErr, 0)

	// Create a synchronous publisher for the selected topic.
	plugin.kafkaSyncPublisher, err = plugin.Kafka.NewSyncPublisher(conn, topic)
	if err != nil {
		return err
	}

	// Create an asynchronous publisher for the selected topic.
	plugin.kafkaAsyncPublisher, err = plugin.Kafka.NewAsyncPublisher(conn, topic, messaging.ToProtoMsgChan(plugin.asyncMessageChannel),
		messaging.ToProtoMsgErrChan(plugin.asyncErrorChannel))
	if err != nil {
		return err
	}

	plugin.kafkaWatcher = plugin.Kafka.NewWatcher("example-plugin")

	// kafkaWatcher.Watch is called to start consuming a topic.
	plugin.subscription = make(chan messaging.ProtoMessage)
	err = plugin.kafkaWatcher.Watch(messaging.ToProtoMsgChan(plugin.subscription), topic)
	if err != nil {
		plugin.Log.Error(err)
	}

	plugin.Log.Info("Initialization of the custom plugin for the Kafka example is completed")

	// Run sync and async kafka consumers.
	go plugin.syncEventHandler()
	go plugin.asyncEventHandler()

	// Run the producer to send notifications.
	go plugin.producer()

	// Verify results and close the example.
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

// Close closes the subscription and the channels used by the async producer.
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

	// Synchronous message with protobuf-encoded data.
	enc := &etcdexample.EtcdExample{
		StringVal: "value",
		Uint32Val: uint32(0),
		BoolVal:   true,
	}
	plugin.Log.Info("Sending Kafka notification (protobuf)")
	err := plugin.kafkaSyncPublisher.Put("proto-key", enc)
	if err != nil {
		plugin.Log.Errorf("Failed to sync-send a proto message, error %v", err)
	} else {
		plugin.Log.Info("Sync proto message sent")
	}

	// Send message with protobuf encoded data asynchronously.
	// Delivery status is propagated back to the application through
	// the configured pair of channels - one for the success events and one for
	// the errors.
	plugin.Log.Info("Sending async Kafka notification (protobuf)")
	err = plugin.kafkaAsyncPublisher.Put("async-proto-key", enc)
	if err != nil {
		plugin.Log.Errorf("Failed to async-send a proto message, error %v", err)
	} else {
		plugin.Log.Info("Async proto message sent")
	}
}

/*************
 * Consumers *
 *************/

// syncEventHandler is a Kafka consumer synchronously processing events from
// a channel associated with a specific topic, partition and a starting offset.
// If a producer sends a message matching this destination criteria, the consumer
// will receive it.
func (plugin *ExamplePlugin) syncEventHandler() {
	plugin.Log.Info("Started Kafka event handler...")

	// Watch on message channel for sync kafka events
	for message := range plugin.subscription {
		plugin.Log.Infof("Received Kafka Message, topic '%s', partition '%v', offset '%v', key: '%s', ",
			message.GetTopic(), message.GetPartition(), message.GetOffset(), message.GetKey())
		// Let it know that this part of the example is done
		plugin.syncCaseDone = true
	}
}

// asyncEventHandler is a Kafka consumer asynchronously processing events from
// a channel associated with a specific topic, partition and a starting offset.
// If a producer sends a message matching this destination criteria, the consumer
// will receive it.
func (plugin *ExamplePlugin) asyncEventHandler() {
	plugin.Log.Info("Started Kafka async event handler...")
	for {
		select {
		case message := <-plugin.asyncMessageChannel:
			plugin.Log.Infof("Received async Kafka Message, topic '%s', partition '%v', offset '%v', key: '%s', ",
				message.GetTopic(), message.GetPartition(), message.GetOffset(), message.GetKey())
			// Let it know that this part of the example is done
			plugin.asyncCaseDone = true
		case err := <-plugin.asyncErrorChannel:
			plugin.Log.Errorf("Failed to publish async message, %v", err)
		}
	}
}
