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

package integration

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	govpp "git.fd.io/govpp.git/core"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	ps "github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var (
	vppPath = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	debug   = flag.Bool("debug", false, "Turn on debug mode.")
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
	if *debug {
		govpp.SetLogLevel(logrus.DebugLevel)
	}
}

type testCtx struct {
	VPP  *exec.Cmd
	Conn *govpp.Connection
	T    *testing.T
}

func setupVPP(t *testing.T) *testCtx {
	if os.Getenv("TRAVIS") != "" {
		t.Skip("skipping test for Travis")
	}
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
		if process.Executable() == *vppPath || process.Executable() == "vpp" {
			t.Fatalf("VPP is already running, PID: %v", process.Pid())
		}
	}

	// remove binapi files from previous run
	var removeFile = func(path string) {
		if err := os.Remove(path); err == nil {
			logf("removed file %q", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("removing file %q failed: %v", path, err)
		}
	}
	// rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api
	removeFile("/dev/shm/vpe-api")
	removeFile("/dev/shm/global_vm")
	removeFile("/dev/shm/db")

	time.Sleep(time.Millisecond * 250)

	logf("starting VPP process: %q", *vppPath)

	cmd := exec.Command(*vppPath, "-c", "/etc/vpp/vpp.conf")
	//cmd.Stderr = os.Stderr
	//cmd.Stdout = os.Stdout

	// ensure that process is killed when current process exits
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting VPP failed: %v", err)
	}

	adapter := govppmux.NewVppAdapter("")

	logf("connecting to VPP..")
	conn, err := govpp.Connect(adapter)
	if err != nil {
		logf("sending KILL signal to VPP")
		if err := cmd.Process.Kill(); err != nil {
			logf("KILL FAILED: %v", err)
		}
		t.Fatalf("connecting to VPP failed: %v", err)
	} else {
		logf("connected successfully")
	}

	return &testCtx{VPP: cmd, T: t, Conn: conn}
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

	logf("sending SIGTERM to VPP")
	if err := ctx.VPP.Process.Signal(syscall.SIGTERM); err != nil {
		ctx.T.Fatalf("sending signal to VPP failed: %v", err)
	}
	if err := ctx.VPP.Wait(); err != nil {
		logf("VPP process wait failed: %v", err)
	}
}

func logf(f string, v ...interface{}) {
	if *debug {
		log.Output(2, fmt.Sprintf(f, v...))
	}
}
