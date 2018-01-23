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

package ifplugin

import (
	"time"

	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/model/interfaces"
)

// Resync configures an initial set of interfaces. Existing Linux interfaces are registered and potentially re-configured.
func (plugin *LinuxInterfaceConfigurator) Resync(interfaces []*interfaces.LinuxInterfaces_Interface) (errs []error) {
	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface begin.")

	start := time.Now()
	defer func() {
		if plugin.Stopwatch != nil {
			timeLog := measure.GetTimeLog("linux-interface resync", plugin.Stopwatch)
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Step 1: Create missing Linux interfaces and recreate existing ones
	for _, iface := range interfaces {
		err := plugin.ConfigureLinuxInterface(iface)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Step 2: Dump pre-existing and currently not managed interfaces in the current namespace.
	err := plugin.LookupLinuxInterfaces()
	if err != nil {
		return []error{err}
	}

	plugin.Log.WithField("cfg", plugin).Debug("RESYNC Interface end. ", errs)

	return
}
