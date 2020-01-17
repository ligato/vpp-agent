package descriptor

import (
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	scheduler "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls"
	srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"
)

const (
	// LocalSIDDescriptorName is the name of the descriptor for VPP LocalSIDs
	SRv6GlobalDescriptorName = "vpp-srv6-global"
)

const (
	defaultEncapSourceAddress = "::"
)

var defaultSRv6Global = &srv6.SRv6Global{
	EncapSourceAddress: defaultEncapSourceAddress,
}

// SRv6GlobalDescriptor teaches KVScheduler how to configure VPP SR Encap Source Address.
type SRv6GlobalDescriptor struct {
	// dependencies
	log       logging.Logger
	srHandler vppcalls.SRv6VppAPI
}

// NewSRv6GlobalDescriptor creates a new instance of the SRv6GlobalDescriptor.
func NewSRv6GlobalDescriptor(srHandler vppcalls.SRv6VppAPI, log logging.PluginLogger) *scheduler.KVDescriptor {
	ctx := &SRv6GlobalDescriptor{
		log:       log.NewLogger("encapsource-descriptor"),
		srHandler: srHandler,
	}

	typedDescr := &adapter.SRv6GlobalDescriptor{
		Name:            SRv6GlobalDescriptorName,
		NBKeyPrefix:     srv6.ModelSRv6Global.KeyPrefix(),
		ValueTypeName:   srv6.ModelSRv6Global.ProtoName(),
		KeySelector:     srv6.ModelSRv6Global.IsKeyValid,
		ValueComparator: ctx.EquivalentSRv6Global,
		Create:          ctx.Create,
		Update:          ctx.Update,
		Delete:          ctx.Delete,
	}
	return adapter.NewSRv6GlobalDescriptor(typedDescr)
}

// EquivalentSRv6Global compares the IP Scan Neighbor values.
func (d *SRv6GlobalDescriptor) EquivalentSRv6Global(key string, oldValue, newValue *srv6.SRv6Global) bool {
	return proto.Equal(withDefaults(oldValue), withDefaults(newValue))
}

// Create adds VPP IP Scan Neighbor.
func (d *SRv6GlobalDescriptor) Create(key string, value *srv6.SRv6Global) (metadata interface{}, err error) {
	return d.Update(key, defaultSRv6Global, withDefaults(value), nil)
}

// Delete deletes VPP IP Scan Neighbor.
func (d *SRv6GlobalDescriptor) Delete(key string, value *srv6.SRv6Global, metadata interface{}) error {
	_, err := d.Update(key, withDefaults(value), defaultSRv6Global, metadata)
	return err
}

// Update modifies VPP IP Scan Neighbor.
func (d *SRv6GlobalDescriptor) Update(key string, oldValue, newValue *srv6.SRv6Global, oldMetadata interface{}) (newMetadata interface{}, err error) {
	if err := d.srHandler.SetEncapsSourceAddress(newValue.EncapSourceAddress); err != nil {
		return nil, err
	}
	return nil, nil
}

func withDefaults(orig *srv6.SRv6Global) *srv6.SRv6Global {
	var val = *orig
	if val.EncapSourceAddress == "" {
		val.EncapSourceAddress = defaultEncapSourceAddress
	}
	return &val
}
