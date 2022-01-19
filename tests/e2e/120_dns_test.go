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
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	. "github.com/onsi/gomega"

	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_iptables "go.ligato.io/vpp-agent/v3/proto/ligato/linux/iptables"
	vpp_dns "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/dns"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	. "go.ligato.io/vpp-agent/v3/tests/e2e"
)

// TestDnsCache tests ability of VPP to act as DNS server with cache capabilities (cache info from upstream DNS server)
func TestDnsCache(t *testing.T) {
	dnsIP4Result := net.ParseIP("10.100.0.1")
	dnsIP6Result := net.ParseIP("fc::1") // fc::/7 is ipv6 private range (like 10.0.0.0/8 for ipv4)

	cases := []struct {
		Name                                 string
		PublicUpstreamDNSServer              net.IP
		QueryDomainName                      string
		UnreachabilityVerificationDomainName string
		SetupModifiers                       []SetupOptModifier
		ExpectedResolvedIPv4Address          net.IP
		ExpectedResolvedIPv6Address          net.IP
		SkipAAAARecordCheck                  bool
		SkipAll                              bool
		SkipReason                           string
	}{
		{
			Name:                                 "Test VPP DNS Cache with google DNS as upstream DNS server",
			PublicUpstreamDNSServer:              net.ParseIP("8.8.8.8"),
			QueryDomainName:                      "www.google.com",
			UnreachabilityVerificationDomainName: "www.sme.sk",
			SkipAAAARecordCheck:                  true, // TODO remove skipping when VPP bug resolved
			SkipReason:                           "VPP bug https://jira.fd.io/browse/VPP-1963",
		}, {
			Name:                                 "Test VPP DNS Cache with cloudflare DNS as upstream DNS server",
			PublicUpstreamDNSServer:              net.ParseIP("1.1.1.1"),
			QueryDomainName:                      "www.google.com",
			UnreachabilityVerificationDomainName: "ubuntu.com",
			SkipAll:                              true, // TODO remove skipping when VPP bug resolved
			SkipReason:                           "VPP bug https://jira.fd.io/browse/VPP-1963",
		}, {
			Name:                                 "Test VPP DNS Cache with coredns container as upstream DNS server",
			QueryDomainName:                      "dnscache." + LigatoDNSHostNameSuffix,
			UnreachabilityVerificationDomainName: "unresolvable." + LigatoDNSHostNameSuffix,
			SetupModifiers: []SetupOptModifier{
				WithDNSServer(WithZonedStaticEntries(LigatoDNSHostNameSuffix,
					fmt.Sprintf("%s %s", dnsIP4Result, "dnscache."+LigatoDNSHostNameSuffix),
					fmt.Sprintf("%s %s", dnsIP6Result, "dnscache."+LigatoDNSHostNameSuffix),
				)),
			},
			ExpectedResolvedIPv4Address: dnsIP4Result,
			ExpectedResolvedIPv6Address: dnsIP6Result,
			SkipAll:                     true, // TODO remove skipping when VPP bug resolved
			SkipReason:                  "VPP bug https://jira.fd.io/browse/VPP-1963",
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			if td.SkipAll {
				t.Skipf("Skipped due to %s", td.SkipReason)
			}

			// test setup
			td.SetupModifiers = append(td.SetupModifiers, WithCustomVPPAgent()) // need iptables to be installed
			ctx := Setup(t, td.SetupModifiers...)
			defer ctx.Teardown()

			// start microservice
			ms := ctx.StartMicroservice("microservice1", useMicroserviceWithDig())

			// configure VPP-Agent container as DNS server
			vppDNSServer := net.ParseIP(ctx.Agent.IPAddress())
			ctx.Expect(vppDNSServer).ShouldNot(BeNil(), "VPP DNS Server container has no IP address")
			upstreamDNSServer := td.PublicUpstreamDNSServer
			if td.PublicUpstreamDNSServer == nil {
				upstreamDNSServer = net.ParseIP(ctx.DNSServer.IPAddress())
				ctx.Expect(upstreamDNSServer).ShouldNot(BeNil(),
					"Local upstream DNS Server container (CoreDNS) has no IP address")
			}
			err := configureVPPAgentAsDNSServer(ctx, vppDNSServer, upstreamDNSServer)
			ctx.Expect(err).ShouldNot(HaveOccurred(), "Configuring changes to VPP-Agent failed with err")

			// Testing resolving DNS query by VPP (it should make request to upstream DNS server)
			//// Testing A (IPv4) record
			resolvedIPAddresses, err := ms.Dig(vppDNSServer, td.QueryDomainName, A)
			ctx.Expect(err).ToNot(HaveOccurred())
			if td.ExpectedResolvedIPv4Address != nil {
				ctx.Expect(resolvedIPAddresses).To(ConsistOf([]net.IP{td.ExpectedResolvedIPv4Address}))
			} else { // external domain have loadbalancers -> can't tell what IP address we get from upstream DNS server
				ctx.Expect(resolvedIPAddresses).NotTo(BeEmpty())
				ctx.Expect(resolvedIPAddresses[0].To4() != nil).Should(BeTrue(), "is not ipv4 address")
			}

			//// Testing AAAA (IPv6) record
			if !td.SkipAAAARecordCheck {
				resolvedIPAddresses, err = ms.Dig(vppDNSServer, td.QueryDomainName, AAAA)
				ctx.Expect(err).ToNot(HaveOccurred())
				if td.ExpectedResolvedIPv6Address != nil {
					ctx.Expect(resolvedIPAddresses).To(ConsistOf([]net.IP{td.ExpectedResolvedIPv6Address}))
				} else { // external domain have loadbalancers -> can't tell what IP address we get from upstream DNS server
					ctx.Expect(resolvedIPAddresses).NotTo(BeEmpty())
					ctx.Expect(resolvedIPAddresses[0].To4() == nil).Should(BeTrue(), "is not ipv6 address")
				}
			}

			// block additional request to upstream DNS server (VPP should have the info
			// already cached and should not need the upstream DNS server anymore)
			if td.PublicUpstreamDNSServer != nil {
				// block request to upstream DNS server
				blockUpstreamDNSServer := &linux_iptables.RuleChain{
					Name:      "blockUpstreamDNSServer",
					Protocol:  linux_iptables.RuleChain_IPV4,
					Table:     linux_iptables.RuleChain_FILTER,
					ChainType: linux_iptables.RuleChain_FORWARD,
					Rules: []string{
						fmt.Sprintf("-j DROP -d %s", upstreamDNSServer.String()),
					},
				}
				ctx.Expect(ctx.GenericClient().ChangeRequest().Update(blockUpstreamDNSServer).
					Send(context.Background())).To(Succeed())
			} else {
				// using local container as DNS server -> the easy way how to block it is to kill it
				ctx.DNSServer.Stop()
			}

			// verify the upstream DNS blocking (without it some mild test/container changes could introduce
			// silent error when VPP will still access upstream DNS server due to ineffective(or still not
			// effectively applied) blocking and therefore test of VPP caching will not test the cache at all)
			ctx.Eventually(func() error {
				_, err := ms.Dig(vppDNSServer, td.UnreachabilityVerificationDomainName, A)
				return err
			}, 3*time.Second).Should(HaveOccurred(),
				"The connection to upstream DNS server is still not severed.")

			// Testing resolving DNS query by VPP from its cache (upstream DNS server requests are blocked)
			//// Testing A (IPv4) record
			resolvedIPAddresses, err = ms.Dig(vppDNSServer, td.QueryDomainName, A)
			ctx.Expect(err).ToNot(HaveOccurred())
			if td.ExpectedResolvedIPv4Address != nil {
				ctx.Expect(resolvedIPAddresses).To(ConsistOf([]net.IP{td.ExpectedResolvedIPv4Address}))
			} else { // external domain have loadbalancers -> can't tell what IP address we get from upstream DNS server
				ctx.Expect(resolvedIPAddresses).NotTo(BeEmpty())
				ctx.Expect(resolvedIPAddresses[0].To4() != nil).Should(BeTrue(), "is not ipv4 address")
			}

			//// Testing AAAA (IPv6) record
			if !td.SkipAAAARecordCheck {
				resolvedIPAddresses, err = ms.Dig(vppDNSServer, td.QueryDomainName, AAAA)
				ctx.Expect(err).ToNot(HaveOccurred())
				if td.ExpectedResolvedIPv6Address != nil {
					ctx.Expect(resolvedIPAddresses).To(ConsistOf([]net.IP{td.ExpectedResolvedIPv6Address}))
				} else { // external domain have loadbalancers -> can't tell what IP address we get from upstream DNS server
					ctx.Expect(resolvedIPAddresses).NotTo(BeEmpty())
					ctx.Expect(resolvedIPAddresses[0].To4() == nil).Should(BeTrue(), "is not ipv6 address")
				}
			}
		})
	}
}

