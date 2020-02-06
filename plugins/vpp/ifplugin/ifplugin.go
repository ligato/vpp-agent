// Copyright (c) 2018 Cisco and/or its affiliates.
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

//go:generate descriptor-adapter --descriptor-name Interface  --value-type *vpp_interfaces.Interface --meta-type *ifaceidx.IfaceMetadata --import "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx" --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name Unnumbered  --value-type *vpp_interfaces.Interface_Unnumbered --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name RxMode  --value-type *vpp_interfaces.Interface --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name RxPlacement  --value-type *vpp_interfaces.Interface_RxPlacement --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name BondedInterface  --value-type *vpp_interfaces.BondLink_BondedInterface --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name Span  --value-type *vpp_interfaces.Span --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces" --output-dir "descriptor"

package ifplugin

import (
	"context"
	"sync"
	"time"

	"github.com/ligato/cn-infra/servicelabel"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	linux_ifcalls "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	vppclient "go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp2001"
)

func init() {
	kvscheduler.AddNonRetryableError(vppclient.ErrPluginDisabled)
}

// Default Go routine count used while retrieving linux configuration
const goRoutineCount = 10

// IfPlugin configures VPP interfaces using GoVPP.
type IfPlugin struct {
	Deps

	// handlers
	ifHandler      vppcalls.InterfaceVppAPI
	linuxIfHandler linux_ifcalls.NetlinkAPIRead

	// index maps
	intfIndex ifaceidx.IfaceMetadataIndex
	dhcpIndex idxmap.NamedMapping

	// descriptors
	linkStateDescriptor *descriptor.LinkStateDescriptor
	dhcpDescriptor      *descriptor.DHCPDescriptor
	spanDescriptor      *descriptor.SpanDescriptor

	// from config file
	defaultMtu uint32

	// state data
	publishStats     bool
	publishLock      sync.Mutex
	statusCheckReg   bool
	watchStatusReg   datasync.WatchRegistration
	resyncStatusChan chan datasync.ResyncEvent
	ifStateChan      chan *interfaces.InterfaceNotification
	ifStateUpdater   *InterfaceStateUpdater

	// go routine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Deps lists dependencies of the interface plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler  kvs.KVScheduler
	VPP          govppmux.API
	ServiceLabel servicelabel.ReaderAPI
	AddrAlloc    netalloc.AddressAllocator
	/*	LinuxIfPlugin and NsPlugin deps are optional,
		but they are required if AFPacket or TAP+TAP_TO_VPP interfaces are used. */
	LinuxIfPlugin descriptor.LinuxPluginAPI
	NsPlugin      nsplugin.API
	// state publishing
	StatusCheck       statuscheck.PluginStatusWriter
	PublishErrors     datasync.KeyProtoValWriter            // TODO: to be used with a generic plugin for publishing errors (not just interfaces and BDs)
	Watcher           datasync.KeyValProtoWatcher           /* for resync of interface state data (PublishStatistics) */
	NotifyStates      datasync.KeyProtoValWriter            /* e.g. Kafka (up/down events only)*/
	PublishStatistics datasync.KeyProtoValWriter            /* e.g. ETCD (with resync) */
	DataSyncs         map[string]datasync.KeyProtoValWriter /* available DBs for PublishStatistics */
	PushNotification  func(notification *vpp.Notification)
}

