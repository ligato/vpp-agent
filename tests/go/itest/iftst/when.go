package iftst

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	vppclient "github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/localclient"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
)

const pluginName = core.PluginName("when_iface")

// WhenIface is a collection of test step methods (see Behavior Driven Development)
// (methods that will be called from test scenarios).
type WhenIface struct {
	NewChange func(name core.PluginName) vppclient.DataChangeDSL
	Log       logging.Logger
}

// StoreIf stores configuration of a given interface in ETCD.
func (step *WhenIface) StoreIf(data *intf.Interfaces_Interface, opts ...interface{}) {
	step.Log.Debug("When_StoreIf begin")
	err := step.NewChange(pluginName).Put().Interface(data).Send().ReceiveReply()
	if err != nil {
		step.Log.Panic(err)
	}
	step.Log.Debug("When_StoreIf end")
}

// DelIf removes configuration of a given interface from ETCD.
func (step *WhenIface) DelIf(data *intf.Interfaces_Interface) {
	step.Log.Debug("When_StoreIf begin")
	err := localclient.DataChangeRequest(pluginName).Delete().Interface(data.Name).Send().ReceiveReply()
	if err != nil {
		step.Log.Panic(err)
	}
	step.Log.Debug("When_StoreIf end")
}

/*
// VppLinkDown sends SwInterfaceSetFlags{LinkUpDown: 0} using VPP mock.
func (step *WhenIface) VppLinkDown(data *intf.Interfaces_Interface, given *testing.GivenAndKW) {
	given.And().VppMock(func(mockVpp *govppmock.VppAdapter) {
		n := data.Name
		idx, _, _ := defaultplugins.GetSwIfIndexes().LookupIdx(n)

		// mock the notification and force its delivery
		mockVpp.MockReply(&interfaces.SwInterfaceSetFlags{
			SwIfIndex:  idx,
			LinkUpDown: 0,
		})
		mockVpp.SendMsg(0, []byte(""))

	})
}

// VppLinkUp sends SwInterfaceSetFlags{LinkUpDown: 1} using VPP mock.
func (step *WhenIface) VppLinkUp(data *intf.Interfaces_Interface, given *testing.GivenAndKW) {
	given.And().VppMock(func(mockVpp *govppmock.VppAdapter) {
		n := data.Name
		idx, _, _ := defaultplugins.GetSwIfIndexes().LookupIdx(n)

		// mock the notification and force its delivery
		mockVpp.MockReply(&interfaces.SwInterfaceSetFlags{
			SwIfIndex:  idx,
			LinkUpDown: 0,
		})
		mockVpp.SendMsg(0, []byte(""))

	})
}

// StoreBfdSession stores configuration of a given BFD session in ETCD.
func (step *WhenIface) StoreBfdSession(data *bfd.SingleHopBFD_Session) {
	log.Debug("When_StoreBfdSession begin")
	k := bfd.SessionKey(data.Interface)
	err := etcdmux.NewRootBroker().Put(servicelabel.GetAgentPrefix()+k, data)
	if err != nil {
		log.Panic(err)
	}
	log.Debug("When_StoreBfdSession end")
}

// StoreBfdAuthKey stores configuration of a given BFD key in ETCD.
func (step *WhenIface) StoreBfdAuthKey(data *bfd.SingleHopBFD_Key) {
	log.Debug("When_StoreBfdKey begin")
	k := bfd.AuthKeysKey(strconv.FormatUint(uint64(data.Id), 10))
	err := etcdmux.NewRootBroker().Put(servicelabel.GetAgentPrefix()+k, data)
	if err != nil {
		log.Panic(err)
	}
	log.Debug("When_StoreBfdKey end")
}

// StoreBfdEchoFunction stores configuration of a given BFD echo function in ETCD.
func (step *WhenIface) StoreBfdEchoFunction(data *bfd.SingleHopBFD_EchoFunction) {
	log.Debug("When_StoreBfdEchoFunction begin")
	k := bfd.EchoFunctionKey(data.EchoSourceInterface)
	err := etcdmux.NewRootBroker().Put(servicelabel.GetAgentPrefix()+k, data)
	if err != nil {
		log.Panic(err)
	}
	log.Debug("When_StoreBfdEchoFunction end")
}

// DelBfdSession removes configuration of a given BFD session from ETCD.
func (step *WhenIface) DelBfdSession(data *bfd.SingleHopBFD_Session) {
	log.Debug("When_DelBfdSession begin")
	k := bfd.SessionKey(data.Interface)
	_, err := etcdmux.NewRootBroker().Delete(servicelabel.GetAgentPrefix() + k)
	if err != nil {
		log.Panic(err)
	}
	log.Debug("When_DelBfdSession end")
}

// DelBfdAuthKey removes configuration of a given BFD key from ETCD.
func (step *WhenIface) DelBfdAuthKey(data *bfd.SingleHopBFD_Key) {
	log.Debug("When_DelBfdKey begin")
	k := bfd.AuthKeysKey(strconv.FormatUint(uint64(data.Id), 10))
	_, err := etcdmux.NewRootBroker().Delete(servicelabel.GetAgentPrefix() + k)
	if err != nil {
		log.Panic(err)
	}
	log.Debug("When_DelBfdKey end")
}

// DelBfdEchoFunction removes configuration of a given BFD echo function from ETCD.
func (step *WhenIface) DelBfdEchoFunction(data *bfd.SingleHopBFD_EchoFunction) {
	log.Debug("When_DelBfdEchoFunction begin")
	k := bfd.EchoFunctionKey(data.EchoSourceInterface)
	_, err := etcdmux.NewRootBroker().Delete(servicelabel.GetAgentPrefix() + k)
	if err != nil {
		log.Panic(err)
	}
	log.Debug("When_DelBfdEchoFunction end")
}
*/
