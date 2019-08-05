// Copyright (c) 2019 PANTHEON.tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package descriptor

import (
	"github.com/go-errors/errors"
	"github.com/ligato/cn-infra/logging"
	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// SpanDescriptorName is the name of the descriptor.
	SpanDescriptorName = "span"
)

// A list of non-retriable errors:
var (
	ErrSpanWithoutInterface = errors.New("VPP SPAN defined without From/To interface")
)

// SpanDescriptor teaches KVScheduler how to configure VPP SPAN.
type SpanDescriptor struct {
	log         logging.Logger
	spanHandler vppcalls.InterfaceVppAPI
	intfIndex   ifaceidx.IfaceMetadataIndex
}

// NewSpanDescriptor creates a new instance of the SpanDescriptor.
func NewSpanDescriptor(spanHandler vppcalls.InterfaceVppAPI, log logging.PluginLogger) (*kvs.KVDescriptor, *SpanDescriptor) {

	ctx := &SpanDescriptor{
		spanHandler: spanHandler,
		log:         log.NewLogger("span-descriptor"),
	}

	typedDescr := &adapter.SpanDescriptor{
		Name:          SpanDescriptorName,
		KeySelector:   interfaces.ModelSpan.IsKeyValid,
		KeyLabel:      interfaces.ModelSpan.StripKeyPrefix,
		NBKeyPrefix:   interfaces.ModelSpan.KeyPrefix(),
		ValueTypeName: interfaces.ModelSpan.ProtoName(),
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Retrieve:      ctx.Retrieve,
		Validate:      ctx.Validate,
		Dependencies:  ctx.Dependencies,
	}

	return adapter.NewSpanDescriptor(typedDescr), ctx
}

// SetInterfaceIndex should be used to provide interface index immediately after
// the descriptor registration.
func (d *SpanDescriptor) SetInterfaceIndex(intfIndex ifaceidx.IfaceMetadataIndex) {
	d.intfIndex = intfIndex
}

// Validate does.
func (d *SpanDescriptor) Validate(key string, value *interfaces.Span) error {
	if value.InterfaceFrom == "" && value.InterfaceTo == "" {
		return kvs.NewInvalidValueError(ErrSpanWithoutInterface,
			"interface_from", "interface_to")
	}
	if value.InterfaceFrom == "" {
		return kvs.NewInvalidValueError(ErrSpanWithoutInterface, "interface_from")
	}
	if value.InterfaceTo == "" {
		return kvs.NewInvalidValueError(ErrSpanWithoutInterface, "interface_to")
	}
	return nil
}

// Create does.
func (d *SpanDescriptor) Create(key string, value *interfaces.Span) (metadata interface{}, err error) {
	ifaceFrom, found := d.intfIndex.LookupByName(value.InterfaceFrom)
	if !found {
		err = errors.Errorf("failed to find InterfaceFrom %s", value.InterfaceFrom)
		d.log.Error(err)
		return nil, err
	}

	ifaceTo, found := d.intfIndex.LookupByName(value.InterfaceTo)
	if !found {
		err = errors.Errorf("failed to find InterfaceTo %s", value.InterfaceTo)
		d.log.Error(err)
		return nil, err
	}

	var isL2 uint8
	if value.IsL2 {
		isL2 = 1
	}

	err = d.spanHandler.AddSpan(ifaceFrom.SwIfIndex, ifaceTo.SwIfIndex, uint8(value.State), isL2)
	if err != nil {
		err = errors.Errorf("failed to add interface span: %v", err)
		d.log.Error(err)
		return nil, err
	}

	return nil, err
}

// Delete does.
func (d *SpanDescriptor) Delete(key string, value *interfaces.Span, metadata interface{}) error {
	var err error
	ifaceFrom, found := d.intfIndex.LookupByName(value.InterfaceFrom)
	if !found {
		err = errors.Errorf("failed to find InterfaceFrom %s", value.InterfaceFrom)
		d.log.Error(err)
		return err
	}

	ifaceTo, found := d.intfIndex.LookupByName(value.InterfaceTo)
	if !found {
		err = errors.Errorf("failed to find InterfaceTo %s", value.InterfaceTo)
		d.log.Error(err)
		return err
	}

	var isL2 uint8
	if value.IsL2 {
		isL2 = 1
	}

	err = d.spanHandler.DelSpan(ifaceFrom.SwIfIndex, ifaceTo.SwIfIndex, isL2)
	if err != nil {
		err = errors.Errorf("failed to delete interface span: %v", err)
		d.log.Error(err)
		return err
	}

	return err
}

// Retrieve returns all records from VPP SPAN table.
func (d *SpanDescriptor) Retrieve(correlate []adapter.SpanKVWithMetadata) (retrieved []adapter.SpanKVWithMetadata, err error) {
	spans, err := d.spanHandler.DumpSpan()
	if err != nil {
		d.log.Error(err)
		return retrieved, err
	}

	var nameFrom, nameTo string
	var exists bool
	for _, s := range spans {
		nameFrom, _, exists = d.intfIndex.LookupBySwIfIndex(s.SwIfIndexFrom)
		if !exists {
			// should this break or error? dunno
			continue
		}
		nameTo, _, exists = d.intfIndex.LookupBySwIfIndex(s.SwIfIndexTo)
		if !exists {
			// should this break or error? dunno
			continue
		}
		var isL2 bool
		if s.IsL2 == 1 {
			isL2 = true
		}

		retrieved = append(retrieved, adapter.SpanKVWithMetadata{
			Key: interfaces.SpanKey(nameFrom, nameTo),
			Value: &interfaces.Span{
				InterfaceFrom: nameFrom,
				InterfaceTo:   nameTo,
				State:         interfaces.Span_State(s.State),
				IsL2:          isL2,
			},
			Origin: kvs.FromNB,
		})
	}
	return retrieved, nil
}

// Dependencies lists both From and To interfaces as dependencies.
func (d *SpanDescriptor) Dependencies(key string, value *interfaces.Span) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: "interface-from",
			Key:   interfaces.InterfaceKey(value.InterfaceFrom),
		},
		{
			Label: "interface-to",
			Key:   interfaces.InterfaceKey(value.InterfaceTo),
		},
	}
}
