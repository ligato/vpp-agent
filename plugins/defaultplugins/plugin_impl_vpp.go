// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package defaultplugins

import (
	"context"

	"sync"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	ifaceidx2 "github.com/ligato/vpp-agent/plugins/linuxplugin/ifaceidx"
)

// no operation writer that helps avoiding NIL pointer based segmentation fault
// used as default if some dependency was not injected
var (
	// no operation writer that helps avoiding NIL pointer based segmentation fault
	// used as default if some dependency was not injected
	noopWriter = &datasync.CompositeKVProtoWriter{Adapters: []datasync.KeyProtoValWriter{}}

	// no operation watcher that helps avoiding NIL pointer based segmentation fault
	// used as default if some dependency was not injected
	noopWatcher = &datasync.CompositeKVProtoWatcher{Adapters: []datasync.KeyValProtoWatcher{}}
)

// Default MTU value. Mtu can be set via defaultplugins config or directly with interface json (higher priority). If none
// is set, use default
const defaultMtu = 9000

// Plugin implements Plugin interface, therefore it can be loaded with other plugins
type Plugin struct {
	Deps

	aclConfigurator *aclplugin.ACLConfigurator
	aclL3L4Indexes  idxvpp.NameToIdxRW
	aclL2Indexes    idxvpp.NameToIdxRW

	swIfIndexes          ifaceidx.SwIfIndexRW
	linuxIfIndexes       ifaceidx2.LinuxIfIndex
	ifConfigurator       *ifplugin.InterfaceConfigurator
	ifStateUpdater       *ifplugin.InterfaceStateUpdater
	ifVppNotifChan       chan govppapi.Message
	ifStateChan          chan *intf.InterfaceStateNotification
	bdVppNotifChan       chan l2plugin.BridgeDomainStateMessage
	bdStateUpdater       *l2plugin.BridgeDomainStateUpdater
	bdStateChan          chan *l2plugin.BridgeDomainStateNotification
	bfdSessionIndexes    idxvpp.NameToIdxRW
	bfdAuthKeysIndexes   idxvpp.NameToIdxRW
	bfdEchoFunctionIndex idxvpp.NameToIdxRW

	bfdConfigurator   *ifplugin.BFDConfigurator
	bdConfigurator    *l2plugin.BDConfigurator
	fibConfigurator   *l2plugin.FIBConfigurator
	xcConfigurator    *l2plugin.XConnectConfigurator
	routeConfigurator *l3plugin.RouteConfigurator
	bdIndexes         bdidx.BDIndexRW
	ifToBdDesIndexes  idxvpp.NameToIdxRW
	ifToBdRealIndexes idxvpp.NameToIdxRW
	fibIndexes        idxvpp.NameToIdxRW
	fibDesIndexes     idxvpp.NameToIdxRW
	xcIndexes         idxvpp.NameToIdxRW
	routeIndexes      idxvpp.NameToIdxRW
	errorIndexes      idxvpp.NameToIdxRW
	ifIdxWatchCh      chan ifaceidx.SwIfIdxDto
	bdIdxWatchCh      chan bdidx.ChangeDto
	linuxIfIdxWatchCh chan ifaceidx2.LinuxIfIndexDto

	resyncConfigChan     chan datasync.ResyncEvent
	resyncStatusChan     chan datasync.ResyncEvent
	changeChan           chan datasync.ChangeEvent //TODO dedicated type abstracted from ETCD
	ifStateNotifications messaging.ProtoPublisher
	ifMtu                uint32

	watchConfigReg datasync.WatchRegistration
	watchStatusReg datasync.WatchRegistration

	errorChannel chan ErrCtx
	errorIdxSeq  uint32

	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	// inject all below
	local.PluginInfraDeps
	Publish           datasync.KeyProtoValWriter
	PublishStatistics datasync.KeyProtoValWriter
	Watch             datasync.KeyValProtoWatcher
	IfStatePub        datasync.KeyProtoValWriter
	GoVppmux          govppmux.API
	Linux             linuxpluginAPI
}

type linuxpluginAPI interface {
	// GetLinuxIfIndexes gives access to mapping of logical names (used in ETCD configuration) to corresponding Linux
	// interface indexes. This mapping is especially helpful for plugins that need to watch for newly added or deleted
	// Linux interfaces.
	GetLinuxIfIndexes() ifaceidx2.LinuxIfIndex
}

// DPConfig holds the value of maximum transmission unit in bytes.
type DPConfig struct {
	Mtu uint32 `json:"mtu"`
}

var (
	// gPlugin holds the global instance of the Plugin
	gPlugin *Plugin
)

