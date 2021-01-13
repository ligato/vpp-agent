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
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
)

const (
	vppAgentLabelKey               = "e2e.test.vppagent"
	vppAgentNamePrefix             = "e2e-test-vppagent-"
	defaultStopContainerTimeoutSec = 3
)

type agent struct {
	*Container
	*AgentOpt
	name string
}

func startAgent(ctx *TestCtx, name string, opt *AgentOpt) (*agent, error) {
	agent := &agent{
		Container: &Container{
			ctx:         ctx,
			logIdentity: "Agent " + name,
			stopTimeout: defaultStopContainerTimeoutSec,
		},
		AgentOpt: opt,
		name:     name,
	}

	// construct vpp-agent container creation options
	log := logrus.WithField("name", name)
	agentLabel := vppAgentNamePrefix + name
	opts := &docker.CreateContainerOptions{
		Context: ctx.ctx,
		Name:    agentLabel,
		Config: &docker.Config{
			Image: opt.Image,
			Labels: map[string]string{
				vppAgentLabelKey: name,
			},
			Env:          opt.Env,
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
	if opt.ContainerOptsHook != nil {
		log.Debugf("calling container opts hook")
		opt.ContainerOptsHook(opts)
	}

	// create/start vpp-agent container (+add logging)
	log.Debugf("starting vpp-agent: %+v", *opts)
	container, err := agent.create(opts, false) // the image is builded locally -> no pull
	if err != nil {
		return nil, fmt.Errorf("failed to create vpp-agent: %v", err)
	}
	if err = agent.start(); err != nil {
		return nil, fmt.Errorf("failed to start vpp-agent: %v", err)
	}
	if err = agent.attachLoggingToContainer(ctx.outputBuf); err != nil {
		return nil, fmt.Errorf("can't attach logging to vpp-agent due to: %v", err)
	}
	log = log.WithField("container", container.Name)
	log = log.WithField("cid", container.ID)
	log.Debugf("vpp-agent started")

	return agent, nil
}

func (agent *agent) IPAddress() string {
	return agent.container.NetworkSettings.IPAddress
}

func (agent *agent) PID() int {
	return agent.container.State.Pid
}

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

func (agent *agent) LinuxInterfaceHandler() linuxcalls.NetlinkAPI {
	ns, err := netns.GetFromPid(agent.PID())
	if err != nil {
		agent.ctx.t.Fatalf("unable to get netns (PID %v)", agent.PID())
	}
	ifHandler := linuxcalls.NewNetLinkHandlerNs(ns, logging.DefaultLogger)
	return ifHandler
}
