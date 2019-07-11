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
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	govppcore "git.fd.io/govpp.git/core"
	"github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/ligato/vpp-agent/plugins/govppmux"
)

var (
	vppPath     = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	vppConfig   = flag.String("vpp-config", "/etc/vpp/vpp.conf", "VPP config file")
	vppSockAddr = flag.String("vpp-sock-addr", "", "VPP binapi socket address")

	debug = flag.Bool("debug", false, "Turn on debug mode.")
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
	Chan           govppapi.Channel
}

func setupVPP(t *testing.T) *testCtx {
	if os.Getenv("TRAVIS") != "" {
		t.Skip("skipping test for Travis")
	}
	t.Logf("=== VPP setup ===")

	RegisterTestingT(t)

	// check if VPP process is not running already
	processes, err := ps.Processes()
	if err != nil {
		t.Fatalf("listing processes failed: %v", err)
	}
	for _, process := range processes {
		if strings.Contains(process.Executable(), "vpp") {
			t.Logf("- found VPP process: %q (PID: %d)", process.Executable(), process.Pid())
		}
		if process.Executable() == *vppPath || process.Executable() == "vpp" || process.Executable() == "vpp_main" {
			t.Fatalf("VPP is already running, PID: %v", process.Pid())
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
	removeFile("/run/vpp-api.sock")

	t.Logf("starting VPP process: %q", *vppPath)

	cmd := exec.Command(*vppPath, "-c", *vppConfig)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	// ensure that process is killed when current process exits
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting VPP failed: %v", err)
	}
	t.Logf("VPP process started (PID: %v)", cmd.Process.Pid)

	adapter := govppmux.NewVppAdapter(*vppSockAddr, false)
	if err := adapter.WaitReady(); err != nil {
		t.Logf("WaitReady failed: %v", err)
	}

	time.Sleep(time.Millisecond * 100)

	t.Logf("connecting to VPP..")
	conn, err := govppcore.Connect(adapter)
	if err != nil {
		t.Logf("sending KILL signal to VPP")
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("killing VPP failed: %v", err)
		}
		if state, err := cmd.Process.Wait(); err != nil {
			t.Logf("VPP process wait failed: %v", err)
		} else {
			t.Logf("VPP killed: %v", state)
		}
		t.Fatalf("connecting to VPP failed: %v", err)
	} else {
		t.Logf("connected to VPP successfully")
	}

	ch, err := conn.NewAPIChannel()
	if err != nil {
		t.Fatalf("creating channel failed: %v", err)
	}

	return &testCtx{
		t:      t,
		VPP:    cmd,
		stderr: &stderr,
		stdout: &stdout,
		Conn:   conn,
		Chan:   ch,
	}
}

func (ctx *testCtx) teardownVPP() {
	ctx.t.Logf("--- VPP teardown ---")

	// disconnect sometimes hangs
	done := make(chan struct{})
	go func() {
		ctx.Chan.Close()
		ctx.Conn.Disconnect()
		close(done)
	}()
	select {
	case <-done:
		ctx.t.Logf("VPP disconnected")

	case <-time.After(time.Second * 1):
		ctx.t.Logf("VPP disconnect timeout")
	}

	time.Sleep(time.Millisecond * 100)

	ctx.t.Logf("sending SIGTERM to VPP")
	if err := ctx.VPP.Process.Signal(syscall.SIGTERM); err != nil {
		ctx.t.Fatalf("sending SIGTERM signal to VPP failed: %v", err)
	}

	exit := make(chan struct{})
	go func() {
		if err := ctx.VPP.Wait(); err != nil {
			ctx.t.Logf("VPP process wait failed: %v", err)
		}
		close(exit)
	}()
	select {
	case <-exit:
		ctx.t.Logf("VPP exited")

	case <-time.After(time.Second * 1):
		ctx.t.Logf("VPP exit timeout")

		ctx.t.Logf("sending SIGKILL to VPP")
		if err := ctx.VPP.Process.Signal(syscall.SIGKILL); err != nil {
			ctx.t.Fatalf("sending SIGKILL signal to VPP failed: %v", err)
		}
	}

}
