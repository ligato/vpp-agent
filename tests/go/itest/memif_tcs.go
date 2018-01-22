package itest

import (
	"testing"

	"git.fd.io/govpp.git/adapter/mock"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/localclient"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/tests/go/itest/iftst"
	"github.com/ligato/vpp-agent/tests/go/itest/testutil"
)

type suiteMemif struct {
	T *testing.T
	testutil.VppAgentT
	testutil.Given
	testutil.When
	testutil.Then
}

func forMemif(t *testing.T) *suiteMemif {
	return &suiteMemif{T: t,
		When: testutil.When{
			WhenIface: iftst.WhenIface{
				Log:       testutil.NewLogger("WhenIface", t),
				NewChange: localclient.DataChangeRequest,
				NewResync: localclient.DataResyncRequest,
			}},
		Then: testutil.Then{
			ThenIface: iftst.ThenIface{
				Log: testutil.NewLogger("ThenIface", t),
				//NewChange: localclient.DataChangeRequest,
				//OperState: testutil.NewStatePub(),
			}},
	}
}

func (s *suiteMemif) setupTestingFlavor(flavor *testutil.VppOnlyTestingFlavor) {
	local.DefaultTransport = syncbase.NewRegistry()
	mockVpp := &mock.VppAdapter{}
	flavor.GoVPP = *testutil.VppMock(mockVpp, iftst.RepliesSuccess)
	//mockVpp.MockReplyHandler(iftst.VppMockHandler(mockVpp))
	/*s.When.NewChange = func(caller core.PluginName) defaultplugins.DataChangeDSL {
		return dbadapter.NewDataChangeDSL(local.NewProtoTxn(local.Get().PropagateChanges))
	}*/
	s.Setup(flavor)
	s.When.VPP = &flavor.VPP
	s.When.MockVpp = mockVpp
	s.Then.VPP = &flavor.VPP
	s.Then.OperState = flavor.IfStatePub
}

// TC01EmptyVppCrudEtcd asserts that data written to ETCD after Agent Starts are processed.
func (s *suiteMemif) TC01EmptyVppCrudEtcd() {
	s.setupTestingFlavor(s.SetupDefault())
	defer s.Teardown()

	s.When.StoreIf(&iftst.Memif100011Slave)
	s.Then.SwIfIndexes().ContainsName(iftst.Memif100011Slave.Name)

	s.When.StoreIf(&iftst.Memif100012)

	s.Then.SwIfIndexes().ContainsName(iftst.Memif100012.Name)

	s.When.DelIf(&iftst.Memif100012)
	s.Then.SwIfIndexes().NotContainsName(iftst.Memif100012.Name)

	//TODO simulate that dump return local interface
}

// TC02EmptyVppResyncAtStartup tests that data written to ETCD before Agent Starts are processed (startup RESYNC).
func (s *suiteMemif) TC02EmptyVppResyncAtStartup() {
	s.setupTestingFlavor(s.SetupDefault())
	defer s.Teardown()

	s.When.ResyncIf(&iftst.Memif100011Slave)
	s.When.ResyncIf(&iftst.Memif100012)

	s.Then.SwIfIndexes().ContainsName(iftst.Memif100011Slave.Name)
	s.Then.SwIfIndexes().ContainsName(iftst.Memif100012.Name)
}

// TC03VppNotificaitonIfDown tests that if state down notification is handled correctly
func (s *suiteMemif) TC03VppNotificaitonIfDown() {
	s.setupTestingFlavor(s.SetupDefault())
	defer s.Teardown()

	s.When.StoreIf(&iftst.Memif100011Slave)
	s.When.StoreIf(&iftst.Memif100012)
	s.Then.SwIfIndexes().ContainsName(iftst.Memif100011Slave.Name)

	s.When.VppLinkDown(&iftst.Memif100011Slave)

	s.Then.IfStateInDB(interfaces.InterfacesState_Interface_DOWN, &iftst.Memif100011Slave)

	s.When.VppLinkDown(&iftst.Memif100012)

	s.Then.IfStateInDB(interfaces.InterfacesState_Interface_DOWN, &iftst.Memif100012)
	s.Then.IfStateInDB(interfaces.InterfacesState_Interface_DOWN, &iftst.Memif100011Slave)

	s.When.VppLinkUp(&iftst.Memif100012)

	s.Then.IfStateInDB(interfaces.InterfacesState_Interface_UP, &iftst.Memif100012)
	s.Then.IfStateInDB(interfaces.InterfacesState_Interface_DOWN, &iftst.Memif100011Slave)

	s.When.VppLinkUp(&iftst.Memif100011Slave)

	s.Then.IfStateInDB(interfaces.InterfacesState_Interface_UP, &iftst.Memif100011Slave)
	s.Then.IfStateInDB(interfaces.InterfacesState_Interface_UP, &iftst.Memif100012)
}
