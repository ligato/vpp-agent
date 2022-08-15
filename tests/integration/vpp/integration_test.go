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

package vpp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
	"go.fd.io/govpp/adapter"
	"go.fd.io/govpp/adapter/socketclient"
	"go.fd.io/govpp/adapter/statsclient"
	govppapi "go.fd.io/govpp/api"
	govppcore "go.fd.io/govpp/core"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"
)

const (
	vppConnectRetryDelay = time.Millisecond * 500
	vppBootDelay         = time.Millisecond * 200
	vppTermDelay         = time.Millisecond * 50
	vppExitTimeout       = time.Second * 1

	defaultVPPConfig = `
		unix {
			nodaemon
			cli-listen /run/vpp/cli.sock
			cli-no-pager
			log /tmp/vpp.log
			full-coredump
		}
		api-trace {
			on
		}
		socksvr {
			socket-name /run/vpp/api.sock
		}
		statseg {
			socket-name /run/vpp/stats.sock
			per-node-counters on
		}
		plugins {
			plugin dpdk_plugin.so { disable }
		}`
	// in older versions of VPP (<=20.09), NAT plugin was also configured via the startup config file
	withNatStartupConf = `
		nat {
			endpoint-dependent
		}`
)

type TestCtx struct {
	t              *testing.T
	Ctx            context.Context
	vppCmd         *exec.Cmd
	stderr, stdout *bytes.Buffer
	Conn           *govppcore.Connection
	StatsConn      *govppcore.StatsConnection
	vppBinapi      govppapi.Channel
	vppStats       govppapi.StatsProvider
	vpp            vppcalls.VppCoreAPI
	versionInfo    *vppcalls.VersionInfo
	vppClient      *vppClient
}

func startVPP(t *testing.T, stdout, stderr io.Writer) *exec.Cmd {
	// check if VPP process is not running already
	processes, err := ps.Processes()
	if err != nil {
		t.Fatalf("listing processes failed: %v", err)
	}
	for _, process := range processes {
		proc := process.Executable()
		if strings.Contains(proc, "vpp") && process.Pid() != os.Getpid() {
			t.Logf(" - found process: %+v", process)
		}
		switch proc {
		case *vppPath, "vpp", "vpp_main":
			t.Fatalf("VPP is already running (PID: %v)", process.Pid())
		}
	}

	// remove binapi files from previous run
	var removeFile = func(path string) {
		if err := os.Remove(path); err == nil {
			t.Logf("removed file %q", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("removing file %q failed: %v", path, err)
		}
	}
	removeFile(*vppSockAddr)

	// ensure VPP runtime directory exists
	if err := os.Mkdir("/run/vpp", 0755); err != nil && !os.IsExist(err) {
		t.Logf("mkdir failed: %v", err)
	}

	// setup VPP process
	vppCmd := exec.Command(*vppPath)
	if *vppConfig != "" {
		vppCmd.Args = append(vppCmd.Args, "-c", *vppConfig)
	} else {
		config := defaultVPPConfig
		if os.Getenv("VPPVER") <= "20.09" {
			config += withNatStartupConf
		}
		vppCmd.Args = append(vppCmd.Args, config)
	}
	if *debug {
		vppCmd.Stderr = os.Stderr
		vppCmd.Stdout = os.Stdout
	} else {
		vppCmd.Stderr = stderr
		vppCmd.Stdout = stdout
	}

	// ensure that process is killed when current process exits
	vppCmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}

	if err := vppCmd.Start(); err != nil {
		t.Fatalf("starting VPP failed: %v", err)
	}

	t.Logf("VPP start OK (PID: %v)", vppCmd.Process.Pid)
	return vppCmd
}

// reRegisterMessage overwrites the original registration of Messages in GoVPP with new message registration.
func reRegisterMessage(x govppapi.Message) {
	typ := reflect.TypeOf(x)
	namecrc := x.GetMessageName() + "_" + x.GetCrcString()
	binapiPath := path.Dir(reflect.TypeOf(x).Elem().PkgPath())
	govppapi.GetRegisteredMessages()[binapiPath][namecrc] = x
	govppapi.GetRegisteredMessageTypes()[binapiPath][typ] = namecrc
}

