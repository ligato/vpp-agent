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
	"regexp"
	"strings"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"go.ligato.io/cn-infra/v2/health/statuscheck/model/status"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/grpc"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentctl "go.ligato.io/vpp-agent/v3/cmd/agentctl/client"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

var (
	covPath       = flag.String("cov", "", "Path to collect coverage data")
	agentHTTPPort = flag.Int("agent-http-port", 9191, "VPP-Agent HTTP port")
	agentGrpcPort = flag.Int("agent-grpc-port", 9111, "VPP-Agent GRPC port")
	debug         = flag.Bool("debug", false, "Turn on debug mode.")
)

const (
	agentInitTimeout     = time.Second * 15
	processExitTimeout   = time.Second * 3
	checkPollingInterval = time.Millisecond * 100
	checkTimeout         = time.Second * 6
	defaultTestShareDir  = "/test-share"
	shareVolumeName      = "share-for-vpp-agent-e2e-tests"
	nameOfDefaultAgent   = "agent0"

	// VPP input nodes for packet tracing (uncomment when needed)
	tapv2InputNode = "virtio-input"
	//tapv1InputNode    = "tapcli-rx"
	//afPacketInputNode = "af-packet-input"
	//memifInputNode    = "memif-input"
)

type TestCtx struct {
	Etcd      *EtcdContainer
	DNSServer *DNSContainer

	t      *testing.T
	ctx    context.Context
	cancel context.CancelFunc

	testDataDir  string
	testShareDir string

	agent         *agent
	agents        map[string]*agent
	dockerClient  *docker.Client
	agentClient   agentctl.APIClient
	microservices map[string]*microservice
	nsCalls       nslinuxcalls.NetworkNamespaceAPI
	vppVersion    string

	outputBuf *bytes.Buffer
	logger    *log.Logger
}

func NewTest(t *testing.T) *TestCtx {
	RegisterTestingT(t)
	// TODO: Do not use global test registration.
	//  It is now deprecated and you should use NewWithT() instead.
	//g := NewWithT(t)

	logrus.Debugf("Environ:\n%v", strings.Join(os.Environ(), "\n"))

	SetDefaultEventuallyPollingInterval(checkPollingInterval)
	SetDefaultEventuallyTimeout(checkTimeout)

	outputBuf := new(bytes.Buffer)
	logger := log.New(outputBuf, "e2e-test: ", log.Lshortfile|log.Lmicroseconds)

	te := &TestCtx{
		t:             t,
		testDataDir:   os.Getenv("TESTDATA_DIR"),
		testShareDir:  defaultTestShareDir,
		agents:        make(map[string]*agent),
		microservices: make(map[string]*microservice),
		nsCalls:       nslinuxcalls.NewSystemHandler(),
		outputBuf:     outputBuf,
		logger:        logger,
	}
	te.ctx, te.cancel = context.WithCancel(context.Background())
	return te
}

func Setup(t *testing.T, options ...SetupOptModifier) *TestCtx {
	// prepare setup options
	opt := DefaultSetupOpt()
	for _, optModifier := range options {
		optModifier(opt)
	}

	testCtx := NewTest(t)

	// connect to the docker daemon
	var err error
	testCtx.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		t.Fatalf("failed to get docker client instance from the environment variables: %v", err)
	}
	if *debug {
		t.Logf("Using docker client endpoint: %+v", testCtx.dockerClient.Endpoint())
	}

	// make sure there are no containers left from the previous run
	removeDanglingAgents(t, testCtx.dockerClient)
	removeDanglingMicroservices(t, testCtx.dockerClient)

	// if setupE2E fails we need to stop started containers
	defer func() {
		if testCtx.t.Failed() || *debug {
			testCtx.dumpLog()
		}
		if testCtx.t.Failed() {
			if testCtx.agent != nil {
				if err := testCtx.agent.terminate(); err != nil {
					t.Logf("failed to terminate vpp-agent: %v", err)
				}
			}
			if testCtx.Etcd != nil {
				if err := testCtx.Etcd.terminate(); err != nil {
					t.Logf("failed to terminate etcd due to: %v", err)
				}
			}
			if testCtx.DNSServer != nil {
				if err := testCtx.DNSServer.terminate(); err != nil {
					t.Logf("failed to terminate DNS server due to: %v", err)
				}
			}
		}
	}()

	// setup DNS server
	if opt.SetupDNSServer {
		testCtx.DNSServer, err = NewDNSContainer(testCtx, extractDNSOptions(opt))
		Expect(err).ShouldNot(HaveOccurred())
	}

	// setup Etcd
	if opt.SetupEtcd {
		testCtx.Etcd, err = NewEtcdContainer(testCtx, extractEtcdOptions(opt))
		Expect(err).ShouldNot(HaveOccurred())
	}

	if opt.SetupAgent {
		SetupVPPAgent(testCtx, extractAgentOptions(opt))
	}

	return testCtx
}

