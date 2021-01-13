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
	"io"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
)

const execTimeout = 10 * time.Second

// Container is represents running docker container
type Container struct {
	ctx         *TestCtx
	container   *docker.Container
	logIdentity string
	stopTimeout uint
}

func (c *Container) create(containerOptions *docker.CreateContainerOptions, pull bool) (*docker.Container, error) {
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

func (c *Container) start() error {
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
	c.container, err = c.inspect(id)
	if err != nil {
		return errors.Errorf("can't update inner %s container reference for id %s "+
			"due to failing container inspect due to: %v", c.logIdentity, id, err)
	}
	return nil
}

// terminate stops and removes the container
func (c *Container) terminate() error {
	if err := c.stop(); err != nil {
		if errors.Is(err, &docker.NoSuchContainer{}) {
			// container no longer exists -> nothing to do (state is the same as after successful termination)
			return nil
		}
		return err
	}
	if err := c.remove(); err != nil {
		return err
	}
	return nil
}

func (c *Container) stop() error {
	err := c.ctx.dockerClient.StopContainer(c.container.ID, c.stopTimeout)
	if errors.Is(err, &docker.NoSuchContainer{}) {
		return err
	} else if err != nil {
		return errors.Errorf("failed to stop %s container: %v", c.logIdentity, err)
	}
	return nil
}

func (c *Container) remove() error {
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

// execCmd executes command inside docker container
func (c *Container) execCmd(cmd string, args ...string) (string, error) {
	opts := docker.CreateExecOptions{
		Context:      c.ctx.ctx,
		Container:    c.container.ID,
		Cmd:          append([]string{cmd}, args...),
		AttachStdout: true,
		AttachStderr: true,
	}
	exec, err := c.ctx.dockerClient.CreateExec(opts)
	if err != nil {
		return "", errors.Errorf("failed to create docker exec for command %v due to: %v", cmd, err)
	}

	ctx, cancel := context.WithTimeout(c.ctx.ctx, execTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	err = c.ctx.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		OutputStream: &stdout,
		ErrorStream:  &stderr,
	})
	if err != nil {
		return "", errors.Errorf("starting of docker exec for command %v failed due to: %v", cmd, err)
	}

	if info, er := c.ctx.dockerClient.InspectExec(exec.ID); er != nil {
		c.ctx.t.Logf("exec inspect failed (ID %v, Cmd %s)s: %v", exec.ID, cmd, er)
	} else {
		c.ctx.logger.Printf("exec details (ID %v, Cmd %s): %+v", exec.ID, cmd, info)
		if info.ExitCode != 0 {
			err = errors.Errorf("exec error (exit code %v): %v", info.ExitCode, stderr.String())
		}
	}
	if strings.TrimSpace(stderr.String()) != "" {
		return "", errors.Errorf("failed exec command %s "+
			"due to nonempty error output: %s", cmd, stderr.String())
	}
	return stdout.String(), err
}

// attachLoggingToContainer attaches nonblocking logging to current container. The logging doesn't use standard
// log output, but it uses provided logOutput argument as its output. This provides more flexibility for
// the caller of this method how the log output can be handled. The only exception is the final container exit
// status that is logged using stadard output.
func (c *Container) attachLoggingToContainer(logOutput io.Writer) error {
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

	log := logrus.WithField("name", c.logIdentity)
	log = log.WithField("container", c.container.Name)
	log = log.WithField("cid", c.container.ID)

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

// Inspect provides actual docker.Container of running container that can be
// used to inspect various things about the container
func (c *Container) Inspect() (*docker.Container, error) {
	return c.inspect(c.container.ID)
}

func (c *Container) inspect(containerID string) (*docker.Container, error) {
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

func (c *Container) parseImageName(imageName string) (repo, tag string, err error) {
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
