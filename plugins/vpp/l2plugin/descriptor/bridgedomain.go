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
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	vpp_ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

const (
	// BridgeDomainDescriptorName is the name of the descriptor for VPP bridge domains.
	BridgeDomainDescriptorName = "vpp-bridge-domain"

	// prefix prepended to internal names of untagged bridge domains to construct
	// unique logical names
	untaggedBDPreffix = "UNTAGGED-"
)

// A list of non-retriable errors:
var (
	// ErrBridgeDomainWithoutName is returned when VPP bridge domain configuration
	// has undefined Name attribute.
	ErrBridgeDomainWithoutName = errors.New("VPP bridge domain defined without logical name")

	// ErrBridgeDomainWithMultipleBVI is returned when bridge domain is defined with
	// multiple BVI interfaces.
	ErrBridgeDomainWithMultipleBVI = errors.New("VPP bridge domain defined with mutliple BVIs")
)

// BridgeDomainDescriptor teaches KVScheduler how to configure VPP bridge domains.
type BridgeDomainDescriptor struct {
	// dependencies
	log       logging.Logger
	bdHandler vppcalls.BridgeDomainVppAPI

	// runtime
	bdIDSeq uint32
}

// NewBridgeDomainDescriptor creates a new instance of the BridgeDomain descriptor.
func NewBridgeDomainDescriptor(bdHandler vppcalls.BridgeDomainVppAPI, log logging.PluginLogger) *BridgeDomainDescriptor {

	return &BridgeDomainDescriptor{
		bdHandler: bdHandler,
		log:       log.NewLogger("bd-descriptor"),
		bdIDSeq:   1,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *BridgeDomainDescriptor) GetDescriptor() *adapter.BridgeDomainDescriptor {
	return &adapter.BridgeDomainDescriptor{
		Name:                 BridgeDomainDescriptorName,
		NBKeyPrefix:          l2.ModelBridgeDomain.KeyPrefix(),
		ValueTypeName:        l2.ModelBridgeDomain.ProtoName(),
		KeySelector:          l2.ModelBridgeDomain.IsKeyValid,
		KeyLabel:             l2.ModelBridgeDomain.StripKeyPrefix,
		ValueComparator:      d.EquivalentBridgeDomains,
		WithMetadata:         true,
		MetadataMapFactory:   d.MetadataFactory,
		Validate:             d.Validate,
		Create:               d.Create,
		Delete:               d.Delete,
		Update:               d.Update,
		UpdateWithRecreate:   d.UpdateWithRecreate,
		Retrieve:             d.Retrieve,
		DerivedValues:        d.DerivedValues,
		RetrieveDependencies: []string{vpp_ifdescriptor.InterfaceDescriptorName},
	}
}

// EquivalentBridgeDomains is case-insensitive comparison function for
// l2.BridgeDomain, also ignoring the order of assigned ARP termination entries.
func (d *BridgeDomainDescriptor) EquivalentBridgeDomains(key string, oldBD, newBD *l2.BridgeDomain) bool {
	// BD parameters
	if !equalBDParameters(oldBD, newBD) {
		return false
	}

	// ARP termination entries
	obsoleteARPs, newARPs := calculateARPDiff(oldBD.GetArpTerminationTable(), newBD.GetArpTerminationTable())
	return len(obsoleteARPs) == 0 && len(newARPs) == 0
}

// MetadataFactory is a factory for index-map customized for VPP bridge domains.
func (d *BridgeDomainDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return idxvpp.NewNameToIndex(d.log, "vpp-bd-index", nil)
}

// Validate validates VPP bridge domain configuration.
func (d *BridgeDomainDescriptor) Validate(key string, bd *l2.BridgeDomain) error {
	if bd.Name == "" {
		return kvs.NewInvalidValueError(ErrBridgeDomainWithoutName, "name")
	}

	// check that BD has defined at most one BVI
	var hasBVI bool
	for _, bdIface := range bd.Interfaces {
		if bdIface.BridgedVirtualInterface {
			if hasBVI {
				return kvs.NewInvalidValueError(ErrBridgeDomainWithMultipleBVI,
					"interfaces.bridged_virtual_interface")
			}
			hasBVI = true
		}
	}
	return nil
}

// Create adds new bridge domain.
func (d *BridgeDomainDescriptor) Create(key string, bd *l2.BridgeDomain) (metadata *idxvpp.OnlyIndex, err error) {
	// allocate new bridge domain ID
	bdIdx := d.bdIDSeq
	d.bdIDSeq++

	// create the bridge domain
	err = d.bdHandler.AddBridgeDomain(bdIdx, bd)
	if err != nil {
		// Note: d.bdIDSeq will be refreshed by Dump
		d.log.Error(err)
		return nil, err
	}

	// add ARP termination entries
	for _, arp := range bd.ArpTerminationTable {
		if err := d.bdHandler.AddArpTerminationTableEntry(bdIdx, arp.PhysAddress, arp.IpAddress); err != nil {
			d.log.Error(err)
			return nil, err
		}
	}

	// fill the metadata
	metadata = &idxvpp.OnlyIndex{
		Index: bdIdx,
	}
	return metadata, nil
}

// Delete removes VPP bridge domain.
func (d *BridgeDomainDescriptor) Delete(key string, bd *l2.BridgeDomain, metadata *idxvpp.OnlyIndex) error {
	err := d.bdHandler.DeleteBridgeDomain(metadata.GetIndex())
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// UpdateWithRecreate returns true if bridge domain base parameters are different.
func (d *BridgeDomainDescriptor) UpdateWithRecreate(key string, oldBD, newBD *l2.BridgeDomain, metadata *idxvpp.OnlyIndex) bool {
	return !equalBDParameters(oldBD, newBD)
}

// Update is able to change ARP termination entries.
func (d *BridgeDomainDescriptor) Update(key string, oldBD, newBD *l2.BridgeDomain, oldMetadata *idxvpp.OnlyIndex) (newMetadata *idxvpp.OnlyIndex, err error) {
	// update ARP termination entries
	bdIdx := oldMetadata.Index
	obsoleteARPs, newARPs := calculateARPDiff(oldBD.GetArpTerminationTable(), newBD.GetArpTerminationTable())
	for _, arp := range obsoleteARPs { // remove obsolete first to avoid collisions
		if err := d.bdHandler.RemoveArpTerminationTableEntry(bdIdx, arp.PhysAddress, arp.IpAddress); err != nil {
			d.log.Error(err)
			return oldMetadata, err
		}
	}
	for _, arp := range newARPs {
		if err := d.bdHandler.AddArpTerminationTableEntry(bdIdx, arp.PhysAddress, arp.IpAddress); err != nil {
			d.log.Error(err)
			return oldMetadata, err
		}
	}

	return oldMetadata, nil
}

// Retrieve returns all configured VPP bridge domains.
func (d *BridgeDomainDescriptor) Retrieve(correlate []adapter.BridgeDomainKVWithMetadata) (retrieved []adapter.BridgeDomainKVWithMetadata, err error) {
	// d.bdIDSeq will be refreshed
	var bdIDSeq uint32 = 1

	// sequence number for untagged interfaces
	var untaggedSeq int

	// dump bridge domains
	bridgeDomains, err := d.bdHandler.DumpBridgeDomains()
	if err != nil {
		d.log.Error(err)
		return retrieved, err
	}
	for _, bd := range bridgeDomains {
		// make sure that bdIDSeq is larger than any of the existing indexes
		if bd.Meta.BdID >= bdIDSeq {
			bdIDSeq = bd.Meta.BdID + 1
		}

		// handle untagged bridge domain - construct name that is unlikely to
		// collide with NB, thus the bridge domain will get removed by resync
		if bd.Bd.Name == "" {
			bd.Bd.Name = fmt.Sprintf("%s%d", untaggedBDPreffix, untaggedSeq)
			untaggedSeq++
		}

		retrieved = append(retrieved, adapter.BridgeDomainKVWithMetadata{
			Key:      l2.BridgeDomainKey(bd.Bd.Name),
			Value:    bd.Bd,
			Metadata: &idxvpp.OnlyIndex{Index: bd.Meta.BdID},
			Origin:   kvs.FromNB,
		})
	}

	// update d.bdIDSeq
	d.bdIDSeq = bdIDSeq

	return retrieved, nil
}

// DerivedValues derives l2.BridgeDomain_Interface for every interface assigned
// to the bridge domain.
func (d *BridgeDomainDescriptor) DerivedValues(key string, bd *l2.BridgeDomain) (derValues []kvs.KeyValuePair) {
	// BD interfaces
	for _, bdIface := range bd.Interfaces {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   l2.BDInterfaceKey(bd.Name, bdIface.Name),
			Value: bdIface,
		})
	}
	return derValues
}

