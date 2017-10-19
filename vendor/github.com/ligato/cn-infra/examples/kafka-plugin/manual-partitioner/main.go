package main

import (
	"time"

	"fmt"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/messaging/kafka/mux"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"os"
	"strconv"
)

//********************************************************************
// This example shows how to use the Agent's Kafka APIs to perform
// synchronous/asynchronous calls and how to watch on these events.
//********************************************************************

var (
	// Flags used to read the input arguments. Applies for both, sync and async message
	offsetMsg    = flag.String("offsetMsg", os.Getenv("KAFKA_OFFSET"), "Use 'latest', 'oldest' or exact number of message offset")
	messageCount = flag.String("messageCount", os.Getenv("MSG_COUNT"), "Number of messages which will be send. Set to '0' to just watch")
)

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

	subscription        chan (messaging.ProtoMessage)
	kafkaSyncPublisher  messaging.ProtoPublisher
	kafkaAsyncPublisher messaging.ProtoPublisher
	kafkaWatcher        messaging.ProtoPartitionWatcher
	// Successfully published kafka message is sent through the message channel.
	// In case of a failure it sent through the error channel.
	asyncSubscription   chan (messaging.ProtoMessage)
	asyncSuccessChannel chan (messaging.ProtoMessage)
	asyncErrorChannel   chan (messaging.ProtoMessageErr)
	// Fields below are used to properly finish the example.
	messagesSent bool
	asyncSuccess bool
	closeChannel *chan struct{}
}

const (
	// Partition sync messages are sent and watched on
	syncMessagePartition = 1
	// Partiton async messages are sent and watched on
	asyncMessagePartition = 2
)

// These vars are applied for both, sync and async case
var (
	// Offset for sync messages watcher
	messageOffset int64
	// How many messages will be sent
	messageCountNum = 10
)

// Consts
const (
	topic1     = "example-sync-topic"
	topic2     = "example-async-topic"
	connection = "example-proto-connection"
	subscriber = "example-part-watcher"
)

// Init initializes and starts producers and consumers.
func (plugin *ExamplePlugin) Init() (err error) {
	// handle flags
	flag.Parse()
	// sync/async offset flag
	if *offsetMsg != "" {
		messageOffset, err = resolveOffset(*offsetMsg)
		if err != nil {
			return fmt.Errorf("incorrect sync offset value %v", *offsetMsg)
		}
	} else {
		plugin.Log.Info("offset arg not set, using default value")
	}
	// message count flag
	if *messageCount != "" {
		messageCountNum, err = resolveMsgCount(*messageCount)
		if err != nil {
			return fmt.Errorf("'messageCount' has to be a number, not %v", *messageCount)
		}
		if messageCountNum < 0 {
			plugin.Log.Warnf("'messageCount' %v is not a positive number, defaulting to 0")
			messageCountNum = 0
		}
	} else {
		plugin.Log.Info("messageCount arg not set, using default value")
	}

	plugin.Log.Infof("Offset: %v, message count: %v", messageOffset, messageCountNum)

	// Create a synchronous and asynchronous publisher.
	// In the manual mode, every publisher has selected its target partition.
	plugin.kafkaSyncPublisher, err = plugin.Kafka.NewSyncPublisherToPartition(connection, topic1, syncMessagePartition)
	if err != nil {
		return err
	}
	// Async publisher requires two more channels to send success/error callback.
	plugin.asyncSuccessChannel = make(chan messaging.ProtoMessage, 0)
	plugin.asyncErrorChannel = make(chan messaging.ProtoMessageErr, 0)
	plugin.kafkaAsyncPublisher, err = plugin.Kafka.NewAsyncPublisherToPartition(connection, topic2, asyncMessagePartition,
		messaging.ToProtoMsgChan(plugin.asyncSuccessChannel), messaging.ToProtoMsgErrChan(plugin.asyncErrorChannel))
	if err != nil {
		return err
	}

	// Initialize sync watcher.
	plugin.kafkaWatcher = plugin.Kafka.NewPartitionWatcher(subscriber)

	// Prepare subscription channel. Relevant kafka messages are send to this
	// channel so that the watcher can read it.
	plugin.subscription = make(chan messaging.ProtoMessage)
	// The watcher is consuming messages on a custom partition and an offset.
	// If there is a producer who stores message to the same partition under
	// the same or a newer offset, the message will be consumed.
	err = plugin.kafkaWatcher.WatchPartition(messaging.ToProtoMsgChan(plugin.subscription), topic1,
		syncMessagePartition, messageOffset)
	if err != nil {
		plugin.Log.Error(err)
	}

	// Prepare subscription channel. Relevant kafka messages are send to this
	// channel so that the watcher can read it
	plugin.asyncSubscription = make(chan messaging.ProtoMessage)
	// The watcher is consuming messages on custom partition and offset.
	// If there is a producer who stores message to the same partition under
	// the same or a newer offset, the message will be consumed.
	err = plugin.kafkaWatcher.WatchPartition(messaging.ToProtoMsgChan(plugin.asyncSubscription), topic2,
		asyncMessagePartition, messageOffset)
	if err != nil {
		plugin.Log.Error(err)
	}

	plugin.Log.Info("Initialization of the custom plugin for the Kafka example is completed")

	// Run sync and async kafka consumers.
	go plugin.syncEventHandler()
	go plugin.asyncEventHandler()

	// Run the producer.
	go plugin.producer()

	// Verify results and close the example if successful.
	go plugin.closeExample()

	return err
}

