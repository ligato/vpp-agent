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
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/ip"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

type ArpConfigurator struct {
	Log logging.Logger

	GoVppmux    govppmux.API
	ArpIndexes  idxvpp.NameToIdxRW
	ArpIndexSeq uint32
	SwIfIndexes ifaceidx.SwIfIndex
	vppChan     *govppapi.Channel

	Stopwatch *measure.Stopwatch
}

// Init initializes ARP configurator
func (plugin *ArpConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing ArpConfigurator")

	// Init VPP API channel
	plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	err = plugin.checkMsgCompatibility()
	if err != nil {
		return err
	}

	return nil
}

// TransformArp converts raw entry data to ARP object
func TransformArp(arpInput *l3.ArpTable_ArpTableEntry, index ifaceidx.SwIfIndex, log logging.Logger) (*vppcalls.ArpEntry, error) {
	if arpInput == nil {
		log.Infof("ARP input is empty")
		return nil, nil
	}
	if arpInput.Interface == "" {
		log.Infof("ARP input does not contain interface")
		return nil, nil
	}
	if arpInput.IpAddress == "" {
		log.Infof("ARP input does not contain IP")
		return nil, nil
	}
	if arpInput.PhysAddress == "" {
		log.Infof("ARP input does not contain MAC")
		return nil, nil
	}

	ifName := arpInput.Interface
	ifIndex, _, exists := index.LookupIdx(ifName)
	if !exists {
		return nil, fmt.Errorf("ARP entry interface %v not found", ifName)
	}

	ipAddr := net.ParseIP(arpInput.IpAddress)
	macAddr, err := net.ParseMAC(arpInput.PhysAddress)
	if err != nil {
		return nil, err
	}
	arp := &vppcalls.ArpEntry{
		Interface:  ifIndex,
		IPAddress:  ipAddr,
		MacAddress: macAddr,
		Static:     arpInput.Static,
	}
	return arp, nil
}

func (plugin *ArpConfigurator) Add(entry *l3.ArpTable_ArpTableEntry) error {
	//plugin.Log.Infof("Creating new ARP entry %v -> %v (%v) for interface %v", entry.IpAddress, entry.PhysAddress, entry.Static, entry.Interface)
	plugin.Log.Infof("Creating ARP entry %+v", *entry)

	// Transform route data
	arp, err := TransformArp(entry, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	if arp == nil {
		return nil
	}
	plugin.Log.Debugf("adding ARP: %+v", *arp)

	// Create and register new route
	err = vppcalls.VppAddArp(arp, plugin.vppChan, measure.GetTimeLog(ip.IPNeighborAddDel{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	arpID := arpIdentifier(arp.Interface, arp.IPAddress.String(), arp.MacAddress.String())
	plugin.ArpIndexes.RegisterName(arpID, plugin.ArpIndexSeq, nil)
	plugin.ArpIndexSeq++
	plugin.Log.Infof("ARP entry %v registered", arpID)

	return nil
}

func (plugin *ArpConfigurator) Diff(entry *l3.ArpTable_ArpTableEntry, prevEntry *l3.ArpTable_ArpTableEntry) error {
	return fmt.Errorf("ARP DIFF NOT IMPLEMENTED")
}

func (plugin *ArpConfigurator) Delete(entry *l3.ArpTable_ArpTableEntry) error {
	plugin.Log.Infof("Deleting ARP entry %+v", *entry)

	// Transform route data
	arp, err := TransformArp(entry, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	if arp == nil {
		return nil
	}
	plugin.Log.Debugf("deleting ARP: %+v", arp)

	// Delete and unregister new route
	err = vppcalls.VppDelArp(arp, plugin.vppChan, measure.GetTimeLog(ip.IPNeighborAddDel{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	arpID := arpIdentifier(arp.Interface, arp.IPAddress.String(), arp.MacAddress.String())
	_, _, found := plugin.ArpIndexes.UnregisterName(arpID)
	if found {
		plugin.Log.Infof("ARP entry %v unregistered", arpID)
	} else {
		plugin.Log.Warnf("Unregister failed, ARP entry %v not found", arpID)
	}

	return nil
}

// Close GOVPP channel
func (plugin *ArpConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// Creates unique identifier which serves as a name in name to index mapping
func arpIdentifier(iface uint32, ip, mac string) string {
	return fmt.Sprintf("arp-iface%v-%v-%v", iface, ip, mac)
}

func (plugin *ArpConfigurator) checkMsgCompatibility() error {
	msgs := []govppapi.Message{
		&ip.IPNeighborAddDel{},
		&ip.IPNeighborAddDelReply{},
	}
	err := plugin.vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		plugin.Log.Error(err)
	}
	return err
}
