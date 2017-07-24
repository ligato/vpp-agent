package main

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/examples/model"
	"github.com/ligato/cn-infra/examples/simple-agent/generic"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/config"
	"os"
	"strconv"
	"strings"
	"time"
)

// *************************************************************************
// This file contains examples of simple Data Broker CRUD operations
// (APIs) including an event handler (watcher). The CRUD operations
// supported by the Data Broker are as follows:
// - Create/Update: dataBroker.Put()
// - Read:          dataBroker.Get()
// - Delete:        dataBroker.Delete()
//
// These functions are called from the REST API. CRUD operations are
// done as single operations and as a part of the transaction
// ************************************************************************/

/********
 * Main *
 ********/

var log logging.Logger

// Main allows running Example Plugin as a statically linked binary with Agent Core Plugins. Close channel and plugins
// required for the example are initialized. Agent is instantiated with generic plugins (ETCD, Kafka, Status check,
// HTTP and Log), resync plugin and example plugin which demonstrates ETCD functionality.
func main() {
	log = logroot.Logger()
	// Init close channel to stop the example
	closeChannel := make(chan struct{}, 1)

	flavour := generic.Flavour{}
	// Resync plugin
	resyncPlugin := &core.NamedPlugin{PluginName: resync.PluginID, Plugin: &resync.Plugin{}}
	// Example plugin (ETCD)
	examplePlugin := &core.NamedPlugin{PluginName: PluginID, Plugin: &ExamplePlugin{ServiceLabel: &flavour.ServiceLabel}}

	// Create new agent
	agent := core.NewAgent(log, 15*time.Second, append(flavour.Plugins(), resyncPlugin, examplePlugin)...)

	// End when the ETCD example is finished
	go closeExample("etcd txn example finished", closeChannel)

	core.EventLoopWithInterrupt(agent, closeChannel)
}

// Stop the agent with desired info message
func closeExample(message string, closeChannel chan struct{}) {
	time.Sleep(12 * time.Second)
	log.Info(message)
	closeChannel <- struct{}{}
}

/**********************
 * Example plugin API *
 **********************/

// PluginID of the custom ETCD plugin
const PluginID core.PluginName = "example-plugin"

/******************
 * Example plugin *
 ******************/

// ExamplePlugin implements Plugin interface which is used to pass custom plugin instances to the agent
type ExamplePlugin struct {
	ServiceLabel        *servicelabel.Plugin
	exampleConfigurator *ExampleConfigurator           // Plugin configurator
	transport           datasync.TransportAdapter      // To access ETCD data
	changeChannel       chan datasync.ChangeEvent      // Channel used by the watcher for change events
	resyncChannel       chan datasync.ResyncEvent      // Channel used by the watcher for resync events
	watchDataReg        datasync.WatchDataRegistration // To subscribe on data change/resync events
}

// Init is the entry point into the plugin that is called by Agent Core when the Agent is coming up.
// The Go native plugin mechanism that was introduced in Go 1.8
func (plugin *ExamplePlugin) Init() error {
	// Initialize plugin fields
	plugin.exampleConfigurator = &ExampleConfigurator{plugin.ServiceLabel}
	plugin.transport = datasync.GetTransport()
	plugin.resyncChannel = make(chan datasync.ResyncEvent)
	plugin.changeChannel = make(chan datasync.ChangeEvent)

	// Start the consumer (ETCD watcher) before the custom plugin configurator is initialized
	go plugin.consumer()

	// Now initialize the plugin configurator
	plugin.exampleConfigurator.Init()

	// Subscribe watcher to be able to watch on data changes and resync events
	plugin.subscribeWatcher()

	log.Info("Initialization of the custom plugin for the ETCD example is completed")

	return nil
}

// Close is called by Agent Core when the Agent is shutting down. It is supposed to clean up resources that were
// allocated by the plugin during its lifetime
func (plugin *ExamplePlugin) Close() error {
	plugin.exampleConfigurator.Close()
	plugin.watchDataReg.Close()
	return nil
}

/*************************
 * Example plugin config *
 *************************/

// ExampleConfigurator usually initializes configuration-specific fields or other tasks (e.g. defines GOVPP channels
// if they are used, checks VPP message compatibility etc.)
type ExampleConfigurator struct {
	ServiceLabel *servicelabel.Plugin
}

// Init members of configurator
func (configurator *ExampleConfigurator) Init() (err error) {
	// There is nothing to init in the example
	log.Info("Custom plugin configurator initialized")

	// Now the configurator is initialized and the watcher is already running (started in plugin initialization),
	// so publisher is used to put data to ETCD
	go func() {
		// Show simple ETCD CRUD
		configurator.etcdPublisher()
		// Show transactions
		configurator.etcdTxnPublisher()
	}()

	return err
}

// Close function for example plugin (just for representation, there is nothing to close in the example)
func (configurator *ExampleConfigurator) Close() {}

/*************
 * ETCD call *
 *************/

const etcdIndex string = "index"

