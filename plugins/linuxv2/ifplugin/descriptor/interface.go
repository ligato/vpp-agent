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
	"reflect"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"

	"github.com/ligato/cn-infra/idxmap"
	scheduler "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/addrs"

	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/ifaceidx"
	iflinuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxv2/model/namespace"
	"github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin"
	nsdescriptor "github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin/descriptor"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin/linuxcalls"
)

const (
	// InterfaceDescriptorName is the name of the descriptor for Linux interfaces.
	InterfaceDescriptorName = "linux-interfaces"

	// defaultEthernetMTU - expected when MTU is not specified in the config.
	defaultEthernetMTU = 1500

	// dependency labels
	tapInterfaceDep = "vpp-tap-interface"
	vethPeerDep     = "veth-peer"
	microserviceDep = "microservice"

	// suffix attached to logical names of duplicate VETH interfaces
	vethDuplicateSuffix = "-DUPLICATE"

	// suffix attached to logical names of VETH interfaces with peers not found by Dump
	vethMissingPeerSuffix = "-MISSING_PEER"
)

// A list of non-retriable errors:
var (
	// ErrUnsupportedLinuxInterfaceType is returned for Linux interfaces of unknown type.
	ErrUnsupportedLinuxInterfaceType = errors.New("unsupported Linux interface type")

	// ErrInterfaceWithoutName is returned when Linux interface configuration has undefined
	// Name attribute.
	ErrInterfaceWithoutName = errors.New("Linux interface defined without logical name")

	// ErrInterfaceWithoutType is returned when Linux interface configuration has undefined
	// Type attribute.
	ErrInterfaceWithoutType = errors.New("Linux interface defined without type")

	// ErrNamespaceWithoutReference is returned when namespace is missing reference.
	ErrInterfaceReferenceMismatch = errors.New("Linux interface reference does not match the interface type")

	// ErrVETHWithoutPeer is returned when VETH interface is missing peer interface
	// reference.
	ErrVETHWithoutPeer = errors.New("VETH interface defined without peer reference")

	// ErrNamespaceWithoutReference is returned when namespace is missing reference.
	ErrNamespaceWithoutReference = errors.New("namespace defined without name")
)

// InterfaceDescriptor teaches KVScheduler how to configure Linux interfaces.
type InterfaceDescriptor struct {
	log          logging.Logger
	serviceLabel servicelabel.ReaderAPI
	ifHandler    iflinuxcalls.NetlinkAPI
	nsPlugin     nsplugin.API
	scheduler    scheduler.KVScheduler
}

