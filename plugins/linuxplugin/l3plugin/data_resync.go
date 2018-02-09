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
	"time"

	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/l3"
)

// Resync configures an initial set of ARPs. Existing Linux ARPs are registered and potentially re-configured.
func (plugin *LinuxArpConfigurator) Resync(arpEntries []*l3.LinuxStaticArpEntries_ArpEntry) (errs []error) {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC ARPs begin.")

	start := time.Now()
	defer func() {
		if plugin.Stopwatch != nil {
			timeLog := measure.GetTimeLog("linux-arp resync", plugin.Stopwatch)
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Create missing arp entries and update existing ones
	for _, entry := range arpEntries {
		err := plugin.ConfigureLinuxStaticArpEntry(entry)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Dump pre-existing not managed arp entries
	err := plugin.LookupLinuxArpEntries()
	if err != nil {
		errs = append(errs, err)
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC ARPs end. ")

	return
}

// Resync configures an initial set of static routes. Existing Linux static routes are registered and potentially re-configured.
func (plugin *LinuxRouteConfigurator) Resync(routes []*l3.LinuxStaticRoutes_Route) (errs []error) {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC static routes begin.")

	start := time.Now()
	defer func() {
		if plugin.Stopwatch != nil {
			timeLog := measure.GetTimeLog("linux-route resync", plugin.Stopwatch)
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Create missing routes and update existing ones
	var defaultEntries []*l3.LinuxStaticRoutes_Route
	for _, entry := range routes {
		// if default entry is found, store it and configure as the last one
		if entry.Default {
			defaultEntries = append(defaultEntries, entry)
			continue
		}
		err := plugin.ConfigureLinuxStaticRoute(entry)
		if err != nil {
			errs = append(errs, err)
		}
	}
	// configure default entries if there are some
	if len(defaultEntries) != 0 {
		for _, defEntry := range defaultEntries {
			err := plugin.ConfigureLinuxStaticRoute(defEntry)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	// Dump pre-existing not managed arp entries
	err := plugin.LookupLinuxRoutes()
	if err != nil {
		errs = append(errs, err)
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC static routes end. ")

	return
}
