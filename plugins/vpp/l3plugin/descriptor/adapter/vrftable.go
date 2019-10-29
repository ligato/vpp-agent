// Code generated by adapter-generator. DO NOT EDIT.

package adapter

import (
	"github.com/golang/protobuf/proto"
	. "go.ligato.io/vpp-agent/v2/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/l3plugin/vrfidx"
	"go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp/l3"
)

////////// type-safe key-value pair with metadata //////////

type VrfTableKVWithMetadata struct {
	Key      string
	Value    *vpp_l3.VrfTable
	Metadata *vrfidx.VRFMetadata
	Origin   ValueOrigin
}

////////// type-safe Descriptor structure //////////

type VrfTableDescriptor struct {
	Name                 string
	KeySelector          KeySelector
	ValueTypeName        string
	KeyLabel             func(key string) string
	ValueComparator      func(key string, oldValue, newValue *vpp_l3.VrfTable) bool
	NBKeyPrefix          string
	WithMetadata         bool
	MetadataMapFactory   MetadataMapFactory
	Validate             func(key string, value *vpp_l3.VrfTable) error
	Create               func(key string, value *vpp_l3.VrfTable) (metadata *vrfidx.VRFMetadata, err error)
	Delete               func(key string, value *vpp_l3.VrfTable, metadata *vrfidx.VRFMetadata) error
	Update               func(key string, oldValue, newValue *vpp_l3.VrfTable, oldMetadata *vrfidx.VRFMetadata) (newMetadata *vrfidx.VRFMetadata, err error)
	UpdateWithRecreate   func(key string, oldValue, newValue *vpp_l3.VrfTable, metadata *vrfidx.VRFMetadata) bool
	Retrieve             func(correlate []VrfTableKVWithMetadata) ([]VrfTableKVWithMetadata, error)
	IsRetriableFailure   func(err error) bool
	DerivedValues        func(key string, value *vpp_l3.VrfTable) []KeyValuePair
	Dependencies         func(key string, value *vpp_l3.VrfTable) []Dependency
	RetrieveDependencies []string /* descriptor name */
}

////////// Descriptor adapter //////////

type VrfTableDescriptorAdapter struct {
	descriptor *VrfTableDescriptor
}

func NewVrfTableDescriptor(typedDescriptor *VrfTableDescriptor) *KVDescriptor {
	adapter := &VrfTableDescriptorAdapter{descriptor: typedDescriptor}
	descriptor := &KVDescriptor{
		Name:                 typedDescriptor.Name,
		KeySelector:          typedDescriptor.KeySelector,
		ValueTypeName:        typedDescriptor.ValueTypeName,
		KeyLabel:             typedDescriptor.KeyLabel,
		NBKeyPrefix:          typedDescriptor.NBKeyPrefix,
		WithMetadata:         typedDescriptor.WithMetadata,
		MetadataMapFactory:   typedDescriptor.MetadataMapFactory,
		IsRetriableFailure:   typedDescriptor.IsRetriableFailure,
		RetrieveDependencies: typedDescriptor.RetrieveDependencies,
	}
	if typedDescriptor.ValueComparator != nil {
		descriptor.ValueComparator = adapter.ValueComparator
	}
	if typedDescriptor.Validate != nil {
		descriptor.Validate = adapter.Validate
	}
	if typedDescriptor.Create != nil {
		descriptor.Create = adapter.Create
	}
	if typedDescriptor.Delete != nil {
		descriptor.Delete = adapter.Delete
	}
	if typedDescriptor.Update != nil {
		descriptor.Update = adapter.Update
	}
	if typedDescriptor.UpdateWithRecreate != nil {
		descriptor.UpdateWithRecreate = adapter.UpdateWithRecreate
	}
	if typedDescriptor.Retrieve != nil {
		descriptor.Retrieve = adapter.Retrieve
	}
	if typedDescriptor.Dependencies != nil {
		descriptor.Dependencies = adapter.Dependencies
	}
	if typedDescriptor.DerivedValues != nil {
		descriptor.DerivedValues = adapter.DerivedValues
	}
	return descriptor
}

func (da *VrfTableDescriptorAdapter) ValueComparator(key string, oldValue, newValue proto.Message) bool {
	typedOldValue, err1 := castVrfTableValue(key, oldValue)
	typedNewValue, err2 := castVrfTableValue(key, newValue)
	if err1 != nil || err2 != nil {
		return false
	}
	return da.descriptor.ValueComparator(key, typedOldValue, typedNewValue)
}

