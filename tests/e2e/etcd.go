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
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-errors/errors"
)

const (
	etcdImage       = "gcr.io/etcd-development/etcd"
	etcdStopTimeout = 1 // seconds
)

// EtcdContainer is represents running ETCD container
type EtcdContainer struct {
	*Container
}

// NewEtcdContainer creates and starts new ETCD container
func NewEtcdContainer(ctx *TestCtx, options ...EtcdOptModifier) (*EtcdContainer, error) {
	ec := &EtcdContainer{
		&Container{
			ctx:         ctx,
			logIdentity: "ETCD",
			stopTimeout: etcdStopTimeout,
		},
	}
	_, err := ec.create(options...)
	if err != nil {
		return nil, errors.Errorf("can't create %s container due to: %v", ec.logIdentity, err)
	}
	if err := ec.start(); err != nil {
		return nil, errors.Errorf("can't start %s container due to: %v", ec.logIdentity, err)
	}
	return ec, nil
}

// Put inserts key-value pair into the ETCD inside its running docker container
func (ec *EtcdContainer) Put(key string, value string) error {
	_, err := ec.execCmd("etcdctl", "put", key, value)
	return err
}

// Get retrieves value for the key from the ETCD that is running in its docker container
func (ec *EtcdContainer) Get(key string) (string, error) {
	return ec.execCmd("etcdctl", "get", key)
}

// GetAll retrieves all key-value pairs from the ETCD that is running in its docker container
func (ec *EtcdContainer) GetAll() (string, error) {
	return ec.execCmd("etcdctl", "get", "", "--prefix=true")
}

func (ec *EtcdContainer) create(options ...EtcdOptModifier) (*docker.Container, error) {
	opts := DefaultEtcdOpt()
	for _, optionModifier := range options {
		optionModifier(opts)
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
		hostConfig.Binds = []string{filepath.Join(ec.ctx.testDataDir, "certs") + ":/etc/certs:ro"}
	} else { // HTTP connection
		cmd = append(cmd,
			"--advertise-client-urls=http://0.0.0.0:2379",
			"--listen-client-urls=http://0.0.0.0:2379",
		)
	}
	if opts.UseTestContainerForNetworking {
		hostConfig.NetworkMode = "container:vpp-agent-e2e-test"
	} else { // separate container networking (default)
		hostConfig.PortBindings = map[docker.Port][]docker.PortBinding{
			"2379/tcp": {{HostIP: "0.0.0.0", HostPort: "2379"}},
		}
	}
	containerOptions := &docker.CreateContainerOptions{
		Name: "e2e-test-etcd",
		Config: &docker.Config{
			Env:   []string{"ETCDCTL_API=3"},
			Image: etcdImage,
			Cmd:   cmd,
		},
		HostConfig: hostConfig,
	}

	return ec.Container.create(containerOptions, true)
}
