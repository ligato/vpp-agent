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

package vppcalls

import (
	"net"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
)

// DNSVppAPI is API boundary for vppcall package access, introduced to properly test code dependent on vppcalls package
type DNSVppAPI interface {
	DNSVPPWrite
	DNSVPPRead
}

// DNSVPPWrite provides write methods for DNS cache server functionality
type DNSVPPWrite interface {
	// EnableDNS make act VPP as DNS cache server
	EnableDNS() error

	// DisableDNS disables functionality that makes VPP act as DNS cache server
	DisableDNS() error

	// AddUpstreamDNSServer adds new upstream DNS Server to the upstream DNS server list
	AddUpstreamDNSServer(serverIPAddress net.IP) error

	// DeleteUpstreamDNSServer removes upstream DNS Server from the upstream DNS server list
	DeleteUpstreamDNSServer(serverIPAddress net.IP) error
}

// DNSVPPRead provides read methods for DNS cache server functionality
type DNSVPPRead interface {
	// TODO check whether dump can be implemented(vppcalls + descriptor's retrieve) (currently there is
	//   no dump binapi or VPP CLI check whether the functionality is enabled - dns cache is not good indicator
	//   because it can't detect feature disabling after first enabling)
	// DumpDNSCache retrieves DNSCache if DNS cache server functionality is enabled, otherwise it returns nil
	// DumpDNSCache() (dnsCache *dns.DNSCache, err error)
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "dns",
	HandlerAPI: (*DNSVppAPI)(nil),
})

type NewHandlerFunc func(vpp.Client, logging.Logger) DNSVppAPI

func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			return c.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			return h(c, a[0].(logging.Logger))
		},
	})
}

func CompatibleDNSHandler(c vpp.Client, log logging.Logger) DNSVppAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c, log).(DNSVppAPI)
	}
	return nil
}
