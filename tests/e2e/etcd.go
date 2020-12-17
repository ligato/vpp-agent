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
)

func (test *TestCtx) pullEtcd() {
	err := test.dockerClient.PullImage(docker.PullImageOptions{
		Repository: etcdImage,
		Tag:        "latest",
	}, docker.AuthConfiguration{})
	if err != nil {
		test.t.Fatalf("failed to pull ETCD image: %v", err)
	}
}

func (test *TestCtx) StartEtcd() string {
	container, err := test.dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: "e2e-test-etcd",
		Config: &docker.Config{
			Image: etcdImage,
			Env:   []string{"ETCDCTL_API=3"},
			Cmd: []string{
				"/usr/local/bin/etcd",
				"--client-cert-auth",
				"--trusted-ca-file=/etc/certs/ca.pem",
				"--cert-file=/etc/certs/cert1.pem",
				"--key-file=/etc/certs/cert1-key.pem",
				"--advertise-client-urls=https://127.0.0.1:2379",
				"--listen-client-urls=https://127.0.0.1:2379",
			},
		},
		HostConfig: &docker.HostConfig{
			NetworkMode: "container:vpp-agent-e2e-test",
			Binds: []string{
				filepath.Join(test.testDataDir, "certs") + ":/etc/certs:ro",
			},
		},
	})
	if err != nil {
		test.t.Fatalf("failed to create ETCD container: %v", err)
	}
	err = test.dockerClient.StartContainer(container.ID, nil)
	if err != nil {
		err = test.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		if err != nil {
			test.t.Errorf("failed to remove ETCD container: %v", err)
		}
		test.t.Fatalf("failed to start ETCD container: %v", err)
	}
	test.t.Logf("started ETCD container %v", container.ID)
	return container.ID
}

func (test *TestCtx) StopEtcd(id string) {
	err := test.dockerClient.StopContainer(id, msStopTimeout)
	if err != nil {
		test.t.Logf("failed to stop ETCD container: %v", err)
	}
	err = test.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    id,
		Force: true,
	})
	if err != nil {
		test.t.Fatalf("failed to remove ETCD container: %v", err)
	}
	test.t.Logf("removed ETCD container %v", id)
}