// equalBDParameters compares all base bridge domain parameters for equality.
func equalBDParameters(bd1, bd2 *l2.BridgeDomain) bool {
	return bd1.ArpTermination == bd2.ArpTermination && bd1.Flood == bd2.Flood &&
		bd1.Forward == bd2.Forward && bd1.Learn == bd2.Learn && bd1.MacAge == bd2.MacAge &&
		bd1.UnknownUnicastFlood == bd2.UnknownUnicastFlood
}

// calculateARPDiff compares two sets of ARP termination entries.
func calculateARPDiff(oldARPs, newARPs []*l2.BridgeDomain_ArpTerminationEntry) (toRemove, toAdd []*l2.BridgeDomain_ArpTerminationEntry) {
	// Resolve ARPs to add
	for _, newARP := range newARPs {
		var exists bool
		for _, oldARP := range oldARPs {
			if equalTerminationARPs(oldARP, newARP) {
				exists = true
				break
			}
		}
		if !exists {
			toAdd = append(toAdd, newARP)
		}
	}
	// Resolve ARPs to remove
	for _, oldARP := range oldARPs {
		var exists bool
		for _, newARP := range newARPs {
			if equalTerminationARPs(oldARP, newARP) {
				exists = true
				break
			}
		}
		if !exists {
			toRemove = append(toRemove, oldARP)
		}
	}

	return toAdd, toRemove
}

// equalTerminationARPs compares two termination ARP entries for equality.
func equalTerminationARPs(arp1, arp2 *l2.BridgeDomain_ArpTerminationEntry) bool {
	// compare MAC addresses
	if strings.ToLower(arp1.PhysAddress) != strings.ToLower(arp2.PhysAddress) {
		return false
	}

	// compare IP addresses
	ip1 := net.ParseIP(arp1.IpAddress)
	ip2 := net.ParseIP(arp2.IpAddress)
	if ip1 == nil || ip2 == nil {
		// if parsing fails, compare as strings
		return strings.ToLower(arp1.IpAddress) == strings.ToLower(arp2.IpAddress)
	}
	return ip1.Equal(ip2)
}
