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

package l3plugin

import (
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
)

// Resync confgures the empty VPP (overwrites the static route)
func (plugin *RouteConfigurator) Resync(staticRoutes []*l3.StaticRoutes_Route) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC routes begin. ")
	// TODO lookup vpp Route Configs

	var wasError error
	if len(staticRoutes) > 0 {
		for _, route := range staticRoutes {
			// VRF ID is already validated at this point
			wasError = plugin.ConfigureRoute(route, string(route.VrfId))
		}
	}
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC routes end. ", wasError)
	return wasError
}