func SetupVPPAgent(testCtx *TestCtx, opts ...AgentOptModifier) {
	// prepare options
	options := DefaultAgentOpt()
	for _, optionsModifier := range opts {
		optionsModifier(options)
	}

	// start agent container
	testCtx.agent = testCtx.StartAgent(nameOfDefaultAgent, opts...) // not passing prepared options due to public visibility

	// interact with the agent using the client from agentctl
	agentAddr := testCtx.agent.IPAddress()
	var err error
	testCtx.agentClient, err = agentctl.NewClientWithOpts(
		agentctl.WithHost(agentAddr),
		agentctl.WithGrpcPort(*agentGrpcPort),
		agentctl.WithHTTPPort(*agentHTTPPort))
	if err != nil {
		testCtx.t.Fatalf("Failed to create VPP-agent client: %v", err)
	}

	// wait to agent to start properly
	Eventually(testCtx.checkAgentReady, agentInitTimeout, checkPollingInterval).Should(Succeed())

	// run initial resync
	if !options.NoManualInitialResync {
		testCtx.syncAgent()
	}

	// fill VPP version (this depends on agentctl and that depends on agent to be set up)
	if version, err := testCtx.ExecVppctl("show version"); err != nil {
		testCtx.t.Fatalf("Retrieving VPP version via vppctl failed: %v", err)
	} else {
		versionParts := strings.SplitN(version, " ", 3)
		if len(versionParts) > 1 {
			testCtx.vppVersion = version
			testCtx.t.Logf("VPP version: %v", testCtx.vppVersion)
		} else {
			testCtx.t.Logf("invalid VPP version: %q", version)
		}
	}
}

// AgentInstanceName provides instance name of VPP-Agent that is created by setup by default. This name is
// used i.e. in ETCD key prefix.
func AgentInstanceName(testCtx *TestCtx) string {
	//TODO API boundaries becomes blurry as tests and support structures are in the same package and there
	// is strong temptation to misuse it and create an unmaintainable dependency mesh -> create different
	// package for test supporting files (setup/teardown/util stuff) and define clear boundaries
	if testCtx.agent != nil {
		return testCtx.agent.name
	}
	return nameOfDefaultAgent
}

func (test *TestCtx) Teardown() {
	if test.t.Failed() || *debug {
		defer test.dumpLog()
	}

	if test.cancel != nil {
		test.cancel()
		test.cancel = nil
	}

	// stop all microservices
	for msName := range test.microservices {
		test.StopMicroservice(msName)
	}

	// close the agentctl client
	if err := test.agentClient.Close(); err != nil {
		test.t.Logf("closing the client failed: %v", err)
	}

	// stop agent
	if test.agent != nil {
		if err := test.agent.terminate(); err != nil {
			test.t.Logf("failed to terminate vpp-agent: %v", err)
		}
	}

	// terminate etcd
	if test.Etcd != nil {
		if err := test.Etcd.terminate(); err != nil {
			test.t.Logf("failed to terminate ETCD: %v", err)
		}
	}

	// terminate DNS server
	if test.DNSServer != nil {
		if err := test.DNSServer.terminate(); err != nil {
			test.t.Logf("failed to terminate DNS server: %v", err)
		}
	}
}

