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
	"os"

	docker "github.com/fsouza/go-dockerclient"
)

// Setup options constants
const (
	HTTPsConnection             = "HTTPsConnection"
	VPPAgentContainerNetworking = "VPPAgentContainerNetworking"
)

// EtcdContainer is represents running ETCD container
type EtcdContainer struct {
	ctx         *TestCtx
	containerID string
}

// NewEtcdContainer creates and starts new ETCD container
func NewEtcdContainer(ctx *TestCtx, options ...*Option) *EtcdContainer {
	ec := &EtcdContainer{
		ctx: ctx,
	}
	container := ec.create(ctx, options...)
	ec.start(ctx, container)
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

func (ec *EtcdContainer) create(ctx *TestCtx, options ...*Option) *docker.Container {
	optionsMap := optionsMap(options)

	// pull image
	err := ctx.dockerClient.PullImage(docker.PullImageOptions{
		Repository: etcdImage,
		Tag:        "latest",
	}, docker.AuthConfiguration{})
	if err != nil {
		ctx.t.Fatalf("failed to pull ETCD image: %v", err)
	}

	// construct command string and container host config
	cmd := []string{
		"/usr/local/bin/etcd",
	}
	hostConfig := &docker.HostConfig{}
	if _, found := optionsMap[HTTPsConnection]; found {
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
	if _, found := optionsMap[VPPAgentContainerNetworking]; found {
		hostConfig.NetworkMode = "container:vpp-agent-e2e-test"
	} else { // separate container networking (default)
		hostConfig.PortBindings = map[docker.Port][]docker.PortBinding{
			"2379/tcp": {{HostIP: "0.0.0.0", HostPort: "2379"}},
		}
	}

	// create container
	container, err := ctx.dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: "e2e-test-etcd",
		Config: &docker.Config{
			Env:   []string{"ETCDCTL_API=3"},
			Image: etcdImage,
			Cmd:   cmd,
		},
		HostConfig: hostConfig,
	})
	if err != nil {
		ctx.t.Fatalf("failed to create ETCD container: %v", err)
	}
	return container
}

// WithEtcdHTTPsConnection is ETCD test setup option that will use HTTPS connection to ETCD (by default it is used
// unsecure HTTP connection)
func WithEtcdHTTPsConnection() *Option {
	return &Option{
		key:   HTTPsConnection,
		value: struct{}{}, // only presence is needed
	}
}

// WithEtcdVPPAgentContainerNetworking is ETCD test setup option that will use VPP-Agent test container for
// networking (by default the ETCD has separate networking)
func WithEtcdVPPAgentContainerNetworking() *Option {
	return &Option{
		key:   VPPAgentContainerNetworking,
		value: struct{}{}, // only presence is needed
	}
}

func (ec *EtcdContainer) start(ctx *TestCtx, container *docker.Container) {
	err := ctx.dockerClient.StartContainer(container.ID, nil)
	if err != nil {
		err = ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		if err != nil {
			ctx.t.Errorf("failed to remove ETCD container: %v", err)
		}
		ctx.t.Fatalf("failed to start ETCD container: %v", err)
	}
	ctx.t.Logf("started ETCD container %v", container.ID)
}

// Terminate stops and removes the ETCD container
func (ec *EtcdContainer) Terminate(ctx *TestCtx) {
	ec.stop(ctx)
	ec.remove(ctx)
}

func (ec *EtcdContainer) stop(ctx *TestCtx) {
	err := ctx.dockerClient.StopContainer(ec.containerID, msStopTimeout)
	if err != nil {
		ctx.t.Logf("failed to stop ETCD container: %v", err)
	}
}

func (ec *EtcdContainer) remove(ctx *TestCtx) {
	err := ctx.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    ec.containerID,
		Force: true,
	})
	if err != nil {
		ctx.t.Fatalf("failed to remove ETCD container: %v", err)
	}
	ctx.t.Logf("removed ETCD container %v", ec.containerID)
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
