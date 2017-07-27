// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils_test

import (
	"github.com/ligato/cn-infra/statuscheck/model/status"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
	"github.com/onsi/gomega"
	"testing"
)

func TestParseKeyAgentPrefix(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, plugStatCfgRev := utils.
		ParseKey("/vnf-agent/{agent-label}/check/status/v1/agent")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(status.AgentStatusPrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{}))
	gomega.Expect(plugStatCfgRev).To(gomega.BeEquivalentTo(status.StatusPrefix))
}

func TestParseKeyInterfaceConfig(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/config/v1/interface/{interface-name}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(interfaces.InterfacePrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{interface-name}"}))
}

func TestParseKeyInterfaceStatus(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/status/v1/interface/{interface-name}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(interfaces.IfStatePrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{interface-name}"}))
}

func TestParseKeyInterfaceError(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/status/v1/interface/error/{interface-name}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(interfaces.IfErrorPrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{interface-name}"}))
}

func TestParseKeyBdConfig(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/config/v1/bd/{bd-name}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(l2.BdPrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{bd-name}"}))
}

func TestParseKeyBdState(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/status/v1/bd/{bd-name}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(l2.BdStatePrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{bd-name}"}))
}

func TestParseKeyBdError(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/status/v1/bd/error/{bd-name}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(l2.BdErrPrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{bd-name}"}))
}

func TestParseKeyFib(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/config/v1/bd/fib/{mac-address}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(l2.FIBPrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{mac-address}"}))
}
