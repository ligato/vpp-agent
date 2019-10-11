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
	"expvar"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	prom "github.com/ligato/cn-infra/rpc/prometheus"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/unrolled/render"

	"github.com/ligato/vpp-agent/api/configurator"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/telemetry/vppcalls"

	_ "github.com/ligato/vpp-agent/plugins/telemetry/vppcalls/vpp1904"
	_ "github.com/ligato/vpp-agent/plugins/telemetry/vppcalls/vpp1908"
	_ "github.com/ligato/vpp-agent/plugins/telemetry/vppcalls/vpp2001_324"
	_ "github.com/ligato/vpp-agent/plugins/telemetry/vppcalls/vpp2001_379"
)

var debug = os.Getenv("DEBUG_TELEMETRY") != ""

// Plugin registers Telemetry Plugin
type Plugin struct {
	Deps

	handler vppcalls.TelemetryVppAPI

	statsPollerServer
	prometheusMetrics

	// From config file
	updatePeriod time.Duration
	disabled     bool
	skipped      map[string]bool

	wg   sync.WaitGroup
	quit chan struct{}
}

// Deps represents dependencies of Telemetry Plugin
type Deps struct {
	infra.PluginDeps
	ServiceLabel servicelabel.ReaderAPI
	GoVppmux     govppmux.StatsAPI
	Prometheus   prom.API
	GRPC         grpc.Server
	HTTPHandlers rest.HTTPHandlers
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
	// This serves as fallback if the config was not found or if the value is not set in config.
	if p.updatePeriod == 0 {
		p.updatePeriod = defaultUpdatePeriod
	}

	if err := p.registerPrometheus(); err != nil {
		return err
	}

	p.statsPollerServer.log = p.Log.NewLogger("stats-poller")
	if err := p.setupStatsPoller(); err != nil {
		return err
	}

	if p.HTTPHandlers != nil {
		p.HTTPHandlers.RegisterHTTPHandler("/metrics/{metric}", metricsHandler, "GET")
	}

	return nil
}

// AfterInit executes after initializion of Telemetry Plugin
func (p *Plugin) AfterInit() error {
	// Do not start polling if telemetry is disabled
	if p.disabled {
		return nil
	}

	p.startPeriodicUpdates()

	return nil
}

func (p *Plugin) setupStatsPoller() error {
	vppCh, err := p.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	defer vppCh.Close()

	h := vppcalls.CompatibleTelemetryHandler(vppCh, p.GoVppmux)
	if h == nil {
		return err
	}
	p.statsPollerServer.handler = h

	if p.GRPC != nil && p.GRPC.GetServer() != nil {
		configurator.RegisterStatsPollerServer(p.GRPC.GetServer(), &p.statsPollerServer)
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
	vppCh, err := p.GoVppmux.NewAPIChannel()
	if err != nil {
		p.Log.Errorf("creating channel failed: %v", err)
		return
	}
	defer vppCh.Close()

	p.handler = vppcalls.CompatibleTelemetryHandler(vppCh, p.GoVppmux)
	if p.handler == nil {
		p.Log.Warnf("no compatible telemetry handler, skipping periodic updates")
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
			formatter.JSON(w, http.StatusNotFound, struct{}{})
			return
		}
		metric := vars["metric"]
		var data proto.Message
		expvar.Do(func(kv expvar.KeyValue) {
			if data != nil {
				return
			}
			if kv.Key == metric {
				if fn, ok := kv.Value.(expvar.Func); ok {
					data = fn.Value().(proto.Message)
					return
				}
			}
		})
		if data == nil {
			formatter.JSON(w, http.StatusNotFound, struct{}{})
			return
		}
		formatter.JSON(w, 200, data)
	}
}
