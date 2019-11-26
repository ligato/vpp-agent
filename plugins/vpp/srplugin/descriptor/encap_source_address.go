package descriptor

import (
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	srv6 "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/srv6"
	scheduler "go.ligato.io/vpp-agent/v2/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/vppcalls"
)

const (
	// LocalSIDDescriptorName is the name of the descriptor for VPP LocalSIDs
	EncapSourceAddressDescriptorName = "vpp-encap-source-address"
)

const (
	defaultEncapSource = "::"
)

var defaultEncapSourceAddress = &srv6.EncapSourceAddress{
	EncapSource: defaultEncapSource,
}

// EncapSourceAddressDescriptor teaches KVScheduler how to configure VPP SR Encap Source Address.
type EncapSourceAddressDescriptor struct {
	// dependencies
	log       logging.Logger
	srHandler vppcalls.SRv6VppAPI
}

// NewEncapSourceAddressDescriptor creates a new instance of the EncapSourceAddressDescriptor.
func NewEncapSourceAddressDescriptor(srHandler vppcalls.SRv6VppAPI, log logging.PluginLogger) *scheduler.KVDescriptor {
	ctx := &EncapSourceAddressDescriptor{
		log:       log.NewLogger("encapsource-descriptor"),
		srHandler: srHandler,
	}

	typedDescr := &adapter.EncapSourceAddressDescriptor{
		Name:            EncapSourceAddressDescriptorName,
		NBKeyPrefix:     srv6.ModelEncapSourceAddress.KeyPrefix(),
		ValueTypeName:   srv6.ModelEncapSourceAddress.ProtoName(),
		KeySelector:     srv6.ModelEncapSourceAddress.IsKeyValid,
		ValueComparator: ctx.EquivalentEncapSourceAddresses,
		Create:          ctx.Create,
		Update:          ctx.Update,
		Delete:          ctx.Delete,
	}
	return adapter.NewEncapSourceAddressDescriptor(typedDescr)
}


// EquivalentEncapSourceAddresses compares the IP Scan Neighbor values.
func (d *EncapSourceAddressDescriptor) EquivalentEncapSourceAddresses(key string, oldValue, newValue *srv6.EncapSourceAddress) bool {
	return proto.Equal(withDefaults(oldValue), withDefaults(newValue))
}

// Create adds VPP IP Scan Neighbor.
func (d *EncapSourceAddressDescriptor) Create(key string, value *srv6.EncapSourceAddress) (metadata interface{}, err error) {
	return d.Update(key, defaultEncapSourceAddress, withDefaults(value), nil)
}

// Delete deletes VPP IP Scan Neighbor.
func (d *EncapSourceAddressDescriptor) Delete(key string, value *srv6.EncapSourceAddress, metadata interface{}) error {
	_, err := d.Update(key, withDefaults(value), defaultEncapSourceAddress, metadata)
	return err
}

// Update modifies VPP IP Scan Neighbor.
func (d *EncapSourceAddressDescriptor) Update(key string, oldValue, newValue *srv6.EncapSourceAddress, oldMetadata interface{}) (newMetadata interface{}, err error) {
	if err := d.srHandler.SetEncapsSourceAddress(newValue.EncapSource); err != nil {
		return nil, err
	}
	return nil, nil
}

func withDefaults(orig *srv6.EncapSourceAddress) *srv6.EncapSourceAddress {
	var val = *orig
	if val.EncapSource == "" {
		val.EncapSource = defaultEncapSource
	}
	return &val
}
