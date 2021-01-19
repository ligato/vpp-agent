//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package e2e_test

import (
	"fmt"
	"net"
	"testing"

	docker "github.com/fsouza/go-dockerclient"

	. "github.com/onsi/gomega"
	. "go.ligato.io/vpp-agent/v3/tests/e2e"
)

func TestDnsCache(t *testing.T) {
	dnsIP4Result := net.ParseIP("10.100.0.1")
	dnsIP6Result := net.ParseIP("fc::1") // fc::/7 is ipv6 private range (like 10.0.0.0/8 for ipv4)

	cases := []struct {
		Name                        string
		PublicUpstreamDNSServer     net.IP
		QueryDomainName             string
		SetupModifiers              []SetupOptModifier
		ExpectedResolvedIPv4Address net.IP
		ExpectedResolvedIPv6Address net.IP
		SkipAAAARecordCheck         bool
		SkipAll                     bool
	}{
		{
			Name:                    "Test VPP DNS Cache with google DNS as upstream DNS server",
			PublicUpstreamDNSServer: net.ParseIP("8.8.8.8"),
			QueryDomainName:         "www.google.com",
			SkipAAAARecordCheck:     true, // TODO add VPP Jira task reference
		}, {
			Name:                    "Test VPP DNS Cache with cloudflare DNS as upstream DNS server",
			PublicUpstreamDNSServer: net.ParseIP("1.1.1.1"),
			QueryDomainName:         "www.google.com",
			SkipAll:                 true, // TODO add VPP Jira task reference
		}, {
			Name:            "Test VPP DNS Cache with coredns container as upstream DNS server",
			QueryDomainName: "dnscache." + LigatoDNSHostNameSuffix,
			SetupModifiers: []SetupOptModifier{
				WithDNSServer(WithZonedStaticEntries(LigatoDNSHostNameSuffix,
					fmt.Sprintf("%s %s", dnsIP4Result, "dnscache."+LigatoDNSHostNameSuffix),
					fmt.Sprintf("%s %s", dnsIP6Result, "dnscache."+LigatoDNSHostNameSuffix),
				)),
			},
			ExpectedResolvedIPv4Address: dnsIP4Result,
			ExpectedResolvedIPv6Address: dnsIP6Result,
			SkipAll:                     true, // TODO add VPP Jira task reference
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			if td.SkipAll {
				t.Skipf("Skipped due to VPP bugs")
			}
			td.SetupModifiers = append(td.SetupModifiers, WithCustomVPPAgent()) // need iptables to be installed
			ctx := Setup(t, td.SetupModifiers...)
			defer ctx.Teardown()

			ms := ctx.StartMicroservice("microservice1", useMicroserviceWithDig())

			vppDNSServer := net.ParseIP(ctx.Agent.IPAddress())
			Expect(vppDNSServer).ShouldNot(BeNil(), "VPP DNS Server container has no IP address")
			upstreamDNSServer := td.PublicUpstreamDNSServer
			if td.PublicUpstreamDNSServer == nil {
				upstreamDNSServer := net.ParseIP(ctx.DNSServer.IPAddress())
				Expect(upstreamDNSServer).ShouldNot(BeNil(),
					"Local upstream DNS Server container (CoreDNS) has no IP address")
			}

			ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "bin", "dns_name_server_add_del", upstreamDNSServer.String()) //"8.8.8.8") //upstreamDNSServer.String()) //TODO debug
			//ctx.Agent.ExecCmd("vppctl", "-s", ":5002", " bin dns_name_server_add_del "2001:4860:4860::8888"
			ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "bin", "dns_enable_disable")
			//ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "dns", "cache add test.ligato.io", "1.2.3.4")

			ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "create", "tap")
			ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "set", "interface", "state", "tap0", "up")
			ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "set", "interface", "ip", "addr", "tap0", "10.10.0.2/24")

			ctx.Agent.ExecCmd("ip", "addr", "add", "10.10.0.3/24", "dev", "tap0")
			ctx.Agent.ExecCmd("ip", "link", "set", "tap0", "up")

			ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "show", "dns", "cache", "verbose")
			ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "show", "dns", "servers")

			// routing all packets out of vpp using tap0 (requesting upstream DNS server + responding to clients of VPP in role of DNS server)
			//ctx.Agent.ExecCmd("vppctl", "-s", ":5002", " ip route add ${DNS_SERVER}/32 via 10.10.0.3
			ctx.Agent.ExecCmd("vppctl", "-s", ":5002", "ip", "route", "add", "0.0.0.0/0", "via", "10.10.0.3")

			// NAT container IP to/from DNS service(tap IP address on VPP end side) (this could be also done by using iproute2: http://linux-ip.net/html/nat-dnat.html//ex-nat-dnat-full)
			//// NAT for communication with client of VPP DNS server and VPP DNS server
			ctx.Agent.ExecCmd("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "tcp", "-d", vppDNSServer.String(), "--dport", "53", "-j", "DNAT", "--to-destination", "10.10.0.2")
			ctx.Agent.ExecCmd("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "-d", vppDNSServer.String(), "--dport", "53", "-j", "DNAT", "--to-destination", "10.10.0.2")
			ctx.Agent.ExecCmd("iptables", "-t", "nat", "-A", "POSTROUTING", "-p", "tcp", "-s", "10.10.0.2", "--sport", "53", "-j", "SNAT", "--to-source", vppDNSServer.String())
			ctx.Agent.ExecCmd("iptables", "-t", "nat", "-A", "POSTROUTING", "-p", "udp", "-s", "10.10.0.2", "--sport", "53", "-j", "SNAT", "--to-source", vppDNSServer.String())
			//// NAT for communication with upstream DNS server
			ctx.Agent.ExecCmd("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "tcp", "-d", vppDNSServer.String(), "--dport", "53053", "-j", "DNAT", "--to-destination", "10.10.0.2")
			ctx.Agent.ExecCmd("iptables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "-d", vppDNSServer.String(), "--dport", "53053", "-j", "DNAT", "--to-destination", "10.10.0.2")
			ctx.Agent.ExecCmd("iptables", "-t", "nat", "-A", "POSTROUTING", "-p", "tcp", "-s", "10.10.0.2", "--sport", "53053", "-j", "SNAT", "--to-source", vppDNSServer.String())
			ctx.Agent.ExecCmd("iptables", "-t", "nat", "-A", "POSTROUTING", "-p", "udp", "-s", "10.10.0.2", "--sport", "53053", "-j", "SNAT", "--to-source", vppDNSServer.String())

			// Testing resolving DNS query by VPP (it should make request to upstream DNS server)
			resolvedIPAddresses, err := ms.Dig(vppDNSServer, td.QueryDomainName, A)
			Expect(err).ToNot(HaveOccurred())
			if td.ExpectedResolvedIPv4Address != nil {
				Expect(resolvedIPAddresses).To(ConsistOf([]net.IP{td.ExpectedResolvedIPv4Address}))
			} else { // external domain have loadbalancers -> can't tell what IP address we get from upstream DNS server
				Expect(resolvedIPAddresses).NotTo(BeEmpty())
				Expect(resolvedIPAddresses[0].To4() != nil).Should(BeTrue(), "is not ipv4 address")
			}

			// doesn't work due to VPP bug (TODO add link to VPP jira bug task)
			if !td.SkipAAAARecordCheck {
				resolvedIPAddresses, err = ms.Dig(vppDNSServer, td.QueryDomainName, AAAA)
				Expect(err).ToNot(HaveOccurred())
				if td.ExpectedResolvedIPv6Address != nil {
					Expect(resolvedIPAddresses).To(ConsistOf([]net.IP{td.ExpectedResolvedIPv6Address}))
				} else { // external domain have loadbalancers -> can't tell what IP address we get from upstream DNS server
					Expect(resolvedIPAddresses).NotTo(BeEmpty())
					Expect(resolvedIPAddresses[0].To4() == nil).Should(BeTrue(), "is not ipv6 address")
				}
			}

			// block additional request to upstream DNS server (VPP should have the info
			// already cached and should not need the upstream DNS server anymore)
			if td.PublicUpstreamDNSServer != nil {
				// block request to upstream DNS server
				ctx.Agent.ExecCmd("iptables", "-A", "OUTPUT", "-j", "DROP", "-d", upstreamDNSServer.String())
			} else {
				// using local container as DNS server -> the easy way how to block it is to kill it
				ctx.TerminateDNSServer() // TODO change call to ctx.DNSServer.Stop()
			}

			// Testing resolving DNS query by VPP from its cache (upstream DNS server requests are blocked)
			resolvedIPAddresses, err = ms.Dig(vppDNSServer, td.QueryDomainName, A)
			Expect(err).ToNot(HaveOccurred())
			if td.ExpectedResolvedIPv4Address != nil {
				Expect(resolvedIPAddresses).To(ConsistOf([]net.IP{td.ExpectedResolvedIPv4Address}))
			} else { // external domain have loadbalancers -> can't tell what IP address we get from upstream DNS server
				Expect(resolvedIPAddresses).NotTo(BeEmpty())
				Expect(resolvedIPAddresses[0].To4() != nil).Should(BeTrue(), "is not ipv4 address")
			}

			// doesn't work due to VPP bug (TODO add link to VPP jira bug task)
			if !td.SkipAAAARecordCheck {
				resolvedIPAddresses, err = ms.Dig(vppDNSServer, td.QueryDomainName, AAAA)
				Expect(err).ToNot(HaveOccurred())
				if td.ExpectedResolvedIPv6Address != nil {
					Expect(resolvedIPAddresses).To(ConsistOf([]net.IP{td.ExpectedResolvedIPv6Address}))
				} else { // external domain have loadbalancers -> can't tell what IP address we get from upstream DNS server
					Expect(resolvedIPAddresses).NotTo(BeEmpty())
					Expect(resolvedIPAddresses[0].To4() == nil).Should(BeTrue(), "is not ipv6 address")
				}
			}
		})
	}
}

func useMicroserviceWithDig() MicroserviceOptModifier {
	return WithMSContainerStartHook(func(opts *docker.CreateContainerOptions) {
		// use different image (+ entrypoint usage in image needs changes in container start)
		opts.Config.Image = "itsthenetwork/alpine-dig"
		opts.Config.Entrypoint = []string{"tail", "-f", "/dev/null"}
		opts.Config.Cmd = []string{}
		opts.HostConfig.NetworkMode = "bridge" // return back to default docker networking
	})
}
