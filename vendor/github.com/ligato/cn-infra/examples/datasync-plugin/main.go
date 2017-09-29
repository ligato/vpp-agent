package main

import (
	"strings"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/utils/safeclose"
	"golang.org/x/net/context"
)

// *************************************************************************
// This example demonstrates the usage of datasync API with etcdv3
// as the data store.
// ExamplePlugin spawns a data publisher and a data consumer (watcher)
// as two separate go routines.
// The publisher executes two operations on the same key: CREATE + UPDATE.
// The consumer is notified with each change and reports the events into
// the log.
// ************************************************************************/

func main() {
	// Init close channel used to stop the example.
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor
	// (combination of ExamplePlugin & cn-infra plugins).
	flavor := ExampleFlavor{closeChan: &exampleFinished}
	plugins := flavor.Plugins()
	agent := core.NewAgent(flavor.LogRegistry().NewLogger("core"), 15*time.Second, plugins...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// ExamplePlugin demonstrates the usage of datasync API.
type ExamplePlugin struct {
	Deps

	changeChannel chan datasync.ChangeEvent  // Channel used by the watcher for change events.
	resyncChannel chan datasync.ResyncEvent  // Channel used by the watcher for resync events.
	context       context.Context            // Used to cancel watching.
	watchDataReg  datasync.WatchRegistration // To subscribe on data change/resync events.
	// Fields below are used to properly finish the example.
	eventCounter uint8
	closeChannel *chan struct{}
}

// Init starts the consumer.
func (plugin *ExamplePlugin) Init() error {
	// Initialize plugin fields.
	plugin.resyncChannel = make(chan datasync.ResyncEvent)
	plugin.changeChannel = make(chan datasync.ChangeEvent)
	plugin.context = context.Background()

	// Start the consumer (ETCD watcher).
	go plugin.consumer()
	// Subscribe watcher to be able to watch on data changes and resync events.
	err := plugin.subscribeWatcher()
	if err != nil {
		return err
	}

	plugin.Log.Info("Initialization of the custom plugin for the datasync example is completed")

	return nil
}

// AfterInit starts the publisher and prepares for the shutdown.
func (plugin *ExamplePlugin) AfterInit() error {

	go plugin.etcdPublisher()

	go plugin.closeExample()

	return nil
}

// etcdPublisher creates a simple data, then demonstrates CREATE and UPDATE
// operations with ETCD.
func (plugin *ExamplePlugin) etcdPublisher() {
	// Wait for the consumer to initialize
	time.Sleep(3 * time.Second)
	plugin.Log.Print("KeyValPublisher started")

	// Convert data into the proto format.
	exampleData := plugin.buildData("string1", 0, true)

	// PUT: demonstrate how to use the Data Broker Put() API to store
	// a simple data structure into ETCD.
	label := etcdKeyPrefixLabel(plugin.ServiceLabel.GetAgentLabel(), "index")
	plugin.Log.Infof("Write data to %v", label)
	plugin.Publisher.Put(label, exampleData)

	// Prepare different set of data.
	plugin.Log.Infof("Update data at %v", label)
	exampleData = plugin.buildData("string2", 1, false)

	// UPDATE: demonstrate how use the Data Broker Put() API to change
	// an already stored data in ETCD.
	plugin.Publisher.Put(label, exampleData)
}

// consumer (watcher) is subscribed to watch on data store changes.
// Changes arrive via data change channel, get identified based on the key
// and printed into the log.
func (plugin *ExamplePlugin) consumer() {
	plugin.Log.Print("KeyValProtoWatcher started")
	for {
		select {
		// WATCH: demonstrate how to receive data change events.
		case dataChng := <-plugin.changeChannel:
			plugin.Log.Printf("Received event: %v", dataChng)
			// If event arrives, the key is extracted and used together with
			// the expected prefix to identify item.
			key := dataChng.GetKey()
			if strings.HasPrefix(key, etcdKeyPrefix(plugin.ServiceLabel.GetAgentLabel())) {
				var value, previousValue etcdexample.EtcdExample
				// The first return value is diff - boolean flag whether previous value exists or not
				err := dataChng.GetValue(&value)
				if err != nil {
					plugin.Log.Error(err)
				}
				diff, err := dataChng.GetPrevValue(&previousValue)
				if err != nil {
					plugin.Log.Error(err)
				}
				plugin.Log.Infof("Event arrived to etcd eventHandler, key %v, update: %v, change type: %v,",
					dataChng.GetKey(), diff, dataChng.GetChangeType())
				// Increase event counter (expecting two events).
				plugin.eventCounter++
			}
			// Here you would test for other event types with one if statement
			// for each key prefix:
			//
			// if strings.HasPrefix(key, etcd prefix) { ... }

		// Here you would also watch for resync events
		// (not published in this example):
		//
		// case resyncEvent := <-plugin.ResyncEvent:
		//   ...

		case <-plugin.context.Done():
			plugin.Log.Warnf("Stop watching events")
		}
	}
}

// subscribeWatcher subscribes for data change and data resync events.
// Events are delivered to the consumer via the selected channels.
// ETCD watcher adapter is used to perform the registration behind the scenes.
func (plugin *ExamplePlugin) subscribeWatcher() (err error) {
	prefix := etcdKeyPrefix(plugin.ServiceLabel.GetAgentLabel())
	plugin.Log.Infof("Prefix: %v", prefix)
	plugin.watchDataReg, err = plugin.Watcher.
		Watch("Example etcd plugin", plugin.changeChannel, plugin.resyncChannel, prefix)
	if err != nil {
		return err
	}

	plugin.Log.Info("KeyValProtoWatcher subscribed")

	return nil
}

func (plugin *ExamplePlugin) closeExample() {
	for {
		// Two events are expected for successful example completion.
		if plugin.eventCounter == 2 {
			// Close the watcher
			plugin.context.Done()
			plugin.Log.Infof("etcd/datasync example finished, sending shutdown ...")
			// Close the example
			*plugin.closeChannel <- struct{}{}
			break
		}
	}
}

// Close shutdowns both the publisher and the consumer.
// Channels used to propagate data resync and data change events are closed
// as well.
func (plugin *ExamplePlugin) Close() error {
	safeclose.CloseAll(plugin.Publisher, plugin.Watcher, plugin.resyncChannel, plugin.changeChannel)
	return nil
}

// Create simple ETCD data structure with provided data values.
func (plugin *ExamplePlugin) buildData(stringVal string, uint32Val uint32, boolVal bool) *etcdexample.EtcdExample {
	return &etcdexample.EtcdExample{
		StringVal: stringVal,
		Uint32Val: uint32Val,
		BoolVal:   boolVal,
	}
}

// The ETCD key prefix used for this example
func etcdKeyPrefix(agentLabel string) string {
	return "/vnf-agent/" + agentLabel + "/api/v1/example/db/simple/"
}

// The ETCD key (the key prefix + label)
func etcdKeyPrefixLabel(agentLabel string, index string) string {
	return etcdKeyPrefix(agentLabel) + index
}
