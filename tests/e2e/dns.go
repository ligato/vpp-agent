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

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-errors/errors"
)

const (
	LigatoDNSHostNameSuffix = "test.ligato.io"
	dnsImage                = "coredns/coredns:1.8.0"
	dnsStopTimeout          = 1 // seconds
)

// DNSContainer is represents running DNS server container
type DNSContainer struct {
	*Container
}

// NewDNSContainer creates and starts new DNS server container
func NewDNSContainer(ctx *TestCtx, options ...DNSOptModifier) (*DNSContainer, error) {
	c := &DNSContainer{
		&Container{
			ctx:         ctx,
			logIdentity: "DNS server",
			stopTimeout: dnsStopTimeout,
		},
	}
	_, err := c.create(options...)
	if err != nil {
		return nil, errors.Errorf("can't create %s container due to: %v", c.logIdentity, err)
	}
	if err := c.start(); err != nil {
		return nil, errors.Errorf("can't start %s container due to: %v", c.logIdentity, err)
	}
	return c, nil
}

func (c *DNSContainer) create(options ...DNSOptModifier) (*docker.Container, error) {
	opts := DefaultDNSOpt()
	for _, optionModifier := range options {
		optionModifier(opts)
	}

	// create configuration files on shared-files docker volume
	hostFileContent := fmt.Sprintf(`
# Hosts file for Domain: %s
# Place entries below in standard hosts file format: ipaddress hostname fqdn
%s
`, opts.DomainNameSuffix, opts.HostsConfig)
	hostsFilepath := CreateFileOnSharedVolume(c.ctx, "staticHosts", hostFileContent)
	corefileContent := fmt.Sprintf(`%s {
    log
    errors
    hosts %s %s
}
. {
	# everything else will be resolved by external public DNS server
    log
    errors
    forward . 8.8.8.8:53
}
`, opts.DomainNameSuffix, hostsFilepath, opts.DomainNameSuffix)
	coreFilepath := CreateFileOnSharedVolume(c.ctx, "Corefile", corefileContent)

	// construct container options
	containerOptions := &docker.CreateContainerOptions{
		Name: "e2e-test-dns",
		Config: &docker.Config{
			Image: dnsImage,
			Cmd:   []string{"-conf", coreFilepath}, // CMD adds only additional parameters to command in ENTRYPOINT
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				shareVolumeName + ":" + c.ctx.testShareDir, // needed for coredns configuration
			},
		},
	}

	return c.Container.create(containerOptions, true)
}
