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
	"github.com/ligato/vpp-agent/clientv1/linux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/model/interfaces"

	vpp_clientv1 "github.com/ligato/vpp-agent/clientv1/defaultplugins"
	vpp_dbadapter "github.com/ligato/vpp-agent/clientv1/defaultplugins/dbadapter"
	vpp_acl "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	vpp_intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	vpp_l3 "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	vpp_bfd "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
)

// NewDataResyncDSL returns a new instance of DataResyncDSL which implements
// the data RESYNC DSL for both Linux and VPP config (inherits dbadapter
// from defaultplugins)
// Transaction <txn> is used to propagate changes to plugins.
// Function <listKeys> is used to list keys with already existing configuration.
func NewDataResyncDSL(txn keyval.ProtoTxn, listKeys func(prefix string) (keyval.ProtoKeyIterator, error)) *DataResyncDSL {
	vppDataResync := vpp_dbadapter.NewDataResyncDSL(txn, listKeys)
	return &DataResyncDSL{txn, []string{}, listKeys, vppDataResync}
}

// DataResyncDSL is an implementation of Domain Specific Language (DSL) for data
// RESYNC of both Linux and VPP configuration.
type DataResyncDSL struct {
	txn      keyval.ProtoTxn
	txnKeys  []string
	listKeys func(prefix string) (keyval.ProtoKeyIterator, error)

	vppDataResync vpp_clientv1.DataResyncDSL
}

// LinuxInterface adds Linux interface to the RESYNC request.
func (dsl *DataResyncDSL) LinuxInterface(val *interfaces.LinuxInterfaces_Interface) linux.DataResyncDSL {
	key := interfaces.InterfaceKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// LinuxArpEntry adds Linux ARP entry to the RESYNC request.
func (dsl *DataResyncDSL) LinuxArpEntry(val *l3.LinuxStaticArpEntries_ArpEntry) linux.DataResyncDSL {
	key := l3.StaticArpKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// LinuxRoute adds Linux route to the RESYNC request.
func (dsl *DataResyncDSL) LinuxRoute(val *l3.LinuxStaticRoutes_Route) linux.DataResyncDSL {
	key := l3.StaticRouteKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// VppInterface adds VPP interface to the RESYNC request.
func (dsl *DataResyncDSL) VppInterface(intf *vpp_intf.Interfaces_Interface) linux.DataResyncDSL {
	dsl.vppDataResync.Interface(intf)
	return dsl
}

// BfdSession adds VPP bidirectional forwarding detection session
// to the RESYNC request.
func (dsl *DataResyncDSL) BfdSession(val *vpp_bfd.SingleHopBFD_Session) linux.DataResyncDSL {
	dsl.vppDataResync.BfdSession(val)
	return dsl
}

// BfdAuthKeys adds VPP bidirectional forwarding detection key to the RESYNC
// request.
func (dsl *DataResyncDSL) BfdAuthKeys(val *vpp_bfd.SingleHopBFD_Key) linux.DataResyncDSL {
	dsl.vppDataResync.BfdAuthKeys(val)
	return dsl
}

// BfdEchoFunction adds VPP bidirectional forwarding detection echo function
// to the RESYNC request.
func (dsl *DataResyncDSL) BfdEchoFunction(val *vpp_bfd.SingleHopBFD_EchoFunction) linux.DataResyncDSL {
	dsl.vppDataResync.BfdEchoFunction(val)
	return dsl
}

// BD adds VPP Bridge Domain to the RESYNC request.
func (dsl *DataResyncDSL) BD(bd *vpp_l2.BridgeDomains_BridgeDomain) linux.DataResyncDSL {
	dsl.vppDataResync.BD(bd)
	return dsl
}

// BDFIB adds VPP L2 FIB to the RESYNC request.
func (dsl *DataResyncDSL) BDFIB(fib *vpp_l2.FibTableEntries_FibTableEntry) linux.DataResyncDSL {
	dsl.vppDataResync.BDFIB(fib)
	return dsl
}

// XConnect adds VPP Cross Connect to the RESYNC request.
func (dsl *DataResyncDSL) XConnect(xcon *vpp_l2.XConnectPairs_XConnectPair) linux.DataResyncDSL {
	dsl.vppDataResync.XConnect(xcon)
	return dsl
}

// StaticRoute adds VPP L3 Static Route to the RESYNC request.
func (dsl *DataResyncDSL) StaticRoute(staticRoute *vpp_l3.StaticRoutes_Route) linux.DataResyncDSL {
	dsl.vppDataResync.StaticRoute(staticRoute)
	return dsl
}

// ACL adds VPP Access Control List to the RESYNC request.
func (dsl *DataResyncDSL) ACL(acl *vpp_acl.AccessLists_Acl) linux.DataResyncDSL {
	dsl.vppDataResync.ACL(acl)
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

// KeySet is a helper type that reuses map keys to store vales as a set.
// The values of the map are nil.
type keySet map[string] /*key*/ interface{} /*nil*/

// Send propagates the request to the plugins.
// It deletes obsolete keys if listKeys() (from constructor) function is not nil.
func (dsl *DataResyncDSL) Send() vpp_clientv1.Reply {

	for dsl.listKeys != nil {
		toBeDeleted := keySet{}

		// fill all known keys associated with the Linux network configuration:
		keys, err := dsl.listKeys(interfaces.InterfaceKeyPrefix())
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
