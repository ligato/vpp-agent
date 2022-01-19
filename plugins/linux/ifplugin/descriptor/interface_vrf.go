// Copyright (c) 2020 Pantheon.tech
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
	"fmt"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/protobuf/proto"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/ifaceidx"
	iflinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
)

const (
	// InterfaceVrfDescriptorName is the name of the descriptor for assigning Linux interfaces into a VRF.
	InterfaceVrfDescriptorName = "linux-interface-vrf"

	// dependency labels
	vrfDeviceDep   = "vrf-device-is-created"
	externalVrfDep = "inserted-into-vrf-externally"
)

// InterfaceVrfDescriptor (un)assigns Linux interface to/from VRF.
type InterfaceVrfDescriptor struct {
	log       logging.Logger
	ifHandler iflinuxcalls.NetlinkAPI
	nsPlugin  nsplugin.API
	intfIndex ifaceidx.LinuxIfMetadataIndex
}

// NewInterfaceVrfDescriptor creates a new instance of InterfaceVrfDescriptor.
func NewInterfaceVrfDescriptor(nsPlugin nsplugin.API,
	ifHandler iflinuxcalls.NetlinkAPI, log logging.PluginLogger) (descr *kvs.KVDescriptor, ctx *InterfaceVrfDescriptor) {

	ctx = &InterfaceVrfDescriptor{
		ifHandler: ifHandler,
		nsPlugin:  nsPlugin,
		log:       log.NewLogger("interface-vrf-descriptor"),
	}
	typedDescr := &adapter.InterfaceVrfDescriptor{
		Name:        InterfaceVrfDescriptorName,
		KeySelector: ctx.IsInterfaceVrfKey,
		ValueComparator: func(_ string, _, _ *interfaces.Interface) bool {
			// compare VRF assignments based on keys, not values that contain also other interface attributes
			// needed by the descriptor
			// FIXME: we can get rid of this hack once we add Context to descriptor methods
			return true
		},
		Create:       ctx.Create,
		Delete:       ctx.Delete,
		Dependencies: ctx.Dependencies,
	}
	descr = adapter.NewInterfaceVrfDescriptor(typedDescr)
	return
}

// SetInterfaceIndex should be used to provide interface index immediately after
// the descriptor registration.
func (d *InterfaceVrfDescriptor) SetInterfaceIndex(intfIndex ifaceidx.LinuxIfMetadataIndex) {
	d.intfIndex = intfIndex
}

// IsInterfaceVrfKey returns true if the key represents assignment of a Linux interface into a VRF.
func (d *InterfaceVrfDescriptor) IsInterfaceVrfKey(key string) bool {
	_, _, _, isVrfKey := interfaces.ParseInterfaceVrfKey(key)
	return isVrfKey
}

// Validate validates derived key.
func (d *InterfaceVrfDescriptor) Validate(key string, emptyVal proto.Message) (err error) {
	_, _, invalidKey, _ := interfaces.ParseInterfaceVrfKey(key)
	if invalidKey {
		return errors.New("invalid key")
	}
	return nil
}

