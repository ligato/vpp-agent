// Copyright (c) 2020 Cisco and/or its affiliates.
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
	"errors"

	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

const (
	FPParamsDescriptorName = "vpp-flowprobe-params"
)

// FPParamsDescriptor configures Flowprobe params for VPP.
type FPParamsDescriptor struct {
	ipfixHandler vppcalls.IpfixVppAPI
	log          logging.Logger
}

// NewFPParamsDescriptor creates a new instance of FPParamsDescriptor.
func NewFPParamsDescriptor(ipfixHandler vppcalls.IpfixVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &FPParamsDescriptor{
		ipfixHandler: ipfixHandler,
		log:          log.NewLogger("flowprobe-params-descriptor"),
	}
	typedDescr := &adapter.FlowProbeParamsDescriptor{
		Name:          FPParamsDescriptorName,
		NBKeyPrefix:   ipfix.ModelFlowprobeParams.KeyPrefix(),
		ValueTypeName: ipfix.ModelFlowprobeParams.ProtoName(),
		KeySelector:   ipfix.ModelFlowprobeParams.IsKeyValid,
		KeyLabel:      ipfix.ModelFlowprobeParams.StripKeyPrefix,
		Validate:      ctx.Validate,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Update:        ctx.Update,
	}
	return adapter.NewFlowProbeParamsDescriptor(typedDescr)
}

// Validate validates Flowprobe params.
func (d *FPParamsDescriptor) Validate(key string, value *ipfix.FlowProbeParams) error {
	d.log.Debug("Validate Flowprobe Params")

	if !(value.GetRecordL2() || value.GetRecordL3() || value.GetRecordL4()) {
		err := errors.New("at least one of record fields (l2, l3, l4) must be set")
		return kvs.NewInvalidValueError(err, "record_l2", "record_l3", "record_l4")
	}

	return nil
}

// Create uses vppcalls to pass Flowprobe params to VPP.
func (d *FPParamsDescriptor) Create(key string, val *ipfix.FlowProbeParams) (metadata interface{}, err error) {
	d.log.Debug("Create Flowprobe Params")
	err = d.ipfixHandler.SetFPParams(val)
	return
}

// Update uses vppcalls to pass Flowprobe params to VPP.
func (d *FPParamsDescriptor) Update(key string, oldVal, newVal *ipfix.FlowProbeParams, oldMetadata interface{}) (newMetadata interface{}, err error) {
	d.log.Debug("Update Flowprobe Params")
	err = d.ipfixHandler.SetFPParams(newVal)
	return
}

// Delete does nothing.
func (d *FPParamsDescriptor) Delete(key string, val *ipfix.FlowProbeParams, metadata interface{}) (err error) {
	d.log.Debug("Delete Flowprobe Params (nothing happens)")
	return
}