func (test *TestCtx) dumpLog() {
	output := test.outputBuf.String()
	test.outputBuf.Reset()
	test.t.Logf("OUTPUT:\n-----------------\n%s\n------------------\n\n", output)
}

func (test *TestCtx) VppRelease() string {
	version := test.vppVersion
	if len(version) > 5 {
		return version[1:6]
	}
	return test.vppVersion
}

func (test *TestCtx) GenericClient() client.GenericClient {
	c, err := test.agentClient.GenericClient()
	if err != nil {
		test.t.Fatalf("Failed to get generic VPP-agent client: %v", err)
	}
	return c
}

func (test *TestCtx) GRPCConn() *grpc.ClientConn {
	conn, err := test.agentClient.GRPCConn()
	if err != nil {
		test.t.Fatalf("Failed to get gRPC connection: %v", err)
	}
	return conn
}

// syncAgent runs downstream resync and returns the list of executed operations.
func (test *TestCtx) syncAgent() (executed kvs.RecordedTxnOps) {
	txn, err := test.agentClient.SchedulerResync(context.Background(), types.SchedulerResyncOptions{
		Retry: true,
	})
	if err != nil {
		test.t.Fatalf("Downstream resync request has failed: %v", err)
	}
	if txn.Start.IsZero() {
		test.t.Fatalf("Downstream resync returned empty transaction record: %v", txn)
	}
	return txn.Executed
}

// AgentInSync checks if the agent NB config and the SB state (VPP+Linux)
// are in-sync.
func (test *TestCtx) AgentInSync() bool {
	ops := test.syncAgent()
	for _, op := range ops {
		if !op.NOOP {
			return false
		}
	}
	return true
}

// ExecCmd executes command in agent and returns stdout, stderr as strings and error.
func (test *TestCtx) ExecCmd(cmd string, args ...string) (stdout, stderr string, err error) {
	test.t.Helper()

	stdout, err = test.agent.execCmd(cmd, args...)
	test.logger.Printf("exec: '%s %s':\nstdout: %v",
		cmd, strings.Join(args, " "), stdout)
	if err != nil {
		logging.Errorf("exec cmd failed: %v", err)
		return
	}

	return
}

// ExecVppctl returns output from vppctl for given action and arguments.
func (test *TestCtx) ExecVppctl(action string, args ...string) (string, error) {
	test.t.Helper()

	cmd := append([]string{"-s", "127.0.0.1:5002", action}, args...)

	stdout, _, err := test.ExecCmd("vppctl", cmd...)
	if err != nil {
		return "", fmt.Errorf("execute `vppctl %s` error: %v", strings.Join(cmd, " "), err)
	}
	if *debug {
		test.t.Logf("executed (vppctl %v): %v", strings.Join(cmd, " "), stdout)
	}

	return stdout, nil
}

func (test *TestCtx) StartMicroservice(name string) *microservice {
	test.t.Helper()

	if _, ok := test.microservices[name]; ok {
		test.t.Fatalf("microservice %q already started", name)
	}
	ms, err := createMicroservice(test, name, test.dockerClient, test.nsCalls)
	if err != nil {
		test.t.Fatalf("creating microservice %q failed: %v", name, err)
	}
	test.microservices[name] = ms
	return ms
}

func (test *TestCtx) StopMicroservice(name string) {
	test.t.Helper()

	ms, found := test.microservices[name]
	if !found {
		// bug inside a test
		test.t.Logf("ERROR: cannot stop unknown microservice %q", name)
	}

	if err := ms.terminate(); err != nil {
		test.t.Logf("ERROR: stopping/removing microservice %q failed: %v", name, err)
	}
	delete(test.microservices, name)
}

func (test *TestCtx) StartAgent(name string, opts ...AgentOptModifier) *agent {
	test.t.Helper()

	if _, ok := test.agents[name]; ok {
		test.t.Fatalf("agent %q already started", name)
	}

	// prepare agent options
	opt := DefaultAgentOpt()
	for _, optModifier := range opts {
		optModifier(opt)
	}
	opt.Env = append(opt.Env, "MICROSERVICE_LABEL="+name)

	agent, err := startAgent(test, name, opt)
	if err != nil {
		test.t.Fatalf("creating agent %q failed: %v", name, err)
	}
	if test.agent == nil {
		test.agent = agent
	}
	test.agents[name] = agent

	return agent
}

