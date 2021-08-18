// Copyright (c) 2021 Pantheon.tech
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

//go:generate descriptor-adapter --descriptor-name DNSCache --value-type *vpp_dns.DNSCache --import "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/dns" --output-dir "descriptor"

package dnsplugin

import (
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	scheduler "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls/vpp2005"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls/vpp2009"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls/vpp2101"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls/vpp2106"
)

// DNSPlugin configures VPP ability to act as DNS cache server.
type DNSPlugin struct {
	Deps

	dnsHandler    vppcalls.DNSVppAPI
	dnsDescriptor *descriptor.DNSCacheDescriptor
}

type Deps struct {
	infra.PluginDeps
	Scheduler   scheduler.KVScheduler
	VPP         govppmux.API
	StatusCheck statuscheck.PluginStatusWriter // optional
}

// Init initializes and registers descriptor for DNS.
func (p *DNSPlugin) Init() error {
	// init handler
	p.dnsHandler = vppcalls.CompatibleDNSHandler(p.VPP, p.Log)

	// init & register descriptor
	dnsDescriptor := descriptor.NewDNSCacheDescriptor(p.dnsHandler, p.Log)
	if err := p.Deps.Scheduler.RegisterKVDescriptor(dnsDescriptor); err != nil {
		return err
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *DNSPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}
