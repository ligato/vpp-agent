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

//go:generate descriptor-adapter --descriptor-name LocalSID --value-type *vpp_srv6.LocalSID --import "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/srv6" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name Policy --value-type *vpp_srv6.Policy --import "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/srv6" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name Steering --value-type *vpp_srv6.Steering --import "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/srv6" --output-dir "descriptor"

package srplugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v2/plugins/govppmux"
	scheduler "go.ligato.io/vpp-agent/v2/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/descriptor"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/vppcalls/vpp2001"
	_ "go.ligato.io/vpp-agent/v2/plugins/vpp/srplugin/vppcalls/vpp2001_324"
)

// SRPlugin configures segment routing.
type SRPlugin struct {
	Deps

	// GoVPP channels
	vppCh govppapi.Channel

	// VPP handler
	srHandler vppcalls.SRv6VppAPI

	// descriptors
	localSIDDescriptor *descriptor.LocalSIDDescriptor
	policyDescriptor   *descriptor.PolicyDescriptor
	steeringDescriptor *descriptor.SteeringDescriptor
}

// Deps lists dependencies of the interface p.
type Deps struct {
	infra.PluginDeps
	Scheduler   scheduler.KVScheduler
	GoVppmux    govppmux.API
	IfPlugin    ifplugin.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init initializes and registers descriptors for Linux ARPs and Routes.
func (p *SRPlugin) Init() error {
	var err error

	// GoVPP channels
	if p.vppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API channel: %v", err)
	}

	// init handlers
	p.srHandler = vppcalls.CompatibleSRv6VppHandler(p.vppCh, p.IfPlugin.GetInterfaceIndex(), p.Log)

	// init & register descriptors
	localSIDDescriptor := descriptor.NewLocalSIDDescriptor(p.srHandler, p.Log)
	policyDescriptor := descriptor.NewPolicyDescriptor(p.srHandler, p.Log)
	steeringDescriptor := descriptor.NewSteeringDescriptor(p.srHandler, p.Log)

	err = p.Deps.Scheduler.RegisterKVDescriptor(
		localSIDDescriptor,
		policyDescriptor,
		steeringDescriptor,
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
