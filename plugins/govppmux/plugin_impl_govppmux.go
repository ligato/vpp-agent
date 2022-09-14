//  Copyright (c) 2021 Cisco and/or its affiliates.
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

package govppmux

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.fd.io/govpp/adapter"
	govppapi "go.fd.io/govpp/api"
	govpp "go.fd.io/govpp/core"
	"go.fd.io/govpp/proxy"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/rpc/rest"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"

	_ "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp2101"
	_ "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp2106"
	_ "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls/vpp2202"
)

var (
	disabledSocketClient = os.Getenv("GOVPPMUX_NOSOCK") != ""
)

// Plugin is the govppmux plugin implementation.
type Plugin struct {
	Deps

	config *Config

	vpeHandler vppcalls.VppCoreAPI

	binapiVersion vpp.Version
	vppAdapter    adapter.VppAPI
	vppConn       *govpp.Connection
	vppConChan    chan govpp.ConnectionEvent
	lastConnErr   error
	vppapiChan    govppapi.Channel
	lastPID       uint32

	onnConnectMu sync.Mutex
	onConnects   []func()

	statsMu      sync.Mutex
	statsAdapter adapter.StatsAPI
	statsConn    *govpp.StatsConnection

	proxy *proxy.Server

	// infoMu synchonizes access to fields
	// vppInfo and lastEvent
	infoMu    sync.Mutex
	vppInfo   VPPInfo
	lastEvent govpp.ConnectionEvent

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Deps defines dependencies for the govppmux plugin.
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

	// set GoVPP config
	govpp.HealthCheckProbeInterval = p.config.HealthCheckProbeInterval
	govpp.HealthCheckReplyTimeout = p.config.HealthCheckReplyTimeout
	govpp.HealthCheckThreshold = p.config.HealthCheckThreshold
	govpp.DefaultReplyTimeout = p.config.ReplyTimeout

	var address string
	useShm := disabledSocketClient || p.config.ConnectViaShm || p.config.ShmPrefix != ""
	if useShm {
		address = p.config.ShmPrefix
	} else {
		address = p.config.BinAPISocketPath
	}

	p.Log.Debugf("found %d registered VPP handlers", len(vpp.GetHandlers()))
	for name, handler := range vpp.GetHandlers() {
		versions := handler.Versions()
		p.Log.Debugf("- handler: %-10s has %d versions: %v", name, len(versions), versions)
	}

	// FIXME: this is a hack for GoVPP bug when register of the same message(same CRC and name) but different
	//  VPP version overwrites the already registered message from one VPP version (map key is only CRC+name
	//  and that didn't change with VPP version, but generated binapi generated 2 different go types for it).
	//  Similar fix exists also for integration tests.
	if err := p.hackForBugInGoVPPMessageCache(address, useShm); err != nil {
		return errors.Errorf("can't apply hack fixing bug in GoVPP "+
			"regarding stream's message type resolving: %v", err)
	}

	// TODO: Async connect & automatic reconnect support is not yet implemented in the agent,
	// so synchronously wait until connected to VPP.
	startTime := time.Now()
	p.Log.Debugf("connecting to VPP..")

	p.vppAdapter = NewVppAdapter(address, useShm)
	p.vppConn, p.vppConChan, err = govpp.AsyncConnect(p.vppAdapter, p.config.RetryConnectCount, p.config.RetryConnectTimeout)
	if err != nil {
		return err
	}
	if err := p.waitForConnectionEvent(p.vppConChan); err != nil {
		return err
	}
	took := time.Since(startTime)
	p.Log.Debugf("connection to VPP established (took %s)", took.Round(time.Millisecond))

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

	if p.config.ProxyEnabled {
		// register binapi messages to gob package (required for proxy)
		msgList := binapi.Versions[p.binapiVersion]
		for _, msg := range msgList.AllMessages() {
			gob.Register(msg)
		}
		err := p.startProxy(NewVppAdapter(address, useShm), NewStatsAdapter(statsSocket))
		if err != nil {
			p.Log.Warnf("VPP proxy failed to start: %v", err)
		} else {
			p.Log.Infof("VPP proxy ready")
		}
	}

	// register REST API handlers
	p.registerHandlers(p.HTTPHandlers)

	return nil
}

