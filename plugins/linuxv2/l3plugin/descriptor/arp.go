// Copyright (c) 2018 Cisco and/or its affiliates.
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

package descriptor

import (
	"net"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/ligato/cn-infra/logging"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"

	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	ifdescriptor "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin/descriptor/adapter"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin/linuxcalls"
	ifmodel "github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
	l3 "github.com/ligato/vpp-agent/plugins/linuxv2/model/l3"
	"github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin/linuxcalls"
)

const (
	// ARPDescriptorName is the name of the descriptor for Linux ARP entries.
	ARPDescriptorName = "linux-arp"

	// dependency labels
	arpInterfaceDep = "interface-exists"

	// minimum number of interfaces to be given to a single Go routine for processing
	// in the Dump operation
	minWorkForGoRoutine = 3
)

// A list of non-retriable errors:
var (
	// ErrARPWithoutInterface is returned when Linux ARP configuration is missing
	// interface reference.
	ErrARPWithoutInterface = errors.New("Linux ARP entry defined without interface reference")

	// ErrARPWithoutIP is returned when Linux ARP configuration is missing IP address.
	ErrARPWithoutIP = errors.New("Linux ARP entry defined without IP address")

	// ErrARPWithInvalidIP is returned when Linux ARP configuration contains IP address that cannot be parsed.
	ErrARPWithInvalidIP = errors.New("Linux ARP entry defined with invalid IP address")

	// ErrARPWithoutHwAddr is returned when Linux ARP configuration is missing
	// MAC address.
	ErrARPWithoutHwAddr = errors.New("Linux ARP entry defined without MAC address")

	// ErrARPWithInvalidHwAddr is returned when Linux ARP configuration contains MAC address that cannot be parsed.
	ErrARPWithInvalidHwAddr = errors.New("Linux ARP entry defined with invalid MAC address")
)

// ARPDescriptor teaches KVScheduler how to configure Linux ARP entries.
type ARPDescriptor struct {
	log       logging.Logger
	l3Handler l3linuxcalls.NetlinkAPI
	ifPlugin  ifplugin.API
	nsPlugin  nsplugin.API
	scheduler scheduler.KVScheduler

	// parallelization of the Dump operation
	dumpGoRoutinesCnt int
}