// Init loads configuration file and registers interface-related descriptors.
func (p *IfPlugin) Init() (err error) {
	// Create plugin context, save cancel function into the plugin handle.
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Read config file and set all related fields
	if err := p.fromConfigFile(); err != nil {
		return err
	}

	// Fills nil dependencies with default values
	p.publishStats = p.PublishStatistics != nil || p.NotifyStates != nil
	p.fixNilPointers()

	// Init handlers
	p.ifHandler = vppcalls.CompatibleInterfaceVppHandler(p.VPP, p.Log)
	if p.ifHandler == nil {
		return errors.New("interface VPP handler is not available")
	}

	if p.LinuxIfPlugin != nil {
		p.linuxIfHandler = linux_ifcalls.NewNetLinkHandler(
			p.NsPlugin,
			p.LinuxIfPlugin.GetInterfaceIndex(),
			p.ServiceLabel.GetAgentPrefix(),
			goRoutineCount, p.Log,
		)
	}

	// Init descriptors

	//   -> base interface descriptor
	ifaceDescriptor, ifaceDescrCtx := descriptor.NewInterfaceDescriptor(p.ifHandler,
		p.AddrAlloc, p.defaultMtu, p.linuxIfHandler, p.LinuxIfPlugin, p.NsPlugin, p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(ifaceDescriptor)
	if err != nil {
		return err
	}
	var withIndex bool
	metadataMap := p.KVScheduler.GetMetadataMap(ifaceDescriptor.Name)
	p.intfIndex, withIndex = metadataMap.(ifaceidx.IfaceMetadataIndex)
	if !withIndex {
		return errors.New("missing index with interface metadata")
	}
	ifaceDescrCtx.SetInterfaceIndex(p.intfIndex)

	//   -> descriptors for derived values / notifications
	var (
		linkStateDescriptor *kvs.KVDescriptor
		dhcpDescriptor      *kvs.KVDescriptor
	)
	dhcpDescriptor, p.dhcpDescriptor = descriptor.NewDHCPDescriptor(p.KVScheduler,
		p.ifHandler, p.intfIndex, p.Log)
	linkStateDescriptor, p.linkStateDescriptor = descriptor.NewLinkStateDescriptor(
		p.KVScheduler, p.ifHandler, p.intfIndex, p.Log)

	rxModeDescriptor := descriptor.NewRxModeDescriptor(p.ifHandler, p.intfIndex, p.Log)
	rxPlacementDescriptor := descriptor.NewRxPlacementDescriptor(p.ifHandler, p.intfIndex, p.Log)
	addrDescriptor := descriptor.NewInterfaceAddressDescriptor(p.ifHandler, p.AddrAlloc, p.intfIndex, p.Log)
	unIfDescriptor := descriptor.NewUnnumberedIfDescriptor(p.ifHandler, p.intfIndex, p.Log)
	bondIfDescriptor, _ := descriptor.NewBondedInterfaceDescriptor(p.ifHandler, p.intfIndex, p.Log)
	vrfDescriptor := descriptor.NewInterfaceVrfDescriptor(p.ifHandler, p.intfIndex, p.Log)
	withAddrDescriptor := descriptor.NewInterfaceWithAddrDescriptor(p.Log)
	spanDescriptor, spanDescriptorCtx := descriptor.NewSpanDescriptor(p.ifHandler, p.Log)
	spanDescriptorCtx.SetInterfaceIndex(p.intfIndex)

	err = p.KVScheduler.RegisterKVDescriptor(
		dhcpDescriptor,
		linkStateDescriptor,
		rxModeDescriptor,
		rxPlacementDescriptor,
		addrDescriptor,
		unIfDescriptor,
		bondIfDescriptor,
		vrfDescriptor,
		withAddrDescriptor,
		spanDescriptor,
	)
	if err != nil {
		return err
	}

	// start watching for DHCP notifications
	p.dhcpIndex = p.KVScheduler.GetMetadataMap(dhcpDescriptor.Name)
	if p.dhcpIndex == nil {
		return errors.New("missing index with DHCP metadata")
	}
	p.dhcpDescriptor.WatchDHCPNotifications(p.ctx)

	// interface state data
	if p.publishStats {
		// subscribe & watch for resync of interface state data
		p.resyncStatusChan = make(chan datasync.ResyncEvent)

		p.wg.Add(1)
		go p.watchStatusEvents()
	}

	// start interface state updater
	p.ifStateChan = make(chan *interfaces.InterfaceNotification, 1000)

	// start interface state publishing
	p.wg.Add(1)
	go p.publishIfStateEvents()

	// Interface state updater
	p.ifStateUpdater = &InterfaceStateUpdater{}

	var n int
	var t time.Time
	ifNotifHandler := func(state *interfaces.InterfaceNotification) {
		select {
		case p.ifStateChan <- state:
			// OK
		default:
			// full
			if time.Since(t) > time.Second {
				p.Log.Debugf("ifStateChan channel is full (%d)", n)
				n = 0
			} else {
				n++
			}
			t = time.Now()
		}
	}

	err = p.ifStateUpdater.Init(p.ctx, p.Log, p.KVScheduler, p.VPP, p.intfIndex,
		ifNotifHandler, p.publishStats)
	if err != nil {
		return err
	}

	if p.publishStats {
		if err = p.subscribeWatcher(); err != nil {
			return err
		}
	}

	return nil
}

func (p *IfPlugin) subscribeWatcher() (err error) {
	keyPrefixes := []string{interfaces.StatePrefix}

	p.Log.Debugf("subscribe to %d status prefixes: %v", len(keyPrefixes), keyPrefixes)

	p.watchStatusReg, err = p.Watcher.Watch("vpp-if-state",
		nil, p.resyncStatusChan, keyPrefixes...)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit delegates the call to ifStateUpdater.
func (p *IfPlugin) AfterInit() error {
	err := p.ifStateUpdater.AfterInit()
	if err != nil {
		return err
	}

	if p.StatusCheck != nil {
		// Register the plugin to status check. Periodical probe is not needed,
		// data change will be reported when changed
		p.StatusCheck.Register(p.PluginName, nil)
		// Notify that status check for the plugins was registered. It will
		// prevent status report errors in case resync is executed before AfterInit.
		p.statusCheckReg = true
	}

	return nil
}

// Close stops all go routines.
func (p *IfPlugin) Close() error {
	// stop publishing of state data
	p.cancel()
	p.wg.Wait()

	// close all resources
	return safeclose.Close(
		// DHCP descriptor (DHCP notification watcher)
		p.dhcpDescriptor,
		// state updater
		p.ifStateUpdater,
		// registrations
		p.watchStatusReg)
}

// GetInterfaceIndex gives read-only access to map with metadata of all configured
// VPP interfaces.
func (p *IfPlugin) GetInterfaceIndex() ifaceidx.IfaceMetadataIndex {
	return p.intfIndex
}

// GetDHCPIndex gives read-only access to (untyped) map with DHCP leases.
// Cast metadata to "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces".DHCPLease
func (p *IfPlugin) GetDHCPIndex() idxmap.NamedMapping {
	return p.dhcpIndex
}

// SetNotifyService sets notification callback for processing VPP notifications.
func (p *IfPlugin) SetNotifyService(notify func(notification *vpp.Notification)) {
	p.PushNotification = notify
}

// fromConfigFile loads plugin attributes from the configuration file.
func (p *IfPlugin) fromConfigFile() error {
	config, err := p.loadConfig()
	if err != nil {
		p.Log.Errorf("Error reading %v config file: %v", p.PluginName, err)
		return err
	}
	if config != nil {
		publishers := datasync.KVProtoWriters{}
		for _, pub := range config.StatusPublishers {
			db, found := p.Deps.DataSyncs[pub]
			if !found {
				p.Log.Warnf("Unknown status publisher %q from config", pub)
				continue
			}
			publishers = append(publishers, db)
			p.Log.Infof("Added status publisher %q from config", pub)
		}
		p.Deps.PublishStatistics = publishers
		if config.MTU != 0 {
			p.defaultMtu = config.MTU
			p.Log.Infof("Default MTU set to %v", p.defaultMtu)
		}
	}
	return nil
}

var (
	// noopWriter (no operation writer) helps avoiding NIL pointer based segmentation fault.
	// It is used as default if some dependency was not injected.
	noopWriter = datasync.KVProtoWriters{}

	// noopWatcher (no operation watcher) helps avoiding NIL pointer based segmentation fault.
	// It is used as default if some dependency was not injected.
	noopWatcher = datasync.KVProtoWatchers{}
)

// fixNilPointers sets noopWriter & nooWatcher for nil dependencies.
func (p *IfPlugin) fixNilPointers() {
	if p.Deps.PublishErrors == nil {
		p.Deps.PublishErrors = noopWriter
		p.Log.Debug("setting default noop writer for PublishErrors dependency")
	}
	if p.Deps.PublishStatistics == nil {
		p.Deps.PublishStatistics = noopWriter
		p.Log.Debug("setting default noop writer for PublishStatistics dependency")
	}
	if p.Deps.NotifyStates == nil {
		p.Deps.NotifyStates = noopWriter
		p.Log.Debug("setting default noop writer for NotifyStatistics dependency")
	}
	if p.Deps.Watcher == nil {
		p.Deps.Watcher = noopWatcher
		p.Log.Debug("setting default noop watcher for Watcher dependency")
	}
}
