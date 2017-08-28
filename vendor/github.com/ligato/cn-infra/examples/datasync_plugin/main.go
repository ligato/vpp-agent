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
	"github.com/ligato/cn-infra/flavors/localdeps"
	"github.com/ligato/cn-infra/logging"
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

var log logging.Logger

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with ExampleFlavor
func main() {
	log = logroot.StandardLogger()
	// Init close channel to stop the example
	exampleFinished := make(chan struct{}, 1)

	flavor := ExampleFlavor{}

	// Create new agent
	agent := core.NewAgent(log, 15*time.Second, append(flavor.Plugins())...)

	// End when the ETCD example is finished
	go closeExample("etcd txn example finished", exampleFinished)

	core.EventLoopWithInterrupt(agent, exampleFinished)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(12 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
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

	injected bool
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
	ef.DatasyncExample.InfraDeps = *ef.FlavorLocal.InfraDeps("datasync-example")
	ef.DatasyncExample.Publisher = &ef.ETCDDataSync
	ef.DatasyncExample.Watcher = &ef.ETCDDataSync

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
}

// Deps is here to group injected dependencies of plugin to not mix with other plugin fields
type Deps struct {
	InfraDeps localdeps.PluginInfraDeps   // injected
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
	return nil
}

// AfterInit is called after every plugin is initialized
func (plugin *ExamplePlugin) AfterInit() error {
	// Start the consumer (ETCD watcher) before the custom plugin configurator is initialized
	go plugin.consumer()

	go plugin.etcdPublisher()

	// Subscribe watcher to be able to watch on data changes and resync events
	err := plugin.subscribeWatcher()
	if err != nil {
		return err
	}

	log.Info("Initialization of the custom plugin for the ETCD example is completed")

	return nil
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

const etcdIndex string = "index"

// KeyProtoValWriter creates a simple data, then demonstrates CRUD operations with ETCD
func (plugin *ExamplePlugin) etcdPublisher() {
	time.Sleep(3 * time.Second)
	log.Print("KeyValPublisher started")

	// Convert data to the generated proto format
	exampleData := plugin.buildData("string1", 0, true)

	// PUT: examplePut demonstrates how to use the Data Broker Put() API to create (or update) a simple data
	// structure into ETCD
	label := etcdKeyPrefixLabel(plugin.InfraDeps.ServiceLabel.GetAgentLabel(), etcdIndex)
	log.Infof("Write data to %v", label)
	plugin.Publisher.Put(label, exampleData)

	// Prepare different set of data
	log.Infof("Update data at %v", label)
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
	log.Print("KeyValProtoWatcher started")
	for {
		select {
		case dataChng := <-plugin.changeChannel:
			log.Print("event")
			// If event arrives, the key is extracted and used together with the expected prefix to
			// identify item
			key := dataChng.GetKey()
			if strings.HasPrefix(key, etcdKeyPrefix(plugin.InfraDeps.ServiceLabel.GetAgentLabel())) {
				var value, previousValue etcd_example.EtcdExample
				// The first return value is diff - boolean flag whether previous value exists or not
				err := dataChng.GetValue(&value)
				if err != nil {
					log.Error(err)
				}
				diff, err := dataChng.GetPrevValue(&previousValue)
				if err != nil {
					log.Error(err)
				}
				log.Infof("Event arrived to etcd eventHandler, key %v, update: %v, change type: %v,",
					dataChng.GetKey(), diff, dataChng.GetChangeType())
			}
			// Another strings.HasPrefix(key, etcd prefix) ...
		case <-plugin.context.Done():
			log.Warnf("Stop watching events")
		}
	}
}

// KeyValProtoWatcher is subscribed to data change channel and resync channel. ETCD watcher adapter is used for this purpose
func (plugin *ExamplePlugin) subscribeWatcher() (err error) {
	prefix := etcdKeyPrefix(plugin.InfraDeps.ServiceLabel.GetAgentLabel())
	log.Infof("Prefix: %v", prefix)
	plugin.watchDataReg, err = plugin.Watcher.
		Watch("Example etcd plugin", plugin.changeChannel, plugin.resyncChannel, prefix)
	if err != nil {
		return err
	}

	log.Info("KeyValProtoWatcher subscribed")

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
