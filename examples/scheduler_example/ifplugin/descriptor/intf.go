package descriptor

import (
	"fmt"
	"strings"

	"github.com/ligato/cn-infra/idxmap"
	. "github.com/ligato/cn-infra/kvscheduler/api"
	. "github.com/ligato/cn-infra/kvscheduler/value/protoval"
	"github.com/ligato/cn-infra/logging/logrus"

	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	. "github.com/ligato/vpp-agent/examples/scheduler_example/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/examples/scheduler_example/ifplugin/ifaceidx"
)

// If model was together with plugin, we could use relative path, e.g.: "../model/interfaces":
//   go:generate adapter-generator --descriptor-name Intf --is-proto --value-type *interfaces.Interfaces_Interface --meta-type *ifaceidx.IfaceMetadata --from-datasync --import "../model/interfaces" --import "../ifaceidx"

//go:generate adapter-generator --descriptor-name Intf --is-proto --value-type *interfaces.Interfaces_Interface --meta-type *ifaceidx.IfaceMetadata --from-datasync --import "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces" --import "../ifaceidx"


// Example how default proto value can be customized:

type InterfaceProtoValue struct {
	ProtoValue
	iface *interfaces.Interfaces_Interface
}

func (ipv *InterfaceProtoValue) Label() string {
	return strings.ToUpper(ipv.iface.Name)
}


// Example of descriptor implementation:

type IntfDescriptorImpl struct {
	IntfDescriptorBase
}

func (intfd *IntfDescriptorImpl) GetName() string {
	return "interface"
}

func (intfd *IntfDescriptorImpl) KeySelector(key string) bool {
	return strings.HasPrefix(key, interfaces.Prefix)
}

func (intfd *IntfDescriptorImpl) NBKeyPrefixes() []string {
	return []string{interfaces.Prefix}
}

func (intfd *IntfDescriptorImpl) WithMetadata() (withMeta bool, customMapFactory MetadataMapFactory) {
	return true, func() idxmap.NamedMappingRW {
		return ifaceidx.NewIfaceIndex(logrus.DefaultLogger(), "interface-index")
	}
}

func (intfd *IntfDescriptorImpl) Build(key string, valueData *interfaces.Interfaces_Interface) (value ProtoValue, err error) {
	return &InterfaceProtoValue{ProtoValue: NewProtoValue(valueData), iface: valueData}, nil
}

func (intfd *IntfDescriptorImpl) Add(key string, value *interfaces.Interfaces_Interface) (metadata *ifaceidx.IfaceMetadata, err error) {
	metadata = &ifaceidx.IfaceMetadata{IpAddresses: value.IpAddresses, SwIfIndex: 10}
	fmt.Printf("Add interface with name:%s, under key:%s, sw_if_index: %d, hw-addr:%s\n", value.Name, key, metadata.SwIfIndex, value.PhysAddress)
	return metadata, nil
}

func (intfd *IntfDescriptorImpl) Modify(key string, oldValue, newValue *interfaces.Interfaces_Interface, oldMetadata *ifaceidx.IfaceMetadata) (newMetadata *ifaceidx.IfaceMetadata, err error) {
	fmt.Printf("Modified interface with name:%s, under key:%s, new-hw-addr:%s,\n",
		oldValue.Name, key, newValue.PhysAddress)
	oldMetadata.IpAddresses = newValue.IpAddresses
	return oldMetadata, nil
}

func (intfd *IntfDescriptorImpl) ModifyHasToRecreate(key string, oldValue, newValue *interfaces.Interfaces_Interface, metadata *ifaceidx.IfaceMetadata) bool {
	return oldValue.Tap.HostIfName != newValue.Tap.HostIfName
}
