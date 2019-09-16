//  Copyright (c) 2019 Cisco and/or its affiliates.
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
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestAgentCtl(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()
	Expect(true).To(BeTrue())

	var cmd *exec.Cmd
	var err error
	var stdout, stderr bytes.Buffer
	var output string
	var matched bool

	// Test if executable is present
	cmd = exec.Command("/agentctl")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	Expect(stdout.Len()).To(Not(BeZero()))

	// command: `agentctl dump vpp.interfaces`
	// expecting at least one interface with `type: SOFTWARE_LOOPBACK`
	stdout.Reset()
	cmd = exec.Command("/agentctl", "dump", "vpp.interfaces")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	Expect(strings.Contains(output, "type: SOFTWARE_LOOPBACK")).To(BeTrue())

	// command: `agentctl generate vpp.interfaces`
	// expecting not empty output
	stdout.Reset()
	cmd = exec.Command("/agentctl", "generate", "vpp.interfaces")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	Expect(stdout.Len()).To(Not(BeZero()))

	// command: `agentctl help`
	// expecting not empty output
	stdout.Reset()
	cmd = exec.Command("/agentctl", "help")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	Expect(stdout.Len()).To(Not(BeZero()))

	// command: `agentctl import /tmp/config1`
	// expecting it to fail due to missing ETCD
	f, err := os.Create("/tmp/config1")
	Expect(err).To(BeNil())
	w := bufio.NewWriter(f)
	_, err = w.WriteString(`config/vpp/v2/interfaces/tap2 {"name":"tap2", "type":"TAP", "enabled":true, "ip_addresses":["10.10.10.10/24"], "tap":{"version": "2"}}`)
	Expect(err).To(BeNil())
	w.Flush()
	stdout.Reset()
	stderr.Reset()
	cmd = exec.Command("/agentctl", "import", "/tmp/config1", "--service-label", "vpp1")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	Expect(err).To(Not(BeNil()))
	Expect(stdout.Len()).To(BeZero())
	output = stderr.String()
	Expect(strings.Contains(output, "Error: connecting to Etcd failed:")).To(BeTrue())

	// command: `agentctl import /tmp/config1 --grpc`
	// expecting to successfully send txn
	stdout.Reset()
	cmd = exec.Command("/agentctl", "import", "/tmp/config1", "--service-label", "vpp1", "--grpc")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	Expect(stdout.String()).To(Equal("importing 1 key vals\n - /vnf-agent/vpp1/config/vpp/v2/interfaces/tap2\nsending via gRPC\n"))

	// command: `agentctl kvdb list`
	// expecting it to fail due to missing ETCD
	stdout.Reset()
	stderr.Reset()
	cmd = exec.Command("/agentctl", "kvdb", "list")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	Expect(err).To(Not(BeNil()))
	Expect(stdout.Len()).To(BeZero())
	output = stderr.String()
	Expect(strings.Contains(output, "ERROR: connecting to Etcd failed:")).To(BeTrue())

	// command: `agentctl log list`
	// expecting level of logger `agent` is `info`
	stdout.Reset()
	cmd = exec.Command("/agentctl", "log", "list")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	matched, err = regexp.MatchString(`agent\s+info`, output)
	Expect(err).To(BeNil())
	Expect(matched).To(BeTrue())

	// command: `agentctl log set agent debug`
	// expecting to change log level of `agent`
	stdout.Reset()
	cmd = exec.Command("/agentctl", "log", "set", "agent", "debug")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	Expect(stdout.String()).To(Equal("logger agent has been set to level debug\n"))

	// command: `agentctl log list`
	// expecting level of logger `agent` is `debug`
	stdout.Reset()
	cmd = exec.Command("/agentctl", "log", "list")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	matched, err = regexp.MatchString(`agent\s+debug`, output)
	Expect(err).To(BeNil())
	Expect(matched).To(BeTrue())

	// command: `agentctl model ls`
	// expecting to find at least info about linux interfaces
	stdout.Reset()
	cmd = exec.Command("/agentctl", "model", "ls")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	matched, err = regexp.MatchString(`linux.interfaces.interface\s+config/linux/interfaces/v2/interface/\s+linux.interfaces.Interface`, output)
	Expect(err).To(BeNil())
	Expect(matched).To(BeTrue())

	// command: `agentctl model inspect vpp.interfaces`
	// expecting to find at least `KeyPrefix` of `vpp.interfaces` model
	stdout.Reset()
	cmd = exec.Command("/agentctl", "model", "inspect", "vpp.interfaces")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	Expect(strings.Contains(output, `"KeyPrefix": "config/vpp/v2/interfaces/",`)).To(BeTrue())

	// command: `agentctl status`
	// expecting to find `UNTAGGED-local0` obtained vpp interface
	stdout.Reset()
	cmd = exec.Command("/agentctl", "status")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	matched, err = regexp.MatchString(`vpp.interfaces\s+UNTAGGED-local0\s+obtained`, output)
	Expect(err).To(BeNil())
	Expect(matched).To(BeTrue())

	// command: `agentctl vpp info`
	// expecting to find at least Version of VPP
	stdout.Reset()
	cmd = exec.Command("/agentctl", "vpp", "info")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	matched, err = regexp.MatchString(`Version:\s+v\d{2}\.\d{2}`, output)
	Expect(err).To(BeNil())
	Expect(matched).To(BeTrue())

	// command: `agentctl vpp cli sh int`
	// expecting to find `local0` interface in output of executed vpp cli command
	stdout.Reset()
	cmd = exec.Command("/agentctl", "vpp", "cli", "sh", "int")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	matched, err = regexp.MatchString(`local0\s+0\s+down\s+0/0/0/0`, output)
	Expect(err).To(BeNil())
	Expect(matched).To(BeTrue())
}