// NewInterfaceDescriptor creates a new instance of the Interface descriptor.
func NewInterfaceDescriptor(
	scheduler scheduler.KVScheduler, serviceLabel servicelabel.ReaderAPI, nsPlugin nsplugin.API,
	ifHandler iflinuxcalls.NetlinkAPI, log logging.PluginLogger) *InterfaceDescriptor {

	return &InterfaceDescriptor{
		scheduler:    scheduler,
		ifHandler:    ifHandler,
		nsPlugin:     nsPlugin,
		serviceLabel: serviceLabel,
		log:          log.NewLogger("-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (intfd *InterfaceDescriptor) GetDescriptor() *adapter.InterfaceDescriptor {
	return &adapter.InterfaceDescriptor{
		Name:               InterfaceDescriptorName,
		KeySelector:        intfd.IsInterfaceKey,
		ValueTypeName:      proto.MessageName(&interfaces.LinuxInterface{}),
		KeyLabel:           intfd.InterfaceNameFromKey,
		ValueComparator:    intfd.EquivalentInterfaces,
		NBKeyPrefix:        interfaces.InterfaceKeyPrefix,
		WithMetadata:       true,
		MetadataMapFactory: intfd.MetadataFactory,
		Add:                intfd.Add,
		Delete:             intfd.Delete,
		Modify:             intfd.Modify,
		ModifyWithRecreate: intfd.ModifyWithRecreate,
		IsRetriableFailure: intfd.IsRetriableFailure,
		Dependencies:       intfd.Dependencies,
		DerivedValues:      intfd.DerivedValues,
		Dump:               intfd.Dump,
		DumpDependencies:   []string{nsdescriptor.MicroserviceDescriptorName},
	}
}

// IsInterfaceKey returns true if the key identifying Linux interface configuration.
func (intfd *InterfaceDescriptor) IsInterfaceKey(key string) bool {
	return strings.HasPrefix(key, interfaces.InterfaceKeyPrefix)
}

// InterfaceNameFromKey returns Linux interface name from the key.
func (intfd *InterfaceDescriptor) InterfaceNameFromKey(key string) string {
	return strings.TrimPrefix(key, interfaces.InterfaceKeyPrefix)
}

// EquivalentInterfaces is case-insensitive comparison function for
// interfaces.LinuxInterface, also ignoring the order of assigned IP addresses.
func (intfd *InterfaceDescriptor) EquivalentInterfaces(key string, intf1, intf2 *interfaces.LinuxInterface) bool {
	// attributes compared as usually:
	if intf1.Name != intf2.Name || intf1.Type != intf2.Type || intf1.Enabled != intf2.Enabled ||
		getHostIfName(intf1) != getHostIfName(intf2) {
		return false
	}
	if intf1.Type == interfaces.LinuxInterface_VETH && getVethPeerName(intf1) != getVethPeerName(intf2) {
		return false
	}
	if !proto.Equal(intf1.Namespace, intf2.Namespace) {
		return false
	}

	// handle undefined TAP temporary name
	if intf1.Type == interfaces.LinuxInterface_AUTO_TAP && getTapTempHostName(intf1) != getTapTempHostName(intf2) {
		return false
	}

	// handle default MTU
	if getInterfaceMTU(intf1) != getInterfaceMTU(intf2) {
		return false
	}

	// compare MAC addresses case-insensitively
	if strings.ToLower(intf1.PhysAddress) != strings.ToLower(intf2.PhysAddress) {
		return false
	}

	// order-irrelevant comparison of IP addresses
	intf1Addrs, err1 := addrs.StrAddrsToStruct(intf1.IpAddresses)
	intf2Addrs, err2 := addrs.StrAddrsToStruct(intf2.IpAddresses)
	if err1 != nil || err2 != nil {
		// one or both of the configurations are invalid, compare lazily
		return reflect.DeepEqual(intf1.IpAddresses, intf2.IpAddresses)
	}
	obsolete, new := addrs.DiffAddr(intf1Addrs, intf2Addrs)
	return len(obsolete) == 0 && len(new) == 0
}

// MetadataFactory is a factory for index-map customized for Linux interfaces.
func (intfd *InterfaceDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return ifaceidx.NewLinuxIfIndex(logrus.DefaultLogger(), "linux-interface-index")
}

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (intfd *InterfaceDescriptor) IsRetriableFailure(err error) bool {
	nonRetriable := []error{
		ErrUnsupportedLinuxInterfaceType,
		ErrInterfaceWithoutName,
		ErrInterfaceWithoutType,
		ErrInterfaceReferenceMismatch,
		ErrVETHWithoutPeer,
		ErrNamespaceWithoutReference,
		}
	for _, nonRetriableErr := range nonRetriable {
		if err == nonRetriableErr {
			return false
		}
	}
	return true
}

// Add creates VETH or configures TAP interface.
func (intfd *InterfaceDescriptor) Add(key string, linuxIf *interfaces.LinuxInterface) (metadata *ifaceidx.LinuxIfMetadata, err error) {
	// validate configuration first
	err = validateInterfaceConfig(linuxIf)
	if err != nil {
		return nil, err
	}

	// create interface based on its type
	switch linuxIf.Type {
	case interfaces.LinuxInterface_VETH:
		metadata, err = intfd.addVETH(key, linuxIf)
	case interfaces.LinuxInterface_AUTO_TAP:
		metadata, err = intfd.addAutoTAP(key, linuxIf)
	default:
		return nil, ErrUnsupportedLinuxInterfaceType
	}

	if err != nil {
		return nil, err
	}

	// move to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := intfd.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
	if err != nil {
		intfd.log.Error(err)
		return nil, err
	}
	defer revert()

	// set interface up
	hostName := getHostIfName(linuxIf)
	if linuxIf.Enabled {
		err = intfd.ifHandler.SetInterfaceUp(hostName)
		if nil != err {
			err = errors.Errorf("failed to set linux interface %s up: %v", linuxIf.Name, err)
			intfd.log.Error(err)
			return nil, err
		}
	}

	// set interface MAC address
	if linuxIf.PhysAddress != "" {
		err = intfd.ifHandler.SetInterfaceMac(hostName, linuxIf.PhysAddress)
		if err != nil {
			err = errors.Errorf("failed to set MAC address %s to linux interface %s: %v",
				linuxIf.PhysAddress, linuxIf.Name, err)
			intfd.log.Error(err)
			return nil, err
		}
	}

	// set interface IP addresses
	ipAddresses, err := addrs.StrAddrsToStruct(linuxIf.IpAddresses)
	if err != nil {
		err = errors.Errorf("failed to convert IP addresses %v for interface %s: %v",
			linuxIf.IpAddresses, linuxIf.Name, err)
		intfd.log.Error(err)
		return nil, err
	}
	for _, ipAddress := range ipAddresses {
		err = intfd.ifHandler.AddInterfaceIP(hostName, ipAddress)
		if err != nil {
			err = errors.Errorf("failed to add IP address %v to linux interface %s: %v",
				ipAddress, linuxIf.Name, err)
			intfd.log.Error(err)
			return nil, err
		}
	}

	// set interface MTU
	if linuxIf.Mtu != 0 {
		mtu := int(linuxIf.Mtu)
		err = intfd.ifHandler.SetInterfaceMTU(hostName, mtu)
		if err != nil {
			err = errors.Errorf("failed to set MTU %d to linux interface %s: %v",
				mtu, linuxIf.Name, err)
			intfd.log.Error(err)
			return nil, err
		}
	}

	return metadata, nil
}

// Delete removes VETH or unconfigures TAP interface.
func (intfd *InterfaceDescriptor) Delete(key string, linuxIf *interfaces.LinuxInterface, metadata *ifaceidx.LinuxIfMetadata) error {
	// move to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := intfd.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
	if err != nil {
		intfd.log.Error(err)
		return err
	}
	defer revert()

	// unassign IP addresses
	ipAddresses, err := addrs.StrAddrsToStruct(linuxIf.IpAddresses)
	if err != nil {
		err = errors.Errorf("failed to convert IP addresses %v for interface %s: %v",
			linuxIf.IpAddresses, linuxIf.Name, err)
		intfd.log.Error(err)
		return err
	}
	for _, ipAddress := range ipAddresses {
		err = intfd.ifHandler.DelInterfaceIP(getHostIfName(linuxIf), ipAddress)
		if err != nil {
			err = errors.Errorf("failed to remove IP address %v from linux interface %s: %v",
				ipAddress, linuxIf.Name, err)
			intfd.log.Error(err)
			return err
		}
	}

	switch linuxIf.Type {
	case interfaces.LinuxInterface_VETH:
		return intfd.deleteVETH(nsCtx, key, linuxIf, metadata)
	case interfaces.LinuxInterface_AUTO_TAP:
		return intfd.deleteAutoTAP(nsCtx, key, linuxIf, metadata)
	}

	err = ErrUnsupportedLinuxInterfaceType
	intfd.log.Error(err)
	return err
}

// Modify is able to change Type-unspecific attributes.
func (intfd *InterfaceDescriptor) Modify(key string, oldLinuxIf, newLinuxIf *interfaces.LinuxInterface, oldMetadata *ifaceidx.LinuxIfMetadata) (newMetadata *ifaceidx.LinuxIfMetadata, err error) {
	oldHostName := getHostIfName(oldLinuxIf)
	newHostName := getHostIfName(newLinuxIf)

	// validate the new configuration first
	err = validateInterfaceConfig(newLinuxIf)
	if err != nil {
		return oldMetadata, err
	}
	// move to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := intfd.nsPlugin.SwitchToNamespace(nsCtx, oldLinuxIf.Namespace)
	if err != nil {
		intfd.log.Error(err)
		return nil, err
	}
	defer revert()

	// update host name
	if oldHostName != newHostName {
		intfd.ifHandler.RenameInterface(oldHostName, newHostName)
		if err != nil {
			intfd.log.Error(err)
			return nil, err
		}
	}

	// update admin status
	if oldLinuxIf.Enabled != newLinuxIf.Enabled {
		if newLinuxIf.Enabled {
			err = intfd.ifHandler.SetInterfaceUp(newHostName)
			if nil != err {
				err = errors.Errorf("failed to set linux interface %s UP: %v", newHostName, err)
				intfd.log.Error(err)
				return nil, err
			}
		} else {
			err = intfd.ifHandler.SetInterfaceDown(newHostName)
			if nil != err {
				err = errors.Errorf("failed to set linux interface %s DOWN: %v", newHostName, err)
				intfd.log.Error(err)
				return nil, err
			}
		}
	}

	// update MAC address
	if newLinuxIf.PhysAddress != "" && newLinuxIf.PhysAddress != oldLinuxIf.PhysAddress {
		err := intfd.ifHandler.SetInterfaceMac(newLinuxIf.HostIfName, newLinuxIf.PhysAddress)
		if err != nil {
			err = errors.Errorf("failed to reconfigure MAC address for linux interface %s: %v",
				newLinuxIf.Name, err)
			intfd.log.Error(err)
			return nil, err
		}
	}

	// IP addresses
	newAddrs, err := addrs.StrAddrsToStruct(newLinuxIf.IpAddresses)
	if err != nil {
		err = errors.Errorf("linux interface modify: failed to convert IP addresses for %s: %v",
			newLinuxIf.Name, err)
		intfd.log.Error(err)
		return nil, err
	}
	oldAddrs, err := addrs.StrAddrsToStruct(oldLinuxIf.IpAddresses)
	if err != nil {
		err = errors.Errorf("linux interface modify: failed to convert IP addresses for %s: %v",
			newLinuxIf.Name, err)
		intfd.log.Error(err)
		return nil, err
	}
	var del, add []*net.IPNet
	del, add = addrs.DiffAddr(newAddrs, oldAddrs)

	for i := range del {
		err := intfd.ifHandler.DelInterfaceIP(newLinuxIf.HostIfName, del[i])
		if nil != err {
			err = errors.Errorf("failed to remove IPv4 address from a Linux interface %s: %v",
				newLinuxIf.Name, err)
			intfd.log.Error(err)
			return nil, err
		}
	}

	for i := range add {
		err := intfd.ifHandler.AddInterfaceIP(newLinuxIf.HostIfName, add[i])
		if nil != err {
			err = errors.Errorf("linux interface modify: failed to add IP addresses %s to %s: %v",
				add[i], newLinuxIf.Name, err)
			intfd.log.Error(err)
			return nil, err
		}
	}

	// MTU
	if getInterfaceMTU(newLinuxIf) != getInterfaceMTU(oldLinuxIf) {
		mtu := getInterfaceMTU(newLinuxIf)
		err := intfd.ifHandler.SetInterfaceMTU(newLinuxIf.HostIfName, mtu)
		if nil != err {
			err = errors.Errorf("failed to reconfigure MTU for the linux interface %s: %v",
				newLinuxIf.Name, err)
			intfd.log.Error(err)
			return nil, err
		}
	}

	// update metadata
	link, err := intfd.ifHandler.GetLinkByName(newHostName)
	if err != nil {
		intfd.log.Error(err)
		return nil, err
	}
	oldMetadata.LinuxIfIndex = link.Attrs().Index
	return oldMetadata, nil
}

// ModifyWithRecreate returns true if Type or Type-specific attributes are different.
func (intfd *InterfaceDescriptor) ModifyWithRecreate(key string, oldLinuxIf, newLinuxIf *interfaces.LinuxInterface, metadata *ifaceidx.LinuxIfMetadata) bool {
	if oldLinuxIf.Type != newLinuxIf.Type {
		return true
	}
	if !proto.Equal(oldLinuxIf.Namespace, newLinuxIf.Namespace) {
		// anything attached to the interface (ARPs, routes, ...) will be re-created as well
		return true
	}
	switch oldLinuxIf.Type {
	case interfaces.LinuxInterface_VETH:
		return getVethPeerName(oldLinuxIf) != getVethPeerName(newLinuxIf)
	case interfaces.LinuxInterface_AUTO_TAP:
		return getTapTempHostName(oldLinuxIf) != getTapTempHostName(newLinuxIf)
	}
	return false
}

// Dependencies lists dependencies for a Linux interface.
func (intfd *InterfaceDescriptor) Dependencies(key string, linuxIf *interfaces.LinuxInterface) []scheduler.Dependency {
	var dependencies []scheduler.Dependency

	// TODO: once VPP-ifplugin is refactored, use reference to derived linux-side of the TAP interface instead
	// (this dependency will not be satisfied as soon as the interface is moved to another ns)
	if linuxIf.Type == interfaces.LinuxInterface_AUTO_TAP {
		dependencies = append(dependencies, scheduler.Dependency{
			Label: tapInterfaceDep,
			Key:   interfaces.InterfaceHostNameKey(getTapTempHostName(linuxIf)),
		})
	}

	// circular dependency between VETH ends
	if linuxIf.Type == interfaces.LinuxInterface_VETH {
		peerName := getVethPeerName(linuxIf)
		if peerName != "" {
			dependencies = append(dependencies, scheduler.Dependency{
				Label: vethPeerDep,
				Key:   interfaces.InterfaceKey(peerName),
			})
		}
	}

	if linuxIf.Namespace != nil && linuxIf.Namespace.Type == namespace.LinuxNetNamespace_NETNS_REF_MICROSERVICE {
		dependencies = append(dependencies, scheduler.Dependency{
			Label: microserviceDep,
			Key:   namespace.MicroserviceKey(linuxIf.Namespace.Reference),
		})
	}

	return dependencies
}

// DerivedValues derives one empty value to represent interface state and also
// one empty value for every IP address assigned to the interface.
func (intfd *InterfaceDescriptor) DerivedValues(key string, linuxIf *interfaces.LinuxInterface) (derValues []scheduler.KeyValuePair) {
	// interface state
	derValues = append(derValues, scheduler.KeyValuePair{
		Key:   interfaces.InterfaceStateKey(linuxIf.Name, linuxIf.Enabled),
		Value: &prototypes.Empty{},
	})
	// IP addresses
	for _, ipAddr := range linuxIf.IpAddresses {
		derValues = append(derValues, scheduler.KeyValuePair{
			Key:   interfaces.InterfaceAddressKey(linuxIf.Name, ipAddr),
			Value: &prototypes.Empty{},
		})
	}
	return derValues
}

// Dump returns all Linux interfaces managed by this agent, attached to the default namespace
// or to one of the configured non-default namespaces.
func (intfd *InterfaceDescriptor) Dump(correlate []adapter.InterfaceKVWithMetadata) ([]adapter.InterfaceKVWithMetadata, error) {
	agentPrefix := intfd.serviceLabel.GetAgentPrefix()
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	nsList := []*namespace.LinuxNetNamespace{nil}              // nil = default namespace, which always should be dumped
	ifCfg := make(map[string]*interfaces.LinuxInterface)       // interface logical name -> interface config (as expected by correlate)
	ifDump := make(map[string]adapter.InterfaceKVWithMetadata) // interface logical name -> interface dump
	indexes := make(map[int]struct{})                          // already dumped interfaces by their Linux indexes

	// process interfaces for correlation to get:
	//  - the set of namespaces to dump
	//  - mapping between interface and the destination namespace
	// beware: the same namespace can have multiple different references (e.g. integration of Contiv with SFC)
	for _, kv := range correlate {
		nsListed := false
		for _, ns := range nsList {
			if proto.Equal(ns, kv.Value.Namespace) {
				nsListed = true
				break
			}
		}
		if !nsListed {
			nsList = append(nsList, kv.Value.Namespace)
		}
		ifCfg[kv.Value.Name] = kv.Value
	}

	// dump every namespace mentioned in the correlate
	var (
		err    error
		revert func()
	)
	for _, nsRef := range nsList {
		// switch to the namespace
		if nsRef != nil {
			revert, err = intfd.nsPlugin.SwitchToNamespace(nsCtx, nsRef)
			if err != nil {
				intfd.log.WithFields(logging.Fields{
					"err":       err,
					"namespace": nsRef,
				}).Debug("Failed to dump namespace")
				continue // continue with the next namespace
			}
		}

		// get all links in the namespace
		links, err := intfd.ifHandler.GetLinkList()
		if err != nil {
			// switch back to the default namespace before returning error
			if nsRef != nil {
				revert()
			}
			intfd.log.Error(err)
			return nil, err
		}

		// dump every interface (at most once!) managed by this agent
		for _, link := range links {
			intf := &interfaces.LinuxInterface{
				Namespace:   nsRef,
				HostIfName:  link.Attrs().Name,
				Enabled:     isInterfaceEnabled(link),
				PhysAddress: link.Attrs().HardwareAddr.String(),
				Mtu:         uint32(link.Attrs().MTU),
			}

			alias := link.Attrs().Alias
			if !strings.HasPrefix(alias, agentPrefix) {
				// skip interface not configured by this agent
				continue
			}
			alias = strings.TrimPrefix(alias, agentPrefix)

			// parse alias to obtain logical references
			var tapTempIfName string
			if link.Type() == (&netlink.Veth{}).Type() {
				var vethPeerIfName string
				intf.Type = interfaces.LinuxInterface_VETH
				intf.Name, vethPeerIfName = parseVethAlias(alias)
				intf.Link = &interfaces.LinuxInterface_VethPeerIfName{VethPeerIfName: vethPeerIfName}
			} else if link.Type() == (&netlink.Tuntap{}).Type() {
				intf.Type = interfaces.LinuxInterface_AUTO_TAP
				intf.Name, tapTempIfName = parseTapAlias(alias)
				intf.Link = &interfaces.LinuxInterface_TapTempIfName{TapTempIfName: tapTempIfName}
			} else {
				// unsupported interface type supposedly configured by agent => print warning
				intfd.log.WithFields(logging.Fields{
					"if-host-name": link.Attrs().Name,
					"namespace":    nsRef,
				}).Warn("Managed interface of unsupported type")
				continue
			}

			// skip interfaces with invalid aliases
			if intf.Name == "" {
				continue
			}

			// skip if this interface was already dumped and this is not the expected
			// namespace from correlation - remember, the same namespace may have
			// multiple different references
			rewrite := false
			if _, dumped := indexes[link.Attrs().Index]; dumped {
				if expCfg, hasExpCfg := ifCfg[intf.Name]; hasExpCfg {
					if proto.Equal(expCfg.Namespace, nsRef) {
						rewrite = true
					}
				}
				if !rewrite {
					continue
				}
			}
			indexes[link.Attrs().Index] = struct{}{}

			// test for duplicity of VETH logical names
			if intf.Type == interfaces.LinuxInterface_VETH {
				if _, duplicate := ifDump[intf.Name]; duplicate && !rewrite {
					// add suffix to the duplicate to make its logical name unique
					// (and not configured by NB so that it will get removed)
					dupIndex := 1
					for intf2 := range ifDump {
						if strings.HasPrefix(intf2, intf.Name + vethDuplicateSuffix) {
							dupIndex++
						}
					}
					intf.Name = intf.Name + vethDuplicateSuffix + strconv.Itoa(dupIndex)
				}
			}

			// dump assigned IP addresses
			addressList, err := intfd.ifHandler.GetAddressList(link.Attrs().Name)
			if err != nil {
				intfd.log.WithFields(logging.Fields{
					"if-host-name": link.Attrs().Name,
					"namespace":    nsRef,
				}).Warn("Failed to read IP addresses")
			}
			for _, address := range addressList {
				if address.Scope == unix.RT_SCOPE_LINK {
					// ignore link-local IPv6 addresses
					continue
				}
				mask, _ := address.Mask.Size()
				addrStr := address.IP.String() + "/" + strconv.Itoa(mask)
				intf.IpAddresses = append(intf.IpAddresses, addrStr)
			}

			// clear attributes unspecified in the config
			// TODO: consider handling of the unspecified attributes in the ValueComparator (by adding origin).
			if expCfg, hasExpCfg := ifCfg[intf.Name]; hasExpCfg {
				if expCfg.PhysAddress == "" {
					intf.PhysAddress = ""
				}
			}

			// build key-value pair for the dumped interface
			ifDump[intf.Name] = adapter.InterfaceKVWithMetadata{
				Key:    interfaces.InterfaceKey(intf.Name),
				Value:  intf,
				Origin: scheduler.FromNB,
				Metadata: &ifaceidx.LinuxIfMetadata{
					LinuxIfIndex: link.Attrs().Index,
					TapTempName:  tapTempIfName,
					Namespace:    nsRef,
				},
			}
		}

		// switch back to the default namespace
		if nsRef != nil {
			revert()
		}
	}

	// first collect VETHs with duplicate logical names
	var dump []adapter.InterfaceKVWithMetadata
	for ifName, kv := range ifDump {
		if kv.Value.Type == interfaces.LinuxInterface_VETH {
			isDuplicate := strings.Contains(ifName, vethDuplicateSuffix)
			// first interface dumped from the set of duplicate VETHs still
			// does not have the vethDuplicateSuffix appended to the name
			_, hasDuplicate := ifDump[ifName + vethDuplicateSuffix + "1"]
			if hasDuplicate {
				kv.Value.Name = ifName + vethDuplicateSuffix + "0"
				kv.Key = interfaces.InterfaceKey(kv.Value.Name)
			}
			if isDuplicate || hasDuplicate {
				// clear peer reference so that Delete removes the VETH-end
				// as standalone
				kv.Value.Link = &interfaces.LinuxInterface_VethPeerIfName{}
				delete(ifDump, ifName)
				dump = append(dump, kv)
			}
		}
	}

	// next collect VETHs with missing peer
	for ifName, kv := range ifDump {
		if kv.Value.Type == interfaces.LinuxInterface_VETH {
			peer, dumped := ifDump[getVethPeerName(kv.Value)]
			if !dumped || getVethPeerName(peer.Value) != kv.Value.Name {
				// append vethMissingPeerSuffix to the logical name so that VETH
				// will get removed during resync
				kv.Value.Name = ifName + vethMissingPeerSuffix
				kv.Key = interfaces.InterfaceKey(kv.Value.Name)
				// clear peer reference so that Delete removes the VETH-end
				// as standalone
				kv.Value.Link = &interfaces.LinuxInterface_VethPeerIfName{}
				delete(ifDump, ifName)
				dump = append(dump, kv)
			}
		}
	}

	// finally collect AUTO-TAPs and valid VETHs
	for _, kv := range ifDump {
		dump = append(dump, kv)
	}

	intfd.log.WithField("dump", dump).Debug("Dumping Linux interfaces")
	return dump, nil
}

// setInterfaceNamespace moves linux interface from the current to the desired
// namespace.
func (intfd *InterfaceDescriptor) setInterfaceNamespace(ctx nslinuxcalls.NamespaceMgmtCtx, ifName string, namespace *namespace.LinuxNetNamespace) error {
	// Get namespace handle.
	ns, err := intfd.nsPlugin.GetNamespaceHandle(ctx, namespace)
	if err != nil {
		return err
	}
	defer ns.Close()

	// Get the interface link handle.
	link, err := intfd.ifHandler.GetLinkByName(ifName)
	if err != nil {
		return errors.Errorf("failed to get link for interface %s: %v", ifName, err)
	}

	// When interface moves from one namespace to another, it loses all its IP addresses, admin status
	// and MTU configuration -- we need to remember the interface configuration before the move
	// and re-configure the interface in the new namespace.
	addresses, isIPv6, err := intfd.getInterfaceAddresses(link.Attrs().Name)
	if err != nil {
		return errors.Errorf("failed to get IP address list from interface %s: %v", link.Attrs().Name, err)
	}

	// Move the interface into the namespace.
	err = intfd.ifHandler.SetLinkNamespace(link, ns)
	if err != nil {
		return errors.Errorf("failed to set interface %s file descriptor: %v", link.Attrs().Name, err)
	}

	// Re-configure interface in its new namespace
	revertNs, err := intfd.nsPlugin.SwitchToNamespace(ctx, namespace)
	if err != nil {
		return errors.Errorf("failed to switch namespace: %v", err)
	}
	defer revertNs()

	if link.Attrs().Flags&net.FlagUp == 1 {
		// Re-enable interface
		err = intfd.ifHandler.SetInterfaceUp(ifName)
		if nil != err {
			return errors.Errorf("failed to re-enable Linux interface `%s`: %v", ifName, err)
		}
	}

	// Re-add IP addresses
	for _, address := range addresses {
		// Skip IPv6 link local address if there is no other IPv6 address
		if !isIPv6 && address.IP.IsLinkLocalUnicast() {
			continue
		}
		err = intfd.ifHandler.AddInterfaceIP(ifName, address)
		if err != nil {
			if err.Error() == "file exists" {
				continue
			}
			return errors.Errorf("failed to re-assign IP address to a Linux interface `%s`: %v", ifName, err)
		}
	}

	// Revert back the MTU config
	err = intfd.ifHandler.SetInterfaceMTU(ifName, link.Attrs().MTU)
	if nil != err {
		return errors.Errorf("failed to re-assign MTU of a Linux interface `%s`: %v", ifName, err)
	}

	return nil
}

// getInterfaceAddresses returns a list of IP addresses assigned to the given linux interface.
// <hasIPv6> is returned as true if a non link-local IPv6 address is among them.
func (intfd *InterfaceDescriptor) getInterfaceAddresses(ifName string) (addresses []*net.IPNet, hasIPv6 bool, err error) {
	// get all assigned IP addresses
	ipAddrs, err := intfd.ifHandler.GetAddressList(ifName)
	if err != nil {
		return nil, false, err
	}

	// iterate over IP addresses to see if there is IPv6 among them
	for _, ipAddr := range ipAddrs {
		if ipAddr.IP.To4() == nil && !ipAddr.IP.IsLinkLocalUnicast() {
			// IP address is version 6 and not a link local address
			hasIPv6 = true
		}
		addresses = append(addresses, ipAddr.IPNet)
	}
	return addresses, hasIPv6, nil
}

// validateInterfaceConfig validates Linux interface configuration.
func validateInterfaceConfig(linuxIf *interfaces.LinuxInterface) error {
	if linuxIf.Name == "" {
		return ErrInterfaceWithoutName
	}
	if linuxIf.Type == interfaces.LinuxInterface_UNDEFINED {
		return ErrInterfaceWithoutType
	}
	if linuxIf.Namespace != nil &&
		(linuxIf.Namespace.Type == namespace.LinuxNetNamespace_NETNS_REF_UNDEFINED ||
	 	linuxIf.Namespace.Reference == "") {
		return ErrNamespaceWithoutReference
	}
	switch linuxIf.Link.(type) {
	case *interfaces.LinuxInterface_TapTempIfName:
		if linuxIf.Type == interfaces.LinuxInterface_VETH {
			return ErrInterfaceReferenceMismatch
		}
	case *interfaces.LinuxInterface_VethPeerIfName:
		if linuxIf.Type == interfaces.LinuxInterface_AUTO_TAP {
			return ErrInterfaceReferenceMismatch
		}
	}
	return nil
}

// getHostIfName returns the interface host name.
func getHostIfName(linuxIf *interfaces.LinuxInterface) string {
	hostIfName := linuxIf.HostIfName
	if hostIfName == "" {
		hostIfName = linuxIf.Name
	}
	return hostIfName
}

// getInterfaceMTU returns the interface MTU.
func getInterfaceMTU(linuxIntf *interfaces.LinuxInterface) int {
	mtu := int(linuxIntf.Mtu)
	if mtu == 0 {
		return defaultEthernetMTU
	}
	return mtu
}

// isInterfaceEnabled returns true if the interface is in the enabled state.
func isInterfaceEnabled(link netlink.Link) bool {
	// - interface of any type is enabled when state is netlink.OperUp,
	// - additionally, VETH may be enabled while the peer is down (OperLowerLayerDown)
	return link.Attrs().OperState == netlink.OperUp ||
		(link.Type() == (&netlink.Veth{}).Type() && link.Attrs().OperState == netlink.OperLowerLayerDown)
}
