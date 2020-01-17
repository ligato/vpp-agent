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
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	ifdescriptor "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/mockcalls"
	l2 "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model"
	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

const (
	// BridgeDomainDescriptorName is the name of the descriptor for mock bridge domains.
	BridgeDomainDescriptorName = "mock-bridge-domain"
)

// Example of some validation errors:
var (
	// ErrBridgeDomainWithoutName is returned when mock bridge domain configuration
	// has undefined Name attribute.
	ErrBridgeDomainWithoutName = errors.New("mock bridge domain defined without logical name")

	// ErrBridgeDomainWithMultipleBVI is returned when mock bridge domain is defined with
	// multiple BVI interfaces.
	ErrBridgeDomainWithMultipleBVI = errors.New("mock bridge domain defined with mutliple BVIs")
)

// BridgeDomainDescriptor teaches KVScheduler how to configure bridge domains
// in the mock SB.
type BridgeDomainDescriptor struct {
	// dependencies
	log       logging.Logger
	bdHandler mockcalls.MockBDAPI
}

// NewBridgeDomainDescriptor creates a new instance of the BridgeDomain descriptor.
func NewBridgeDomainDescriptor(bdHandler mockcalls.MockBDAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	// descriptors are supposed to be stateless and this principle is not broken
	// here - we only need to keep context consisting of references to logger
	// and the BD handler for mock SB, to be used inside the CRUD methods.
	descrCtx := &BridgeDomainDescriptor{
		bdHandler: bdHandler,
		log:       log.NewLogger("mock-bd-descriptor"),
	}

	// use adapter to convert typed descriptor into generic descriptor API
	typedDescr := &adapter.BridgeDomainDescriptor{
		Name:               BridgeDomainDescriptorName,
		NBKeyPrefix:        l2.ModelBridgeDomain.KeyPrefix(),
		ValueTypeName:      l2.ModelBridgeDomain.ProtoName(),
		KeySelector:        l2.ModelBridgeDomain.IsKeyValid,
		KeyLabel:           l2.ModelBridgeDomain.StripKeyPrefix,
		ValueComparator:    descrCtx.EquivalentBridgeDomains,
		WithMetadata:       true,
		MetadataMapFactory: descrCtx.MetadataFactory,
		Validate:           descrCtx.Validate,
		Create:             descrCtx.Create,
		Delete:             descrCtx.Delete,

		// Note: no need for Update operation - the interfaces are derived out
		//       and updated as separate key-value pairs, whereas a change in
		//       the Name actually results in a different key and creates a new
		//       BD altogether.

		Retrieve:      descrCtx.Retrieve,
		DerivedValues: descrCtx.DerivedValues,

		// Retrieve interfaces first to have the index with interface metadata
		// up-to-date when Retrieve for bridge domains is called, which then uses
		// the index to translate interface names to the corresponding integer
		// handles used in the mock SB.
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewBridgeDomainDescriptor(typedDescr)
}

// EquivalentBridgeDomains always returns true - this may seems strange, but
// two revisions of the same BD have the same key, therefore they must equal
// in Name, which is included in the key. The interfaces may differ, but
// they are derived out and updated as separate key-value pairs.
func (d *BridgeDomainDescriptor) EquivalentBridgeDomains(key string, oldBD, newBD *l2.BridgeDomain) bool {
	return true
}

// MetadataFactory is a factory for index-map customized for mock bridge domains.
func (d *BridgeDomainDescriptor) MetadataFactory() idxmap.NamedMappingRW {
	return idxvpp.NewNameToIndex(d.log, "mock-bd-index", nil)
}

// Validate validates mock bridge domain configuration.
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
	sbBDHandle, err := d.bdHandler.CreateBridgeDomain(bd.Name)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	// fill the metadata and return
	metadata = &idxvpp.OnlyIndex{
		Index: sbBDHandle,
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

// Retrieve returns all configured mock bridge domains.
func (d *BridgeDomainDescriptor) Retrieve(correlate []adapter.BridgeDomainKVWithMetadata) (retrieved []adapter.BridgeDomainKVWithMetadata, err error) {
	bds, err := d.bdHandler.DumpBridgeDomains()
	if err != nil {
		return nil, err
	}

	for sbBDHandle, bd := range bds {
		retrieved = append(retrieved, adapter.BridgeDomainKVWithMetadata{
			Key:      models.Key(bd),
			Value:    bd,
			Metadata: &idxvpp.OnlyIndex{Index: sbBDHandle},
			Origin:   kvs.FromNB, // not considering OBTAINED BDs in our simplified example
		})
	}
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
