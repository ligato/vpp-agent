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
	"io"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/client"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

var debug = flag.Bool("debug", false, "Turn on debug mode.")

const (
	checkPollingInterval = time.Millisecond * 100
	checkTimeout         = time.Second * 6
	shareDir             = "/test-share"
	shareVolumeName      = "share-for-vpp-agent-e2e-tests"
	mainAgentName        = "agent0"

	// VPP input nodes for packet tracing (uncomment when needed)
	tapv2InputNode = "virtio-input"
	// tapv1InputNode    = "tapcli-rx"
	// afPacketInputNode = "af-packet-input"
	// memifInputNode    = "memif-input"
)

// TestCtx represents data context fur currently running test
type TestCtx struct {
	*gomega.WithT

	Agent     *Agent // the main agent (first agent in multi-agent test scenario)
	Etcd      *Etcd
	DNSServer *DNSServer

	Logger *log.Logger

	t      *testing.T
	ctx    context.Context
	cancel context.CancelFunc

	testDataDir  string
	testShareDir string

	agents        map[string]*Agent
	dockerClient  *docker.Client
	microservices map[string]*Microservice
	nsCalls       nslinuxcalls.NetworkNamespaceAPI
	vppVersion    string

	outputBuf *bytes.Buffer
}

// ComponentRuntime represents running instance of test topology component. Different implementation can
// handle test topology components in different environments (docker container, k8s pods, VMs,...)
type ComponentRuntime interface {
	CommandExecutor

	// Start starts instance of test topology component
	Start(options interface{}) error

	// Stop stops instance of test topology component
	Stop(options ...interface{}) error

	// IPAddress provides ip address for connecting to the component
	IPAddress() string

	// TODO replace PID() this with some namespace handler(that is what it is used for) because this
	//  won't help in certain runtime implementations (non-local runtimes)

	// PID provides process id of the main process in component
	PID() int
}

// CommandExecutor gives test topology components the ability to perform (linux) commands
type CommandExecutor interface {
	// ExecCmd executes command inside runtime environment
	ExecCmd(cmd string, args ...string) (stdout, stderr string, err error)
}

// Pinger gives test topology components the ability to perform pinging (pinging from them to other places)
type Pinger interface {
	CommandExecutor

	// Ping <destAddress> from inside of the container.
	Ping(destAddress string, opts ...PingOptModifier) error

	// PingAsCallback can be used to ping repeatedly inside the assertions "Eventually"
	// and "Consistently" from Omega.
	PingAsCallback(destAddress string, opts ...PingOptModifier) func() error
}

// Diger gives test topology components the ability to perform dig command (DNS-query linux tool)
type Diger interface {
	CommandExecutor

	// Dig calls linux tool "dig" that query DNS server for domain name (queryDomain) and return records associated
	// of given type (requestedInfo) associated with the domain name.
	Dig(dnsServer net.IP, queryDomain string, requestedInfo DNSRecordType) ([]net.IP, error)
}

// NewTest creates new TestCtx for given runnin test
func NewTest(t *testing.T) *TestCtx {
	g := gomega.NewWithT(t)

	g.SetDefaultEventuallyPollingInterval(checkPollingInterval)
	g.SetDefaultEventuallyTimeout(checkTimeout)

	logging.Debugf("Environ:\n%v", strings.Join(os.Environ(), "\n"))

	outputBuf := new(bytes.Buffer)
	var logW io.Writer
	if *debug {
		logW = os.Stderr // io.MultiWriter(os.Stderr, outputBuf)
	} else {
		logW = outputBuf
	}

	prefix := fmt.Sprintf("[E2E-TEST::%v] ", t.Name())
	logger := log.New(logW, prefix, log.Lshortfile|log.Lmicroseconds)

	te := &TestCtx{
		WithT:         g,
		t:             t,
		testDataDir:   os.Getenv("TESTDATA_DIR"),
		testShareDir:  shareDir,
		agents:        make(map[string]*Agent),
		microservices: make(map[string]*Microservice),
		nsCalls:       nslinuxcalls.NewSystemHandler(),
		outputBuf:     outputBuf,
		Logger:        logger,
	}
	te.ctx, te.cancel = context.WithCancel(context.Background())
	return te
}

