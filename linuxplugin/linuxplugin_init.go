package linuxplugin

import (
	"context"
	"sync"

	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/safeclose"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
)

// Plugin implements Plugin interface, therefore it can be loaded with other plugins
type Plugin struct {
	transport datasync.TransportAdapter
	ifIndexes idxvpp.NameToIdxRW

	ifConfigurator *LinuxInterfaceConfigurator

	resyncChan chan datasync.ResyncEvent
	changeChan chan datasync.ChangeEvent // TODO dedicated type abstracted from ETCD

	watchDataReg datasync.WatchDataRegistration

	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

var (
	// gPlugin holds the global instance of the Plugin
	gPlugin *Plugin
)

// plugin function is used in api to access the plugin instance. It panics if the plugin instance is not initialized.
func plugin() *Plugin {
	if gPlugin == nil {
		log.Panic("Trying to access the Linux Interface Plugin but it is still not initialized")
	}
	return gPlugin
}

// Init gets handlers for ETCD, Kafka and delegates them to ifConfigurator
func (plugin *Plugin) Init() error {
	var err error
	plugin.transport = datasync.GetTransport()

	log.Debug("Initializing Linux interface plugin")

	plugin.resyncChan = make(chan datasync.ResyncEvent)
	plugin.changeChan = make(chan datasync.ChangeEvent)

	// create plugin context, save cancel function into the plugin handle
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())

	// run event handler go routines
	go plugin.watchEvents(ctx)

	// Interface indexes
	plugin.ifIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "linux_if_indexes", nil)

	// Linux interface configurator
	plugin.ifConfigurator = &LinuxInterfaceConfigurator{}
	plugin.ifConfigurator.Init(plugin.ifIndexes)

	err = plugin.subscribeWatcher()
	if err != nil {
		return err
	}

	gPlugin = plugin

	return nil
}

// Close cleans up the resources
func (plugin *Plugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	_, err := safeclose.CloseAll(plugin.watchDataReg, plugin.changeChan, plugin.resyncChan,
		plugin.ifConfigurator)

	return err
}
