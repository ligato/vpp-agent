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

	var cmd *exec.Cmd
	var err error
	var stdout, stderr bytes.Buffer
	var matched bool

	// file created below is required to test `import` action
	err = createFileWithContent(
		"/tmp/config1",
		`config/vpp/v2/interfaces/tap1 {"name":"tap1", "type":"TAP", "enabled":true, "ip_addresses":["10.10.10.10/24"], "tap":{"version": "2"}}`,
	)
	Expect(err).To(BeNil(), "Failed to create file required by one of the tests")

	tests := []struct {
		name                 string
		cmd                  string
		expectErr            bool
		expectNotEmptyStdout bool
		expectStdout         string
		expectInStdout       string
		expectReStdout       string
		expectInStderr       string
	}{
		{
			name:                 "Check if executable is present",
			cmd:                  "--help",
			expectNotEmptyStdout: true,
		},
		{
			name:           "Test `dump all` action",
			cmd:            "dump all",
			expectInStdout: "type: SOFTWARE_LOOPBACK",
		},
		{
			name:           "Test `dump vpp.*` action",
			cmd:            `dump vpp.*`,
			expectInStdout: "type: SOFTWARE_LOOPBACK",
		},
		{
			name:           "Test `dump` action with bad model",
			cmd:            "dump NoSuchModel",
			expectErr:      true,
			expectInStderr: "no models found for [\"NoSuchModel\"]",
		},
		{
			name:           "Test `dump` action with one bad model",
			cmd:            "dump NoSuchModel vpp.interfaces",
			expectInStdout: "type: SOFTWARE_LOOPBACK",
		},
		{
			name:           "Test `dump --view=SB` action",
			cmd:            "dump vpp.interfaces --view=SB",
			expectInStdout: "type: SOFTWARE_LOOPBACK",
		},
		{
			name:           "Test `dump --view=NB` action",
			cmd:            "dump vpp.interfaces --view=NB",
			expectReStdout: `KEY\s+VALUE\s+ORIGIN\s+METADATA`,
		},
		{
			name:           "Test `dump --view=cached` action",
			cmd:            "dump vpp.interfaces --view=cached",
			expectInStdout: "type: SOFTWARE_LOOPBACK",
		},
		{
			name:           "Test `dump` with JSON format",
			cmd:            "dump vpp.interfaces -f=json",
			expectReStdout: `"Value": {\s+"name": "UNTAGGED-local0",`,
		},
		{
			name:           "Test `dump` with YAML format",
			cmd:            "dump vpp.interfaces -f=yaml",
			expectReStdout: `Value:\s+name: UNTAGGED-local0`,
		},

		{
			name:         "Test `dump` with custom format",
			cmd:          `dump vpp.interfaces -f "{{range.}}Name:{{.Value.Name}}{{end}}"`,
			expectStdout: `"Name:UNTAGGED-local0"`,
		},
		{
			name:                 "Test `generate` action",
			cmd:                  "generate vpp.interfaces",
			expectNotEmptyStdout: true,
		},
		{
			name:           "Test `generate` action with not exsiting model",
			cmd:            "generate NoSuchModel",
			expectErr:      true,
			expectInStderr: "no model found for: NoSuchModel",
		},
		{
			name:           "Test `generate` action to yaml",
			cmd:            "generate vpp.interfaces -f=yaml",
			expectInStdout: "type: UNDEFINED_TYPE",
		},
		{
			name:           "Test `generate` action to json",
			cmd:            "generate vpp.interfaces -f=json",
			expectInStdout: `"type": "UNDEFINED_TYPE",`,
		},
		{
			name:           "Test `generate` action to json (oneline)",
			cmd:            "generate vpp.interfaces -f=json --oneline",
			expectInStdout: `{"name":"","type":"UNDEFINED_TYPE",`,
		},
		/*{
			// This test depends on file (/tmp/config1) which was created before.
			name:           "Test `import` action",
			cmd:            "import /tmp/config1 --service-label vpp1",
			expectErr:      true,
			expectInStderr: "connecting to Etcd failed",
		},
		{
			// This test depends on file (/tmp/config1) which was created before.
			name:         "Test `import` action (grpc)",
			cmd:          "import /tmp/config1 --service-label vpp1 --grpc",
			expectStdout: "importing 1 key vals\n - /vnf-agent/vpp1/config/vpp/v2/interfaces/tap1\nsending via gRPC\n",
		},*/
		{
			name:           "Test `kvdb list` action",
			cmd:            "kvdb list",
			expectErr:      true,
			expectInStderr: "connecting to Etcd failed",
		},
		{
			name:           "Test `log list` action",
			cmd:            "log list",
			expectReStdout: `agent\s+info`,
		},
		{
			name:         "Test `log set` action",
			cmd:          "log set agent debug",
			expectStdout: "logger agent has been set to level debug\n",
		},
		{
			// This test depends on previous one.
			name:           "Test `log list` action",
			cmd:            "log list",
			expectReStdout: `agent\s+debug`,
		},
		{
			name:           "Test `model ls` action",
			cmd:            "model ls",
			expectReStdout: `linux.interfaces.interface\s+config\s+ligato.linux.interfaces.Interface`,
		},
		{
			name:           "Test `model inspect` action",
			cmd:            "model inspect vpp.interfaces",
			expectInStdout: `"KeyPrefix": "config/vpp/v2/interfaces/",`,
		},
		{
			name:           "Test `model inspect` action (no models)",
			cmd:            "model inspect NoSuchModel",
			expectErr:      true,
			expectInStderr: "no model found for provided prefix: NoSuchModel",
		},
		{
			name:           "Test `model inspect` action (multiple models)",
			cmd:            "model inspect vpp.",
			expectErr:      true,
			expectInStderr: "multiple models found with provided prefix: vpp.",
		},
		{
			name:           "Test `status` action",
			cmd:            "status",
			expectInStdout: `State: OK`,
		},
		{
			name:         "Test `status` action (with format)",
			cmd:          "status -f {{.AgentStatus.State}}",
			expectStdout: "OK",
		},
		{
			name:           "Test `values` action",
			cmd:            "values",
			expectReStdout: `vpp.interfaces\s+UNTAGGED-local0\s+obtained`,
		},
		/*{
			name:           "Test `values` action (with model)",
			cmd:            "values vpp.proxyarp-global",
			expectReStdout: `vpp.proxyarp-global\s+obtained `,
		},*/
		{
			name:           "Test `vpp info` action",
			cmd:            "vpp info",
			expectReStdout: `Version:\s+v\d{2}\.\d{2}`,
		},
		{
			name:           "Test `vpp cli` action",
			cmd:            "vpp cli sh int",
			expectReStdout: `local0\s+0\s+down\s+0/0/0/0`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset both buffers to be empty before test
			stdout.Reset()
			stderr.Reset()

			// Run command
			cmd = exec.Command("/agentctl", strings.Split(test.cmd, " ")...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()

			if test.expectErr {
				Expect(err).To(Not(BeNil()),
					"Command `%s` should fail\n",
					test.cmd,
				)
			} else {
				Expect(err).To(BeNil(),
					"Command `%s` should not fail. Got err: %v\nStderr:\n%s\n",
					test.cmd, err, stderr.String(),
				)
			}

			// Check STDOUT:
			if test.expectNotEmptyStdout {
				Expect(stdout.Len()).To(Not(BeZero()),
					"Stdout should not be empty\n",
				)
			}

			if test.expectStdout != "" {
				Expect(stdout.String()).To(Equal(test.expectStdout),
					"Want stdout: \n%s\nGot stdout: \n%s\n",
					test.expectStdout, stdout.String(),
				)
			}

			if test.expectInStdout != "" {
				Expect(strings.Contains(stdout.String(), test.expectInStdout)).To(BeTrue(),
					"Want in stdout: \n%s\nGot stdout: \n%s\n",
					test.expectInStdout, stdout.String(),
				)
			}

			if test.expectReStdout != "" {
				matched, err = regexp.MatchString(test.expectReStdout, stdout.String())
				Expect(err).To(BeNil())
				Expect(matched).To(BeTrue(),
					"Want stdout to contain any match of the regular expression: \n`%s`\nGot stdout: \n%s\n",
					test.expectReStdout, stdout.String(),
				)
			}

			// Check STDERR:
			if test.expectInStderr != "" {
				Expect(strings.Contains(stderr.String(), test.expectInStderr)).To(BeTrue(),
					"Want in stderr: \n%s\nGot stderr: \n%s\n",
					test.expectInStderr, stderr.String(),
				)
			}
		})
	}
}

func createFileWithContent(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	_, err = w.WriteString(content)
	if err != nil {
		return err
	}
	w.Flush()

	return nil
}
