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

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/ligato/cn-infra/logging"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"

	ifmodel "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	l3 "github.com/ligato/vpp-agent/api/models/linux/l3"
	netalloc_api "github.com/ligato/vpp-agent/api/models/netalloc"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin"
	ifdescriptor "github.com/ligato/vpp-agent/plugins/linux/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/linux/l3plugin/descriptor/adapter"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linux/nsplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/netalloc"
	netalloc_descr "github.com/ligato/vpp-agent/plugins/netalloc/descriptor"
)

const (
	// ARPDescriptorName is the name of the descriptor for Linux ARP entries.
	ARPDescriptorName = "linux-arp"

	// dependency labels
	arpInterfaceDep   = "interface-is-up"
	arpInterfaceIPDep = "interface-has-ip-address"

	// minimum number of interfaces to be given to a single Go routine for processing
	// in the Retrieve operation
	minWorkForGoRoutine = 3
)

// A list of non-retriable errors:
var (
	// ErrARPWithoutInterface is returned when Linux ARP configuration is missing
	// interface reference.
	ErrARPWithoutInterface = errors.New("Linux ARP entry defined without interface reference")

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
	addrAlloc netalloc.AddressAllocator
	scheduler kvs.KVScheduler

	// parallelization of the Retrieve operation
	goRoutinesCnt int
}

