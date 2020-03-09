// Copyright (c) 2018 Cisco and/or its affiliates.
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

package dbadapter

import (
	"go.ligato.io/cn-infra/v2/db/keyval"

	linuxclient "go.ligato.io/vpp-agent/v3/clientv2/linux"
	vppclient "go.ligato.io/vpp-agent/v3/clientv2/vpp"
	"go.ligato.io/vpp-agent/v3/clientv2/vpp/dbadapter"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_iptables "go.ligato.io/vpp-agent/v3/proto/ligato/linux/iptables"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
	acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
	punt "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/punt"
	stn "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/stn"
)

// NewDataResyncDSL returns a new instance of DataResyncDSL which implements
// the data RESYNC DSL for both Linux and VPP config (inherits dbadapter
// from vppplugin).
// Transaction <txn> is used to propagate changes to plugins.
// Function <listKeys> is used to list keys with already existing configuration.
func NewDataResyncDSL(txn keyval.ProtoTxn, listKeys func(prefix string) (keyval.ProtoKeyIterator, error)) *DataResyncDSL {
	vppDataResync := dbadapter.NewDataResyncDSL(txn, listKeys)
	return &DataResyncDSL{txn, []string{}, listKeys, vppDataResync}
}

// DataResyncDSL is an implementation of Domain Specific Language (DSL) for data
// RESYNC of both Linux and VPP configuration.
type DataResyncDSL struct {
	txn      keyval.ProtoTxn
	txnKeys  []string
	listKeys func(prefix string) (keyval.ProtoKeyIterator, error)

	vppDataResync vppclient.DataResyncDSL
}

