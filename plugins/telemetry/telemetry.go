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
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	prom "github.com/ligato/cn-infra/rpc/prometheus"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/pkg/errors"
	"github.com/unrolled/render"

	"go.ligato.io/vpp-agent/v3/pkg/metrics"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"

	_ "go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls/vpp2001"
)

var debug = os.Getenv("DEBUG_TELEMETRY") != ""

// Plugin registers Telemetry Plugin
type Plugin struct {
	Deps

	handler vppcalls.TelemetryVppAPI

	statsPollerServer
	prometheusMetrics

	// From config file
	updatePeriod       time.Duration
	disabled           bool
	prometheusDisabled bool
	skipped            map[string]bool

	wg   sync.WaitGroup
	quit chan struct{}
}

type InterfaceIndexProvider interface {
	// GetInterfaceIndex gives read-only access to map with metadata of all configured
	// VPP interfaces.
	GetInterfaceIndex() ifaceidx.IfaceMetadataIndex
}

// Deps represents dependencies of Telemetry Plugin
type Deps struct {
	infra.PluginDeps
	ServiceLabel servicelabel.ReaderAPI
	VPP          govppmux.API
	Prometheus   prom.API
	GRPC         grpc.Server
	HTTPHandlers rest.HTTPHandlers
	IfPlugin     InterfaceIndexProvider
}

// Init initializes Telemetry Plugin
func (p *Plugin) Init() error {
	p.quit = make(chan struct{})
	p.skipped = make(map[string]bool, 0)

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
		// Disable prometheus metrics if set by config
		if config.PrometheusDisabled {
			p.Log.Info("Prometheus metrics disabled via config file")
			p.prometheusDisabled = true
		} else {
			// This prevents setting the update period to less than 5 seconds,
			// which can have significant performance hit.
			if config.PollingInterval > minimumUpdatePeriod {
				p.updatePeriod = config.PollingInterval
				p.Log.Infof("polling period changed to %v", p.updatePeriod)
			} else if config.PollingInterval > 0 {
				p.Log.Warnf("polling period has to be at least %s, using default: %v",
					minimumUpdatePeriod, defaultUpdatePeriod)
			}
			// Store map of skipped metrics
			for _, skip := range config.Skipped {
				p.skipped[skip] = true
			}
		}
	}

	// Register prometheus
	if !p.prometheusDisabled {
		if p.updatePeriod == 0 {
			p.updatePeriod = defaultUpdatePeriod
		}
		if err := p.registerPrometheus(); err != nil {
			return err
		}
	}

	// Setup stats poller
	p.statsPollerServer.log = p.Log.NewLogger("stats-poller")
	if err := p.setupStatsPoller(); err != nil {
		return errors.WithMessage(err, "setting up stats poller failed")
	}

	if p.HTTPHandlers != nil {
		p.HTTPHandlers.RegisterHTTPHandler("/metrics/{metric}", metricsHandler, "GET")
	}

	return nil
}

// AfterInit executes after initializion of Telemetry Plugin
func (p *Plugin) AfterInit() error {
	// Do not start polling if telemetry is disabled
	if p.disabled || p.prometheusDisabled {
		return nil
	}

	p.startPeriodicUpdates()

	return nil
}

func (p *Plugin) setupStatsPoller() error {
	h := vppcalls.CompatibleTelemetryHandler(p.VPP)
	if h == nil {
		return fmt.Errorf("VPP telemetry handler unavailable")
	}
	p.statsPollerServer.handler = h
	p.ifIndex = p.IfPlugin.GetInterfaceIndex()

	if p.GRPC != nil && p.GRPC.GetServer() != nil {
		configurator.RegisterStatsPollerServiceServer(p.GRPC.GetServer(), &p.statsPollerServer)
	}

	return nil
}

// Close is used to clean up resources used by Telemetry Plugin
func (p *Plugin) Close() error {
	close(p.quit)
	p.wg.Wait()
	return nil
}

func (p *Plugin) startPeriodicUpdates() {
	p.handler = vppcalls.CompatibleTelemetryHandler(p.VPP)
	if p.handler == nil {
		p.Log.Warnf("VPP telemetry handler unavailable, skipping periodic updates")
		return
	}

	p.wg.Add(1)
	go p.periodicUpdates()
}

// periodic updates for the metrics data
func (p *Plugin) periodicUpdates() {
	defer p.wg.Done()

	p.Log.Debugf("starting periodic updates (%v)", p.updatePeriod)

	tick := time.NewTicker(p.updatePeriod)
	for {
		select {
		// Delay period between updates
		case <-tick.C:
			ctx := context.Background()
			p.updatePrometheus(ctx)

		// Plugin has stopped.
		case <-p.quit:
			p.Log.Debugf("stopping periodic updates")
			return
		}
	}
}

func (p *Plugin) tracef(f string, a ...interface{}) {
	if debug && p.Log.GetLevel() >= logging.DebugLevel {
		s := fmt.Sprintf(f, a...)
		if len(s) > 250 {
			p.Log.Debugf("%s... (%d bytes omitted) ...%s", s[:200], len(s)-250, s[len(s)-50:])
			return
		}
		p.Log.Debug(s)
	}
}

func metricsHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		if vars == nil {
			_ = formatter.JSON(w, http.StatusNotFound, struct{}{})
			return
		}
		metric := vars["metric"]
		model, err := models.DefaultRegistry.GetModel(metric)
		if err != nil {
			_ = formatter.JSON(w, http.StatusNotFound, struct{ Error string }{err.Error()})
			return
		}
		data := model.NewInstance()
		if err := metrics.Retrieve(data); err != nil {
			_ = formatter.JSON(w, http.StatusInternalServerError, struct{ Error string }{err.Error()})
			return
		}
		_ = formatter.JSON(w, 200, data)
	}
}