// Setup setups the testing environment according to options
func Setup(t *testing.T, optMods ...SetupOptModifier) *TestCtx {
	testCtx := NewTest(t)

	// prepare setup options
	opts := DefaultSetupOpt(testCtx)
	for _, mod := range optMods {
		mod(opts)
	}

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
			if testCtx.Agent != nil {
				if err := testCtx.Agent.Stop(); err != nil {
					t.Logf("failed to stop vpp-agent: %v", err)
				}
			}
			if testCtx.Etcd != nil {
				if err := testCtx.Etcd.Stop(); err != nil {
					t.Logf("failed to stop etcd due to: %v", err)
				}
			}
			if testCtx.DNSServer != nil {
				if err := testCtx.DNSServer.Stop(); err != nil {
					t.Logf("failed to stop DNS server due to: %v", err)
				}
			}
		}
	}()

	// setup DNS server
	if opts.SetupDNSServer {
		testCtx.DNSServer, err = NewDNSServer(testCtx, opts.DNSOptMods...)
		testCtx.Expect(err).ShouldNot(HaveOccurred())
	}

	// setup Etcd
	if opts.SetupEtcd {
		testCtx.Etcd, err = NewEtcd(testCtx, opts.EtcdOptMods...)
		testCtx.Expect(err).ShouldNot(HaveOccurred())
	}

	// setup main VPP-Agent
	if opts.SetupAgent {
		testCtx.Agent = testCtx.StartAgent(mainAgentName, opts.AgentOptMods...)

		// fill VPP version (this depends on agentctl and that depends on agent to be set up)
		if version, err := testCtx.Agent.ExecVppctl("show version"); err != nil {
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

	return testCtx
}

// AgentInstanceName provides instance name of VPP-Agent that is created by setup by default. This name is
// used i.e. in ETCD key prefix.
func AgentInstanceName(testCtx *TestCtx) string {
	// TODO API boundaries becomes blurry as tests and support structures are in the same package and there
	// is strong temptation to misuse it and create an unmaintainable dependency mesh -> create different
	// package for test supporting files (setup/teardown/util stuff) and define clear boundaries
	if testCtx.Agent != nil {
		return testCtx.Agent.name
	}
	return mainAgentName
}

// Teardown perform test cleanup
func (test *TestCtx) Teardown() {
	if test.t.Failed() || *debug {
		defer test.dumpLog()
	}

	if test.cancel != nil {
		test.cancel()
		test.cancel = nil
	}

	// terminate all agents and close their clients
	for name, agent := range test.agents {
		if err := agent.Stop(); err != nil {
			test.t.Logf("failed to stop agent %s: %v", name, err)
		}
	}

	// stop all microservices
	for name, ms := range test.microservices {
		if err := ms.Stop(); err != nil {
			test.t.Logf("failed to stop microservice %s: %v", name, err)
		}
	}

	// terminate etcd
	if test.Etcd != nil {
		if err := test.Etcd.Stop(); err != nil {
			test.t.Logf("failed to stop ETCD: %v", err)
		}
	}

	// terminate DNS server
	if test.DNSServer != nil {
		if err := test.DNSServer.Stop(); err != nil {
			test.t.Logf("failed to stop DNS server: %v", err)
		}
	}
}

func (test *TestCtx) dumpLog() {
	output := test.outputBuf.String()
	test.outputBuf.Reset()
	test.t.Logf("OUTPUT:\n------------------\n%s\n------------------\n\n", output)
}

// VppRelease provides VPP version of VPP in default VPP-Agent test component
func (test *TestCtx) VppRelease() string {
	version := test.vppVersion
	version = strings.TrimPrefix(version, "vpp ")
	if len(version) > 5 {
		return version[1:6]
	}
	return version
}

// GenericClient provides generic client for communication with default VPP-Agent test component
func (test *TestCtx) GenericClient() client.GenericClient {
	test.t.Helper()
	return test.Agent.GenericClient()
}

// GRPCConn provides GRPC client connection for communication with default VPP-Agent test component
func (test *TestCtx) GRPCConn() *grpc.ClientConn {
	test.t.Helper()
	return test.Agent.GRPCConn()
}

// AgentInSync checks if the agent NB config and the SB state (VPP+Linux) are in-sync.
func (test *TestCtx) AgentInSync() bool {
	test.t.Helper()
	return test.Agent.IsInSync()
}

// ExecCmd executes command in agent and returns stdout, stderr as strings and error.
func (test *TestCtx) ExecCmd(cmd string, args ...string) (stdout, stderr string, err error) {
	test.t.Helper()
	return test.Agent.ExecCmd(cmd, args...)
}

// ExecVppctl returns output from vppctl for given action and arguments.
func (test *TestCtx) ExecVppctl(action string, args ...string) (string, error) {
	test.t.Helper()
	return test.Agent.ExecVppctl(action, args...)
}

// StartMicroservice starts microservice according to given options
func (test *TestCtx) StartMicroservice(name string, optMods ...MicroserviceOptModifier) *Microservice {
	test.t.Helper()

	if _, ok := test.microservices[name]; ok {
		test.t.Fatalf("microservice %s already started", name)
	}
	ms, err := NewMicroservice(test, name, test.nsCalls, optMods...)
	if err != nil {
		test.t.Fatalf("creating microservice %s failed: %v", name, err)
	}
	test.microservices[name] = ms
	return ms
}

// StopMicroservice stops microservice with given name
func (test *TestCtx) StopMicroservice(name string) {
	test.t.Helper()

	ms, found := test.microservices[name]
	if !found {
		// bug inside a test
		test.t.Logf("ERROR: cannot stop unknown microservice %s", name)
	}
	if err := ms.Stop(); err != nil {
		test.t.Logf("ERROR: stopping microservice %s failed: %v", name, err)
	}
	delete(test.microservices, name)
}

// StartAgent starts new VPP-Agent with given name and according to options
func (test *TestCtx) StartAgent(name string, optMods ...AgentOptModifier) *Agent {
	test.t.Helper()

	if _, ok := test.agents[name]; ok {
		test.t.Fatalf("agent %s already started", name)
	}
	agent, err := NewAgent(test, name, optMods...)
	if err != nil {
		test.t.Fatalf("creating agent %s failed: %v", name, err)
	}
	if test.Agent == nil {
		test.Agent = agent
	}
	test.agents[name] = agent
	return agent
}

// StopAgent stops VPP-Agent with given name
func (test *TestCtx) StopAgent(name string) {
	test.t.Helper()

	agent, found := test.agents[name]
	if !found {
		// bug inside a test
		test.t.Logf("ERROR: cannot stop unknown agent %s", name)
	}
	if err := agent.Stop(); err != nil {
		test.t.Logf("ERROR: stopping agent %s failed: %v", name, err)
	}
	if test.Agent.name == name {
		test.Agent = nil
	}
	delete(test.agents, name)
}

// GetRunningMicroservice retrieves already running microservice by its name.
func (test *TestCtx) GetRunningMicroservice(msName string) *Microservice {
	ms, found := test.microservices[msName]
	if !found {
		// bug inside a test
		test.t.Fatalf("cannot ping from unknown microservice '%s'", msName)
	}
	return ms
}

// PingFromMs pings <dstAddress> from the microservice <msName>
// Deprecated: use ctx.AlreadyRunningMicroservice(msName).Ping(dstAddress, opts...) instead (or
// ms := ctx.StartMicroservice; ms.Ping(dstAddress, opts...))
func (test *TestCtx) PingFromMs(msName, dstAddress string, opts ...PingOptModifier) error {
	test.t.Helper()
	return test.GetRunningMicroservice(msName).Ping(dstAddress, opts...)
}

// PingFromMsClb can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
// Deprecated: use ctx.AlreadyRunningMicroservice(msName).PingAsCallback(dstAddress, opts...) instead (or
// ms := ctx.StartMicroservice; ms.PingAsCallback(dstAddress, opts...))
func (test *TestCtx) PingFromMsClb(msName, dstAddress string, opts ...PingOptModifier) func() error {
	test.t.Helper()
	return test.GetRunningMicroservice(msName).PingAsCallback(dstAddress, opts...)
}

// PingFromVPP pings <dstAddress> from inside the VPP.
func (test *TestCtx) PingFromVPP(destAddress string) error {
	test.t.Helper()
	return test.Agent.PingFromVPP(destAddress)
}

// PingFromVPPClb can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (test *TestCtx) PingFromVPPClb(destAddress string) func() error {
	test.t.Helper()
	return test.Agent.PingFromVPPAsCallback(destAddress)
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
	test.Logger.Printf("%s connection <from-ms=%s, dest=%s:%d, to-ms=%s, server=%s:%d>\n",
		protocol, fromMs, toAddr, toPort, toMs, listenAddr, listenPort)
	stopPacketTrace()
	test.Logger.Printf("%s connection <from-ms=%s, dest=%s:%d, to-ms=%s, server=%s:%d> => outcome: %s\n",
		protocol, fromMs, toAddr, toPort, toMs, listenAddr, listenPort, outcome)

	return err
}

