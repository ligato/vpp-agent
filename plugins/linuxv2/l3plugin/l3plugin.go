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

//go:generate protoc --proto_path=../model/l3 --proto_path=${GOPATH}/src --gogo_out=../model/l3 l3.proto
//go:generate adapter-generator --descriptor-name ARP --is-proto --value-type *l3.LinuxStaticARPEntry --from-datasync --import "../model/l3" --output-dir "descriptor"
//go:generate adapter-generator --descriptor-name Route --is-proto --value-type *l3.LinuxStaticRoute --from-datasync --import "../model/l3" --output-dir "descriptor"

package l3plugin

import (
	"github.com/ligato/cn-infra/infra"
	scheduler "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/logging/measure"

	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	"github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin"
)

// L3Plugin configures Linux routes and ARP entries using Netlink API.
type L3Plugin struct {
	Deps

	// From configuration file
	disabled  bool
	stopwatch *measure.Stopwatch

	// system handlers
	l3Handler linuxcalls.NetlinkAPI

	// descriptors
	arpDescriptor   *descriptor.ARPDescriptor
	routeDescriptor *descriptor.RouteDescriptor
}

// Deps lists dependencies of the interface plugin.
type Deps struct {
	infra.PluginDeps
	Scheduler scheduler.KVScheduler
	NsPlugin  nsplugin.API
	IfPlugin  ifplugin.API
}

// Config holds the nsplugin configuration.
type Config struct {
	Stopwatch bool `json:"stopwatch"`
	Disabled  bool `json:"disabled"`
}

// Init initializes and registers descriptors for Linux ARPs and Routes.
func (plugin *L3Plugin) Init() error {
	// parse configuration file
	config, err := plugin.retrieveConfig()
	if err != nil {
		return err
	}
	if config != nil {
		if config.Disabled {
			plugin.disabled = true
			plugin.Log.Infof("Disabling Linux L3 plugin")
			return nil
		}
		if config.Stopwatch {
			plugin.Log.Infof("stopwatch enabled for %v", plugin.PluginName)
			plugin.stopwatch = measure.NewStopwatch("Linux-L3Plugin", plugin.Log)
		} else {
			plugin.Log.Infof("stopwatch disabled for %v", plugin.PluginName)
		}
	} else {
		plugin.Log.Infof("stopwatch disabled for %v", plugin.PluginName)
	}

	// init handlers
	plugin.l3Handler = linuxcalls.NewNetLinkHandler(plugin.stopwatch)

	// init & register descriptors
	arpDescriptor := adapter.NewARPDescriptor(descriptor.NewARPDescriptor(
		plugin.Scheduler, plugin.IfPlugin, plugin.NsPlugin, plugin.l3Handler, plugin.Log))

	routeDescriptor := adapter.NewRouteDescriptor(descriptor.NewRouteDescriptor(
		plugin.Scheduler, plugin.IfPlugin, plugin.NsPlugin, plugin.l3Handler, plugin.Log))

	plugin.Deps.Scheduler.RegisterKVDescriptor(arpDescriptor)
	plugin.Deps.Scheduler.RegisterKVDescriptor(routeDescriptor)

	return nil
}

// Close does nothing here.
func (plugin *L3Plugin) Close() error {
	return nil
}

// retrieveConfig loads L3Plugin configuration file.
func (plugin *L3Plugin) retrieveConfig() (*Config, error) {
	config := &Config{}
	found, err := plugin.Cfg.LoadValue(config)
	if !found {
		plugin.Log.Debug("Linux L3Plugin config not found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	plugin.Log.Debug("Linux L3Plugin config found")
	return config, err
}
