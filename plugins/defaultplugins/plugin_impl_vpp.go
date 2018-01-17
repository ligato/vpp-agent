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
	"os"

	"sync"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/nsidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	ifaceLinux "github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	"github.com/namsral/flag"
)

// defaultpluigns specific flags
var (
	// skip resync flag
	skipResyncFlag = flag.Bool("skip-vpp-resync", false, "Skip defaultplugins resync with VPP")
)

var (
	// noopWriter (no operation writer) helps avoiding NIL pointer based segmentation fault.
	// It is used as default if some dependency was not injected.
	noopWriter = &datasync.CompositeKVProtoWriter{Adapters: []datasync.KeyProtoValWriter{}}

	// noopWatcher (no operation watcher) helps avoiding NIL pointer based segmentation fault.
	// It is used as default if some dependency was not injected.
	noopWatcher = &datasync.CompositeKVProtoWatcher{Adapters: []datasync.KeyValProtoWatcher{}}
)

// VPP resync strategy. Can be set in defaultplugins.conf. If no strategy is set, the default behavior is defined by 'fullResync'
const (
	// fullResync calls the full resync for every default plugin
	fullResync = "full"
	// optimizeColdStart checks existence of the configured interface on the VPP (except local0). If there are any, the full
	// resync is executed, otherwise it's completely skipped.
	// Note: resync will be skipped also in case there is not configuration in VPP but exists in etcd
	optimizeColdStart = "optimize"
	// resync is skipped in any case
	skipResync = "skip"
)

