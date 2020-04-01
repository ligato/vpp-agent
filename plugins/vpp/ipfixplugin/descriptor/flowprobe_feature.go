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
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

const (
	FPFeatureDescriptorName = "vpp-flowprobe-feature"
)

// FPFeatureDescriptor configures Flowprobe feature for VPP.
type FPFeatureDescriptor struct {
	ipfixHandler vppcalls.IpfixVppAPI
	log          logging.Logger
}

// NewFPFeatureDescriptor creates a new instance of FPFeatureDescriptor.
func NewFPFeatureDescriptor(ipfixHandler vppcalls.IpfixVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &FPFeatureDescriptor{
		ipfixHandler: ipfixHandler,
		log:          log.NewLogger("flowprobe-feature-descriptor"),
	}
	typedDescr := &adapter.FlowProbeFeatureDescriptor{
		Name:          FPFeatureDescriptorName,
		NBKeyPrefix:   ipfix.ModelFlowprobeFeature.KeyPrefix(),
		ValueTypeName: ipfix.ModelFlowprobeFeature.ProtoName(),
		KeySelector:   ipfix.ModelFlowprobeFeature.IsKeyValid,
		KeyLabel:      ipfix.ModelFlowprobeFeature.StripKeyPrefix,
		Validate:      ctx.Validate,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Dependencies:  ctx.Dependencies,
	}
	return adapter.NewFlowProbeFeatureDescriptor(typedDescr)
}

// Validate does nothing.
func (d *FPFeatureDescriptor) Validate(key string, value *ipfix.FlowProbeFeature) error {
	return nil
}

// Create uses vppcalls to pass Flowprobe feature configuration for interface to VPP.
func (d *FPFeatureDescriptor) Create(key string, val *ipfix.FlowProbeFeature) (metadata interface{}, err error) {
	err = d.ipfixHandler.AddFPFeature(val)
	return
}

// Delete uses vppcalls to remove Flowprobe feature configuration for interface..
func (d *FPFeatureDescriptor) Delete(key string, val *ipfix.FlowProbeFeature, metadata interface{}) (err error) {
	err = d.ipfixHandler.DelFPFeature(val)
	return
}

// Dependencies sets Flowprobe params as a dependency which must be created
// before enabling Flowprobe feature on an interface.
func (d *FPFeatureDescriptor) Dependencies(key string, value *ipfix.FlowProbeFeature) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: "flowprobe-params",
			Key:   ipfix.FlowprobeParamsKey(),
		},
	}
}
