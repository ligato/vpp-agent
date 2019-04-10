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
	"strconv"

	"github.com/go-errors/errors"
	"github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
	abf "github.com/ligato/vpp-agent/api/models/vpp/abf"
	acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/abfidx"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/abfplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/aclplugin/aclidx"
	ifdescriptor "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/descriptor"
)

const (
	// ABFDescriptorName is descriptor name
	ABFDescriptorName = "vpp-abf"

	// dependency labels
	aclDep = "acl-exists"
)

// A list of non-retriable errors:
var (
	// ErrABFInvalidIndex is returned when ABF configuration is defined with invalid index (not a number).
	ErrABFInvalidIndex = errors.New("ABF configuration contains invalid index")

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

// NewABFDescriptor is constructor for ABF descriptor
func NewABFDescriptor(abfHandler vppcalls.ABFVppAPI, aclIndex aclidx.ACLMetadataIndex,
	logger logging.PluginLogger) *ABFDescriptor {
	return &ABFDescriptor{
		log:        logger.NewLogger("abf-descriptor"),
		aclIndex:   aclIndex,
		abfHandler: abfHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *ABFDescriptor) GetDescriptor() *adapter.ABFDescriptor {
	return &adapter.ABFDescriptor{
		Name:          ABFDescriptorName,
		NBKeyPrefix:   abf.ModelABF.KeyPrefix(),
		ValueTypeName: abf.ModelABF.ProtoName(),
		KeySelector:   abf.ModelABF.IsKeyValid,
		KeyLabel:      abf.ModelABF.StripKeyPrefix,
		WithMetadata:  true,
		MetadataMapFactory: func() idxmap.NamedMappingRW {
			return abfidx.NewABFIndex(d.log, "vpp-abf-index")
		},
		ValueComparator:      d.EquivalentABFs,
		Validate:             d.Validate,
		Create:               d.Create,
		Delete:               d.Delete,
		UpdateWithRecreate:   d.UpdateWithRecreate,
		Retrieve:             d.Retrieve,
		IsRetriableFailure:   d.IsRetriableFailure,
		DerivedValues:        d.DerivedValues,
		Dependencies:         d.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}
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
	if abfData.Index == "" {
		return api.NewInvalidValueError(ErrABFInvalidIndex, "index")
	}
	_, err := strconv.Atoi(abfData.Index)
	if err != nil {
		return api.NewInvalidValueError(ErrABFInvalidIndex, "index")
	}
	if abfData.AclName == "" {
		return api.NewInvalidValueError(ErrABFWithoutACL, "acl_name")
	}
	return nil
}

// Create validates ABF (mainly index), verifies ACL existence and configures ABF policy. Attached interfaces
// are put to metadata together with the ABF index to make it available for other ABF descriptors.
func (d *ABFDescriptor) Create(key string, abfData *abf.ABF) (*abfidx.ABFMetadata, error) {
	abfIdx, err := getABFIndex(abfData)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	// get ACL index
	aclData, exists := d.aclIndex.LookupByName(abfData.AclName)
	if !exists {
		err := errors.Errorf("failed to obtain metadata for ACL %s", abfData.AclName)
		d.log.Error(err)
		return nil, err
	}

	// add new ABF policy
	if err := d.abfHandler.AddAbfPolicy(uint32(abfIdx), aclData.Index, abfData.ForwardingPaths); err != nil {
		d.log.Error(err)
		return nil, err
	}

	// fill the metadata
	metadata := &abfidx.ABFMetadata{
		Index:    uint32(abfIdx),
		Attached: abfData.AttachedInterfaces,
	}

	return metadata, nil
}

// Delete removes ABF policy
func (d *ABFDescriptor) Delete(key string, abfData *abf.ABF, metadata *abfidx.ABFMetadata) error {
	// get ACL index
	aclData, exists := d.aclIndex.LookupByName(abfData.AclName)
	if !exists {
		err := errors.Errorf("failed to obtain metadata for ACL %s", abfData.AclName)
		d.log.Error(err)
	}

	return d.abfHandler.DeleteAbfPolicy(metadata.Index, aclData.Index, abfData.ForwardingPaths)
}

// UpdateWithRecreate is always set to true since there is no binary API to specialy handle a part
// of the config - the whole ABF policy needs to be removed and created.
func (d *ABFDescriptor) UpdateWithRecreate(key string, oldAbfData, newAbfData *abf.ABF, oldMetadata *abfidx.ABFMetadata) bool {
	// always recreate
	return true
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

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (d *ABFDescriptor) IsRetriableFailure(err error) bool {
	return !(err == ErrABFInvalidIndex || err == ErrABFWithoutACL)
}

// DerivedValues returns list of derived values for ABF.
func (d *ABFDescriptor) DerivedValues(key string, value *abf.ABF) (derived []api.KeyValuePair) {
	for _, attachedIf := range value.GetAttachedInterfaces() {
		derived = append(derived, api.KeyValuePair{
			Key:   abf.ToABFInterfaceKey(value.Index, attachedIf.InputInterface),
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
				oldPath.Vrf == newPath.Vrf &&
				oldPath.Weight == newPath.Weight &&
				oldPath.Preference == newPath.Preference &&
				oldPath.Afi == newPath.Afi &&
				oldPath.RpfId == newPath.RpfId &&
				oldPath.ViaLabel == newPath.ViaLabel &&
				oldPath.Local == newPath.Local &&
				oldPath.Drop == newPath.Drop &&
				oldPath.UdpEncap == newPath.UdpEncap &&
				oldPath.Unreachable == newPath.Unreachable &&
				oldPath.Prohibit == newPath.Prohibit &&
				oldPath.ResolveHost == newPath.ResolveHost &&
				oldPath.ResolveAttached == newPath.ResolveAttached &&
				oldPath.Dvr == newPath.Dvr &&
				oldPath.SourceLookup == newPath.SourceLookup &&
				equivalentLabelStacks(oldPath.LabelStack, newPath.LabelStack) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equivalentLabelStacks(oldLabelStack, newLabelStack []*abf.ABF_ForwardingPath_Label) bool {
	if len(oldLabelStack) != len(newLabelStack) {
		return false
	}
	for _, oldLabelStackEntry := range oldLabelStack {
		var found bool
		for _, newLabelStackEntry := range newLabelStack {
			if proto.Equal(oldLabelStackEntry, newLabelStackEntry) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func getABFIndex(abfData *abf.ABF) (index int, err error) {
	index, err = strconv.Atoi(abfData.Index)
	if err != nil {
		return index, errors.Errorf("cannot convert ABF index %s: %v", abfData.Index, err)
	}
	return index, nil
}