// Create puts interface into a VRF.
func (d *InterfaceVrfDescriptor) Create(key string, iface *interfaces.Interface) (metadata interface{}, err error) {
	ifaceName, vrf, _, _ := interfaces.ParseInterfaceVrfKey(key)
	ifMeta, found := d.intfIndex.LookupByName(ifaceName)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return nil, err
	}
	vrfMeta, found := d.intfIndex.LookupByName(vrf)
	if !found {
		err = errors.Errorf("failed to find VRF device %s", vrf)
		d.log.Error(err)
		return nil, err
	}

	if iface.Type == interfaces.Interface_EXISTING {
		// Interface is managed externally, including its assignment into the VRF.
		// While dependencies allow us to require that the interface and the VRF both exist
		// and that the interface is inside *some* VRF, it is not possible to express requirement
		// that the actual VRF is the same as the desired one. Therefore we check the condition here
		// and return error if it is not the case, thus preventing items depending on this
		// from being created. Once the interface is re-assigned to the proper VRF, this kv will be
		// re-created with success.
		ifaceLink, err := d.ifHandler.GetLinkByIndex(ifMeta.LinuxIfIndex)
		if err != nil {
			err = fmt.Errorf("failed to obtain interface %s link: %w", iface, err)
			d.log.Error(err)
			return nil, err
		}
		if ifaceLink.Attrs().MasterIndex != vrfMeta.LinuxIfIndex {
			err = fmt.Errorf("existing interface %s is not inside VRF %s", iface, vrf)
			d.log.Error(err)
			return nil, err
		}
		return nil, nil
	}

	// switch to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}
	defer revert()

	err = d.ifHandler.PutInterfaceIntoVRF(ifMeta.HostIfName, vrfMeta.HostIfName)
	if err != nil {
		err = errors.WithMessagef(err, "failed to put interface '%s' into VRF '%s'",
			ifMeta.HostIfName, vrfMeta.HostIfName)
	}
	return nil, err
}

// Delete removes interface from VRF.
func (d *InterfaceVrfDescriptor) Delete(key string, iface *interfaces.Interface, metadata interface{}) (err error) {
	ifaceName, vrf, _, _ := interfaces.ParseInterfaceVrfKey(key)
	if iface.Type == interfaces.Interface_EXISTING {
		// interface is managed externally, nothing to do here
		return nil
	}

	ifMeta, found := d.intfIndex.LookupByName(ifaceName)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface)
		d.log.Error(err)
		return err
	}
	vrfMeta, found := d.intfIndex.LookupByName(vrf)
	if !found {
		err = errors.Errorf("failed to find VRF device %s", vrf)
		d.log.Error(err)
		return err
	}

	// switch to the namespace with the interface
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, ifMeta.Namespace)
	if err != nil {
		if _, ok := err.(*nsplugin.UnavailableMicroserviceErr); ok {
			// Assume that the delete was called by scheduler because the namespace
			// was removed. Do not return error in this case.
			d.log.Debugf("Interface %s assumed to be unassigned from VRF %s, required namespace %+v does not exist",
				iface, vrf, ifMeta.Namespace)
			return nil
		}
		d.log.Error(err)
		return err
	}
	defer revert()

	err = d.ifHandler.RemoveInterfaceFromVRF(ifMeta.HostIfName, vrfMeta.HostIfName)
	if err != nil {
		err = errors.WithMessagef(err, "failed to remove interface '%s' from VRF '%s'",
			ifMeta.HostIfName, vrfMeta.HostIfName)
	}
	return err
}

// Dependencies lists the VRF device as the only dependency.
func (d *InterfaceVrfDescriptor) Dependencies(key string, iface *interfaces.Interface) (deps []kvs.Dependency) {
	_, vrf, _, _ := interfaces.ParseInterfaceVrfKey(key)
	if vrf != "" {
		deps = append(deps, kvs.Dependency{
			Label: vrfDeviceDep,
			Key:   interfaces.InterfaceKey(vrf),
		})
	}
	if iface.Type == interfaces.Interface_EXISTING {
		// Interface is added into the VRF externally.
		// KV dependencies do not allow us to fully express this dependency - we can express requirement
		// that the interface is inside *some* VRF, but it not may be the desired one. This is because
		// the VRF host name is not yet known at this point and therefore it is not possible to build the
		// key to depend on.
		// Verification that the desired and actual VRFs are the same is therefore done in the Create method
		// and error is returned if it is not the case.
		deps = append(deps, kvs.Dependency{
			Label: externalVrfDep,
			AnyOf: kvs.AnyOfDependency{
				KeyPrefixes: []string{
					interfaces.InterfaceHostNameWithVrfKey(getHostIfName(iface), ""),
				},
			},
		})
	}
	return deps
}
