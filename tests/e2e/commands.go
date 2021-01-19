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

package e2e

import (
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/go-errors/errors"
)

// DNSRecordType represent types of records associated with domain name in DNS server
type DNSRecordType int

const (
	A DNSRecordType = iota
	AAAA
)

var (
	dnsRecordTypeNames = map[DNSRecordType]string{
		A:    "A",
		AAAA: "AAAA",
	}
	digRecordRegexps = map[DNSRecordType]*regexp.Regexp{
		A:    regexp.MustCompile(`\tA\t([^\n]*)\n`),
		AAAA: regexp.MustCompile(`\tAAAA\t([^\n]*)\n`),
	}
	pingRegexp = regexp.MustCompile("\n([0-9]+) packets transmitted, ([0-9]+) packets received, ([0-9]+)% packet loss")
)

// Dig calls linux tool "dig" that query DNS server for domain name (queryDomain) and return records associated
// of given type (requestedInfo) associated with the domain name.
func (c *ContainerRuntime) Dig(dnsServer net.IP, queryDomain string, requestedInfo DNSRecordType) ([]net.IP, error) {
	c.ctx.t.Helper()

	// call dig in container
	args := []string{fmt.Sprintf("@%s", dnsServer), // target DNS server
		"+time=1",                               // minimize max request time (for request that don't have answer)
		"+tries=1",                              // minimize retries (try count = initial request + retries) (for request that don't have answers)
		"-t", dnsRecordTypeNames[requestedInfo], // requested record type
		queryDomain,
	}
	stdout, _, err := c.ExecCmd("dig", args...)
	if err != nil {
		return nil, errors.Errorf("execution of linux command dig failed due to: %v", err)
	}

	// parse output of dig linux command
	ipAddresses := make([]net.IP, 0)
	for _, match := range digRecordRegexps[requestedInfo].FindAllSubmatch([]byte(stdout), -1) {
		ipAddressStr := string(match[1])
		ipAddress := net.ParseIP(ipAddressStr)
		if ipAddress == nil {
			return nil, errors.Errorf("can't parse %s record value %s as ip address. Probably regular "+
				"expression matching issue for dig output:\n %s", dnsRecordTypeNames[requestedInfo],
				ipAddressStr, stdout)
		}
		c.ctx.Logger.Printf("Linux dig command got for queried domain %s an %s record %s",
			queryDomain, dnsRecordTypeNames[requestedInfo], ipAddressStr)

		ipAddresses = append(ipAddresses, ipAddress)
	}
	return ipAddresses, nil
}

// PingAsCallback can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (c *ContainerRuntime) PingAsCallback(destAddress string, opts ...PingOptModifier) func() error {
	return func() error {
		return c.Ping(destAddress, opts...)
	}
}

// Ping <destAddress> from inside of the container.
func (c *ContainerRuntime) Ping(destAddress string, opts ...PingOptModifier) error {
	c.ctx.t.Helper()

	ping := NewPingOpts(opts...)
	args := append(ping.args(), destAddress)

	stdout, _, err := c.ExecCmd("ping", args...)
	if err != nil {
		return err
	}

	matches := pingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	c.ctx.Logger.Printf("Linux ping from %s container to %s: sent=%d, received=%d, loss=%d%%",
		c.logIdentity, destAddress, sent, recv, loss)

	if sent == 0 || loss > ping.AllowedLoss {
		return fmt.Errorf("failed to ping '%s': %s", destAddress, matches[0])
	}
	return nil
}

func parsePingOutput(output string, matches []string) (sent int, recv int, loss int, err error) {
	if len(matches) != 4 {
		err = fmt.Errorf("unexpected output from ping: %s", output)
		return
	}
	sent, err = strconv.Atoi(matches[1])
	if err != nil {
		err = fmt.Errorf("failed to parse the sent packet count: %v", err)
		return
	}
	recv, err = strconv.Atoi(matches[2])
	if err != nil {
		err = fmt.Errorf("failed to parse the received packet count: %v", err)
		return
	}
	loss, err = strconv.Atoi(matches[3])
	if err != nil {
		err = fmt.Errorf("failed to parse the loss percentage: %v", err)
		return
	}
	return
}
