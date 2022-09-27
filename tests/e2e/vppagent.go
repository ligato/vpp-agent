//  Copyright (c) 2020 Cisco and/or its affiliates.
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
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-errors/errors"
	"github.com/vishvananda/netns"
	"go.ligato.io/cn-infra/v2/health/statuscheck/model/status"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	ctl "go.ligato.io/vpp-agent/v3/cmd/agentctl/client"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"

	. "github.com/onsi/gomega"
)

const (
	agentImage       = "ligato/vpp-agent:latest"
	agentLabelKey    = "e2e.test.vppagent"
	agentNamePrefix  = "e2e-test-vppagent-"
	agentInitTimeout = 15 // seconds
	agentStopTimeout = 3  // seconds
)

var vppPingRegexp = regexp.MustCompile("Statistics: ([0-9]+) sent, ([0-9]+) received, ([0-9]+)% packet loss")

// Agent represents running VPP-Agent test component
type Agent struct {
	ComponentRuntime
	client ctl.APIClient
	ctx    *TestCtx
	name   string
}

// NewAgent creates and starts new VPP-Agent container
func NewAgent(ctx *TestCtx, name string, optMods ...AgentOptModifier) (*Agent, error) {
	// compute options
	opts := DefaultAgentOpt(ctx, name)
	for _, mod := range optMods {
		mod(opts)
	}

	// create struct for Agent
	agent := &Agent{
		ComponentRuntime: opts.Runtime,
		ctx:              ctx,
		name:             name,
	}

	// get runtime specific options and start agent in runtime environment
	startOpts, err := opts.RuntimeStartOptions(ctx, opts)
	if err != nil {
		return nil, errors.Errorf("can't get agent %s start option for runtime due to: %v", name, err)
	}
	err = agent.Start(startOpts)
	if err != nil {
		return nil, errors.Errorf("can't start agent %s due to: %v", name, err)
	}
	agent.client, err = ctl.NewClient(agent.IPAddress())
	if err != nil {
		return nil, errors.Errorf("can't create client for %s due to: %v", name, err)
	}

	agent.ctx.Eventually(agent.checkReady, agentInitTimeout, checkPollingInterval).Should(Succeed())
	if opts.InitialResync {
		agent.Sync()
	}
	return agent, nil
}

func (agent *Agent) Stop(options ...interface{}) error {
	if err := agent.ComponentRuntime.Stop(options); err != nil {
		// not additionally cleaning up after attempting to stop test topology component because
		// it would lock access to further inspection of this component (i.e. why it won't stop)
		return err
	}
	// cleanup
	if err := agent.client.Close(); err != nil {
		return err
	}
	delete(agent.ctx.agents, agent.name)
	return nil
}

// AgentStartOptionsForContainerRuntime translates AgentOpt to options for ComponentRuntime.Start(option)
// method implemented by ContainerRuntime
func AgentStartOptionsForContainerRuntime(ctx *TestCtx, options interface{}) (interface{}, error) {
	opts, ok := options.(*AgentOpt)
	if !ok {
		return nil, errors.Errorf("expected AgentOpt but got %+v", options)
	}

	// construct vpp-agent container creation options
	agentLabel := agentNamePrefix + opts.Name
	createOpts := &docker.CreateContainerOptions{
		Context: ctx.ctx,
		Name:    agentLabel,
		Config: &docker.Config{
			Image: opts.Image,
			Labels: map[string]string{
				agentLabelKey: opts.Name,
			},
			Env:          opts.Env,
			AttachStderr: true,
			AttachStdout: true,
		},
		HostConfig: &docker.HostConfig{
			PublishAllPorts: true,
			Privileged:      true,
			PidMode:         "host",
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
				ctx.testDataDir + ":/testdata:ro",
				filepath.Join(ctx.testDataDir, "certs") + ":/etc/certs:ro",
				shareVolumeName + ":" + ctx.testShareDir,
			},
		},
	}
	if opts.ContainerOptsHook != nil {
		opts.ContainerOptsHook(createOpts)
	}
	return &ContainerStartOptions{
		ContainerOptions: createOpts,
		Pull:             false,
		AttachLogs:       true,
	}, nil
}