func (test *TestCtx) StopAgent(name string) {
	test.t.Helper()

	agent, found := test.agents[name]
	if !found {
		// bug inside a test
		test.t.Logf("ERROR: cannot stop unknown agent %q", name)
	}
	if err := agent.terminate(); err != nil {
		test.t.Logf("ERROR: terminating agent %q failed: %v", name, err)
	}
	if test.agent.name == name {
		test.agent = nil
	}
	delete(test.agents, name)
}

// PingFromMs pings <dstAddress> from the microservice <msName>
func (test *TestCtx) PingFromMs(msName, dstAddress string, opts ...pingOpt) error {
	test.t.Helper()
	ms, found := test.microservices[msName]
	if !found {
		// bug inside a test
		test.t.Fatalf("cannot ping from unknown microservice '%s'", msName)
	}
	return ms.ping(dstAddress, opts...)
}

// PingFromMsClb can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (test *TestCtx) PingFromMsClb(msName, dstAddress string, opts ...pingOpt) func() error {
	return func() error {
		return test.PingFromMs(msName, dstAddress, opts...)
	}
}

var (
	vppPingRegexp = regexp.MustCompile("Statistics: ([0-9]+) sent, ([0-9]+) received, ([0-9]+)% packet loss")
)

// PingFromVPP pings <dstAddress> from inside the VPP.
func (test *TestCtx) PingFromVPP(destAddress string) error {
	test.t.Helper()

	// run ping on VPP using vppctl
	stdout, err := test.ExecVppctl("ping", destAddress)
	if err != nil {
		return err
	}

	// parse output
	matches := vppPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	test.logger.Printf("VPP ping %s: sent=%d, received=%d, loss=%d%%",
		destAddress, sent, recv, loss)

	if sent == 0 || loss >= 50 {
		return fmt.Errorf("failed to ping '%s': %s", destAddress, matches[0])
	}
	return nil
}

// PingFromVPPClb can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (test *TestCtx) PingFromVPPClb(destAddress string) func() error {
	return func() error {
		return test.PingFromVPP(destAddress)
	}
}

func (test *TestCtx) TestConnection(
	fromMs, toMs, toAddr, listenAddr string,
	toPort, listenPort uint16, udp bool,
	traceVPPNodes ...string,
) error {
	test.t.Helper()

	const (
		connTimeout    = 3 * time.Second
		srvExitTimeout = 500 * time.Millisecond
		reqData        = "Hi server!"
		respData       = "Hi client!"
	)

	clientMs, found := test.microservices[fromMs]
	if !found {
		// bug inside a test
		test.t.Fatalf("client microservice %q not found", fromMs)
	}
	serverMs, found := test.microservices[toMs]
	if !found {
		// bug inside a test
		test.t.Fatalf("server microservice %q not found", toMs)
	}

	serverAddr := fmt.Sprintf("%s:%d", listenAddr, listenPort)
	clientAddr := fmt.Sprintf("%s:%d", toAddr, toPort)

	srvRet := make(chan error, 1)
	srvCtx, cancelSrv := context.WithCancel(context.Background())
	runServer := func() {
		defer close(srvRet)
		if udp {
			simpleUDPServer(srvCtx, serverMs, serverAddr, reqData, respData, srvRet)
		} else {
			simpleTCPServer(srvCtx, serverMs, serverAddr, reqData, respData, srvRet)
		}
	}

	clientRet := make(chan error, 1)
	runClient := func() {
		defer close(clientRet)
		if udp {
			simpleUDPClient(clientMs, clientAddr,
				reqData, respData, connTimeout, clientRet)
		} else {
			simpleTCPClient(clientMs, clientAddr,
				reqData, respData, connTimeout, clientRet)
		}
	}

	stopPacketTrace := test.startPacketTrace(traceVPPNodes...)

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
	test.logger.Printf("%s connection <from-ms=%s, dest=%s:%d, to-ms=%s, server=%s:%d>\n",
		protocol, fromMs, toAddr, toPort, toMs, listenAddr, listenPort)
	stopPacketTrace()
	test.logger.Printf("%s connection <from-ms=%s, dest=%s:%d, to-ms=%s, server=%s:%d> => outcome: %s\n",
		protocol, fromMs, toAddr, toPort, toMs, listenAddr, listenPort, outcome)

	return err
}

