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
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

const (
	// FPFeatureDescriptorName is the name of the descriptor for
	// VPP Flowprobe Feature configuration.
	FPFeatureDescriptorName = "vpp-flowprobe-feature"
)

// Validation errors:
var (
	// ErrIfaceNotDefined returned when interface in confiugration is empty string.
	ErrIfaceNotDefined = errors.New("missing interface name for Flowprobe Feature")
)

// FPFeatureDescriptor configures Flowprobe Feature for VPP.
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
		WithMetadata:  true,
		Validate:      ctx.Validate,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Retrieve:      ctx.Retrieve,
		Dependencies:  ctx.Dependencies,
	}
	return adapter.NewFlowProbeFeatureDescriptor(typedDescr)
}

// Validate checks if Flowprobe Feature configuration is good to send to VPP.
func (d *FPFeatureDescriptor) Validate(key string, value *ipfix.FlowProbeFeature) error {
	if value.GetInterface() == "" {
		return kvs.NewInvalidValueError(ErrIfaceNotDefined, "interface")
	}
	return nil
}

// Create uses vppcalls to pass Flowprobe Feature configuration for interface to VPP.
func (d *FPFeatureDescriptor) Create(key string, val *ipfix.FlowProbeFeature) (metadata interface{}, err error) {
	err = d.ipfixHandler.AddFPFeature(val)
	return val, err
}

// Delete uses vppcalls to remove Flowprobe Feature configuration for interface.
func (d *FPFeatureDescriptor) Delete(key string, val *ipfix.FlowProbeFeature, metadata interface{}) (err error) {
	err = d.ipfixHandler.DelFPFeature(val)
	return
}

// Dependencies sets Flowprobe Params as a dependency which must be created
// before enabling Flowprobe Feature on an interface.
func (d *FPFeatureDescriptor) Dependencies(key string, val *ipfix.FlowProbeFeature) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: "flowprobe-params",
			Key:   ipfix.FlowprobeParamsKey(),
		},
		{
			Label: "interface",
			Key:   interfaces.InterfaceKey(val.Interface),
		},
	}
}

// Retrieve hopes that configuration in correlate is actual configuration in VPP.
// As soon as VPP devs will add dump API calls, this methods should be fixed.
// TODO: waiting for https://jira.fd.io/browse/VPP-1861.
//
// Also, this method sets metadata, so descriptor for Flowprobe Params would know
// that there are some interfaces with Flowprobe Feature enabled.
func (d *FPFeatureDescriptor) Retrieve(correlate []adapter.FlowProbeFeatureKVWithMetadata) (retrieved []adapter.FlowProbeFeatureKVWithMetadata, err error) {
	corr := make([]adapter.FlowProbeFeatureKVWithMetadata, len(correlate))
	for i, c := range correlate {
		corr[i] = adapter.FlowProbeFeatureKVWithMetadata{
			Key:      c.Key,
			Value:    c.Value,
			Metadata: c.Value,
			Origin:   c.Origin,
		}
	}
	return corr, nil
}