// NewARPDescriptor creates a new instance of the ARP descriptor.
func NewARPDescriptor(
	scheduler scheduler.KVScheduler, ifPlugin ifplugin.API, nsPlugin nsplugin.API,
	l3Handler l3linuxcalls.NetlinkAPI, log logging.PluginLogger, dumpGoRoutinesCnt int) *ARPDescriptor {

	return &ARPDescriptor{
		scheduler:         scheduler,
		l3Handler:         l3Handler,
		ifPlugin:          ifPlugin,
		nsPlugin:          nsPlugin,
		dumpGoRoutinesCnt: dumpGoRoutinesCnt,
		log:               log.NewLogger("arp-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *ARPDescriptor) GetDescriptor() *adapter.ARPDescriptor {
	return &adapter.ARPDescriptor{
		Name:               ARPDescriptorName,
		KeySelector:        d.IsARPKey,
		ValueTypeName:      proto.MessageName(&l3.StaticARPEntry{}),
		ValueComparator:    d.EquivalentARPs,
		NBKeyPrefix:        l3.StaticArpKeyPrefix,
		Add:                d.Add,
		Delete:             d.Delete,
		Modify:             d.Modify,
		IsRetriableFailure: d.IsRetriableFailure,
		Dependencies:       d.Dependencies,
		Dump:               d.Dump,
		DumpDependencies:   []string{ifdescriptor.InterfaceDescriptorName},
	}
}

// IsARPKey returns <true> if the key identifies a Linux ARP configuration.
func (d *ARPDescriptor) IsARPKey(key string) bool {
	return strings.HasPrefix(key, l3.StaticArpKeyPrefix)
}

// EquivalentARPs is case-insensitive comparison function for l3.LinuxStaticARPEntry.
func (d *ARPDescriptor) EquivalentARPs(key string, oldArp, NewArp *l3.StaticARPEntry) bool {
	// interfaces compared as usually:
	if oldArp.Interface != NewArp.Interface {
		return false
	}

	// compare MAC addresses case-insensitively
	if strings.ToLower(oldArp.HwAddress) != strings.ToLower(NewArp.HwAddress) {
		return false
	}

	// compare IP addresses converted to net.IPNet
	return equalAddrs(oldArp.IpAddress, NewArp.IpAddress)
}

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (d *ARPDescriptor) IsRetriableFailure(err error) bool {
	nonRetriable := []error{
		ErrARPWithoutInterface,
		ErrARPWithoutIP,
		ErrARPWithInvalidIP,
		ErrARPWithoutHwAddr,
		ErrARPWithInvalidHwAddr,
	}
	for _, nonRetriableErr := range nonRetriable {
		if err == nonRetriableErr {
			return false
		}
	}
	return true
}

// Add creates ARP entry.
func (d *ARPDescriptor) Add(key string, arp *l3.StaticARPEntry) (metadata interface{}, err error) {
	err = d.updateARPEntry(arp, "add", d.l3Handler.SetARPEntry)
	return nil, err
}

// Delete removes ARP entry.
func (d *ARPDescriptor) Delete(key string, arp *l3.StaticARPEntry, metadata interface{}) error {
	return d.updateARPEntry(arp, "delete", d.l3Handler.DelARPEntry)
}

// Modify is able to change MAC address of the ARP entry.
func (d *ARPDescriptor) Modify(key string, oldARP, newARP *l3.StaticARPEntry, oldMetadata interface{}) (newMetadata interface{}, err error) {
	err = d.updateARPEntry(newARP, "modify", d.l3Handler.SetARPEntry)
	return nil, err
}

// updateARPEntry adds, modifies or deletes an ARP entry.
func (d *ARPDescriptor) updateARPEntry(arp *l3.StaticARPEntry, actionName string, actionClb func(arpEntry *netlink.Neigh) error) error {
	var err error

	// validate the configuration first
	if arp.Interface == "" {
		err = ErrARPWithoutInterface
		d.log.Error(err)
		return err
	}
	if arp.IpAddress == "" {
		err = ErrARPWithoutIP
		d.log.Error(err)
		return err
	}
	if arp.HwAddress == "" {
		err = ErrARPWithoutHwAddr
		d.log.Error(err)
		return err
	}

	// Prepare ARP entry object
	neigh := &netlink.Neigh{}

	// Get interface metadata
	ifMeta, found := d.ifPlugin.GetInterfaceIndex().LookupByName(arp.Interface)
	if !found || ifMeta == nil {
		err = errors.Errorf("failed to obtain metadata for interface %s", arp.Interface)
		d.log.Error(err)
		return err
	}

	// set link index
	neigh.LinkIndex = ifMeta.LinuxIfIndex

	// set IP address
	ipAddr := net.ParseIP(arp.IpAddress)
	if ipAddr == nil {
		err = ErrARPWithInvalidIP
		d.log.Error(err)
		return err
	}
	neigh.IP = ipAddr

	// set MAC address
	mac, err := net.ParseMAC(arp.HwAddress)
	if err != nil {
		err = ErrARPWithInvalidHwAddr
		d.log.Error(err)
		return err
	}
	neigh.HardwareAddr = mac

	// set ARP entry state (always permanent for static ARPs configured by the agent)
	neigh.State = netlink.NUD_PERMANENT

	// set ip family based on the IP address
	if neigh.IP.To4() != nil {
		neigh.Family = netlink.FAMILY_V4
	} else {
		neigh.Family = netlink.FAMILY_V6
	}

	// move to the namespace of the associated interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revertNs, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
	if err != nil {
		err = errors.Errorf("failed to switch namespace: %v", err)
		d.log.Error(err)
		return err
	}
	defer revertNs()

	// update ARP entry in the interface namespace
	err = actionClb(neigh)
	if err != nil {
		err = errors.Errorf("failed to %s linux ARP entry: %v", actionName, err)
		d.log.Error(err)
		return err
	}

	return nil
}

// Dependencies lists dependencies for a Linux ARP entry.
func (d *ARPDescriptor) Dependencies(key string, arp *l3.StaticARPEntry) []scheduler.Dependency {
	// the associated interface must exist and be UP
	if arp.Interface != "" {
		return []scheduler.Dependency{
			{
				Label: arpInterfaceDep,
				Key:   ifmodel.InterfaceStateKey(arp.Interface, true),
			},
		}
	}
	return nil
}

// arpDump is used as the return value sent via channel by dumpARPs().
type arpDump struct {
	arps []adapter.ARPKVWithMetadata
	err  error
}

// Dump returns all ARP entries associated with interfaces managed by this agent.
func (d *ARPDescriptor) Dump(correlate []adapter.ARPKVWithMetadata) ([]adapter.ARPKVWithMetadata, error) {
	var dump []adapter.ARPKVWithMetadata
	interfaces := d.ifPlugin.GetInterfaceIndex().ListAllInterfaces()
	goRoutinesCnt := len(interfaces) / minWorkForGoRoutine
	if goRoutinesCnt == 0 {
		goRoutinesCnt = 1
	}
	if goRoutinesCnt > d.dumpGoRoutinesCnt {
		goRoutinesCnt = d.dumpGoRoutinesCnt
	}
	dumpCh := make(chan arpDump, goRoutinesCnt)

	// invoke multiple go routines for more efficient parallel dumping
	for idx := 0; idx < goRoutinesCnt; idx++ {
		if goRoutinesCnt > 1 {
			go d.dumpARPs(interfaces, idx, goRoutinesCnt, dumpCh)
		} else {
			d.dumpARPs(interfaces, idx, goRoutinesCnt, dumpCh)
		}
	}

	// collect results from the go routines
	for idx := 0; idx < goRoutinesCnt; idx++ {
		arpDump := <-dumpCh
		if arpDump.err != nil {
			return dump, arpDump.err
		}
		dump = append(dump, arpDump.arps...)
	}

	return dump, nil
}

// dumpARPs is run by a separate go routine to dump all ARP entries associated
// with every <goRoutineIdx>-th interface.
func (d *ARPDescriptor) dumpARPs(interfaces []string, goRoutineIdx, goRoutinesCnt int, dumpCh chan<- arpDump) {
	var dump arpDump
	ifMetaIdx := d.ifPlugin.GetInterfaceIndex()
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()

	for i := goRoutineIdx; i < len(interfaces); i += goRoutinesCnt {
		ifName := interfaces[i]
		// get interface metadata
		ifMeta, found := ifMetaIdx.LookupByName(ifName)
		if !found || ifMeta == nil {
			dump.err = errors.Errorf("failed to obtain metadata for interface %s", ifName)
			d.log.Error(dump.err)
			break
		}

		// switch to the namespace of the interface
		revertNs, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
		if err != nil {
			// namespace and all the ARPs it had contained no longer exist
			d.log.WithFields(logging.Fields{
				"err":       err,
				"namespace": ifMeta.Namespace,
			}).Warn("Failed to dump namespace")
			continue
		}

		// get ARPs assigned to this interface
		arps, err := d.l3Handler.GetARPEntries(ifMeta.LinuxIfIndex)
		revertNs()
		if err != nil {
			dump.err = err
			d.log.Error(dump.err)
			break
		}

		// convert each ARP from Netlink representation to the NB representation
		for _, arp := range arps {
			if arp.IP.IsLinkLocalMulticast() {
				// skip link-local multi-cast ARPs until there is a requirement to support them as well
				continue
			}
			ipAddr := arp.IP.String()
			hwAddr := arp.HardwareAddr.String()

			dump.arps = append(dump.arps, adapter.ARPKVWithMetadata{
				Key: l3.StaticArpKey(ifName, ipAddr),
				Value: &l3.StaticARPEntry{
					Interface: ifName,
					IpAddress: ipAddr,
					HwAddress: hwAddr,
				},
				Origin: scheduler.UnknownOrigin, // let the scheduler to determine the origin
			})
		}
	}

	dumpCh <- dump
}