// NewARPDescriptor creates a new instance of the ARP descriptor.
func NewARPDescriptor(
	scheduler kvs.KVScheduler, ifPlugin ifplugin.API, nsPlugin nsplugin.API, addrAlloc netalloc.AddressAllocator,
	l3Handler l3linuxcalls.NetlinkAPI, log logging.PluginLogger, goRoutinesCnt int) *kvs.KVDescriptor {

	ctx := &ARPDescriptor{
		scheduler:     scheduler,
		l3Handler:     l3Handler,
		ifPlugin:      ifPlugin,
		nsPlugin:      nsPlugin,
		addrAlloc:     addrAlloc,
		goRoutinesCnt: goRoutinesCnt,
		log:           log.NewLogger("arp-descriptor"),
	}

	typedDescr := &adapter.ARPDescriptor{
		Name:                 ARPDescriptorName,
		NBKeyPrefix:          l3.ModelARPEntry.KeyPrefix(),
		ValueTypeName:        l3.ModelARPEntry.ProtoName(),
		KeySelector:          l3.ModelARPEntry.IsKeyValid,
		KeyLabel:             l3.ModelARPEntry.StripKeyPrefix,
		ValueComparator:      ctx.EquivalentARPs,
		Validate:             ctx.Validate,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Update:               ctx.Update,
		Retrieve:             ctx.Retrieve,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{
			netalloc_descr.IPAllocDescriptorName,
			ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewARPDescriptor(typedDescr)
}

// EquivalentARPs is case-insensitive comparison function for l3.LinuxARPEntry.
// Only MAC addresses are compared - interface and IP address are part of the key
// which is already given to be the same for the two values.
func (d *ARPDescriptor) EquivalentARPs(key string, oldArp, NewArp *l3.ARPEntry) bool {
	// compare MAC addresses case-insensitively
	return strings.ToLower(oldArp.HwAddress) == strings.ToLower(NewArp.HwAddress)
}

// Validate validates ARP entry configuration.
func (d *ARPDescriptor) Validate(key string, arp *l3.ARPEntry) (err error) {
	if arp.Interface == "" {
		return kvs.NewInvalidValueError(ErrARPWithoutInterface, "interface")
	}
	if arp.HwAddress == "" {
		return kvs.NewInvalidValueError(ErrARPWithoutHwAddr, "hw_address")
	}
	return d.addrAlloc.ValidateIPAddress(arp.IpAddress, "", "ip_address", netalloc.GWRefAllowed)
}

// Create creates ARP entry.
func (d *ARPDescriptor) Create(key string, arp *l3.ARPEntry) (metadata interface{}, err error) {
	err = d.updateARPEntry(arp, "add", d.l3Handler.SetARPEntry)
	return nil, err
}

// Delete removes ARP entry.
func (d *ARPDescriptor) Delete(key string, arp *l3.ARPEntry, metadata interface{}) error {
	return d.updateARPEntry(arp, "delete", d.l3Handler.DelARPEntry)
}

// Update is able to change MAC address of the ARP entry.
func (d *ARPDescriptor) Update(key string, oldARP, newARP *l3.ARPEntry, oldMetadata interface{}) (newMetadata interface{}, err error) {
	err = d.updateARPEntry(newARP, "modify", d.l3Handler.SetARPEntry)
	return nil, err
}

// updateARPEntry adds, modifies or deletes an ARP entry.
func (d *ARPDescriptor) updateARPEntry(arp *l3.ARPEntry, actionName string, actionClb func(arpEntry *netlink.Neigh) error) error {
	var err error

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
	ipAddr, err := d.addrAlloc.GetOrParseIPAddress(arp.IpAddress, "",
		netalloc_api.IPAddressForm_ADDR_ONLY)
	if err != nil {
		d.log.Error(err)
		return err
	}
	neigh.IP = ipAddr.IP

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
func (d *ARPDescriptor) Dependencies(key string, arp *l3.ARPEntry) (deps []kvs.Dependency) {
	// the associated interface must exist, but also must be UP and have at least
	// one IP address assigned (to be in the L3 mode)
	if arp.Interface != "" {
		deps = []kvs.Dependency{
			{
				Label: arpInterfaceDep,
				Key:   ifmodel.InterfaceStateKey(arp.Interface, true),
			},
			{
				Label: arpInterfaceIPDep,
				AnyOf: kvs.AnyOfDependency{
					KeyPrefixes: []string{ifmodel.InterfaceAddressPrefix(arp.Interface)},
				},
			},
		}
	}
	// if IP is only a symlink to netalloc address pool, then wait for it to be allocated first
	allocDep, hasAllocDep := d.addrAlloc.GetAddressAllocDep(arp.IpAddress, "", "")
	if hasAllocDep {
		deps = append(deps, allocDep)
	}
	return deps
}

// retrievedARPs is used as the return value sent via channel by retrieveARPs().
type retrievedARPs struct {
	arps []adapter.ARPKVWithMetadata
	err  error
}

// Retrieve returns all ARP entries associated with interfaces managed by this agent.
func (d *ARPDescriptor) Retrieve(correlate []adapter.ARPKVWithMetadata) ([]adapter.ARPKVWithMetadata, error) {
	var values []adapter.ARPKVWithMetadata

	hwLabel := func(arp *l3.ARPEntry) string {
		return arp.Interface + "/" + strings.ToLower(arp.HwAddress)
	}
	expCfg := make(map[string]*l3.ARPEntry) // Interface+MAC -> expected ARP config
	for _, kv := range correlate {
		expCfg[hwLabel(kv.Value)] = kv.Value
	}

	interfaces := d.ifPlugin.GetInterfaceIndex().ListAllInterfaces()
	goRoutinesCnt := len(interfaces) / minWorkForGoRoutine
	if goRoutinesCnt == 0 {
		goRoutinesCnt = 1
	}
	if goRoutinesCnt > d.goRoutinesCnt {
		goRoutinesCnt = d.goRoutinesCnt
	}
	ch := make(chan retrievedARPs, goRoutinesCnt)

	// invoke multiple go routines for more efficient parallel ARP retrieval
	for idx := 0; idx < goRoutinesCnt; idx++ {
		if goRoutinesCnt > 1 {
			go d.retrieveARPs(interfaces, idx, goRoutinesCnt, ch)
		} else {
			d.retrieveARPs(interfaces, idx, goRoutinesCnt, ch)
		}
	}

	// collect results from the go routines
	for idx := 0; idx < goRoutinesCnt; idx++ {
		retrieved := <-ch
		if retrieved.err != nil {
			return values, retrieved.err
		}
		// correlate IP addresses with netalloc references (if any) from the expected config
		for _, arp := range retrieved.arps {
			if expCfg, hasExpCfg := expCfg[hwLabel(arp.Value)]; hasExpCfg {
				arp.Value.IpAddress = d.addrAlloc.CorrelateRetrievedIPs(
					[]string{expCfg.IpAddress}, []string{arp.Value.IpAddress},
					"", netalloc_api.IPAddressForm_ADDR_ONLY)[0]
				// recreate key in case the IP address was replaced with a netalloc link
				arp.Key = l3.ArpKey(arp.Value.Interface, arp.Value.IpAddress)
			}
			values = append(values, arp)
		}
	}

	return values, nil
}

// retrieveARPs is run by a separate go routine to retrieve all ARP entries associated
// with every <goRoutineIdx>-th interface.
func (d *ARPDescriptor) retrieveARPs(interfaces []string, goRoutineIdx, goRoutinesCnt int, ch chan<- retrievedARPs) {
	var retrieved retrievedARPs
	ifMetaIdx := d.ifPlugin.GetInterfaceIndex()
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()

	for i := goRoutineIdx; i < len(interfaces); i += goRoutinesCnt {
		ifName := interfaces[i]
		// get interface metadata
		ifMeta, found := ifMetaIdx.LookupByName(ifName)
		if !found || ifMeta == nil {
			retrieved.err = errors.Errorf("failed to obtain metadata for interface %s", ifName)
			d.log.Error(retrieved.err)
			break
		}

		// switch to the namespace of the interface
		revertNs, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
		if err != nil {
			// namespace and all the ARPs it had contained no longer exist
			d.log.WithFields(logging.Fields{
				"err":       err,
				"namespace": ifMeta.Namespace,
			}).Warn("Failed to retrieve ARPs from the namespace")
			continue
		}

		// get ARPs assigned to this interface
		arps, err := d.l3Handler.GetARPEntries(ifMeta.LinuxIfIndex)
		revertNs()
		if err != nil {
			retrieved.err = err
			d.log.Error(retrieved.err)
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

			retrieved.arps = append(retrieved.arps, adapter.ARPKVWithMetadata{
				Key: l3.ArpKey(ifName, ipAddr),
				Value: &l3.ARPEntry{
					Interface: ifName,
					IpAddress: ipAddr,
					HwAddress: hwAddr,
				},
				Origin: kvs.UnknownOrigin, // let the scheduler to determine the origin
			})
		}
	}

	ch <- retrieved
}
