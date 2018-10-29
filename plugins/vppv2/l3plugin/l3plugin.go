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

//go:generate descriptor-adapter --descriptor-name StaticRoute --value-type *l3.StaticRoute --import "../model/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name ARPEntry --value-type *l3.ARPEntry --import "../model/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name ProxyARP --value-type *l3.ProxyARP --import "../model/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name ProxyARPInterface --value-type *l3.ProxyARP_Interface --import "../model/l3" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name IPScanNeighbor --value-type *l3.IPScanNeighbor --import "../model/l3" --output-dir "descriptor"

package l3plugin

import (
	"context"
	"sync"

	govppapi "git.fd.io/govpp.git/api"

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vppv2/l3plugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/l3plugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/l3plugin/vppcalls"
	"github.com/pkg/errors"
)

// L3Plugin configures Linux routes and ARP entries using Netlink API.
type L3Plugin struct {
	Deps

	// From configuration file
	disabled bool

	// GoVPP channels
	vppCh govppapi.Channel

	// VPP handlers
	routeHandler    vppcalls.RouteVppAPI
	arpandler       vppcalls.ArpVppAPI
	proxyArpHandler vppcalls.ProxyArpVppAPI
	ipNeigh         vppcalls.IPNeighVppAPI

	// descriptors
	routeDescriptor          *descriptor.RouteDescriptor
	arpDescriptor            *descriptor.ArpDescriptor
	proxyArpDescriptor       *descriptor.ProxyArpDescriptor
	proxyArpIfaceDescriptor  *descriptor.ProxyArpInterfaceDescriptor
	ipScanNeighborDescriptor *descriptor.IPScanNeighborDescriptor

	// go routine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Deps lists dependencies of the interface p.
type Deps struct {
	infra.PluginDeps
	Scheduler scheduler.KVScheduler
	GoVppmux  govppmux.API
	IfPlugin  ifplugin.API
}

// Config holds the l3plugin configuration.
type Config struct {
	Disabled bool `json:"disabled"`
}

// Init initializes and registers descriptors for Linux ARPs and Routes.
func (p *L3Plugin) Init() error {
	// parse configuration file
	config, err := p.retrieveConfig()
	if err != nil {
		return err
	}
	if config != nil {
		if config.Disabled {
			p.disabled = true
			p.Log.Infof("Disabling Linux L3 plugin")
			return nil
		}
	}

	// GoVPP channels
	if p.vppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API channel: %v", err)
	}

	// init handlers
	p.routeHandler = vppcalls.NewRouteVppHandler(p.vppCh, p.IfPlugin.GetInterfaceIndex(), nil)
	p.arpandler = vppcalls.NewArpVppHandler(p.vppCh, p.IfPlugin.GetInterfaceIndex(), nil)
	p.proxyArpHandler = vppcalls.NewProxyArpVppHandler(p.vppCh, p.IfPlugin.GetInterfaceIndex(), nil)
	p.ipNeigh = vppcalls.NewIPNeighVppHandler(p.vppCh, nil)

	// init & register descriptors
	routeDescriptor := adapter.NewStaticRouteDescriptor(descriptor.NewRouteDescriptor(
		p.Scheduler, p.routeHandler, p.Log).GetDescriptor())
	arpDescriptor := adapter.NewARPEntryDescriptor(descriptor.NewArpDescriptor(
		p.Scheduler, p.arpandler, p.Log).GetDescriptor())
	proxyArpDescriptor := adapter.NewProxyARPDescriptor(descriptor.NewProxyArpDescriptor(
		p.Scheduler, p.proxyArpHandler, p.Log).GetDescriptor())
	proxyArpIfaceDescriptor := adapter.NewProxyARPInterfaceDescriptor(descriptor.NewProxyArpInterfaceDescriptor(
		p.Scheduler, p.proxyArpHandler, p.Log).GetDescriptor())
	ipScanNeighborDescriptor := adapter.NewIPScanNeighborDescriptor(descriptor.NewIPScanNeighborDescriptor(
		p.Scheduler, p.ipNeigh, p.Log).GetDescriptor())

	p.Deps.Scheduler.RegisterKVDescriptor(routeDescriptor)
	p.Deps.Scheduler.RegisterKVDescriptor(arpDescriptor)
	p.Deps.Scheduler.RegisterKVDescriptor(proxyArpDescriptor)
	p.Deps.Scheduler.RegisterKVDescriptor(proxyArpIfaceDescriptor)
	p.Deps.Scheduler.RegisterKVDescriptor(ipScanNeighborDescriptor)

	return nil
}

// Close does nothing here.
func (p *L3Plugin) Close() error {
	return nil
}

// retrieveConfig loads L3Plugin configuration file.
func (p *L3Plugin) retrieveConfig() (*Config, error) {
	config := &Config{}
	found, err := p.Cfg.LoadValue(config)
	if !found {
		p.Log.Debug("Linux L3Plugin config not found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Log.Debug("Linux L3Plugin config found")
	return config, err
}
