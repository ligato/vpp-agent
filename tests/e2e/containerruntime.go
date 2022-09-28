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

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/pkg/stringid"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-errors/errors"
	"github.com/segmentio/textio"
	"go.ligato.io/cn-infra/v2/logging"
)

const containerExecTimeout = 10 * time.Second

// ContainerRuntime represents docker container environments for one component of test topology
type ContainerRuntime struct {
	ctx         *TestCtx
	container   *docker.Container
	logIdentity string
	stopTimeout uint
}

// ContainerStartOptions are options for ComponentRuntime.Start(option) method implemented by ContainerRuntime
type ContainerStartOptions struct {
	ContainerOptions *docker.CreateContainerOptions
	Pull             bool
	AttachLogs       bool
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
	_, err := c.createContainer(opts.ContainerOptions, opts.Pull)
	if err != nil {
		return errors.Errorf("can't create %s container due to: %v", c.logIdentity, err)
	}
	log := logging.DefaultLogger.WithField("name", c.logIdentity)
	log.Debugf("starting container: %+v", *opts)
	if err := c.startContainer(); err != nil {
		return errors.Errorf("can't start %s container due to: %v", c.logIdentity, err)
	}
	log = log.WithField("container", c.container.Name)
	log = log.WithField("cid", stringid.TruncateID(c.container.ID))
	log.Debugf("container started")

	// attach logs (using one buffer from testctx -> all logs from all containers are merged together)
	if opts.AttachLogs {
		logWriter := textio.NewPrefixWriter(c.ctx.outputBuf, fmt.Sprintf("[container::%s/%v] ", c.container.Name, stringid.TruncateID(c.container.ID)))
		if err = c.attachLoggingToContainer(logWriter); err != nil {
			return errors.Errorf("can't attach logging to %s container due to: %v", c.logIdentity, err)
		}
	}
	return nil
}

// Stop stops and removes container
func (c *ContainerRuntime) Stop(options ...interface{}) error {
	if err := c.stopContainer(); err != nil {
		if errors.Is(err, &docker.NoSuchContainer{}) {
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
func (c *ContainerRuntime) ExecCmd(cmd string, args ...string) (stdout, stderr string, err error) {
	c.ctx.Logger.Printf("[container:%v] ExecCmd(%s, %v)", c.container.ID, cmd, args)

	opts := docker.CreateExecOptions{
		Context:      c.ctx.ctx,
		Container:    c.container.ID,
		Cmd:          append([]string{cmd}, args...),
		AttachStdout: true,
		AttachStderr: true,
	}
	exec, err := c.ctx.dockerClient.CreateExec(opts)
	if err != nil {
		err = errors.Errorf("failed to create docker exec for command %v due to: %v", cmd, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.ctx.ctx, containerExecTimeout)
	defer cancel()

	var stdoutBuf, stderrBuf bytes.Buffer
	err = c.ctx.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		OutputStream: &stdoutBuf,
		ErrorStream:  &stderrBuf,
	})
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	cmdStr := fmt.Sprintf("`%s %s`", cmd, strings.Join(args, " "))

	c.ctx.Logger.Printf("docker exec: %v:\nstdout(%d): %v\nstderr(%d): %v", cmdStr, len(stdout), stdout, len(stderr), stderr)

	if err != nil {
		errMsg := fmt.Sprintf("exec command %v failed due to: %v", cmdStr, err)
		c.ctx.Logger.Printf(errMsg)
		err = errors.Errorf(errMsg)
		return
	}

	if info, er := c.ctx.dockerClient.InspectExec(exec.ID); er != nil {
		c.ctx.t.Logf("exec inspect failed (ID %v, Cmd %s)s: %v", exec.ID, cmdStr, er)
		err = errors.Errorf("inspect exec error: %v", err)
	} else {
		c.ctx.Logger.Printf("exec details (ID %v, Cmd %s): %+v", exec.ID, cmdStr, info)
		if info.ExitCode != 0 {
			err = errors.Errorf("exec error (exit code %v): %v", info.ExitCode, stderr)
		}
	}

	return
}

// IPAddress provides ip address for connecting to the component
func (c *ContainerRuntime) IPAddress() string {
	return c.container.NetworkSettings.IPAddress
}

// PID provides process id of the main process in component
func (c *ContainerRuntime) PID() int {
	return c.container.State.Pid
}

func (c *ContainerRuntime) createContainer(containerOptions *docker.CreateContainerOptions,
	pull bool) (*docker.Container, error) {
	// pull image
	if pull {
		repo, tag, err := c.parseImageName(containerOptions.Config.Image)
		if err != nil {
			return nil, errors.Errorf("can't parse docker image %s "+
				"due to: %v", containerOptions.Config.Image, err)
		}

		err = c.ctx.dockerClient.PullImage(docker.PullImageOptions{
			Repository: repo,
			Tag:        tag,
		}, docker.AuthConfiguration{})
		if err != nil {
			return nil, errors.Errorf("failed to pull %s image: %v", c.logIdentity, err)
		}
	}

	// create container
	var err error
	c.container, err = c.ctx.dockerClient.CreateContainer(*containerOptions)
	if err != nil {
		return nil, errors.Errorf("failed to create %s container: %v", c.logIdentity, err)
	}
	return c.container, nil
}

func (c *ContainerRuntime) startContainer() error {
	if c.container == nil {
		return errors.Errorf("Reference to docker client container is nil. " +
			"Please use create() before start().")
	}

	// start container
	err := c.ctx.dockerClient.StartContainer(c.container.ID, nil)
	if err != nil {
		errRemove := c.ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    c.container.ID,
			Force: true,
		})
		if errRemove != nil {
			return errors.Errorf("failed to remove %s container: %v "+
				"(after failed start due to: %v)", c.logIdentity, errRemove, err)
		}
		return errors.Errorf("failed to start %s container: %v", c.logIdentity, err)
	}
	c.ctx.t.Logf("started %s container %v", c.logIdentity, c.container.ID)

	// update container reference (some attributes of container change by starting the container)
	id := c.container.ID
	c.container, err = c.inspectContainer(id)
	if err != nil {
		return errors.Errorf("can't update inner %s container reference for id %s "+
			"due to failing container inspect due to: %v", c.logIdentity, id, err)
	}
	return nil
}

