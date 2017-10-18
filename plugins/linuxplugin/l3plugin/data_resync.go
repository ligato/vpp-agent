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
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
	"time"
)

// Resync configures an initial set of ARPs. Existing Linux ARPs are registered and potentially re-configured.
func (plugin *LinuxArpConfigurator) Resync(interfaces []*l3.LinuxStaticArpEntries_ArpEntry) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC ARPs begin.")
	start := time.Now()
	defer func() {
		if plugin.Stopwatch != nil {
			timeLog := measure.GetTimeLog("linux-arp resync", plugin.Stopwatch)
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// todo implement

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC ARPs end. ")

	return nil
}

// Resync configures an initial set of static routes. Existing Linux static routes are registered and potentially re-configured.
func (plugin *LinuxRouteConfigurator) Resync(interfaces []*l3.LinuxStaticRoutes_Route) error {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC static routes begin.")
	start := time.Now()
	defer func() {
		if plugin.Stopwatch != nil {
			timeLog := measure.GetTimeLog("linux-route resync", plugin.Stopwatch)
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// todo implement

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC static routes end. ")

	return nil
}
