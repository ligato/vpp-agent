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
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"go.ligato.io/vpp-agent/v3/pkg/models"

	"github.com/golang/protobuf/proto"
	prototypes "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"go.ligato.io/cn-infra/v2/servicelabel"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/ifaceidx"
	iflinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	nsdescriptor "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/descriptor"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
	netalloc_descr "go.ligato.io/vpp-agent/v3/plugins/netalloc/descriptor"
	vpp_ifaceidx "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	namespace "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	netalloc_api "go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	vpp_intf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const (
	// InterfaceDescriptorName is the name of the descriptor for Linux interfaces.
	InterfaceDescriptorName = "linux-interface"

	// default MTU - expected when MTU is not specified in the config.
	defaultEthernetMTU = 1500
	defaultLoopbackMTU = 65536

	// dependency labels
	existingHostInterfaceDep = "host-interface-exists"
	tapInterfaceDep          = "vpp-tap-interface-exists"
	vethPeerDep              = "veth-peer-exists"
	microserviceDep          = "microservice-available"

	// suffix attached to logical names of duplicate VETH interfaces
	vethDuplicateSuffix = "-DUPLICATE"

	// suffix attached to logical names of VETH interfaces with peers not found by Retrieve
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

	// ErrTAPWithoutVPPReference is returned when TAP_TO_VPP interface is missing reference to VPP TAP.
	ErrTAPWithoutVPPReference = errors.New("TAP_TO_VPP interface defined without reference to VPP TAP")

	// ErrTAPRequiresVPPIfPlugin is returned when TAP_TO_VPP is supposed to be configured but VPP ifplugin
	// is not loaded.
	ErrTAPRequiresVPPIfPlugin = errors.New("TAP_TO_VPP interface requires VPP interface plugin to be loaded")

	// ErrNamespaceWithoutReference is returned when namespace is missing reference.
	ErrNamespaceWithoutReference = errors.New("namespace defined without name")

	// ErrExistingWithNamespace is returned when namespace is specified for
	// EXISTING interface.
	ErrExistingWithNamespace = errors.New("EXISTING interface defined with namespace")

	// ErrInvalidIPWithMask is returned when address is invalid or mask is missing
	ErrInvalidIPWithMask = errors.New("IP with mask is not valid")

	// ErrLoopbackAlreadyConfigured is returned when multiple logical NB interfaces tries to configure the same loopback
	ErrLoopbackAlreadyConfigured = errors.New("loopback already configured")

	// ErrLoopbackNotFound is returned if loopback interface can not be found
	ErrLoopbackNotFound = errors.New("loopback not found")
)

// InterfaceDescriptor teaches KVScheduler how to configure Linux interfaces.
type InterfaceDescriptor struct {
	log          logging.Logger
	serviceLabel servicelabel.ReaderAPI
	ifHandler    iflinuxcalls.NetlinkAPI
	nsPlugin     nsplugin.API
	vppIfPlugin  VPPIfPluginAPI
	addrAlloc    netalloc.AddressAllocator

	// runtime
	intfIndex ifaceidx.LinuxIfMetadataIndex
}

// VPPIfPluginAPI is defined here to avoid import cycles.
type VPPIfPluginAPI interface {
	// GetInterfaceIndex gives read-only access to map with metadata of all configured
	// VPP interfaces.
	GetInterfaceIndex() vpp_ifaceidx.IfaceMetadataIndex
}

// NewInterfaceDescriptor creates a new instance of the Interface descriptor.
func NewInterfaceDescriptor(
	serviceLabel servicelabel.ReaderAPI, nsPlugin nsplugin.API, vppIfPlugin VPPIfPluginAPI,
	addrAlloc netalloc.AddressAllocator, log logging.PluginLogger) (descr *kvs.KVDescriptor,
	ctx *InterfaceDescriptor) {

	// descriptor context
	ctx = &InterfaceDescriptor{
		nsPlugin:     nsPlugin,
		vppIfPlugin:  vppIfPlugin,
		addrAlloc:    addrAlloc,
		serviceLabel: serviceLabel,
		log:          log.NewLogger("if-descriptor"),
	}

	typedDescr := &adapter.InterfaceDescriptor{
		Name:               InterfaceDescriptorName,
		NBKeyPrefix:        interfaces.ModelInterface.KeyPrefix(),
		ValueTypeName:      interfaces.ModelInterface.ProtoName(),
		KeySelector:        interfaces.ModelInterface.IsKeyValid,
		KeyLabel:           interfaces.ModelInterface.StripKeyPrefix,
		ValueComparator:    ctx.EquivalentInterfaces,
		WithMetadata:       true,
		MetadataMapFactory: ctx.MetadataFactory,
		Validate:           ctx.Validate,
		Create:             ctx.Create,
		Delete:             ctx.Delete,
		Update:             ctx.Update,
		UpdateWithRecreate: ctx.UpdateWithRecreate,
		Retrieve:           ctx.Retrieve,
		IsRetriableFailure: ctx.IsRetriableFailure,
		DerivedValues:      ctx.DerivedValues,
		Dependencies:       ctx.Dependencies,
		RetrieveDependencies: []string{
			// refresh the pool of allocated IP addresses first
			netalloc_descr.IPAllocDescriptorName,
			nsdescriptor.MicroserviceDescriptorName},
	}
	descr = adapter.NewInterfaceDescriptor(typedDescr)
	return
}