// waitForConnectionEvent waits for Connected event from govpp
func (p *Plugin) waitForConnectionEvent(vppConChan chan govpp.ConnectionEvent) error {
	for {
		event, ok := <-vppConChan
		if !ok {
			return errors.Errorf("VPP connection state channel closed")
		}
		if event.State == govpp.Connected {
			break
		} else if event.State == govpp.Failed || event.State == govpp.Disconnected {
			return errors.Errorf("unable to establish connection to VPP (%v)", event.Error)
		} else {
			p.Log.Debugf("VPP connection state: %+v", event)
		}
	}
	return nil
}

func (p *Plugin) hackForBugInGoVPPMessageCache(address string, useShm bool) error {
	// connect to VPP
	startTime := time.Now()
	vppAdapter := NewVppAdapter(address, useShm)
	vppConn, vppConChan, err := govpp.AsyncConnect(vppAdapter, p.config.RetryConnectCount, p.config.RetryConnectTimeout)
	if err != nil {
		return err
	}
	if err := p.waitForConnectionEvent(vppConChan); err != nil {
		return err
	}
	took := time.Since(startTime)
	p.Log.Debugf("first connection to VPP established (took %s)", took.Round(time.Millisecond))

	if vppConn == nil {
		return fmt.Errorf("VPP connection is nil")
	}

	// detect binary API version
	vppapiChan, err := vppConn.NewAPIChannel()
	if err != nil {
		return err
	}
	binapiVersion, err := binapi.CompatibleVersion(vppapiChan)
	if err != nil {
		return err
	}

	// overwrite messages with messages from correct VPP version
	for _, msg := range binapi.Versions[binapiVersion].AllMessages() {
		reRegisterMessage(msg)
	}

	// disconnect from VPP (GoVPP is caching the messages that we want to override
	// by first connection to VPP -> we must disconnect and reconnect later again)
	vppConn.Disconnect()

	return nil
}

// reRegisterMessage overwrites the original registration of Messages in GoVPP with new message registration.
func reRegisterMessage(x govppapi.Message) {
	typ := reflect.TypeOf(x)
	namecrc := x.GetMessageName() + "_" + x.GetCrcString()
	binapiPath := path.Dir(reflect.TypeOf(x).Elem().PkgPath())
	govppapi.GetRegisteredMessages()[binapiPath][namecrc] = x
	govppapi.GetRegisteredMessageTypes()[binapiPath][typ] = namecrc
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

	if p.proxy != nil {
		p.proxy.DisconnectBinapi()
		p.proxy.DisconnectStats()
	}

	return nil
}

func (p *Plugin) Version() vpp.Version {
	return p.binapiVersion
}

func (p *Plugin) CheckCompatiblity(msgs ...govppapi.Message) error {
	p.infoMu.Lock()
	defer p.infoMu.Unlock()
	if p.vppapiChan == nil {
		apiChan, err := p.vppConn.NewAPIChannel()
		if err != nil {
			return err
		}
		p.vppapiChan = apiChan
	}
	return p.vppapiChan.CheckCompatiblity(msgs...)
}

func (p *Plugin) Stats() govppapi.StatsProvider {
	if p.statsConn == nil {
		return nil
	}
	return p
}

func (p *Plugin) BinapiVersion() vpp.Version {
	return p.binapiVersion
}

// VPPInfo returns information about VPP session.
func (p *Plugin) VPPInfo() VPPInfo {
	p.infoMu.Lock()
	defer p.infoMu.Unlock()
	return p.vppInfo
}

// IsPluginLoaded returns true if plugin is loaded.
func (p *Plugin) IsPluginLoaded(plugin string) bool {
	p.infoMu.Lock()
	defer p.infoMu.Unlock()
	for _, p := range p.vppInfo.Plugins {
		if p.Name == plugin {
			return true
		}
	}
	return false
}

