package descriptor

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/aclidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/acl"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	"github.com/pkg/errors"
)

const (
	// ACLToInterfaceDescriptorName is name for descriptor
	ACLToInterfaceDescriptorName = "ACLInterface"

	// dependency labels
	interfaceDep = "acl-interface-existence"
)

type ACLToInterfaceDescriptor struct {
	log        logging.Logger
	aclHandler vppcalls.ACLVppAPI
	aclIndex   aclidx.AclMetadataIndex
}

// NewACLToInterfaceDescriptor returns new ACLInterface descriptor
func NewACLToInterfaceDescriptor(aclIndex aclidx.AclMetadataIndex, aclHandler vppcalls.ACLVppAPI, log logging.PluginLogger) *ACLToInterfaceDescriptor {
	return &ACLToInterfaceDescriptor{
		log:        log,
		aclIndex:   aclIndex,
		aclHandler: aclHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration with the KVScheduler.
func (d *ACLToInterfaceDescriptor) GetDescriptor() *scheduler.KVDescriptor {
	return &scheduler.KVDescriptor{
		Name:         ACLToInterfaceDescriptorName,
		KeySelector:  d.IsACLInterfaceKey,
		Add:          d.Add,
		Delete:       d.Delete,
		Dependencies: d.Dependencies,
	}
}

// IsACLInterfaceKey returns true if the key is identifying ACL interface (derived value)
func (d *ACLToInterfaceDescriptor) IsACLInterfaceKey(key string) bool {
	_, _, _, isACLToInterfaceKey := acl.ParseACLToInterfaceKey(key)
	return isACLToInterfaceKey
}

// Add enables DHCP client.
func (d *ACLToInterfaceDescriptor) Add(key string, emptyVal proto.Message) (metadata scheduler.Metadata, err error) {
	aclName, ifName, _, _ := acl.ParseACLToInterfaceKey(key)

	d.log.Warnf(" ADD: %v %v", aclName, ifName)

	aclMeta, found := d.aclIndex.LookupByName(aclName)
	if !found {
		err = errors.Errorf("failed to obtain metadata for ACL %s", aclName)
		d.log.Error(err)
		return nil, err
	}

	if aclMeta.L2 {
		if err := d.aclHandler.AddMACIPACLToInterface(aclMeta.Index, ifName); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("not implemented yet")
	}

	return nil, err
}

// Delete disables DHCP client.
func (d *ACLToInterfaceDescriptor) Delete(key string, emptyVal proto.Message, metadata scheduler.Metadata) error {

	return fmt.Errorf("not implemented yet")
}

// Dependencies lists the interface as the only dependency for the binding.
func (d *ACLToInterfaceDescriptor) Dependencies(key string, emptyVal proto.Message) []scheduler.Dependency {
	_, ifName, _, _ := acl.ParseACLToInterfaceKey(key)
	return []scheduler.Dependency{
		{
			Label: interfaceDep,
			Key:   interfaces.InterfaceKey(ifName),
		},
	}
}
