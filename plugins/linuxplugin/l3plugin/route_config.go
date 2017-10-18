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

//go:generate protoc --proto_path=model --gogo_out=model model/l3/l3.proto

package l3plugin

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
)

// LinuxRouteConfigurator watches for any changes in the configuration of static routes as modelled by the proto file
// "model/l3/l3.proto" and stored in ETCD under the key "/vnf-agent/{vnf-agent}/linux/config/v1/route".
// Updates received from the northbound API are compared with the Linux network configuration and differences
// are applied through the Netlink AP
type LinuxRouteConfigurator struct {
	Log logging.Logger

	rtIndexes l3idx.LinuxRouteIndexRW

	// Time measurement
	Stopwatch *measure.Stopwatch // timer used to measure and store time

}

// Init initializes static route configurator and starts goroutines
func (plugin *LinuxRouteConfigurator) Init(rtIndexes l3idx.LinuxRouteIndexRW) error {
	plugin.Log.Debug("Initializing LinuxRouteConfigurator")
	plugin.rtIndexes = rtIndexes

	return nil
}

// Close closes all goroutines started during Init
func (plugin *LinuxRouteConfigurator) Close() error {
	return nil
}

// ConfigureLinuxStaticRoute reacts to a new northbound Linux static route config by creating and configuring
// the route in the host network stack through Netlink API.
func (plugin *LinuxRouteConfigurator) ConfigureLinuxStaticRoute(route *l3.LinuxStaticRoutes_Route) error {
	// todo implement
	return nil
}

// ModifyLinuxStaticRoute applies changes in the NB configuration of a Linux static route into the host network stack
// through Netlink API.
func (plugin *LinuxRouteConfigurator) ModifyLinuxStaticRoute(newRoute *l3.LinuxStaticRoutes_Route, oldRoute *l3.LinuxStaticRoutes_Route) error {
	// todo implement
	return nil
}

// DeleteLinuxStaticRoute reacts to a removed NB configuration of a Linux static route entry.
func (plugin *LinuxRouteConfigurator) DeleteLinuxStaticRoute(route *l3.LinuxStaticRoutes_Route) error {
	// todo implement
	return nil
}
