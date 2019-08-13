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
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"git.fd.io/govpp.git/adapter"
	govppapi "git.fd.io/govpp.git/api"
	govpp "git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/pkg/errors"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"

	_ "github.com/ligato/vpp-agent/plugins/govppmux/vppcalls/vpp1901"
	_ "github.com/ligato/vpp-agent/plugins/govppmux/vppcalls/vpp1904"
	_ "github.com/ligato/vpp-agent/plugins/govppmux/vppcalls/vpp1908"
)

var (
	disabledSocketClient = os.Getenv("GOVPPMUX_NOSOCK") != ""
)

// Plugin implements the govppmux plugin interface.
type Plugin struct {
	Deps

	vppConn      *govpp.Connection
	vppAdapter   adapter.VppAPI
	statsConn    govppapi.StatsProvider
	statsAdapter adapter.StatsAPI
	vppConChan   chan govpp.ConnectionEvent
	lastConnErr  error

	// mu protects fields below (vppInfo, lastEvent)
	infoMu    sync.Mutex
	vppInfo   VPPInfo
	lastEvent govpp.ConnectionEvent

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
	HTTPHandlers rest.HTTPHandlers
	StatusCheck  statuscheck.PluginStatusWriter
	Resync       *resync.Plugin
}

// Init is the entry point called by Agent Core. A single binary-API connection to VPP is established.
func (p *Plugin) Init() (err error) {
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

	// register REST API handlers
	p.registerHandlers(p.HTTPHandlers)

	if p.vppAdapter == nil {
		var address string
		useShm := disabledSocketClient || p.config.ConnectViaShm || p.config.ShmPrefix != ""
		if useShm {
			address = p.config.ShmPrefix
		} else {
			address = p.config.BinAPISocketPath
		}
		p.vppAdapter = NewVppAdapter(address, useShm)
	} else {
		// this is used for testing purposes
		p.Log.Info("Reusing existing vppAdapter")
	}

	// TODO: Async connect & automatic reconnect support is not yet implemented in the agent,
	// so synchronously wait until connected to VPP.
	startTime := time.Now()
	p.Log.Debugf("connecting to VPP..")

	p.vppConn, p.vppConChan, err = govpp.AsyncConnect(p.vppAdapter, p.config.RetryConnectCount, p.config.RetryConnectTimeout)
	if err != nil {
		return err
	}

	// wait for connection event
	for {
		event := <-p.vppConChan
		if event.State == govpp.Connected {
			break
		} else if event.State == govpp.Failed || event.State == govpp.Disconnected {
			return errors.Errorf("unable to establish connection to VPP (%v)", event.Error)
		} else {
			p.Log.Debugf("VPP connection state: %+v", event)
		}
	}

	connectDur := time.Since(startTime)
	p.Log.Debugf("connection to VPP established (took %s)", connectDur.Round(time.Millisecond))

	if err := p.updateVPPInfo(); err != nil {
		return errors.WithMessage(err, "retrieving VPP info failed")
	}

	// Connect to VPP status socket
	var statsSocket string
	if p.config.StatsSocketPath != "" {
		statsSocket = p.config.StatsSocketPath
	} else {
		statsSocket = adapter.DefaultStatsSocket
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

// AfterInit reports status check.
func (p *Plugin) AfterInit() error {
	// Register providing status reports (push mode)
	p.StatusCheck.Register(p.PluginName, nil)
	p.StatusCheck.ReportStateChange(p.PluginName, statuscheck.OK, nil)

	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	p.wg.Add(1)
	go p.handleVPPConnectionEvents(ctx)

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

// VPPInfo returns information about VPP session.
func (p *Plugin) VPPInfo() (VPPInfo, error) {
	p.infoMu.Lock()
	defer p.infoMu.Unlock()
	return p.vppInfo, nil
}

func (p *Plugin) updateVPPInfo() error {
	if p.vppConn == nil {
		return fmt.Errorf("VPP connection is nil")
	}

	vppAPIChan, err := p.vppConn.NewAPIChannel()
	if err != nil {
		return err
	}
	defer vppAPIChan.Close()

	vpeHandler := vppcalls.CompatibleVpeHandler(vppAPIChan)

	version, err := vpeHandler.RunCli("show version verbose")
	if err != nil {
		p.Log.Warnf("RunCli error: %v", err)
	} else {
		p.Log.Debugf("vpp# show version verbose\n%s", version)
	}

	cmdline, err := vpeHandler.RunCli("show version cmdline")
	if err != nil {
		p.Log.Warnf("RunCli error: %v", err)
	} else {
		out := strings.Replace(cmdline, "\n", "", -1)
		p.Log.Debugf("vpp# show version cmdline:\n%s", out)
	}

	ver, err := vpeHandler.GetVersionInfo()
	if err != nil {
		return err
	}

	p.Log.Infof("VPP version: %v", ver.Version)

	vpe, err := vpeHandler.GetVpeInfo()
	if err != nil {
		return err
	}

	p.Log.WithFields(logging.Fields{
		"PID":      vpe.PID,
		"ClientID": vpe.ClientIdx,
	}).Debugf("loaded %d VPP modules: %v", len(vpe.ModuleVersions), vpe.ModuleVersions)

	p.infoMu.Lock()
	p.vppInfo = VPPInfo{
		Connected:   true,
		VersionInfo: *ver,
		VpeInfo:     *vpe,
	}
	p.infoMu.Unlock()

	return nil
}

// handleVPPConnectionEvents handles VPP connection events.
func (p *Plugin) handleVPPConnectionEvents(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case event := <-p.vppConChan:
			if event.State == govpp.Connected {
				if err := p.updateVPPInfo(); err != nil {
					p.Log.Errorf("updating VPP info failed: %v", err)
				}

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
			} else if event.State == govpp.Failed || event.State == govpp.Disconnected {
				p.infoMu.Lock()
				p.vppInfo.Connected = false
				p.infoMu.Unlock()

				p.lastConnErr = errors.Errorf("VPP connection lost (event: %+v)", event)
				p.StatusCheck.ReportStateChange(p.PluginName, statuscheck.Error, p.lastConnErr)
			} else {
				p.Log.Debugf("VPP connection state: %+v", event)
			}

			p.infoMu.Lock()
			p.lastEvent = event
			p.infoMu.Unlock()

		case <-ctx.Done():
			return
		}
	}
}
