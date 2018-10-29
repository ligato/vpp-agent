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

package dbadapter

import (
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/clientv2/vpp"
	"github.com/ligato/vpp-agent/plugins/vpp/model/bfd"
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l4"
	"github.com/ligato/vpp-agent/plugins/vpp/model/stn"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/acl"
	intf "github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/l2"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/l3"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/nat"
)

// NewDataResyncDSL returns a new instance of DataResyncDSL which implements
// the data RESYNC DSL for VPP configuration.
// Transaction <txn> is used to propagate changes to plugins.
// Function <listKeys> is used to list keys with already existing configuration.
func NewDataResyncDSL(txn keyval.ProtoTxn, listKeys func(prefix string) (keyval.ProtoKeyIterator, error)) *DataResyncDSL {
	return &DataResyncDSL{txn, []string{}, listKeys}
}

// DataResyncDSL is an implementation of Domain Specific Language (DSL) for data
// RESYNC of VPP configuration.
type DataResyncDSL struct {
	txn      keyval.ProtoTxn
	txnKeys  []string
	listKeys func(prefix string) (keyval.ProtoKeyIterator, error)
}

// Interface adds VPP interface to the RESYNC request.
func (dsl *DataResyncDSL) Interface(val *intf.Interface) vppclient.DataResyncDSL {
	key := intf.InterfaceKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// ACL adds Access Control List to the RESYNC request.
func (dsl *DataResyncDSL) ACL(val *acl.Acl) vppclient.DataResyncDSL {
	key := acl.Key(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BfdSession adds bidirectional forwarding detection session to the RESYNC
// request.
func (dsl *DataResyncDSL) BfdSession(val *bfd.SingleHopBFD_Session) vppclient.DataResyncDSL {
	key := bfd.SessionKey(val.Interface)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BfdAuthKeys adds bidirectional forwarding detection key to the RESYNC
// request.
func (dsl *DataResyncDSL) BfdAuthKeys(val *bfd.SingleHopBFD_Key) vppclient.DataResyncDSL {
	key := bfd.AuthKeysKey(string(val.Id))
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BfdEchoFunction adds bidirectional forwarding detection echo function
// to the RESYNC request.
func (dsl *DataResyncDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) vppclient.DataResyncDSL {
	key := bfd.EchoFunctionKey(val.EchoSourceInterface)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BD adds Bridge Domain to the RESYNC request.
func (dsl *DataResyncDSL) BD(val *l2.BridgeDomain) vppclient.DataResyncDSL {
	key := l2.BridgeDomainKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BDFIB adds Bridge Domain to the RESYNC request.
func (dsl *DataResyncDSL) BDFIB(val *l2.FIBEntry) vppclient.DataResyncDSL {
	key := l2.FIBKey(val.BridgeDomain, val.PhysAddress)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// XConnect adds Cross Connect to the RESYNC request.
func (dsl *DataResyncDSL) XConnect(val *l2.XConnectPair) vppclient.DataResyncDSL {
	key := l2.XConnectKey(val.ReceiveInterface)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// StaticRoute adds L3 Static Route to the RESYNC request.
func (dsl *DataResyncDSL) StaticRoute(val *l3.StaticRoute) vppclient.DataResyncDSL {
	key := l3.RouteKey(val.VrfId, val.DstNetwork, val.NextHopAddr)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// ProxyArp adds L3 proxy ARP to the RESYNC request.
func (dsl *DataResyncDSL) ProxyArp(proxyArp *l3.ProxyARP) vppclient.DataResyncDSL {
	key := l3.ProxyARPKey
	dsl.txn.Put(key, proxyArp)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// Arp adds L3 ARP entry to the RESYNC request.
func (dsl *DataResyncDSL) Arp(val *l3.ARPEntry) vppclient.DataResyncDSL {
	key := l3.ArpEntryKey(val.Interface, val.IpAddress)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// IPScanNeighbor adds L3 IP Scan Neighbor to the RESYNC request.
func (dsl *DataResyncDSL) IPScanNeighbor(ipScanNeigh *l3.IPScanNeighbor) vppclient.DataResyncDSL {
	key := l3.IPScanNeighborKey
	dsl.txn.Put(key, ipScanNeigh)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// L4Features adds L4Features to the RESYNC request
func (dsl *DataResyncDSL) L4Features(val *l4.L4Features) vppclient.DataResyncDSL {
	key := l4.FeatureKey()
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// AppNamespace adds Application Namespace to the RESYNC request
func (dsl *DataResyncDSL) AppNamespace(val *l4.AppNamespaces_AppNamespace) vppclient.DataResyncDSL {
	key := l4.AppNamespacesKey(val.NamespaceId)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// StnRule adds Stn rule to the RESYNC request.
func (dsl *DataResyncDSL) StnRule(val *stn.STN_Rule) vppclient.DataResyncDSL {
	key := stn.Key(val.RuleName)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// NAT44Global adds global NAT44 configuration to the RESYNC request.
func (dsl *DataResyncDSL) NAT44Global(nat44 *nat.Nat44Global) vppclient.DataResyncDSL {
	key := nat.GlobalNAT44Key
	dsl.txn.Put(key, nat44)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// DNAT44 adds DNAT44 configuration to the RESYNC request
func (dsl *DataResyncDSL) DNAT44(nat44 *nat.DNat44) vppclient.DataResyncDSL {
	key := nat.DNAT44Key(nat44.Label)
	dsl.txn.Put(key, nat44)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// IPSecSA adds request to create a new Security Association
func (dsl *DataResyncDSL) IPSecSA(sa *ipsec.SecurityAssociations_SA) vppclient.DataResyncDSL {
	key := ipsec.SAKey(sa.Name)
	dsl.txn.Put(key, sa)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// IPSecSPD adds request to create a new Security Policy Database
func (dsl *DataResyncDSL) IPSecSPD(spd *ipsec.SecurityPolicyDatabases_SPD) vppclient.DataResyncDSL {
	key := ipsec.SPDKey(spd.Name)
	dsl.txn.Put(key, spd)
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

		// fill all known keys of one VPP:

		keys, err := dsl.listKeys(intf.Prefix)
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l2.BDPrefix)
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l2.XConnectPrefix)
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l3.RoutePrefix)
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l3.ArpPrefix)
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

	err := dsl.txn.Commit()

	return &Reply{err: err}
}
