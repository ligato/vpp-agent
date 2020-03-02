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
	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	vppIfDescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
)

const (
	// IPSecSPDDescriptorName is the name of the descriptor for VPP IPSec SPD
	// configurations.
	IPSecSPDDescriptorName = "vpp-ipsec-spd"
)

// A list of non-retriable errors:
var (
	// ErrIPSecSPDWithoutIndex is returned when VPP security policy database
	// configuration was defined without index.
	ErrIPSecSPDWithoutIndex = errors.New("VPP IPSec security policy database defined without index")

	// ErrIPSecSPDInvalidIndex is returned when VPP security policy database
	// configuration was defined with non-numerical index.
	ErrIPSecSPDInvalidIndex = errors.New("VPP IPSec security policy database defined with invalid index")

	// ErrSPDWithoutSA is returned when VPP security policy entry has undefined
	// security association attribute.
	ErrSPDWithoutSA = errors.New("VPP SPD policy entry defined without security association name")
)

// IPSecSPDDescriptor teaches KVScheduler how to configure IPSec SPD in VPP.
type IPSecSPDDescriptor struct {
	// dependencies
	log          logging.Logger
	ipSecHandler vppcalls.IPSecVppAPI

	// runtime
	spdIDSeq uint32
}

// NewIPSecSPDDescriptor creates a new instance of the IPSec SPD descriptor.
func NewIPSecSPDDescriptor(ipSecHandler vppcalls.IPSecVppAPI, log logging.PluginLogger) *IPSecSPDDescriptor {
	return &IPSecSPDDescriptor{
		ipSecHandler: ipSecHandler,
		log:          log.NewLogger("ipsec-spd-descriptor"),
		spdIDSeq:     1,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *IPSecSPDDescriptor) GetDescriptor() *adapter.SPDDescriptor {
	return &adapter.SPDDescriptor{
		Name:                 IPSecSPDDescriptorName,
		NBKeyPrefix:          ipsec.ModelSecurityPolicyDatabase.KeyPrefix(),
		ValueTypeName:        ipsec.ModelSecurityPolicyDatabase.ProtoName(),
		KeySelector:          ipsec.ModelSecurityPolicyDatabase.IsKeyValid,
		KeyLabel:             ipsec.ModelSecurityPolicyDatabase.StripKeyPrefix,
		ValueComparator:      d.EquivalentIPSecSPDs,
		WithMetadata:         true,
		MetadataMapFactory:   d.MetadataFactory,
		Create:               d.Create,
		Delete:               d.Delete,
		Retrieve:             d.Retrieve,
		DerivedValues:        d.DerivedValues,
		RetrieveDependencies: []string{vppIfDescriptor.InterfaceDescriptorName},
	}
}

// EquivalentIPSecSPDs is case-insensitive comparison function for
// ipsec.SecurityPolicyDatabase, also ignoring the order of assigned
// interfaces and/or policy entries.
func (d *IPSecSPDDescriptor) EquivalentIPSecSPDs(key string, oldSPD, newSPD *ipsec.SecurityPolicyDatabase) bool {
	// SPD interfaces
	obsoleteIfs, newIfs := calculateInterfacesDiff(oldSPD.GetInterfaces(), newSPD.GetInterfaces())
	if len(obsoleteIfs) != 0 || len(newIfs) != 0 {
		return false
	}

	// SPD policy entries
	obsoletePes, newPes := calculatePolicyEntriesDiff(oldSPD.GetPolicyEntries(), newSPD.GetPolicyEntries())
	return len(obsoletePes) == 0 && len(newPes) == 0
}

// MetadataFactory is a factory for index-map customized for VPP security policy databases.
func (d *IPSecSPDDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return idxvpp.NewNameToIndex(d.log, "vpp-spd-index", nil)
}

// Create adds a new IPSec security policy database.
func (d *IPSecSPDDescriptor) Create(key string, spd *ipsec.SecurityPolicyDatabase) (metadata *idxvpp.OnlyIndex, err error) {
	// allocate new SPD ID
	spdIdx := d.spdIDSeq
	d.spdIDSeq++

	// create a new SPD with index
	err = d.ipSecHandler.AddSPD(spdIdx)
	if err != nil {
		// Note: d.spdIDSeq will be refreshed by Dump
		d.log.Error(err)
		return nil, err
	}

	// fill the metadata
	metadata = &idxvpp.OnlyIndex{
		Index: spdIdx,
	}
	return metadata, nil
}

// Delete removes VPP IPSec security policy database.
func (d *IPSecSPDDescriptor) Delete(key string, spd *ipsec.SecurityPolicyDatabase, metadata *idxvpp.OnlyIndex) error {
	err := d.ipSecHandler.DeleteSPD(metadata.GetIndex())
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// Retrieve returns all configured VPP security policy databases.
func (d *IPSecSPDDescriptor) Retrieve(correlate []adapter.SPDKVWithMetadata) (dump []adapter.SPDKVWithMetadata, err error) {
	// dump security policy associations
	spds, err := d.ipSecHandler.DumpIPSecSPD()
	if err != nil {
		d.log.Error(err)
		return dump, err
	}
	for _, spd := range spds {
		dump = append(dump, adapter.SPDKVWithMetadata{
			Key:      ipsec.SPDKey(spd.Spd.Index),
			Value:    spd.Spd,
			Metadata: &idxvpp.OnlyIndex{Index: spd.Spd.Index},
			Origin:   kvs.FromNB,
		})
	}

	return dump, nil
}

// DerivedValues derives ipsec.SecurityPolicyDatabase_Interface for every interface assigned
// assigned to the SPD and ipsec.SecurityPolicyDatabase_PolicyEntry for every policy entry
// assigned to the SPD
func (d *IPSecSPDDescriptor) DerivedValues(key string, spd *ipsec.SecurityPolicyDatabase) (derValues []kvs.KeyValuePair) {
	// SPD interfaces
	for _, spdIface := range spd.Interfaces {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   ipsec.SPDInterfaceKey(spd.Index, spdIface.Name),
			Value: spdIface,
		})
	}

	// SPD policy entries
	for _, spdPe := range spd.PolicyEntries {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   ipsec.SPDPolicyKey(spd.Index, spdPe.SaIndex),
			Value: spdPe,
		})
	}

	return derValues
}

