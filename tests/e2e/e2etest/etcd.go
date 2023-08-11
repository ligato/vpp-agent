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

package e2etest

import (
	"path/filepath"

	moby "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/go-errors/errors"
)

const (
	etcdImage       = "gcr.io/etcd-development/etcd:v3.5.5"
	etcdStopTimeout = 1 // seconds
)

// Etcd is represents running ETCD
type Etcd struct {
	ComponentRuntime
	ctx *TestCtx
}

// NewEtcd creates and starts new ETCD container
func NewEtcd(ctx *TestCtx, optMods ...EtcdOptModifier) (*Etcd, error) {
	// compute options
	opts := DefaultEtcdOpt(ctx)
	for _, mod := range optMods {
		mod(opts)
	}

	// create struct for ETCD server
	etcd := &Etcd{
		ComponentRuntime: opts.Runtime,
		ctx:              ctx,
	}

	// get runtime specific options and start ETCD in runtime environment
	startOpts, err := opts.RuntimeStartOptions(ctx, opts)
	if err != nil {
		return nil, errors.Errorf("can't get ETCD start option for runtime due to: %v", err)
	}
	err = etcd.Start(startOpts)
	if err != nil {
		return nil, errors.Errorf("can't start ETCD due to: %v", err)
	}
	return etcd, nil
}

// Put inserts key-value pair into the ETCD inside its running docker container
func (ec *Etcd) Put(key string, value string) error {
	_, _, err := ec.ExecCmd("etcdctl", "put", key, value)
	return err
}

// Get retrieves value for the key from the ETCD that is running in its docker container
func (ec *Etcd) Get(key string) (string, error) {
	stdout, _, err := ec.ExecCmd("etcdctl", "get", key)
	return stdout, err
}

// GetAll retrieves all key-value pairs from the ETCD that is running in its docker container
func (ec *Etcd) GetAll() (string, error) {
	stdout, _, err := ec.ExecCmd("etcdctl", "get", "", "--prefix=true")
	return stdout, err
}

// ETCDStartOptionsForContainerRuntime translates EtcdOpt to options for ComponentRuntime.Start(option)
// method implemented by ContainerRuntime
func ETCDStartOptionsForContainerRuntime(ctx *TestCtx, options interface{}) (interface{}, error) {
	opts, ok := options.(*EtcdOpt)
	if !ok {
		return nil, errors.Errorf("expected EtcdOpt but got %+v", options)
	}

	// construct command string and container host config
	cmd := []string{
		"/usr/local/bin/etcd",
	}
	hostConfig := &container.HostConfig{}
	if opts.UseHTTPS {
		cmd = append(cmd,
			"--client-cert-auth",
			"--trusted-ca-file=/etc/certs/ca.pem",
			"--cert-file=/etc/certs/cert1.pem",
			"--key-file=/etc/certs/cert1-key.pem",
			"--advertise-client-urls=https://127.0.0.1:2379",
			"--listen-client-urls=https://127.0.0.1:2379",
		)
		hostConfig.Binds = []string{filepath.Join(ctx.DataDir, "certs") + ":/etc/certs:ro"}
	} else { // HTTP connection
		cmd = append(cmd,
			"--advertise-client-urls=http://0.0.0.0:2379",
			"--listen-client-urls=http://0.0.0.0:2379",
		)
	}
	if opts.UseTestContainerForNetworking {
		hostConfig.NetworkMode = "container:vpp-agent-e2e-test"
	} else { // separate container networking (default)
		hostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
			"2379/tcp": {{HostIP: "0.0.0.0", HostPort: "2379"}},
		}
	}
	containerConfig := &moby.ContainerCreateConfig{
		Name: "e2e-test-etcd",
		Config: &container.Config{
			Env:   []string{"ETCDCTL_API=3"},
			Image: etcdImage,
			Cmd:   cmd,
		},
		HostConfig: hostConfig,
	}

	return &ContainerStartOptions{
		ContainerConfig: containerConfig,
		Pull:            true,
	}, nil
}
