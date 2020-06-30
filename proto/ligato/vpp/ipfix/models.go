//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp_ipfix

import (
	"go.ligato.io/vpp-agent/v3/pkg/models"
)

// ModuleName is the module name used for models.
const ModuleName = "vpp.ipfix"

var (
	ModelIPFIX = models.Register(&IPFIX{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "ipfix",
	})

	ModelFlowprobeParams = models.Register(&FlowProbeParams{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "flowprobe-params",
	})

	ModelFlowprobeFeature = models.Register(&FlowProbeFeature{}, models.Spec{
		Module:  ModuleName,
		Version: "v2",
		Type:    "flowprobe-feature",
	}, models.WithNameTemplate("{{.Interface}}"))
)

// IPFIXKey returns the prefix used in ETCD to store vpp IPFIX config.
func IPFIXKey() string {
	return models.Key(&IPFIX{})
}

// FlowprobeParamsKey returns the prefix used in ETCD
// to store vpp Flowprobe params config.
func FlowprobeParamsKey() string {
	return models.Key(&FlowProbeParams{})
}

// FlowprobeFeatureKey returns the prefix used in ETCD
// to store vpp Flowprobe feature config.
func FlowprobeFeatureKey(iface string) string {
	return models.Key(&FlowProbeFeature{Interface: iface})
}
