// Copyright (c) 2020 Pantheon.tech
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
	"net"

	"github.com/go-errors/errors"
	"go.ligato.io/cn-infra/v2/logging"
	scheduler "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls"
	dns "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/dns"
)

const (
	// DNSCacheDescriptorName is the name of the descriptor for VPP DNS cache server functionality
	DNSCacheDescriptorName = "vpp-dns-cache"
)

// DNSCacheDescriptor teaches KVScheduler how to configure VPP to act as DNS cache server.
type DNSCacheDescriptor struct {
	// dependencies
	log        logging.Logger
	dnsHandler vppcalls.DNSVppAPI
}

// NewDNSCacheDescriptor creates a new instance of the DNSCache descriptor.
func NewDNSCacheDescriptor(dnsHandler vppcalls.DNSVppAPI, log logging.PluginLogger) *scheduler.KVDescriptor {
	ctx := &DNSCacheDescriptor{
		log:        log.NewLogger("dnscache-descriptor"),
		dnsHandler: dnsHandler,
	}

	typedDescr := &adapter.DNSCacheDescriptor{
		Name:          DNSCacheDescriptorName,
		KeySelector:   dns.ModelDNSCache.IsKeyValid,
		ValueTypeName: dns.ModelDNSCache.ProtoName(),
		KeyLabel:      dns.ModelDNSCache.StripKeyPrefix,
		NBKeyPrefix:   dns.ModelDNSCache.KeyPrefix(),
		Validate:      ctx.ValidateDNSCache,
		Create:        ctx.Create,
		Delete:        ctx.Delete,
	}
	return adapter.NewDNSCacheDescriptor(typedDescr)
}

// ValidateDNSCache validates content of DNS cache server configuration
func (d *DNSCacheDescriptor) ValidateDNSCache(key string, dnsCache *dns.DNSCache) error {
	if len(dnsCache.UpstreamDnsServers) == 0 {
		return scheduler.NewInvalidValueError(
			errors.New("at least one upstream DNS server must be defined"), "upstreamDnsServers")
	}
	for _, serverIpAddress := range dnsCache.UpstreamDnsServers {
		if net.ParseIP(serverIpAddress) == nil {
			return scheduler.NewInvalidValueError(errors.Errorf("failed to parse upstream DNS Server IP "+
				"address %s, should be a valid ipv4/ipv6 address", serverIpAddress), "upstreamDnsServers")
		}
	}
	return nil
}

// Create enables and configures DNS functionality in VPP using VPP's binary api
func (d *DNSCacheDescriptor) Create(key string, value *dns.DNSCache) (metadata interface{}, err error) {
	for _, serverIPString := range value.UpstreamDnsServers {
		// Note: net.ParseIP should be always successful thanks to validation
		if err := d.dnsHandler.AddUpstreamDNSServer(net.ParseIP(serverIPString)); err != nil {
			return nil, errors.Errorf("can't add upstream DNS server "+
				"with IP %s due to: %v", serverIPString, err)
		}
	}
	if err := d.dnsHandler.EnableDNS(); err != nil {
		return nil, errors.Errorf("failed to enable DNS due to: %v", err)
	}
	return nil, nil
}

// Delete disables (and removes configuration) DNS functionality in VPP using VPP's binary api
func (d *DNSCacheDescriptor) Delete(key string, value *dns.DNSCache, metadata interface{}) error {
	if err := d.dnsHandler.DisableDNS(); err != nil {
		return errors.Errorf("failed to disable DNS due to: %v", err)
	}
	for _, serverIPString := range value.UpstreamDnsServers {
		// Note: net.ParseIP should be always successful thanks to validation
		if err := d.dnsHandler.DeleteUpstreamDNSServer(net.ParseIP(serverIPString)); err != nil {
			return errors.Errorf("can't remove upstream DNS server "+
				"with IP %s due to: %v", serverIPString, err)
		}
	}
	return nil
}
