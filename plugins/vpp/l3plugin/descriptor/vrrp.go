//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package descriptor

import (
	"net"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// VrrpDescriptorName is the name of the descriptor.
	VrrpDescriptorName = "vrrp"

	// Dependency labels.
	vrrpEntryInterfaceDep    = "interface-exists"
	vrrpDescriptorLoggerName = "vrrp-descriptor"

	maxUint8  = 255
	maxUint16 = 65535
)

// A list of validation errors.
var (
	ErrMissingInterface = errors.New("missing interface")
	ErrInvalidAddrNum   = errors.New("addrs quantity should be > 0 && <= 255")
	ErrInvalidVrID      = errors.New("vr_id should be > 0 && <= 255")
	ErrInvalidPriority  = errors.New("priority should be > 0 && <= 255")
	ErrInvalidInterval  = errors.New("interval should be > 0 && <= 65535")
	ErrInvalidVrrpIP    = errors.New("invalid IP address")
	ErrInvalidIPVer     = errors.New("ipv6_flag does not correspond to IP version of the provided address")
	ErrInvalidInterface = errors.New("interface does not exist")
)

// VrrpDescriptor teaches KVScheduler how to configure VPP VRRPs.
type VrrpDescriptor struct {
	log         logging.Logger
	vrrpHandler vppcalls.VrrpVppAPI
}

// NewVrrpDescriptor creates a new instance of the VrrpDescriptor.
func NewVrrpDescriptor(vrrpHandler vppcalls.VrrpVppAPI,
	log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &VrrpDescriptor{
		log:         log.NewLogger(vrrpDescriptorLoggerName),
		vrrpHandler: vrrpHandler,
	}

	typedDescr := &adapter.VRRPEntryDescriptor{
		Name:                 VrrpDescriptorName,
		NBKeyPrefix:          l3.ModelVRRPEntry.KeyPrefix(),
		ValueTypeName:        l3.ModelVRRPEntry.ProtoName(),
		KeySelector:          l3.ModelVRRPEntry.IsKeyValid,
		KeyLabel:             l3.ModelVRRPEntry.StripKeyPrefix,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Update:               ctx.Update,
		UpdateWithRecreate:   ctx.UpdateWithRecreate,
		Validate:             ctx.Validate,
		Retrieve:             ctx.Retrieve,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}

	return adapter.NewVRRPEntryDescriptor(typedDescr)
}

// Validate returns error if given VRRP is not valid.
func (d *VrrpDescriptor) Validate(key string, vrrp *l3.VRRPEntry) error {
	if vrrp.Interface == "" {
		return kvs.NewInvalidValueError(ErrMissingInterface, "interface")
	}

	if len(vrrp.Addrs) > maxUint8 || len(vrrp.Addrs) == 0 {
		return kvs.NewInvalidValueError(ErrInvalidAddrNum, "addrs")
	}

	if vrrp.GetVrId() > maxUint8 || vrrp.GetVrId() == 0 {
		return kvs.NewInvalidValueError(ErrInvalidVrID, "vr_id")
	}

	if vrrp.GetPriority() > maxUint8 || vrrp.GetPriority() == 0 {
		return kvs.NewInvalidValueError(ErrInvalidPriority, "priority")
	}

	if vrrp.GetInterval() > maxUint16 || vrrp.GetInterval() == 0 {
		return kvs.NewInvalidValueError(ErrInvalidInterval, "interval")
	}

	var ip net.IP
	for _, addr := range vrrp.Addrs {
		ip = net.ParseIP(addr)
		if ip == nil {
			return kvs.NewInvalidValueError(ErrInvalidVrrpIP, "addr")
		}

		if ip.To4() == nil && !vrrp.Ipv6Flag || ip.To4() != nil && vrrp.Ipv6Flag {
			return kvs.NewInvalidValueError(ErrInvalidIPVer, "addr")
		}
	}
	return nil
}

