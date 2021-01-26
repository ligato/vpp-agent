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
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-errors/errors"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/vishvananda/netns"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
)

const (
	vppAgentLabelKey               = "e2e.test.vppagent"
	vppAgentNamePrefix             = "e2e-test-vppagent-"
	defaultStopContainerTimeoutSec = 3
)

var vppPingRegexp = regexp.MustCompile("Statistics: ([0-9]+) sent, ([0-9]+) received, ([0-9]+)% packet loss")

// Agent represents running VPP-Agent test component
type Agent struct {
	ComponentRuntime
	ctx  *TestCtx
	name string
}

func startAgent(ctx *TestCtx, name string, opts *AgentOpt) (*Agent, error) {
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
	return agent, nil
}

// AgentStartOptionsForContainerRuntime translates AgentOpt to options for ComponentRuntime.Start(option)
// method implemented by ContainerRuntime
func AgentStartOptionsForContainerRuntime(ctx *TestCtx, options interface{}) (interface{}, error) {
	opts, ok := options.(*AgentOpt)
	if !ok {
		return nil, errors.Errorf("expected AgentOpt but got %+v", options)
	}

	// construct vpp-agent container creation options
	agentLabel := vppAgentNamePrefix + opts.Name
	createOpts := &docker.CreateContainerOptions{
		Context: ctx.ctx,
		Name:    agentLabel,
		Config: &docker.Config{
			Image: opts.Image,
			Labels: map[string]string{
				vppAgentLabelKey: opts.Name,
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
			"label": {vppAgentLabelKey},
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
			t.Logf("removed existing vpp-agent: %s", container.Labels[vppAgentLabelKey])
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

// ExecVppctl returns output from vppctl for given action and arguments.
func (agent *Agent) ExecVppctl(action string, args ...string) (string, error) {
	cmd := append([]string{"-s", "127.0.0.1:5002", action}, args...)
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
