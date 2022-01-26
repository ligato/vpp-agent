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
	"math"
	"net"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

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

	// The minimum value in milliseconds that can be used as interval.
	centisecondInMilliseconds = 10
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
		ValueComparator:      ctx.EquivalentVRRPs,
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

	if len(vrrp.IpAddresses) > math.MaxUint8 || len(vrrp.IpAddresses) == 0 {
		return kvs.NewInvalidValueError(ErrInvalidAddrNum, "ip_addresses")
	}

	if vrrp.GetVrId() > math.MaxUint8 || vrrp.GetVrId() == 0 {
		return kvs.NewInvalidValueError(ErrInvalidVrID, "vr_id")
	}

	if vrrp.GetPriority() > math.MaxUint8 || vrrp.GetPriority() == 0 {
		return kvs.NewInvalidValueError(ErrInvalidPriority, "priority")
	}

	if vrrp.GetInterval() > math.MaxUint16 || vrrp.GetInterval() < centisecondInMilliseconds {
		return kvs.NewInvalidValueError(ErrInvalidInterval, "interval")
	}

	var ip net.IP
	var isIpv6 bool
	for idx, addr := range vrrp.IpAddresses {
		ip = net.ParseIP(addr)
		if ip == nil {
			return kvs.NewInvalidValueError(ErrInvalidVrrpIP, "ip_addresses")
		}

		if idx == 0 && ip.To4() == nil {
			isIpv6 = true
			continue
		}

		if ip.To4() == nil && !isIpv6 || ip.To4() != nil && isIpv6 {
			return kvs.NewInvalidValueError(ErrInvalidIPVer, "ip_addresses")
		}
	}
	return nil
}

// Create adds VPP VRRP entry.
func (d *VrrpDescriptor) Create(key string, vrrp *l3.VRRPEntry) (interface{}, error) {
	if err := d.vrrpHandler.VppAddVrrp(vrrp); err != nil {
		if errors.Is(vppcalls.ErrVRRPUnsupported, err) {
			d.log.Debugf("Unsupported action: %v", err)
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
				d.log.Debugf("Unsupported action: %v", err)
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

	return !allFieldsWhithoutEnabledEquals(oldVRRPEntry, newVRRPEntry)
}

// Update updates VPP VRRP entry.
func (d *VrrpDescriptor) Update(_ string, oldVRRPEntry, newVRRPEntry *l3.VRRPEntry, _ interface{}) (
	_ interface{}, err error) {

	if newVRRPEntry.Enabled {
		err = d.vrrpHandler.VppStartVrrp(newVRRPEntry)
	} else {
		err = d.vrrpHandler.VppStopVrrp(newVRRPEntry)
	}
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
			Key:    l3.VrrpEntryKey(entry.Vrrp.Interface, entry.Vrrp.VrId),
			Value:  entry.Vrrp,
			Origin: kvs.FromNB,
		})
	}
	return retrieved, nil
}

// EquivalentVRRPs is a comparison function for l3.VRRPEntry.
func (d *VrrpDescriptor) EquivalentVRRPs(_ string, oldVRRPEntry, newVRRPEntry *l3.VRRPEntry) bool {
	if proto.Equal(oldVRRPEntry, newVRRPEntry) {
		return true
	}
	if oldVRRPEntry.Enabled != newVRRPEntry.Enabled {
		return false
	}
	return allFieldsWhithoutEnabledEquals(oldVRRPEntry, newVRRPEntry)
}

// allFieldsWhithoutEnabledEquals returns true if all entrys' fields are equal,
// without checking the Enabled field.
func allFieldsWhithoutEnabledEquals(entry1, entry2 *l3.VRRPEntry) bool {
	if entry1.Interface != entry2.Interface ||
		!intervalEquals(entry1.Interval, entry2.Interval) ||
		entry1.Priority != entry2.Priority ||
		entry1.VrId != entry2.VrId ||
		entry1.Accept != entry2.Accept ||
		entry1.Preempt != entry2.Preempt ||
		entry1.Unicast != entry2.Unicast ||
		len(entry1.IpAddresses) != len(entry2.IpAddresses) {
		return false
	}

	for i := 0; i < len(entry1.IpAddresses); i++ {
		if entry1.IpAddresses[i] != entry2.IpAddresses[i] {
			return false
		}
	}
	return true
}

// intervalEquals returns true if i1 and i2 are equal in centisonds.
func intervalEquals(i1, i2 uint32) bool {
	if i1/centisecondInMilliseconds == i2/centisecondInMilliseconds {
		return true
	}
	return false
}