// SetInterfaceIndex should be used to provide interface index immediately after
// the descriptor registration.
func (d *InterfaceDescriptor) SetInterfaceIndex(intfIndex ifaceidx.LinuxIfMetadataIndex) {
	d.intfIndex = intfIndex
}

// SetInterfaceHandler provides interface handler to the descriptor immediately after
// the registration.
func (d *InterfaceDescriptor) SetInterfaceHandler(ifHandler iflinuxcalls.NetlinkAPI) {
	d.ifHandler = ifHandler
}

// EquivalentInterfaces is case-insensitive comparison function for
// interfaces.LinuxInterface, also ignoring the order of assigned IP addresses.
func (d *InterfaceDescriptor) EquivalentInterfaces(key string, oldIntf, newIntf *interfaces.Interface) bool {
	// attributes compared as usually:
	if oldIntf.Name != newIntf.Name ||
		oldIntf.Type != newIntf.Type ||
		oldIntf.Enabled != newIntf.Enabled ||
		oldIntf.LinkOnly != newIntf.LinkOnly ||
		getHostIfName(oldIntf) != getHostIfName(newIntf) {
		return false
	}
	if oldIntf.Type == interfaces.Interface_VETH {
		if oldIntf.GetVeth().GetPeerIfName() != newIntf.GetVeth().GetPeerIfName() {
			return false
		}
		// handle default config for checksum offloading
		if getRxChksmOffloading(oldIntf) != getRxChksmOffloading(newIntf) ||
			getTxChksmOffloading(oldIntf) != getTxChksmOffloading(newIntf) {
			return false
		}
	}
	if oldIntf.Type == interfaces.Interface_TAP_TO_VPP &&
		oldIntf.GetTap().GetVppTapIfName() != newIntf.GetTap().GetVppTapIfName() {
		return false
	}
	if !proto.Equal(oldIntf.Namespace, newIntf.Namespace) {
		return false
	}

	// handle default MTU
	if getInterfaceMTU(oldIntf) != getInterfaceMTU(newIntf) {
		return false
	}

	// for link-only everything else is ignored
	if oldIntf.LinkOnly {
		return true
	}

	// compare MAC addresses case-insensitively (also handle unspecified MAC address)
	if newIntf.PhysAddress != "" &&
		strings.ToLower(oldIntf.PhysAddress) != strings.ToLower(newIntf.PhysAddress) {
		return false
	}

	return true
}

// MetadataFactory is a factory for index-map customized for Linux interfaces.
func (d *InterfaceDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return ifaceidx.NewLinuxIfIndex(logrus.DefaultLogger(), "linux-interface-index")
}

