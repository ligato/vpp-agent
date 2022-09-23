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
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func TestInfoVersionHandler(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	version, err := ctx.Agent.Client().AgentVersion(context.Background())
	ctx.Expect(err).ToNot(HaveOccurred())
	ctx.Expect(version.App).ToNot(BeEmpty())
	ctx.Expect(version.Version).ToNot(BeEmpty())
	ctx.Expect(version.GitCommit).ToNot(BeEmpty())
	ctx.Expect(version.GitBranch).ToNot(BeEmpty())
	ctx.Expect(version.BuildUser).ToNot(BeEmpty())
	ctx.Expect(version.BuildHost).ToNot(BeEmpty())
	ctx.Expect(version.BuildTime).ToNot(BeZero())
	ctx.Expect(version.GoVersion).ToNot(BeEmpty())
	ctx.Expect(version.OS).ToNot(BeEmpty())
	ctx.Expect(version.Arch).ToNot(BeEmpty())
}

func TestJsonschema(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	res, err := ctx.Agent.Client().HTTPClient().Get("http://" + ctx.Agent.Client().AgentHost() + ":9191/info/configuration/jsonschema")
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		t.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("BODY: %s", body)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatal(err)
	}
}
