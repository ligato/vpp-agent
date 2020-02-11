//  Copyright (c) 2019 Cisco and/or its affiliates.
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
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	prototypes "github.com/golang/protobuf/ptypes/empty"

	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin/abfidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/aclidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/descriptor"
	ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
	acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const (
	// ABFDescriptorName is descriptor name
	ABFDescriptorName = "vpp-abf"

	// dependency labels
	aclDep = "acl-exists"
)

// A list of non-retriable errors:
var (
	// ErrABFWithoutACL is returned when ABF configuration does not contain associated access list.
	ErrABFWithoutACL = errors.New("ABF configuration defined without ACL")
)

// ABFDescriptor is descriptor for ABF
type ABFDescriptor struct {
	// dependencies
	log        logging.Logger
	abfHandler vppcalls.ABFVppAPI

	// runtime
	aclIndex aclidx.ACLMetadataIndex
}

// NewABFDescriptor is constructor for ABF descriptor and returns descriptor
// suitable for registration (via adapter) with the KVScheduler.
func NewABFDescriptor(
	abfHandler vppcalls.ABFVppAPI,
	aclIndex aclidx.ACLMetadataIndex,
	logger logging.PluginLogger,
) *api.KVDescriptor {
	ctx := &ABFDescriptor{
		log:        logger.NewLogger("abf-descriptor"),
		aclIndex:   aclIndex,
		abfHandler: abfHandler,
	}
	typedDescr := &adapter.ABFDescriptor{
		Name:          ABFDescriptorName,
		NBKeyPrefix:   abf.ModelABF.KeyPrefix(),
		ValueTypeName: abf.ModelABF.ProtoName(),
		KeySelector:   abf.ModelABF.IsKeyValid,
		KeyLabel:      abf.ModelABF.StripKeyPrefix,
		WithMetadata:  true,
		MetadataMapFactory: func() idxmap.NamedMappingRW {
			return abfidx.NewABFIndex(ctx.log, "vpp-abf-index")
		},
		ValueComparator:      ctx.EquivalentABFs,
		Validate:             ctx.Validate,
		Create:               ctx.Create,
		Delete:               ctx.Delete,
		Retrieve:             ctx.Retrieve,
		DerivedValues:        ctx.DerivedValues,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName, descriptor.ACLDescriptorName},
	}
	return adapter.NewABFDescriptor(typedDescr)
}

// EquivalentABFs compares related ACL name, list of attached interfaces and forwarding paths to
// specify ABS equality.
func (d *ABFDescriptor) EquivalentABFs(key string, oldABF, newABF *abf.ABF) bool {
	// check index and associated ACL
	if oldABF.AclName != newABF.AclName {
		return false
	}

	// compare attached interfaces
	if len(oldABF.AttachedInterfaces) != len(newABF.AttachedInterfaces) {
		return false
	}
	if !equivalentABFAttachedInterfaces(oldABF.AttachedInterfaces, newABF.AttachedInterfaces) {
		return false
	}

	// compare forwarding paths
	if len(oldABF.ForwardingPaths) != len(newABF.ForwardingPaths) {
		return false
	}
	return equivalentABFForwardingPaths(oldABF.ForwardingPaths, newABF.ForwardingPaths)
}

// Validate validates VPP ABF configuration.
func (d *ABFDescriptor) Validate(key string, abfData *abf.ABF) error {
	if abfData.AclName == "" {
		return api.NewInvalidValueError(ErrABFWithoutACL, "acl_name")
	}
	return nil
}

// Create validates ABF (mainly index), verifies ACL existence and configures ABF policy. Attached interfaces
// are put to metadata together with the ABF index to make it available for other ABF descriptors.
func (d *ABFDescriptor) Create(key string, abfData *abf.ABF) (*abfidx.ABFMetadata, error) {
	// get ACL index
	aclData, exists := d.aclIndex.LookupByName(abfData.AclName)
	if !exists {
		err := errors.Errorf("failed to obtain metadata for ACL %s", abfData.AclName)
		d.log.Error(err)
		return nil, err
	}

	// add new ABF policy
	if err := d.abfHandler.AddAbfPolicy(abfData.Index, aclData.Index, abfData.ForwardingPaths); err != nil {
		d.log.Error(err)
		return nil, err
	}

	// fill the metadata
	metadata := &abfidx.ABFMetadata{
		Index:    abfData.Index,
		Attached: abfData.AttachedInterfaces,
	}

	return metadata, nil
}

// Delete removes ABF policy
func (d *ABFDescriptor) Delete(key string, abfData *abf.ABF, metadata *abfidx.ABFMetadata) error {
	// ACL ID is not required
	return d.abfHandler.DeleteAbfPolicy(metadata.Index, abfData.ForwardingPaths)
}

// Retrieve returns ABF policies from the VPP.
func (d *ABFDescriptor) Retrieve(correlate []adapter.ABFKVWithMetadata) (abfs []adapter.ABFKVWithMetadata, err error) {
	// Retrieve VPP configuration.
	abfPolicies, err := d.abfHandler.DumpABFPolicy()
	if err != nil {
		return nil, errors.Errorf("failed to dump ABF policy: %v", err)
	}

	for _, abfPolicy := range abfPolicies {
		abfs = append(abfs, adapter.ABFKVWithMetadata{
			Key:   abf.Key(abfPolicy.ABF.Index),
			Value: abfPolicy.ABF,
			Metadata: &abfidx.ABFMetadata{
				Index:    abfPolicy.Meta.PolicyID,
				Attached: abfPolicy.ABF.AttachedInterfaces,
			},
			Origin: api.FromNB,
		})
	}

	return abfs, nil
}

// DerivedValues returns list of derived values for ABF.
func (d *ABFDescriptor) DerivedValues(key string, value *abf.ABF) (derived []api.KeyValuePair) {
	for _, attachedIf := range value.GetAttachedInterfaces() {
		derived = append(derived, api.KeyValuePair{
			Key:   abf.ToInterfaceKey(value.Index, attachedIf.InputInterface),
			Value: &prototypes.Empty{},
		})
	}
	return derived
}

// A list of ABF dependencies (ACL + forwarding path interfaces).
func (d *ABFDescriptor) Dependencies(key string, abfData *abf.ABF) (dependencies []api.Dependency) {
	// forwarding path interfaces
	for _, abfDataLabel := range abfData.ForwardingPaths {
		if abfDataLabel.InterfaceName != "" {
			dependencies = append(dependencies, api.Dependency{
				Label: interfaceDep,
				Key:   vpp_interfaces.InterfaceKey(abfDataLabel.InterfaceName),
			})
		}
	}
	// access list
	dependencies = append(dependencies, api.Dependency{
		Label: aclDep,
		Key:   acl.Key(abfData.AclName),
	})

	return dependencies
}

func equivalentABFAttachedInterfaces(oldIfs, newIfs []*abf.ABF_AttachedInterface) bool {
	for _, oldIf := range oldIfs {
		var found bool
		for _, newIf := range newIfs {
			if proto.Equal(oldIf, newIf) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equivalentABFForwardingPaths(oldPaths, newPaths []*abf.ABF_ForwardingPath) bool {
	for _, oldPath := range oldPaths {
		var found bool
		for _, newPath := range newPaths {
			if oldPath.InterfaceName == newPath.InterfaceName &&
				oldPath.NextHopIp == newPath.NextHopIp &&
				oldPath.Weight == newPath.Weight &&
				oldPath.Preference == newPath.Preference &&
				oldPath.Dvr == newPath.Dvr {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