func hackForBugInGoVPPMessageCache(t *testing.T, adapter adapter.VppAPI, vppCmd *exec.Cmd) error {
	// connect to VPP
	conn, apiChannel, _ := connectToBinAPI(t, adapter, vppCmd)
	binapiVersion, err := binapi.CompatibleVersion(apiChannel)
	if err != nil {
		return err
	}

	// overwrite messages with messages from correct VPP version
	for _, msg := range binapi.Versions[binapiVersion].AllMessages() {
		reRegisterMessage(msg)
	}

	// disconnect from VPP (GoVPP is caching the messages that we want to override
	// by first connection to VPP -> we must disconnect and reconnect later again)
	disconnectBinAPI(t, conn, apiChannel, nil)

	return nil
}

func setupVPP(t *testing.T) *TestCtx {
	RegisterTestingT(t)

	start := time.Now()

	ctx := context.TODO()

	// start VPP process
	var stderr, stdout bytes.Buffer
	vppCmd := startVPP(t, &stdout, &stderr)
	vppPID := uint32(vppCmd.Process.Pid)

	// if setupVPP fails we need stop the VPP process
	defer func() {
		if t.Failed() {
			stopVPP(t, vppCmd)
		}
	}()

	// wait until the socket is ready
	adapter := socketclient.NewVppClient(*vppSockAddr)
	if err := adapter.WaitReady(); err != nil {
		t.Logf("WaitReady error: %v", err)
	}
	time.Sleep(vppBootDelay)

	// FIXME: this is a hack for GoVPP bug when register of the same message(same CRC and name) but different
	//  VPP version overwrites the already registered message from one VPP version (map key is only CRC+name
	//  and that didn't change with VPP version, but generated binapi generated 2 different go types for it).
	//  Similar fix exists also for govppmux.
	if err := hackForBugInGoVPPMessageCache(t, adapter, vppCmd); err != nil {
		t.Fatal("can't apply hack fixing bug in GoVPP regarding stream's message type resolving")
	}

	// connect to VPP's binary API
	conn, apiChannel, vppClient := connectToBinAPI(t, adapter, vppCmd)
	vpeHandler := vppcalls.CompatibleHandler(vppClient)

	// retrieve VPP version
	versionInfo, err := vpeHandler.GetVersion(ctx)
	if err != nil {
		t.Fatalf("getting version info failed: %v", err)
	}
	t.Logf("VPP version: %v", versionInfo.Version)
	if versionInfo.Version == "" {
		t.Fatal("expected VPP version to not be empty")
	}
	// verify connected session
	vpeInfo, err := vpeHandler.GetSession(ctx)
	if err != nil {
		t.Fatalf("getting vpp info failed: %v", err)
	}
	if vpeInfo.PID != vppPID {
		t.Fatalf("expected VPP PID to be %v, got %v", vppPID, vpeInfo.PID)
	}

	vppClient.vpp = vpeHandler

	// connect to stats
	statsClient := statsclient.NewStatsClient("")
	statsConn, err := govppcore.ConnectStats(statsClient)
	if err != nil {
		t.Logf("connecting to VPP stats API failed: %v", err)
	} else {
		vppClient.stats = statsConn
	}

	t.Logf("-> VPP ready (took %v)", time.Since(start).Seconds())

	return &TestCtx{
		t:           t,
		Ctx:         ctx,
		versionInfo: versionInfo,
		vpp:         vpeHandler,
		vppCmd:      vppCmd,
		stderr:      &stderr,
		stdout:      &stdout,
		Conn:        conn,
		vppBinapi:   apiChannel,
		vppStats:    statsConn,
		vppClient:   vppClient,
	}
}

func connectToBinAPI(t *testing.T, adapter adapter.VppAPI, vppCmd *exec.Cmd) (*govppcore.Connection, govppapi.Channel, *vppClient) {
	connectRetry := func(retries int) (conn *govppcore.Connection, err error) {
		for i := 1; i <= retries; i++ {
			conn, err = govppcore.Connect(adapter)
			if err != nil {
				t.Logf("attempt #%d failed: %v, retrying in %v", i, err, vppConnectRetryDelay)
				time.Sleep(vppConnectRetryDelay)
				continue
			}
			return
		}
		return nil, fmt.Errorf("failed to connect after %d retries", retries)
	}

	// connect to binapi
	conn, err := connectRetry(int(*vppRetry))
	if err != nil {
		t.Errorf("connecting to VPP failed: %v", err)
		if err := vppCmd.Process.Kill(); err != nil {
			t.Fatalf("killing VPP failed: %v", err)
		}
		if state, err := vppCmd.Process.Wait(); err != nil {
			t.Logf("VPP wait failed: %v", err)
		} else {
			t.Logf("VPP wait OK: %v", state)
		}
		t.FailNow()
	}

	apiChannel, err := conn.NewAPIChannel()
	if err != nil {
		t.Fatalf("creating channel failed: %v", err)
	}

	vppClient := &vppClient{
		t:    t,
		conn: conn,
		ch:   apiChannel,
	}
	return conn, apiChannel, vppClient
}

