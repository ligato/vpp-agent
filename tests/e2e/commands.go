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
	linuxDigRegexp = regexp.MustCompile("\n([0-9]+) packets transmitted, ([0-9]+) packets received, ([0-9]+)% packet loss")
)

// Dig calls linux tool "dig" that query DNS server for domain name (queryDomain) and return records associated
// of given type (requestedInfo) associated with the domain name.
func (c *Container) Dig(dnsServer net.IP, queryDomain string, requestedInfo DNSRecordType) ([]net.IP, error) {
	c.ctx.t.Helper()

	args := []string{fmt.Sprintf("@%s", dnsServer), "-t", dnsRecordTypeNames[requestedInfo], queryDomain}
	stdout, err := c.execCmd("dig", args...)
	if err != nil {
		return nil, errors.Errorf("execution of linux command dig failed due to: %v", err)
	}

	fmt.Print(stdout)
	// TODO parse IP addresses from answer section of dig command output and return it as result
	return nil, nil
}

type pingOptions struct {
	allowedLoss int    // percentage of allowed loss for success
	outIface    string // outgoing interface name
	maxTimeout  int    // timeout in seconds before ping exits
	count       int    // number of pings
}

func newPingOpts(opts ...pingOpt) *pingOptions {
	popts := &pingOptions{
		allowedLoss: 49, // by default at least half of the packets should get through
		maxTimeout:  4,
	}
	popts.init(opts...)
	return popts
}

func (ping *pingOptions) init(opts ...pingOpt) {
	for _, o := range opts {
		o(ping)
	}
}

func (ping *pingOptions) args() []string {
	var args []string
	if ping.maxTimeout > 0 {
		args = append(args, "-w", fmt.Sprint(ping.maxTimeout))
	}
	if ping.count > 0 {
		args = append(args, "-c", fmt.Sprint(ping.count))
	}
	if ping.outIface != "" {
		args = append(args, "-I", ping.outIface)
	}
	return args
}

type pingOpt func(opts *pingOptions)

func pingWithAllowedLoss(maxLoss int) pingOpt {
	return func(opts *pingOptions) {
		opts.allowedLoss = maxLoss
	}
}

func pingWithOutInterface(iface string) pingOpt {
	return func(opts *pingOptions) {
		opts.outIface = iface
	}
}

// ping <destAddress> from inside of the container.
func (c *Container) ping(destAddress string, opts ...pingOpt) error {
	c.ctx.t.Helper()

	ping := newPingOpts(opts...)
	args := append(ping.args(), destAddress)

	stdout, err := c.execCmd("ping", args...)
	if err != nil {
		return err
	}

	matches := linuxPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	c.ctx.logger.Printf("Linux ping from %s container to %s: sent=%d, received=%d, loss=%d%%",
		c.logIdentity, destAddress, sent, recv, loss)

	if sent == 0 || loss > ping.allowedLoss {
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