// TODO this is runtime specific -> integrate it into runtime concept
func removeDanglingAgents(t *testing.T, dockerClient *docker.Client) {
	// remove any running vpp-agents prior to starting a new test
	containers, err := dockerClient.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"label": {agentLabelKey},
		},
	})
	if err != nil {
		t.Fatalf("failed to list existing vpp-agents: %v", err)
	}
	for _, container := range containers {
		err = dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		if err != nil {
			t.Fatalf("failed to remove existing vpp-agents: %v", err)
		} else {
			t.Logf("removed existing vpp-agent: %s", container.Labels[agentLabelKey])
		}
	}
}

func (agent *Agent) LinuxInterfaceHandler() linuxcalls.NetlinkAPI {
	ns, err := netns.GetFromPid(agent.PID())
	if err != nil {
		agent.ctx.t.Fatalf("unable to get netns (PID %v)", agent.PID())
	}
	ifHandler := linuxcalls.NewNetLinkHandlerNs(ns, logging.DefaultLogger)
	return ifHandler
}

func (agent *Agent) Client() ctl.APIClient {
	return agent.client
}

// GenericClient provides generic client for communication with default VPP-Agent test component
func (agent *Agent) GenericClient() client.GenericClient {
	c, err := agent.client.GenericClient()
	if err != nil {
		agent.ctx.t.Fatalf("Failed to get generic VPP-agent client: %v", err)
	}
	return c
}

// GRPCConn provides GRPC client connection for communication with default VPP-Agent test component
func (agent *Agent) GRPCConn() *grpc.ClientConn {
	conn, err := agent.client.GRPCConn()
	if err != nil {
		agent.ctx.t.Fatalf("Failed to get gRPC connection: %v", err)
	}
	return conn
}

// Sync runs downstream resync and returns the list of executed operations.
func (agent *Agent) Sync() kvs.RecordedTxnOps {
	txn, err := agent.client.SchedulerResync(context.Background(), types.SchedulerResyncOptions{
		Retry: true,
	})
	if err != nil {
		agent.ctx.t.Fatalf("Downstream resync request has failed: %v", err)
	}
	if txn.Start.IsZero() {
		agent.ctx.t.Fatalf("Downstream resync returned empty transaction record: %v", txn)
	}
	return txn.Executed
}

// IsInSync checks if the agent NB config and the SB state (VPP+Linux) are in-sync.
func (agent *Agent) IsInSync() bool {
	ops := agent.Sync()
	for _, op := range ops {
		if !op.NOOP {
			return false
		}
	}
	return true
}

func (agent *Agent) checkReady() error {
	agentStatus, err := agent.client.Status(agent.ctx.ctx)
	if err != nil {
		return fmt.Errorf("query to get %s status failed: %v", agent.name, err)
	}
	agentPlugin, ok := agentStatus.PluginStatus["VPPAgent"]
	if !ok {
		return fmt.Errorf("%s plugin status missing", agent.name)
	}
	if agentPlugin.State != status.OperationalState_OK {
		return fmt.Errorf("%s status: %v", agent.name, agentPlugin.State.String())
	}
	return nil
}

// ExecVppctl returns output from vppctl for given action and arguments.
func (agent *Agent) ExecVppctl(action string, args ...string) (string, error) {
	cmd := append([]string{"-s", "/run/vpp/cli.sock", action}, args...)
	stdout, _, err := agent.ExecCmd("vppctl", cmd...)
	if err != nil {
		return "", fmt.Errorf("execute `vppctl %s` error: %v", strings.Join(cmd, " "), err)
	}
	if *debug {
		agent.ctx.t.Logf("executed (vppctl %v): %v", strings.Join(cmd, " "), stdout)
	}

	return stdout, nil
}

// PingFromVPPAsCallback can be used to ping repeatedly inside the assertions "Eventually"
// and "Consistently" from Omega.
func (agent *Agent) PingFromVPPAsCallback(destAddress string) func() error {
	return func() error {
		return agent.PingFromVPP(destAddress)
	}
}

