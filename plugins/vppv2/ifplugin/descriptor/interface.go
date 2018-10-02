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
	"reflect"
	"strings"

	"github.com/go-errors/errors"
	"github.com/gogo/protobuf/proto"

	"github.com/ligato/cn-infra/idxmap"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/addrs"

	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	linux_ifdescriptor "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/descriptor"
)

const (
	// InterfaceDescriptorName is the name of the descriptor for VPP interfaces.
	InterfaceDescriptorName = "vpp-interfaces"
)

// A list of non-retriable errors:
var (
	// ErrUnsupportedVPPInterfaceType is returned for VPP interfaces of unknown type.
	ErrUnsupportedVPPInterfaceType = errors.New("unsupported VPP interface type")

	// ErrInterfaceWithoutName is returned when VPP interface configuration has undefined
	// Name attribute.
	ErrInterfaceWithoutName = errors.New("VPP interface defined without logical name")

	// ErrInterfaceWithoutType is returned when VPP interface configuration has undefined
	// Type attribute.
	ErrInterfaceWithoutType = errors.New("VPP interface defined without type")
)

// InterfaceDescriptor teaches KVScheduler how to configure VPP interfaces.
type InterfaceDescriptor struct {
	log           logging.Logger
	defaultMtu    uint32
	linuxIfPlugin linux_ifplugin.API /* optional, provide if TAP+AUTO_TAP interfaces are used */
}

// NewInterfaceDescriptor creates a new instance of the Interface descriptor.
func NewInterfaceDescriptor(defaultMtu uint32, linuxIfPlugin linux_ifplugin.API, log logging.PluginLogger) *InterfaceDescriptor {

	return &InterfaceDescriptor{
		defaultMtu:     defaultMtu,
		linuxIfPlugin:  linuxIfPlugin,
		log:            log.NewLogger("-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (intfd *InterfaceDescriptor) GetDescriptor() *adapter.InterfaceDescriptor {
	return &adapter.InterfaceDescriptor{
		Name:               InterfaceDescriptorName,
		KeySelector:        intfd.IsInterfaceKey,
		ValueTypeName:      proto.MessageName(&interfaces.Interface{}),
		KeyLabel:           intfd.InterfaceNameFromKey,
		ValueComparator:    intfd.EquivalentInterfaces,
		NBKeyPrefix:        interfaces.Prefix,
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
		// If Linux-IfPlugin is loaded, dump it first.
		DumpDependencies:   []string{linux_ifdescriptor.InterfaceDescriptorName},
	}
}

// IsInterfaceKey returns true if the key is identifying VPP interface configuration.
func (intfd *InterfaceDescriptor) IsInterfaceKey(key string) bool {
	return strings.HasPrefix(key, interfaces.Prefix)
}

// InterfaceNameFromKey returns VPP interface name from the key.
func (intfd *InterfaceDescriptor) InterfaceNameFromKey(key string) string {
	name, _ := interfaces.ParseNameFromKey(key)
	return name
}

// EquivalentInterfaces is case-insensitive comparison function for
// interfaces.Interface, also ignoring the order of assigned IP addresses.
func (intfd *InterfaceDescriptor) EquivalentInterfaces(key string, intf1, intf2 *interfaces.Interface) bool {
	// attributes compared as usually:
	if intf1.Name != intf2.Name || intf1.Type != intf2.Type || intf1.Enabled != intf2.Enabled ||
		intf1.Vrf != intf2.Vrf || intf1.SetDhcpClient != intf2.SetDhcpClient {
		return false
	}
	if !proto.Equal(intf1.Unnumbered, intf2.Unnumbered) || !proto.Equal(intf1.RxModeSettings, intf2.RxModeSettings) ||
		!proto.Equal(intf1.RxPlacementSettings, intf2.RxPlacementSettings) {
		return false
	}

	// TODO: link, mtu

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

// MetadataFactory is a factory for index-map customized for VPP interfaces.
func (intfd *InterfaceDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return ifaceidx.NewIfaceIndex(logrus.DefaultLogger(), "vpp-interface-index")
}

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (intfd *InterfaceDescriptor) IsRetriableFailure(err error) bool {
	nonRetriable := []error{
		ErrUnsupportedVPPInterfaceType,
		ErrInterfaceWithoutName,
		ErrInterfaceWithoutType,
		}
	for _, nonRetriableErr := range nonRetriable {
		if err == nonRetriableErr {
			return false
		}
	}
	return true
}

// Add creates VETH or configures TAP interface.
func (intfd *InterfaceDescriptor) Add(key string, intf *interfaces.Interface) (metadata *ifaceidx.IfaceMetadata, err error) {
	// TODO
	return metadata, nil
}

// Delete removes VPP interface.
func (intfd *InterfaceDescriptor) Delete(key string, intf *interfaces.Interface, metadata *ifaceidx.IfaceMetadata) error {
	// TODO
	return nil
}

// Modify is able to change Type-unspecific attributes.
func (intfd *InterfaceDescriptor) Modify(key string, oldIntf, newIntf *interfaces.Interface, oldMetadata *ifaceidx.IfaceMetadata) (newMetadata *ifaceidx.IfaceMetadata, err error) {
	// TODO
	return oldMetadata, nil
}

// ModifyWithRecreate returns true if Type or Type-specific attributes are different.
func (intfd *InterfaceDescriptor) ModifyWithRecreate(key string, oldIntf, newIntf *interfaces.Interface, metadata *ifaceidx.IfaceMetadata) bool {
	// TODO
	return false
}

// Dependencies lists dependencies for a VPP interface.
func (intfd *InterfaceDescriptor) Dependencies(key string, intf *interfaces.Interface) []scheduler.Dependency {
	var dependencies []scheduler.Dependency

	// TODO

	return dependencies
}

// DerivedValues derives:
//  - empty value from a TAP interface to represent its Linux-side
//  - one empty value for every IP address assigned to the interface.
func (intfd *InterfaceDescriptor) DerivedValues(key string, intf *interfaces.Interface) (derValues []scheduler.KeyValuePair) {
	// TODO

	return derValues
}

// Dump returns all configured VPP interfaces.
func (intfd *InterfaceDescriptor) Dump(correlate []adapter.InterfaceKVWithMetadata) ([]adapter.InterfaceKVWithMetadata, error) {
	var dump []adapter.InterfaceKVWithMetadata
	// TODO

	intfd.log.WithField("dump", dump).Debug("Dumping VPP interfaces")
	return dump, nil
}

// validateInterfaceConfig validates VPP interface configuration.
func validateInterfaceConfig(intf *interfaces.Interface) error {
	if intf.Name == "" {
		return ErrInterfaceWithoutName
	}
	if intf.Type == interfaces.Interface_UNDEFINED {
		return ErrInterfaceWithoutType
	}
	return nil
}