// Plugin implements Plugin interface, therefore it can be loaded with other plugins.
type Plugin struct {
	Deps

	// ACL plugin fields
	aclConfigurator *aclplugin.ACLConfigurator
	aclL3L4Indexes  idxvpp.NameToIdxRW
	aclL2Indexes    idxvpp.NameToIdxRW

	// Interface plugin fields
	ifConfigurator       *ifplugin.InterfaceConfigurator
	swIfIndexes          ifaceidx.SwIfIndexRW
	linuxIfIndexes       ifaceLinux.LinuxIfIndex
	ifStateUpdater       *ifplugin.InterfaceStateUpdater
	ifVppNotifChan       chan govppapi.Message
	ifStateChan          chan *intf.InterfaceStateNotification
	ifStateNotifications messaging.ProtoPublisher
	ifIdxWatchCh         chan ifaceidx.SwIfIdxDto
	linuxIfIdxWatchCh    chan ifaceLinux.LinuxIfIndexDto
	stnConfigurator      *ifplugin.StnConfigurator
	stnAllIndexes        idxvpp.NameToIdxRW
	stnUnstoredIndexes   idxvpp.NameToIdxRW

	// Bridge domain fields
	bdConfigurator    *l2plugin.BDConfigurator
	bdIndexes         bdidx.BDIndexRW
	ifToBdDesIndexes  idxvpp.NameToIdxRW
	ifToBdRealIndexes idxvpp.NameToIdxRW
	bdVppNotifChan    chan l2plugin.BridgeDomainStateMessage
	bdStateUpdater    *l2plugin.BridgeDomainStateUpdater
	bdStateChan       chan *l2plugin.BridgeDomainStateNotification
	bdIdxWatchCh      chan bdidx.ChangeDto

	// Bidirectional forwarding detection fields
	bfdSessionIndexes    idxvpp.NameToIdxRW
	bfdAuthKeysIndexes   idxvpp.NameToIdxRW
	bfdEchoFunctionIndex idxvpp.NameToIdxRW
	bfdConfigurator      *ifplugin.BFDConfigurator

	// Forwarding information base fields
	fibConfigurator *l2plugin.FIBConfigurator
	fibIndexes      idxvpp.NameToIdxRW
	fibDesIndexes   idxvpp.NameToIdxRW

	// xConnect fields
	xcConfigurator *l2plugin.XConnectConfigurator
	xcIndexes      idxvpp.NameToIdxRW

	// L3 route fields
	routeConfigurator *l3plugin.RouteConfigurator
	routeIndexes      l3idx.RouteIndexRW

	// L3 arp fields
	arpConfigurator *l3plugin.ArpConfigurator
	arpIndexes      l3idx.ARPIndexRW

	// L4 fields
	l4Configurator      *l4plugin.L4Configurator
	namespaceIndexes    nsidx.AppNsIndexRW
	notConfAppNsIndexes nsidx.AppNsIndexRW

	// Error handler
	errorIndexes idxvpp.NameToIdxRW
	errorChannel chan ErrCtx
	errorIdxSeq  uint32

	// Resync
	resyncConfigChan chan datasync.ResyncEvent
	resyncStatusChan chan datasync.ResyncEvent
	changeChan       chan datasync.ChangeEvent //TODO dedicated type abstracted from ETCD
	watchConfigReg   datasync.WatchRegistration
	watchStatusReg   datasync.WatchRegistration
	omittedPrefixes  []string // list of keys which won't be resynced

	// From config file
	ifMtu          uint32
	resyncStrategy string

	// Common
	enableStopwatch bool
	statusCheckReg  bool
	cancel          context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg              sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Deps groups injected dependencies of plugin so that they do not mix with
// other plugin fieldsMtu.
type Deps struct {
	// inject all below
	local.PluginInfraDeps

	Publish           datasync.KeyProtoValWriter
	PublishStatistics datasync.KeyProtoValWriter
	Watch             datasync.KeyValProtoWatcher
	IfStatePub        datasync.KeyProtoValWriter
	GoVppmux          govppmux.API
	Linux             linuxpluginAPI

	DataSyncs map[string]datasync.KeyProtoValWriter
}

type linuxpluginAPI interface {
	// GetLinuxIfIndexes gives access to mapping of logical names (used in ETCD configuration) to corresponding Linux
	// interface indexes. This mapping is especially helpful for plugins that need to watch for newly added or deleted
	// Linux interfaces.
	GetLinuxIfIndexes() ifaceLinux.LinuxIfIndex
}

// DPConfig holds the defaultpluigns configuration.
type DPConfig struct {
	Mtu              uint32   `json:"mtu"`
	Stopwatch        bool     `json:"stopwatch"`
	Strategy         string   `json:"strategy"`
	StatusPublishers []string `json:"status-publishers"`
}

// DisableResync can be used to disable resync for one or more key prefixes
func (plugin *Plugin) DisableResync(keyPrefix ...string) {
	plugin.Log.Infof("Keys with prefixes %v will be skipped", keyPrefix)
	plugin.omittedPrefixes = keyPrefix
}

// GetSwIfIndexes gives access to mapping of logical names (used in ETCD configuration) to sw_if_index.
// This mapping is helpful if other plugins need to configure VPP by the Binary API that uses sw_if_index input.
func (plugin *Plugin) GetSwIfIndexes() ifaceidx.SwIfIndex {
	return plugin.swIfIndexes
}

// GetBfdSessionIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_session_indexes.
func (plugin *Plugin) GetBfdSessionIndexes() idxvpp.NameToIdx {
	return plugin.bfdSessionIndexes
}

// GetBfdAuthKeyIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_auth_keys.
func (plugin *Plugin) GetBfdAuthKeyIndexes() idxvpp.NameToIdx {
	return plugin.bfdAuthKeysIndexes
}

// GetBfdEchoFunctionIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_echo_function
func (plugin *Plugin) GetBfdEchoFunctionIndexes() idxvpp.NameToIdx {
	return plugin.bfdEchoFunctionIndex
}

// GetBDIndexes gives access to mapping of logical names (used in ETCD configuration) as bd_indexes.
func (plugin *Plugin) GetBDIndexes() bdidx.BDIndex {
	return plugin.bdIndexes
}

// GetFIBIndexes gives access to mapping of logical names (used in ETCD configuration) as fib_indexes.
func (plugin *Plugin) GetFIBIndexes() idxvpp.NameToIdx {
	return plugin.fibIndexes
}

// GetXConnectIndexes gives access to mapping of logical names (used in ETCD configuration) as xc_indexes.
func (plugin *Plugin) GetXConnectIndexes() idxvpp.NameToIdx {
	return plugin.xcIndexes
}

// GetAppNsIndexes gives access to mapping of app-namespace logical names (used in ETCD configuration)
// to their respective indices as assigned by VPP.
func (plugin *Plugin) GetAppNsIndexes() nsidx.AppNsIndex {
	return plugin.namespaceIndexes
}

// DumpACL returns a list of all configured ACL entires
func (plugin *Plugin) DumpACL() (acls []*acl.AccessLists_Acl, err error) {
	return plugin.aclConfigurator.DumpACL()
}

// Init gets handlers for ETCD and Messaging and delegates them to ifConfigurator & ifStateUpdater.
func (plugin *Plugin) Init() error {
	plugin.Log.Debug("Initializing default plugins")
	// handle flag
	flag.Parse()

	// read config file and set all related fields
	config, err := plugin.retrieveDPConfig()
	if err != nil {
		return err
	}
	if config != nil {
		publishers := &datasync.CompositeKVProtoWriter{}
		for _, pub := range config.StatusPublishers {
			db, found := plugin.Deps.DataSyncs[pub]
			if !found {
				plugin.Log.Warnf("Unknown status publisher %q from config", pub)
				continue
			}
			publishers.Adapters = append(publishers.Adapters, db)
			plugin.Log.Infof("Added status publisher %q from config", pub)
		}
		plugin.Deps.PublishStatistics = publishers
		if config.Mtu != 0 {
			plugin.ifMtu = config.Mtu
			plugin.Log.Info("Default MTU set to %v", plugin.ifMtu)
		}
		plugin.enableStopwatch = config.Stopwatch
		if plugin.enableStopwatch {
			plugin.Log.Infof("stopwatch enabled for %v", plugin.PluginName)
		} else {
			plugin.Log.Infof("stopwatch disabled for %v", plugin.PluginName)
		}
		// return skip (if set) or value from config
		plugin.resyncStrategy = plugin.resolveResyncStrategy(config.Strategy)
		plugin.Log.Infof("VPP resync strategy is set to %v", plugin.resyncStrategy)
	} else {
		plugin.Log.Infof("stopwatch disabled for %v", plugin.PluginName)
		// return skip (if set) or full
		plugin.resyncStrategy = plugin.resolveResyncStrategy(fullResync)
		plugin.Log.Infof("VPP resync strategy config not found, set to %v", plugin.resyncStrategy)
	}

	plugin.fixNilPointers()

	plugin.ifStateNotifications = plugin.Deps.IfStatePub

	// All channels that are used inside of publishIfStateEvents or watchEvents must be created in advance!
	plugin.ifStateChan = make(chan *intf.InterfaceStateNotification, 100)
	plugin.bdStateChan = make(chan *l2plugin.BridgeDomainStateNotification, 100)
	plugin.resyncConfigChan = make(chan datasync.ResyncEvent)
	plugin.resyncStatusChan = make(chan datasync.ResyncEvent)
	plugin.changeChan = make(chan datasync.ChangeEvent)
	plugin.ifIdxWatchCh = make(chan ifaceidx.SwIfIdxDto, 100)
	plugin.bdIdxWatchCh = make(chan bdidx.ChangeDto, 100)
	plugin.linuxIfIdxWatchCh = make(chan ifaceLinux.LinuxIfIndexDto, 100)
	plugin.errorChannel = make(chan ErrCtx, 100)

	// Create plugin context, save cancel function into the plugin handle.
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())

	//FIXME Run the following go routines later than following init*() calls - just before Watch().

	// Run event handler go routines.
	go plugin.publishIfStateEvents(ctx)
	go plugin.publishBdStateEvents(ctx)
	go plugin.watchEvents(ctx)

	// Run error handler.
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
	err = plugin.initL4(ctx)
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

	return nil
}

