package main

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/utils/safeclose"
)

//********************************************************************
// This example shows how to use the Agent's Kafka APIs to perform
// synchronous/asynchronous calls and how to watch on these events.
//********************************************************************

func main() {
	// Init close channel used to stop the example
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

	subscription       chan (messaging.ProtoMessage)
	kafkaSyncPublisher messaging.ProtoPublisher
	kafkaWatcher       messaging.ProtoPartitionWatcher
	// Fields below are used to properly finish the example.
	initialized  bool // auxiliary flag that marks plugin as initialized
	messagesSent bool
	syncCaseDone bool
	closeChannel *chan struct{}
}

const (
	// Number of sync messages sent. Ensure that syncMessageCount >= syncMessageOffset
	syncMessageCount = 10
	// Partition sync messages are sent and watched on
	syncMessagePartition = 1
	// Offset for sync messages watcher
	syncMessageOffset = 0
)

// Topics
const (
	topic1     = "example-sync-topic"
	connection = "example-proto-connection"
)

// Init initializes and starts producers and consumers.
func (plugin *ExamplePlugin) Init() (err error) {
	// Create a synchronous publisher.
	// In the manual mode, every publisher has selected its target partition.
	plugin.kafkaSyncPublisher, err = plugin.Kafka.NewSyncPublisherToPartition(connection, topic1, syncMessagePartition)
	if err != nil {
		return err
	}

	// Prepare subscription channel. Relevant kafka messages are send to this
	// channel so that the watcher can read it.
	plugin.subscription = make(chan messaging.ProtoMessage)

	plugin.Log.Info("Initialization of the custom plugin for the Kafka example is completed")

	// Run sync and async kafka consumers.
	go plugin.syncEventHandler()

	// Run the producer.
	go plugin.producer()

	// Verify results and close the example if successful.
	go plugin.closeExample()

	plugin.initialized = true

	return err
}

func (plugin *ExamplePlugin) closeExample() {
	for {
		if plugin.syncCaseDone && plugin.messagesSent {
			time.Sleep(2 * time.Second)
			plugin.Log.Info("kafka example finished, sending shutdown ...")
			*plugin.closeChannel <- struct{}{}
			break
		}
	}
}

// Close closes the subscription and the channels used by the async producer.
func (plugin *ExamplePlugin) Close() error {
	safeclose.Close(plugin.subscription)
	return nil
}

/*************
 * Producers *
 *************/

// producer sends messages to a desired topic and in the manual mode also
// to a specified partition.
func (plugin *ExamplePlugin) producer() {
	// Wait for the both event handlers to initialize.
	time.Sleep(2 * time.Second)

	// Synchronous message with protobuf-encoded data.
	enc := &etcdexample.EtcdExample{
		StringVal: "sync-dummy-message",
		Uint32Val: uint32(0),
		BoolVal:   true,
	}
	// Send several sync messages with offsets 0,1,...
	plugin.Log.Infof("Sending %v Kafka notifications (protobuf) ...", syncMessageCount)
	for i := 0; i < syncMessageCount; i++ {
		err := plugin.kafkaSyncPublisher.Put("proto-key", enc)
		if err != nil {
			plugin.Log.Errorf("Failed to sync-send a proto message, error %v", err)
		}
	}

	plugin.messagesSent = true
}

/*************
 * Consumers *
 *************/

// syncEventHandler is a Kafka consumer synchronously processing events from
// a channel associated with a specific topic, partition and a starting offset.
// If a producer sends a message matching this destination criteria, the consumer
// will receive it.
func (plugin *ExamplePlugin) syncEventHandler() {
	plugin.Log.Info("Started Kafka sync event handler...")

	for !plugin.initialized {
		continue
	}

	time.Sleep(1 * time.Second)

	// Initialize sync watcher.
	plugin.kafkaWatcher = plugin.Kafka.NewPartitionWatcher("example-part-watcher")

	// The watcher is consuming messages on a custom partition and an offset.
	// If there is a producer who stores message to the same partition under
	// the same or a newer offset, the message will be consumed.
	err := plugin.kafkaWatcher.WatchPartition(messaging.ToProtoMsgChan(plugin.subscription), topic1,
		syncMessagePartition, syncMessageOffset)
	if err != nil {
		plugin.Log.Error(err)
	}

	// Producer sends several messages (set in syncMessageCount).
	// Consumer should receive only messages from desired partition and offset.
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
		if messageCounter >= (syncMessageCount - syncMessageOffset) {
			plugin.syncCaseDone = true
		}
	}

}
