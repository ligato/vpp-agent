package defaultplugins

import (
	"context"

	"sync"

	log "github.com/ligato/cn-infra/logging/logrus"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/messaging/kafka"
	"github.com/ligato/cn-infra/messaging/kafka/mux"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/defaultplugins/aclplugin"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/ifaceidx"
	intf "github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/bdidx"
	"github.com/ligato/vpp-agent/defaultplugins/l3plugin"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/linuxplugin"
)

// Plugin implements Plugin interface, therefore it can be loaded with other plugins
type Plugin struct {
	Transport    datasync.TransportAdapter `inject:""`
	ServiceLabel *servicelabel.Plugin
	Kafka        kafka.Mux
	//TODO Kafka PubSub `inject:""` instead of kafkaConn

	aclConfigurator *aclplugin.ACLConfigurator
	aclL3L4Indexes  idxvpp.NameToIdxRW
	aclL2Indexes    idxvpp.NameToIdxRW

	swIfIndexes          ifaceidx.SwIfIndexRW
	linuxIfIndexes       idxvpp.NameToIdx
	ifConfigurator       *ifplugin.InterfaceConfigurator
	ifStateUpdater       *ifplugin.InterfaceStateUpdater
	ifVppNotifChan       chan govppapi.Message
	ifStateChan          chan *intf.InterfaceStateNotification
	bdVppNotifChan       chan l2plugin.BridgeDomainStateMessage
	bdStateUpdater       *l2plugin.BridgeDomainStateUpdater
	bdVppNotifChan       chan govppapi.Message
	bdStateChan          chan *l2plugin.BridgeDomainStateNotification
	bfdSessionIndexes    idxvpp.NameToIdxRW
	bfdAuthKeysIndexes   idxvpp.NameToIdxRW
	bfdEchoFunctionIndex idxvpp.NameToIdxRW

	bdConfigurator    *l2plugin.BDConfigurator
	fibConfigurator   *l2plugin.FIBConfigurator
	xcConfigurator    *l2plugin.XConnectConfigurator
	bdIndexes         bdidx.BDIndexRW
	ifToBdDesIndexes  idxvpp.NameToIdxRW
	ifToBdRealIndexes idxvpp.NameToIdxRW
	fibIndexes        idxvpp.NameToIdxRW
	fibDesIndexes     idxvpp.NameToIdxRW
	xcIndexes         idxvpp.NameToIdxRW
	errorIndexes      idxvpp.NameToIdxRW
	ifIdxWatchCh      chan ifaceidx.SwIfIdxDto
	bdIdxWatchCh      chan bdidx.ChangeDto
	linuxIfIdxWatchCh chan idxvpp.NameToIdxDto

	routeConfigurator *l3plugin.RouteConfigurator

	resyncConfigChan chan datasync.ResyncEvent
	resyncStatusChan chan datasync.ResyncEvent
	changeChan       chan datasync.ChangeEvent //TODO dedicated type abstracted from ETCD
	kafkaConn        *mux.Connection

	watchConfigReg datasync.WatchDataRegistration
	watchStatusReg datasync.WatchDataRegistration

	errorChannel chan ErrCtx
	errorIdxSeq  uint32

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
		log.Panic("Trying to access the Interface Plugin but it is still not initialized")
	}
	return gPlugin
}

// Init gets handlers for ETCD, Kafka and delegates them to ifConfigurator & ifStateUpdater
func (plugin *Plugin) Init() error {
	plugin.Transport = datasync.GetTransport()

	log.Debug("Initializing interface plugin")

	if plugin.Kafka != nil {
		plugin.kafkaConn = plugin.Kafka.NewConnection(string(PluginID))
	}

	// all channels that are used inside of publishIfStateEvents or watchEvents must be created in advance!
	plugin.ifStateChan = make(chan *intf.InterfaceStateNotification, 100)
	plugin.bdStateChan = make(chan *l2plugin.BridgeDomainStateNotification, 100)
	plugin.resyncConfigChan = make(chan datasync.ResyncEvent)
	plugin.resyncStatusChan = make(chan datasync.ResyncEvent)
	plugin.changeChan = make(chan datasync.ChangeEvent)
	plugin.ifIdxWatchCh = make(chan ifaceidx.SwIfIdxDto, 100)
	plugin.bdIdxWatchCh = make(chan bdidx.ChangeDto, 100)
	plugin.linuxIfIdxWatchCh = make(chan idxvpp.NameToIdxDto, 100)
	plugin.errorChannel = make(chan ErrCtx, 100)

	// create plugin context, save cancel function into the plugin handle
	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())

	// run event handler go routines
	go plugin.publishIfStateEvents(ctx)
	go plugin.publishBdStateEvents(ctx)
	go plugin.watchEvents(ctx)

	// run error handler
	go plugin.changePropagateError()

	err := plugin.initIF(ctx)
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

