// Code generated by adapter-generator. DO NOT EDIT.

package adapter

import (
	"github.com/golang/protobuf/proto"
	. "go.ligato.io/vpp-agent/v2/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/aclplugin/aclidx"
	"go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp/acl"
)

////////// type-safe key-value pair with metadata //////////

type ACLKVWithMetadata struct {
	Key      string
	Value    *vpp_acl.ACL
	Metadata *aclidx.ACLMetadata
	Origin   ValueOrigin
}

////////// type-safe Descriptor structure //////////

type ACLDescriptor struct {
	Name                 string
	KeySelector          KeySelector
	ValueTypeName        string
	KeyLabel             func(key string) string
	ValueComparator      func(key string, oldValue, newValue *vpp_acl.ACL) bool
	NBKeyPrefix          string
	WithMetadata         bool
	MetadataMapFactory   MetadataMapFactory
	Validate             func(key string, value *vpp_acl.ACL) error
	Create               func(key string, value *vpp_acl.ACL) (metadata *aclidx.ACLMetadata, err error)
	Delete               func(key string, value *vpp_acl.ACL, metadata *aclidx.ACLMetadata) error
	Update               func(key string, oldValue, newValue *vpp_acl.ACL, oldMetadata *aclidx.ACLMetadata) (newMetadata *aclidx.ACLMetadata, err error)
	UpdateWithRecreate   func(key string, oldValue, newValue *vpp_acl.ACL, metadata *aclidx.ACLMetadata) bool
	Retrieve             func(correlate []ACLKVWithMetadata) ([]ACLKVWithMetadata, error)
	IsRetriableFailure   func(err error) bool
	DerivedValues        func(key string, value *vpp_acl.ACL) []KeyValuePair
	Dependencies         func(key string, value *vpp_acl.ACL) []Dependency
	RetrieveDependencies []string /* descriptor name */
}

////////// Descriptor adapter //////////

type ACLDescriptorAdapter struct {
	descriptor *ACLDescriptor
}

func NewACLDescriptor(typedDescriptor *ACLDescriptor) *KVDescriptor {
	adapter := &ACLDescriptorAdapter{descriptor: typedDescriptor}
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

func (da *ACLDescriptorAdapter) ValueComparator(key string, oldValue, newValue proto.Message) bool {
	typedOldValue, err1 := castACLValue(key, oldValue)
	typedNewValue, err2 := castACLValue(key, newValue)
	if err1 != nil || err2 != nil {
		return false
	}
	return da.descriptor.ValueComparator(key, typedOldValue, typedNewValue)
}

func (da *ACLDescriptorAdapter) Validate(key string, value proto.Message) (err error) {
	typedValue, err := castACLValue(key, value)
	if err != nil {
		return err
	}
	return da.descriptor.Validate(key, typedValue)
}

func (da *ACLDescriptorAdapter) Create(key string, value proto.Message) (metadata Metadata, err error) {
	typedValue, err := castACLValue(key, value)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Create(key, typedValue)
}

func (da *ACLDescriptorAdapter) Update(key string, oldValue, newValue proto.Message, oldMetadata Metadata) (newMetadata Metadata, err error) {
	oldTypedValue, err := castACLValue(key, oldValue)
	if err != nil {
		return nil, err
	}
	newTypedValue, err := castACLValue(key, newValue)
	if err != nil {
		return nil, err
	}
	typedOldMetadata, err := castACLMetadata(key, oldMetadata)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Update(key, oldTypedValue, newTypedValue, typedOldMetadata)
}

func (da *ACLDescriptorAdapter) Delete(key string, value proto.Message, metadata Metadata) error {
	typedValue, err := castACLValue(key, value)
	if err != nil {
		return err
	}
	typedMetadata, err := castACLMetadata(key, metadata)
	if err != nil {
		return err
	}
	return da.descriptor.Delete(key, typedValue, typedMetadata)
}

func (da *ACLDescriptorAdapter) UpdateWithRecreate(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
	oldTypedValue, err := castACLValue(key, oldValue)
	if err != nil {
		return true
	}
	newTypedValue, err := castACLValue(key, newValue)
	if err != nil {
		return true
	}
	typedMetadata, err := castACLMetadata(key, metadata)
	if err != nil {
		return true
	}
	return da.descriptor.UpdateWithRecreate(key, oldTypedValue, newTypedValue, typedMetadata)
}

func (da *ACLDescriptorAdapter) Retrieve(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	var correlateWithType []ACLKVWithMetadata
	for _, kvpair := range correlate {
		typedValue, err := castACLValue(kvpair.Key, kvpair.Value)
		if err != nil {
			continue
		}
		typedMetadata, err := castACLMetadata(kvpair.Key, kvpair.Metadata)
		if err != nil {
			continue
		}
		correlateWithType = append(correlateWithType,
			ACLKVWithMetadata{
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

func (da *ACLDescriptorAdapter) DerivedValues(key string, value proto.Message) []KeyValuePair {
	typedValue, err := castACLValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.DerivedValues(key, typedValue)
}

func (da *ACLDescriptorAdapter) Dependencies(key string, value proto.Message) []Dependency {
	typedValue, err := castACLValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.Dependencies(key, typedValue)
}

////////// Helper methods //////////

func castACLValue(key string, value proto.Message) (*vpp_acl.ACL, error) {
	typedValue, ok := value.(*vpp_acl.ACL)
	if !ok {
		return nil, ErrInvalidValueType(key, value)
	}
	return typedValue, nil
}

func castACLMetadata(key string, metadata Metadata) (*aclidx.ACLMetadata, error) {
	if metadata == nil {
		return nil, nil
	}
	typedMetadata, ok := metadata.(*aclidx.ACLMetadata)
	if !ok {
		return nil, ErrInvalidMetadataType(key)
	}
	return typedMetadata, nil
}