// Resolves resync strategy. Skip resync flag is also evaluated here and it has priority regardless
// the resync strategy parameter.
func (plugin *Plugin) resolveResyncStrategy(strategy string) string {
	// first check skip resync flag
	if *skipResyncFlag {
		return skipResync
		// else check if strategy is set in configfile
	} else if strategy == fullResync || strategy == optimizeColdStart {
		return strategy
	}
	plugin.Log.Infof("Resync strategy %v is not known, setting up the full resync", strategy)
	return fullResync
}

// fixNilPointers sets noopWriter & nooWatcher for nil dependencies.
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
	plugin.Log.Infof("Init interface plugin")
	// configurator loggers
	ifLogger := plugin.Log.NewLogger("-if-conf")
	ifStateLogger := plugin.Log.NewLogger("-if-state")
	bfdLogger := plugin.Log.NewLogger("-bfd-conf")
	stnLogger := plugin.Log.NewLogger("-stn-conf")
	// Interface indexes
	plugin.swIfIndexes = ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(ifLogger, plugin.PluginName,
		"sw_if_indexes", ifaceidx.IndexMetadata))

	// Get pointer to the map with Linux interface indices.
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

	var stopwatch *measure.Stopwatch
	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("InterfaceConfigurator", ifLogger)
	}
	plugin.ifConfigurator = &ifplugin.InterfaceConfigurator{
		Log:          ifLogger,
		GoVppmux:     plugin.GoVppmux,
		ServiceLabel: plugin.ServiceLabel,
		Linux:        plugin.Linux,
		Stopwatch:    stopwatch,
	}
	plugin.ifConfigurator.Init(plugin.swIfIndexes, plugin.ifMtu, plugin.ifVppNotifChan)

	plugin.Log.Debug("ifConfigurator Initialized")

	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("BFDConfigurator", bfdLogger)
	}
	plugin.bfdConfigurator = &ifplugin.BFDConfigurator{
		Log:          bfdLogger,
		GoVppmux:     plugin.GoVppmux,
		ServiceLabel: plugin.ServiceLabel,
		SwIfIndexes:  plugin.swIfIndexes,
		BfdIDSeq:     1,
		Stopwatch:    stopwatch,
	}
	plugin.bfdConfigurator.Init(plugin.bfdSessionIndexes, plugin.bfdAuthKeysIndexes, plugin.bfdEchoFunctionIndex, BfdRemovedAuthKeys)

	plugin.Log.Debug("bfdConfigurator Initialized")

	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("stnConfigurator", stnLogger)
	}

	plugin.stnAllIndexes = nametoidx.NewNameToIdx(stnLogger, plugin.PluginName, "stn-all-indexes", nil)
	plugin.stnUnstoredIndexes = nametoidx.NewNameToIdx(stnLogger, plugin.PluginName, "stn-unstored-indexes", nil)

	plugin.stnConfigurator = &ifplugin.StnConfigurator{
		Log:                 bfdLogger,
		GoVppmux:            plugin.GoVppmux,
		SwIfIndexes:         plugin.swIfIndexes,
		StnUnstoredIndexes:  plugin.stnUnstoredIndexes,
		StnAllIndexes:       plugin.stnAllIndexes,
		StnUnstoredIndexSeq: 1,
		StnAllIndexSeq:      1,
		Stopwatch:           stopwatch,
	}
	plugin.stnConfigurator.Init()

	plugin.Log.Debug("stnConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initACL(ctx context.Context) error {
	plugin.Log.Infof("Init ACL plugin")
	// logger
	aclLogger := plugin.Log.NewLogger("-acl-plugin")
	var err error
	plugin.aclL3L4Indexes = nametoidx.NewNameToIdx(aclLogger, plugin.PluginName, "acl_l3_l4_indexes", nil)

	plugin.aclL2Indexes = nametoidx.NewNameToIdx(aclLogger, plugin.PluginName, "acl_l2_indexes", nil)

	var stopwatch *measure.Stopwatch
	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("ACLConfigurator", aclLogger)
	}
	plugin.aclConfigurator = &aclplugin.ACLConfigurator{
		Log:            aclLogger,
		GoVppmux:       plugin.GoVppmux,
		ACLL3L4Indexes: plugin.aclL3L4Indexes,
		ACLL2Indexes:   plugin.aclL2Indexes,
		SwIfIndexes:    plugin.swIfIndexes,
		Stopwatch:      stopwatch,
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
	plugin.Log.Infof("Init L2 plugin")
	// loggers
	bdLogger := plugin.Log.NewLogger("-l2-bd-conf")
	bdStateLogger := plugin.Log.NewLogger("-l2-bd-state")
	fibLogger := plugin.Log.NewLogger("-l2-fib-conf")
	xcLogger := plugin.Log.NewLogger("-l2-xc-conf")
	// Bridge domain indices
	plugin.bdIndexes = bdidx.NewBDIndex(nametoidx.NewNameToIdx(bdLogger, plugin.PluginName,
		"bd_indexes", bdidx.IndexMetadata))

	// Interface to bridge domain indices - desired state
	plugin.ifToBdDesIndexes = nametoidx.NewNameToIdx(bdLogger, plugin.PluginName, "if_to_bd_des_indexes", nil)

	// Interface to bridge domain indices - current state

	plugin.ifToBdRealIndexes = nametoidx.NewNameToIdx(bdLogger, plugin.PluginName, "if_to_bd_real_indexes", nil)

	var stopwatch *measure.Stopwatch
	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("BDConfigurator", bdLogger)
	}
	plugin.bdConfigurator = &l2plugin.BDConfigurator{
		Log:                bdLogger,
		GoVppmux:           plugin.GoVppmux,
		SwIfIndexes:        plugin.swIfIndexes,
		BdIndexes:          plugin.bdIndexes,
		BridgeDomainIDSeq:  1,
		IfToBdIndexes:      plugin.ifToBdDesIndexes,
		IfToBdRealStateIdx: plugin.ifToBdRealIndexes,
		Stopwatch:          stopwatch,
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

	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("FIBConfigurator", fibLogger)
	}
	plugin.fibConfigurator = &l2plugin.FIBConfigurator{
		Log:           fibLogger,
		GoVppmux:      plugin.GoVppmux,
		SwIfIndexes:   plugin.swIfIndexes,
		BdIndexes:     plugin.bdIndexes,
		IfToBdIndexes: plugin.ifToBdDesIndexes,
		FibIndexes:    plugin.fibIndexes,
		FibIndexSeq:   1,
		FibDesIndexes: plugin.fibDesIndexes,
		Stopwatch:     stopwatch,
	}

	// L2 xConnect indexes

	plugin.xcIndexes = nametoidx.NewNameToIdx(xcLogger, plugin.PluginName, "xc_indexes", nil)

	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("XConnectConfigurator", xcLogger)
	}
	plugin.xcConfigurator = &l2plugin.XConnectConfigurator{
		Log:         xcLogger,
		GoVppmux:    plugin.GoVppmux,
		SwIfIndexes: plugin.swIfIndexes,
		XcIndexes:   plugin.xcIndexes,
		XcIndexSeq:  1,
		Stopwatch:   stopwatch,
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
	plugin.Log.Infof("Init L3 plugin")
	routeLogger := plugin.Log.NewLogger("-l3-route-conf")
	plugin.routeIndexes = l3idx.NewRouteIndex(
		nametoidx.NewNameToIdx(routeLogger, plugin.PluginName, "route_indexes", nil))
	routeCachedIndexes := l3idx.NewRouteIndex(
		nametoidx.NewNameToIdx(plugin.Log, plugin.PluginName, "route_cached_indexes", nil))

	var stopwatch *measure.Stopwatch
	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("RouteConfigurator", routeLogger)
	}
	plugin.routeConfigurator = &l3plugin.RouteConfigurator{
		Log:              routeLogger,
		GoVppmux:         plugin.GoVppmux,
		RouteIndexes:     plugin.routeIndexes,
		RouteIndexSeq:    1,
		SwIfIndexes:      plugin.swIfIndexes,
		RouteCachedIndex: routeCachedIndexes,
		Stopwatch:        stopwatch,
	}

	arpLogger := plugin.Log.NewLogger("-l3-arp-conf")
	// ARP configuration indices
	plugin.arpIndexes = l3idx.NewARPIndex(nametoidx.NewNameToIdx(arpLogger, plugin.PluginName, "arp_indexes", nil))
	// ARP cache indices
	arpCache := l3idx.NewARPIndex(nametoidx.NewNameToIdx(arpLogger, plugin.PluginName, "arp_cache", nil))
	// ARP deleted indices
	arpDeleted := l3idx.NewARPIndex(nametoidx.NewNameToIdx(arpLogger, plugin.PluginName, "arp_unnasigned", nil))

	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("ArpConfigurator", arpLogger)
	}
	plugin.arpConfigurator = &l3plugin.ArpConfigurator{
		Log:         arpLogger,
		GoVppmux:    plugin.GoVppmux,
		ARPIndexes:  plugin.arpIndexes,
		ARPCache:    arpCache,
		ARPDeleted:  arpDeleted,
		ARPIndexSeq: 1,
		SwIfIndexes: plugin.swIfIndexes,
		Stopwatch:   stopwatch,
	}

	if err := plugin.routeConfigurator.Init(); err != nil {
		return err
	}

	plugin.Log.Debug("routeConfigurator Initialized")

	if err := plugin.arpConfigurator.Init(); err != nil {
		return err
	}
	plugin.Log.Debug("arpConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initL4(ctx context.Context) error {
	plugin.Log.Infof("Init L4 plugin")
	l4Logger := plugin.Log.NewLogger("-l4-plugin")
	plugin.namespaceIndexes = nsidx.NewAppNsIndex(nametoidx.NewNameToIdx(l4Logger, plugin.PluginName,
		"namespace_indexes", nil))
	plugin.notConfAppNsIndexes = nsidx.NewAppNsIndex(nametoidx.NewNameToIdx(l4Logger, plugin.PluginName,
		"not_configured_namespace_indexes", nil))

	var stopwatch *measure.Stopwatch
	if plugin.enableStopwatch {
		stopwatch = measure.NewStopwatch("L4Configurator", l4Logger)
	}
	plugin.l4Configurator = &l4plugin.L4Configurator{
		Log:                l4Logger,
		GoVppmux:           plugin.GoVppmux,
		AppNsIndexes:       plugin.namespaceIndexes,
		NotConfiguredAppNs: plugin.notConfAppNsIndexes,
		AppNsIdxSeq:        1,
		SwIfIndexes:        plugin.swIfIndexes,
		Stopwatch:          stopwatch,
	}
	err := plugin.l4Configurator.Init()
	if err != nil {
		return err
	}

	plugin.Log.Debug("l4Configurator Initialized")

	return nil
}

