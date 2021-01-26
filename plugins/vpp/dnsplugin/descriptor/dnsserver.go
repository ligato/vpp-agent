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
	// DNSServerDescriptorName is the name of the descriptor for VPP DNS server functionality
	DNSServerDescriptorName = "vpp-dns-server"
)

// DNSServerDescriptor teaches KVScheduler how to configure VPP to act as DNS server.
type DNSServerDescriptor struct {
	// dependencies
	log        logging.Logger
	dnsHandler vppcalls.DNSVppAPI
}

// NewDNServerDescriptor creates a new instance of the DNSServer descriptor.
func NewDNServerDescriptor(dnsHandler vppcalls.DNSVppAPI, log logging.PluginLogger) *scheduler.KVDescriptor {
	ctx := &DNSServerDescriptor{
		log:        log.NewLogger("dnsserver-descriptor"),
		dnsHandler: dnsHandler,
	}

	typedDescr := &adapter.DNSServerDescriptor{
		Name:            DNSServerDescriptorName,
		KeySelector:     dns.ModelDNSServer.IsKeyValid,
		ValueTypeName:   dns.ModelDNSServer.ProtoName(),
		KeyLabel:        dns.ModelDNSServer.StripKeyPrefix,
		NBKeyPrefix:     dns.ModelDNSServer.KeyPrefix(),
		Validate:        ctx.ValidateDNSServers,
		ValueComparator: ctx.EquivalentDNSServers,
		Create:          ctx.Create,
		Delete:          ctx.Delete,
		Update:          ctx.Update,
	}
	return adapter.NewDNSServerDescriptor(typedDescr)
}

// ValidateDNSServers validates content of DNS server configuration
func (d *DNSServerDescriptor) ValidateDNSServers(key string, dnsServer *dns.DNSServer) error {
	if len(dnsServer.UpstreamDnsServers) == 0 {
		return scheduler.NewInvalidValueError(
			errors.New("at least one upstream DNS server must be defined"), "upstreamDnsServers")
	}
	for _, serverIpAddress := range dnsServer.UpstreamDnsServers {
		if net.ParseIP(serverIpAddress) == nil {
			return scheduler.NewInvalidValueError(errors.Errorf("failed to parse upstream DNS Server IP "+
				"address %s, should be a valid ipv4/ipv6 address", serverIpAddress), "upstreamDnsServers")
		}
	}
	return nil
}

// EquivalentDNSServers determines whether 2 DNS servers are logically equal. This comparison takes
// into consideration also semantics that couldn't be modeled into proto models (i.e. upstream servers are
// IP addresses and not only strings)
func (d *DNSServerDescriptor) EquivalentDNSServers(key string, oldDNSServer, newDNSServer *dns.DNSServer) bool {
	if (oldDNSServer.UpstreamDnsServers == nil) != (newDNSServer.UpstreamDnsServers == nil) {
		return false
	}
	if len(oldDNSServer.UpstreamDnsServers) != len(newDNSServer.UpstreamDnsServers) {
		return false
	}

	for i := range newDNSServer.UpstreamDnsServers {
		if newDNSServer.UpstreamDnsServers[i] != oldDNSServer.UpstreamDnsServers[i] {
			return false
		}
	}
	return true
}

// Create enables and configures DNS functionality in VPP using VPP's binary api
func (d *DNSServerDescriptor) Create(key string, value *dns.DNSServer) (metadata interface{}, err error) {
	if err := d.updateUpstreamDNSServerList(nil, value.UpstreamDnsServers); err != nil {
		return nil, err
	}
	if err := d.dnsHandler.EnableDNS(); err != nil {
		return nil, errors.Errorf("failed to enable DNS due to: %v", err)
	}
	return nil, nil
}

// Delete disables (and removes configuration) DNS functionality in VPP using VPP's binary api
func (d *DNSServerDescriptor) Delete(key string, value *dns.DNSServer, metadata interface{}) error {
	if err := d.dnsHandler.DisableDNS(); err != nil {
		return errors.Errorf("failed to disable DNS due to: %v", err)
	}
	if err := d.updateUpstreamDNSServerList(value.UpstreamDnsServers, nil); err != nil {
		return err
	}
	return nil
}

// Update updates DNS Server configuration of already enabled DNS functionality
func (d *DNSServerDescriptor) Update(key string, oldValue, newValue *dns.DNSServer,
	oldMetadata interface{}) (newMetadata interface{}, err error) {
	if err := d.updateUpstreamDNSServerList(oldValue.UpstreamDnsServers, newValue.UpstreamDnsServers); err != nil {
		return nil, errors.Errorf("can't update upstream DNS servers due to: %v", err)
	}
	return nil, nil
}

func (d *DNSServerDescriptor) updateUpstreamDNSServerList(oldServers, newServers []string) error {
	// no insertion into existing upstream DNS Server list, just adding at the end of list  -> remove all
	// and then add new servers one by one
	if oldServers != nil {
		for _, serverIPString := range oldServers {
			// Note: net.ParseIP should be always successful thanks to validation
			if err := d.dnsHandler.DeleteUpstreamDNSServer(net.ParseIP(serverIPString)); err != nil {
				return errors.Errorf("can't remove upstream DNS server "+
					"with IP %s due to: %v", serverIPString, err)
			}
		}
	}
	if newServers != nil {
		for _, serverIPString := range newServers {
			// Note: net.ParseIP should be always successful thanks to validation
			if err := d.dnsHandler.AddUpstreamDNSServer(net.ParseIP(serverIPString)); err != nil {
				return errors.Errorf("can't add upstream DNS server "+
					"with IP %s due to: %v", serverIPString, err)
			}
		}
	}
	return nil
}
