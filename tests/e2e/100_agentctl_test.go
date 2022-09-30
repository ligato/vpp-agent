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
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestAgentCtlCommands(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	var err error
	var stdout, stderr string

	// File created below is required to test `import` action.
	config1File := ctx.testShareDir + "/agentctl-config1.yaml"
	_, err = createFileWithContent(
		config1File,
		`config/vpp/v2/interfaces/tap1 {"name":"tap1", "type":"TAP", "enabled":true, "ip_addresses":["10.10.10.10/24"], "tap":{"version": "2"}}`,
	)
	ctx.Expect(err).To(BeNil(), "Failed to create file required by one of the tests")
	// cleanup the file
	defer func() {
		err = os.Remove(config1File)
		ctx.Expect(err).To(BeNil())
	}()

	// These update files created below are required to test `get` and `update` action with labels.
	// All tests using `agentctl get` depend on the existence of these files.
	nextDummyIf := dummyIfFactory(ctx)
	updateLabels := []string{"if=dummy", "\"if=dummy\",\"source=test\"", "\"if=differentvalue\",\"source=test\"", "", "\"onlykey=\""}
	for _, ul := range updateLabels {
		file, err := createFileWithContent(nextDummyIf())
		ctx.Expect(err).To(BeNil(), "Failed to create file required by one of the tests")
		stdout, _, err = ctx.Agent.ExecCmd("agentctl", "config", "update", file, "--labels="+ul)
		ctx.Expect(err).ToNot(HaveOccurred())
		// ctx.Expect(stderr).To(BeEmpty()) TODO: uncomment this once the warning log has been cleaned up
		ctx.Expect(stdout).To(ContainSubstring("OK"))

		// cleanup the file
		defer func() {
			err = os.Remove(file)
			ctx.Expect(err).ToNot(HaveOccurred())
		}()
	}

	// Parsing these labels should result in an error.
	wrongUpdateLabels := []string{"\"=onlyvalue\"", "\"duplicatekey=foo\",\"duplicatekey=bar\"", "\"\""}
	for _, wul := range wrongUpdateLabels {
		file, err := createFileWithContent(nextDummyIf())
		ctx.Expect(err).To(BeNil(), "Failed to create file required by one of the tests")
		stdout, _, err = ctx.Agent.ExecCmd("agentctl", "config", "update", file, "--labels="+wul)
		ctx.Expect(err).To(HaveOccurred())
		// ctx.Expect(stderr).To(BeEmpty()) TODO: uncomment this once the warning log has been cleaned up
		ctx.Expect(stdout).ToNot(ContainSubstring("OK"))

		// cleanup the file
		defer func() {
			err = os.Remove(file)
			ctx.Expect(err).ToNot(HaveOccurred())
		}()
	}

	type KeyVal struct {
		Key   string
		Value interface{}
	}
	tests := []struct {
		name                 string
		cmd                  string
		expectErr            bool
		expectNotEmptyStdout bool
		expectStdout         string
		expectInStdout       string
		expectNotInStdout    string
		expectReStdout       string
		expectNotReStdout    string
		expectInStderr       string
		expectJsonKeyVals    []KeyVal
	}{
		{
			name:                 "Check if executable is present",
			cmd:                  "--help",
			expectNotEmptyStdout: true,
		},
		{
			name:           "Test `config get`",
			cmd:            "config get",
			expectInStdout: "type: DUMMY",
			expectReStdout: "name: dummyif(0|1|2|3|4)",
		},
		{
			name:              "Test `config get` with full label",
			cmd:               "config get --labels=\"if=dummy\"",
			expectInStdout:    "type: DUMMY",
			expectReStdout:    "name: dummyif(0|1)",
			expectNotReStdout: "name: dummyif(2|3|4)",
		},
		{
			name:              "Test `config get` with label key",
			cmd:               "config get --labels=\"if\"",
			expectInStdout:    "type: DUMMY",
			expectReStdout:    "name: dummyif(0|1|2)",
			expectNotReStdout: "name: dummyif(4|5)",
		},
		{
			name:              "Test `config get` with label key",
			cmd:               "config get --labels=\"if=\"",
			expectInStdout:    "type: DUMMY",
			expectReStdout:    "name: dummyif(0|1|2)",
			expectNotReStdout: "name: dummyif(3|4)",
		},
		{
			name:              "Test `config get` with multiple full labels",
			cmd:               "config get --labels=\"if=dummy\" --labels=\"source=test\"",
			expectInStdout:    "type: DUMMY",
			expectReStdout:    "name: dummyif(1)",
			expectNotReStdout: "name: dummyif(0|2|3|4)",
		},
		{
			name:              "Test `config get` with multiple label keys",
			cmd:               "config get --labels=\"if\",\"source\"",
			expectInStdout:    "type: DUMMY",
			expectReStdout:    "name: dummyif(1|2)",
			expectNotReStdout: "name: dummyif(0|3|4)",
		},
		{
			name:              "Test `config get` with multiple label keys",
			cmd:               "config get --labels=\"if=\",\"source=\"",
			expectInStdout:    "type: DUMMY",
			expectReStdout:    "name: dummyif(1|2)",
			expectNotReStdout: "name: dummyif(0|3|4)",
		},
		{
			name:              "Test `config get` with multiple label keys",
			cmd:               "config get --labels=\"if\" --labels=\"source=\"",
			expectInStdout:    "type: DUMMY",
			expectReStdout:    "name: dummyif(1|2)",
			expectNotReStdout: "name: dummyif(0|3|4)",
		},
		{
			name:              "Test `config get` with label key and full label",
			cmd:               "config get --labels=\"if=dummy\",\"source\"",
			expectInStdout:    "type: DUMMY",
			expectReStdout:    "name: dummyif(1)",
			expectNotReStdout: "name: dummyif(0|2|3|4)",
		},
		{
			name:              "Test `config get` with bad label",
			cmd:               "config get --labels=\"missingkey=missingvalue\"",
			expectInStdout:    "linuxConfig: {}",
			expectNotReStdout: "name: dummyif(0|1|2|3|4)",
		},
		{
			name:              "Test `config get` with bad label",
			cmd:               "config get --labels=\"missingkey\"",
			expectInStdout:    "linuxConfig: {}",
			expectNotReStdout: "name: dummyif(0|1|2|3|4)",
		},
		{
			name:              "Test `config get` with bad label",
			cmd:               "config get --labels=\"missingkey\",\"if=dummy\"",
			expectInStdout:    "linuxConfig: {}",
			expectNotReStdout: "name: dummyif(0|1|2|3|4)",
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
			expectInStderr: "no matching models found for [\"NoSuchModel\"]",
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
			expectReStdout: `MODEL[\s\|]+ORIGIN[\s\|]+VALUE[\s\|]+METADATA`,
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
			cmd:          `dump vpp.interfaces -f "{{range.}}Name:{{.Value.name}}{{end}}"`,
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
			name: "Test `generate` action to json",
			cmd:  "generate vpp.interfaces -f=json",
			expectJsonKeyVals: []KeyVal{
				{"type", "UNDEFINED_TYPE"},
			},
		},
		{
			name: "Test `generate` action to json (oneline)",
			cmd:  "generate vpp.interfaces -f=json --oneline",
			expectJsonKeyVals: []KeyVal{
				{"type", "UNDEFINED_TYPE"},
			},
		},
		{
			// This test depends on file (agentctl-config1.yaml) which was created before.
			name:           "Test `import` action",
			cmd:            "import " + config1File,
			expectErr:      true,
			expectInStderr: "connecting to Etcd failed",
		},
		{
			// This test depends on file (agentctl-config1.yaml) which was created before.
			name:         "Test `import` action (grpc)",
			cmd:          "import " + config1File + " --grpc",
			expectStdout: "importing 1 key-value pairs\n - config/vpp/v2/interfaces/tap1\nsending via gRPC\n",
		},
		{
			name:           "Test `kvdb list` action",
			cmd:            "kvdb list",
			expectErr:      true,
			expectInStderr: "connecting to Etcd failed",
		},
		{
			name:           "Test `log list` action",
			cmd:            "log list",
			expectReStdout: `agent\s+(trace|debug|info)`,
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
			name:           "Test `models` action",
			cmd:            "models",
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
			expectReStdout: `State:\s*OK`,
		},
		{
			name:         "Test `status` action (with format)",
			cmd:          "status -f {{.Status.AgentStatus.State}}",
			expectStdout: "OK",
		},
		{
			name:           "Test `values` action",
			cmd:            "values",
			expectReStdout: `vpp.interfaces\s+UNTAGGED-local0\s+obtained`,
		},
		{
			name:           "Test `values` action (with model)",
			cmd:            "values vpp.proxyarp-global",
			expectReStdout: `vpp.proxyarp-global\s+obtained `,
		},
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
			g := NewWithT(t)

			stdout, stderr, err = ctx.Agent.ExecCmd("agentctl", strings.Split(test.cmd, " ")...)

			if test.expectErr {
				g.Expect(err).To(HaveOccurred(),
					"Expected command `%s` to fail\n", test.cmd)
			} else {
				g.Expect(err).ToNot(HaveOccurred(),
					"Expected command `%s` not to fail, but failed with err: %v\nStderr:\n%s\n", test.cmd, err, stderr)
			}
			// Check STDOUT:
			if test.expectNotEmptyStdout {
				g.Expect(stdout).ToNot(BeEmpty(),
					"Stdout should not be empty\n")
			}
			if test.expectStdout != "" {
				g.Expect(stdout).To(Equal(test.expectStdout),
					"Expected output not equal stdout")
			}
			if test.expectInStdout != "" {
				g.Expect(stdout).To(ContainSubstring(test.expectInStdout),
					"Expected string not found in stdout")
			}
			if test.expectNotInStdout != "" {
				g.Expect(stdout).ToNot(ContainSubstring(test.expectNotInStdout),
					"Unexpected string found in stdout")
			}
			if test.expectJsonKeyVals != nil {
				var data map[string]interface{}
				err := json.Unmarshal([]byte(stdout), &data)
				if err != nil {
					t.Fatal(err)
				}
				var matchers []types.GomegaMatcher
				for _, kv := range test.expectJsonKeyVals {
					matchers = append(matchers, HaveKeyWithValue(kv.Key, kv.Value))
				}
				g.Expect(data).To(SatisfyAll(matchers...), "Expected key-value not found in JSON data from stdout")
			}
			if test.expectReStdout != "" {
				g.Expect(stdout).To(MatchRegexp(test.expectReStdout),
					"Expect regexp %q to match stdout for command %q, stdout:\n%s",
					test.expectReStdout, test.cmd, stdout)
			}
			if test.expectNotReStdout != "" {
				g.Expect(stdout).ToNot(MatchRegexp(test.expectNotReStdout),
					"Expect regexp %q to not match stdout for command %q, stdout:\n%s,",
					test.expectReStdout, test.cmd, stdout)
			}
			// Check STDERR:
			if test.expectInStderr != "" {
				g.Expect(stderr).To(ContainSubstring(test.expectInStderr),
					"Want in stderr: \n%s\nGot stderr: \n%s\n", test.expectInStderr, stderr)
			}
		})
	}
}

