package iftst

import (
	govppmock "git.fd.io/govpp.git/adapter/mock"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	vppclient "github.com/ligato/vpp-agent/clientv1/vpp"
	"github.com/ligato/vpp-agent/clientv1/vpp/localclient"
	"github.com/ligato/vpp-agent/plugins/vpp"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
)

const pluginName = core.PluginName("when_iface")

// WhenIface is a collection of test step methods (see Behavior Driven Development)
// (methods that will be called from test scenarios).
type WhenIface struct {
	NewChange func(name core.PluginName) vppclient.DataChangeDSL
	NewResync func(name core.PluginName) vppclient.DataResyncDSL
	Log       logging.Logger
	VPP       vpp.API
	MockVpp   *govppmock.VppAdapter
}

// ResyncIf stores configuration of a given interface in ETCD.
func (when *WhenIface) ResyncIf(data1, data2 *intf.Interfaces_Interface) {
	when.Log.Debug("When_ResyncIf begin")
	err := when.NewResync(pluginName).Interface(data1).Interface(data2).Send().ReceiveReply()
	if err != nil {
		when.Log.Panic(err)
	}
	when.Log.Debug("When_ResyncIf end")
}

// StoreIf stores configuration of a given interface in ETCD.
func (when *WhenIface) StoreIf(data *intf.Interfaces_Interface, opts ...interface{}) {
	when.Log.Debug("When_StoreIf begin")
	err := when.NewChange(pluginName).Put().Interface(data).Send().ReceiveReply()
	if err != nil {
		when.Log.Panic(err)
	}
	when.Log.Debug("When_StoreIf end")
}

// StoreIf stores configuration of a given interface in ETCD.
func (when *WhenIface) Put(fn func(dsl vppclient.PutDSL) vppclient.PutDSL) {
	err := fn(when.NewChange(pluginName).Put()).Send().ReceiveReply()
	if err != nil {
		when.Log.Panic(err)
	}
}

// DelIf removes configuration of a given interface from ETCD.
func (when *WhenIface) DelIf(data *intf.Interfaces_Interface) {
	when.Log.Debug("When_StoreIf begin")
	err := localclient.DataChangeRequest(pluginName).Delete().Interface(data.Name).Send().ReceiveReply()
	if err != nil {
		when.Log.Panic(err)
	}
	when.Log.Debug("When_StoreIf end")
}

// VppLinkUp sends interface event link up using VPP mock.
func (when *WhenIface) VppLinkUp(data *intf.Interfaces_Interface) {
	idx, _, exists := when.VPP.GetSwIfIndexes().LookupIdx(data.Name)
	if !exists {
		when.Log.Panicf("swIfIndex for %q doesnt exist", data.Name)
	}

	logrus.DefaultLogger().Infof("- VppLinkUp idx:%v", idx)

	// mock the notification and force its delivery
	when.MockVpp.MockReply(&interfaces.SwInterfaceEvent{
		SwIfIndex: idx,
		//AdminUpDown: 1,
		LinkUpDown: 1,
	})
	when.MockVpp.SendMsg(0, []byte(""))

	logrus.DefaultLogger().Info("~ VppLinkUp")
}

// VppLinkDown sends interface event link down using VPP mock.
func (when *WhenIface) VppLinkDown(data *intf.Interfaces_Interface) {
	idx, _, exists := when.VPP.GetSwIfIndexes().LookupIdx(data.Name)
	if !exists {
		when.Log.Panicf("swIfIndex for %q doesnt exist", data.Name)
	}

	logrus.DefaultLogger().Info("- VppLinkDown")

	// mock the notification and force its delivery
	when.MockVpp.MockReply(&interfaces.SwInterfaceEvent{
		SwIfIndex:  idx,
		LinkUpDown: 0,
	})
	when.MockVpp.SendMsg(0, []byte(""))

	logrus.DefaultLogger().Info("~ VppLinkDown")
}

/*
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
