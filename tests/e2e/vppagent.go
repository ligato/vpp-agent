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
	"context"
	"errors"
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
	vppAgentDefaultImg = "ligato/vpp-agent:latest"
	vppAgentLabelKey   = "e2e.test.vppagent"
	vppAgentNamePrefix = "e2e-test-vppagent-"
)

type vppAgentOpt struct {
	Image   string
	Env     []string
	UseEtcd bool
}

type vppAgent struct {
	ctx         *TestCtx
	name        string
	container   *docker.Container
	closeWaiter docker.CloseWaiter
	*vppAgentOpt
}

func RunVppAgent(ctx *TestCtx, name string, opt *vppAgentOpt) *vppAgent {
	log := logrus.WithField("name", name)

	agentLabel := vppAgentNamePrefix + name

	opts := docker.CreateContainerOptions{
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
			},
		},
	}

	log.Debugf("starting vpp-agent: %+v", opts)

	container, err := ctx.dockerClient.CreateContainer(opts)
	if err != nil {
		ctx.t.Fatalf("failed to create vpp-agent: %v", err)
	}

	err = ctx.dockerClient.StartContainer(container.ID, nil)
	if err != nil {
		ctx.t.Fatalf("failed to start vpp-agent: %v", err)
	}
	container, err = ctx.dockerClient.InspectContainerWithOptions(docker.InspectContainerOptions{
		Context: ctx.ctx,
		ID:      container.ID,
	})
	if err != nil {
		ctx.t.Fatalf("failed to inspect vpp-agent: %v", err)
	}

	log = log.WithField("container", container.Name)
	log = log.WithField("cid", container.ID)

	closeWaiter, err := ctx.dockerClient.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    container.ID,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		Logs:         true,
		OutputStream: ctx.outputBuf,
		ErrorStream:  ctx.outputBuf,
	})
	if err != nil {
		ctx.t.Fatalf("failed to attach vpp-agent: %v", err)
	}

	log.Debugf("vpp-agent started")

	go func() {
		err := closeWaiter.Wait()
		if err != nil {
			log.Warnf("vpp-agent exited: %v", err)
		} else {
			log.Debugf("vpp-agent exited OK")
		}
	}()

	return &vppAgent{
		ctx:         ctx,
		name:        name,
		container:   container,
		closeWaiter: closeWaiter,
		vppAgentOpt: opt,
	}
}

func (agent *vppAgent) IPAddress() string {
	return agent.container.NetworkSettings.IPAddress
}

func (agent *vppAgent) PID() int {
	return agent.container.State.Pid
}

func resetVppAgents(t *testing.T, dockerClient *docker.Client) {
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

func (agent *vppAgent) removeContainer(force bool) error {
	return agent.ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:            agent.container.ID,
		Force:         force,
		RemoveVolumes: false,
	})
}

const defaultStopContainerTimeoutSec = 3

func (agent *vppAgent) Stop() error {
	err := agent.ctx.dockerClient.StopContainer(agent.container.ID, defaultStopContainerTimeoutSec)
	if errors.Is(err, &docker.NoSuchContainer{}) {
		return nil // skip remove if not found
	} else if err != nil {
		return err
	}
	return agent.removeContainer(true)
}

// Exec allows to execute command **inside** the vpp-agent - i.e. not just
// inside the network namespace of the vpp-agent, but inside the container
// as a whole.
func (agent *vppAgent) Exec(cmd string, args ...string) (string, string, error) {
	opts := docker.CreateExecOptions{
		Context:      agent.ctx.ctx,
		Container:    agent.container.ID,
		Cmd:          append([]string{cmd}, args...),
		AttachStdout: true,
		AttachStderr: true,
	}
	exec, err := agent.ctx.dockerClient.CreateExec(opts)
	if err != nil {
		agent.ctx.t.Fatalf("docker create exec: %v", err)
	}

	ctx, cancel := context.WithTimeout(agent.ctx.ctx, execTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	err = agent.ctx.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		OutputStream: &stdout,
		ErrorStream:  &stderr,
	})
	if err != nil {
		agent.ctx.t.Fatalf("start exec failed: %v", err)
	}

	if info, er := agent.ctx.dockerClient.InspectExec(exec.ID); er != nil {
		agent.ctx.t.Logf("exec inspect failed (ID %v): %v", exec.ID, er)
	} else {
		agent.ctx.logger.Printf("exec details (ID %v): %+v", exec.ID, info)
		if info.ExitCode != 0 {
			err = fmt.Errorf("exec error (exit code %v): %v", info.ExitCode, stderr.String())
		}
	}
	return stdout.String(), stderr.String(), err
}

// Ping <destAddress> from inside of the vpp-agent.
func (agent *vppAgent) Ping(targetAddr string, opts ...pingOpt) error {
	agent.ctx.t.Helper()

	params := &pingOptions{
		allowedLoss: 49, // by default at least half of the packets should get through
	}
	for _, o := range opts {
		o(params)
	}

	args := []string{"-w", "4"}
	if params.outIface != "" {
		args = append(args, "-I", params.outIface)
	}
	args = append(args, targetAddr)

	stdout, _, err := agent.Exec("ping", args...)
	if err != nil {
		return err
	}

	matches := linuxPingRegexp.FindStringSubmatch(stdout)
	sent, recv, loss, err := parsePingOutput(stdout, matches)
	if err != nil {
		return err
	}
	agent.ctx.logger.Printf("VPP-Agent ping %s: sent=%d, received=%d, loss=%d%%",
		targetAddr, sent, recv, loss)

	if sent == 0 || loss > params.allowedLoss {
		return fmt.Errorf("failed to ping '%s': %s", targetAddr, matches[0])
	}
	return nil
}

func (agent *vppAgent) LinuxInterfaceHandler() linuxcalls.NetlinkAPI {
	ns, err := netns.GetFromPid(agent.PID())
	if err != nil {
		agent.ctx.t.Fatalf("unable to get netns (PID %v)", agent.PID())
	}
	ifHandler := linuxcalls.NewNetLinkHandlerNs(ns, logging.DefaultLogger)
	return ifHandler
}
