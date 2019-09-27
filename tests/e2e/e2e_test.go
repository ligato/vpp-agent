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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gogo/protobuf/proto"
	"github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"

	"github.com/ligato/cn-infra/health/probe"
	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/ligato/vpp-agent/api/genericmanager"
	"github.com/ligato/vpp-agent/client"
	"github.com/ligato/vpp-agent/client/remoteclient"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linux/nsplugin/linuxcalls"
	"github.com/ligato/vpp-agent/tests/e2e/utils"
)

var (
	vppPath       = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	vppConfig     = flag.String("vpp-config", "", "VPP config file")
	vppSockAddr   = flag.String("vpp-sock-addr", "", "VPP binapi socket address")
	covPath       = flag.String("cov", "", "Path to collect coverage data")
	agentHTTPPort = flag.Int("agent-http-port", 9191, "VPP-Agent HTTP port")
	agentGrpcPort = flag.Int("agent-grpc-port", 9111, "VPP-Agent GRPC port")
	debugHTTP     = flag.Bool("debug-http", false, "Enable HTTP client debugging")

	vppPingRegexp = regexp.MustCompile("Statistics: ([0-9]+) sent, ([0-9]+) received, ([0-9]+)% packet loss")
)

const (
	agentInitTimeout     = time.Second * 15
	processExitTimeout   = time.Second * 3
	checkPollingInterval = time.Millisecond * 100
	checkTimeout         = time.Second * 6

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
}

type testCtx struct {
	t             *testing.T
	VPP           *exec.Cmd
	agent         *exec.Cmd
	dockerClient  *docker.Client
	microservices map[string]*microservice
	nsCalls       nslinuxcalls.NetworkNamespaceAPI
	httpClient    *utils.HTTPClient
	grpcConn      *grpc.ClientConn
	grpcClient    client.ConfigClient
}

func setupE2E(t *testing.T) *testCtx {
	if os.Getenv("TRAVIS") != "" {
		t.Skip("skipping test for Travis")
	}
	RegisterTestingT(t)

	SetDefaultEventuallyPollingInterval(checkPollingInterval)
	SetDefaultEventuallyTimeout(checkTimeout)

	// connect to the docker daemon
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		t.Fatalf("failed to get docker client instance from the environment variables: %v", err)
	}
	t.Logf("Using docker client endpoint: %+v", dockerClient.Endpoint())

	// make sure there are no microservices left from the previous run
	resetMicroservices(t, dockerClient)

	// check if VPP process is not running already
	assertProcessNotRunning(t, "vpp", "vpp_main", *vppPath)

	// remove binapi files from previous run
	if *vppSockAddr != "" {
		removeFile(t, *vppSockAddr)
	}
	if err := os.Mkdir("/run/vpp", 0755); err != nil && !os.IsExist(err) {
		t.Logf("mkdir failed: %v", err)
	}

	// start VPP process
	var vppArgs []string
	if *vppConfig != "" {
		vppArgs = []string{"-c", *vppConfig}
	} else {
		vppArgs = []string{vppConf}
	}
	vppCmd := startProcess(t, "VPP", *vppPath, vppArgs...)

	// start the agent
	assertProcessNotRunning(t, "vpp_agent")

	var agentArgs []string

	if *covPath != "" {
		e2eCovPath := fmt.Sprintf("%s/%d.out", *covPath, time.Now().Unix())
		agentArgs = []string{"-test.coverprofile", e2eCovPath}
	}

	agentCmd := startProcess(t, "VPP-Agent", "/vpp-agent", agentArgs...)

	// prepare HTTP client for access to REST API of the agent
	httpAddr := fmt.Sprintf(":%d", *agentHTTPPort)
	httpClient := utils.NewHTTPClient(httpAddr)

	if *debugHTTP {
		httpClient.Log = logrus.NewLogger("http-client")
		httpClient.Log.SetLevel(logging.DebugLevel)
	}

	waitUntilAgentReady(t, httpClient)

	// connect with agent via GRPC
	grpcAddr := fmt.Sprintf(":%d", *agentGrpcPort)
	grpcConn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect to VPP-agent via gRPC: %v", err)
	}
	grpcClient := remoteclient.NewClientGRPC(genericmanager.NewGenericManagerClient(grpcConn))

	// run initial resync
	syncAgent(t, httpClient)

	return &testCtx{
		t:             t,
		VPP:           vppCmd,
		dockerClient:  dockerClient,
		agent:         agentCmd,
		microservices: make(map[string]*microservice),
		nsCalls:       nslinuxcalls.NewSystemHandler(),
		httpClient:    httpClient,
		grpcConn:      grpcConn,
		grpcClient:    grpcClient,
	}
}

func (ctx *testCtx) teardownE2E() {
	ctx.t.Logf("-----------------")

	// stop all microservices
	for msName := range ctx.microservices {
		ctx.stopMicroservice(msName)
	}

	// close gRPC connection
	ctx.grpcConn.Close()

	// terminate agent
	stopProcess(ctx.t, ctx.agent, "VPP-Agent")

	// terminate VPP
	stopProcess(ctx.t, ctx.VPP, "VPP")
}

// syncAgent runs downstream resync and returns the list of executed operations.
func (ctx *testCtx) syncAgent() (executed kvs.RecordedTxnOps) {
	return syncAgent(ctx.t, ctx.httpClient)
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

// execVppctl returns output from vppctl for given action and arguments.
func (ctx *testCtx) execVppctl(action string, args ...string) (string, error) {
	var stdout bytes.Buffer
	command := append([]string{action}, args...)
	cmd := exec.Command("vppctl", command...)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("could not execute `vppctl %s`: %v", strings.Join(command, " "), err)
	}
	return stdout.String(), nil
}

