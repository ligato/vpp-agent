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

package e2e

import (
	"fmt"
	"testing"
)

func TestDnsCache(t *testing.T) {
	dnsIP4Result := "10.100.0.1"
	dnsIP6Result := "fc::1" // fc::/7 is ipv6 private range (like 10.0.0.0/8 for ipv4)
	ctx := Setup(t, WithDNSServer(WithZonedStaticEntries(LigatoDNSHostNameSuffix,
		fmt.Sprintf("%s %s.%s", dnsIP4Result, "dnscache", LigatoDNSHostNameSuffix),
		fmt.Sprintf("%s %s.%s", dnsIP6Result, "dnscache", LigatoDNSHostNameSuffix),
	)))
	defer ctx.Teardown()

	//time.Sleep(100000 * time.Hour)
	// TODO configure VPP as DNS server(with coredns server in container as its upstream DNS server that VPP
	//  consults for records that it doesn't know nothing about) and use dig linux command to check its functionality
}
