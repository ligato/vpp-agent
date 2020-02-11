// Copyright (c) 2019 Bell Canada, Pantheon Technologies and/or its affiliates.
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

//go:generate descriptor-adapter --descriptor-name LocalSID --value-type *vpp_srv6.LocalSID --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name Policy --value-type *vpp_srv6.Policy --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name Steering --value-type *vpp_srv6.Steering --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name SRv6Global --value-type *vpp_srv6.SRv6Global --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6" --output-dir "descriptor"

package srplugin

import (
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	scheduler "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls/vpp2001"
)

// SRPlugin configures segment routing.
type SRPlugin struct {
	Deps

	// VPP handler
	srHandler vppcalls.SRv6VppAPI

	// descriptors
	localSIDDescriptor *descriptor.LocalSIDDescriptor
	policyDescriptor   *descriptor.PolicyDescriptor
	steeringDescriptor *descriptor.SteeringDescriptor
}

type Deps struct {
	infra.PluginDeps
	Scheduler   scheduler.KVScheduler
	VPP         govppmux.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init initializes and registers descriptors for Linux ARPs and Routes.
func (p *SRPlugin) Init() error {
	var err error

	// init handlers
	p.srHandler = vppcalls.CompatibleSRv6Handler(p.VPP, p.IfPlugin.GetInterfaceIndex(), p.Log)

	// init & register descriptors
	localSIDDescriptor := descriptor.NewLocalSIDDescriptor(p.srHandler, p.Log)
	policyDescriptor := descriptor.NewPolicyDescriptor(p.srHandler, p.Log)
	steeringDescriptor := descriptor.NewSteeringDescriptor(p.srHandler, p.Log)
	encapSourceAddressDescriptor := descriptor.NewSRv6GlobalDescriptor(p.srHandler, p.Log)

	err = p.Deps.Scheduler.RegisterKVDescriptor(
		localSIDDescriptor,
		policyDescriptor,
		steeringDescriptor,
		encapSourceAddressDescriptor,
	)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *SRPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}