/*func TestAgentCtlSecureGrpcWithClientCertRequired(t *testing.T) {
	// WARNING: Do not use grpc connection created in `setupE2E` in
	// this test (though I don't know why you would but anyway).
	// By default `grpc.Dial` is non-blocking and connecting happens
	// in the background, so `setupE2E` function does not know about
	// any errors. With securing grpc on the agent (by replacing
	// grpc.conf with grpc-secure.conf) that client won't be able
	// to establish connection because it's not configured for this
	// secure case.

	t.Log("Replacing `GRPC_CONFIG` value with /etc/grpc-secure-full.conf")
	defer func(oldVal string) {
		t.Logf("Setting `GRPC_CONFIG` back to %q", oldVal)
		os.Setenv("GRPC_CONFIG", oldVal)
	}(os.Getenv("GRPC_CONFIG"))
	os.Setenv("GRPC_CONFIG", "/etc/grpc-secure-full.conf")

	ctx := Setup(t)
	defer ctx.Teardown()

	t.Log("Try without any TLS")
	_, stderr, err := ctx.ExecCmd(
		"/agentctl", "--debug", "dump", "vpp.interfaces",
	)
ctx.Expect(err).To(Not(BeNil()))
ctx.Expect(strings.Contains(stderr, "rpc error")).To(BeTrue(),
		"Want in stderr: \n\"rpc error\"\nGot stderr: \n%s\n", stderr,
	)

	t.Log("Try with TLS enabled via flag --insecure-tls, but without cert and key (note: server configured to check those files)")
	_, stderr, err = ctx.ExecCmd(
		"/agentctl", "--debug", "--insecure-tls", "dump", "vpp.interfaces",
	)
ctx.Expect(err).To(Not(BeNil()))
ctx.Expect(strings.Contains(stderr, "rpc error")).To(BeTrue(),
		"Want in stderr: \n\"rpc error\"\nGot stderr: \n%s\n", stderr,
	)

	t.Log("Try with fully configured TLS via config file")
	stdout, stderr, err := ctx.ExecCmd(
		"/agentctl", "--debug", "--config-dir=/etc/.agentctl", "dump", "vpp.interfaces",
	)
ctx.Expect(err).To(BeNil(),
		"Should not fail. Got err: %v\nStderr:\n%s\n", err, stderr,
	)
ctx.Expect(len(stdout)).To(Not(BeZero()))
}*/

