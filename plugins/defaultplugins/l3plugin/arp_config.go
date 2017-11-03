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

var msgCompatibilityARP = []govppapi.Message{
	&ip.IPNeighborAddDel{},
	&ip.IPNeighborAddDelReply{},
}

// ArpConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 arp entries as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/arp". Updates received from the northbound API
// are compared with the VPP run-time configuration and differences are applied through the VPP binary API.
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

	if err := plugin.vppChan.CheckMessageCompatibility(msgCompatibilityARP...); err != nil {
		plugin.Log.Error(err)
		return err
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

// AddArp processes the NB config and propagates it to bin api call
func (plugin *ArpConfigurator) AddArp(entry *l3.ArpTable_ArpTableEntry) error {
	plugin.Log.Infof("Creating ARP entry %v", *entry)

	// Transform arp data
	arp, err := TransformArp(entry, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	if arp == nil {
		return nil
	}
	plugin.Log.Debugf("adding ARP: %+v", *arp)

	// Create and register new arp entry
	err = vppcalls.VppAddArp(arp, plugin.vppChan,
		measure.GetTimeLog(ip.IPNeighborAddDel{}, plugin.Stopwatch))
	if err != nil {
		return err
	}
	arpID := arpIdentifier(arp.Interface, arp.IPAddress.String(), arp.MacAddress.String())
	plugin.ArpIndexes.RegisterName(arpID, plugin.ArpIndexSeq, nil)
	plugin.ArpIndexSeq++
	plugin.Log.Infof("ARP entry %v registered", arpID)

	return nil
}

// ChangeArp processes the NB config and propagates it to bin api call
func (plugin *ArpConfigurator) ChangeArp(entry *l3.ArpTable_ArpTableEntry, prevEntry *l3.ArpTable_ArpTableEntry) error {
	plugin.Log.Infof("Change ARP entry %v to %v", *prevEntry, *entry)

	if err := plugin.DeleteArp(prevEntry); err != nil {
		return err
	}
	if err := plugin.AddArp(entry); err != nil {
		return err
	}

	return nil
}

// DeleteArp processes the NB config and propagates it to bin api call
func (plugin *ArpConfigurator) DeleteArp(entry *l3.ArpTable_ArpTableEntry) error {
	plugin.Log.Infof("Deleting ARP entry %v", *entry)

	// Transform arp data
	arp, err := TransformArp(entry, plugin.SwIfIndexes, plugin.Log)
	if err != nil {
		return err
	}
	if arp == nil {
		return nil
	}
	plugin.Log.Debugf("deleting ARP: %+v", arp)

	// Delete and unregister new arp
	err = vppcalls.VppDelArp(arp, plugin.vppChan,
		measure.GetTimeLog(ip.IPNeighborAddDel{}, plugin.Stopwatch))
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
