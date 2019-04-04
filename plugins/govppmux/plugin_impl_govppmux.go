// Copyright (c) 2017 Cisco and/or its affiliates.
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

package govppmux

import (
	"context"
	"sync"
	"time"

	"git.fd.io/govpp.git/adapter"
	govppapi "git.fd.io/govpp.git/api"
	govpp "git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/logging/measure/model/apitrace"
	"github.com/pkg/errors"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"

	_ "github.com/ligato/vpp-agent/plugins/govppmux/vppcalls/vpp1810"
	_ "github.com/ligato/vpp-agent/plugins/govppmux/vppcalls/vpp1901"
)

// Default path to socket for VPP stats
const defaultStatsSocket = "/run/vpp/stats.sock"

// Plugin implements the govppmux plugin interface.
type Plugin struct {
	Deps

	vppConn      *govpp.Connection
	vppAdapter   adapter.VppAPI
	statsConn    govppapi.StatsProvider
	statsAdapter adapter.StatsAPI
	vppConChan   chan govpp.ConnectionEvent

	lastConnErr error

	config *Config

	// Cancel can be used to cancel all goroutines and their jobs inside of the plugin.
	cancel context.CancelFunc

	// Plugin-wide tracer instance used to trace and time-measure binary API calls. Can be nil if not set.
	tracer measure.Tracer

	// Wait group allows to wait until all goroutines of the plugin have finished.
	wg sync.WaitGroup
}

// Deps groups injected dependencies of plugin
// so that they do not mix with other plugin fields.
type Deps struct {
	infra.PluginDeps
	StatusCheck statuscheck.PluginStatusWriter
	Resync      *resync.Plugin
}

// Config groups the configurable parameter of GoVpp.
type Config struct {
	TraceEnabled             bool          `json:"trace-enabled"`
	ReconnectResync          bool          `json:"resync-after-reconnect"`
	HealthCheckProbeInterval time.Duration `json:"health-check-probe-interval"`
	HealthCheckReplyTimeout  time.Duration `json:"health-check-reply-timeout"`
	HealthCheckThreshold     int           `json:"health-check-threshold"`
	ReplyTimeout             time.Duration `json:"reply-timeout"`
	// The prefix prepended to the name used for shared memory (SHM) segments. If not set,
	// shared memory segments are created directly in the SHM directory /dev/shm.
	ShmPrefix       string `json:"shm-prefix"`
	StatsSocketName string `json:"stats-socket-name"`
	// How many times can be request resent in case vpp is suddenly disconnected.
	RetryRequestCount int `json:"retry-request-count"`
	// Time between request resend attempts. Default is 500ms.
	RetryRequestTimeout time.Duration `json:"retry-request-timeout"`
	// How many times can be connection request resent in case the vpp is not reachable.
	RetryConnectCount int `json:"retry-connect-count"`
	// Time between connection request resend attempts. Default is 1s.
	RetryConnectTimeout time.Duration `json:"retry-connect-timeout"`
}

func defaultConfig() *Config {
	return &Config{
		HealthCheckProbeInterval: time.Second,
		HealthCheckReplyTimeout:  250 * time.Millisecond,
		HealthCheckThreshold:     1,
		ReplyTimeout:             time.Second,
		RetryRequestTimeout:      500 * time.Millisecond,
		RetryConnectTimeout:      time.Second,
	}
}

func (p *Plugin) loadConfig() (*Config, error) {
	cfg := defaultConfig()

	found, err := p.Cfg.LoadValue(cfg)
	if err != nil {
		return nil, err
	} else if found {
		p.Log.Debugf("config loaded from file %q", p.Cfg.GetConfigName())
	} else {
		p.Log.Debugf("config file %q not found, using default config", p.Cfg.GetConfigName())
	}

	return cfg, nil
}

