// Copyright (c) 2018 Cisco and/or its affiliates.
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

package l3plugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
)

// IPNeighConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 IP scan neighbor, as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/ipneigh". Configuration uses single key, since
// the configuration is global-like.
type IPNeighConfigurator struct {
	log logging.Logger

	// VPP channel
	vppChan govppapi.Channel
	// VPP API channel
	ipNeighHandler vppcalls.IPNeighVppAPI

	// Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// Init VPP channel and vppcalls handler
func (p *IPNeighConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, enableStopwatch bool) (err error) {
	// Logger
	p.log = logger.NewLogger("-l3-ip-neigh-conf")
	p.log.Debugf("Initializing proxy ARP configurator")

	// Configurator-wide stopwatch instance
	if enableStopwatch {
		p.stopwatch = measure.NewStopwatch("IPScan-Neigh-configurator", p.log)
	}

	// VPP channel
	p.vppChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// VPP API handler
	p.ipNeighHandler = vppcalls.NewIPNeighVppHandler(p.vppChan, p.log, p.stopwatch)

	return nil
}

// Close VPP channel
func (p *IPNeighConfigurator) Close() error {
	return safeclose.Close(p.vppChan)
}

// Set puts desired IP scan neighbor configuration to the VPP
func (p *IPNeighConfigurator) Set(config *l3.IPScanNeighbor) error {
	if err := p.ipNeighHandler.SetIPScanNeighbor(config); err != nil {
		return err
	}

	p.log.Debugf("IP scan neighbor set to %v", config.Mode)

	return nil
}

// Unset returns IP scan neighbor configuration to default
func (p *IPNeighConfigurator) Unset() error {
	defaultCfg := &l3.IPScanNeighbor{
		Mode: l3.IPScanNeighbor_DISABLED,
	}

	if err := p.ipNeighHandler.SetIPScanNeighbor(defaultCfg); err != nil {
		return err
	}

	p.log.Debug("IP scan neighbor set to default")

	return nil
}