// Validate validates Linux interface configuration.
func (d *InterfaceDescriptor) Validate(key string, linuxIf *interfaces.Interface) error {
	// validate name (this should never happen, since key is derived from name)
	if linuxIf.GetName() == "" {
		return kvs.NewInvalidValueError(ErrInterfaceWithoutName, "name")
	}

	// validate namespace
	if ns := linuxIf.GetNamespace(); ns != nil {
		if ns.GetType() == namespace.NetNamespace_UNDEFINED || ns.GetReference() == "" {
			return kvs.NewInvalidValueError(ErrNamespaceWithoutReference, "namespace")
		}
	}

	// validate type
	switch linuxIf.GetType() {
	case interfaces.Interface_EXISTING:
		if linuxIf.GetLink() != nil {
			return kvs.NewInvalidValueError(ErrInterfaceReferenceMismatch, "link")
		}
		// For now support only the same namespace as the agent.
		if linuxIf.GetNamespace() != nil {
			return kvs.NewInvalidValueError(ErrExistingWithNamespace, "namespace")
		}
	case interfaces.Interface_LOOPBACK:
		if linuxIf.GetLink() != nil {
			return kvs.NewInvalidValueError(ErrInterfaceReferenceMismatch, "link")
		}
	case interfaces.Interface_TAP_TO_VPP:
		if d.vppIfPlugin == nil {
			return ErrTAPRequiresVPPIfPlugin
		}
	case interfaces.Interface_UNDEFINED:
		return kvs.NewInvalidValueError(ErrInterfaceWithoutType, "type")
	}

	// validate link
	switch linuxIf.GetLink().(type) {
	case *interfaces.Interface_Tap:
		if linuxIf.GetType() != interfaces.Interface_TAP_TO_VPP {
			return kvs.NewInvalidValueError(ErrInterfaceReferenceMismatch, "link")
		}
		if linuxIf.GetTap().GetVppTapIfName() == "" {
			return kvs.NewInvalidValueError(ErrTAPWithoutVPPReference, "vpp_tap_if_name")
		}
	case *interfaces.Interface_Veth:
		if linuxIf.GetType() != interfaces.Interface_VETH {
			return kvs.NewInvalidValueError(ErrInterfaceReferenceMismatch, "link")
		}
		if linuxIf.GetVeth().GetPeerIfName() == "" {
			return kvs.NewInvalidValueError(ErrVETHWithoutPeer, "peer_if_name")
		}
	}

	return nil
}

