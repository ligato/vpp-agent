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
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

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
	debug       = flag.Bool("debug", false, "Turn on debug mode.")
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
	if *debug {
		govppcore.SetLogLevel(logrus.DebugLevel)
	}
}

type testCtx struct {
	VPP            *exec.Cmd
	stderr, stdout *bytes.Buffer
	Conn           *govppcore.Connection
	T              *testing.T
}

func setupVPP(t *testing.T) *testCtx {
	/*if os.Getenv("TRAVIS") != "" {
		t.Skip("skipping test for Travis")
	}*/
	logf("=== VPP setup ===")

	RegisterTestingT(t)

	// check if VPP process is not running already
	processes, err := ps.Processes()
	if err != nil {
		t.Fatalf("listing processes failed: %v", err)
	}
	for _, process := range processes {
		if strings.Contains(process.Executable(), "vpp") {
			logf("- found VPP process: %q (PID: %d)", process.Executable(), process.Pid())
		}
		if process.Executable() == *vppPath || process.Executable() == "vpp" || process.Executable() == "vpp_main" {
			t.Fatalf("VPP is already running, PID: %v", process.Pid())
		}
	}

	logf("starting VPP process: %q", *vppPath)

	cmd := exec.Command(*vppPath, "-c", *vppConfig)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	// ensure that process is killed when current process exits
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting VPP failed: %v", err)
	}

	time.Sleep(time.Millisecond * 250)

	adapter := govppmux.NewVppAdapter(*vppSockAddr, false)

	logf("connecting to VPP..")
	conn, err := govppcore.Connect(adapter)
	if err != nil {
		logf("sending KILL signal to VPP")
		if err := cmd.Process.Kill(); err != nil {
			logf("KILL FAILED: %v", err)
		}
		t.Fatalf("connecting to VPP failed: %v", err)
	} else {
		logf("connected successfully")
	}

	return &testCtx{
		VPP:    cmd,
		stderr: &stderr,
		stdout: &stdout,
		T:      t,
		Conn:   conn,
	}
}

func (ctx *testCtx) teardownVPP() {
	logf("--- VPP teardown ---")

	// disconnect sometimes hangs
	done := make(chan struct{})
	go func() {
		ctx.Conn.Disconnect()
		close(done)
	}()
	select {
	case <-done:
		logf("VPP disconnected")

	case <-time.After(time.Second * 1):
		logf("VPP disconnect timeout")
	}

	time.Sleep(time.Millisecond * 100)

	logf("sending SIGTERM to VPP")
	if err := ctx.VPP.Process.Signal(syscall.SIGTERM); err != nil {
		ctx.T.Fatalf("sending SIGTERM signal to VPP failed: %v", err)
	}

	exit := make(chan struct{})
	go func() {
		if err := ctx.VPP.Wait(); err != nil {
			logf("VPP process wait failed: %v", err)
		}
		close(exit)
	}()
	select {
	case <-exit:
		logf("VPP exited")

	case <-time.After(time.Second * 1):
		logf("VPP exit timeout")

		logf("sending SIGKILL to VPP")
		if err := ctx.VPP.Process.Signal(syscall.SIGKILL); err != nil {
			ctx.T.Fatalf("sending SIGKILL signal to VPP failed: %v", err)
		}
	}

}

func logf(f string, v ...interface{}) {
	if *debug {
		log.Output(2, fmt.Sprintf(f, v...))
	}
}
