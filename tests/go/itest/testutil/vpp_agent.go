package testutil

import (
	"git.fd.io/govpp.git/adapter/mock"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	localsync "github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/tests/go/itest/iftst"
)

// VppAgentT is similar to testing.T in golang packages.
type VppAgentT struct {
	//*testing.T
	agent *core.Agent
}

// Given is a composition of multiple test step methods (see BDD Given keyword).
type Given struct {
	//MockVpp *mock.VppAdapter
}

// When is a composition of multiple test step methods (see BDD When keyword).
type When struct {
	iftst.WhenIface
	//when_l2.WhenL2
	//when_l3.WhenL3
}

// Then is a composition of multiple test step methods (see BDD Then keyword).
type Then struct {
	iftst.ThenIface
	//then_l2.ThenL2
	//then_l3.ThenL3
}

// SetupDefault setups default behaviour of mocks and delegates to Setup(Flavor).
func (t *VppAgentT) SetupDefault() (flavor *VppOnlyTestingFlavor) {
	flavor = &VppOnlyTestingFlavor{
		IfStatePub: NewIfStatePub(),
		//GoVPP: *VppMock(t.MockVpp, iftst.RepliesSuccess /*, given_l3.RepliesSuccess*/),
	}

	//t.Setup(flavor)
	return flavor
}

// Setup registers gomega and starts the agent with the flavor argument.
func (t *VppAgentT) Setup(flavor core.Flavor) {
	//gomega.RegisterTestingT(t.T)

	t.agent = core.NewAgent(flavor)
	err := t.agent.Start()
	if err != nil {
		logrus.DefaultLogger().Panic(err)
	}
}

// Teardown stops the agent.
func (t *VppAgentT) Teardown() {
	if t.agent != nil {
		if err := t.agent.Stop(); err != nil {
			logrus.DefaultLogger().Panic(err)
		}
	}
}

// VppMock allows to mock go VPP plugin in a flavor.
func VppMock(vppMock *mock.VppAdapter, vppMockSetups ...func(adapter *mock.VppAdapter)) *govppmux.GOVPPPlugin {
	//vppMock := &mock.VppAdapter{}
	for _, vppMockSetup := range vppMockSetups {
		vppMockSetup(vppMock)
	}
	return govppmux.FromExistingAdapter(vppMock)
}

// VppOnlyTestingFlavor glues together multiple plugins to mange VPP and linux interfaces configuration using local client.
type VppOnlyTestingFlavor struct {
	*local.FlavorLocal

	IfStatePub *MockIfStatePub

	GoVPP govppmux.GOVPPPlugin
	VPP   defaultplugins.Plugin

	injected bool
}

// MockIfStatePub is mocking for interface state publishing.
type MockIfStatePub struct {
	states map[string]*intf.InterfacesState_Interface
}

// Put is mocked implementation for interface state publishing.
func (m *MockIfStatePub) Put(key string, data proto.Message, opts ...datasync.PutOption) error {
	logrus.DefaultLogger().Warnf("-> MyStatePub.Put(key: %v, data: %#v, opts: %v)", key, data, opts)
	//var state intf.InterfacesState_Interface
	if state, ok := data.(*intf.InterfacesState_Interface); ok {
		m.states[state.Name] = state
	} else {
		logrus.DefaultLogger().Warnf("invalid type received")
	}
	return nil
}

// InterfaceState returns state from mocked interface state publisher.
func (m *MockIfStatePub) InterfaceState(ifaceName string, ifState *intf.InterfacesState_Interface) (bool, error) {
	state, found := m.states[ifaceName]
	logrus.DefaultLogger().Warnf("-> InterfaceState(%v) - %+v", ifaceName, state)
	if found {
		*ifState = *state
	}
	return found, nil
}

// NewIfStatePub returns new instance of MockIfStatePub.
func NewIfStatePub() *MockIfStatePub {
	return &MockIfStatePub{
		states: make(map[string]*intf.InterfacesState_Interface),
	}
}

// Inject sets object references.
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
	f.VPP.Deps.IfStatePub = f.IfStatePub
	//f.VPP.Deps.Publish = StatePub

	//TODO f.VPP.Deps.Publish = local_sync.Get()

	return false
}

// Plugins combines Generic Plugins and Standard VPP Plugins.
func (f *VppOnlyTestingFlavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}

/*
// SetupDefault setups default behaviour of mocks and delegates to Setup(Flavor).
func (t *VppAgentT) SetupDefault() (flavor *VppOnlyTestingFlavor) {
	flavor = &VppOnlyTestingFlavor{
		GoVPP: *VppMock(iftst.RepliesSuccess),
	}

	t.Setup(flavor)

	return flavor
}

// Setup registers gomega and starts the agent with the flavor argument.
func (t *VppAgentT) Setup(flavor core.Flavor) {
	//gomega.RegisterTestingT(t.T)

	t.agent = core.NewAgent(flavor)
	err := t.agent.Start()
	if err != nil {
		logrus.DefaultLogger().Panic(err)
	}
}

// Teardown stops the agent.
func (t *VppAgentT) Teardown() {
	if t.agent != nil {
		err := t.agent.Stop()
		if err != nil {
			logrus.DefaultLogger().Panic(err)
		}
	}
}

// VppMock allows to mock go VPP plugin in a flavor.
func VppMock(vppMockSetups ...func(adapter *mock.VppAdapter)) *govppmux.GOVPPPlugin {
	vppMock := &mock.VppAdapter{}
	for _, vppMockSetup := range vppMockSetups {
		vppMockSetup(vppMock)
	}
	return govppmux.FromExistingAdapter(vppMock)
}
*/