func TestAgentCtlSecureGrpc(t *testing.T) {
	// WARNING: Do not use grpc connection created in `setupE2E` in
	// this test (though I don't know why you would but anyway).
	// By default `grpc.Dial` is non-blocking and connecting happens
	// in the background, so `setupE2E` function does not know about
	// any errors. With securing grpc on the agent (by replacing
	// grpc.conf with grpc-secure.conf) that client won't be able
	// to establish connection because it's not configured for this
	// secure case.

	t.Log("Replacing `GRPC_CONFIG` value with /testdata/grpc-secure.conf")
	defer func(oldVal string) {
		t.Logf("Setting `GRPC_CONFIG` back to %q", oldVal)
		os.Setenv("GRPC_CONFIG", oldVal)
	}(os.Getenv("GRPC_CONFIG"))
	os.Setenv("GRPC_CONFIG", "/testdata/grpc-secure.conf")

	ctx := Setup(t)
	defer ctx.Teardown()

	ctx.ExecCmd("bash", "-c", "set -x; ls /testdata; cat /testdata/agentctl.conf")

	t.Log("Try without any TLS")
	_, stderr, err := ctx.ExecCmd(
		"agentctl", "--debug", "dump", "vpp.interfaces")
	ctx.Expect(err).To(Not(BeNil()))
	ctx.Expect(stderr).To(ContainSubstring("rpc error"), "Expected string not found in stderr")

	t.Log("Try with TLS enabled via flag --insecure-tls. Should work because server is not configured to check client certs.")
	stdout, stderr, err := ctx.ExecCmd(
		"agentctl", "--debug", "--insecure-tls", "dump", "vpp.interfaces")
	ctx.Expect(err).To(Not(BeNil()))
	ctx.Expect(stdout).To(BeEmpty())
	ctx.Expect(stderr).To(ContainSubstring("dump failed:"))

	t.Log("Try with fully configured TLS via config file")
	stdout, stderr, err = ctx.ExecCmd(
		"agentctl", "--debug", "--config=/testdata/agentctl.conf", "dump", "vpp.interfaces")
	ctx.Expect(err).To(Not(BeNil()))
	ctx.Expect(stdout).To(BeEmpty())
	ctx.Expect(stderr).To(ContainSubstring("dump failed:"))
}