func (test *TestCtx) GetValueState(value proto.Message) kvscheduler.ValueState {
	key := models.Key(value)
	return test.getValueStateByKey(key, "")
}

func (test *TestCtx) GetValueStateByKey(key string) kvscheduler.ValueState {
	return test.getValueStateByKey(key, "")
}

func (test *TestCtx) GetDerivedValueState(baseValue proto.Message, derivedKey string) kvscheduler.ValueState {
	key := models.Key(baseValue)
	return test.getValueStateByKey(key, derivedKey)
}

func (test *TestCtx) getValueStateByKey(key, derivedKey string) kvscheduler.ValueState {
	values, err := test.agentClient.SchedulerValues(context.Background(), types.SchedulerValuesOptions{
		Key: key,
	})
	if err != nil {
		test.t.Fatalf("Request to obtain value status has failed: %v", err)
	}
	if len(values) != 1 {
		test.t.Fatalf("Expected single value status, got status for %d values", len(values))
	}
	st := values[0]
	if st.GetValue().GetKey() != key {
		test.t.Fatalf("Received value status for unexpected key: %v", st)
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
func (test *TestCtx) GetValueStateClb(value proto.Message) func() kvscheduler.ValueState {
	return func() kvscheduler.ValueState {
		return test.GetValueState(value)
	}
}

// GetDerivedValueStateClb can be used to repeatedly check derived value state inside
// the assertions "Eventually" and "Consistently" from Omega.
func (test *TestCtx) GetDerivedValueStateClb(baseValue proto.Message, derivedKey string) func() kvscheduler.ValueState {
	return func() kvscheduler.ValueState {
		return test.GetDerivedValueState(baseValue, derivedKey)
	}
}

// GetValueMetadata retrieves metadata associated with the given value.
func (test *TestCtx) GetValueMetadata(value proto.Message, view kvs.View) (metadata interface{}) {
	key, err := models.GetKey(value)
	if err != nil {
		test.t.Fatalf("Failed to get key for value %v: %v", value, err)
	}
	model, err := models.GetModelFor(value)
	if err != nil {
		test.t.Fatalf("Failed to get model for value %v: %v", value, err)
	}
	kvDump, err := test.agentClient.SchedulerDump(context.Background(), types.SchedulerDumpOptions{
		KeyPrefix: model.KeyPrefix(),
		View:      view.String(),
	})
	if err != nil {
		test.t.Fatalf("Request to dump values failed: %v", err)
	}
	for _, kv := range kvDump {
		if kv.Key == key {
			return kv.Metadata
		}
	}
	return nil
}

func (test *TestCtx) startPacketTrace(nodes ...string) (stopTrace func()) {
	const tracePacketsMax = 100
	for i, node := range nodes {
		if i == 0 {
			_, err := test.ExecVppctl("clear trace")
			if err != nil {
				test.t.Errorf("Failed to clear the packet trace: %v", err)
			}
		}
		_, err := test.ExecVppctl("trace add", fmt.Sprintf("%s %d", node, tracePacketsMax))
		if err != nil {
			test.t.Errorf("Failed to add packet trace for node '%s': %v", node, err)
		}
	}
	return func() {
		if len(nodes) == 0 {
			return
		}
		traces, err := test.ExecVppctl("show trace")
		if err != nil {
			test.t.Errorf("Failed to show packet trace: %v", err)
			return
		}
		test.logger.Printf("Packet trace:\n%s\n", traces)
	}
}

func (test *TestCtx) checkAgentReady() error {
	agentStatus, err := test.agentClient.Status(test.ctx)
	if err != nil {
		return fmt.Errorf("query to get agent status failed: %v", err)
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
