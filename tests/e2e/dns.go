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

// DNSServer is represents running DNS server
type DNSServer struct {
	ComponentRuntime
	ctx *TestCtx
}

// NewDNSServer creates and starts new DNS server container
func NewDNSServer(ctx *TestCtx, optMods ...DNSOptModifier) (*DNSServer, error) {
	// compute options
	opts := DefaultDNSOpt(ctx)
	for _, mod := range optMods {
		mod(opts)
	}

	// create struct for DNS server
	dnsServer := &DNSServer{
		ComponentRuntime: opts.Runtime,
		ctx:              ctx,
	}

	// get runtime specific options and start DNS server in runtime environment
	startOpts, err := opts.RuntimeStartOptions(ctx, opts)
	if err != nil {
		return nil, errors.Errorf("can't get DNSServer start option for runtime due to: %v", err)
	}
	err = dnsServer.Start(startOpts)
	if err != nil {
		return nil, errors.Errorf("can't start DNS server due to: %v", err)
	}

	return dnsServer, nil
}

func (dns *DNSServer) Stop(options ...interface{}) error {
	if err := dns.ComponentRuntime.Stop(options); err != nil {
		// not additionally cleaning up after attempting to stop test topology component because
		// it would lock access to further inspection of this component (i.e. why it won't stop)
		return err
	}
	// cleanup
	dns.ctx.DNSServer = nil
	return nil
}

// DNSServerStartOptionsForContainerRuntime translates DNSOpt to options for ComponentRuntime.Start(option)
// method implemented by ContainerRuntime
func DNSServerStartOptionsForContainerRuntime(ctx *TestCtx, options interface{}) (interface{}, error) {
	opts, ok := options.(*DNSOpt)
	if !ok {
		return nil, errors.Errorf("expected DNSOpt but got %+v", options)
	}

	// create configuration files on shared-files docker volume
	hostFileContent := fmt.Sprintf(`
# Hosts file for Domain: %s
# Place entries below in standard hosts file format: ipaddress hostname fqdn
%s
`, opts.DomainNameSuffix, opts.HostsConfig)
	hostsFilepath := CreateFileOnSharedVolume(ctx, "staticHosts", hostFileContent)
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
	coreFilepath := CreateFileOnSharedVolume(ctx, "Corefile", corefileContent)

	// construct container options
	containerOptions := &docker.CreateContainerOptions{
		Name: "e2e-test-dns",
		Config: &docker.Config{
			Image: dnsImage,
			Cmd:   []string{"-conf", coreFilepath}, // CMD adds only additional parameters to command in ENTRYPOINT
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				shareVolumeName + ":" + ctx.testShareDir, // needed for coredns configuration
			},
		},
	}

	return &ContainerStartOptions{
		ContainerOptions: containerOptions,
		Pull:             true,
	}, nil
}