// Create adds VPP VRRP entry.
func (d *VrrpDescriptor) Create(key string, vrrp *l3.VRRPEntry) (interface{}, error) {
	if err := d.vrrpHandler.VppAddVrrp(vrrp); err != nil {
		if errors.Is(vppcalls.ErrVRRPUnsupported, err) {
			d.log.Debugf("Unsupported action: ", err)
		}
		return nil, err
	}

	if vrrp.Enabled {
		if err := d.vrrpHandler.VppStartVrrp(vrrp); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Delete removes VPP VRRP entry.
func (d *VrrpDescriptor) Delete(key string, vrrp *l3.VRRPEntry, metadata interface{}) error {
	if vrrp.Enabled {
		if err := d.vrrpHandler.VppStopVrrp(vrrp); err != nil {
			if errors.Is(vppcalls.ErrVRRPUnsupported, err) {
				d.log.Debugf("Unsupported action: ", err)
			}
			return err
		}
	}
	if err := d.vrrpHandler.VppDelVrrp(vrrp); err != nil {
		return err
	}
	return nil
}

// UpdateWithRecreate returns true if a VRRP update needs to be performed via re-crate.
func (d *VrrpDescriptor) UpdateWithRecreate(_ string, oldVRRPEntry, newVRRPEntry *l3.VRRPEntry, _ interface{}) bool {
	if oldVRRPEntry.Enabled == newVRRPEntry.Enabled {
		// Something has changed, but it is not about enabled/disabled.
		return true
	}

	if oldVRRPEntry.Interface == newVRRPEntry.Interface &&
		oldVRRPEntry.Interval == newVRRPEntry.Interval &&
		oldVRRPEntry.Priority == newVRRPEntry.Priority &&
		oldVRRPEntry.VrId == newVRRPEntry.VrId &&
		oldVRRPEntry.Ipv6Flag == newVRRPEntry.Ipv6Flag &&
		oldVRRPEntry.AcceptFlag == newVRRPEntry.AcceptFlag &&
		oldVRRPEntry.PreemtpFlag == newVRRPEntry.PreemtpFlag &&
		oldVRRPEntry.UnicastFlag == newVRRPEntry.UnicastFlag &&
		len(oldVRRPEntry.Addrs) == len(newVRRPEntry.Addrs) {
		return false
	}

	for i := 0; i < len(oldVRRPEntry.Addrs); i++ {
		if oldVRRPEntry.Addrs[i] != newVRRPEntry.Addrs[i] {
			return true
		}
	}

	return true // Something changed except VRRP Enabled = recreate
}

// Update updates VPP VRRP entry.
func (d *VrrpDescriptor) Update(_ string, oldVRRPEntry, newVRRPEntry *l3.VRRPEntry, _ interface{}) (
	_ interface{}, err error) {

	if newVRRPEntry.Enabled {
		err = d.vrrpHandler.VppStartVrrp(newVRRPEntry)
	} else {
		err = d.vrrpHandler.VppStopVrrp(newVRRPEntry)
	}
	d.log.Debugf("Unsupported action: ", err)

	return nil, err
}

// Dependencies lists dependencies for a VPP VRRP entry.
func (d *VrrpDescriptor) Dependencies(key string, vrrp *l3.VRRPEntry) (deps []kvs.Dependency) {
	// the interface must exist
	if vrrp.Interface != "" {
		deps = append(deps, kvs.Dependency{
			Label: vrrpEntryInterfaceDep,
			Key:   interfaces.InterfaceKey(vrrp.Interface),
		})
	}
	return deps
}

// Retrieve returns all configured VPP VRRP entries.
func (d *VrrpDescriptor) Retrieve(correlate []adapter.VRRPEntryKVWithMetadata) (
	retrieved []adapter.VRRPEntryKVWithMetadata, err error,
) {
	entries, err := d.vrrpHandler.DumpVrrpEntries()

	for _, entry := range entries {

		retrieved = append(retrieved, adapter.VRRPEntryKVWithMetadata{
			Key:    l3.VrrpEntryKey(entry.Interface, entry.VrId),
			Value:  entry,
			Origin: kvs.UnknownOrigin,
		})
	}

	return retrieved, nil
}

// IsRetriableFailure returns false if error is one of errors
// defined at the top of this file as non-retriable.
func (d *VrrpDescriptor) IsRetriableFailure(err error) bool {
	if errors.Is(err, vppcalls.ErrVRRPUnsupported) {
		return false
	}
	return true
}
