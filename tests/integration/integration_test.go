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

	govpp "git.fd.io/govpp.git/core"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	"github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
)

var (
	vppPath = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	debug   = flag.Bool("debug", false, "Turn on debug mode.")
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	govpp.SetControlPingMessages(&vpe.ControlPing{}, &vpe.ControlPingReply{})
}

func logf(f string, v ...interface{}) {
	if *debug {
		log.Output(2, fmt.Sprintf(f, v...))
	}
}

type testCtx struct {
	VPP  *exec.Cmd
	Conn *govpp.Connection
	T    *testing.T
}

func setupTest(t *testing.T) *testCtx {
	if os.Getenv("TRAVIS") != "" {
		t.Skip("skipping test for Travis")
	}

	RegisterTestingT(t)

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

	var removeFile = func(path string) {
		if err := os.Remove("/dev/shm/vpe-api"); err != nil && !os.IsNotExist(err) {
			t.Fatalf("removing file %q failed: %v", path, err)
		} else {
			logf("removed file %q", path)
		}
	}
	// rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api
	removeFile("/dev/shm/vpe-api")
	removeFile("/dev/shm/global_vm")
	removeFile("/dev/shm/db")

	logf("starting VPP process: %q", *vppPath)

	cmd := exec.Command(*vppPath, "-c", "/etc/vpp/vpp.conf")
	//cmd.Stderr = os.Stderr
	//cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting VPP failed: %v", err)
	}

	logf("connecting to VPP..")

	adapter := govppmux.NewVppAdapter("")
	conn, err := govpp.Connect(adapter)
	if err != nil {
		if err := cmd.Process.Kill(); err != nil {
			logf("KILL FAILED: %v", err)
		}
		t.Fatalf("connecting to VPP failed: %v", err)
	}

	return &testCtx{VPP: cmd, T: t, Conn: conn}
}

func (ctx *testCtx) teardown() {
	ctx.Conn.Disconnect()

	if err := ctx.VPP.Process.Signal(syscall.SIGTERM); err != nil {
		ctx.T.Fatalf("sending signal to VPP failed: %v", err)
	}
	if err := ctx.VPP.Wait(); err != nil {
		logf("VPP process wait failed: %v", err)
	}
}

func TestVersion(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardown()

	channel, err := ctx.Conn.NewAPIChannel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()

	info, err := vppcalls.GetVersionInfo(channel)
	if err != nil {
		t.Fatalf("getting version info failed: %v", err)
		return
	}

	t.Logf("version info: %+v", info)
}