// NumValues returns number of values found under the given model
func (test *TestCtx) NumValues(value proto.Message, view kvs.View) int {
	test.t.Helper()
	return test.Agent.NumValues(value, view)
}

// GetValue retrieves value(s) as seen by the given view
func (test *TestCtx) GetValue(value proto.Message, view kvs.View) proto.Message {
	test.t.Helper()
	return test.Agent.GetValue(value, view)
}

// GetValueMetadata retrieves metadata associated with the given value.
func (test *TestCtx) GetValueMetadata(value proto.Message, view kvs.View) (metadata interface{}) {
	test.t.Helper()
	return test.Agent.GetValueMetadata(value, view)
}

func (test *TestCtx) GetValueState(value proto.Message) kvscheduler.ValueState {
	test.t.Helper()
	return test.Agent.GetValueState(value)
}

func (test *TestCtx) GetValueStateByKey(key string) kvscheduler.ValueState {
	test.t.Helper()
	return test.Agent.GetValueStateByKey(key)
}

func (test *TestCtx) GetDerivedValueState(baseValue proto.Message, derivedKey string) kvscheduler.ValueState {
	test.t.Helper()
	return test.Agent.GetDerivedValueState(baseValue, derivedKey)
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

func (test *TestCtx) startPacketTrace(nodes ...string) (stopTrace func()) {
	const tracePacketsMax = 100
	for i, node := range nodes {
		if i == 0 {
			_, err := test.Agent.ExecVppctl("clear trace")
			if err != nil {
				test.t.Errorf("Failed to clear the packet trace: %v", err)
			}
		}
		_, err := test.Agent.ExecVppctl("trace add", fmt.Sprintf("%s %d", node, tracePacketsMax))
		if err != nil {
			test.t.Errorf("Failed to add packet trace for node '%s': %v", node, err)
		}
	}
	return func() {
		if len(nodes) == 0 {
			return
		}
		traces, err := test.Agent.ExecVppctl("show trace")
		if err != nil {
			test.t.Errorf("Failed to show packet trace: %v", err)
			return
		}
		test.Logger.Printf("Packet trace:\n%s\n", traces)
	}
}

func supportsLinuxVRF() bool {
	if os.Getenv("GITHUB_WORKFLOW") != "" {
		// Linux VRFs are not enabled by default in the github workflow runners
		// Notes:
		// generally, run this to check system support for VRFs:
		//  	modinfo vrf
		// in the container, you can check if kernel module for VRFs is loaded:
		//  	ls /sys/module/vrf
		// TODO: figure out how to enable support for linux VRFs
		return false
	}
	if os.Getenv("TRAVIS") != "" {
		// Linux VRFs are seemingly not supported on Ubuntu Xenial, which is used in Travis CI to run the tests.
		// TODO: remove once we upgrade to Ubuntu Bionic or newer
		return false
	}
	return true
}