func (ctx *TestCtx) teardownVPP() {
	disconnectBinAPI(ctx.t, ctx.Conn, ctx.vppBinapi, ctx.StatsConn)
	stopVPP(ctx.t, ctx.vppCmd)
}

func disconnectBinAPI(t *testing.T, conn *govppcore.Connection, vppBinapi govppapi.Channel,
	statsConn *govppcore.StatsConnection) {
	// disconnect sometimes hangs
	done := make(chan struct{})
	go func() {
		if statsConn != nil {
			statsConn.Disconnect()
		}
		vppBinapi.Close()
		conn.Disconnect()
		close(done)
	}()
	select {
	case <-done:
		time.Sleep(vppTermDelay)
	case <-time.After(vppExitTimeout):
		t.Logf("VPP disconnect timeout")
	}
}

func stopVPP(t *testing.T, vppCmd *exec.Cmd) {
	if err := vppCmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("sending SIGTERM to VPP failed: %v", err)
	}
	// wait until VPP exits
	exit := make(chan struct{})
	go func() {
		if err := vppCmd.Wait(); err != nil {
			var exiterr *exec.ExitError
			if errors.As(err, &exiterr) && strings.Contains(exiterr.Error(), "core dumped") {
				t.Logf("VPP process CRASHED: %s", exiterr.Error())
			} else {
				t.Logf("VPP process wait failed: %v", err)
			}
		} else {
			t.Logf("VPP exit OK")
		}
		close(exit)
	}()
	select {
	case <-exit:
		// exited
	case <-time.After(vppExitTimeout):
		t.Logf("VPP exit timeout")
		t.Logf("sending SIGKILL to VPP..")
		if err := vppCmd.Process.Signal(syscall.SIGKILL); err != nil {
			t.Fatalf("sending SIGKILL to VPP failed: %v", err)
		}
	}
}

type vppClient struct {
	t       *testing.T
	conn    *govppcore.Connection
	ch      govppapi.Channel
	stats   govppapi.StatsProvider
	vpp     vppcalls.VppCoreAPI
	version vpp.Version
}

func (v *vppClient) NewAPIChannel() (govppapi.Channel, error) {
	return v.conn.NewAPIChannel()
}

func (v *vppClient) NewStream(ctx context.Context, options ...govppapi.StreamOption) (govppapi.Stream, error) {
	return v.conn.NewStream(ctx, options...)
}

func (v *vppClient) Invoke(ctx context.Context, req govppapi.Message, reply govppapi.Message) error {
	return v.conn.Invoke(ctx, req, reply)
}

func (v *vppClient) Version() vpp.Version {
	return v.version
}

func (v *vppClient) BinapiVersion() vpp.Version {
	vppapiChan, err := v.conn.NewAPIChannel()
	if err != nil {
		v.t.Fatalf("Can't create new API channel (to get binary API version) due to: %v", err)
	}
	binapiVersion, err := binapi.CompatibleVersion(vppapiChan)
	if err != nil {
		v.t.Fatalf("Can't get binary API version due to: %v", err)
	}
	return binapiVersion
}

func (v *vppClient) CheckCompatiblity(msgs ...govppapi.Message) error {
	return v.ch.CheckCompatiblity(msgs...)
}

func (v *vppClient) Stats() govppapi.StatsProvider {
	return v.stats
}

func (v *vppClient) IsPluginLoaded(plugin string) bool {
	ctx := context.Background()
	plugins, err := v.vpp.GetPlugins(ctx)
	if err != nil {
		v.t.Fatalf("GetPlugins failed: %v", plugins)
	}
	for _, p := range plugins {
		if p.Name == plugin {
			return true
		}
	}
	return false
}

func (v *vppClient) OnReconnect(h func()) {
	// no-op
}