// Init is the entry point called by Agent Core. A single binary-API connection to VPP is established.
func (p *Plugin) Init() error {
	var err error

	if p.config, err = p.loadConfig(); err != nil {
		return err
	}

	p.Log.Debugf("config: %+v", p.config)
	govpp.HealthCheckProbeInterval = p.config.HealthCheckProbeInterval
	govpp.HealthCheckReplyTimeout = p.config.HealthCheckReplyTimeout
	govpp.HealthCheckThreshold = p.config.HealthCheckThreshold
	govpp.DefaultReplyTimeout = p.config.ReplyTimeout
	if p.config.TraceEnabled {
		p.tracer = measure.NewTracer("govpp-mux")
		p.Log.Info("VPP API trace enabled")
	}

	if p.vppAdapter == nil {
		p.vppAdapter = NewVppAdapter(p.config.ShmPrefix)
	} else {
		// this is used for testing purposes
		p.Log.Info("Reusing existing vppAdapter")
	}

	startTime := time.Now()
	p.vppConn, p.vppConChan, err = govpp.AsyncConnect(p.vppAdapter, p.config.RetryConnectCount, p.config.RetryConnectTimeout)
	if err != nil {
		return err
	}

	// TODO: Async connect & automatic reconnect support is not yet implemented in the agent,
	// so synchronously wait until connected to VPP.
	status := <-p.vppConChan
	if status.State != govpp.Connected {
		return errors.New("unable to connect to VPP")
	}
	vppConnectTime := time.Since(startTime)

	vppAPIChan, err := p.vppConn.NewAPIChannel()
	if err != nil {
		p.Log.Error("getting new api channel failed:", err)
		return err
	}
	defer vppAPIChan.Close()

	vpeHandler := vppcalls.CompatibleVpeHandler(vppAPIChan)

	ver, err := vpeHandler.GetVersionInfo()
	if err != nil {
		p.Log.Errorf("retrieving version info failed: %v", err)
		return err
	}
	info, err := vpeHandler.GetVpeInfo()
	if err != nil {
		p.Log.Errorf("retrieving vpe info failed: %v", err)
		return err
	}
	p.Log.Infof("Connected to VPP %v [PID:%d, ClientIdx:%d] (took %s)",
		ver.Version, info.PID, info.ClientIdx, vppConnectTime.Truncate(time.Millisecond))
	p.Log.Debugf("loaded module versions: %v", info.ModuleVersions)

	// Register providing status reports (push mode)
	p.StatusCheck.Register(p.PluginName, nil)
	p.StatusCheck.ReportStateChange(p.PluginName, statuscheck.OK, nil)

	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	go p.handleVPPConnectionEvents(ctx)

	// Connect to VPP status socket
	var statsSocket string
	if p.config.StatsSocketName != "" {
		statsSocket = p.config.StatsSocketName
	} else {
		statsSocket = defaultStatsSocket
	}
	statsAdapter := NewStatsAdapter(statsSocket)
	if statsAdapter == nil {
		p.Log.Warnf("Unable to connect to the VPP statistics socket, nil stats adapter", err)
	} else if p.statsConn, err = govpp.ConnectStats(statsAdapter); err != nil {
		p.Log.Warnf("Unable to connect to the VPP statistics socket, %v", err)
		p.statsAdapter = nil
	}

	return nil
}

// Close cleans up the resources allocated by the govppmux plugin.
func (p *Plugin) Close() error {
	p.cancel()
	p.wg.Wait()

	defer func() {
		if p.vppConn != nil {
			p.vppConn.Disconnect()
		}
		if p.statsAdapter != nil {
			if err := p.statsAdapter.Disconnect(); err != nil {
				p.Log.Errorf("VPP statistics socket adapter disconnect error: %v", err)
			}
		}
	}()

	return nil
}

// NewAPIChannel returns a new API channel for communication with VPP via govpp core.
// It uses default buffer sizes for the request and reply Go channels.
//
// Example of binary API call from some plugin using GOVPP:
//      ch, _ := govpp_mux.NewAPIChannel()
//      ch.SendRequest(req).ReceiveReply
func (p *Plugin) NewAPIChannel() (govppapi.Channel, error) {
	ch, err := p.vppConn.NewAPIChannel()
	if err != nil {
		return nil, err
	}
	retryCfg := retryConfig{
		p.config.RetryRequestCount,
		p.config.RetryRequestTimeout,
	}
	return newGovppChan(ch, retryCfg, p.tracer), nil
}

