package itest

import (
	"testing"

	"git.fd.io/govpp.git/adapter/mock"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/localclient"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/ipsec"
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

// TC04
func (s *suiteMemif) TC04() {
	s.setupTestingFlavor(s.SetupDefault())
	defer s.Teardown()

	s.When.Put(func(put defaultplugins.PutDSL) defaultplugins.PutDSL {
		return put.IPSecSA(&IPsecSA20)
	})
	s.When.Put(func(put defaultplugins.PutDSL) defaultplugins.PutDSL {
		return put.IPSecSA(&IPsecSA10)
	})
	s.When.Put(func(put defaultplugins.PutDSL) defaultplugins.PutDSL {
		return put.IPSecSPD(&IPsecSPD1)
	})
	s.Then.ContainsIPSecSA(IPsecSA10.Name)
	s.Then.ContainsIPSecSA(IPsecSA20.Name)
	s.Then.ContainsIPSecSPD(IPsecSPD1.Name)

	s.When.StoreIf(&iftst.Memif100011Master)
	s.Then.SwIfIndexes().ContainsName(iftst.Memif100011Master.Name)
	s.When.DelIf(&iftst.Memif100011Master)
	s.When.StoreIf(&iftst.Memif100011Master)
}

var IPsecSA10 = ipsec.SecurityAssociations_SA{
	Name:      "sa10",
	Spi:       1001,
	Protocol:  ipsec.SecurityAssociations_SA_ESP,
	CryptoAlg: ipsec.CryptoAlgorithm_AES_CBC_128,
	CryptoKey: "4a506a794f574265564551694d653768",
	IntegAlg:  ipsec.IntegAlgorithm_SHA1_96,
	IntegKey:  "4339314b55523947594d6d3547666b45764e6a58",
}

var IPsecSA20 = ipsec.SecurityAssociations_SA{
	Name:      "sa20",
	Spi:       1000,
	Protocol:  ipsec.SecurityAssociations_SA_ESP,
	CryptoAlg: ipsec.CryptoAlgorithm_AES_CBC_128,
	CryptoKey: "4a506a794f574265564551694d653768",
	IntegAlg:  ipsec.IntegAlgorithm_SHA1_96,
	IntegKey:  "4339314b55523947594d6d3547666b45764e6a58",
}

var IPsecSPD1 = ipsec.SecurityPolicyDatabases_SPD{
	Name: "spd1",
	Interfaces: []*ipsec.SecurityPolicyDatabases_SPD_Interface{
		{Name: "memif1"},
	},
	PolicyEntries: []*ipsec.SecurityPolicyDatabases_SPD_PolicyEntry{
		{
			Priority:   100,
			IsOutbound: false,
			Action:     ipsec.SecurityPolicyDatabases_SPD_PolicyEntry_BYPASS,
			Protocol:   50,
		}, {
			Priority:   100,
			IsOutbound: true,
			Action:     ipsec.SecurityPolicyDatabases_SPD_PolicyEntry_BYPASS,
			Protocol:   50,
		}, {
			Priority:        10,
			IsOutbound:      false,
			Action:          ipsec.SecurityPolicyDatabases_SPD_PolicyEntry_PROTECT,
			RemoteAddrStart: "10.0.0.1",
			RemoteAddrStop:  "10.0.0.1",
			LocalAddrStart:  "10.0.0.2",
			LocalAddrStop:   "10.0.0.2",
			Sa:              "sa20",
		}, {
			Priority:        10,
			IsOutbound:      true,
			Action:          ipsec.SecurityPolicyDatabases_SPD_PolicyEntry_PROTECT,
			RemoteAddrStart: "10.0.0.1",
			RemoteAddrStop:  "10.0.0.1",
			LocalAddrStart:  "10.0.0.2",
			LocalAddrStop:   "10.0.0.2",
			Sa:              "sa10",
		},
	},
}
