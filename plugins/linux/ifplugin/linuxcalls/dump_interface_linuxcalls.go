//  Copyright (c) 2018 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// +build !windows,!darwin

package linuxcalls

import (
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	namespaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
)

const (
	// defaultLoopbackName is the name used to access loopback interface in linux
	// host_if_name field in config is effectively ignored
	DefaultLoopbackName = "lo"

	// minimum number of namespaces to be given to a single Go routine for processing
	// in the Retrieve operation
	minWorkForGoRoutine = 3
)

// retrievedIfaces is used as the return value sent via channel by retrieveInterfaces().
type retrievedInterfaces struct {
	interfaces []*InterfaceDetails
	stats      []*InterfaceStatistics
	err        error
}

// DumpInterfaces retrieves all linux interfaces from default namespace and from all
// the other namespaces based on known linux interfaces from the index map.
func (h *NetLinkHandler) DumpInterfaces() ([]*InterfaceDetails, error) {
	return h.DumpInterfacesFromNamespaces(h.getKnownNamespaces())
}

// DumpInterfaceStats retrieves statistics for all linux interfaces from default namespace
// and from all the other namespaces based on known linux interfaces from the index map.
func (h *NetLinkHandler) DumpInterfaceStats() ([]*InterfaceStatistics, error) {
	return h.DumpInterfaceStatsFromNamespaces(h.getKnownNamespaces())
}

// DumpInterfacesFromNamespaces requires context in form of the namespace list of which linux interfaces
// will be retrieved. If no context is provided, interfaces only from the default namespace are retrieved.
func (h *NetLinkHandler) DumpInterfacesFromNamespaces(nsList []*namespaces.NetNamespace) ([]*InterfaceDetails, error) {
	// Always retrieve from the default namespace
	if len(nsList) == 0 {
		nsList = []*namespaces.NetNamespace{nil}
	}
	// Determine the number of go routines to invoke
	goRoutinesCnt := len(nsList) / minWorkForGoRoutine
	if goRoutinesCnt == 0 {
		goRoutinesCnt = 1
	}
	if goRoutinesCnt > h.goRoutineCount {
		goRoutinesCnt = h.goRoutineCount
	}
	ch := make(chan retrievedInterfaces, goRoutinesCnt)

	// Invoke multiple go routines for more efficient parallel interface retrieval
	for idx := 0; idx < goRoutinesCnt; idx++ {
		if goRoutinesCnt > 1 {
			go h.retrieveInterfaces(nsList, idx, goRoutinesCnt, ch)
		} else {
			h.retrieveInterfaces(nsList, idx, goRoutinesCnt, ch)
		}
	}

	// receive results from the go routines
	var linuxIfs []*InterfaceDetails
	for idx := 0; idx < goRoutinesCnt; idx++ {
		retrieved := <-ch
		if retrieved.err != nil {
			return nil, retrieved.err
		}
		linuxIfs = append(linuxIfs, retrieved.interfaces...)
	}
	return linuxIfs, nil
}

// DumpInterfaceStatsFromNamespaces requires context in form of the namespace list of which linux interface stats
// will be retrieved. If no context is provided, interface stats only from the default namespace interfaces
// are retrieved.
func (h *NetLinkHandler) DumpInterfaceStatsFromNamespaces(nsList []*namespaces.NetNamespace) ([]*InterfaceStatistics, error) {
	// Always retrieve from the default namespace
	if len(nsList) == 0 {
		nsList = []*namespaces.NetNamespace{nil}
	}
	// Determine the number of go routines to invoke
	goRoutinesCnt := len(nsList) / minWorkForGoRoutine
	if goRoutinesCnt == 0 {
		goRoutinesCnt = 1
	}
	if goRoutinesCnt > h.goRoutineCount {
		goRoutinesCnt = h.goRoutineCount
	}
	ch := make(chan retrievedInterfaces, goRoutinesCnt)

	// Invoke multiple go routines for more efficient parallel interface retrieval
	for idx := 0; idx < goRoutinesCnt; idx++ {
		if goRoutinesCnt > 1 {
			go h.retrieveInterfaces(nsList, idx, goRoutinesCnt, ch)
		} else {
			h.retrieveInterfaces(nsList, idx, goRoutinesCnt, ch)
		}
	}

	// receive results from the go routines
	var linuxStats []*InterfaceStatistics
	for idx := 0; idx < goRoutinesCnt; idx++ {
		retrieved := <-ch
		if retrieved.err != nil {
			return nil, retrieved.err
		}
		linuxStats = append(linuxStats, retrieved.stats...)
	}
	return linuxStats, nil
}

