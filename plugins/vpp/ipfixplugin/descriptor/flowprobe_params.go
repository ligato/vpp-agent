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

	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

const (
	// FPParamsDescriptorName is the name of the descriptor for
	// VPP Flowprobe Params configuration.
	FPParamsDescriptorName = "vpp-flowprobe-params"
)

var (
	// Validation errors:

	// ErrAllRecordFieldsDisabled returned when all record fields are set to false.
	ErrAllRecordFieldsDisabled = errors.New("at least one of record fields (l2, l3, l4) must be set")

	// Non-retriable errors:

	// ErrFeatureEnabled informs the reason why Flowprobe Params can not be changed.
	ErrFeatureEnabled = errors.New("can not change Flowprobe Params when Flowprobe Feature enabled on some interface")
)

// FPParamsDescriptor configures Flowprobe Params for VPP.
type FPParamsDescriptor struct {
	ipfixHandler vppcalls.IpfixVppAPI
	featureMap   idxmap.NamedMapping
	log          logging.Logger
}

// NewFPParamsDescriptor creates a new instance of FPParamsDescriptor.
func NewFPParamsDescriptor(ipfixHandler vppcalls.IpfixVppAPI, featureMap idxmap.NamedMapping, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &FPParamsDescriptor{
		ipfixHandler: ipfixHandler,
		featureMap:   featureMap,
		log:          log.NewLogger("flowprobe-params-descriptor"),
	}
	typedDescr := &adapter.FlowProbeParamsDescriptor{
		Name:               FPParamsDescriptorName,
		NBKeyPrefix:        ipfix.ModelFlowprobeParams.KeyPrefix(),
		ValueTypeName:      ipfix.ModelFlowprobeParams.ProtoName(),
		KeySelector:        ipfix.ModelFlowprobeParams.IsKeyValid,
		KeyLabel:           ipfix.ModelFlowprobeParams.StripKeyPrefix,
		IsRetriableFailure: ctx.IsRetriableFailure,
		Validate:           ctx.Validate,
		Create:             ctx.Create,
		Delete:             ctx.Delete,
		Update:             ctx.Update,
		Retrieve:           ctx.Retrieve,
	}
	return adapter.NewFlowProbeParamsDescriptor(typedDescr)
}

// Validate checks if Flowprobe Params are good to send to VPP.
func (d *FPParamsDescriptor) Validate(key string, value *ipfix.FlowProbeParams) error {
	if !(value.GetRecordL2() || value.GetRecordL3() || value.GetRecordL4()) {
		return kvs.NewInvalidValueError(ErrAllRecordFieldsDisabled,
			"record_l2", "record_l3", "record_l4")
	}

	return nil
}

// IsRetriableFailure returns false if error is one of errors
// defined at the top of this file as non-retriable.
func (d *FPParamsDescriptor) IsRetriableFailure(err error) bool {
	if errors.Is(err, ErrFeatureEnabled) {
		return false
	}
	return true
}

// Create passes Flowprobe Params to Update method.
func (d *FPParamsDescriptor) Create(key string, val *ipfix.FlowProbeParams) (metadata interface{}, err error) {
	_, err = d.Update(key, nil, val, nil)
	return
}

// Update uses vppcalls to pass Flowprobe Params to VPP.
func (d *FPParamsDescriptor) Update(key string, oldVal, newVal *ipfix.FlowProbeParams, oldMetadata interface{}) (newMetadata interface{}, err error) {
	// Check if there is at least one Flowporbe Feature configured on some interface.
	if len(d.featureMap.ListAllNames()) > 0 {
		err = ErrFeatureEnabled
		return
	}
	err = d.ipfixHandler.SetFPParams(newVal)
	return
}

// Delete does nothing.
//
// Since all Flowprobe Features are dependent on Flowprobe Params,
// calling this method will also disable (move to "pending" state)
// all Flowprobe Features on interfaces.
//
// All the work will be done by KVScheduler :)
func (d *FPParamsDescriptor) Delete(key string, val *ipfix.FlowProbeParams, metadata interface{}) (err error) {
	return
}

// Retrieve hopes that configuration in correlate is actual configuration in VPP.
// As soon as VPP devs will add dump API calls, this methods should be fixed.
func (d *FPParamsDescriptor) Retrieve(correlate []adapter.FlowProbeParamsKVWithMetadata) (retrieved []adapter.FlowProbeParamsKVWithMetadata, err error) {
	return correlate, nil
}