// LinuxInterface adds Linux interface to the RESYNC request.
func (dsl *DataResyncDSL) LinuxInterface(val *linux_interfaces.Interface) linuxclient.DataResyncDSL {
	key := linux_interfaces.InterfaceKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// LinuxArpEntry adds Linux ARP entry to the RESYNC request.
func (dsl *DataResyncDSL) LinuxArpEntry(val *linux_l3.ARPEntry) linuxclient.DataResyncDSL {
	key := linux_l3.ArpKey(val.Interface, val.IpAddress)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// LinuxRoute adds Linux route to the RESYNC request.
func (dsl *DataResyncDSL) LinuxRoute(val *linux_l3.Route) linuxclient.DataResyncDSL {
	key := linux_l3.RouteKey(val.DstNetwork, val.OutgoingInterface)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// IptablesRuleChain adds iptables rule chain to the RESYNC request.
func (dsl *DataResyncDSL) IptablesRuleChain(val *linux_iptables.RuleChain) linuxclient.DataResyncDSL {
	key := linux_iptables.RuleChainKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// VppInterface adds VPP interface to the RESYNC request.
func (dsl *DataResyncDSL) VppInterface(intf *interfaces.Interface) linuxclient.DataResyncDSL {
	dsl.vppDataResync.Interface(intf)
	return dsl
}

// Span adds VPP span to the RESYNC request.
func (dsl *DataResyncDSL) Span(span *interfaces.Span) linuxclient.DataResyncDSL {
	dsl.vppDataResync.Span(span)
	return dsl
}

// ACL adds VPP Access Control List to the RESYNC request.
func (dsl *DataResyncDSL) ACL(acl *acl.ACL) linuxclient.DataResyncDSL {
	dsl.vppDataResync.ACL(acl)
	return dsl
}

// ABF adds ACL-based forwarding to the RESYNC request.
func (dsl *DataResyncDSL) ABF(abf *abf.ABF) linuxclient.DataResyncDSL {
	dsl.vppDataResync.ABF(abf)
	return dsl
}

/*// BfdSession adds VPP bidirectional forwarding detection session
// to the RESYNC request.
func (dsl *DataResyncDSL) BfdSession(val *bfd.SingleHopBFD_Session) linuxclient.DataResyncDSL {
	dsl.vppDataResync.BfdSession(val)
	return dsl
}

// BfdAuthKeys adds VPP bidirectional forwarding detection key to the RESYNC
// request.
func (dsl *DataResyncDSL) BfdAuthKeys(val *bfd.SingleHopBFD_Key) linuxclient.DataResyncDSL {
	dsl.vppDataResync.BfdAuthKeys(val)
	return dsl
}

// BfdEchoFunction adds VPP bidirectional forwarding detection echo function
// to the RESYNC request.
func (dsl *DataResyncDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) linuxclient.DataResyncDSL {
	dsl.vppDataResync.BfdEchoFunction(val)
	return dsl
}*/

// BD adds VPP Bridge Domain to the RESYNC request.
func (dsl *DataResyncDSL) BD(bd *l2.BridgeDomain) linuxclient.DataResyncDSL {
	dsl.vppDataResync.BD(bd)
	return dsl
}

// BDFIB adds VPP L2 FIB to the RESYNC request.
func (dsl *DataResyncDSL) BDFIB(fib *l2.FIBEntry) linuxclient.DataResyncDSL {
	dsl.vppDataResync.BDFIB(fib)
	return dsl
}

// XConnect adds VPP Cross Connect to the RESYNC request.
func (dsl *DataResyncDSL) XConnect(xcon *l2.XConnectPair) linuxclient.DataResyncDSL {
	dsl.vppDataResync.XConnect(xcon)
	return dsl
}

// VrfTable adds VPP VRF table to the RESYNC request.
func (dsl *DataResyncDSL) VrfTable(vrfTable *l3.VrfTable) linuxclient.DataResyncDSL {
	dsl.vppDataResync.VrfTable(vrfTable)
	return dsl
}

// StaticRoute adds VPP L3 Static Route to the RESYNC request.
func (dsl *DataResyncDSL) StaticRoute(staticRoute *l3.Route) linuxclient.DataResyncDSL {
	dsl.vppDataResync.StaticRoute(staticRoute)
	return dsl
}

// Arp adds VPP L3 ARP to the RESYNC request.
func (dsl *DataResyncDSL) Arp(arp *l3.ARPEntry) linuxclient.DataResyncDSL {
	dsl.vppDataResync.Arp(arp)
	return dsl
}

// ProxyArp adds L3 proxy ARP to the RESYNC request.
func (dsl *DataResyncDSL) ProxyArp(proxyArp *l3.ProxyARP) linuxclient.DataResyncDSL {
	dsl.vppDataResync.ProxyArp(proxyArp)
	return dsl
}

// IPScanNeighbor adds L3 IP Scan Neighbor to the RESYNC request.
func (dsl *DataResyncDSL) IPScanNeighbor(ipScanNeigh *l3.IPScanNeighbor) linuxclient.DataResyncDSL {
	dsl.vppDataResync.IPScanNeighbor(ipScanNeigh)

	return dsl
}

/*// L4Features adds L4 features to the RESYNC request
func (dsl *DataResyncDSL) L4Features(val *l4.L4Features) linuxclient.DataResyncDSL {
	dsl.vppDataResync.L4Features(val)
	return dsl
}

// AppNamespace adds VPP Application namespaces to the RESYNC request
func (dsl *DataResyncDSL) AppNamespace(appNs *l4.AppNamespaces_AppNamespace) linuxclient.DataResyncDSL {
	dsl.vppDataResync.AppNamespace(appNs)
	return dsl
}*/

// StnRule adds Stn rule to the RESYNC request.
func (dsl *DataResyncDSL) StnRule(stn *stn.Rule) linuxclient.DataResyncDSL {
	dsl.vppDataResync.StnRule(stn)
	return dsl
}

// NAT44Global adds global NAT44 configuration to the RESYNC request.
func (dsl *DataResyncDSL) NAT44Global(nat44 *nat.Nat44Global) linuxclient.DataResyncDSL {
	dsl.vppDataResync.NAT44Global(nat44)
	return dsl
}

// DNAT44 adds DNAT44 configuration to the RESYNC request
func (dsl *DataResyncDSL) DNAT44(nat44 *nat.DNat44) linuxclient.DataResyncDSL {
	dsl.vppDataResync.DNAT44(nat44)
	return dsl
}

// NAT44Interface adds NAT44 interface configuration to the RESYNC request.
func (dsl *DataResyncDSL) NAT44Interface(natIf *nat.Nat44Interface) linuxclient.DataResyncDSL {
	key := models.Key(natIf)
	dsl.txn.Put(key, natIf)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// NAT44AddressPool adds NAT44 address pool configuration to the RESYNC request.
func (dsl *DataResyncDSL) NAT44AddressPool(pool *nat.Nat44AddressPool) linuxclient.DataResyncDSL {
	key := models.Key(pool)
	dsl.txn.Put(key, pool)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// IPSecSA adds request to RESYNC a new Security Association
func (dsl *DataResyncDSL) IPSecSA(sa *ipsec.SecurityAssociation) linuxclient.DataResyncDSL {
	dsl.vppDataResync.IPSecSA(sa)
	return dsl
}

// IPSecSPD adds request to RESYNC a new Security Policy Database
func (dsl *DataResyncDSL) IPSecSPD(spd *ipsec.SecurityPolicyDatabase) linuxclient.DataResyncDSL {
	dsl.vppDataResync.IPSecSPD(spd)
	return dsl
}

// IPSecTunnelProtection adds request to RESYNC an IPSec tunnel protection
func (dsl *DataResyncDSL) IPSecTunnelProtection(tp *ipsec.TunnelProtection) linuxclient.DataResyncDSL {
	dsl.vppDataResync.IPSecTunnelProtection(tp)
	return dsl
}

// PuntIPRedirect adds request to RESYNC a rule used to punt L3 traffic via interface.
func (dsl *DataResyncDSL) PuntIPRedirect(val *punt.IPRedirect) linuxclient.DataResyncDSL {
	dsl.vppDataResync.PuntIPRedirect(val)
	return dsl
}

// PuntToHost adds request to RESYNC a rule used to punt L4 traffic to a host.
func (dsl *DataResyncDSL) PuntToHost(val *punt.ToHost) linuxclient.DataResyncDSL {
	dsl.vppDataResync.PuntToHost(val)
	return dsl
}

// PuntException adds request to create or update exception to punt specific packets.
func (dsl *DataResyncDSL) PuntException(val *punt.Exception) linuxclient.DataResyncDSL {
	key := models.Key(val)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// AppendKeys is a helper function that fills the keySet <keys> with values
// pointed to by the iterator <it>.
func appendKeys(keys *keySet, it keyval.ProtoKeyIterator) {
	for {
		k, _, stop := it.GetNext()
		if stop {
			break
		}

		(*keys)[k] = nil
	}
}

// KeySet is a helper type that reuses map keys to store values as a set.
// The values of the map are nil.
type keySet map[string] /*key*/ interface{} /*nil*/

// Send propagates the request to the plugins.
// It deletes obsolete keys if listKeys() (from constructor) function is not nil.
func (dsl *DataResyncDSL) Send() vppclient.Reply {

	for dsl.listKeys != nil {
		toBeDeleted := keySet{}

		// fill all known keys associated with the Linux network configuration:
		keys, err := dsl.listKeys(interfaces.ModelInterface.KeyPrefix())
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)

		// remove keys that are part of the transaction
		for _, txnKey := range dsl.txnKeys {
			delete(toBeDeleted, txnKey)
		}

		for delKey := range toBeDeleted {
			dsl.txn.Delete(delKey)
		}

		break
	}

	return dsl.vppDataResync.Send()
}
