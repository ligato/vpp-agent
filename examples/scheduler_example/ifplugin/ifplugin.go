package ifplugin

import (
	. "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/vpp-agent/examples/scheduler_example/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/examples/scheduler_example/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/examples/scheduler_example/ifplugin/ifaceidx"
	"github.com/pkg/errors"
)

type IfPlugin struct {
	Deps

	intfIndex ifaceidx.IfaceMetadataIndex
}

type Deps struct {
	Scheduler KVScheduler
}

func (p *IfPlugin) Init() error {
	descriptor := adapter.NewIntfDescriptor(&descriptor.IntfDescriptorImpl{})
	p.Deps.Scheduler.RegisterKVDescriptor(descriptor)

	var withIndex bool
	metadataMap := p.Deps.Scheduler.GetMetadataMap(descriptor.GetName())
	p.intfIndex, withIndex = metadataMap.(ifaceidx.IfaceMetadataIndex)
	if !withIndex {
		return errors.New("missing index with interface metadata")
	}
	return nil
}

func (p *IfPlugin) GetInterfaceIndex() ifaceidx.IfaceMetadataIndex {
	return p.intfIndex
}
