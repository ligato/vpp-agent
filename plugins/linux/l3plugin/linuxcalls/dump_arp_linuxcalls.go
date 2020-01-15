// Copyright (c) 2019 Cisco and/or its affiliates.
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

// +build !windows,!darwin

package linuxcalls

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
)

// retrievedARPs is used as the return value sent via channel by retrieveARPs().
type retrievedARPs struct {
	arps []*ArpDetails
	err  error
}

// GetARPEntries reads all configured static ARP entries for given interface.
// <interfaceIdx> works as filter, if set to zero, all arp entries in the namespace
// are returned
func (h *NetLinkHandler) GetARPEntries(interfaceIdx int) ([]netlink.Neigh, error) {
	return netlink.NeighList(interfaceIdx, 0)
}

// DumpARPEntries reads all ARP entries and returns them as details
// with proto-modeled ARP data and additional metadata
func (h *NetLinkHandler) DumpARPEntries() ([]*ArpDetails, error) {
	interfaces := h.ifIndexes.ListAllInterfaces()
	goRoutinesCnt := len(interfaces) / minWorkForGoRoutine
	if goRoutinesCnt == 0 {
		goRoutinesCnt = 1
	}
	if goRoutinesCnt > h.goRoutineCount {
		goRoutinesCnt = h.goRoutineCount
	}
	ch := make(chan retrievedARPs, goRoutinesCnt)

	// invoke multiple go routines for more efficient parallel ARP retrieval
	for idx := 0; idx < goRoutinesCnt; idx++ {
		if goRoutinesCnt > 1 {
			go h.retrieveARPs(interfaces, idx, goRoutinesCnt, ch)
		} else {
			h.retrieveARPs(interfaces, idx, goRoutinesCnt, ch)
		}
	}

	// collect results from the go routines
	var arpDetails []*ArpDetails
	for idx := 0; idx < goRoutinesCnt; idx++ {
		retrieved := <-ch
		if retrieved.err != nil {
			return nil, retrieved.err
		}
		arpDetails = append(arpDetails, retrieved.arps...)
	}

	return arpDetails, nil
}

// retrieveARPs is run by a separate go routine to retrieve all ARP entries associated
// with every <goRoutineIdx>-th interface.
func (h *NetLinkHandler) retrieveARPs(interfaces []string, goRoutineIdx, goRoutinesCnt int, ch chan<- retrievedARPs) {
	var retrieved retrievedARPs
	nsCtx := linuxcalls.NewNamespaceMgmtCtx()

	for i := goRoutineIdx; i < len(interfaces); i += goRoutinesCnt {
		ifName := interfaces[i]
		// get interface metadata
		ifMeta, found := h.ifIndexes.LookupByName(ifName)
		if !found || ifMeta == nil {
			retrieved.err = errors.Errorf("failed to obtain metadata for interface %s", ifName)
			h.log.Error(retrieved.err)
			break
		}

		// switch to the namespace of the interface
		revertNs, err := h.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
		if err != nil {
			// namespace and all the ARPs it had contained no longer exist
			h.log.WithFields(logging.Fields{
				"err":       err,
				"namespace": ifMeta.Namespace,
			}).Warn("Failed to retrieve ARPs from the namespace")
			continue
		}

		// get ARPs assigned to this interface
		arps, err := h.GetARPEntries(ifMeta.LinuxIfIndex)
		revertNs()
		if err != nil {
			retrieved.err = err
			h.log.Error(retrieved.err)
			break
		}

		// convert each ARP from Netlink representation to the ARP details
		for _, arp := range arps {
			if arp.IP.IsLinkLocalMulticast() {
				// skip link-local multi-cast ARPs until there is a requirement to support them as well
				continue
			}
			retrieved.arps = append(retrieved.arps, &ArpDetails{
				ARP: &linux_l3.ARPEntry{
					Interface: ifName,
					IpAddress: arp.IP.String(),
					HwAddress: arp.HardwareAddr.String(),
				},
				Meta: &ArpMeta{
					InterfaceIndex: uint32(arp.LinkIndex),
					IPFamily:       uint32(arp.Family),
					VNI:            uint32(arp.VNI),
				},
			})
		}
	}

	ch <- retrieved
}