// Obtain all linux namespaces known to the Linux plugin
func (h *NetLinkHandler) getKnownNamespaces() []*namespaces.NetNamespace {
	// Add default namespace
	nsList := []*namespaces.NetNamespace{nil}
	for _, ifName := range h.ifIndexes.ListAllInterfaces() {
		if metadata, exists := h.ifIndexes.LookupByName(ifName); exists {
			if metadata == nil {
				h.log.Warnf("metadata for %s are nil", ifName)
				continue
			}
			nsListed := false
			for _, ns := range nsList {
				if proto.Equal(ns, metadata.Namespace) {
					nsListed = true
					break
				}
			}
			if !nsListed {
				nsList = append(nsList, metadata.Namespace)
			}
		}
	}
	return nsList
}

// GetVethAlias returns alias for Linux VETH interface managed by the agent.
// The alias stores the VETH logical name together with the peer (logical) name.
func GetVethAlias(vethName, peerName string) string {
	return vethName + "/" + peerName
}

// ParseVethAlias parses out VETH logical name together with the peer name from the alias.
func ParseVethAlias(alias string) (vethName, peerName string) {
	aliasParts := strings.Split(alias, "/")
	vethName = aliasParts[0]
	if len(aliasParts) > 1 {
		peerName = aliasParts[1]
	}
	return
}

// GetTapAlias returns alias for Linux TAP interface managed by the agent.
// The alias stores the TAP_TO_VPP logical name together with VPP-TAP logical name
// and the host interface name as originally set by VPP side.
func GetTapAlias(linuxIf *interfaces.Interface, origHostIfName string) string {
	return linuxIf.Name + "/" + linuxIf.GetTap().GetVppTapIfName() + "/" + origHostIfName
}

// ParseTapAlias parses out TAP_TO_VPP logical name together with the name of the
// linked VPP-TAP and the original TAP host interface name.
func ParseTapAlias(alias string) (linuxTapName, vppTapName, origHostIfName string) {
	aliasParts := strings.Split(alias, "/")
	linuxTapName = aliasParts[0]
	if len(aliasParts) > 1 {
		vppTapName = aliasParts[1]
	}
	if len(aliasParts) > 2 {
		origHostIfName = aliasParts[2]
	}
	return
}

