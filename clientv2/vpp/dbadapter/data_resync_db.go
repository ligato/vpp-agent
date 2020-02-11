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
	"context"

	"go.ligato.io/cn-infra/v2/db/keyval"

	"go.ligato.io/vpp-agent/v3/pkg/models"

	vppclient "go.ligato.io/vpp-agent/v3/clientv2/vpp"
	orch "go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
	acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	intf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
	punt "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/punt"
	stn "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/stn"
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

// Span adds VPP span to the RESYNC request.
func (dsl *DataResyncDSL) Span(val *intf.Span) vppclient.DataResyncDSL {
	key := intf.SpanKey(val.InterfaceFrom, val.InterfaceTo)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// ACL adds Access Control List to the RESYNC request.
func (dsl *DataResyncDSL) ACL(val *acl.ACL) vppclient.DataResyncDSL {
	key := acl.Key(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// ABF adds ACL-based forwarding to the RESYNC request.
func (dsl *DataResyncDSL) ABF(val *abf.ABF) vppclient.DataResyncDSL {
	key := models.Key(val)
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

// VrfTable adds VRF table to the RESYNC request.
func (dsl *DataResyncDSL) VrfTable(val *l3.VrfTable) vppclient.DataResyncDSL {
	key := l3.VrfTableKey(val.Id, val.Protocol)
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
func (dsl *DataResyncDSL) StaticRoute(val *l3.Route) vppclient.DataResyncDSL {
	key := models.Key(val)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// ProxyArp adds L3 proxy ARP to the RESYNC request.
func (dsl *DataResyncDSL) ProxyArp(proxyArp *l3.ProxyARP) vppclient.DataResyncDSL {
	key := models.Key(&l3.ProxyARP{})
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
	key := models.Key(&l3.IPScanNeighbor{})
	dsl.txn.Put(key, ipScanNeigh)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// StnRule adds Stn rule to the RESYNC request.
func (dsl *DataResyncDSL) StnRule(val *stn.Rule) vppclient.DataResyncDSL {
	key := stn.Key(val.Interface, val.IpAddress)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// NAT44Global adds global NAT44 configuration to the RESYNC request.
func (dsl *DataResyncDSL) NAT44Global(nat44 *nat.Nat44Global) vppclient.DataResyncDSL {
	key := models.Key(&nat.Nat44Global{})
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

// NAT44Interface adds NAT44 interface configuration to the RESYNC request.
func (dsl *DataResyncDSL) NAT44Interface(natIf *nat.Nat44Interface) vppclient.DataResyncDSL {
	key := models.Key(natIf)
	dsl.txn.Put(key, natIf)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// NAT44AddressPool adds NAT44 address pool configuration to the RESYNC request.
func (dsl *DataResyncDSL) NAT44AddressPool(pool *nat.Nat44AddressPool) vppclient.DataResyncDSL {
	key := models.Key(pool)
	dsl.txn.Put(key, pool)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// IPSecSA adds request to create a new Security Association
func (dsl *DataResyncDSL) IPSecSA(sa *ipsec.SecurityAssociation) vppclient.DataResyncDSL {
	key := ipsec.SAKey(sa.Index)
	dsl.txn.Put(key, sa)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// IPSecSPD adds request to create a new Security Policy Database
func (dsl *DataResyncDSL) IPSecSPD(spd *ipsec.SecurityPolicyDatabase) vppclient.DataResyncDSL {
	key := ipsec.SPDKey(spd.Index)
	dsl.txn.Put(key, spd)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// PuntIPRedirect adds request to RESYNC a rule used to punt L3 traffic via interface.
func (dsl *DataResyncDSL) PuntIPRedirect(val *punt.IPRedirect) vppclient.DataResyncDSL {
	key := punt.IPRedirectKey(val.L3Protocol, val.TxInterface)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// PuntToHost adds request to RESYNC a rule used to punt L4 traffic to a host.
func (dsl *DataResyncDSL) PuntToHost(val *punt.ToHost) vppclient.DataResyncDSL {
	key := punt.ToHostKey(val.L3Protocol, val.L4Protocol, val.Port)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// PuntException adds request to create or update exception to punt specific packets.
func (dsl *DataResyncDSL) PuntException(val *punt.Exception) vppclient.DataResyncDSL {
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

		// fill all known keys of one VPP:

		keys, err := dsl.listKeys(intf.ModelInterface.KeyPrefix())
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l2.ModelBridgeDomain.KeyPrefix())
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l2.ModelXConnectPair.KeyPrefix())
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l3.ModelRoute.KeyPrefix())
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l3.ModelARPEntry.KeyPrefix())
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

	ctx := context.Background()
	ctx = orch.DataSrcContext(ctx, "localclient")
	err := dsl.txn.Commit(ctx)

	return &Reply{err: err}
}
