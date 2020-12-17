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
	"fmt"
	"os"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-errors/errors"
)

const (
	etcdImage       = "gcr.io/etcd-development/etcd"
	etcdStopTimeout = 1 // seconds
)

// TODO unify whether return errors or use test context's log fatal

// EtcdContainer is represents running ETCD container
type EtcdContainer struct {
	ctx         *TestCtx
	containerID string
}

// NewEtcdContainer creates and starts new ETCD container
func NewEtcdContainer(ctx *TestCtx, options ...EtcdOptModifier) *EtcdContainer {
	ec := &EtcdContainer{
		ctx: ctx,
	}
	container := ec.create(options...)
	ec.start(container)
	ec.containerID = container.ID
	return ec
}

// Put inserts key-value pair into the ETCD inside its running docker container
func (ec *EtcdContainer) Put(key string, value string) error {
	_, err := ec.exec("etcdctl", "put", key, value)
	return err
}

// Get retrieves value for the key from the ETCD that is running in its docker container
func (ec *EtcdContainer) Get(key string) (string, error) {
	return ec.exec("etcdctl", "get", key)
}

// Get retrieves all key-value pairs from the ETCD that is running in its docker container
func (ec *EtcdContainer) GetAll() (string, error) {
	return ec.exec("etcdctl", "get", "", "--prefix=true")
}

// Inspect provides docker.Container of running ETCD container that can be
// used to inspect various things about ETCD container
func (ec *EtcdContainer) Inspect() *docker.Container {
	container, err := ec.ctx.dockerClient.InspectContainer(ec.containerID)
	if err != nil {
		ec.ctx.t.Fatalf("failed to inspect container with ID %v due to: %v", ec.containerID, err)
	}
	return container
}

func (ec *EtcdContainer) create(options ...EtcdOptModifier) *docker.Container {
	opts := DefaultEtcdOpt()
	for _, optionModifier := range options {
		optionModifier(opts)
	}

	// pull image
	err := ec.ctx.dockerClient.PullImage(docker.PullImageOptions{
		Repository: etcdImage,
		Tag:        "latest",
	}, docker.AuthConfiguration{})
	if err != nil {
		ec.ctx.t.Fatalf("failed to pull ETCD image: %v", err)
	}

	// construct command string and container host config
	cmd := []string{
		"/usr/local/bin/etcd",
	}
	hostConfig := &docker.HostConfig{}
	if opts.UseHTTPS {
		cmd = append(cmd,
			"--client-cert-auth",
			"--trusted-ca-file=/etc/certs/ca.pem",
			"--cert-file=/etc/certs/cert1.pem",
			"--key-file=/etc/certs/cert1-key.pem",
			"--advertise-client-urls=https://127.0.0.1:2379",
			"--listen-client-urls=https://127.0.0.1:2379",
		)
		hostConfig.Binds = []string{os.Getenv("CERTS_PATH") + ":/etc/certs"}
	} else { // HTTP connection
		cmd = append(cmd,
			"--advertise-client-urls=http://0.0.0.0:2379",
			"--listen-client-urls=http://0.0.0.0:2379",
		)
	}
	if opts.UseAgentContainerForNetworking {
		hostConfig.NetworkMode = fmt.Sprintf("container:e2e-test-vppagent-%v", AgentInstanceName(ec.ctx))
	} else { // separate container networking (default)
		hostConfig.PortBindings = map[docker.Port][]docker.PortBinding{
			"2379/tcp": {{HostIP: "0.0.0.0", HostPort: "2379"}},
		}
	}

	// create container
	container, err := ec.ctx.dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: "e2e-test-etcd",
		Config: &docker.Config{
			Env:   []string{"ETCDCTL_API=3"},
			Image: etcdImage,
			Cmd:   cmd,
		},
		HostConfig: hostConfig,
	})
	if err != nil {
		ec.ctx.t.Fatalf("failed to create ETCD container: %v", err)
	}
	return container
}

func (ec *EtcdContainer) start(container *docker.Container) {
	err := ec.ctx.dockerClient.StartContainer(container.ID, nil)
	if err != nil {
		err = ec.ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		if err != nil {
			ec.ctx.t.Errorf("failed to remove ETCD container: %v", err)
		}
		ec.ctx.t.Fatalf("failed to start ETCD container: %v", err)
	}
	ec.ctx.t.Logf("started ETCD container %v", container.ID)
}

// Terminate stops and removes the ETCD container
func (ec *EtcdContainer) Terminate() error {
	if err := ec.stop(); err != nil {
		return err
	}
	if err := ec.remove(); err != nil {
		return err
	}
	return nil
}

func (ec *EtcdContainer) stop() error {
	err := ec.ctx.dockerClient.StopContainer(ec.containerID, etcdStopTimeout)
	if err != nil {
		return errors.Errorf("failed to stop ETCD container: %v", err)
	}
	return nil
}

func (ec *EtcdContainer) remove() error {
	err := ec.ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    ec.containerID,
		Force: true,
	})
	if err != nil {
		return errors.Errorf("failed to remove ETCD container: %v", err)
	}
	ec.ctx.t.Logf("removed ETCD container %v", ec.containerID)
	return nil
}

// exec executes command inside Etcd container
func (ec *EtcdContainer) exec(cmdName string, args ...string) (output string, err error) {
	execCtx, err := ec.ctx.dockerClient.CreateExec(docker.CreateExecOptions{
		AttachStdout: true,
		Cmd:          append([]string{cmdName}, args...),
		Container:    ec.containerID,
	})
	if err != nil {
		ec.ctx.t.Fatalf("failed to create docker exec instance for exec in etcd container: %v", err)
	}

	var stdout bytes.Buffer
	err = ec.ctx.dockerClient.StartExec(execCtx.ID, docker.StartExecOptions{
		OutputStream: &stdout,
	})
	return stdout.String(), err
}
