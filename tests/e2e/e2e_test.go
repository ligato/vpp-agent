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

package e2e

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	govppcore "git.fd.io/govpp.git/core"
	"github.com/fsouza/go-dockerclient"
	"github.com/gogo/protobuf/proto"
	"github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	vnf_agent "github.com/ligato/cn-infra/agent"
	"github.com/ligato/vpp-agent/pkg/models"
	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

var (
	vppPath     = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	vppConfig   = flag.String("vpp-config", "", "VPP config file")
	vppSockAddr = flag.String("vpp-sock-addr", "", "VPP binapi socket address")

	debug = flag.Bool("debug", false, "Turn on debug mode.")

	vppPingRegexp = regexp.MustCompile("Statistics: ([0-9]+) sent, ([0-9]+) received, ([0-9]+)% packet loss")
)

const (
	agentResyncTimeout = time.Second * 15
	vppExitTimeout     = time.Second * 1

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
	t                    *testing.T
	VPP                  *exec.Cmd
	vppStderr, vppStdout *bytes.Buffer
	plugins              *vppAgent
	agent                vnf_agent.Agent
	dockerClient         *docker.Client
	microservices        map[string]*microservice
	vppChan              govppapi.Channel
	vpe                  vppcalls.VpeVppAPI
}

func setupE2E(t *testing.T) *testCtx {
	if os.Getenv("TRAVIS") != "" {
		t.Skip("skipping test for Travis")
	}
	RegisterTestingT(t)

	// connect to the docker daemon
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		t.Fatalf("failed to get docker client instance from the environment variables: %v", err)
	}
	t.Logf("Using docker client endpoint: %+v", dockerClient.Endpoint())

	// make sure there are no microservices left from the previous run
	resetMicroservices(t, dockerClient)

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
	if *vppSockAddr != "" {
		removeFile(*vppSockAddr)
	}
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

	// start the agent
	inSync := make(chan struct{})
	plugins := setupAgent(inSync)
	agent := vnf_agent.NewAgent(
		vnf_agent.AllPlugins(plugins),
	)
	go func() {
		if err := agent.Run(); err != nil {
			t.Fatalf("agent failed with error: %v", err)
		}
	}()
	select {
	case <-inSync:
		t.Logf("agent is in-sync with VPP")

	case <-time.After(agentResyncTimeout):
		t.Fatal("agent resync timeout")
	}

	// create VPE handler just in case some test needs to run VPP CLIs, etc.
	vppChan, err := plugins.GoVppMux.NewAPIChannel()
	if err != nil {
		t.Fatalf("failed to create channel to VPP: %v", err)
	}
	vpe := vppcalls.CompatibleVpeHandler(vppChan)

	return &testCtx{
		t:             t,
		VPP:           vppCmd,
		vppStderr:     &stderr,
		vppStdout:     &stdout,
		dockerClient:  dockerClient,
		plugins:       plugins,
		agent:         agent,
		microservices: make(map[string]*microservice),
		vppChan:       vppChan,
		vpe:           vpe,
	}
}