// retrieveInterfaces is run by a separate go routine to retrieve all interfaces
// present in every <goRoutineIdx>-th network namespace from the list.
func (h *NetLinkHandler) retrieveInterfaces(nsList []*namespaces.NetNamespace, goRoutineIdx, goRoutinesCnt int, ch chan<- retrievedInterfaces) {
	var retrieved retrievedInterfaces

	nsCtx := linuxcalls.NewNamespaceMgmtCtx()
	for i := goRoutineIdx; i < len(nsList); i += goRoutinesCnt {
		nsRef := nsList[i]
		// switch to the namespace
		revert, err := h.nsPlugin.SwitchToNamespace(nsCtx, nsRef)
		if err != nil {
			h.log.WithField("namespace", nsRef).Warn("Failed to switch namespace:", err)
			continue // continue with the next namespace
		}

		// get all links in the namespace
		links, err := h.GetLinkList()
		if err != nil {
			h.log.Error("Failed to get link list:", err)
			// switch back to the default namespace before returning error
			revert()
			retrieved.err = err
			break
		}

		// retrieve every interface managed by this agent
		for _, link := range links {
			iface := &interfaces.Interface{
				Namespace:   nsRef,
				HostIfName:  link.Attrs().Name,
				PhysAddress: link.Attrs().HardwareAddr.String(),
				Mtu:         uint32(link.Attrs().MTU),
			}

			alias := link.Attrs().Alias
			if !strings.HasPrefix(alias, h.agentPrefix) {
				// skip interface not configured by this agent
				continue
			}
			alias = strings.TrimPrefix(alias, h.agentPrefix)

			// parse alias to obtain logical references
			if link.Type() == "veth" {
				iface.Type = interfaces.Interface_VETH
				var vethPeerIfName string
				iface.Name, vethPeerIfName = ParseVethAlias(alias)
				iface.Link = &interfaces.Interface_Veth{
					Veth: &interfaces.VethLink{
						PeerIfName: vethPeerIfName,
					},
				}
			} else if link.Type() == "tuntap" || link.Type() == "tun" /* not defined in vishvananda */ {
				iface.Type = interfaces.Interface_TAP_TO_VPP
				var vppTapIfName string
				iface.Name, vppTapIfName, _ = ParseTapAlias(alias)
				iface.Link = &interfaces.Interface_Tap{
					Tap: &interfaces.TapLink{
						VppTapIfName: vppTapIfName,
					},
				}
			} else if link.Attrs().Name == DefaultLoopbackName {
				iface.Type = interfaces.Interface_LOOPBACK
				iface.Name = alias
			} else {
				// unsupported interface type supposedly configured by agent => print warning
				h.log.WithFields(logging.Fields{
					"if-host-name": link.Attrs().Name,
					"namespace":    nsRef,
				}).Warnf("Managed interface of unsupported type: %s", link.Type())
				continue
			}

			// skip interfaces with invalid aliases
			if iface.Name == "" {
				continue
			}

			// retrieve addresses, MTU, etc.
			h.retrieveLinkDetails(link, iface, nsRef)

			// build interface details
			retrieved.interfaces = append(retrieved.interfaces, &InterfaceDetails{
				Interface: iface,
				Meta: &InterfaceMeta{
					LinuxIfIndex:  link.Attrs().Index,
					ParentIndex:   link.Attrs().ParentIndex,
					MasterIndex:   link.Attrs().MasterIndex,
					OperState:     uint8(link.Attrs().OperState),
					Flags:         link.Attrs().RawFlags,
					Encapsulation: link.Attrs().EncapType,
					NumRxQueues:   link.Attrs().NumRxQueues,
					NumTxQueues:   link.Attrs().NumTxQueues,
					TxQueueLen:    link.Attrs().TxQLen,
				},
			})

			// build interface statistics
			retrieved.stats = append(retrieved.stats, &InterfaceStatistics{
				Name:         iface.Name,
				Type:         iface.Type,
				LinuxIfIndex: link.Attrs().Index,
				RxPackets:    link.Attrs().Statistics.RxPackets,
				TxPackets:    link.Attrs().Statistics.TxPackets,
				RxBytes:      link.Attrs().Statistics.RxBytes,
				TxBytes:      link.Attrs().Statistics.TxBytes,
				RxErrors:     link.Attrs().Statistics.RxErrors,
				TxErrors:     link.Attrs().Statistics.TxErrors,
				RxDropped:    link.Attrs().Statistics.TxDropped,
				TxDropped:    link.Attrs().Statistics.RxDropped,
			})
		}

		// switch back to the default namespace
		revert()
	}

	ch <- retrieved
}

// retrieveLinkDetails retrieves link details common to all interface types (e.g. addresses).
func (h *NetLinkHandler) retrieveLinkDetails(link netlink.Link, iface *interfaces.Interface, nsRef *namespaces.NetNamespace) {
	var err error
	// read interface status
	iface.Enabled, err = h.IsInterfaceUp(link.Attrs().Name)
	if err != nil {
		h.log.WithFields(logging.Fields{
			"if-host-name": link.Attrs().Name,
			"namespace":    nsRef,
		}).Warn("Failed to read interface status:", err)
	}

	// read assigned IP addresses
	addressList, err := h.GetAddressList(link.Attrs().Name)
	if err != nil {
		h.log.WithFields(logging.Fields{
			"if-host-name": link.Attrs().Name,
			"namespace":    nsRef,
		}).Warn("Failed to read address list:", err)
	}
	for _, address := range addressList {
		if address.Scope == unix.RT_SCOPE_LINK {
			// ignore link-local IPv6 addresses
			continue
		}
		mask, _ := address.Mask.Size()
		addrStr := address.IP.String() + "/" + strconv.Itoa(mask)
		iface.IpAddresses = append(iface.IpAddresses, addrStr)
	}

	// read checksum offloading
	if iface.Type == interfaces.Interface_VETH {
		rxOn, txOn, err := h.GetChecksumOffloading(link.Attrs().Name)
		if err != nil {
			h.log.WithFields(logging.Fields{
				"if-host-name": link.Attrs().Name,
				"namespace":    nsRef,
			}).Warn("Failed to read checksum offloading:", err)
		} else {
			if !rxOn {
				iface.GetVeth().RxChecksumOffloading = interfaces.VethLink_CHKSM_OFFLOAD_DISABLED
			}
			if !txOn {
				iface.GetVeth().TxChecksumOffloading = interfaces.VethLink_CHKSM_OFFLOAD_DISABLED
			}
		}
	}
}
