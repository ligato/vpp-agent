package testutil

import (
	"git.fd.io/govpp.git/adapter/mock"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/cn-infra/core"
	"testing"
	"github.com/onsi/gomega"
	"github.com/ligato/cn-infra/logging/logroot"
	localsync "github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"time"
	"github.com/ligato/vpp-agent/tests/go/itest/iftst"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
)

// VppAgentT is simmilar to what testing.T is in golang packages.
type VppAgentT struct {
	*testing.T
	agent *core.Agent
}

// Given is composition of multiple test step methods (see BDD Given keyword)
type Given struct {
}

// When is composition of multiple test step methods (see BDD When keyword)
type When struct {
	iftst.WhenIface
	//when_l2.WhenL2
	//when_l3.WhenL3
}

// Then is composition of multiple test step methods (see BDD Then keyword)
type Then struct {
	iftst.ThenIface
	//then_l2.ThenL2
	//then_l3.ThenL3
}

// VppOnlyTestingFlavor glues together multiple plugins to mange VPP and linux interfaces configuration using local client.
type VppOnlyTestingFlavor struct {
	*local.FlavorLocal
	GoVPP    govppmux.GOVPPPlugin
	VPP      defaultplugins.Plugin
	injected bool
}

// Inject sets object references
func (f *VppOnlyTestingFlavor) Inject() bool {
	if f.injected {
		return true
	}
	f.injected = true

	if f.FlavorLocal == nil {
		f.FlavorLocal = &local.FlavorLocal{}
	}
	f.FlavorLocal.Inject()

	f.GoVPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("govpp")
	f.VPP.Deps.PluginInfraDeps = *f.FlavorLocal.InfraDeps("default-plugins")
	//nil: f.VPP.Deps.Linux
	f.VPP.Deps.GoVppmux = &f.GoVPP
	f.VPP.Deps.Watch = localsync.Get()
	//nil: f.VPP.Deps.Messaging

	//TODO f.VPP.Deps.Publish = local_sync.Get()

	return false
}

// Plugins combines Generic Plugins and Standard VPP Plugins
func (f *VppOnlyTestingFlavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

// SetupDefault setups default behaviour of mocks and delegates to Setup(Flavor)
func (t *VppAgentT) SetupDefault() (flavor *VppOnlyTestingFlavor) {
	flavor = &VppOnlyTestingFlavor{
		GoVPP: *VppMock(iftst.RepliesSuccess /*, given_l3.RepliesSuccess*/),
	}

	t.Setup(flavor)

	return flavor
}

// Setup registers gomega and starts the agent with the flavor argument
func (t *VppAgentT) Setup(flavor core.Flavor) {
	gomega.RegisterTestingT(t.T)

	agent := core.NewAgent(logroot.StandardLogger(), 2000*time.Second, flavor.Plugins()...)
	err := agent.Start()
	if err != nil {
		logroot.StandardLogger().Panic(err)
	}
}

// Teardown stops the agent
func (t *VppAgentT) Teardown() {
	if t.agent != nil {
		err := t.agent.Stop()
		if err != nil {
			logroot.StandardLogger().Panic(err)
		}
	}
}

// VppMock allows to flavors the Vpp Mock
func VppMock(vppMockSetups ... func(adapter *mock.VppAdapter)) *govppmux.GOVPPPlugin {
	vppMock := &mock.VppAdapter{}
	for _, vppMockSetup := range vppMockSetups {
		vppMockSetup(vppMock)
	}
	return govppmux.FromExistingAdapter(vppMock)
}