func (da *VrfTableDescriptorAdapter) Validate(key string, value proto.Message) (err error) {
	typedValue, err := castVrfTableValue(key, value)
	if err != nil {
		return err
	}
	return da.descriptor.Validate(key, typedValue)
}

func (da *VrfTableDescriptorAdapter) Create(key string, value proto.Message) (metadata Metadata, err error) {
	typedValue, err := castVrfTableValue(key, value)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Create(key, typedValue)
}

func (da *VrfTableDescriptorAdapter) Update(key string, oldValue, newValue proto.Message, oldMetadata Metadata) (newMetadata Metadata, err error) {
	oldTypedValue, err := castVrfTableValue(key, oldValue)
	if err != nil {
		return nil, err
	}
	newTypedValue, err := castVrfTableValue(key, newValue)
	if err != nil {
		return nil, err
	}
	typedOldMetadata, err := castVrfTableMetadata(key, oldMetadata)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Update(key, oldTypedValue, newTypedValue, typedOldMetadata)
}

func (da *VrfTableDescriptorAdapter) Delete(key string, value proto.Message, metadata Metadata) error {
	typedValue, err := castVrfTableValue(key, value)
	if err != nil {
		return err
	}
	typedMetadata, err := castVrfTableMetadata(key, metadata)
	if err != nil {
		return err
	}
	return da.descriptor.Delete(key, typedValue, typedMetadata)
}

func (da *VrfTableDescriptorAdapter) UpdateWithRecreate(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
	oldTypedValue, err := castVrfTableValue(key, oldValue)
	if err != nil {
		return true
	}
	newTypedValue, err := castVrfTableValue(key, newValue)
	if err != nil {
		return true
	}
	typedMetadata, err := castVrfTableMetadata(key, metadata)
	if err != nil {
		return true
	}
	return da.descriptor.UpdateWithRecreate(key, oldTypedValue, newTypedValue, typedMetadata)
}

func (da *VrfTableDescriptorAdapter) Retrieve(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	var correlateWithType []VrfTableKVWithMetadata
	for _, kvpair := range correlate {
		typedValue, err := castVrfTableValue(kvpair.Key, kvpair.Value)
		if err != nil {
			continue
		}
		typedMetadata, err := castVrfTableMetadata(kvpair.Key, kvpair.Metadata)
		if err != nil {
			continue
		}
		correlateWithType = append(correlateWithType,
			VrfTableKVWithMetadata{
				Key:      kvpair.Key,
				Value:    typedValue,
				Metadata: typedMetadata,
				Origin:   kvpair.Origin,
			})
	}

	typedValues, err := da.descriptor.Retrieve(correlateWithType)
	if err != nil {
		return nil, err
	}
	var values []KVWithMetadata
	for _, typedKVWithMetadata := range typedValues {
		kvWithMetadata := KVWithMetadata{
			Key:      typedKVWithMetadata.Key,
			Metadata: typedKVWithMetadata.Metadata,
			Origin:   typedKVWithMetadata.Origin,
		}
		kvWithMetadata.Value = typedKVWithMetadata.Value
		values = append(values, kvWithMetadata)
	}
	return values, err
}

func (da *VrfTableDescriptorAdapter) DerivedValues(key string, value proto.Message) []KeyValuePair {
	typedValue, err := castVrfTableValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.DerivedValues(key, typedValue)
}

func (da *VrfTableDescriptorAdapter) Dependencies(key string, value proto.Message) []Dependency {
	typedValue, err := castVrfTableValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.Dependencies(key, typedValue)
}

////////// Helper methods //////////

func castVrfTableValue(key string, value proto.Message) (*vpp_l3.VrfTable, error) {
	typedValue, ok := value.(*vpp_l3.VrfTable)
	if !ok {
		return nil, ErrInvalidValueType(key, value)
	}
	return typedValue, nil
}

func castVrfTableMetadata(key string, metadata Metadata) (*vrfidx.VRFMetadata, error) {
	if metadata == nil {
		return nil, nil
	}
	typedMetadata, ok := metadata.(*vrfidx.VRFMetadata)
	if !ok {
		return nil, ErrInvalidMetadataType(key)
	}
	return typedMetadata, nil
}
