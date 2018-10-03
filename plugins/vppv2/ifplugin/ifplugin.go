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

//go:generate descriptor-adapter --descriptor-name Interface  --value-type *interfaces.Interface --meta-type *ifaceidx.IfaceMetadata --import "../model/interfaces" --import "ifaceidx" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name Unnumbered  --value-type *interfaces.Interface_Unnumbered --import "../model/interfaces" --output-dir "descriptor"

package ifplugin

import (
	"context"
	"os"
	"sync"

	"github.com/go-errors/errors"
	govppapi "git.fd.io/govpp.git/api"

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/utils/safeclose"

	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	linux_ifcalls "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/linuxcalls"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/dhcp"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

const (
	// vppStatusPublishersEnv is the name of the environment variable used to
	// override state publishers from the configuration file.
	vppStatusPublishersEnv = "VPP_STATUS_PUBLISHERS"
)

var (
	// noopWriter (no operation writer) helps avoiding NIL pointer based segmentation fault.
	// It is used as default if some dependency was not injected.
	noopWriter = datasync.KVProtoWriters{}

	// noopWatcher (no operation watcher) helps avoiding NIL pointer based segmentation fault.
	// It is used as default if some dependency was not injected.
	noopWatcher = datasync.KVProtoWatchers{}
)

// IfPlugin configures VPP interfaces using GoVPP.
type IfPlugin struct {
	Deps

	// GoVPP
	vppCh     govppapi.Channel

	// handlers
	ifHandler      vppcalls.IfVppAPI
	linuxIfHandler linux_ifcalls.NetlinkAPI

	// descriptors
	ifDescriptor   *descriptor.InterfaceDescriptor
	unIfDescriptor *descriptor.UnnumberedIfDescriptor
	dhcpDescriptor *descriptor.DHCPDescriptor

	// index maps
	intfIndex ifaceidx.IfaceMetadataIndex
	dhcpIndex idxmap.NamedMapping

	// From config file
	enableStopwatch bool
	stopwatch       *measure.Stopwatch // timer used to measure and store time
	defaultMtu      uint32

	// state data
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

// Deps lists dependencies of the interface p.
type Deps struct {
	infra.PluginDeps
	Scheduler     scheduler.KVScheduler
	GoVppmux      govppmux.API

	/* optional, provide if AFPacket or TAP+AUTO_TAP interfaces are used */
	LinuxIfPlugin linux_ifplugin.API

	// state publishing
	StatusCheck       statuscheck.PluginStatusWriter
	PublishErrors     datasync.KeyProtoValWriter
	Watcher           datasync.KeyValProtoWatcher /* for resync of interface state data (PublishStatistics) */
	NotifyStatistics  datasync.KeyProtoValWriter  /* e.g. Kafka */
	PublishStatistics datasync.KeyProtoValWriter  /* e.g. ETCD (with resync) */
	DataSyncs         map[string]datasync.KeyProtoValWriter /* available DBs for PublishStatistics */
	// TODO: GRPCSvc           rpc.GRPCService
}

// Config holds the vpp-plugin configuration.
type Config struct {
	Mtu              uint32   `json:"mtu"`
	Stopwatch        bool     `json:"stopwatch"`
	StatusPublishers []string `json:"status-publishers"`
}

// Init loads configuration file and registers interface-related descriptors.
func (p *IfPlugin) Init() error {
	var err error
	// Read config file and set all related fields
	p.fromConfigFile()

	// Fills nil dependencies with default values
	p.fixNilPointers()

	// Plugin-wide stopwatch instance
	if p.enableStopwatch {
		p.stopwatch = measure.NewStopwatch(string(p.PluginName), p.Log)
	}

	// VPP channel
	if p.vppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API channel: %v", err)
	}

	// init handlers
	p.ifHandler = vppcalls.NewIfVppHandler(p.vppCh, p.Log, p.stopwatch)
	if p.LinuxIfPlugin != nil {
		p.linuxIfHandler = linux_ifcalls.NewNetLinkHandler(p.stopwatch)
	}

	// init descriptors
	p.ifDescriptor = descriptor.NewInterfaceDescriptor(p.ifHandler, p.defaultMtu,
		p.linuxIfHandler, p.LinuxIfPlugin, p.Log)
	ifDescriptor := adapter.NewInterfaceDescriptor(p.ifDescriptor.GetDescriptor())
	p.unIfDescriptor = descriptor.NewUnnumberedIfDescriptor(p.ifHandler, p.Log)
	unIfDescriptor := adapter.NewUnnumberedDescriptor(p.unIfDescriptor.GetDescriptor())
	p.dhcpDescriptor = descriptor.NewDHCPDescriptor(p.Scheduler, p.ifHandler, p.Log)
	dhcpDescriptor := p.dhcpDescriptor.GetDescriptor()

	// register descriptors
	p.Deps.Scheduler.RegisterKVDescriptor(ifDescriptor)
	p.Deps.Scheduler.RegisterKVDescriptor(unIfDescriptor)
	p.Deps.Scheduler.RegisterKVDescriptor(dhcpDescriptor)

	// obtain read-only reference to index map
	var withIndex bool
	metadataMap := p.Deps.Scheduler.GetMetadataMap(ifDescriptor.Name)
	p.intfIndex, withIndex = metadataMap.(ifaceidx.IfaceMetadataIndex)
	if !withIndex {
		return errors.New("missing index with interface metadata")
	}
	p.dhcpIndex = p.Deps.Scheduler.GetMetadataMap(dhcpDescriptor.Name)

	// pass read-only index map to descriptors
	p.ifDescriptor.SetInterfaceIndex(p.intfIndex)
	p.unIfDescriptor.SetInterfaceIndex(p.intfIndex)
	p.dhcpDescriptor.SetInterfaceIndex(p.intfIndex)

	// start watching for DHCP notifications
	dhcpChan := make(chan govppapi.Message, 1)
	if _, err := p.vppCh.SubscribeNotification(dhcpChan, &dhcp.DHCPComplEvent{}); err != nil {
		return err
	}
	p.dhcpDescriptor.StartWatchingDHCP(dhcpChan)

	// Create plugin context, save cancel function into the plugin handle.
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// subscribe & watch for resync of interface state data
	p.resyncStatusChan = make(chan datasync.ResyncEvent)
	p.watchStatusReg, err = p.Watcher.
		Watch("VPP interface state data", nil, p.resyncStatusChan,
		interfaces.StatePrefix)
	if err != nil {
		return err
	}
	go p.watchStatusEvents()

	// start interface state publishing
	go p.publishIfStateEvents()

	// start interface state updater
	p.ifStateChan = make(chan *interfaces.InterfaceNotification, 100)
	// Interface state updater
	p.ifStateUpdater = &InterfaceStateUpdater{}
	if err := p.ifStateUpdater.Init(p.ctx, p.Log, p.GoVppmux, p.intfIndex, func(state *interfaces.InterfaceNotification) {
		select {
		case p.ifStateChan <- state:
			// OK
		default:
			p.Log.Debug("Unable to send to the ifStateNotifications channel - channel buffer full.")
		}
	}); err != nil {
		return p.ifStateUpdater.LogError(err)
	}

	p.Log.Debug("ifStateUpdater Initialized")

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

	// stop watching of DHCP notifications
	p.dhcpDescriptor.StopWatchingDHCP()

	// close all channels
	safeclose.CloseAll(p.resyncStatusChan, p.ifStateChan)
	return nil
}

// GetInterfaceIndex gives read-only access to map with metadata of all configured
// VPP interfaces.
func (p *IfPlugin) GetInterfaceIndex() ifaceidx.IfaceMetadataIndex {
	return p.intfIndex
}

// GetDHCPIndex gives read-only access to (untyped) map with DHCP leases.
// Cast metadata to "github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces".DHCPLease
func (p *IfPlugin) GetDHCPIndex() idxmap.NamedMapping {
	return p.dhcpIndex
}

// fromConfigFile loads plugin attributes from the configuration file.
func (p *IfPlugin) fromConfigFile() {
	config, err := p.loadConfig()
	if err != nil {
		p.Log.Errorf("Error reading %v config file: %v", p.PluginName, err)
		return
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
		if config.Mtu != 0 {
			p.defaultMtu = config.Mtu
			p.Log.Infof("Default MTU set to %v", p.defaultMtu)
		}

		if config.Stopwatch {
			p.enableStopwatch = true
			p.Log.Info("stopwatch enabled for %v", p.PluginName)
		} else {
			p.Log.Info("stopwatch disabled for %v", p.PluginName)
		}
	} else {
		p.Log.Infof("stopwatch disabled for %v", p.PluginName)
	}
}

// loadConfig loads configuration file.
func (p *IfPlugin) loadConfig() (*Config, error) {
	config := &Config{}

	found, err := p.Cfg.LoadValue(config)
	if err != nil {
		return nil, err
	} else if !found {
		p.Log.Debugf("%v config not found", p.PluginName)
		return nil, nil
	}
	p.Log.Debugf("%v config found: %+v", p.PluginName, config)

	if pubs := os.Getenv(vppStatusPublishersEnv); pubs != "" {
		p.Log.Debugf("status publishers from env: %v", pubs)
		config.StatusPublishers = append(config.StatusPublishers, pubs)
	}

	return config, err
}

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
	if p.Deps.NotifyStatistics == nil {
		p.Deps.NotifyStatistics = noopWriter
		p.Log.Debug("setting default noop writer for NotifyStatistics dependency")
	}
	if p.Deps.Watcher == nil {
		p.Deps.Watcher = noopWatcher
		p.Log.Debug("setting default noop watcher for Watcher dependency")
	}
}