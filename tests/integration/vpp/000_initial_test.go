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

package vpp

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
)

func TestPing(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	vpp := vppcalls.CompatibleHandler(test.vppClient)

	Expect(vpp.Ping(test.Ctx)).To(Succeed())

	session, err := vpp.GetSession(test.Ctx)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.PID).To(BeEquivalentTo(test.vppCmd.Process.Pid))
}

func TestGetVersion(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	vpp := vppcalls.CompatibleHandler(test.vppClient)

	info, err := vpp.GetVersion(test.Ctx)
	Expect(err).ToNot(HaveOccurred())
	Expect(info.Version).To(BePrintable(), "Version should be printable string:\n\t%#v", info)
}

func TestGetPlugins(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	vpp := vppcalls.CompatibleHandler(test.vppClient)

	plugins, err := vpp.GetPlugins(test.Ctx)
	Expect(err).ToNot(HaveOccurred())
	t.Logf("%d plugins: %v", len(plugins), plugins)
	Expect(plugins).ToNot(BeEmpty())

	// GetModules return empty list with VPP 20.01
	/*modules, err := vpp.GetModules(test.Ctx)
	Expect(err).ToNot(HaveOccurred())
	t.Logf("%d modules: %v", len(modules), modules)
	Expect(modules).ToNot(BeEmpty())*/
}
