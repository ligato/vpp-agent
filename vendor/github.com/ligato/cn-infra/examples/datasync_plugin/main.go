package main

import (
	"strings"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"golang.org/x/net/context"
)

// *************************************************************************
// This file contains examples of simple publisher operations
// (APIs) including an event handler (watcher).
//
// These functions are called from the REST API. Put() operations are
// done as single operations and as a part of the transaction
// ************************************************************************/

/********
 * Main *
 ********/

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with ExampleFlavor
func main() {
	log := logroot.StandardLogger()
	// Init close channel to stop the example
	exampleFinished := make(chan struct{}, 1)

	// Start Agent with ExampleFlavor (combination of ExamplePlugin & reused cn-infra plugins)
	flavor := ExampleFlavor{closeChan: &exampleFinished}
	agent := core.NewAgent(log, 15*time.Second, append(flavor.Plugins())...)
	core.EventLoopWithInterrupt(agent, exampleFinished)
}

/**********
 * Flavor *
 **********/

// ETCD flag to load config
func init() {
	flag.String("etcdv3-config", "etcd.conf",
		"Location of the Etcd configuration file")
}

// ExampleFlavor is a set of plugins required for the datasync example.
type ExampleFlavor struct {
	// Local flavor to access to Infra (logger, service label, status check)
	*local.FlavorLocal
	// Resync orchestrator
	ResyncOrch resync.Plugin
	// Etcd plugin
	ETCD etcdv3.Plugin
	// Etcd sync which manages and injects connection
	ETCDDataSync kvdbsync.Plugin
	// Example plugin
	DatasyncExample ExamplePlugin
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
	// Init Resync, ETCD + ETCD sync
	ef.ResyncOrch.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("resync-orch")
	ef.ETCD.Deps.PluginInfraDeps = *ef.FlavorLocal.InfraDeps("etcdv3")
	ef.ETCDDataSync.Deps.PluginLogDeps = *ef.FlavorLocal.LogDeps("etcdv3-datasync")
	ef.ETCDDataSync.KvPlugin = &ef.ETCD
	ef.ETCDDataSync.ResyncOrch = &ef.ResyncOrch
	ef.ETCDDataSync.ServiceLabel = &ef.FlavorLocal.ServiceLabel
	// Inject infra + transport (publisher, watcher) to example plugin
	ef.DatasyncExample.PluginInfraDeps = *ef.FlavorLocal.InfraDeps("datasync-example")
	ef.DatasyncExample.Publisher = &ef.ETCDDataSync
	ef.DatasyncExample.Watcher = &ef.ETCDDataSync
	ef.DatasyncExample.closeChannel = ef.closeChan

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

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	Deps

	changeChannel chan datasync.ChangeEvent  // Channel used by the watcher for change events
	resyncChannel chan datasync.ResyncEvent  // Channel used by the watcher for resync events
	context       context.Context            // Used to cancel watching
	watchDataReg  datasync.WatchRegistration // To subscribe on data change/resync events
	// Fields below are used to properly finish the example
	eventCounter uint8
	closeChannel *chan struct{}
}

// Deps is a helper struct which is grouping all dependencies injected to the plugin
type Deps struct {
	local.PluginInfraDeps                 // injected
	Publisher datasync.KeyProtoValWriter  // injected - To write ETCD data
	Watcher   datasync.KeyValProtoWatcher // injected - To watch ETCD data
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() error {
	// Initialize plugin fields
	plugin.resyncChannel = make(chan datasync.ResyncEvent)
	plugin.changeChannel = make(chan datasync.ChangeEvent)
	plugin.context = context.Background()


	// Start the consumer (ETCD watcher) before the custom plugin configurator is initialized
	go plugin.consumer()
	// Subscribe watcher to be able to watch on data changes and resync events
	err := plugin.subscribeWatcher()
	if err != nil {
		return err
	}


	plugin.Log.Info("Initialization of the custom plugin for the ETCD example is completed")

	return nil
}

// AfterInit is called after every plugin is initialized
func (plugin *ExamplePlugin) AfterInit() error {

	go plugin.etcdPublisher()

	go plugin.closeExample()

	return nil
}

func (plugin *ExamplePlugin) closeExample() {
	for {
		// Two events are expected for successful example completion
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

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	safeclose.CloseAll(plugin.Publisher, plugin.Watcher, plugin.resyncChannel, plugin.changeChannel)
	return nil
}

/*************
 * ETCD call *
 *************/

// KeyProtoValWriter creates a simple data, then demonstrates CRUD operations with ETCD
func (plugin *ExamplePlugin) etcdPublisher() {
	// Wait for the consumer to initialize
	time.Sleep(3 * time.Second)
	plugin.Log.Print("KeyValPublisher started")

	// Convert data to the generated proto format
	exampleData := plugin.buildData("string1", 0, true)

	// PUT: examplePut demonstrates how to use the Data Broker Put() API to create (or update) a simple data
	// structure into ETCD
	label := etcdKeyPrefixLabel(plugin.ServiceLabel.GetAgentLabel(), "index")
	plugin.Log.Infof("Write data to %v", label)
	plugin.Publisher.Put(label, exampleData)

	// Prepare different set of data
	plugin.Log.Infof("Update data at %v", label)
	exampleData = plugin.buildData("string2", 1, false)

	// UPDATE: Put() performs both create operations (if index does not exist) and update operations
	// (if the index exists)
	plugin.Publisher.Put(label, exampleData)
}

// The ETCD key prefix used for this example
func etcdKeyPrefix(agentLabel string) string {
	return "/vnf-agent/" + agentLabel + "/api/v1/example/db/simple/"
}

// The ETCD key (the key prefix + label)
func etcdKeyPrefixLabel(agentLabel string, index string) string {
	return etcdKeyPrefix(agentLabel) + index
}

/***********
 * KeyValProtoWatcher *
 ***********/

// Consumer (watcher) is subscribed to watch on data store changes. Change arrives via data change channel and
// its key is parsed
func (plugin *ExamplePlugin) consumer() {
	plugin.Log.Print("KeyValProtoWatcher started")
	for {
		select {
		case dataChng := <-plugin.changeChannel:
			// If event arrives, the key is extracted and used together with the expected prefix to
			// identify item
			key := dataChng.GetKey()
			if strings.HasPrefix(key, etcdKeyPrefix(plugin.ServiceLabel.GetAgentLabel())) {
				var value, previousValue etcd_example.EtcdExample
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
				// Increase event counter (expecting two events)
				plugin.eventCounter++
			}
			// Another strings.HasPrefix(key, etcd prefix) ...
		case <-plugin.context.Done():
			plugin.Log.Warnf("Stop watching events")
		}
	}
}

// KeyValProtoWatcher is subscribed to data change channel and resync channel. ETCD watcher adapter is used for this purpose
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

// Create simple ETCD data structure with provided data values
func (plugin *ExamplePlugin) buildData(stringVal string, uint32Val uint32, boolVal bool) *etcd_example.EtcdExample {
	return &etcd_example.EtcdExample{
		StringVal: stringVal,
		Uint32Val: uint32Val,
		BoolVal:   boolVal,
	}
}
