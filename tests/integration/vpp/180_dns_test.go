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

package vpp

import (
	"net"
	"testing"

	"github.com/go-errors/errors"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/dnsplugin"
)

var (
	upstreamDNSServer1 = net.ParseIP("8.8.8.8") // google dns 1
	upstreamDNSServer2 = net.ParseIP("8.8.4.4") // google dns 2
	upstreamDNSServer3 = net.ParseIP("1.1.1.1") // cloudflare dns
)

// TestDNSServerCRUD test basic CRUD scenario (currently testing only call capabilities without VPP state verification)
func TestDNSServerCRUD(t *testing.T) {
	// TODO add read operation testing/assertion of VPP state - currently is missing VPP dumping functionality
	//  due to inability to find out whether DNS functionality is swithed on or off (the only way is to make
	//  DNS request, but no binary or CLI API can tell it) => testing here blindly VPP state by non-nil errors
	//  from commands
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	// ignoring unsupported VPP versions
	release := ctx.versionInfo.Release()

	// get VPP handler
	dnsHandler := vppcalls.CompatibleDNSHandler(ctx.vppClient, logrus.NewLogger("test-dns"))
	Expect(dnsHandler).ToNot(BeNil(), errors.Errorf("dns vpp handler for VPP %s is not available", release))

	// create DNS server
	Expect(dnsHandler.AddUpstreamDNSServer(upstreamDNSServer1)).To(Succeed())
	Expect(dnsHandler.AddUpstreamDNSServer(upstreamDNSServer2)).To(Succeed())
	Expect(dnsHandler.EnableDNS()).To(Succeed())

	// update upstream DNS server for DNS server configuration
	Expect(dnsHandler.DeleteUpstreamDNSServer(upstreamDNSServer1)).To(Succeed())
	Expect(dnsHandler.AddUpstreamDNSServer(upstreamDNSServer3)).To(Succeed())

	// delete DNS server and cleanup
	Expect(dnsHandler.DisableDNS()).To(Succeed())
	Expect(dnsHandler.DeleteUpstreamDNSServer(upstreamDNSServer2)).To(Succeed())
	Expect(dnsHandler.DeleteUpstreamDNSServer(upstreamDNSServer3)).To(Succeed())
}