func (p *Plugin) updateVPPInfo() (err error) {
	if p.vppConn == nil {
		return fmt.Errorf("VPP connection is nil")
	}

	p.vppapiChan, err = p.vppConn.NewAPIChannel()
	if err != nil {
		return err
	}
	p.binapiVersion, err = binapi.CompatibleVersion(p.vppapiChan)
	if err != nil {
		return err
	}

	p.vpeHandler, err = vppcalls.NewHandler(p)
	if err != nil {
		return errors.New("no compatible VPP handler found")
	}

	ctx := context.TODO()

	version, err := p.vpeHandler.RunCli(ctx, "show version verbose")
	if err != nil {
		p.Log.Warnf("RunCli error: %v", err)
	} else {
		p.Log.Debugf("vpp# show version verbose\n%s", version)
	}
	cmdline, err := p.vpeHandler.RunCli(ctx, "show version cmdline")
	if err != nil {
		p.Log.Warnf("RunCli error: %v", err)
	} else {
		out := strings.Replace(cmdline, "\n", "", -1)
		p.Log.Debugf("vpp# show version cmdline:\n%s", out)
	}

	ver, err := p.vpeHandler.GetVersion(ctx)
	if err != nil {
		return err
	}
	session, err := p.vpeHandler.GetSession(ctx)
	if err != nil {
		return err
	}
	p.Log.WithFields(logging.Fields{
		"PID":      session.PID,
		"ClientID": session.ClientIdx,
	}).Infof("VPP version: %v", ver.Version)

	modules, err := p.vpeHandler.GetModules(ctx)
	if err != nil {
		return err
	}
	p.Log.Debugf("VPP has %d core modules: %v", len(modules), modules)

	plugins, err := p.vpeHandler.GetPlugins(ctx)
	if err != nil {
		return err
	}
	// sort plugins by name
	sort.Slice(plugins, func(i, j int) bool { return plugins[i].Name < plugins[j].Name })

	p.Log.Debugf("VPP loaded %d plugins", len(plugins))
	for _, plugin := range plugins {
		p.Log.Debugf(" - plugin: %v", plugin)
	}

	p.infoMu.Lock()
	p.vppInfo = VPPInfo{
		Connected:   true,
		VersionInfo: *ver,
		SessionInfo: *session,
		Plugins:     plugins,
	}
	p.infoMu.Unlock()

	if p.lastPID != 0 && p.lastPID != session.PID {
		p.Log.Warnf("VPP has restarted (previous PID: %d)", p.lastPID)
		p.onConnect()
	}
	p.lastPID = session.PID

	return nil
}

func (p *Plugin) OnReconnect(fn func()) {
	p.onnConnectMu.Lock()
	defer p.onnConnectMu.Unlock()
	p.onConnects = append(p.onConnects, fn)
}

func (p *Plugin) onConnect() {
	p.onnConnectMu.Lock()
	defer p.onnConnectMu.Unlock()
	for _, fn := range p.onConnects {
		fn()
	}
}

// handleVPPConnectionEvents handles VPP connection events.
func (p *Plugin) handleVPPConnectionEvents(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case event, ok := <-p.vppConChan:
			if !ok {
				p.lastConnErr = errors.Errorf("VPP connection state channel closed")
				p.StatusCheck.ReportStateChange(p.PluginName, statuscheck.Error, p.lastConnErr)
				return
			}

			p.Log.Debugf("VPP connection state changed: %+v", event)

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

				// TODO: fix reconnecting after reaching maximum reconnect attempts
				//		current implementation wont work with already created govpp channels
				// keep reconnecting
				/*if event.State == govpp.Failed {
					p.vppConn, p.vppConChan, _ = govpp.AsyncConnect(p.vppAdapter, p.config.RetryConnectCount, p.config.RetryConnectTimeout)
				}*/
			} else if event.State == govpp.NotResponding {
				p.infoMu.Lock()
				p.vppInfo.Connected = false
				p.infoMu.Unlock()

				p.lastConnErr = errors.Errorf("VPP is not responding (event: %+v)", event)
				p.StatusCheck.ReportStateChange(p.PluginName, statuscheck.Error, p.lastConnErr)
			} else {
				p.Log.Warnf("unknown VPP connection state: %+v", event)
			}

			p.infoMu.Lock()
			p.lastEvent = event
			p.infoMu.Unlock()

		case <-ctx.Done():
			return
		}
	}
}

func (p *Plugin) startProxy(vppapi adapter.VppAPI, statsapi adapter.StatsAPI) (err error) {
	p.Log.Debugf("starting VPP proxy")

	p.proxy, err = proxy.NewServer()
	if err != nil {
		return errors.Wrap(err, "creating proxy failed")
	}
	if err = p.proxy.ConnectBinapi(vppapi); err != nil {
		return errors.Wrap(err, "connecting binapi for proxy failed")
	}
	if err = p.proxy.ConnectStats(statsapi); err != nil {
		return errors.Wrap(err, "connecting stats for proxy failed")
	}
	return nil
}
