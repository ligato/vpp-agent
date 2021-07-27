//  Copyright (c) 2021 Cisco and/or its affiliates.
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

//go:generate descriptor-adapter --descriptor-name IPFIX  --value-type *vpp_ipfix.IPFIX --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name FlowProbeFeature  --value-type *vpp_ipfix.FlowProbeFeature --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name FlowProbeParams  --value-type *vpp_ipfix.FlowProbeParams --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix" --output-dir "descriptor"

package ipfixplugin

import (
	"errors"

	"go.ligato.io/cn-infra/v2/infra"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls/vpp2005"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls/vpp2009"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls/vpp2101"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls/vpp2106"
)

// IPFIXPlugin is a plugin that manages IPFIX configuration in VPP.
// IPFIX - IP Flow Information Export (IPFIX).
// It allows to:
//   - configure export of Flowprobe information;
//   - configure Flowprobe Params;
//   - enable/disable Flowprobe Feature for an interface.
//
// Things to rememmber:
//   - Flowprobe Feature can not be configured for any interface,
//     if Flowprobe Params were not set.
//   - Flowprobe Params can not be changed,
//     if Flowprobe Feature was enabled for at least one interface.
type IPFIXPlugin struct {
	Deps

	// VPP handler.
	ipfixHandler vppcalls.IpfixVppAPI
}

// Deps represents dependencies for the plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
	VPP         govppmux.API
	IfPlugin    ifplugin.API
}

// Init initializes IPFIX plugin.
func (p *IPFIXPlugin) Init() (err error) {
	// Even with IPFIX being part of a core of VPP, without Flowprobe plugin
	// there would be no information to export, hence no point in this plugin.
	if !p.VPP.IsPluginLoaded("flowprobe") {
		p.Log.Warnf("VPP plugin Flowprobe was disabled by VPP")
		return nil
	}

	p.ipfixHandler = vppcalls.CompatibleIpfixVppHandler(p.VPP, p.IfPlugin.GetInterfaceIndex(), p.Log)
	if p.ipfixHandler == nil {
		return errors.New("IPFIX VPP handler is not available")
	}

	ipfixDescriptor := descriptor.NewIPFIXDescriptor(p.ipfixHandler, p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(ipfixDescriptor)
	if err != nil {
		return err
	}

	fpFeatureDescriptor := descriptor.NewFPFeatureDescriptor(p.ipfixHandler, p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(fpFeatureDescriptor)
	if err != nil {
		return err
	}

	// Descriptor for Flowprobe Params will use `fpFeatureMM` to check
	// if Flowprobe Params can be updated. If at least one item is in this
	// map, than there is at least one interface with Flowprobe Feature
	// enabled, hence Flowprobe Params update is not allowed.
	fpFeatureMM := p.KVScheduler.GetMetadataMap(fpFeatureDescriptor.Name)

	fpParamsDescriptor := descriptor.NewFPParamsDescriptor(p.ipfixHandler, fpFeatureMM, p.Log)
	err = p.KVScheduler.RegisterKVDescriptor(fpParamsDescriptor)
	if err != nil {
		return err
	}

	return nil
}
