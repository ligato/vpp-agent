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
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	vppAgentImage      = "ligato/vpp-agent"
	vppAgentImageTag   = "v3.2.0"
	vppAgentLabelKey   = "e2e.test.vppagent"
	vppAgentNamePrefix = "e2e-test-vppagent-"
)

type vppAgent struct {
	ctx         *TestCtx
	name        string
	container   *docker.Container
	closeWaiter docker.CloseWaiter
}

func runVppAgent(ctx *TestCtx, name string) *vppAgent {
	agentLabel := vppAgentNamePrefix + name
	container, err := ctx.dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: agentLabel,
		Config: &docker.Config{
			Env: []string{
				"MICROSERVICE_LABEL=" + agentLabel,
				"CERTS_PATH=" + os.Getenv("CERTS_PATH"),
				"INITIAL_LOGLVL=" + os.Getenv("INITIAL_LOGLVL"),
				"ETCD_CONFIG=disabled",
			},
			Image: vppAgentImage + ":" + vppAgentImageTag,
			Labels: map[string]string{
				vppAgentLabelKey: name,
			},
			AttachStderr: true,
			AttachStdout: true,
		},
		HostConfig: &docker.HostConfig{
			// networking configured via VPP in E2E tests
			PublishAllPorts: true,
			Privileged:      true,
			PidMode:         "host",
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
				os.Getenv("CERTS_PATH") + ":/etc/certs",
			},
		},
	})
	if err != nil {
		ctx.t.Fatalf("failed to create vpp-agent '%s': %v", name, err)
	}
	err = ctx.dockerClient.StartContainer(container.ID, nil)
	if err != nil {
		ctx.t.Fatalf("failed to start vpp-agent '%s': %v", name, err)
	}
	container, err = ctx.dockerClient.InspectContainer(container.ID)
	if err != nil {
		ctx.t.Fatalf("failed to inspect vpp-agent '%s': %v", name, err)
	}
	closeWaiter, err := ctx.dockerClient.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    container.ID,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		OutputStream: ctx.outputBuf,
		ErrorStream:  ctx.outputBuf,
	})
	if err != nil {
		ctx.t.Fatalf("failed to attach vpp-agent '%s': %v", name, err)
	}
	ctx.t.Logf("vpp-agent %s started (ID: %v)", name, container.ID)
	return &vppAgent{
		ctx:         ctx,
		name:        name,
		container:   container,
		closeWaiter: closeWaiter,
	}
}

func resetVppAgents(t *testing.T, dockerClient *docker.Client) {
	// pull image for vpp-agent
	err := dockerClient.PullImage(docker.PullImageOptions{
		Repository: vppAgentImage,
		Tag:        vppAgentImageTag,
	}, docker.AuthConfiguration{})
	if err != nil {
		t.Fatalf("failed to pull image '%s:%s' for vpp-agent: %v", vppAgentImage, vppAgentImageTag, err)
	}
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

func (agent *vppAgent) stop() error {
	err := agent.ctx.dockerClient.StopContainer(agent.container.ID, 3)
	if err != nil && !errors.Is(err, &docker.NoSuchContainer{}) {
		return err
	}
	return agent.ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    agent.container.ID,
		Force: true,
	})
}

// exec allows to execute command **inside** the vpp-agent - i.e. not just
// inside the network namespace of the vpp-agent, but inside the container
// as a whole.
func (agent *vppAgent) exec(cmd string, args ...string) (output, outerr string, err error) {
	execCtx, err := agent.ctx.dockerClient.CreateExec(docker.CreateExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          append([]string{cmd}, args...),
		Container:    agent.container.ID,
	})
	if err != nil {
		agent.ctx.t.Fatalf("failed to create docker exec instance: %v", err)
	}
	var stdout, stderr bytes.Buffer
	err = agent.ctx.dockerClient.StartExec(execCtx.ID, docker.StartExecOptions{
		OutputStream: &stdout,
		ErrorStream:  &stderr,
	})
	return stdout.String(), stderr.String(), err
}

// ping <destAddress> from inside of the vpp-agent.
func (agent *vppAgent) ping(destAddress string, opts ...pingOpt) error {
	agent.ctx.t.Helper()

	params := &pingOpts{
		allowedLoss: 49, // by default at least half of the packets should get through
	}
	for _, o := range opts {
		o(params)
	}

	args := []string{"-w", "4"}
	if params.outIface != "" {
		args = append(args, "-I", params.outIface)
	}
	args = append(args, destAddress)

	stdout, _, err := agent.exec("ping", args...)
	if err != nil {
		return err
	}

	matches := linuxPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	agent.ctx.logger.Printf("VPP-Agent ping %s: sent=%d, received=%d, loss=%d%%",
		destAddress, sent, recv, loss)

	if sent == 0 || loss > params.allowedLoss {
		return fmt.Errorf("failed to ping '%s': %s", destAddress, matches[0])
	}
	return nil
}
