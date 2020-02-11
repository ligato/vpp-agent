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
	"strconv"

	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

const (
	// SPDInterfaceDescriptorName is the name of the descriptor for bindings between
	// VPP IPSec security policy database and interfaces.
	SPDInterfaceDescriptorName = "vpp-spd-interface"

	// dependency labels
	interfaceDep = "interface-exists"
)

// SPDInterfaceDescriptor teaches KVScheduler how to put interface into VPP
// security policy database
type SPDInterfaceDescriptor struct {
	// dependencies
	log          logging.Logger
	ipSecHandler vppcalls.IPSecVppAPI
}

// NewSPDInterfaceDescriptor creates a new instance of the SPDInterface descriptor.
func NewSPDInterfaceDescriptor(ipSecHandler vppcalls.IPSecVppAPI, log logging.PluginLogger) *SPDInterfaceDescriptor {
	return &SPDInterfaceDescriptor{
		log:          log.NewLogger("spd-interface-descriptor"),
		ipSecHandler: ipSecHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *SPDInterfaceDescriptor) GetDescriptor() *adapter.SPDInterfaceDescriptor {
	return &adapter.SPDInterfaceDescriptor{
		Name:          SPDInterfaceDescriptorName,
		KeySelector:   d.IsSPDInterfaceKey,
		ValueTypeName: proto.MessageName(&ipsec.SecurityPolicyDatabase{}),
		Create:        d.Create,
		Delete:        d.Delete,
		Dependencies:  d.Dependencies,
	}
}

// IsSPDInterfaceKey returns true if the key is identifying binding between
// VPP security policy database and interface.
func (d *SPDInterfaceDescriptor) IsSPDInterfaceKey(key string) bool {
	_, _, isSPDIfaceKey := ipsec.ParseSPDInterfaceKey(key)
	return isSPDIfaceKey
}

// Create puts interface into security policy database.
func (d *SPDInterfaceDescriptor) Create(key string, spdIf *ipsec.SecurityPolicyDatabase_Interface) (metadata interface{}, err error) {
	// get security policy database index
	spdIdx, _, isSPDIfKey := ipsec.ParseSPDInterfaceKey(key)
	if !isSPDIfKey {
		err = errors.Errorf("provided key is not a derived SPD <=> interface binding key %s", key)
		d.log.Error(err)
		return nil, err
	}

	// convert to numeric index
	spdID, err := strconv.Atoi(spdIdx)
	if err != nil {
		err = errors.Errorf("provided SPD index is not a valid value %s", spdIdx)
		d.log.Error(err)
		return nil, err
	}

	// put interface into the security policy database
	err = d.ipSecHandler.AddSPDInterface(uint32(spdID), spdIf)
	if err != nil {
		d.log.Error(err)
		return nil, err

	}
	return nil, nil
}

// Delete removes interface from security policy database.
func (d *SPDInterfaceDescriptor) Delete(key string, spdIf *ipsec.SecurityPolicyDatabase_Interface, metadata interface{}) (err error) {
	// get security policy database index
	spdIdx, _, isSPDIfKey := ipsec.ParseSPDInterfaceKey(key)
	if !isSPDIfKey {
		err = errors.Errorf("provided key is not a derived SPD <=> interface binding key %s", key)
		d.log.Error(err)
		return err
	}

	// convert to numeric index
	spdID, err := strconv.Atoi(spdIdx)
	if err != nil {
		err = errors.Errorf("provided SPD index is not a valid value %s", spdIdx)
		d.log.Error(err)
		return err
	}

	err = d.ipSecHandler.DeleteSPDInterface(uint32(spdID), spdIf)
	if err != nil {
		d.log.Error(err)
		return err

	}
	return nil
}

// Dependencies lists the interface as the only dependency for the binding.
func (d *SPDInterfaceDescriptor) Dependencies(key string, value *ipsec.SecurityPolicyDatabase_Interface) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: interfaceDep,
			Key:   interfaces.InterfaceKey(value.Name),
		},
	}
}
