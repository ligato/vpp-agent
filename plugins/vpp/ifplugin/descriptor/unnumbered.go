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
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// UnnumberedIfDescriptorName is the name of the descriptor for the unnumbered
	// config-subsection of VPP interfaces.
	UnnumberedIfDescriptorName = "vpp-unnumbered-interface"

	// dependency labels
	unnumberedInterfaceHasIPDep = "unnumbered-interface-has-IP"
)

// UnnumberedIfDescriptor sets/unsets VPP interfaces as unnumbered.
// Values = Interface_Unnumbered{} derived from interfaces where IsUnnumbered==true
type UnnumberedIfDescriptor struct {
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI
	ifIndex   ifaceidx.IfaceMetadataIndex
}

// NewUnnumberedIfDescriptor creates a new instance of UnnumberedIfDescriptor.
func NewUnnumberedIfDescriptor(ifHandler vppcalls.InterfaceVppAPI, ifIndex ifaceidx.IfaceMetadataIndex,
	log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &UnnumberedIfDescriptor{
		ifHandler: ifHandler,
		ifIndex:   ifIndex,
		log:       log.NewLogger("unif-descriptor"),
	}

	typedDescr := &adapter.UnnumberedDescriptor{
		Name:          UnnumberedIfDescriptorName,
		KeySelector:   ctx.IsUnnumberedInterfaceKey,
		ValueTypeName: proto.MessageName(&interfaces.Interface_Unnumbered{}),
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Dependencies:  ctx.Dependencies,
	}

	return adapter.NewUnnumberedDescriptor(typedDescr)
}

// IsUnnumberedInterfaceKey returns true if the key is identifying unnumbered
// VPP interface.
func (d *UnnumberedIfDescriptor) IsUnnumberedInterfaceKey(key string) bool {
	_, isValid := interfaces.ParseNameFromUnnumberedKey(key)
	return isValid
}

// Create sets interface as unnumbered.
func (d *UnnumberedIfDescriptor) Create(key string, unIntf *interfaces.Interface_Unnumbered) (metadata interface{}, err error) {
	ifName, _ := interfaces.ParseNameFromUnnumberedKey(key)

	ifMeta, found := d.ifIndex.LookupByName(ifName)
	if !found {
		err = errors.Errorf("failed to find unnumbered interface %s", ifName)
		d.log.Error(err)
		return nil, err
	}

	ifWithIPMeta, found := d.ifIndex.LookupByName(unIntf.InterfaceWithIp)
	if !found {
		err = errors.Errorf("failed to find interface %s referenced by unnumbered interface %s",
			unIntf.InterfaceWithIp, ifName)
		d.log.Error(err)
		return nil, err
	}

	err = d.ifHandler.SetUnnumberedIP(ifMeta.SwIfIndex, ifWithIPMeta.SwIfIndex)
	if err != nil {
		d.log.Error(err)
	}
	return nil, err
}

// Delete un-sets interface as unnumbered.
func (d *UnnumberedIfDescriptor) Delete(key string, unIntf *interfaces.Interface_Unnumbered, metadata interface{}) error {
	ifName, _ := interfaces.ParseNameFromUnnumberedKey(key)

	ifMeta, found := d.ifIndex.LookupByName(ifName)
	if !found {
		err := errors.Errorf("failed to find unnumbered interface %s", ifName)
		d.log.Error(err)
		return err
	}

	err := d.ifHandler.UnsetUnnumberedIP(ifMeta.SwIfIndex)
	if err != nil {
		d.log.Error(err)
	}

	return err
}

// Dependencies lists dependencies for an unnumbered VPP interface.
func (d *UnnumberedIfDescriptor) Dependencies(key string, unIntf *interfaces.Interface_Unnumbered) (deps []kvs.Dependency) {
	// link between unnumbered interface and the referenced interface with IP address
	// - satisfied as along as the referenced interface is configured and has at least
	//   one IP address assigned
	deps = []kvs.Dependency{
		{
			Label: unnumberedInterfaceHasIPDep,
			Key:   interfaces.InterfaceWithIPKey(unIntf.InterfaceWithIp),
		},
	}

	// interface has to be assigned to VRF before setting as unnumbered
	iface, _ := interfaces.ParseNameFromUnnumberedKey(key)
	deps = append(deps, kvs.Dependency{
		Label: interfaceInVrfDep,
		AnyOf: kvs.AnyOfDependency{
			KeyPrefixes: []string{interfaces.InterfaceVrfKeyPrefix(iface)},
		},
	})
	return
}
