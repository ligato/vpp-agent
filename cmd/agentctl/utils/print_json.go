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

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
	"github.com/logrusorgru/aurora.git"
	"sort"
	"strings"
)

const (
	indent    = "  "
	emptyJSON = "{}"
)

// PrintDataAsJSON prints ETCD data in JSON format
func (ed EtcdDump) PrintDataAsJSON(filter []string) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	keys := ed.getSortedKeys()
	var wasError error

	vpps, isData := processFilter(keys, filter)
	if !isData {
		fmt.Fprintf(buffer, "No data to display for VPPS: %s\n", vpps)
		return buffer, wasError
	}

	for _, key := range keys {
		if isNotInFilter(key, vpps) {
			continue
		}

		vd, _ := ed[key]
		// Obtain raw data
		ifaceConfDataRoot, ifaceConfKeys := getInterfaceConfigData(vd.Interfaces)
		ifaceStateDataRoot, ifaceStateKeys := getInterfaceStateData(vd.Interfaces)
		l2DataRoot, l2Keys := getL2Data(vd.BridgeDomains)
		fibDataRoot, fibKeys := getFIBData(vd.FibTableEntries)

		// Interface config data
		jsConfData, err := json.MarshalIndent(ifaceConfDataRoot, "", indent)
		if err != nil {
			wasError = err
		}
		// Interface state data
		jsStateData, err := json.MarshalIndent(ifaceStateDataRoot, "", indent)
		if err != nil {
			wasError = err
		}
		// L2 data
		jsL2Data, err := json.MarshalIndent(l2DataRoot, "", indent)
		if err != nil {
			wasError = err
		}

		// FIB data
		jsFIBData, err := json.MarshalIndent(fibDataRoot, "", indent)
		if err != nil {
			wasError = err
		}

		// Add data to buffer
		if string(jsConfData) != emptyJSON {
			printLabel(buffer, key+": - INTERFACE CONFIG\n", indent, ifaceConfKeys)
			fmt.Fprintf(buffer, "%s\n", jsConfData)
		}
		if string(jsStateData) != emptyJSON {
			printLabel(buffer, key+": - INTERFACE STATE\n", indent, ifaceStateKeys)
			fmt.Fprintf(buffer, "%s\n", jsStateData)
		}
		if string(jsL2Data) != emptyJSON {
			printLabel(buffer, key+": - BRIDGE DOMAINS\n", indent, l2Keys)
			fmt.Fprintf(buffer, "%s\n", jsL2Data)
		}
		if string(jsFIBData) != emptyJSON {
			printLabel(buffer, key+": - FIB TABLE\n", indent, fibKeys)
			fmt.Fprintf(buffer, "%s\n", jsFIBData)
		}

	}

	return buffer, wasError
}

// Function returns a list of VPPs which will be shown according to provided filter. If the filter is empty, all VPPs
// will be shown. If there is nothing to show because of filter, isData flag is returned as false
func processFilter(keys []string, filter []string) ([]string, bool) {
	var vpps []string
	if len(filter) > 0 {
		// Ignore all parameters but first
		vpps = strings.Split(filter[0], ",")
	} else {
		// Show all if there is no filter
		vpps = keys
	}
	var isData bool
	// Find at leas one match
	for _, key := range keys {
		for _, vpp := range vpps {
			if key == vpp {
				isData = true
				break
			}
		}
	}
	return vpps, isData
}

// Returns true if provided key is present in filter, false otherwise
func isNotInFilter(key string, filter []string) bool {
	for _, itemInFilter := range filter {
		if itemInFilter == key {
			return false
		}
	}
	return true
}

// Get interface config data and create full interface config proto structure
func getInterfaceConfigData(interfaceData map[string]InterfaceWithMD) (*interfaces.Interfaces, []string) {
	// Config data
	ifaceRoot := interfaces.Interfaces{}
	ifaces := []*interfaces.Interfaces_Interface{}
	var keyset []string
	for _, ifaceData := range interfaceData {
		if ifaceData.Config != nil {
			iface := ifaceData.Config.Interfaces_Interface
			ifaces = append(ifaces, iface)
			keyset = append(keyset, ifaceData.Config.Key)
		}
	}
	sort.Strings(keyset)
	ifaceRoot.Interface = ifaces

	return &ifaceRoot, keyset
}

// Get interface state data and create full interface state proto structure
func getInterfaceStateData(interfaceData map[string]InterfaceWithMD) (*interfaces.InterfacesState, []string) {
	// Status data
	ifaceStateRoot := interfaces.InterfacesState{}
	ifaceStates := []*interfaces.InterfacesState_Interface{}
	var keyset []string
	for _, ifaceData := range interfaceData {
		if ifaceData.State != nil {
			ifaceState := ifaceData.State.InterfacesState_Interface
			ifaceStates = append(ifaceStates, ifaceState)
			keyset = append(keyset, ifaceData.State.Key)
		}
	}
	sort.Strings(keyset)
	ifaceStateRoot.Interface = ifaceStates

	return &ifaceStateRoot, keyset
}

// Get l2 data and create full l2 bridge domains proto structure
func getL2Data(l2ConfigData map[string]BdWithMD) (*l2.BridgeDomains, []string) {
	l2Root := l2.BridgeDomains{}
	l2Data := []*l2.BridgeDomains_BridgeDomain{}
	var keyset []string
	for _, bdData := range l2ConfigData {
		bd := bdData.BridgeDomains_BridgeDomain
		l2Data = append(l2Data, bd)
		keyset = append(keyset, bdData.Key)
	}
	sort.Strings(keyset)
	l2Root.BridgeDomains = l2Data

	return &l2Root, keyset
}

// Get FIB data and create full FIB proto structure
func getFIBData(fibData FibTableWithMD) (*l2.FibTableEntries, []string) {
	fibRoot := l2.FibTableEntries{}
	fibRoot.FibTableEntry = fibData.FibTable
	var keyset []string
	for _, fib := range fibData.FibTable {
		keyset = append(keyset, l2.FIBPrefix+fib.PhysAddress)
	}
	sort.Strings(keyset)

	return &fibRoot, keyset
}

// Print label before every data structure including used keys
func printLabel(buffer *bytes.Buffer, label string, prefix string, keyset []string) {
	// Format output - find longest string in label to make label nicer
	labelLength := len(label)
	for _, key := range keyset {
		if len(key) > labelLength {
			labelLength = len(key)
		}
	}
	ub := prefix + strings.Repeat("-", labelLength) + "\n"

	// Print label
	fmt.Fprintf(buffer, ub)
	fmt.Fprintf(buffer, "%s%s\n", prefix, aurora.Bold(label))
	fmt.Fprintf(buffer, "%s%s\n", prefix, "Keys:")
	for _, key := range keyset {
		if key != "" {
			fmt.Fprintf(buffer, "%s%s\n", prefix, key)
		}
	}
	fmt.Fprintf(buffer, ub)
}
