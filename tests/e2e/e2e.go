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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/health/probe"
	"go.ligato.io/cn-infra/v2/health/statuscheck/model/status"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/client/remoteclient"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	"go.ligato.io/vpp-agent/v3/tests/e2e/utils"
)

var (
	vppPath       = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	vppConfig     = flag.String("vpp-config", "", "VPP config file")
	vppSockAddr   = flag.String("vpp-sock-addr", "", "VPP binapi socket address")
	covPath       = flag.String("cov", "", "Path to collect coverage data")
	agentHTTPPort = flag.Int("agent-http-port", 9191, "VPP-Agent HTTP port")
	agentGrpcPort = flag.Int("agent-grpc-port", 9111, "VPP-Agent GRPC port")
	debugHTTP     = flag.Bool("debug-http", false, "Enable HTTP client debugging")
	debug         = flag.Bool("debug", false, "Turn on debug mode.")
)

const (
	etcdImage = "gcr.io/etcd-development/etcd"

	agentInitTimeout     = time.Second * 15
	processExitTimeout   = time.Second * 3
	checkPollingInterval = time.Millisecond * 100
	checkTimeout         = time.Second * 6
	execTimeout          = 10 * time.Second

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
			socket-name /run/vpp/api.sock
		}
		statseg {
			socket-name /run/vpp/stats.sock
		}
		plugins {
			plugin dpdk_plugin.so { disable }
		}
		nat {
			endpoint-dependent
		}`

	// VPP input nodes for packet tracing (uncomment when needed)
	tapv2InputNode = "virtio-input"
	//tapv1InputNode    = "tapcli-rx"
	//afPacketInputNode = "af-packet-input"
	//memifInputNode    = "memif-input"
)

type TestCtx struct {
	t             *testing.T
	VPP           *exec.Cmd
	agent         *exec.Cmd
	dockerClient  *docker.Client
	microservices map[string]*microservice
	nsCalls       nslinuxcalls.NetworkNamespaceAPI
	httpClient    *utils.HTTPClient
	grpcConn      *grpc.ClientConn
	grpcClient    client.GenericClient
	vppVersion    string
	outputBuf     *bytes.Buffer
	logger        *log.Logger
}

func Setup(t *testing.T) *TestCtx {
	// TODO: Do not use global test registration.
	//  It is now deprecated and you should use NewWithT() instead.
	RegisterTestingT(t)

	testCtx := &TestCtx{
		t:             t,
		microservices: make(map[string]*microservice),
		nsCalls:       nslinuxcalls.NewSystemHandler(),
		outputBuf:     new(bytes.Buffer),
	}
	testCtx.logger = log.New(testCtx.outputBuf, "e2e-test: ", log.Lshortfile)

	// if setupE2E fails we need to stop started processes
	defer func() {
		if testCtx.t.Failed() || *debug {
			testCtx.dumpLog()
		}
		if testCtx.t.Failed() {
			if testCtx.agent != nil {
				stopProcess(testCtx.t, testCtx.agent, "VPP-Agent")
			}
			if testCtx.VPP != nil {
				stopProcess(testCtx.t, testCtx.VPP, "VPP")
			}
		}
	}()

	var err error

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
	testCtx.VPP = startProcess(t, "VPP", nil, testCtx.outputBuf, testCtx.outputBuf, *vppPath, vppArgs...)

	SetDefaultEventuallyPollingInterval(checkPollingInterval)
	SetDefaultEventuallyTimeout(checkTimeout)

	// connect to the docker daemon
	testCtx.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		t.Fatalf("failed to get docker client instance from the environment variables: %v", err)
	}
	if *debug {
		t.Logf("Using docker client endpoint: %+v", testCtx.dockerClient.Endpoint())
	}

	// make sure there are no microservices left from the previous run
	resetMicroservices(t, testCtx.dockerClient)

	// start the agent
	assertProcessNotRunning(t, "vpp_agent")

	var agentArgs []string

	if *covPath != "" {
		e2eCovPath := fmt.Sprintf("%s/%d.out", *covPath, time.Now().Unix())
		agentArgs = []string{"-test.coverprofile", e2eCovPath}
	}
	testCtx.agent = startProcess(t, "VPP-Agent", nil, testCtx.outputBuf, testCtx.outputBuf, "/vpp-agent", agentArgs...)

	// prepare HTTP client for access to REST API of the agent
	httpAddr := fmt.Sprintf(":%d", *agentHTTPPort)
	testCtx.httpClient = utils.NewHTTPClient(httpAddr)

	if *debugHTTP {
		testCtx.httpClient.Log = logrus.NewLogger("http-client")
		testCtx.httpClient.Log.SetLevel(logging.DebugLevel)
	}

	Eventually(testCtx.checkAgentReady, agentInitTimeout, checkPollingInterval).
		Should(Succeed())

	// connect with agent via GRPC
	grpcAddr := fmt.Sprintf(":%d", *agentGrpcPort)
	testCtx.grpcConn, err = grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect to VPP-agent via gRPC: %v", err)
	}
	testCtx.grpcClient, err = remoteclient.NewClientGRPC(testCtx.grpcConn)
	if err != nil {
		t.Fatalf("Failed to create remote GRPC client: %v", err)
	}

	// run initial resync
	syncAgent(t, testCtx.httpClient)

	if version, err := testCtx.ExecVppctl("show version"); err != nil {
		t.Fatalf("Retrieving VPP version via vppctl failed: %v", err)
	} else {
		version = strings.SplitN(version, " ", 3)[1]
		testCtx.vppVersion = version
		t.Logf("VPP version: %v", testCtx.vppVersion)
	}

	return testCtx
}

func (ctx *TestCtx) VppRelease() string {
	version := ctx.vppVersion
	if len(version) > 5 {
		return version[1:6]
	}
	return ctx.vppVersion
}

func (ctx *TestCtx) GenericClient() client.GenericClient {
	return ctx.grpcClient
}

func (ctx *TestCtx) Teardown() {
	if ctx.t.Failed() || *debug {
		ctx.dumpLog()
	}

	// stop all microservices
	for msName := range ctx.microservices {
		ctx.StopMicroservice(msName)
	}

	// close gRPC connection
	if err := ctx.grpcConn.Close(); err != nil {
		ctx.t.Logf("closing grpc connection failed: %v", err)
	}

	// terminate agent
	stopProcess(ctx.t, ctx.agent, "VPP-Agent")

	// terminate VPP
	stopProcess(ctx.t, ctx.VPP, "VPP")
}

func (ctx *TestCtx) StartEtcd() string {
	err := ctx.dockerClient.PullImage(docker.PullImageOptions{
		Repository: etcdImage,
		Tag:        "latest",
	}, docker.AuthConfiguration{})
	if err != nil {
		ctx.t.Fatalf("failed to pull ETCD image: %v", err)
	}
	container, err := ctx.dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: "e2e-test-etcd",
		Config: &docker.Config{
			Env:   []string{"ETCDCTL_API=3"},
			Image: etcdImage,
			Cmd: []string{
				"/usr/local/bin/etcd",
				"--client-cert-auth",
				"--trusted-ca-file=/etc/certs/ca.pem",
				"--cert-file=/etc/certs/cert1.pem",
				"--key-file=/etc/certs/cert1-key.pem",
				"--advertise-client-urls=https://127.0.0.1:2379",
				"--listen-client-urls=https://127.0.0.1:2379",
			},
		},
		HostConfig: &docker.HostConfig{
			NetworkMode: "container:vpp-agent-e2e-test",
			Binds:       []string{os.Getenv("CERTS_PATH") + ":/etc/certs"},
		},
	})
	if err != nil {
		ctx.t.Fatalf("failed to create ETCD container: %v", err)
	}
	err = ctx.dockerClient.StartContainer(container.ID, nil)
	if err != nil {
		err = ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		if err != nil {
			ctx.t.Errorf("failed to remove ETCD container: %v", err)
		}
		ctx.t.Fatalf("failed to start ETCD container: %v", err)
	}
	ctx.t.Logf("started ETCD container %v", container.ID)
	return container.ID
}

func (ctx *TestCtx) StopEtcd(id string) {
	err := ctx.dockerClient.StopContainer(id, msStopTimeout)
	if err != nil {
		ctx.t.Logf("failed to stop ETCD container: %v", err)
	}
	err = ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    id,
		Force: true,
	})
	if err != nil {
		ctx.t.Fatalf("failed to remove ETCD container: %v", err)
	}
	ctx.t.Logf("removed ETCD container %v", id)
}

// syncAgent runs downstream resync and returns the list of executed operations.
func (ctx *TestCtx) syncAgent() (executed kvs.RecordedTxnOps) {
	return syncAgent(ctx.t, ctx.httpClient)
}

// AgentInSync checks if the agent NB config and the SB state (VPP+Linux)
// are in-sync.
func (ctx *TestCtx) AgentInSync() bool {
	ops := ctx.syncAgent()
	for _, op := range ops {
		if !op.NOOP {
			return false
		}
	}
	return true
}

// ExecCmd executes command and returns stdout, stderr as strings and error.
func (ctx *TestCtx) ExecCmd(cmd string, args ...string) (string, string, error) {
	ctx.t.Helper()
	ctx.logger.Printf("exec: '%s %s'", cmd, strings.Join(args, " "))
	var stdout, stderr bytes.Buffer
	execCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(execTimeout))
	defer cancel()
	c := exec.CommandContext(execCtx, cmd, args...)
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if *debug {
		if strings.TrimSpace(stdout.String()) != "" {
			ctx.logger.Printf(" stdout:\n%s", stdout.String())
		}
		if strings.TrimSpace(stderr.String()) != "" {
			ctx.logger.Printf(" stderr:\n%s", stderr.String())
		}
	}
	return stdout.String(), stderr.String(), err
}

// ExecVppctl returns output from vppctl for given action and arguments.
func (ctx *TestCtx) ExecVppctl(action string, args ...string) (string, error) {
	ctx.t.Helper()
	command := append([]string{action}, args...)
	stdout, _, err := ctx.ExecCmd("vppctl", command...)
	if err != nil {
		return "", fmt.Errorf("could not execute `vppctl %s`: %v", strings.Join(command, " "), err)
	}
	return stdout, nil
}

func (ctx *TestCtx) StartMicroservice(msName string) (ms *microservice) {
	ms = createMicroservice(ctx, msName, ctx.dockerClient, ctx.nsCalls)
	ctx.microservices[msName] = ms
	return ms
}

func (ctx *TestCtx) StopMicroservice(msName string) {
	ctx.t.Helper()
	ms, found := ctx.microservices[msName]
	if !found {
		// bug inside a test
		ctx.t.Logf("ERROR: cannot stop unknown microservice '%s'", msName)
	}
	if err := ms.stop(); err != nil {
		ctx.t.Logf("ERROR: stopping microservice failed '%s': %v", msName, err)
	}
	delete(ctx.microservices, msName)
}

// PingFromMs pings <dstAddress> from the microservice <msName>
func (ctx *TestCtx) PingFromMs(msName, dstAddress string, opts ...pingOpt) error {
	ctx.t.Helper()
	ms, found := ctx.microservices[msName]
	if !found {
		// bug inside a test
		ctx.t.Fatalf("cannot ping from unknown microservice '%s'", msName)
	}
	return ms.ping(dstAddress, opts...)
}

// PingFromMsClb can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (ctx *TestCtx) PingFromMsClb(msName, dstAddress string, opts ...pingOpt) func() error {
	return func() error {
		return ctx.PingFromMs(msName, dstAddress, opts...)
	}
}

var vppPingRegexp = regexp.MustCompile("Statistics: ([0-9]+) sent, ([0-9]+) received, ([0-9]+)% packet loss")

// PingFromVPP pings <dstAddress> from inside the VPP.
func (ctx *TestCtx) PingFromVPP(destAddress string) error {
	ctx.t.Helper()
	// run ping on VPP using vppctl
	stdout, err := ctx.ExecVppctl("ping", destAddress)
	if err != nil {
		return err
	}

	// parse output
	matches := vppPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	ctx.logger.Printf("VPP ping %s: sent=%d, received=%d, loss=%d%%",
		destAddress, sent, recv, loss)

	if sent == 0 || loss >= 50 {
		return fmt.Errorf("failed to ping '%s': %s", destAddress, matches[0])
	}
	return nil
}

// PingFromVPPClb can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (ctx *TestCtx) PingFromVPPClb(destAddress string) func() error {
	return func() error {
		return ctx.PingFromVPP(destAddress)
	}
}

func (ctx *TestCtx) TestConnection(fromMs, toMs, toAddr, listenAddr string,
	toPort, listenPort uint16, udp bool, traceVPPNodes ...string) error {
	ctx.t.Helper()

	const (
		connTimeout    = 3 * time.Second
		srvExitTimeout = 500 * time.Millisecond
		reqData        = "Hi server!"
		respData       = "Hi client!"
	)

	clientMs, found := ctx.microservices[fromMs]
	if !found {
		// bug inside a test
		ctx.t.Fatalf("cannot run TCP/UDP client from unknown microservice '%s'", fromMs)
	}
	serverMs, found := ctx.microservices[toMs]
	if !found {
		// bug inside a test
		ctx.t.Fatalf("cannot run TCP/UDP server inside unknown microservice '%s'", toMs)
	}

	srvRet := make(chan error, 1)
	srvCtx, cancelSrv := context.WithCancel(context.Background())
	runServer := func() {
		if udp {
			simpleUDPServer(srvCtx, serverMs, fmt.Sprintf("%s:%d", listenAddr, listenPort),
				reqData, respData, srvRet)
		} else {
			simpleTCPServer(srvCtx, serverMs, fmt.Sprintf("%s:%d", listenAddr, listenPort),
				reqData, respData, srvRet)
		}
		close(srvRet)
	}

	clientRet := make(chan error, 1)
	runClient := func() {
		if udp {
			simpleUDPClient(clientMs, fmt.Sprintf("%s:%d", toAddr, toPort),
				reqData, respData, connTimeout, clientRet)
		} else {
			simpleTCPClient(clientMs, fmt.Sprintf("%s:%d", toAddr, toPort),
				reqData, respData, connTimeout, clientRet)
		}
		close(clientRet)
	}

	stopPacketTrace := ctx.startPacketTrace(traceVPPNodes...)

	go runServer()
	go runClient()
	err := <-clientRet

	// give server some time to exit gracefully, then force it to stop
	var srvErr error
	select {
	case srvErr = <-srvRet:
		// now that the client has read all the data, the server is safe to stop
		// and close the connection
		cancelSrv()
	case <-time.After(srvExitTimeout):
		cancelSrv()
		srvErr = <-srvRet
	}
	if err == nil {
		err = srvErr
	}

	// log info about connection
	protocol := "TCP"
	if udp {
		protocol = "UDP"
	}
	outcome := "OK"
	if err != nil {
		outcome = err.Error()
	}
	ctx.logger.Printf("%s connection <from-ms=%s, dest=%s:%d, to-ms=%s, server=%s:%d>\n",
		protocol, fromMs, toAddr, toPort, toMs, listenAddr, listenPort)
	stopPacketTrace()
	ctx.logger.Printf("%s connection <from-ms=%s, dest=%s:%d, to-ms=%s, server=%s:%d> => outcome: %s\n",
		protocol, fromMs, toAddr, toPort, toMs, listenAddr, listenPort, outcome)
	return err
}

func (ctx *TestCtx) GetValueState(value proto.Message) kvscheduler.ValueState {
	key := models.Key(value)
	return ctx.getValueStateByKey(key, "")
}

func (ctx *TestCtx) GetValueStateByKey(key string) kvscheduler.ValueState {
	return ctx.getValueStateByKey(key, "")
}

func (ctx *TestCtx) GetDerivedValueState(baseValue proto.Message, derivedKey string) kvscheduler.ValueState {
	key := models.Key(baseValue)
	return ctx.getValueStateByKey(key, derivedKey)
}

func (ctx *TestCtx) getValueStateByKey(key, derivedKey string) kvscheduler.ValueState {
	q := fmt.Sprintf(`/scheduler/status?key=%s`, url.QueryEscape(key))
	resp, err := ctx.httpClient.GET(q)
	if err != nil {
		ctx.t.Fatalf("Request to obtain value status has failed: %v", err)
	}
	st := kvscheduler.BaseValueStatus{}
	if err := json.Unmarshal(resp, &st); err != nil {
		ctx.t.Fatalf("Reply with value status cannot be decoded: %v", err)
	}
	if st.GetValue().GetKey() != key {
		ctx.t.Fatalf("Received value status for unexpected key: %v", st)
	}
	if derivedKey != "" {
		for _, derVal := range st.DerivedValues {
			if derVal.Key == derivedKey {
				return derVal.State
			}
		}
		return kvscheduler.ValueState_NONEXISTENT
	}
	return st.GetValue().GetState()
}

// GetValueStateClb can be used to repeatedly check value state inside the assertions
// "Eventually" and "Consistently" from Omega.
func (ctx *TestCtx) GetValueStateClb(value proto.Message) func() kvscheduler.ValueState {
	return func() kvscheduler.ValueState {
		return ctx.GetValueState(value)
	}
}

// GetDerivedValueStateClb can be used to repeatedly check derived value state inside
// the assertions "Eventually" and "Consistently" from Omega.
func (ctx *TestCtx) GetDerivedValueStateClb(baseValue proto.Message, derivedKey string) func() kvscheduler.ValueState {
	return func() kvscheduler.ValueState {
		return ctx.GetDerivedValueState(baseValue, derivedKey)
	}
}

// GetValueMetadata retrieves metadata associated with the given value.
func (ctx *TestCtx) GetValueMetadata(value proto.Message, view kvs.View) (metadata interface{}) {
	type KeyWithMetadata struct {
		Key      string
		Metadata interface{}
	}
	key, err := models.GetKey(value)
	if err != nil {
		ctx.t.Fatalf("Failed to get key for value %v: %v", value, err)
	}
	model, err := models.GetModelFor(value)
	if err != nil {
		ctx.t.Fatalf("Failed to get model for value %v: %v", value, err)
	}
	q := fmt.Sprintf(`/scheduler/dump?key-prefix=%s&view=%v`, url.QueryEscape(model.KeyPrefix()), view)
	resp, err := ctx.httpClient.GET(q)
	if err != nil {
		ctx.t.Fatalf("Request to dump values failed: %v", err)
	}
	var kmDump []KeyWithMetadata
	if err := json.Unmarshal(resp, &kmDump); err != nil {
		ctx.t.Fatalf("Reply with dumped values cannot be decoded: %v", err)
	}
	for _, km := range kmDump {
		if km.Key == key {
			return km.Metadata
		}
	}
	return nil
}

func (ctx *TestCtx) startPacketTrace(nodes ...string) (stopTrace func()) {
	const tracePacketsMax = 100
	for i, node := range nodes {
		if i == 0 {
			_, err := ctx.ExecVppctl("clear trace")
			if err != nil {
				ctx.t.Errorf("Failed to clear the packet trace: %v", err)
			}
		}
		_, err := ctx.ExecVppctl("trace add", fmt.Sprintf("%s %d", node, tracePacketsMax))
		if err != nil {
			ctx.t.Errorf("Failed to add packet trace for node '%s': %v", node, err)
		}
	}
	return func() {
		if len(nodes) == 0 {
			return
		}
		traces, err := ctx.ExecVppctl("show trace")
		if err != nil {
			ctx.t.Errorf("Failed to show packet trace: %v", err)
			return
		}
		ctx.logger.Printf("Packet trace:\n%s\n", traces)
	}
}

func (ctx *TestCtx) dumpLog() {
	ctx.t.Logf("OUTPUT:\n-----------------\n%s\n------------------\n\n", ctx.outputBuf)
}

func syncAgent(t *testing.T, httpClient *utils.HTTPClient) (executed kvs.RecordedTxnOps) {
	resp, err := httpClient.POST("/scheduler/downstream-resync?retry=true", struct{}{})
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

func (ctx *TestCtx) checkAgentReady() error {
	resp, err := ctx.httpClient.GET("/readiness")
	if err != nil {
		return err
	}
	var agentStatus probe.ExposedStatus
	if err := json.Unmarshal(resp, &agentStatus); err != nil {
		return fmt.Errorf("decoding reply from /readiness failed: %w", err)
	}
	agent, ok := agentStatus.PluginStatus["VPPAgent"]
	if !ok {
		return fmt.Errorf("agent status missing")
	}
	if agent.State != status.OperationalState_OK {
		return fmt.Errorf("agent status: %v", agent.State.String())
	}
	return nil
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

func startProcess(t *testing.T, name string, stdin io.Reader, stdout, stderr io.Writer, path string, args ...string) *exec.Cmd {
	cmd := exec.Command(path, args...)
	cmd.Stdin = stdin
	cmd.Stderr = stdout
	cmd.Stdout = stderr
	// ensure that process is killed when current process exits
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting process %s failed: %v", name, err)
	}
	pid := uint32(cmd.Process.Pid)
	t.Logf("%s started (PID: %v)", name, pid)
	return cmd
}

func stopProcess(t *testing.T, cmd *exec.Cmd, name string) {
	// terminate process
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Logf("sending SIGTERM to %s failed: %v", name, err)
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
			t.Errorf("sending SIGKILL to %s failed: %v", name, err)
		}
	}
}

// TCP or UDP connection request
type connectionRequest struct {
	conn net.Conn
	err  error
}

func simpleTCPServer(ctx context.Context, ms *microservice, addr string, expReqMsg, respMsg string, done chan<- error) {
	// move to the network namespace where server should listen
	exitNetNs := ms.enterNetNs()
	defer exitNetNs()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		done <- err
		return
	}
	defer listener.Close()

	// accept single connection
	newConn := make(chan connectionRequest, 1)
	go func() {
		conn, err := listener.Accept()
		newConn <- connectionRequest{conn: conn, err: err}
		close(newConn)
	}()

	// wait for connection
	var cr connectionRequest
	select {
	case <-ctx.Done():
		done <- fmt.Errorf("tcp server listening on %s was canceled", addr)
		return
	case cr = <-newConn:
		if cr.err != nil {
			done <- fmt.Errorf("accept failed with: %v", cr.err)
			return
		}
		defer cr.conn.Close()
	}

	// communicate with the client
	commRv := make(chan error, 1)
	go func() {
		defer close(commRv)
		// receive message from the client
		message, err := bufio.NewReader(cr.conn).ReadString('\n')
		if err != nil {
			commRv <- fmt.Errorf("failed to read data from client: %v", err)
			return
		}
		// send response to the client
		_, err = cr.conn.Write([]byte(respMsg + "\n"))
		if err != nil {
			commRv <- fmt.Errorf("failed to send data to client: %v", err)
			return
		}
		// check if the exchanged data are as expected
		message = strings.TrimRight(message, "\n")
		if message != expReqMsg {
			commRv <- fmt.Errorf("unexpected message received from client ('%s' vs. '%s')",
				message, expReqMsg)
			return
		}
		commRv <- nil
	}()

	// wait for the message exchange to execute
	select {
	case <-ctx.Done():
		done <- fmt.Errorf("tcp server listening on %s was canceled", addr)
		return
	case err = <-commRv:
		done <- err
	}

	// do not close until client confirms reception of the message
	<-ctx.Done()
}

func simpleUDPServer(ctx context.Context, ms *microservice, addr string, expReqMsg, respMsg string, done chan<- error) {
	const maxBufferSize = 1024
	// move to the network namespace where server should listen
	exitNetNs := ms.enterNetNs()
	defer exitNetNs()

	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		done <- err
		return
	}
	defer conn.Close()

	// communicate with the client
	commRv := make(chan error, 1)
	go func() {
		defer close(commRv)
		// receive message from the client
		buffer := make([]byte, maxBufferSize)
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			commRv <- fmt.Errorf("failed to read data from client: %v", err)
			return
		}
		message := string(buffer[:n])
		// send response to the client
		_, err = conn.WriteTo([]byte(respMsg+"\n"), addr)
		if err != nil {
			commRv <- fmt.Errorf("failed to send data to client: %v", err)
			return
		}
		// check if the exchanged data are as expected
		message = strings.TrimRight(message, "\n")
		if message != expReqMsg {
			commRv <- fmt.Errorf("unexpected message received from client ('%s' vs. '%s')",
				message, expReqMsg)
			return
		}
		commRv <- nil
	}()

	// wait for the message exchange to execute
	select {
	case <-ctx.Done():
		done <- fmt.Errorf("udp server listening on %s was canceled", addr)
		return
	case err = <-commRv:
		done <- err
	}

	// do not close until client confirms reception of the message
	<-ctx.Done()
}

func simpleTCPClient(ms *microservice, addr string, reqMsg, expRespMsg string, timeout time.Duration, done chan<- error) {
	// try to connect with the server
	newConn := make(chan connectionRequest, 1)
	go func() {
		// move to the network namespace from which the connection should be initiated
		exitNetNs := ms.enterNetNs()
		defer exitNetNs()
		start := time.Now()
		for {
			conn, err := net.Dial("tcp", addr)
			if err != nil && time.Since(start) < timeout {
				time.Sleep(checkPollingInterval)
				continue
			}
			newConn <- connectionRequest{conn: conn, err: err}
			break
		}
		close(newConn)
	}()

	simpleTCPOrUDPClient(newConn, addr, reqMsg, expRespMsg, timeout, done)
}

func simpleUDPClient(ms *microservice, addr string, reqMsg, expRespMsg string, timeout time.Duration, done chan<- error) {
	// try to connect with the server
	newConn := make(chan connectionRequest, 1)
	go func() {
		// move to the network namespace from which the connection should be initiated
		exitNetNs := ms.enterNetNs()
		defer exitNetNs()
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			newConn <- connectionRequest{conn: nil, err: err}
		} else {
			start := time.Now()
			for {
				conn, err := net.DialUDP("udp", nil, udpAddr)
				if err != nil && time.Since(start) < timeout {
					time.Sleep(checkPollingInterval)
					continue
				}
				newConn <- connectionRequest{conn: conn, err: err}
				break
			}
		}
		close(newConn)
	}()

	simpleTCPOrUDPClient(newConn, addr, reqMsg, expRespMsg, timeout, done)
}

func simpleTCPOrUDPClient(newConn chan connectionRequest, addr, reqMsg, expRespMsg string,
	timeout time.Duration, done chan<- error) {

	// wait for connection
	var cr connectionRequest
	select {
	case <-time.After(timeout):
		done <- fmt.Errorf("connection to %s timed out", addr)
		return
	case cr = <-newConn:
		if cr.err != nil {
			done <- fmt.Errorf("dial failed with: %v", cr.err)
			return
		}
		defer cr.conn.Close()
	}

	// communicate with the server
	commRv := make(chan error, 1)
	go func() {
		defer close(commRv)
		// send message to the server
		_, err := cr.conn.Write([]byte(reqMsg + "\n"))
		if err != nil {
			commRv <- fmt.Errorf("failed to send data to the server: %v", err)
			return
		}
		// listen for reply
		start := time.Now()
		var message string
		for {
			message, err = bufio.NewReader(cr.conn).ReadString('\n')
			if err != nil && time.Since(start) < timeout {
				time.Sleep(checkPollingInterval)
				continue
			}
			if err != nil {
				commRv <- fmt.Errorf("failed to read data from server: %v", err)
				return
			}
			break
		}
		// check if the exchanged data are as expected
		message = strings.TrimRight(message, "\n")
		if message != expRespMsg {
			commRv <- fmt.Errorf("unexpected message received from server ('%s' vs. '%s')",
				message, expRespMsg)
			return
		}
		commRv <- nil
	}()

	// wait for the message exchange to execute
	select {
	case <-time.After(timeout):
		done <- fmt.Errorf("communication with %s timed out", addr)
	case err := <-commRv:
		done <- err
	}
}

func removeFile(t *testing.T, path string) {
	if err := os.Remove(path); err == nil {
		t.Logf("removed file %q", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("removing file %q failed: %v", path, err)
	}
}

// parseVPPTable parses table returned by one of the VPP show commands.
func parseVPPTable(table string) (parsed []map[string]string) {
	lines := strings.Split(table, "\r\n")
	if len(lines) == 0 {
		return
	}
	head := lines[0]
	rows := lines[1:]

	var columns []string
	for _, column := range strings.Split(head, " ") {
		if column != "" {
			columns = append(columns, column)
		}
	}
	for _, row := range rows {
		parsedRow := make(map[string]string)
		i := 0
		for _, cell := range strings.Split(row, " ") {
			if cell == "" {
				continue
			}
			if i >= len(columns) {
				break
			}
			parsedRow[columns[i]] = cell
			i++
		}
		if len(parsedRow) > 0 {
			parsed = append(parsed, parsedRow)
		}
	}
	return
}