func (c *ContainerRuntime) stopContainer() error {
	err := c.ctx.dockerClient.StopContainer(c.container.ID, c.stopTimeout)
	if errors.Is(err, &docker.NoSuchContainer{}) {
		return err
	} else if err != nil {
		return errors.Errorf("failed to stop %s container: %v", c.logIdentity, err)
	}
	return nil
}

func (c *ContainerRuntime) removeContainer() error {
	err := c.ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    c.container.ID,
		Force: true,
	})
	if err != nil {
		return errors.Errorf("failed to remove %s container: %v", c.logIdentity, err)
	}
	c.ctx.t.Logf("removed %s container %v", c.logIdentity, c.container.ID)
	return nil
}

// attachLoggingToContainer attaches nonblocking logging to current container. The logging doesn't use standard
// log output, but it uses provided logOutput argument as its output. This provides more flexibility for
// the caller of this method how the log output can be handled. The only exception is the final container exit
// status that is logged using stadard output.
func (c *ContainerRuntime) attachLoggingToContainer(logOutput io.Writer) error {
	closeWaiter, err := c.ctx.dockerClient.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    c.container.ID,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		Logs:         true,
		OutputStream: logOutput,
		ErrorStream:  logOutput,
	})
	if err != nil {
		return errors.Errorf("failed to attach logging to %s container: %v", c.logIdentity, err)
	}

	log := logging.DefaultLogger.WithField("name", c.logIdentity)
	log = log.WithField("container", c.container.Name)
	log = log.WithField("cid", stringid.TruncateID(c.container.ID))

	go func() {
		err := closeWaiter.Wait()
		if err != nil {
			log.Warnf("%s container exited: %v", c.logIdentity, err)
		} else {
			log.Debugf("%s container exited OK", c.logIdentity)
		}
	}()
	return nil
}

func (c *ContainerRuntime) inspectContainer(containerID string) (*docker.Container, error) {
	container, err := c.ctx.dockerClient.InspectContainerWithOptions(docker.InspectContainerOptions{
		Context: c.ctx.ctx,
		ID:      containerID,
	})
	if err != nil {
		return nil, errors.Errorf("failed to inspect %s container with ID %v due to: %v",
			c.logIdentity, containerID, err)
	}
	return container, nil
}

func (c *ContainerRuntime) parseImageName(imageName string) (repo, tag string, err error) {
	repo = imageName
	tag = "latest"
	if strings.Contains(imageName, ":") {
		split := strings.Split(imageName, ":")
		if len(split) != 2 {
			return repo, tag, errors.Errorf("image %s has is not valid "+
				"due too many repo-tag separator characters", imageName)
		}
		repo = split[0]
		tag = split[1]
	}
	return
}