// Publisher creates a simple data, then demonstrates CRUD operations with ETCD
func (configurator *ExampleConfigurator) etcdPublisher() {
	// Get data broker to communicate with ETCD
	cfg := &etcdv3.Config{}

	configFile := os.Getenv("ETCDV3_CONFIG")
	if configFile != "" {
		err := config.ParseConfigFromYamlFile(configFile, cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
	etcdConfig, err := etcdv3.ConfigToClientv3(cfg)
	if err != nil {
		log.Fatal(err)
	}

	bDB, _ := etcdv3.NewEtcdConnectionWithBytes(*etcdConfig, log)
	dataBroker := kvproto.NewProtoWrapperWithSerializer(bDB, &keyval.SerializerJSON{}).
		NewBroker(configurator.ServiceLabel.GetAgentPrefix())

	time.Sleep(3 * time.Second)

	// Convert data to the generated proto format
	exampleData := configurator.buildData("string1", 0, true)

	// PUT: examplePut demonstrates how to use the Data Broker Put() API to create (or update) a simple data
	// structure into ETCD
	dataBroker.Put(etcdKeyPrefixLabel(configurator.ServiceLabel.GetAgentLabel(), etcdIndex), exampleData)

	// Prepare different set of data
	exampleData = configurator.buildData("string2", 1, false)

	// UPDATE: Put() performs both create operations (if index does not exist) and update operations
	// (if the index exists)
	dataBroker.Put(etcdKeyPrefixLabel(configurator.ServiceLabel.GetAgentLabel(), etcdIndex), exampleData)

	// GET: exampleGet demonstrates how to use the Data Broker Get() API to read a simple data structure from ETCD
	result := etcd_example.EtcdExample{}
	found, _, err := dataBroker.GetValue(etcdKeyPrefixLabel(configurator.ServiceLabel.GetAgentLabel(), etcdIndex), &result)
	if err != nil {
		log.Error(err)
	}
	if found {
		log.Infof("Data read from ETCD data store. Values: %v, %v, %v",
			result.StringVal, result.Uint32Val, result.BoolVal)
	} else {
		log.Error("Data not found")
	}

	// DELETE: demonstrates how to use the Data Broker Delete() API to delete a simple data structure from ETCD
	dataBroker.Delete(etcdKeyPrefixLabel(configurator.ServiceLabel.GetAgentLabel(), etcdIndex))
}

// Publisher creates a simple data, then demonstrates transaction operations with ETCD
func (configurator *ExampleConfigurator) etcdTxnPublisher() {
	log.Info("Preparing bridge domain data")
	// Get data broker to communicate with ETCD
	cfg := &etcdv3.Config{}

	configFile := os.Getenv("ETCDV3_CONFIG")
	if configFile != "" {
		err := config.ParseConfigFromYamlFile(configFile, cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
	etcdConfig, err := etcdv3.ConfigToClientv3(cfg)
	if err != nil {
		log.Fatal(err)
	}

	bDB, _ := etcdv3.NewEtcdConnectionWithBytes(*etcdConfig, log)
	dataBroker := kvproto.NewProtoWrapperWithSerializer(bDB, &keyval.SerializerJSON{}).
		NewBroker(configurator.ServiceLabel.GetAgentPrefix())

	time.Sleep(3 * time.Second)

	// This is how to use the Data Broker Txn API to create a new transaction. It is called from the HTTP handler
	// when a user triggers the creation of a new transaction via REST
	putTxn := dataBroker.NewTxn()
	for i := 1; i <= 3; i++ {
		exampleData1 := configurator.buildData("string", uint32(i), true)
		// putTxn.Put demonstrates how to use the Data Broker Txn Put() API. It is called from the HTTP handler
		// when a user invokes the REST API to add a new Put() operation to the transaction
		putTxn = putTxn.Put(etcdKeyPrefixLabel(configurator.ServiceLabel.GetAgentLabel(), etcdIndex+strconv.Itoa(i)), exampleData1)
	}
	// putTxn.Commit() demonstrates how to use the Data Broker Txn Commit() API. It is called from the HTTP handler
	// when a user invokes the REST API to commit a transaction.
	err = putTxn.Commit()
	if err != nil {
		log.Error(err)
	}
	// Another transaction chain to demonstrate delete operations. Put and Delete operations can be used together
	// within one transaction
	deleteTxn := dataBroker.NewTxn()
	for i := 1; i <= 3; i++ {
		// deleteTxn.Delete demonstrates how to use the Data Broker Txn Delete() API. It is called from the
		// HTTP handler when a user invokes the REST API to add a new Delete() operation to the transaction.
		// Put and Delete operations can be combined in the same transaction chain
		deleteTxn = deleteTxn.Delete(etcdKeyPrefixLabel(configurator.ServiceLabel.GetAgentLabel(), etcdIndex+strconv.Itoa(i)))
	}

	// Commit transactions to data store. Transaction executes multiple operations in a more efficient way in
	// contrast to executing them one by one.
	err = deleteTxn.Commit()
	if err != nil {
		log.Error(err)
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

/***********
 * Watcher *
 ***********/

// Consumer (watcher) is subscribed to watch on data store changes. Change arrives via data change channel and
// its key is parsed
func (plugin *ExamplePlugin) consumer() {
	log.Print("Watcher started")
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
		}
	}
}

// Watcher is subscribed to data change channel and resync channel. ETCD transport adapter is used for this purpose
func (plugin *ExamplePlugin) subscribeWatcher() (err error) {
	plugin.watchDataReg, err = plugin.transport.
		WatchData("Example etcd plugin", plugin.changeChannel, plugin.resyncChannel, etcdKeyPrefix(plugin.ServiceLabel.GetAgentLabel()))
	if err != nil {
		return err
	}

	log.Info("Watcher subscribed")

	return nil
}

// Create simple ETCD data structure with provided data values
func (configurator *ExampleConfigurator) buildData(stringVal string, uint32Val uint32, boolVal bool) *etcd_example.EtcdExample {
	return &etcd_example.EtcdExample{
		StringVal: stringVal,
		Uint32Val: uint32Val,
		BoolVal:   boolVal,
	}
}