func (plugin *Plugin) retrieveDPConfig() (*DPConfig, error) {
	config := &DPConfig{}

	found, err := plugin.PluginConfig.GetValue(config)
	if err != nil {
		return nil, err
	} else if !found {
		plugin.Log.Warn("defaultplugins config not found")
		return nil, nil
	}
	plugin.Log.Debugf("defaultplugins config found: %+v", config)

	if pubs := os.Getenv("DP_STATUS_PUBLISHERS"); pubs != "" {
		plugin.Log.Debugf("status publishers from env: %v", pubs)
		config.StatusPublishers = append(config.StatusPublishers, pubs)
	}

	return config, err
}

func (plugin *Plugin) initErrorHandler() error {
	ehLogger := plugin.Log.NewLogger("-error-handler")
	plugin.errorIndexes = nametoidx.NewNameToIdx(ehLogger, plugin.PluginName, "error_indexes", nil)

	// Init mapping index
	plugin.errorIdxSeq = 1
	return nil
}

// AfterInit delegates the call to ifStateUpdater.
func (plugin *Plugin) AfterInit() error {
	plugin.Log.Debug("vpp plugins AfterInit begin")

	err := plugin.ifStateUpdater.AfterInit()
	if err != nil {
		return err
	}

	if plugin.StatusCheck != nil {
		// Register the plugin to status check. Periodical probe is not needed,
		// data change will be reported when changed
		plugin.StatusCheck.Register(plugin.PluginName, nil)
		// Notify that status check for default plugins was registered. It will
		// prevent status report errors in case resync is executed before AfterInit
		plugin.statusCheckReg = true
	}

	plugin.Log.Debug("vpp plugins AfterInit finished successfully")

	return nil
}

// Close cleans up the resources.
func (plugin *Plugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	_, err := safeclose.CloseAll(plugin.watchStatusReg, plugin.watchConfigReg, plugin.changeChan,
		plugin.resyncStatusChan, plugin.resyncConfigChan,
		plugin.ifConfigurator, plugin.ifStateUpdater, plugin.ifVppNotifChan, plugin.errorChannel,
		plugin.bdVppNotifChan, plugin.bdConfigurator, plugin.fibConfigurator, plugin.bfdConfigurator,
		plugin.xcConfigurator, plugin.routeConfigurator, plugin.arpConfigurator)

	return err
}
