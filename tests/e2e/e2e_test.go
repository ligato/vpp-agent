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

	"github.com/fsouza/go-dockerclient"
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/health/probe"
	"github.com/mitchellh/go-ps"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/ligato/vpp-agent/api/genericmanager"
	"github.com/ligato/vpp-agent/client"
	"github.com/ligato/vpp-agent/client/remoteclient"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linux/nsplugin/linuxcalls"
)

var (
	vppPath       = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	vppConfig     = flag.String("vpp-config", "", "VPP config file")
	vppSockAddr   = flag.String("vpp-sock-addr", "", "VPP binapi socket address")
	agentHTTPPort = flag.Int("agent-http-port", 9191, "VPP-Agent HTTP port")
	agentGrpcPort = flag.Int("agent-grpc-port", 9111, "VPP-Agent GRPC port")

	vppPingRegexp = regexp.MustCompile("Statistics: ([0-9]+) sent, ([0-9]+) received, ([0-9]+)% packet loss")
)

const (
	agentInitTimeout   = time.Second * 15
	processExitTimeout = time.Second * 3

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

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
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
	vppCmd := startProcess(t, "VPP", nil, os.Stdout, os.Stderr, *vppPath, vppArgs...)

	// start the agent
	assertProcessNotRunning(t, "vpp_agent")
	agentCmd := startProcess(t, "VPP-Agent", nil, os.Stdout, os.Stderr, "/vpp-agent")

	// prepare HTTP client for access to REST API of the agent
	httpAddr := fmt.Sprintf(":%d", *agentHTTPPort)
	httpClient := utils.NewHTTPClient(httpAddr)

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

func (ctx *testCtx) testConnection(fromMs, toMs, toAddr, listenAddr string,
	toPort, listenPort uint16, udp bool, traceVPPNodes ...string) error {

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
		// now that server exited, mark context as done
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
	fmt.Printf(
		"%s connection <from-ms=%s, dest=%s:%d, to-ms=%s, server=%s:%d> packet trace:\n",
		protocol, fromMs, toAddr, toPort, toMs, listenAddr, listenPort)
	stopPacketTrace()
	fmt.Printf(
		"%s connection <from-ms=%s, dest=%s:%d, to-ms=%s, server=%s:%d> outcome: %s\n",
		protocol, fromMs, toAddr, toPort, toMs, listenAddr, listenPort, outcome)
	return err
}

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

func (ctx *testCtx) startPacketTrace(nodes ...string) (stopTrace func()) {
	const maxPackets = 100
	for i, node := range nodes {
		if i == 0 {
			_, err := ctx.execVppctl("clear trace")
			if err != nil {
				ctx.t.Fatalf("Failed to clear the packet trace: %v", err)
			}
		}
		_, err := ctx.execVppctl("trace add", fmt.Sprintf("%s %d", node, maxPackets))
		if err != nil {
			ctx.t.Fatalf("Failed to add packet trace for node '%s': %v", node, err)
		}
	}

	return func() {
		if len(nodes) == 0 {
			return
		}
		stdout, err := ctx.execVppctl("show trace")
		if err != nil {
			ctx.t.Fatalf("Failed to show packet trace: %v", err)
		}
		fmt.Println(stdout)
	}
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

func waitUntilAgentReady(t *testing.T, httpClient *utils.HTTPClient) {
	start := time.Now()
	for {
		select {
		case <-time.After(100 * time.Millisecond):
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

func startProcess(t *testing.T, name string, stdin io.Reader, stdout, stderr io.Writer,
	path string, args ...string) *exec.Cmd {

	cmd := exec.Command(path)
	cmd.Args = append(cmd.Args, args...)
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stderr = stdout
	}
	if stderr != nil {
		cmd.Stdout = stderr
	}

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
	case err = <-commRv:
		done <- err
	}
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
	case err = <-commRv:
		done <- err
	}
}

func simpleTCPClient(ms *microservice, addr string, reqMsg, expRespMsg string, timeout time.Duration, done chan<- error) {
	// try to connect with the server
	newConn := make(chan connectionRequest, 1)
	go func() {
		// move to the network namespace from which the connection should be initiated
		exitNetNs := ms.enterNetNs()
		defer exitNetNs()
		conn, err := net.Dial("tcp", addr)
		newConn <- connectionRequest{conn: conn, err: err}
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
			conn, err := net.DialUDP("udp", nil, udpAddr)
			newConn <- connectionRequest{conn: conn, err: err}
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
		_, err := fmt.Fprintf(cr.conn, reqMsg+"\n")
		if err != nil {
			commRv <- fmt.Errorf("failed to send data to the server: %v", err)
			return
		}
		// listen for reply
		message, err := bufio.NewReader(cr.conn).ReadString('\n')
		if err != nil {
			commRv <- fmt.Errorf("failed to read data from server: %v", err)
			return
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