// plugin function is used in api to access the plugin instance. It panics if the plugin instance is not initialized.
func plugin() *Plugin {
	if gPlugin == nil {
		panic("Trying to access the Interface Plugin but it is still not initialized")
	}
	return gPlugin
}

// Init gets handlers for ETCD, Messaging and delegates them to ifConfigurator & ifStateUpdater
func (plugin *Plugin) Init() error {
	plugin.Log.Debug("Initializing interface plugin")

	plugin.fixNilPointers()

	plugin.ifStateNotifications = plugin.Deps.IfStatePub
	config, err := plugin.retrieveMtuConfig()
	if err != nil {
		return err
	}
	if config != nil {
		plugin.ifMtu = config.Mtu
		plugin.Log.Infof("Mtu read from config us set to %v", plugin.ifMtu)
	} else {
		plugin.ifMtu = defaultMtu
		plugin.Log.Infof("Mtu config not found, set to default value %v", plugin.ifMtu)
	}

	// all channels that are used inside of publishIfStateEvents or watchEvents must be created in advance!
	plugin.ifStateChan = make(chan *intf.InterfaceStateNotification, 100)
	plugin.bdStateChan = make(chan *l2plugin.BridgeDomainStateNotification, 100)
	plugin.resyncConfigChan = make(chan datasync.ResyncEvent)
	plugin.resyncStatusChan = make(chan datasync.ResyncEvent)
	plugin.changeChan = make(chan datasync.ChangeEvent)
	plugin.ifIdxWatchCh = make(chan ifaceidx.SwIfIdxDto, 100)
	plugin.bdIdxWatchCh = make(chan bdidx.ChangeDto, 100)
	plugin.linuxIfIdxWatchCh = make(chan ifaceidx2.LinuxIfIndexDto, 100)
	plugin.errorChannel = make(chan ErrCtx, 100)

	// create plugin context, save cancel function into the plugin handle
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())

	//FIXME run following go routines later than following init*() calls - just before Watch()

	// run event handler go routines
	go plugin.publishIfStateEvents(ctx)
	go plugin.publishBdStateEvents(ctx)
	go plugin.watchEvents(ctx)

	// run error handler
	go plugin.changePropagateError()

	err = plugin.initIF(ctx)
	if err != nil {
		return err
	}
	err = plugin.initACL(ctx)
	if err != nil {
		return err
	}
	err = plugin.initL2(ctx)
	if err != nil {
		return err
	}
	err = plugin.initL3(ctx)
	if err != nil {
		return err
	}

	err = plugin.initErrorHandler()
	if err != nil {
		return err
	}

	err = plugin.subscribeWatcher()
	if err != nil {
		return err
	}

	gPlugin = plugin

	return nil
}

// fixNilPointers sets noopWriter & nooWatcher for nil dependencies
func (plugin *Plugin) fixNilPointers() {
	if plugin.Deps.Publish == nil {
		plugin.Deps.Publish = noopWriter
		plugin.Log.Debug("setting default noop writer for Publish dependency")
	}
	if plugin.Deps.PublishStatistics == nil {
		plugin.Deps.PublishStatistics = noopWriter
		plugin.Log.Debug("setting default noop writer for PublishStatistics dependency")
	}
	if plugin.Deps.IfStatePub == nil {
		plugin.Deps.IfStatePub = noopWriter
		plugin.Log.Debug("setting default noop writer for IfStatePub dependency")
	}
	if plugin.Deps.Watch == nil {
		plugin.Deps.Watch = noopWatcher
		plugin.Log.Debug("setting default noop watcher for Watch dependency")
	}
}

