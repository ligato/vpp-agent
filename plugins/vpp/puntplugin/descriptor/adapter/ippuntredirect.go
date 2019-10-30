// Code generated by adapter-generator. DO NOT EDIT.

package adapter

import (
	"github.com/golang/protobuf/proto"
	. "go.ligato.io/vpp-agent/v2/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v2/proto/ligato/vpp/punt"
)

////////// type-safe key-value pair with metadata //////////

type IPPuntRedirectKVWithMetadata struct {
	Key      string
	Value    *vpp_punt.IPRedirect
	Metadata interface{}
	Origin   ValueOrigin
}

////////// type-safe Descriptor structure //////////

type IPPuntRedirectDescriptor struct {
	Name                 string
	KeySelector          KeySelector
	ValueTypeName        string
	KeyLabel             func(key string) string
	ValueComparator      func(key string, oldValue, newValue *vpp_punt.IPRedirect) bool
	NBKeyPrefix          string
	WithMetadata         bool
	MetadataMapFactory   MetadataMapFactory
	Validate             func(key string, value *vpp_punt.IPRedirect) error
	Create               func(key string, value *vpp_punt.IPRedirect) (metadata interface{}, err error)
	Delete               func(key string, value *vpp_punt.IPRedirect, metadata interface{}) error
	Update               func(key string, oldValue, newValue *vpp_punt.IPRedirect, oldMetadata interface{}) (newMetadata interface{}, err error)
	UpdateWithRecreate   func(key string, oldValue, newValue *vpp_punt.IPRedirect, metadata interface{}) bool
	Retrieve             func(correlate []IPPuntRedirectKVWithMetadata) ([]IPPuntRedirectKVWithMetadata, error)
	IsRetriableFailure   func(err error) bool
	DerivedValues        func(key string, value *vpp_punt.IPRedirect) []KeyValuePair
	Dependencies         func(key string, value *vpp_punt.IPRedirect) []Dependency
	RetrieveDependencies []string /* descriptor name */
}

////////// Descriptor adapter //////////

type IPPuntRedirectDescriptorAdapter struct {
	descriptor *IPPuntRedirectDescriptor
}

func NewIPPuntRedirectDescriptor(typedDescriptor *IPPuntRedirectDescriptor) *KVDescriptor {
	adapter := &IPPuntRedirectDescriptorAdapter{descriptor: typedDescriptor}
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

func (da *IPPuntRedirectDescriptorAdapter) ValueComparator(key string, oldValue, newValue proto.Message) bool {
	typedOldValue, err1 := castIPPuntRedirectValue(key, oldValue)
	typedNewValue, err2 := castIPPuntRedirectValue(key, newValue)
	if err1 != nil || err2 != nil {
		return false
	}
	return da.descriptor.ValueComparator(key, typedOldValue, typedNewValue)
}

func (da *IPPuntRedirectDescriptorAdapter) Validate(key string, value proto.Message) (err error) {
	typedValue, err := castIPPuntRedirectValue(key, value)
	if err != nil {
		return err
	}
	return da.descriptor.Validate(key, typedValue)
}

func (da *IPPuntRedirectDescriptorAdapter) Create(key string, value proto.Message) (metadata Metadata, err error) {
	typedValue, err := castIPPuntRedirectValue(key, value)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Create(key, typedValue)
}

func (da *IPPuntRedirectDescriptorAdapter) Update(key string, oldValue, newValue proto.Message, oldMetadata Metadata) (newMetadata Metadata, err error) {
	oldTypedValue, err := castIPPuntRedirectValue(key, oldValue)
	if err != nil {
		return nil, err
	}
	newTypedValue, err := castIPPuntRedirectValue(key, newValue)
	if err != nil {
		return nil, err
	}
	typedOldMetadata, err := castIPPuntRedirectMetadata(key, oldMetadata)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Update(key, oldTypedValue, newTypedValue, typedOldMetadata)
}

func (da *IPPuntRedirectDescriptorAdapter) Delete(key string, value proto.Message, metadata Metadata) error {
	typedValue, err := castIPPuntRedirectValue(key, value)
	if err != nil {
		return err
	}
	typedMetadata, err := castIPPuntRedirectMetadata(key, metadata)
	if err != nil {
		return err
	}
	return da.descriptor.Delete(key, typedValue, typedMetadata)
}

func (da *IPPuntRedirectDescriptorAdapter) UpdateWithRecreate(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
	oldTypedValue, err := castIPPuntRedirectValue(key, oldValue)
	if err != nil {
		return true
	}
	newTypedValue, err := castIPPuntRedirectValue(key, newValue)
	if err != nil {
		return true
	}
	typedMetadata, err := castIPPuntRedirectMetadata(key, metadata)
	if err != nil {
		return true
	}
	return da.descriptor.UpdateWithRecreate(key, oldTypedValue, newTypedValue, typedMetadata)
}

func (da *IPPuntRedirectDescriptorAdapter) Retrieve(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	var correlateWithType []IPPuntRedirectKVWithMetadata
	for _, kvpair := range correlate {
		typedValue, err := castIPPuntRedirectValue(kvpair.Key, kvpair.Value)
		if err != nil {
			continue
		}
		typedMetadata, err := castIPPuntRedirectMetadata(kvpair.Key, kvpair.Metadata)
		if err != nil {
			continue
		}
		correlateWithType = append(correlateWithType,
			IPPuntRedirectKVWithMetadata{
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

func (da *IPPuntRedirectDescriptorAdapter) DerivedValues(key string, value proto.Message) []KeyValuePair {
	typedValue, err := castIPPuntRedirectValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.DerivedValues(key, typedValue)
}

func (da *IPPuntRedirectDescriptorAdapter) Dependencies(key string, value proto.Message) []Dependency {
	typedValue, err := castIPPuntRedirectValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.Dependencies(key, typedValue)
}

////////// Helper methods //////////

func castIPPuntRedirectValue(key string, value proto.Message) (*vpp_punt.IPRedirect, error) {
	typedValue, ok := value.(*vpp_punt.IPRedirect)
	if !ok {
		return nil, ErrInvalidValueType(key, value)
	}
	return typedValue, nil
}

func castIPPuntRedirectMetadata(key string, metadata Metadata) (interface{}, error) {
	if metadata == nil {
		return nil, nil
	}
	typedMetadata, ok := metadata.(interface{})
	if !ok {
		return nil, ErrInvalidMetadataType(key)
	}
	return typedMetadata, nil
}