func (ctx *testCtx) teardownE2E() {
	ctx.t.Logf("-----------------")

	// stop all microservices
	for msName := range ctx.microservices {
		ctx.stopMicroservice(msName)
	}

	// stop the agent
	if err := ctx.agent.Stop(); err != nil {
		ctx.t.Fatalf("agent shutdown failed: %v", err)
	} else {
		ctx.t.Logf("agent shutdown OK")
	}

	// terminate VPP
	ctx.vppChan.Close()
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

// syncAgent runs downstream resync and returns the list of executed operations.
func (ctx *testCtx) syncAgent() (executed kvs.RecordedTxnOps) {
	txn := ctx.plugins.KvScheduler.StartNBTransaction()
	txnCtx := context.Background()
	txnCtx = kvs.WithResync(txnCtx, kvs.DownstreamResync, true)
	txnSeqNum, err := txn.Commit(txnCtx)
	if err != nil {
		ctx.t.Fatalf("failed to sync agent with VPP: %v", err)
	}
	txnRec := ctx.plugins.KvScheduler.GetRecordedTransaction(txnSeqNum)
	return txnRec.Executed
}

// agentInSync checks if the agent NB config and the SB state (VPP+Linux)
// are in-sync.
func (ctx *testCtx) agentInSync() bool {
	ops := ctx.syncAgent()
	for _, op := range ops {
		if !op.NOOP {
			return false
		}
	}
	return true
}

func (ctx *testCtx) startMicroservice(msName string) (ms *microservice) {
	ms = createMicroservice(ctx.t, msName, ctx.dockerClient, ctx.plugins.NSPlugin)
	ctx.microservices[msName] = ms
	return ms
}

func (ctx *testCtx) stopMicroservice(msName string) {
	ms, found := ctx.microservices[msName]
	if !found {
		// bug inside a test
		ctx.t.Fatalf("cannot stop unknown microservice '%s'", msName)
	}
	if err := ms.stop(); err != nil {
		ctx.t.Fatalf("failed to stop microservice '%s': %v", msName, err)
	}
	delete(ctx.microservices, msName)
}

// pingFromMs pings <dstAddress> from the microservice <msName>
func (ctx *testCtx) pingFromMs(msName, dstAddress string) error {
	ms, found := ctx.microservices[msName]
	if !found {
		// bug inside a test
		ctx.t.Fatalf("cannot ping from unknown microservice '%s'", msName)
	}
	return ms.ping(dstAddress)
}

// pingFromMsClb can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (ctx *testCtx) pingFromMsClb(msName, dstAddress string) func() error {
	return func() error {
		return ctx.pingFromMs(msName, dstAddress)
	}
}

// pingFromVPP pings <dstAddress> from inside the VPP.
func (ctx *testCtx) pingFromVPP(destAddress string, allowedLoss ...int) error {
	var stdout bytes.Buffer

	// run ping on VPP using vppctl
	cmd := exec.Command("vppctl", "ping", destAddress)
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return err
	}

	// parse output
	matches := vppPingRegexp.FindStringSubmatch(stdout.String())
	sent, recv, loss, err := parsePingOutput(stdout.String(), matches)
	if err != nil {
		return err
	}
	ctx.t.Logf("VPP ping %s: sent=%d, received=%d, loss=%d%%",
		destAddress, sent, recv, loss)

	maxLoss := 49 // by default at least half of the packets should ge through
	if len(allowedLoss) > 0 {
		maxLoss = allowedLoss[0]
	}
	if sent == 0 || loss > maxLoss {
		return fmt.Errorf("failed to ping '%s': %s", destAddress, matches[0])
	}
	return nil
}

// pingFromVPPClb can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (ctx *testCtx) pingFromVPPClb(destAddress string) func() error {
	return func() error {
		return ctx.pingFromVPP(destAddress)
	}
}

func (ctx *testCtx) testConnection(fromMs, toMs, dstAddr, listenAddr string, port uint16, udp bool) error {
	// TODO (run nc client and server)
	return nil
}

func (ctx *testCtx) getValueState(value proto.Message) kvs.ValueState {
	key := models.Key(value)
	status := ctx.plugins.KvScheduler.GetValueStatus(key)
	return status.GetValue().GetState()
}

func (ctx *testCtx) getValueStateByKey(key string) kvs.ValueState {
	status := ctx.plugins.KvScheduler.GetValueStatus(key)
	return status.GetValue().GetState()
}

// getValueStateClb can be used to repeatedly check value state inside the assertions
// "Eventually" and "Consistently" from Omega.
func (ctx *testCtx) getValueStateClb(value proto.Message) func() kvs.ValueState {
	return func() kvs.ValueState {
		return ctx.getValueState(value)
	}
}
