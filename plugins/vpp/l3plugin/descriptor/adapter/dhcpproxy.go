// Code generated by adapter-generator. DO NOT EDIT.

package adapter

import (
	"github.com/golang/protobuf/proto"
	. "go.ligato.io/vpp-agent/v2/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp/l3"
)

////////// type-safe key-value pair with metadata //////////

type DHCPProxyKVWithMetadata struct {
	Key      string
	Value    *vpp_l3.DHCPProxy
	Metadata interface{}
	Origin   ValueOrigin
}

////////// type-safe Descriptor structure //////////

type DHCPProxyDescriptor struct {
	Name                 string
	KeySelector          KeySelector
	ValueTypeName        string
	KeyLabel             func(key string) string
	ValueComparator      func(key string, oldValue, newValue *vpp_l3.DHCPProxy) bool
	NBKeyPrefix          string
	WithMetadata         bool
	MetadataMapFactory   MetadataMapFactory
	Validate             func(key string, value *vpp_l3.DHCPProxy) error
	Create               func(key string, value *vpp_l3.DHCPProxy) (metadata interface{}, err error)
	Delete               func(key string, value *vpp_l3.DHCPProxy, metadata interface{}) error
	Update               func(key string, oldValue, newValue *vpp_l3.DHCPProxy, oldMetadata interface{}) (newMetadata interface{}, err error)
	UpdateWithRecreate   func(key string, oldValue, newValue *vpp_l3.DHCPProxy, metadata interface{}) bool
	Retrieve             func(correlate []DHCPProxyKVWithMetadata) ([]DHCPProxyKVWithMetadata, error)
	IsRetriableFailure   func(err error) bool
	DerivedValues        func(key string, value *vpp_l3.DHCPProxy) []KeyValuePair
	Dependencies         func(key string, value *vpp_l3.DHCPProxy) []Dependency
	RetrieveDependencies []string /* descriptor name */
}

////////// Descriptor adapter //////////

type DHCPProxyDescriptorAdapter struct {
	descriptor *DHCPProxyDescriptor
}

func NewDHCPProxyDescriptor(typedDescriptor *DHCPProxyDescriptor) *KVDescriptor {
	adapter := &DHCPProxyDescriptorAdapter{descriptor: typedDescriptor}
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

func (da *DHCPProxyDescriptorAdapter) ValueComparator(key string, oldValue, newValue proto.Message) bool {
	typedOldValue, err1 := castDHCPProxyValue(key, oldValue)
	typedNewValue, err2 := castDHCPProxyValue(key, newValue)
	if err1 != nil || err2 != nil {
		return false
	}
	return da.descriptor.ValueComparator(key, typedOldValue, typedNewValue)
}

func (da *DHCPProxyDescriptorAdapter) Validate(key string, value proto.Message) (err error) {
	typedValue, err := castDHCPProxyValue(key, value)
	if err != nil {
		return err
	}
	return da.descriptor.Validate(key, typedValue)
}

func (da *DHCPProxyDescriptorAdapter) Create(key string, value proto.Message) (metadata Metadata, err error) {
	typedValue, err := castDHCPProxyValue(key, value)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Create(key, typedValue)
}

func (da *DHCPProxyDescriptorAdapter) Update(key string, oldValue, newValue proto.Message, oldMetadata Metadata) (newMetadata Metadata, err error) {
	oldTypedValue, err := castDHCPProxyValue(key, oldValue)
	if err != nil {
		return nil, err
	}
	newTypedValue, err := castDHCPProxyValue(key, newValue)
	if err != nil {
		return nil, err
	}
	typedOldMetadata, err := castDHCPProxyMetadata(key, oldMetadata)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Update(key, oldTypedValue, newTypedValue, typedOldMetadata)
}

func (da *DHCPProxyDescriptorAdapter) Delete(key string, value proto.Message, metadata Metadata) error {
	typedValue, err := castDHCPProxyValue(key, value)
	if err != nil {
		return err
	}
	typedMetadata, err := castDHCPProxyMetadata(key, metadata)
	if err != nil {
		return err
	}
	return da.descriptor.Delete(key, typedValue, typedMetadata)
}

func (da *DHCPProxyDescriptorAdapter) UpdateWithRecreate(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
	oldTypedValue, err := castDHCPProxyValue(key, oldValue)
	if err != nil {
		return true
	}
	newTypedValue, err := castDHCPProxyValue(key, newValue)
	if err != nil {
		return true
	}
	typedMetadata, err := castDHCPProxyMetadata(key, metadata)
	if err != nil {
		return true
	}
	return da.descriptor.UpdateWithRecreate(key, oldTypedValue, newTypedValue, typedMetadata)
}

func (da *DHCPProxyDescriptorAdapter) Retrieve(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	var correlateWithType []DHCPProxyKVWithMetadata
	for _, kvpair := range correlate {
		typedValue, err := castDHCPProxyValue(kvpair.Key, kvpair.Value)
		if err != nil {
			continue
		}
		typedMetadata, err := castDHCPProxyMetadata(kvpair.Key, kvpair.Metadata)
		if err != nil {
			continue
		}
		correlateWithType = append(correlateWithType,
			DHCPProxyKVWithMetadata{
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

func (da *DHCPProxyDescriptorAdapter) DerivedValues(key string, value proto.Message) []KeyValuePair {
	typedValue, err := castDHCPProxyValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.DerivedValues(key, typedValue)
}

func (da *DHCPProxyDescriptorAdapter) Dependencies(key string, value proto.Message) []Dependency {
	typedValue, err := castDHCPProxyValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.Dependencies(key, typedValue)
}

////////// Helper methods //////////

func castDHCPProxyValue(key string, value proto.Message) (*vpp_l3.DHCPProxy, error) {
	typedValue, ok := value.(*vpp_l3.DHCPProxy)
	if !ok {
		return nil, ErrInvalidValueType(key, value)
	}
	return typedValue, nil
}

func castDHCPProxyMetadata(key string, metadata Metadata) (interface{}, error) {
	if metadata == nil {
		return nil, nil
	}
	typedMetadata, ok := metadata.(interface{})
	if !ok {
		return nil, ErrInvalidMetadataType(key)
	}
	return typedMetadata, nil
}