func TestAgentCtlSecureETCD(t *testing.T) {
	ctx := Setup(t, WithEtcd(WithEtcdHTTPsConnection(), WithEtcdTestContainerNetworking()))
	defer ctx.Teardown()

	// test without any TLS
	t.Run("no TLS", func(t *testing.T) {
		_, _, err := ctx.ExecCmd("agentctl", "--debug", "kvdb", "list")
		ctx.Expect(err).To(Not(BeNil()))
	})

	// test with TLS enabled via flag --insecure-tls, but without cert and key (note: server configured to check those files)
	t.Run("insecure TLS", func(t *testing.T) {
		_, _, err := ctx.ExecCmd("agentctl", "--debug", "--insecure-tls", "kvdb", "list")
		ctx.Expect(err).To(Not(BeNil()))
	})

	// test with fully configured TLS via config file
	/*t.Run("fully cofigured TLS", func(t *testing.T) {
	  	_, stderr, err := ctx.ExecCmd("/agentctl", "--debug", "--config-dir=/etc/.agentctl", "kvdb", "list")
	  ctx.Expect(err).To(BeNil(), "Should not fail. Got err: %v\nStderr:\n%s\n", err, stderr)
	  })*/
}

func createFileWithContent(path, content string) (string, error) {
	f, err := os.Create(path)
	if err != nil {
		return path, err
	}
	w := bufio.NewWriter(f)
	_, err = w.WriteString(content)
	if err != nil {
		return path, err
	}
	w.Flush()
	return path, nil
}

func dummyIfFactory(ctx *TestCtx) func() (string, string) {
	seq := 0
	return func() (string, string) {
		strseq := strconv.Itoa(seq)
		file := ctx.testShareDir + "/agentctl-dummyif" + strseq + ".yaml"
		content := `linuxConfig:
  interfaces:
  - name: "dummyif` + strseq + `"
    type: DUMMY
    enabled: true
    ipAddresses:
    - 9.9.9.9/24`
		seq += 1
		return file, content
	}
}