// calculateInterfacesDiff compares two sets of SPD interfaces entries.
func calculateInterfacesDiff(oldIfs, newIfs []*ipsec.SecurityPolicyDatabase_Interface) (toRemove, toAdd []*ipsec.SecurityPolicyDatabase_Interface) {
	// Resolve interfaces to add
	for _, newIf := range newIfs {
		var exists bool
		for _, oldIf := range oldIfs {
			if newIf.Name == oldIf.Name {
				exists = true
				break
			}
		}
		if !exists {
			toAdd = append(toAdd, newIf)
		}
	}
	// Resolve interfaces to remove
	for _, oldIf := range oldIfs {
		var exists bool
		for _, newIf := range newIfs {
			if oldIf.Name == newIf.Name {
				exists = true
				break
			}
		}
		if !exists {
			toRemove = append(toRemove, oldIf)
		}
	}

	return toAdd, toRemove
}

// calculateInterfacesDiff compares two sets of SPD interfaces entries.
func calculatePolicyEntriesDiff(oldPes, newPes []*ipsec.SecurityPolicyDatabase_PolicyEntry) (toRemove, toAdd []*ipsec.SecurityPolicyDatabase_PolicyEntry) {
	// Resolve interfaces to add
	for _, newPe := range newPes {
		var exists bool
		for _, oldPe := range oldPes {
			if equalPolicyEntries(newPe, oldPe) {
				exists = true
				break
			}
		}
		if !exists {
			toAdd = append(toAdd, newPe)
		}
	}
	// Resolve interfaces to remove
	for _, oldPe := range oldPes {
		var exists bool
		for _, newPe := range newPes {
			if equalPolicyEntries(newPe, oldPe) {
				exists = true
				break
			}
		}
		if !exists {
			toRemove = append(toRemove, oldPe)
		}
	}

	return toAdd, toRemove
}

// equalPolicyEntries compares two SPD policy entries for equality.
func equalPolicyEntries(pe1, pe2 *ipsec.SecurityPolicyDatabase_PolicyEntry) bool {
	if !equalPolicyEntriesBase(pe1, pe2) {
		return false
	}

	// compare remote start addresses
	if !equalPolicyEntriesIPAddress(pe1.RemoteAddrStart, pe2.RemoteAddrStop) {
		return false
	}

	// compare remote stop addresses
	if !equalPolicyEntriesIPAddress(pe1.RemoteAddrStop, pe2.RemoteAddrStop) {
		return false
	}

	// compare local start addresses
	if !equalPolicyEntriesIPAddress(pe1.LocalAddrStart, pe2.LocalAddrStart) {
		return false
	}

	// compare local stop addresses
	if !equalPolicyEntriesIPAddress(pe1.LocalAddrStop, pe2.LocalAddrStop) {
		return false
	}

	return true
}

// equalPolicyEntriesBase compares base parameters of two policy entries (except IP addresses)
func equalPolicyEntriesBase(pe1, pe2 *ipsec.SecurityPolicyDatabase_PolicyEntry) bool {
	return pe1.Priority == pe2.Priority &&
		pe1.IsOutbound == pe2.IsOutbound &&
		pe1.Protocol == pe2.Protocol &&
		pe1.RemotePortStart == pe2.RemotePortStart &&
		pe1.RemotePortStop == pe2.RemotePortStop &&
		pe1.LocalPortStart == pe2.LocalPortStop &&
		pe1.Action == pe2.Action
}

// equalPolicyEntriesIPAddress compare two policy entries IP addresses
func equalPolicyEntriesIPAddress(peIP1, peIP2 string) bool {
	ip1 := net.ParseIP(peIP1)
	ip2 := net.ParseIP(peIP2)
	if ip1 == nil || ip2 == nil {
		// if parsing fails, compare as strings
		return strings.ToLower(peIP1) != strings.ToLower(peIP2)
	}
	return ip1.Equal(ip2)
}