// NewAPIChannelBuffered returns a new API channel for communication with VPP via govpp core.
// It allows to specify custom buffer sizes for the request and reply Go channels.
//
// Example of binary API call from some plugin using GOVPP:
//      ch, _ := govpp_mux.NewAPIChannelBuffered(100, 100)
//      ch.SendRequest(req).ReceiveReply
func (p *Plugin) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (govppapi.Channel, error) {
	ch, err := p.vppConn.NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize)
	if err != nil {
		return nil, err
	}
	retryCfg := retryConfig{
		p.config.RetryRequestCount,
		p.config.RetryRequestTimeout,
	}
	return newGovppChan(ch, retryCfg, p.tracer), nil
}

// GetTrace returns all trace entries measured so far
func (p *Plugin) GetTrace() *apitrace.Trace {
	if !p.config.TraceEnabled {
		return nil
	}
	return p.tracer.Get()
}

// ListStats returns all stats names
func (p *Plugin) ListStats(prefixes ...string) ([]string, error) {
	if p.statsAdapter == nil {
		return nil, nil
	}
	return p.statsAdapter.ListStats(prefixes...)
}

// DumpStats returns all stats with name, type and value
func (p *Plugin) DumpStats(prefixes ...string) ([]*adapter.StatEntry, error) {
	if p.statsAdapter == nil {
		return nil, nil
	}
	return p.statsAdapter.DumpStats(prefixes...)
}

// GetSystemStats retrieves system statistics of the connected VPP instance like Vector rate, Input rate, etc.
func (p *Plugin) GetSystemStats() (*govppapi.SystemStats, error) {
	if p.statsConn == nil || p.statsConn.(*govpp.StatsConnection) == nil {
		return nil, nil
	}
	return p.statsConn.GetSystemStats()
}

// GetNodeStats retrieves a list of Node VPP counters (vectors, clocks, ...)
func (p *Plugin) GetNodeStats() (*govppapi.NodeStats, error) {
	if p.statsConn == nil || p.statsConn.(*govpp.StatsConnection) == nil {
		return nil, nil
	}
	return p.statsConn.GetNodeStats()
}

// GetInterfaceStats retrieves all counters related to the VPP interfaces
func (p *Plugin) GetInterfaceStats() (*govppapi.InterfaceStats, error) {
	if p.statsConn == nil || p.statsConn.(*govpp.StatsConnection) == nil {
		return nil, nil
	}
	return p.statsConn.GetInterfaceStats()
}

// GetErrorStats retrieves VPP error counters
func (p *Plugin) GetErrorStats(names ...string) (*govppapi.ErrorStats, error) {
	if p.statsConn == nil || p.statsConn.(*govpp.StatsConnection) == nil {
		return nil, nil
	}
	return p.statsConn.GetErrorStats()
}

// handleVPPConnectionEvents handles VPP connection events.
func (p *Plugin) handleVPPConnectionEvents(ctx context.Context) {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case status := <-p.vppConChan:
			if status.State == govpp.Connected {
				vppAPIChan, err := p.vppConn.NewAPIChannel()
				if err != nil {
					p.Log.Error("getting new api channel failed:", err)
					return
				}
				defer vppAPIChan.Close()

				vpeHandler := vppcalls.CompatibleVpeHandler(vppAPIChan)
				vpeHandler.GetVersionInfo()
				vpeHandler.GetVpeInfo()
				if p.config.ReconnectResync && p.lastConnErr != nil {
					p.Log.Info("Starting resync after VPP reconnect")
					if p.Resync != nil {
						p.Resync.DoResync()
						p.lastConnErr = nil
					} else {
						p.Log.Warn("Expected resync after VPP reconnect could not start because of missing Resync plugin")
					}
				}
				p.StatusCheck.ReportStateChange(p.PluginName, statuscheck.OK, nil)
			} else {
				p.lastConnErr = errors.Errorf("VPP disconnected (status: %v)", status)
				p.StatusCheck.ReportStateChange(p.PluginName, statuscheck.Error, p.lastConnErr)
			}

		case <-ctx.Done():
			return
		}
	}
}
