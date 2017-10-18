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

// LinuxArpConfigurator watches for any changes in the configuration of static ARPs as modelled by the proto file
// "model/l3/l3.proto" and stored in ETCD under the key "/vnf-agent/{vnf-agent}/linux/config/v1/arp".
// Updates received from the northbound API are compared with the Linux network configuration and differences
// are applied through the Netlink AP
type LinuxArpConfigurator struct {
	Log logging.Logger

	arpIndexes l3idx.LinuxARPIndexRW

	// Time measurement
	Stopwatch *measure.Stopwatch // timer used to measure and store time

}

// Init initializes ARP configurator and starts goroutines
func (plugin *LinuxArpConfigurator) Init(arpIndexes l3idx.LinuxARPIndexRW) error {
	plugin.Log.Debug("Initializing LinuxArpConfigurator")
	plugin.arpIndexes = arpIndexes

	return nil
}

// Close closes all goroutines started during Init
func (plugin *LinuxArpConfigurator) Close() error {
	return nil
}

// ConfigureLinuxStaticArpEntry reacts to a new northbound Linux ARP entry config by creating and configuring
// the entry in the host network stack through Netlink API.
func (plugin *LinuxArpConfigurator) ConfigureLinuxStaticArpEntry(arpEntry *l3.LinuxStaticArpEntries_ArpEntry) error {
	// todo implement
	return nil
}

// ModifyLinuxStaticArpEntry applies changes in the NB configuration of a Linux ARP into the host network stack
// through Netlink API.
func (plugin *LinuxArpConfigurator) ModifyLinuxStaticArpEntry(newArpEntry *l3.LinuxStaticArpEntries_ArpEntry, oldArpEntry *l3.LinuxStaticArpEntries_ArpEntry) error {
	// todo implement
	return nil
}

// DeleteLinuxStaticArpEntry reacts to a removed NB configuration of a Linux ARP entry.
func (plugin *LinuxArpConfigurator) DeleteLinuxStaticArpEntry(arpEntry *l3.LinuxStaticArpEntries_ArpEntry) error {
	// todo implement
	return nil
}