// PingFromVPP pings <dstAddress> from inside the VPP.
func (agent *Agent) PingFromVPP(destAddress string) error {
	// run ping on VPP using vppctl
	stdout, err := agent.ExecVppctl("ping", destAddress)
	if err != nil {
		return err
	}

	// parse output
	matches := vppPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	agent.ctx.Logger.Printf("VPP ping %s: sent=%d, received=%d, loss=%d%%",
		destAddress, sent, recv, loss)

	if sent == 0 || loss >= 50 {
		return fmt.Errorf("failed to ping '%s': %s", destAddress, matches[0])
	}
	return nil
}

func (agent *Agent) getKVDump(value proto.Message, view kvs.View) []kvs.RecordedKVWithMetadata {
	model, err := models.GetModelFor(value)
	if err != nil {
		agent.ctx.t.Fatalf("Failed to get model for value %v: %v", value, err)
	}
	kvDump, err := agent.client.SchedulerDump(context.Background(), types.SchedulerDumpOptions{
		KeyPrefix: model.KeyPrefix(),
		View:      view.String(),
	})
	if err != nil {
		agent.ctx.t.Fatalf("Request to dump values failed: %v", err)
	}
	return kvDump
}

// GetValue retrieves value(s) as seen by the given view
func (agent *Agent) GetValue(value proto.Message, view kvs.View) proto.Message {
	key, err := models.GetKey(value)
	if err != nil {
		agent.ctx.t.Fatalf("Failed to get key for value %v: %v", value, err)
	}
	kvDump := agent.getKVDump(value, view)
	for _, kv := range kvDump {
		if kv.Key == key {
			return kv.Value.Message
		}
	}
	return nil
}

// GetValueMetadata retrieves metadata associated with the given value.
func (agent *Agent) GetValueMetadata(value proto.Message, view kvs.View) (metadata interface{}) {
	key, err := models.GetKey(value)
	if err != nil {
		agent.ctx.t.Fatalf("Failed to get key for value %v: %v", value, err)
	}
	kvDump := agent.getKVDump(value, view)
	for _, kv := range kvDump {
		if kv.Key == key {
			return kv.Metadata
		}
	}
	return nil
}

// NumValues returns number of values found under the given model
func (agent *Agent) NumValues(value proto.Message, view kvs.View) int {
	return len(agent.getKVDump(value, view))
}

func (agent *Agent) getValueStateByKey(key, derivedKey string) kvscheduler.ValueState {
	values, err := agent.client.SchedulerValues(context.Background(), types.SchedulerValuesOptions{
		Key: key,
	})
	if err != nil {
		agent.ctx.t.Fatalf("Request to obtain value status has failed: %v", err)
	}
	if len(values) != 1 {
		agent.ctx.t.Fatalf("Expected single value status, got status for %d values", len(values))
	}
	st := values[0]
	if st.GetValue().GetKey() != key {
		agent.ctx.t.Fatalf("Received value status for unexpected key: %v", st)
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

func (agent *Agent) GetValueState(value proto.Message) kvscheduler.ValueState {
	key := models.Key(value)
	return agent.getValueStateByKey(key, "")
}

func (agent *Agent) GetValueStateClb(value proto.Message) func() kvscheduler.ValueState {
	return func() kvscheduler.ValueState {
		return agent.GetValueState(value)
	}
}

func (agent *Agent) GetValueStateByKey(key string) kvscheduler.ValueState {
	return agent.getValueStateByKey(key, "")
}

func (agent *Agent) GetValueStateByKeyClb(key string) func() kvscheduler.ValueState {
	return func() kvscheduler.ValueState {
		return agent.GetValueStateByKey(key)
	}
}

func (agent *Agent) GetDerivedValueState(baseValue proto.Message, derivedKey string) kvscheduler.ValueState {
	key := models.Key(baseValue)
	return agent.getValueStateByKey(key, derivedKey)
}

func (agent *Agent) GetDerivedValueStateClb(baseValue proto.Message, derivedKey string) func() kvscheduler.ValueState {
	return func() kvscheduler.ValueState {
		return agent.GetDerivedValueState(baseValue, derivedKey)
	}
}