func (plugin *Plugin) initIF(ctx context.Context) error {
	// Interface indexes
	plugin.swIfIndexes = ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logroot.Logger(), PluginID,
		"sw_if_indexes", ifaceidx.IndexMetadata))

	// get pointer to the map with Linux interface indexes
	plugin.linuxIfIndexes = linuxplugin.GetIfIndexes()

	// BFD session
	plugin.bfdSessionIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "bfd_session_indexes", nil)

	// BFD key
	plugin.bfdAuthKeysIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "bfd_auth_keys_indexes", nil)

	// BFD echo function
	plugin.bfdEchoFunctionIndex = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "bfd_echo_function_index", nil)

	// BFD echo function
	BfdRemovedAuthKeys := nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "bfd_removed_auth_keys", nil)

	plugin.ifVppNotifChan = make(chan govppapi.Message, 100)
	plugin.ifStateUpdater = &ifplugin.InterfaceStateUpdater{}
	plugin.ifStateUpdater.Init(ctx, plugin.swIfIndexes, plugin.ifVppNotifChan, func(state *intf.InterfaceStateNotification) {
		select {
		case plugin.ifStateChan <- state:
			// OK
		default:
			log.Debug("Unable to send to the ifState channel - channel buffer full.")
		}
	})

	log.Debug("ifStateUpdater Initialized")

	plugin.ifConfigurator = &ifplugin.InterfaceConfigurator{ServiceLabel: plugin.ServiceLabel}
	plugin.ifConfigurator.Init(plugin.swIfIndexes, plugin.ifVppNotifChan)

	log.Debug("ifConfigurator Initialized")

	plugin.bfdConfigurator = &ifplugin.BFDConfigurator{
		ServiceLabel: plugin.ServiceLabel,
		SwIfIndexes:  plugin.swIfIndexes,
		BfdIDSeq:     1,
	}
	plugin.bfdConfigurator.Init(plugin.bfdSessionIndexes, plugin.bfdAuthKeysIndexes, plugin.bfdEchoFunctionIndex, BfdRemovedAuthKeys)

	log.Debug("bfdConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initACL(ctx context.Context) error {
	var err error
	plugin.aclL3L4Indexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "acl_l3_l4_indexes", nil)

	plugin.aclL2Indexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "acl_l2_indexes", nil)

	plugin.aclConfigurator = &aclplugin.ACLConfigurator{
		ACLL3L4Indexes: plugin.aclL3L4Indexes,
		ACLL2Indexes:   plugin.aclL2Indexes,
		SwIfIndexes:    plugin.swIfIndexes,
	}

	// Init ACL plugin
	err = plugin.aclConfigurator.Init()
	if err != nil {
		return err
	}
	log.Debug("aclConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initL2(ctx context.Context) error {
	// Bridge domain indexes
	plugin.bdIndexes = bdidx.NewBDIndex(nametoidx.NewNameToIdx(logroot.Logger(), PluginID,
		"bd_indexes", bdidx.IndexMetadata))

	// Interface to bridge domain indexes - desired state
	plugin.ifToBdDesIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "if_to_bd_des_indexes", nil)

	// Interface to bridge domain indexes - current state

	plugin.ifToBdRealIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "if_to_bd_real_indexes", nil)

	plugin.bdConfigurator = &l2plugin.BDConfigurator{
		SwIfIndexes:        plugin.swIfIndexes,
		BdIndexes:          plugin.bdIndexes,
		BridgeDomainIDSeq:  1,
		IfToBdIndexes:      plugin.ifToBdDesIndexes,
		IfToBdRealStateIdx: plugin.ifToBdRealIndexes,
	}

	plugin.bdVppNotifChan = make(chan l2plugin.BridgeDomainStateMessage, 100)
	plugin.bdStateUpdater = &l2plugin.BridgeDomainStateUpdater{}
	plugin.bdStateUpdater.Init(ctx, plugin.bdIndexes, plugin.swIfIndexes, plugin.bdVppNotifChan, func(state *l2plugin.BridgeDomainStateNotification) {
		select {
		case plugin.bdStateChan <- state:
			// OK
		default:
			log.Debug("Unable to send to the bdState channel: buffer is full.")
		}
	})

	// FIB indexes
	plugin.fibIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "fib_indexes", nil)

	plugin.fibConfigurator = &l2plugin.FIBConfigurator{
		SwIfIndexes:   plugin.swIfIndexes,
		BdIndexes:     plugin.bdIndexes,
		IfToBdIndexes: plugin.ifToBdDesIndexes,
		FibIndexes:    plugin.fibIndexes,
		FibIndexSeq:   1,
		FibDesIndexes: plugin.fibDesIndexes,
	}

	// L2 xConnect indexes

	plugin.xcIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "xc_indexes", nil)

	plugin.xcConfigurator = &l2plugin.XConnectConfigurator{
		SwIfIndexes: plugin.swIfIndexes,
		XcIndexes:   plugin.xcIndexes,
		XcIndexSeq:  1,
	}

	// Init
	err := plugin.bdConfigurator.Init(plugin.bdVppNotifChan)
	if err != nil {
		return err
	}

	log.Debug("bdConfigurator Initialized")

	err = plugin.fibConfigurator.Init()
	if err != nil {
		return err
	}

	log.Debug("fibConfigurator Initialized")

	err = plugin.xcConfigurator.Init()
	if err != nil {
		return err
	}

	log.Debug("xcConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initL3(ctx context.Context) error {
	plugin.routeConfigurator = &l3plugin.RouteConfigurator{
		SwIfIndexes: plugin.swIfIndexes,
	}
	err := plugin.routeConfigurator.Init()
	if err != nil {
		return err
	}

	log.Debug("routeConfigurator Initialized")

	return nil
}

func (plugin *Plugin) initErrorHandler() error {

	plugin.errorIndexes = nametoidx.NewNameToIdx(logroot.Logger(), PluginID, "error_indexes", nil)

	// Init mapping index
	plugin.errorIdxSeq = 1
	return nil
}

// AfterInit delegates to ifStateUpdater
func (plugin *Plugin) AfterInit() error {
	log.Debug("vpp plugins AfterInit begin")

	//err := plugin.ifConfigurator.AfterInit()
	err := plugin.ifStateUpdater.AfterInit()
	if err != nil {
		return err
	}

	log.Debug("vpp plugins AfterInit finished succesfully")

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
