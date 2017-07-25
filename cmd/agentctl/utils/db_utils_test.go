package utils_test

import (
	"github.com/ligato/cn-infra/statuscheck/model/status"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"github.com/onsi/gomega"
	"testing"
)

func TestCheckStatus(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, plugStatCfgRev := utils.
		ParseKey("/vnf-agent/{agent-label}/check/status/v1/agent")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(status.AgentStatusPrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{}))
	gomega.Expect(plugStatCfgRev).To(gomega.BeEquivalentTo(status.StatusPrefix))
}

func TestIfStatus(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/status/v1/interface/{interface-name}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(interfaces.IfStatePrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{interface-name}"}))
}

func TestIfConfig(t *testing.T) {
	gomega.RegisterTestingT(t)
	label, dataType, params, _ := utils.
		ParseKey("/vnf-agent/{agent-label}/vpp/config/v1/interface/{interface-name}")

	gomega.Expect(label).To(gomega.BeEquivalentTo("{agent-label}"))
	gomega.Expect(dataType).To(gomega.BeEquivalentTo(interfaces.InterfacePrefix))
	gomega.Expect(params).To(gomega.BeEquivalentTo([]string{"{interface-name}"}))
}