func (plugin *ExamplePlugin) closeExample() {
	for {
		if plugin.messagesSent && plugin.asyncSuccess {
			time.Sleep(2 * time.Second)
			err := plugin.kafkaWatcher.StopWatchPartition(topic1, syncMessagePartition, messageOffset)
			if err != nil {
				plugin.Log.Errorf("Error while stopping watcher: %v", err)
			} else {
				plugin.Log.Info("Sync watcher closed")
			}
			err = plugin.kafkaWatcher.StopWatchPartition(topic2, asyncMessagePartition, messageOffset)
			if err != nil {
				plugin.Log.Errorf("Error while stopping watcher: %v", err)
			} else {
				plugin.Log.Info("Async watcher closed")
			}
			plugin.Log.Info("kafka example finished, sending shutdown ...")
			*plugin.closeChannel <- struct{}{}
			break
		}
	}
}

// Close closes the subscription and the channels used by the async producer.
func (plugin *ExamplePlugin) Close() error {
	safeclose.CloseAll(plugin.subscription, plugin.asyncErrorChannel, plugin.asyncSuccessChannel)
	return nil
}

/*************
 * Producers *
 *************/

// producer sends messages to a desired topic and in the manual mode also to a specified partition. Tho number of messages
// sent can be set with flag
func (plugin *ExamplePlugin) producer() {
	// Wait for the both event handlers to initialize.
	time.Sleep(2 * time.Second)

	// Synchronous message with protobuf-encoded data.
	enc := &etcdexample.EtcdExample{
		StringVal: "sync-dummy-message",
		Uint32Val: uint32(0),
		BoolVal:   true,
	}
	// Send several sync messages with offsets offsetLast+1, offsetLast+2,...
	plugin.Log.Infof("Sending %v sync Kafka notifications (protobuf) ...", messageCountNum)
	for i := 0; i < messageCountNum; i++ {
		err := plugin.kafkaSyncPublisher.Put("proto-key", enc)
		if err != nil {
			plugin.Log.Errorf("Failed to sync-send a proto message, error %v", err)
		}
	}

	// Send message with protobuf encoded data asynchronously.
	// Delivery status is propagated back to the application through
	// the configured pair of channels - one for the success events and one for
	// the errors.
	plugin.Log.Infof("Sending %v async Kafka notifications (protobuf) ...", messageCountNum)
	for i := 0; i < messageCountNum; i++ {
		err := plugin.kafkaAsyncPublisher.Put("async-proto-key", enc)
		if err != nil {
			plugin.Log.Errorf("Failed to async-send a proto message, error %v", err)
		}
	}

	// Mark that all messages were sent
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

	// Producer sends several messages (set in messageCount).
	// Consumer should receive only messages from desired partition and offset.
	receivedMessageCounter := 0
	for message := range plugin.subscription {
		plugin.Log.Infof("Received sync Kafka Message, topic '%s', partition '%v', offset '%v', key: '%s', ",
			message.GetTopic(), message.GetPartition(), message.GetOffset(), message.GetKey())
		// Note: mark the offset if required
		receivedMessageCounter++
		if message.GetPartition() != syncMessagePartition {
			plugin.Log.Errorf("Received sync message with unexpected partition: %v", message.GetOffset())
		}
		if message.GetOffset() < messageOffset {
			plugin.Log.Errorf("Received sync message with unexpected offset: %v", message.GetOffset())
		}
	}

}

// asyncEventHandler is a Kafka consumer asynchronously processing events from
// a channel associated with a specific topic, partition and a starting offset.
// If a producer sends a message matching this destination criteria, the consumer
// will receive it.
func (plugin *ExamplePlugin) asyncEventHandler() {
	plugin.Log.Info("Started Kafka async event handler...")
	asyncSuccessCounter := 0
	if messageCountNum == 0 {
		plugin.asyncSuccess = true
	}

	for {
		select {
		// Channel subscribed with watcher
		case message := <-plugin.asyncSubscription:
			plugin.Log.Infof("Received async Kafka Message, topic '%s', partition '%v', offset '%v', key: '%s', ",
				message.GetTopic(), message.GetPartition(), message.GetOffset(), message.GetKey())
			// Note: mark the offset if required
			if message.GetPartition() != asyncMessagePartition {
				plugin.Log.Errorf("Received async message with unexpected partition: %v", message.GetOffset())
			}
			if message.GetOffset() < messageOffset {
				plugin.Log.Errorf("Received async message with unexpected offset: %v", message.GetOffset())
			}
		// Success callback channel
		case message := <-plugin.asyncSuccessChannel:
			plugin.Log.Infof("Async message successfully delivered, topic '%s', partition '%v', offset '%v', key: '%s', ",
				message.GetTopic(), message.GetPartition(), message.GetOffset(), message.GetKey())
			// Note: mark the offset if required
			asyncSuccessCounter++
			if asyncSuccessCounter == messageCountNum {
				plugin.asyncSuccess = true
			}
		// Error callback channel
		case err := <-plugin.asyncErrorChannel:
			plugin.Log.Errorf("Failed to publish async message, %v", err)
		}
	}
}

func resolveOffset(offset string) (int64, error) {
	if offset == "latest" {
		return mux.OffsetNewest, nil
	} else if offset == "oldest" {
		return mux.OffsetOldest, nil
	} else {
		result, err := strconv.Atoi(offset)
		return int64(result), err
	}
}

func resolveMsgCount(count string) (int, error) {
	result, err := strconv.Atoi(count)
	return result, err
}
