package itest

import (
	//test_if "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/testing"
	"github.com/ligato/vpp-agent/tests/go/itest/iftst"
	"github.com/ligato/vpp-agent/tests/go/itest/testutil"
	"testing"
)

type suiteMemif struct {
	T *testing.T
	testutil.VppAgentT
	testutil.Given
	testutil.When
	testutil.Then
}

func (t *suiteMemif) SetupTestingFlavor(flavor *testutil.VppOnlyTestingFlavor) {
	t.Then.VPP = &flavor.VPP
}

// TC01EmptyVppCrudEtcd asserts that data written to ETCD after Agent Starts are processed.
func (t *suiteMemif) TC01EmptyVppCrudEtcd() {
	t.SetupTestingFlavor(t.SetupDefault())
	defer t.Teardown()

	t.When.StoreIf(&iftst.Memif100011Slave)
	t.Then.SwIfIndexes().ContainsName(iftst.Memif100011Slave.Name)

	t.When.StoreIf(&iftst.Memif100012)

	t.Then.SwIfIndexes().ContainsName(iftst.Memif100012.Name)

	t.When.DelIf(&iftst.Memif100012)
	t.Then.SwIfIndexes().NotContainsName(iftst.Memif100012.Name)

	//TODO simulate that dump return local interface
}

/*
// TC02EmptyVppResyncAtStartup tests that data written to ETCD before Agent Starts are processed (startup RESYNC).
func (t *suiteMemif) TC02EmptyVppResyncAtStartup() {
	t.Given(t.T).VppMock(given.RepliesSuccess).
		And().StartedAgent(append(Plugins(), Init("ETCD before startup", func() error {
		t.When.StoreIf(&Memif100011Slave)
		t.When.StoreIf(&Memif100012)
		return nil
	})))
	defer Teardown(t.T)

	t.Then.SwIfIndexes().ContainsName(iftst.Memif100011Slave.Name)
	t.Then.SwIfIndexes().ContainsName(iftst.Memif100012.Name)
}

//suiteMemif03VppNotificaitonIfDown test that if state down notification is handled correctly
func (t *suiteMemif) TC03VppNotificaitonIfDown() {
	ctx := Given(t.T).VppMock(given.RepliesSuccess).
		And().StartedAgent(Plugins())
	defer Teardown(t.T)
	t.When.StoreIf(&iftst.Memif100011Slave)
	t.When.StoreIf(&iftst.Memif100012)
	t.Then.SwIfIndexes().ContainsName(iftst.Memif100011Slave.Name)

	t.When.VppLinkDown(&iftst.Memif100011Slave, ctx)

	t.Then.IfStateInDB(intf.InterfacesState_Interface_DOWN, &iftst.Memif100011Slave)

	t.When.VppLinkDown(&iftst.Memif100012, ctx)

	t.Then.IfStateInDB(intf.InterfacesState_Interface_DOWN, &iftst.Memif100012)
	t.Then.IfStateInDB(intf.InterfacesState_Interface_DOWN, &iftst.Memif100011Slave)

	t.When.VppLinkUp(&iftst.Memif100012, ctx)

	t.Then.IfStateInDB(intf.InterfacesState_Interface_UP, &iftst.Memif100012)
	t.Then.IfStateInDB(intf.InterfacesState_Interface_DOWN, &iftst.Memif100011Slave)

	t.When.VppLinkUp(&iftst.Memif100011Slave, ctx)

	t.Then.IfStateInDB(intf.InterfacesState_Interface_UP, &iftst.Memif100011Slave)
	t.Then.IfStateInDB(intf.InterfacesState_Interface_UP, &iftst.Memif100012)
}
*/
