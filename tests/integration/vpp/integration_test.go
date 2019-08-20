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
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/adapter/statsclient"
	govppapi "git.fd.io/govpp.git/api"
	govppcore "git.fd.io/govpp.git/core"
	"github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
)

var (
	vppPath     = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	vppConfig   = flag.String("vpp-config", "", "VPP config file")
	vppSockAddr = flag.String("vpp-sock-addr", "", "VPP binapi socket address")

	debug = flag.Bool("debug", false, "Turn on debug mode.")
)

const (
	vppConnectRetries    = 3
	vppConnectRetryDelay = time.Millisecond * 500
	vppBootDelay         = time.Millisecond * 200
	vppTermDelay         = time.Millisecond * 50
	vppExitTimeout       = time.Second * 1

	vppConf = `
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
			default
		}
		statseg {
			default
			per-node-counters on
		}
		plugins {
			plugin dpdk_plugin.so { disable }
		}`
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
	if *debug {
		govppcore.SetLogLevel(logrus.DebugLevel)
	}
}

type testCtx struct {
	t              *testing.T
	VPP            *exec.Cmd
	stderr, stdout *bytes.Buffer
	Conn           *govppcore.Connection
	StatsConn      *govppcore.StatsConnection
	vppBinapi      govppapi.Channel
	vppStats       govppapi.StatsProvider
	vpe            vppcalls.VpeVppAPI
	versionInfo    *vppcalls.VersionInfo
}

func setupVPP(t *testing.T) *testCtx {
	if os.Getenv("TRAVIS") != "" {
		t.Skip("skipping test for Travis")
	}

	RegisterTestingT(t)

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

	var removeFile = func(path string) {
		if err := os.Remove(path); err == nil {
			t.Logf("removed file %q", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("removing file %q failed: %v", path, err)
		}
	}
	// remove binapi files from previous run
	removeFile(*vppSockAddr)
	if err := os.Mkdir("/run/vpp", 0755); err != nil && !os.IsExist(err) {
		t.Logf("mkdir failed: %v", err)
	}

	var stderr, stdout bytes.Buffer

	vppCmd := exec.Command(*vppPath)
	if *vppConfig != "" {
		vppCmd.Args = append(vppCmd.Args, "-c", *vppConfig)
	} else {
		vppCmd.Args = append(vppCmd.Args, vppConf)
	}
	vppCmd.Stderr = &stderr
	vppCmd.Stdout = &stdout
	// ensure that process is killed when current process exits
	vppCmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	if err := vppCmd.Start(); err != nil {
		t.Fatalf("starting VPP failed: %v", err)
	}
	vppPID := uint32(vppCmd.Process.Pid)
	t.Logf("VPP start OK (PID: %v)", vppPID)

	adapter := socketclient.NewVppClient(*vppSockAddr)

	// wait until the socket is ready
	if err := adapter.WaitReady(); err != nil {
		t.Logf("WaitReady failed: %v", err)
	}
	time.Sleep(vppBootDelay)

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
	conn, err := connectRetry(vppConnectRetries)
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

	ch, err := conn.NewAPIChannel()
	if err != nil {
		t.Fatalf("creating channel failed: %v", err)
	}

	vpeHandler := vppcalls.CompatibleVpeHandler(ch)
	versionInfo, err := vpeHandler.GetVersionInfo()
	if err != nil {
		t.Fatalf("getting version info failed: %v", err)
	}
	t.Logf("VPP version: %v", versionInfo.Version)
	if versionInfo.Version == "" {
		t.Fatal("expected VPP version to not be empty")
	}
	vpeInfo, err := vpeHandler.GetVpeInfo()
	if err != nil {
		t.Fatalf("getting vpe info failed: %v", err)
	}
	if vpeInfo.PID != vppPID {
		t.Fatalf("expected VPP PID to be %v, got %v", vppPID, vpeInfo.PID)
	}

	statsClient := statsclient.NewStatsClient("")
	statsConn, err := govppcore.ConnectStats(statsClient)
	if err != nil {
		t.Fatalf("connecting to VPP stats API failed: %v", err)
	}

	t.Logf("---------------")

	return &testCtx{
		t:           t,
		versionInfo: versionInfo,
		vpe:         vpeHandler,
		VPP:         vppCmd,
		stderr:      &stderr,
		stdout:      &stdout,
		Conn:        conn,
		vppBinapi:   ch,
		vppStats:    statsConn,
	}
}

func (ctx *testCtx) teardownVPP() {
	ctx.t.Logf("-----------------")

	// disconnect sometimes hangs
	done := make(chan struct{})
	go func() {
		ctx.StatsConn.Disconnect()
		ctx.vppBinapi.Close()
		ctx.Conn.Disconnect()
		close(done)
	}()
	select {
	case <-done:
		time.Sleep(vppTermDelay)

	case <-time.After(vppExitTimeout):
		ctx.t.Logf("VPP disconnect timeout")
	}

	if err := ctx.VPP.Process.Signal(syscall.SIGTERM); err != nil {
		ctx.t.Fatalf("sending SIGTERM to VPP failed: %v", err)
	}

	// wait until VPP exits
	exit := make(chan struct{})
	go func() {
		if err := ctx.VPP.Wait(); err != nil {
			ctx.t.Logf("VPP process wait failed: %v", err)
		}
		close(exit)
	}()
	select {
	case <-exit:
		ctx.t.Logf("VPP exit OK")

	case <-time.After(vppExitTimeout):
		ctx.t.Logf("VPP exit timeout")
		ctx.t.Logf("sending SIGKILL to VPP..")
		if err := ctx.VPP.Process.Signal(syscall.SIGKILL); err != nil {
			ctx.t.Fatalf("sending SIGKILL to VPP failed: %v", err)
		}
	}
}