func (plugin *Plugin) initIF(ctx context.Context) error {
	// configurator loggers
	ifLogger := plugin.Log.NewLogger("-if-conf")
	ifStateLogger := plugin.Log.NewLogger("-if-state")
	bfdLogger := plugin.Log.NewLogger("-bfd-conf")
	// Interface indexes
	plugin.swIfIndexes = ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(ifLogger, plugin.PluginName,
		"sw_if_indexes", ifaceidx.IndexMetadata))

	// get pointer to the map with Linux interface indexes
	if plugin.Linux != nil {
		plugin.linuxIfIndexes = plugin.Linux.GetLinuxIfIndexes()
	} else {
		plugin.linuxIfIndexes = nil
	}

	// BFD session
	plugin.bfdSessionIndexes = nametoidx.NewNameToIdx(bfdLogger, plugin.PluginName, "bfd_session_indexes", nil)

	// BFD key
	plugin.bfdAuthKeysIndexes = nametoidx.NewNameToIdx(bfdLogger, plugin.PluginName, "bfd_auth_keys_indexes", nil)

	// BFD echo function
	plugin.bfdEchoFunctionIndex = nametoidx.NewNameToIdx(bfdLogger, plugin.PluginName, "bfd_echo_function_index", nil)

	// BFD echo function
	BfdRemovedAuthKeys := nametoidx.NewNameToIdx(bfdLogger, plugin.PluginName, "bfd_removed_auth_keys", nil)

	plugin.ifVppNotifChan = make(chan govppapi.Message, 100)
	plugin.ifStateUpdater = &ifplugin.InterfaceStateUpdater{Log: ifStateLogger, GoVppmux: plugin.GoVppmux}
	plugin.ifStateUpdater.Init(ctx, plugin.swIfIndexes, plugin.ifVppNotifChan, func(state *intf.InterfaceStateNotification) {
		select {
		case plugin.ifStateChan <- state:
			// OK
		default:
			plugin.Log.Debug("Unable to send to the ifStateNotifications channel - channel buffer full.")
		}
	})

	plugin.Log.Debug("ifStateUpdater Initialized")

	plugin.ifConfigurator = &ifplugin.InterfaceConfigurator{
		Log:          ifLogger,
		GoVppmux:     plugin.GoVppmux,
		ServiceLabel: plugin.ServiceLabel,
		Linux:        plugin.Linux,
	}
	plugin.ifConfigurator.Init(plugin.swIfIndexes, plugin.ifMtu, plugin.ifVppNotifChan)

	plugin.Log.Debug("ifConfigurator Initialized")

	plugin.bfdConfigurator = &ifplugin.BFDConfigurator{
		Log:          bfdLogger,
		GoVppmux:     plugin.GoVppmux,
		ServiceLabel: plugin.ServiceLabel,
		SwIfIndexes:  plugin.swIfIndexes,
		BfdIDSeq:     1,
	}
	plugin.bfdConfigurator.Init(plugin.bfdSessionIndexes, plugin.bfdAuthKeysIndexes, plugin.bfdEchoFunctionIndex, BfdRemovedAuthKeys)

	plugin.Log.Debug("bfdConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initACL(ctx context.Context) error {
	// logger
	aclLogger := plugin.Log.NewLogger("-acl-plugin")
	var err error
	plugin.aclL3L4Indexes = nametoidx.NewNameToIdx(aclLogger, plugin.PluginName, "acl_l3_l4_indexes", nil)

	plugin.aclL2Indexes = nametoidx.NewNameToIdx(aclLogger, plugin.PluginName, "acl_l2_indexes", nil)

	plugin.aclConfigurator = &aclplugin.ACLConfigurator{
		Log:            aclLogger,
		GoVppmux:       plugin.GoVppmux,
		ACLL3L4Indexes: plugin.aclL3L4Indexes,
		ACLL2Indexes:   plugin.aclL2Indexes,
		SwIfIndexes:    plugin.swIfIndexes,
	}

	// Init ACL plugin
	err = plugin.aclConfigurator.Init()
	if err != nil {
		return err
	}
	plugin.Log.Debug("aclConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initL2(ctx context.Context) error {
	// loggers
	bdLogger := plugin.Log.NewLogger("-l2-bd-conf")
	bdStateLogger := plugin.Log.NewLogger("-l2-bd-state")
	fibLogger := plugin.Log.NewLogger("-l2-fib-conf")
	xcLogger := plugin.Log.NewLogger("-l2-xc-conf")
	// Bridge domain indexes
	plugin.bdIndexes = bdidx.NewBDIndex(nametoidx.NewNameToIdx(bdLogger, plugin.PluginName,
		"bd_indexes", bdidx.IndexMetadata))

	// Interface to bridge domain indexes - desired state
	plugin.ifToBdDesIndexes = nametoidx.NewNameToIdx(bdLogger, plugin.PluginName, "if_to_bd_des_indexes", nil)

	// Interface to bridge domain indexes - current state

	plugin.ifToBdRealIndexes = nametoidx.NewNameToIdx(bdLogger, plugin.PluginName, "if_to_bd_real_indexes", nil)

	plugin.bdConfigurator = &l2plugin.BDConfigurator{
		Log:                bdLogger,
		GoVppmux:           plugin.GoVppmux,
		SwIfIndexes:        plugin.swIfIndexes,
		BdIndexes:          plugin.bdIndexes,
		BridgeDomainIDSeq:  1,
		IfToBdIndexes:      plugin.ifToBdDesIndexes,
		IfToBdRealStateIdx: plugin.ifToBdRealIndexes,
	}

	// Bridge domain state and state updater
	plugin.bdVppNotifChan = make(chan l2plugin.BridgeDomainStateMessage, 100)
	plugin.bdStateUpdater = &l2plugin.BridgeDomainStateUpdater{Log: bdStateLogger, GoVppmux: plugin.GoVppmux}
	plugin.bdStateUpdater.Init(ctx, plugin.bdIndexes, plugin.swIfIndexes, plugin.bdVppNotifChan, func(state *l2plugin.BridgeDomainStateNotification) {
		select {
		case plugin.bdStateChan <- state:
			// OK
		default:
			plugin.Log.Debug("Unable to send to the bdState channel: buffer is full.")
		}
	})

	// FIB indexes
	plugin.fibIndexes = nametoidx.NewNameToIdx(fibLogger, plugin.PluginName, "fib_indexes", nil)

	plugin.fibConfigurator = &l2plugin.FIBConfigurator{
		Log:           fibLogger,
		GoVppmux:      plugin.GoVppmux,
		SwIfIndexes:   plugin.swIfIndexes,
		BdIndexes:     plugin.bdIndexes,
		IfToBdIndexes: plugin.ifToBdDesIndexes,
		FibIndexes:    plugin.fibIndexes,
		FibIndexSeq:   1,
		FibDesIndexes: plugin.fibDesIndexes,
	}

	// L2 xConnect indexes

	plugin.xcIndexes = nametoidx.NewNameToIdx(xcLogger, plugin.PluginName, "xc_indexes", nil)

	plugin.xcConfigurator = &l2plugin.XConnectConfigurator{
		Log:         xcLogger,
		GoVppmux:    plugin.GoVppmux,
		SwIfIndexes: plugin.swIfIndexes,
		XcIndexes:   plugin.xcIndexes,
		XcIndexSeq:  1,
	}

	// Init
	err := plugin.bdConfigurator.Init(plugin.bdVppNotifChan)
	if err != nil {
		return err
	}

	plugin.Log.Debug("bdConfigurator Initialized")

	err = plugin.fibConfigurator.Init()
	if err != nil {
		return err
	}

	plugin.Log.Debug("fibConfigurator Initialized")

	err = plugin.xcConfigurator.Init()
	if err != nil {
		return err
	}

	plugin.Log.Debug("xcConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initL3(ctx context.Context) error {
	l3Logger := plugin.Log.NewLogger("-l3-plugin")
	plugin.routeIndexes = nametoidx.NewNameToIdx(l3Logger, plugin.PluginName, "route_indexes", nil)

	plugin.routeConfigurator = &l3plugin.RouteConfigurator{
		Log:           l3Logger,
		GoVppmux:      plugin.GoVppmux,
		RouteIndexes:  plugin.routeIndexes,
		RouteIndexSeq: 1,
		SwIfIndexes:   plugin.swIfIndexes,
	}
	err := plugin.routeConfigurator.Init()
	if err != nil {
		return err
	}

	plugin.Log.Debug("routeConfigurator Initialized")

	return nil
}

func (plugin *Plugin) retrieveMtuConfig() (*DPConfig, error) {
	config := &DPConfig{}
	found, err := plugin.PluginConfig.GetValue(config)
	if !found {
		plugin.Log.Debug("Mtu config not found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	plugin.Log.Debug("config found, Mtu value %v", config.Mtu)
	return config, err
}

func (plugin *Plugin) initErrorHandler() error {
	ehLogger := plugin.Log.NewLogger("-error-handler")
	plugin.errorIndexes = nametoidx.NewNameToIdx(ehLogger, plugin.PluginName, "error_indexes", nil)

	// Init mapping index
	plugin.errorIdxSeq = 1
	return nil
}

// AfterInit delegates to ifStateUpdater
func (plugin *Plugin) AfterInit() error {
	plugin.Log.Debug("vpp plugins AfterInit begin")

	err := plugin.ifStateUpdater.AfterInit()
	if err != nil {
		return err
	}

	plugin.Log.Debug("vpp plugins AfterInit finished successfully")

	return nil
}

// Close cleans up the resources
func (plugin *Plugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	_, err := safeclose.CloseAll(plugin.watchStatusReg, plugin.watchConfigReg, plugin.changeChan,
		plugin.resyncStatusChan, plugin.resyncConfigChan,
		plugin.ifConfigurator, plugin.ifStateUpdater, plugin.ifVppNotifChan, plugin.errorChannel,
		plugin.bdVppNotifChan, plugin.bdConfigurator, plugin.fibConfigurator, plugin.bfdConfigurator,
		plugin.xcConfigurator, plugin.routeConfigurator)

	return err
}
