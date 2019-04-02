//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package telemetry

import (
	"sync"
	"time"

	"github.com/ligato/cn-infra/infra"
	prom "github.com/ligato/cn-infra/rpc/prometheus"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/telemetry/vppcalls"

	_ "github.com/ligato/vpp-agent/plugins/telemetry/vppcalls/vpp1810"
	_ "github.com/ligato/vpp-agent/plugins/telemetry/vppcalls/vpp1901"
)

const (
	// default period between updates
	defaultUpdatePeriod = time.Second * 30
	// minimum period between updates
	minimumUpdatePeriod = time.Second * 5
)

// Plugin registers Telemetry Plugin
type Plugin struct {
	Deps

	handler vppcalls.TelemetryVppAPI

	prometheusMetrics

	// From config file
	updatePeriod time.Duration
	disabled     bool

	wg   sync.WaitGroup
	quit chan struct{}
}

// Deps represents dependencies of Telemetry Plugin
type Deps struct {
	infra.PluginDeps
	ServiceLabel servicelabel.ReaderAPI
	GoVppmux     govppmux.API
	Prometheus   prom.API
}

type runtimeStats struct {
	threadName string
	threadID   uint
	itemName   string
	metrics    map[string]prometheus.Gauge
}

type memoryStats struct {
	threadName string
	threadID   uint
	metrics    map[string]prometheus.Gauge
}

type buffersStats struct {
	threadID  uint
	itemName  string
	itemIndex uint
	metrics   map[string]prometheus.Gauge
}

type nodeCounterStats struct {
	itemName string
	metrics  map[string]prometheus.Gauge
}

// Init initializes Telemetry Plugin
func (p *Plugin) Init() error {
	p.quit = make(chan struct{})

	// Telemetry config file
	config, err := p.loadConfig()
	if err != nil {
		return err
	}
	if config != nil {
		// If telemetry is not enabled, skip plugin initialization
		if config.Disabled {
			p.Log.Info("Telemetry plugin disabled via config file")
			p.disabled = true
			return nil
		}
		// This prevents setting the update period to less than 5 seconds,
		// which can have significant performance hit.
		if config.PollingInterval > minimumUpdatePeriod {
			p.updatePeriod = config.PollingInterval
			p.Log.Infof("polling period changed to %v", p.updatePeriod)
		} else if config.PollingInterval > 0 {
			p.Log.Warnf("polling period has to be at least %s, using default: %v",
				minimumUpdatePeriod, defaultUpdatePeriod)
		}
	}
	// This serves as fallback if the config was not found or if the value is not set in config.
	if p.updatePeriod == 0 {
		p.updatePeriod = defaultUpdatePeriod
	}

	if err := p.registerPrometheus(); err != nil {
		return err
	}

	return nil
}

// AfterInit executes after initializion of Telemetry Plugin
func (p *Plugin) AfterInit() error {
	// Do not start polling if telemetry is disabled
	if p.disabled {
		return nil
	}

	p.wg.Add(1)
	go p.periodicUpdates()

	return nil
}

// Close is used to clean up resources used by Telemetry Plugin
func (p *Plugin) Close() error {
	close(p.quit)
	p.wg.Wait()
	return nil
}

// periodic updates for the metrics data
func (p *Plugin) periodicUpdates() {
	defer p.wg.Done()

	// Create GoVPP channel
	vppCh, err := p.GoVppmux.NewAPIChannel()
	if err != nil {
		p.Log.Errorf("creating channel failed: %v", err)
		return
	}
	defer vppCh.Close()

	p.handler = vppcalls.CompatibleTelemetryHandler(vppCh)

	p.Log.Debugf("starting periodic updates (%v)", p.updatePeriod)

	for {
		select {
		// Delay period between updates
		case <-time.After(p.updatePeriod):
			p.updatePrometheus()

		// Plugin has stopped.
		case <-p.quit:
			p.Log.Debugf("stopping periodic updates")
			return
		}
	}
}
