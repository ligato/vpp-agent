// Copyright (c) 2020 Pantheon.tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2etest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	moby "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/stringid"
	"github.com/go-errors/errors"
	"github.com/segmentio/textio"
	"go.ligato.io/cn-infra/v2/logging"
)

const containerExecTimeout = 10 * time.Second

// ContainerRuntime represents docker container environments for one component of test topology
type ContainerRuntime struct {
	ctx         *TestCtx
	container   *moby.ContainerJSON
	logIdentity string
	stopTimeout int
}

// ContainerStartOptions are options for ComponentRuntime.Start(option) method implemented by ContainerRuntime
type ContainerStartOptions struct {
	ContainerConfig *moby.ContainerCreateConfig
	Pull            bool
	AttachLogs      bool
}

// Start creates and starts container
func (c *ContainerRuntime) Start(options interface{}) error {
	// get options
	if options == nil {
		return errors.Errorf("can't start container without any information")
	}
	opts, ok := options.(*ContainerStartOptions)
	if !ok {
		return errors.Errorf("provided runtime start options "+
			"are not for container component runtime (%v)", options)
	}

	// create and start container
	id, err := c.createContainer(opts.ContainerConfig, opts.Pull)
	if err != nil {
		return errors.Errorf("can't create %s container due to: %v", c.logIdentity, err)
	}
	log := logging.DefaultLogger.WithField("name", c.logIdentity)
	log.Debugf("starting container: %+v", *opts)
	if err := c.startContainer(id); err != nil {
		return errors.Errorf("can't start %s container due to: %v", c.logIdentity, err)
	}
	c.container, err = c.inspectContainer(id)
	if err != nil {
		return err
	}
	log = log.WithField("container", c.container.Name)
	log = log.WithField("cid", stringid.TruncateID(id))
	log.Debugf("container started")

	// attach logs (using one buffer from testctx -> all logs from all containers are merged together)
	if opts.AttachLogs {
		logOutput := textio.NewPrefixWriter(c.ctx.logWriter, fmt.Sprintf("[container::%s/%v] ", c.container.Name, stringid.TruncateID(c.container.ID)))
		if err = c.attachLoggingToContainer(logOutput); err != nil {
			return errors.Errorf("can't attach logging to %s container due to: %v", c.logIdentity, err)
		}
	}
	return nil
}

// Stop stops and removes container
func (c *ContainerRuntime) Stop(options ...interface{}) error {
	if err := c.stopContainer(); err != nil {
		if errdefs.IsNotFound(err) {
			// container no longer exists -> nothing to do about container (state is the same
			// as after successful termination)
			return nil
		}
		return err
	}
	if err := c.removeContainer(); err != nil {
		return err
	}
	return nil
}

// ExecCmd executes command inside docker container
func (c *ContainerRuntime) ExecCmd(cmd string, args ...string) (string, string, error) {
	c.ctx.Logger.Printf("[container:%v] ExecCmd(%s, %v)", c.container.ID, cmd, args)

	config := moby.ExecConfig{
		Cmd:          append([]string{cmd}, args...),
		AttachStdout: true,
		AttachStderr: true,
	}
	exec, err := c.ctx.dockerClient.ContainerExecCreate(c.ctx.ctx, c.container.ID, config)
	if err != nil {
		err = errors.Errorf("failed to create docker exec for command %v due to: %v", cmd, err)
		return "", "", err
	}

	hijacked, err := c.ctx.dockerClient.ContainerExecAttach(c.ctx.ctx, exec.ID, moby.ExecStartCheck{})
	if err != nil {
		return "", "", errors.Errorf("failed to attach docker exec for command %s to container %s due to: %v", cmd, c.container.ID, err)
	}
	defer hijacked.Close()

	ctx, cancel := context.WithTimeout(c.ctx.ctx, containerExecTimeout)
	defer cancel()

	var stdoutBuf, stderrBuf bytes.Buffer

	err = c.ctx.dockerClient.ContainerExecStart(ctx, exec.ID, moby.ExecStartCheck{})
	if err != nil {
		return "", "", errors.Errorf("failed to start docker exec for command %s due to: %v", cmd, err)
	}

	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, hijacked.Reader)
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	cmdStr := fmt.Sprintf("`%s %s`", cmd, strings.Join(args, " "))

	if cmdStr != "`vppctl -s /run/vpp/cli.sock show trace`" {
		c.ctx.Logger.Printf("docker exec: %v:\nstdout(%d): %v\nstderr(%d): %v", cmdStr, len(stdout), stdout, len(stderr), stderr)
	}

	if err != nil {
		errMsg := fmt.Sprintf("exec command %v failed due to: %v", cmdStr, err)
		c.ctx.Logger.Printf(errMsg)
		err = errors.Errorf(errMsg)
		return stdout, stderr, err
	}

	if info, e := c.ctx.dockerClient.ContainerExecInspect(c.ctx.ctx, exec.ID); err != nil {
		c.ctx.t.Logf("exec inspect failed (ID %v, Cmd %s)s: %v", exec.ID, cmdStr, err)
		err = errors.Errorf("inspect exec error: %v", e)
	} else {
		c.ctx.Logger.Printf("exec details (ID %v, Cmd %s): %+v", exec.ID, cmdStr, info)
		if info.ExitCode != 0 {
			err = errors.Errorf("exec error (exit code %v): %v", info.ExitCode, stderr)
		}
	}

	return stdout, stderr, err
}