// configureVPPAgentAsDNSServer configures VPP-Agent container to act as DNS server.
func configureVPPAgentAsDNSServer(ctx *TestCtx, vppDNSServer, upstreamDNSServer net.IP) error {
	/* VPP-Agent as DNS Server topology:

	   +--------------------------------------------+
	   |                VPP-Agent container         |
	   |                                            |
	   |   +----------------+  +----------------+   |
	   |   |    VPP-Agent   |  |      VPP       |   |
	   |   |                |  |                |   |
	   |   |                |  |                |   |
	   |   |                |  |       + vppTap |   |
	   |   |                |  |       |        |   |
	   |   +------+---------+  +----------------+   |
	   |          |                    |            |
	   |          |                    + linuxTap   |
	   |          |                    |            |
	   |   +------------------------------------+   |
	   |   |      |   Linux Kernel     |        |   |
	   |   |  +---+--------------------+------+ |   |
	   |   |  |            NAT                | |   |
	   |   |  +--------------+----------------+ |   |
	   |   |                 |                  |   |
	   |   +------------------------------------+   |
	   |                     |                      |
	   +--------------------------------------------+
	                         |
	                         + Default container interface
	*/

	const (
		vppTapName         = "tap1"
		linuxTapName       = "tap1-host"
		vppTapIP           = "10.10.0.2"
		linuxTapIP         = "10.10.0.3"
		vppTapIPWithMask   = vppTapIP + "/24"
		linuxTapIPWithMask = linuxTapIP + "/24"
	)

	// configure VPP to act as DNS cache server
	dnsCacheServer := &vpp_dns.DNSCache{
		UpstreamDnsServers: []string{upstreamDNSServer.String()},
	}

	// tap tunnel from VPP to container linux environment
	vppTap := &vpp_interfaces.Interface{
		Name:        vppTapName,
		Type:        vpp_interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{vppTapIPWithMask},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:    2,
				HostIfName: linuxTapName,
				//ToMicroservice: MsNamePrefix + msName,
			},
		},
	}
	linuxTap := &linux_interfaces.Interface{
		Name:        linuxTapName,
		Type:        linux_interfaces.Interface_TAP_TO_VPP,
		Enabled:     true,
		HostIfName:  linuxTapName,
		IpAddresses: []string{linuxTapIPWithMask},
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: vppTapName,
			},
		},
	}

	// routing all packets out of vpp using tap tunnel
	// (requesting upstream DNS server + responding to clients of VPP in role of DNS server)
	vppRouteOut := &vpp_l3.Route{
		DstNetwork:        "0.0.0.0/0",
		NextHopAddr:       linuxTapIP,
		OutgoingInterface: vppTapName,
	}

	// NAT translation so that VPP-Agent container IP address with standard DNS port (53) acts as DNS Server service.
	// The VPP handles DNS packets from its inner tap tunnel end, but packets arrive at default container interface.
	// So additional path to join these 2 places is done (tap + linux routing), but that is not enough due to
	// using container ip address as DNS Server service address. The DNS packet must be forwarded to VPP, but it
	// thinks that it has already arrived where it is supposed to be (external IP of container), so the destination
	// address must be translated to arrive in VPP. This happens by configuring linux kernel's prerouting chain of
	// NAT table. The answer to DNS request must be also translated (changed source IP) by using postrouting chain
	// of NAT table. The DNS traffic should normally stay only on port 53, but VPP contacts upstream DNS server
	// (to consult unknown DNS domain names) with source port 53053 and that makes trouble when answer return
	// from these upstream DNS servers. Hence, the NAT also for 53053 port.
	vppDNSCommunicationPrerouting := &linux_iptables.RuleChain{
		Name:      "VPPDNSCommunicationPrerouting",
		Protocol:  linux_iptables.RuleChain_IPV4,
		Table:     linux_iptables.RuleChain_NAT,
		ChainType: linux_iptables.RuleChain_PREROUTING,
		Rules: []string{
			// NAT for communication between client and VPP DNS server
			fmt.Sprintf("-p tcp -d %s --dport 53 -j DNAT --to-destination %s", vppDNSServer.String(), vppTapIP),
			fmt.Sprintf("-p udp -d %s --dport 53 -j DNAT --to-destination %s", vppDNSServer.String(), vppTapIP),
			// NAT for communication with upstream DNS server
			fmt.Sprintf("-p tcp -d %s --dport 53053 -j DNAT --to-destination %s", vppDNSServer.String(), vppTapIP),
			fmt.Sprintf("-p udp -d %s --dport 53053 -j DNAT --to-destination %s", vppDNSServer.String(), vppTapIP),
		},
	}
	vppDNSCommunicationPostrouting := &linux_iptables.RuleChain{
		Name:      "VPPDNSCommunicationPostrouting",
		Protocol:  linux_iptables.RuleChain_IPV4,
		Table:     linux_iptables.RuleChain_NAT,
		ChainType: linux_iptables.RuleChain_POSTROUTING,
		Rules: []string{
			// NAT for communication between client and VPP DNS server
			fmt.Sprintf("-p tcp -s %s --sport 53 -j SNAT --to-source %s", vppTapIP, vppDNSServer.String()),
			fmt.Sprintf("-p udp -s %s --sport 53 -j SNAT --to-source %s", vppTapIP, vppDNSServer.String()),
			// NAT for communication with upstream DNS server
			fmt.Sprintf("-p tcp -s %s --sport 53053 -j SNAT --to-source %s", vppTapIP, vppDNSServer.String()),
			fmt.Sprintf("-p udp -s %s --sport 53053 -j SNAT --to-source %s", vppTapIP, vppDNSServer.String()),
		},
	}

	// apply the configuration
	req := ctx.GenericClient().ChangeRequest()
	err := req.Update(
		dnsCacheServer,
		vppTap,
		linuxTap,
		vppRouteOut,
		vppDNSCommunicationPrerouting,
		vppDNSCommunicationPostrouting,
	).Send(context.Background())
	return err
}

// useMicroserviceWithDig provides modifier for using specialized microservice image for using linux dig tool
func useMicroserviceWithDig() MicroserviceOptModifier {
	return WithMSContainerStartHook(func(opts *docker.CreateContainerOptions) {
		// use different image (+ entrypoint usage in image needs changes in container start)
		opts.Config.Image = "itsthenetwork/alpine-dig"
		opts.Config.Entrypoint = []string{"tail", "-f", "/dev/null"}
		opts.Config.Cmd = []string{}
		opts.HostConfig.NetworkMode = "bridge" // return back to default docker networking
	})
}