func (ctx *testCtx) startMicroservice(msName string) (ms *microservice) {
	ms = createMicroservice(ctx.t, msName, ctx.dockerClient, ctx.nsCalls)
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
func (ctx *testCtx) pingFromVPP(destAddress string) error {
	// run ping on VPP using vppctl
	stdout, err := ctx.execVppctl("ping", destAddress)
	if err != nil {
		return err
	}

	// parse output
	matches := vppPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	ctx.t.Logf("VPP ping %s: sent=%d, received=%d, loss=%d%%",
		destAddress, sent, recv, loss)

	if sent == 0 || loss >= 50 {
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

/*func (ctx *testCtx) testConnection(fromMs, toMs, dstAddr, listenAddr string, port uint16, udp bool) error {
	// TODO (run nc client and server)
	return nil
}*/

func (ctx *testCtx) getValueState(value proto.Message) kvs.ValueState {
	key := models.Key(value)
	return ctx.getValueStateByKey(key)
}

func (ctx *testCtx) getValueStateByKey(key string) kvs.ValueState {
	q := fmt.Sprintf(`/scheduler/status?key=%s`, url.QueryEscape(key))
	resp, err := ctx.httpClient.GET(q)
	if err != nil {
		ctx.t.Fatalf("Request to obtain value status has failed: %v", err)
	}
	status := kvs.BaseValueStatus{}
	if err := json.Unmarshal(resp, &status); err != nil {
		ctx.t.Fatalf("Reply with value status cannot be decoded: %v", err)
	}
	if status.GetValue().GetKey() != key {
		ctx.t.Fatalf("Received value status for unexpected key: %v", status)
	}
	return status.GetValue().GetState()
}

// getValueStateClb can be used to repeatedly check value state inside the assertions
// "Eventually" and "Consistently" from Omega.
func (ctx *testCtx) getValueStateClb(value proto.Message) func() kvs.ValueState {
	return func() kvs.ValueState {
		return ctx.getValueState(value)
	}
}

func syncAgent(t *testing.T, httpClient *utils.HTTPClient) (executed kvs.RecordedTxnOps) {
	resp, err := httpClient.POST("/scheduler/downstream-resync", struct{}{})
	if err != nil {
		t.Fatalf("Downstream resync request has failed: %v", err)
	}
	txn := kvs.RecordedTxn{}
	if err := json.Unmarshal(resp, &txn); err != nil {
		t.Fatalf("Downstream resync reply cannot be decoded: %v", err)
	}
	if txn.Start.IsZero() {
		t.Fatalf("Downstream resync returned empty transaction record: %v", txn)
	}
	return txn.Executed
}

func waitUntilAgentReady(t *testing.T, httpClient *utils.HTTPClient) {
	start := time.Now()
	for {
		select {
		case <-time.After(checkPollingInterval):
			if time.Since(start) > agentInitTimeout {
				t.Fatalf("agent failed to initialize within the timeout period of %v",
					agentInitTimeout)
			}
			resp, err := httpClient.GET("/readiness")
			if err != nil {
				continue
			}
			agentStatus := probe.ExposedStatus{}
			if err := json.Unmarshal(resp, &agentStatus); err != nil {
				t.Fatalf("Agent readiness reply cannot be decoded: %v", err)
			}
			if agentStatus, ok := agentStatus.PluginStatus["VPPAgent"]; ok {
				if agentStatus.State == status.OperationalState_OK {
					t.Logf("agent ready, took %v", time.Since(start))
					return
				}
			}
		}
	}
}

func assertProcessNotRunning(t *testing.T, name string, aliases ...string) {
	processes, err := ps.Processes()
	if err != nil {
		t.Fatalf("listing processes failed: %v", err)
	}
	for _, process := range processes {
		proc := process.Executable()
		if strings.Contains(proc, name) && process.Pid() != os.Getpid() {
			t.Logf(" - found process: %+v", process)
		}
		aliases := append(aliases, name)
		for _, alias := range aliases {
			if alias == proc {
				t.Fatalf("%s is already running (PID: %v)", name, process.Pid())
			}
		}
	}
}

func startProcess(t *testing.T, name, path string, args ...string) *exec.Cmd {
	cmd := exec.Command(path)
	cmd.Args = append(cmd.Args, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// ensure that process is killed when current process exits
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting %s failed: %v", name, err)
	}
	pid := uint32(cmd.Process.Pid)
	t.Logf("%s start OK (PID: %v)", name, pid)
	return cmd
}

func stopProcess(t *testing.T, cmd *exec.Cmd, name string) {
	// terminate process
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("sending SIGTERM to %s failed: %v", name, err)
	}

	// wait until process exits
	exit := make(chan struct{})
	go func() {
		if err := cmd.Wait(); err != nil {
			t.Logf("%s process wait failed: %v", name, err)
		}
		close(exit)
	}()
	select {
	case <-exit:
		t.Logf("%s exit OK", name)

	case <-time.After(processExitTimeout):
		t.Logf("%s exit timeout", name)
		t.Logf("sending SIGKILL to %s..", name)
		if err := cmd.Process.Signal(syscall.SIGKILL); err != nil {
			t.Fatalf("sending SIGKILL to %s failed: %v", name, err)
		}
	}
}

func removeFile(t *testing.T, path string) {
	if err := os.Remove(path); err == nil {
		t.Logf("removed file %q", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("removing file %q failed: %v", path, err)
	}
}