// IPAddress provides ip address for connecting to the component
func (c *ContainerRuntime) IPAddress() string {
	return c.container.NetworkSettings.IPAddress
}

// PID provides process id of the main process in component
func (c *ContainerRuntime) PID() int {
	return c.container.State.Pid
}

func (c *ContainerRuntime) createContainer(config *moby.ContainerCreateConfig, pull bool) (string, error) {
	// pull image
	if pull {
		image := config.Config.Image
		_, err := c.ctx.dockerClient.ImagePull(c.ctx.ctx, image, moby.ImagePullOptions{})
		if err != nil {
			return "", errors.Errorf("failed to pull %s image: %v", c.logIdentity, err)
		}
	}

	// create container
	resp, err := c.ctx.dockerClient.ContainerCreate(
		c.ctx.ctx,
		config.Config,
		config.HostConfig,
		config.NetworkingConfig,
		config.Platform,
		config.Name,
	)
	if err != nil {
		return "", errors.Errorf("failed to create %s container: %v", c.logIdentity, err)
	}
	return resp.ID, nil
}

func (c *ContainerRuntime) startContainer(id string) error {
	err := c.ctx.dockerClient.ContainerStart(c.ctx.ctx, id, moby.ContainerStartOptions{})
	if err != nil {
		errRemove := c.ctx.dockerClient.ContainerRemove(c.ctx.ctx, id, moby.ContainerRemoveOptions{Force: true})
		if errRemove != nil {
			return errors.Errorf("failed to remove %s container: %v "+
				"(after failed start due to: %v)", c.logIdentity, errRemove, err)
		}
		return errors.Errorf("failed to start %s container: %v", c.logIdentity, err)
	}
	c.ctx.t.Logf("started %s container %v", c.logIdentity, id)
	return nil
}

func (c *ContainerRuntime) stopContainer() error {
	err := c.ctx.dockerClient.ContainerStop(c.ctx.ctx, c.container.ID, container.StopOptions{Timeout: &c.stopTimeout})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return err
		}
		return errors.Errorf("failed to stop %s container: %v", c.logIdentity, err)
	}
	return nil
}

func (c *ContainerRuntime) removeContainer() error {
	err := c.ctx.dockerClient.ContainerRemove(c.ctx.ctx, c.container.ID, moby.ContainerRemoveOptions{Force: true})
	if err != nil {
		return errors.Errorf("failed to remove %s container: %v", c.logIdentity, err)
	}
	c.ctx.t.Logf("removed %s container %v", c.logIdentity, c.container.ID)
	return nil
}

// attachLoggingToContainer attaches nonblocking logging to current container. The logging doesn't use standard
// log output, but it uses provided logOutput argument as its output. This provides more flexibility for
// the caller of this method how the log output can be handled. The only exception is the final container exit
// status that is logged using standard output.
func (c *ContainerRuntime) attachLoggingToContainer(logOutput io.Writer) error {
	hijacked, err := c.ctx.dockerClient.ContainerAttach(c.ctx.ctx, c.container.ID, moby.ContainerAttachOptions{
		Logs:   true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return errors.Errorf("failed to attach logging to %s container: %v", c.logIdentity, err)
	}

	log := logging.DefaultLogger.WithField("name", c.logIdentity)
	log = log.WithField("container", c.container.Name)
	log = log.WithField("cid", stringid.TruncateID(c.container.ID))

	go func() {
		defer hijacked.Close()
		_, err := stdcopy.StdCopy(logOutput, logOutput, hijacked.Reader)
		if err != nil {
			log.Warnf("%s container exited: %v", c.logIdentity, err)
		} else {
			log.Debugf("%s container exited OK", c.logIdentity)
		}
	}()
	return nil
}

func (c *ContainerRuntime) inspectContainer(id string) (*moby.ContainerJSON, error) {
	info, err := c.ctx.dockerClient.ContainerInspect(c.ctx.ctx, id)
	if err != nil {
		return nil, errors.Errorf("failed to get info about %s container with ID %v due to: %v",
			c.logIdentity, id, err)
	}
	return &info, nil
}
