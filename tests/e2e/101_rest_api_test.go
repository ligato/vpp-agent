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
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestInfoVersionHandler(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	resp, err := ctx.httpClient.GET("/info/version")
	Expect(err).ToNot(HaveOccurred())
	var version = struct {
		App       string
		Version   string
		GitCommit string
		GitBranch string
		BuildUser string
		BuildHost string
		BuildTime int64
		GoVersion string
		OS        string
		Arch      string
	}{}
	err = json.Unmarshal(resp, &version)
	Expect(err).ToNot(HaveOccurred())
	Expect(version.App).ToNot(BeEmpty())
	Expect(version.Version).ToNot(BeEmpty())
	Expect(version.GitCommit).ToNot(BeEmpty())
	Expect(version.GitBranch).ToNot(BeEmpty())
	Expect(version.BuildUser).ToNot(BeEmpty())
	Expect(version.BuildHost).ToNot(BeEmpty())
	Expect(version.BuildTime).ToNot(BeZero())
	Expect(version.GoVersion).ToNot(BeEmpty())
	Expect(version.OS).ToNot(BeEmpty())
	Expect(version.Arch).ToNot(BeEmpty())
}