// Create creates VETH or configures TAP interface.
func (d *InterfaceDescriptor) Create(key string, linuxIf *interfaces.Interface) (metadata *ifaceidx.LinuxIfMetadata, err error) {
	// move to the default namespace
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert1, err := d.nsPlugin.SwitchToNamespace(nsCtx, nil)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	defer revert1()

	// create interface based on its type
	switch linuxIf.Type {
	case interfaces.Interface_VETH:
		metadata, err = d.createVETH(nsCtx, key, linuxIf)
	case interfaces.Interface_TAP_TO_VPP:
		metadata, err = d.createTAPToVPP(nsCtx, key, linuxIf)
	case interfaces.Interface_LOOPBACK:
		metadata, err = d.createLoopback(nsCtx, linuxIf)
	case interfaces.Interface_EXISTING:
		// We expect that the interface already exists, therefore nothing needs to be done.
		// We just get the metadata for the interface.
		getMetadata := func(linuxIf *interfaces.Interface) (*ifaceidx.LinuxIfMetadata, error) {
			link, err := d.ifHandler.GetLinkByName(getHostIfName(linuxIf))
			if err != nil {
				d.log.Error(err)
				return nil, err
			}
			return &ifaceidx.LinuxIfMetadata{
				Namespace:    linuxIf.GetNamespace(),
				LinuxIfIndex: link.Attrs().Index,
			}, nil
		}
		metadata, err = getMetadata(linuxIf)
	default:
		return nil, ErrUnsupportedLinuxInterfaceType
	}
	if err != nil {
		d.log.Errorf("creating %v interface failed: %+v", linuxIf.GetType(), err)
		return nil, err
	}

	metadata.HostIfName = getHostIfName(linuxIf)

	// move to the namespace with the interface
	revert2, err := d.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	defer revert2()

	// set interface up
	hostName := getHostIfName(linuxIf)
	if linuxIf.Enabled {
		err = d.ifHandler.SetInterfaceUp(hostName)
		if nil != err {
			err = errors.Errorf("failed to set linux interface %s up: %v", linuxIf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	// set checksum offloading
	if linuxIf.Type == interfaces.Interface_VETH {
		rxOn := getRxChksmOffloading(linuxIf)
		txOn := getTxChksmOffloading(linuxIf)
		err = d.ifHandler.SetChecksumOffloading(hostName, rxOn, txOn)
		if err != nil {
			err = errors.Errorf("failed to configure checksum offloading (rx=%t,tx=%t) for linux interface %s: %v",
				rxOn, txOn, linuxIf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	// set interface MTU
	if linuxIf.Mtu != 0 {
		mtu := int(linuxIf.Mtu)
		err = d.ifHandler.SetInterfaceMTU(hostName, mtu)
		if err != nil {
			err = errors.Errorf("failed to set MTU %d to linux interface %s: %v",
				mtu, linuxIf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	if linuxIf.GetLinkOnly() {
		// addresses are configured externally
		return metadata, nil
	}

	// set interface MAC address
	if linuxIf.PhysAddress != "" {
		err = d.ifHandler.SetInterfaceMac(hostName, linuxIf.PhysAddress)
		if err != nil {
			err = errors.Errorf("failed to set MAC address %s to linux interface %s: %v",
				linuxIf.PhysAddress, linuxIf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	return metadata, nil
}

// Delete removes VETH or unconfigures TAP interface.
func (d *InterfaceDescriptor) Delete(key string, linuxIf *interfaces.Interface, metadata *ifaceidx.LinuxIfMetadata) error {
	// move to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
	if err != nil {
		d.log.Error("switch to namespace failed:", err)
		return err
	}
	defer revert()

	switch linuxIf.Type {
	case interfaces.Interface_VETH:
		return d.deleteVETH(nsCtx, key, linuxIf, metadata)
	case interfaces.Interface_TAP_TO_VPP:
		return d.deleteAutoTAP(nsCtx, key, linuxIf, metadata)
	case interfaces.Interface_LOOPBACK:
		return d.deleteLoopback(nsCtx, linuxIf)
	case interfaces.Interface_EXISTING:
		// We only need to unconfigure the interface.
		// Nothing else needs to be done.
		return nil
	}

	err = ErrUnsupportedLinuxInterfaceType
	d.log.Error(err)
	return err
}

// Update is able to change Type-unspecific attributes.
func (d *InterfaceDescriptor) Update(key string, oldLinuxIf, newLinuxIf *interfaces.Interface, oldMetadata *ifaceidx.LinuxIfMetadata) (newMetadata *ifaceidx.LinuxIfMetadata, err error) {
	oldHostName := getHostIfName(oldLinuxIf)
	newHostName := getHostIfName(newLinuxIf)

	// move to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, oldLinuxIf.Namespace)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	defer revert()

	// update host name
	if oldHostName != newHostName {
		if err := d.ifHandler.RenameInterface(oldHostName, newHostName); err != nil {
			d.log.Error("renaming interface failed:", err)
			return nil, err
		}
	}

	// update admin status
	if oldLinuxIf.Enabled != newLinuxIf.Enabled {
		if newLinuxIf.Enabled {
			err = d.ifHandler.SetInterfaceUp(newHostName)
			if nil != err {
				err = errors.Errorf("failed to set linux interface %s UP: %v", newHostName, err)
				d.log.Error(err)
				return nil, err
			}
		} else {
			err = d.ifHandler.SetInterfaceDown(newHostName)
			if nil != err {
				err = errors.Errorf("failed to set linux interface %s DOWN: %v", newHostName, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	// update MAC address
	if !newLinuxIf.GetLinkOnly() {
		if newLinuxIf.PhysAddress != "" && newLinuxIf.PhysAddress != oldLinuxIf.PhysAddress {
			err := d.ifHandler.SetInterfaceMac(newHostName, newLinuxIf.PhysAddress)
			if err != nil {
				err = errors.Errorf("failed to reconfigure MAC address for linux interface %s: %v",
					newLinuxIf.Name, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	// MTU
	if getInterfaceMTU(newLinuxIf) != getInterfaceMTU(oldLinuxIf) {
		mtu := getInterfaceMTU(newLinuxIf)
		err := d.ifHandler.SetInterfaceMTU(newHostName, mtu)
		if nil != err {
			err = errors.Errorf("failed to reconfigure MTU for the linux interface %s: %v",
				newLinuxIf.Name, err)
			d.log.Error(err)
			return nil, err
		}
	}

	// update checksum offloading
	if newLinuxIf.Type == interfaces.Interface_VETH {
		rxOn := getRxChksmOffloading(newLinuxIf)
		txOn := getTxChksmOffloading(newLinuxIf)
		if rxOn != getRxChksmOffloading(oldLinuxIf) || txOn != getTxChksmOffloading(oldLinuxIf) {
			err = d.ifHandler.SetChecksumOffloading(newHostName, rxOn, txOn)
			if err != nil {
				err = errors.Errorf("failed to reconfigure checksum offloading (rx=%t,tx=%t) for linux interface %s: %v",
					rxOn, txOn, newLinuxIf.Name, err)
				d.log.Error(err)
				return nil, err
			}
		}
	}

	// update metadata
	link, err := d.ifHandler.GetLinkByName(newHostName)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	oldMetadata.LinuxIfIndex = link.Attrs().Index
	oldMetadata.HostIfName = newHostName
	return oldMetadata, nil
}

// UpdateWithRecreate returns true if Type or Type-specific attributes are different.
func (d *InterfaceDescriptor) UpdateWithRecreate(key string, oldLinuxIf, newLinuxIf *interfaces.Interface, metadata *ifaceidx.LinuxIfMetadata) bool {
	if oldLinuxIf.Type != newLinuxIf.Type {
		return true
	}
	if oldLinuxIf.LinkOnly != newLinuxIf.LinkOnly {
		return true
	}
	if !proto.Equal(oldLinuxIf.Namespace, newLinuxIf.Namespace) {
		// anything attached to the interface (ARPs, routes, ...) will be re-created as well
		return true
	}
	switch oldLinuxIf.Type {
	case interfaces.Interface_VETH:
		return oldLinuxIf.GetVeth().GetPeerIfName() != newLinuxIf.GetVeth().GetPeerIfName()
	case interfaces.Interface_TAP_TO_VPP:
		return oldLinuxIf.GetTap().GetVppTapIfName() != newLinuxIf.GetTap().GetVppTapIfName()
	}
	return false
}

// Dependencies lists dependencies for a Linux interface.
func (d *InterfaceDescriptor) Dependencies(key string, linuxIf *interfaces.Interface) []kvs.Dependency {
	var dependencies []kvs.Dependency

	// EXISTING depends on a referenced Linux interface in the default namespace
	if linuxIf.Type == interfaces.Interface_EXISTING {
		dependencies = append(dependencies, kvs.Dependency{
			Label: existingHostInterfaceDep,
			Key:   interfaces.InterfaceHostNameKey(getHostIfName(linuxIf)),
		})
	}

	if linuxIf.Type == interfaces.Interface_TAP_TO_VPP {
		// dependency on VPP TAP
		dependencies = append(dependencies, kvs.Dependency{
			Label: tapInterfaceDep,
			Key:   vpp_intf.InterfaceKey(linuxIf.GetTap().GetVppTapIfName()),
		})
	}

	// circular dependency between VETH ends
	if linuxIf.Type == interfaces.Interface_VETH {
		peerName := linuxIf.GetVeth().GetPeerIfName()
		if peerName != "" {
			dependencies = append(dependencies, kvs.Dependency{
				Label: vethPeerDep,
				Key:   interfaces.InterfaceKey(peerName),
			})
		}
	}

	if linuxIf.GetNamespace().GetType() == namespace.NetNamespace_MICROSERVICE {
		dependencies = append(dependencies, kvs.Dependency{
			Label: microserviceDep,
			Key:   namespace.MicroserviceKey(linuxIf.Namespace.Reference),
		})
	}

	return dependencies
}

// DerivedValues derives one empty value to represent interface state and also
// one empty value for every IP address assigned to the interface.
func (d *InterfaceDescriptor) DerivedValues(key string, linuxIf *interfaces.Interface) (derValues []kvs.KeyValuePair) {
	// interface state
	derValues = append(derValues, kvs.KeyValuePair{
		Key:   interfaces.InterfaceStateKey(linuxIf.Name, linuxIf.Enabled),
		Value: &prototypes.Empty{},
	})
	if !linuxIf.GetLinkOnly() {
		// IP addresses
		for _, ipAddr := range linuxIf.IpAddresses {
			derValues = append(derValues, kvs.KeyValuePair{
				Key:   interfaces.InterfaceAddressKey(linuxIf.Name, ipAddr, netalloc_api.IPAddressSource_STATIC),
				Value: &prototypes.Empty{},
			})
		}
	}
	return derValues
}

// retrievedIfaces is used as the return value sent via channel by retrieveInterfaces().
type retrievedIfaces struct {
	interfaces []adapter.InterfaceKVWithMetadata
	err        error
}

func (d *InterfaceDescriptor) IsRetriableFailure(err error) bool {
	if err == ErrLoopbackAlreadyConfigured {
		return false
	}
	return true
}

// Retrieve returns all Linux interfaces managed by this agent, attached to the default namespace
// or to one of the configured non-default namespaces.
func (d *InterfaceDescriptor) Retrieve(correlate []adapter.InterfaceKVWithMetadata) ([]adapter.InterfaceKVWithMetadata, error) {
	nsList := []*namespace.NetNamespace{nil}              // nil = default namespace, which always should be listed for interfaces
	ifCfg := make(map[string]*interfaces.Interface)       // interface logical name -> interface config (as expected by correlate)
	expExisting := make(map[string]*interfaces.Interface) // EXISTING interface host name -> expected interface config

	// process interfaces for correlation to get:
	//  - the set of namespaces to list for interfaces
	//  - mapping between interface name and the configuration for correlation
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
		if kv.Value.Type == interfaces.Interface_EXISTING {
			expExisting[getHostIfName(kv.Value)] = kv.Value
		}
	}

	// Obtain interface details - all interfaces with metadata
	ifDetails, err := d.ifHandler.DumpInterfacesFromNamespaces(nsList)
	if err != nil {
		return nil, errors.Errorf("Failed to retrieve linux interfaces: %v", err)
	}
	// interface logical name -> interface data
	ifaces := make(map[string]adapter.InterfaceKVWithMetadata)
	// already retrieved interfaces by their Linux indexes
	indexes := make(map[int]struct{})

	for _, ifDetail := range ifDetails {
		// Transform linux interface details to the type-safe value with metadata
		kv := adapter.InterfaceKVWithMetadata{
			Origin: kvs.FromNB,
			Value:  ifDetail.Interface,
			Metadata: &ifaceidx.LinuxIfMetadata{
				LinuxIfIndex: ifDetail.Meta.LinuxIfIndex,
				Namespace:    ifDetail.Interface.GetNamespace(),
				VPPTapName:   ifDetail.Interface.GetTap().GetVppTapIfName(),
				HostIfName:   ifDetail.Interface.HostIfName,
			},
			Key: interfaces.InterfaceKey(ifDetail.Interface.Name),
		}

		// skip if this interface was already retrieved and this is not the expected
		// namespace from correlation - remember, the same namespace may have
		// multiple different references
		var rewrite bool
		if _, alreadyRetrieved := indexes[kv.Metadata.LinuxIfIndex]; alreadyRetrieved {
			if expCfg, hasExpCfg := ifCfg[kv.Value.Name]; hasExpCfg {
				if proto.Equal(expCfg.Namespace, kv.Value.Namespace) {
					rewrite = true
				}
			}
			if !rewrite {
				continue
			}
		}
		indexes[kv.Metadata.LinuxIfIndex] = struct{}{}

		// test for duplicity of VETH logical names
		if kv.Value.Type == interfaces.Interface_VETH {
			if _, duplicate := ifaces[kv.Value.Name]; duplicate && !rewrite {
				// add suffix to the duplicate to make its logical name unique
				// (and not configured by NB so that it will get removed)
				dupIndex := 1
				for intf2 := range ifaces {
					if strings.HasPrefix(intf2, kv.Value.Name+vethDuplicateSuffix) {
						dupIndex++
					}
				}
				kv.Value.Name = kv.Value.Name + vethDuplicateSuffix + strconv.Itoa(dupIndex)
				kv.Key = interfaces.InterfaceKey(kv.Value.Name)
			}
		}
		// correlate link_only attribute
		if expCfg, hasExpCfg := ifCfg[kv.Value.Name]; hasExpCfg {
			kv.Value.LinkOnly = expCfg.GetLinkOnly()
		}
		ifaces[kv.Value.Name] = kv
	}

	// first collect VETHs with duplicate logical names
	var values []adapter.InterfaceKVWithMetadata
	for ifName, kv := range ifaces {
		if kv.Value.Type == interfaces.Interface_VETH {
			isDuplicate := strings.Contains(ifName, vethDuplicateSuffix)
			// first interface retrieved from the set of duplicate VETHs still
			// does not have the vethDuplicateSuffix appended to the name
			_, hasDuplicate := ifaces[ifName+vethDuplicateSuffix+"1"]
			if hasDuplicate {
				kv.Value.Name = ifName + vethDuplicateSuffix + "0"
				kv.Key = interfaces.InterfaceKey(kv.Value.Name)
			}
			if isDuplicate || hasDuplicate {
				// clear peer reference so that Delete removes the VETH-end
				// as standalone
				kv.Value.Link = &interfaces.Interface_Veth{}
				delete(ifaces, ifName)
				values = append(values, kv)
			}
		}
	}

	// next collect VETHs with missing peer
	for ifName, kv := range ifaces {
		if kv.Value.Type == interfaces.Interface_VETH {
			peer, retrieved := ifaces[kv.Value.GetVeth().GetPeerIfName()]
			if !retrieved || peer.Value.GetVeth().GetPeerIfName() != kv.Value.Name {
				// append vethMissingPeerSuffix to the logical name so that VETH
				// will get removed during resync
				kv.Value.Name = ifName + vethMissingPeerSuffix
				kv.Key = interfaces.InterfaceKey(kv.Value.Name)
				// clear peer reference so that Delete removes the VETH-end
				// as standalone
				kv.Value.Link = &interfaces.Interface_Veth{}
				delete(ifaces, ifName)
				values = append(values, kv)
			}
		}
	}

	// collect AUTO-TAPs and valid VETHs
	for _, kv := range ifaces {
		values = append(values, kv)
	}

	// retrieve EXISTING interfaces
	existingIfaces, err := d.retrieveExistingInterfaces(expExisting)
	if err != nil {
		return nil, err
	}
	for _, kv := range existingIfaces {
		values = append(values, kv)
	}

	// correlate IP addresses with netalloc references from the expected config
	for _, kv := range values {
		if expCfg, hasExpCfg := ifCfg[kv.Value.Name]; hasExpCfg {
			kv.Value.IpAddresses = d.addrAlloc.CorrelateRetrievedIPs(
				expCfg.IpAddresses, kv.Value.IpAddresses,
				kv.Value.Name, netalloc_api.IPAddressForm_ADDR_WITH_MASK)
		}
	}

	return values, nil
}

// retrieveExistingInterfaces retrieves already created Linux interface - i.e. not created
// by this agent = type EXISTING.
func (d *InterfaceDescriptor) retrieveExistingInterfaces(expected map[string]*interfaces.Interface) ([]adapter.InterfaceKVWithMetadata, error) {
	var retrieved []adapter.InterfaceKVWithMetadata

	// get all links in the default namespace
	links, err := d.ifHandler.GetLinkList()
	if err != nil {
		d.log.Error("Failed to get link list:", err)
		return nil, err
	}
	for _, link := range links {
		expCfg, isExp := expected[link.Attrs().Name]
		if !isExp {
			// do not touch existing interfaces which are not configured by the agent
			continue
		}
		iface := &interfaces.Interface{
			Name:        expCfg.GetName(),
			Type:        interfaces.Interface_EXISTING,
			HostIfName:  link.Attrs().Name,
			PhysAddress: link.Attrs().HardwareAddr.String(),
			Mtu:         uint32(link.Attrs().MTU),
			LinkOnly:    expCfg.LinkOnly,
		}

		// retrieve addresses, MTU, etc.
		d.retrieveLinkDetails(link, iface, nil)

		// build key-value pair for the retrieved interface
		retrieved = append(retrieved, adapter.InterfaceKVWithMetadata{
			Key:    models.Key(iface),
			Value:  iface,
			Origin: kvs.FromNB,
			Metadata: &ifaceidx.LinuxIfMetadata{
				LinuxIfIndex: link.Attrs().Index,
				HostIfName:   link.Attrs().Name,
			},
		})
	}

	return retrieved, nil
}

// retrieveLinkDetails retrieves link details common to all interface types (e.g. addresses).
func (d *InterfaceDescriptor) retrieveLinkDetails(link netlink.Link, iface *interfaces.Interface, nsRef *namespace.NetNamespace) {
	var err error
	// read interface status
	iface.Enabled, err = d.ifHandler.IsInterfaceUp(link.Attrs().Name)
	if err != nil {
		d.log.WithFields(logging.Fields{
			"if-host-name": link.Attrs().Name,
			"namespace":    nsRef,
		}).Warn("Failed to read interface status:", err)
	}

	// read assigned IP addresses
	addressList, err := d.ifHandler.GetAddressList(link.Attrs().Name)
	if err != nil {
		d.log.WithFields(logging.Fields{
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
		rxOn, txOn, err := d.ifHandler.GetChecksumOffloading(link.Attrs().Name)
		if err != nil {
			d.log.WithFields(logging.Fields{
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

// setInterfaceNamespace moves linux interface from the current to the desired
// namespace.
func (d *InterfaceDescriptor) setInterfaceNamespace(ctx nslinuxcalls.NamespaceMgmtCtx, ifName string, namespace *namespace.NetNamespace) error {
	// Get namespace handle.
	ns, err := d.nsPlugin.GetNamespaceHandle(ctx, namespace)
	if err != nil {
		return err
	}
	defer ns.Close()

	// Get the interface link handle.
	link, err := d.ifHandler.GetLinkByName(ifName)
	if err != nil {
		return errors.Errorf("failed to get link for interface %s: %v", ifName, err)
	}

	// When interface moves from one namespace to another, it loses all its IP addresses, admin status
	// and MTU configuration -- we need to remember the interface configuration before the move
	// and re-configure the interface in the new namespace.
	addresses, isIPv6, err := d.getInterfaceAddresses(link.Attrs().Name)
	if err != nil {
		return errors.Errorf("failed to get IP address list from interface %s: %v", link.Attrs().Name, err)
	}
	enabled, err := d.ifHandler.IsInterfaceUp(ifName)
	if err != nil {
		return errors.Errorf("failed to get admin status of the interface %s: %v", link.Attrs().Name, err)
	}

	// Move the interface into the namespace.
	if err := d.ifHandler.SetLinkNamespace(link, ns); err != nil {
		return errors.Errorf("failed to set interface %s file descriptor: %v", link.Attrs().Name, err)
	}

	// Re-configure interface in its new namespace
	revertNs, err := d.nsPlugin.SwitchToNamespace(ctx, namespace)
	if err != nil {
		return errors.Errorf("failed to switch namespace: %v", err)
	}
	defer revertNs()

	if enabled {
		// Re-enable interface
		err = d.ifHandler.SetInterfaceUp(ifName)
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
		if err := d.ifHandler.AddInterfaceIP(ifName, address); err != nil {
			if err.Error() == "file exists" {
				continue
			}
			return errors.Errorf("failed to re-assign IP address to a Linux interface `%s`: %v", ifName, err)
		}
	}

	// Revert back the MTU config
	err = d.ifHandler.SetInterfaceMTU(ifName, link.Attrs().MTU)
	if nil != err {
		return errors.Errorf("failed to re-assign MTU of a Linux interface `%s`: %v", ifName, err)
	}

	return nil
}

// getInterfaceAddresses returns a list of IP addresses assigned to the given linux interface.
// <hasIPv6> is returned as true if a non link-local IPv6 address is among them.
func (d *InterfaceDescriptor) getInterfaceAddresses(ifName string) (addresses []*net.IPNet, hasIPv6 bool, err error) {
	// get all assigned IP addresses
	ipAddrs, err := d.ifHandler.GetAddressList(ifName)
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

// getHostIfName returns the interface host name.
func getHostIfName(linuxIf *interfaces.Interface) string {
	if linuxIf.Type == interfaces.Interface_LOOPBACK {
		return iflinuxcalls.DefaultLoopbackName
	}
	hostIfName := linuxIf.HostIfName
	if hostIfName == "" {
		hostIfName = linuxIf.Name
	}
	return hostIfName
}

// getInterfaceMTU returns the interface MTU.
func getInterfaceMTU(linuxIntf *interfaces.Interface) int {
	mtu := int(linuxIntf.Mtu)
	if mtu == 0 {
		if linuxIntf.Type == interfaces.Interface_LOOPBACK {
			return defaultLoopbackMTU
		}
		return defaultEthernetMTU
	}
	return mtu
}

func getRxChksmOffloading(linuxIntf *interfaces.Interface) (rxOn bool) {
	return isChksmOffloadingOn(linuxIntf.GetVeth().GetRxChecksumOffloading())
}

func getTxChksmOffloading(linuxIntf *interfaces.Interface) (txOn bool) {
	return isChksmOffloadingOn(linuxIntf.GetVeth().GetTxChecksumOffloading())
}

func isChksmOffloadingOn(offloading interfaces.VethLink_ChecksumOffloading) bool {
	switch offloading {
	case interfaces.VethLink_CHKSM_OFFLOAD_DEFAULT:
		return true // enabled by default
	case interfaces.VethLink_CHKSM_OFFLOAD_ENABLED:
		return true
	case interfaces.VethLink_CHKSM_OFFLOAD_DISABLED:
		return false
	}
	return true
}

func getSysctl(name string) (string, error) {
	fullName := filepath.Join("/proc/sys", strings.Replace(name, ".", "/", -1))
	fullName = filepath.Clean(fullName)
	data, err := ioutil.ReadFile(fullName)
	if err != nil {
		return "", err
	}
	return string(data[:len(data)-1]), nil
}

func setSysctl(name, value string) (string, error) {
	fullName := filepath.Join("/proc/sys", strings.Replace(name, ".", "/", -1))
	fullName = filepath.Clean(fullName)
	if err := ioutil.WriteFile(fullName, []byte(value), 0644); err != nil {
		return "", err
	}
	return getSysctl(name)
}
