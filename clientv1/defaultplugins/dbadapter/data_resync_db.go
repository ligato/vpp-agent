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
	"net"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
)

// NewDataResyncDSL is a constructor
func NewDataResyncDSL(txn keyval.ProtoTxn, listKeys func(prefix string) (keyval.ProtoKeyIterator, error)) *DataResyncDSL {
	return &DataResyncDSL{txn, []string{}, listKeys}
}

// DataResyncDSL is used to conveniently assign all the data that are needed for the RESYNC
// This is implementation of Domain Specific Language (DSL) for data RESYNC of the VPP configuration.
type DataResyncDSL struct {
	txn      keyval.ProtoTxn
	txnKeys  []string
	listKeys func(prefix string) (keyval.ProtoKeyIterator, error)
}

// Interface add Bridge Domain to the RESYNC request
func (dsl *DataResyncDSL) Interface(val *interfaces.Interfaces_Interface) defaultplugins.DataResyncDSL {
	key := intf.InterfaceKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BfdSession BFD session to the RESYNC request
func (dsl *DataResyncDSL) BfdSession(val *bfd.SingleHopBFD_Session) defaultplugins.DataResyncDSL {
	key := bfd.SessionKey(val.Interface)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BfdKeys BFD key to the RESYNC request
func (dsl *DataResyncDSL) BfdKeys(val *bfd.SingleHopBFD_Key) defaultplugins.DataResyncDSL {
	key := bfd.AuthKeysKey(string(val.Id))
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BfdEchoFunction BFD echo function to the RESYNC request
func (dsl *DataResyncDSL) BfdEchoFunction(val *bfd.SingleHopBFD_EchoFunction) defaultplugins.DataResyncDSL {
	key := bfd.EchoFunctionKey(val.EchoSourceInterface)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BD add Bridge Domain to the RESYNC request
func (dsl *DataResyncDSL) BD(val *l2.BridgeDomains_BridgeDomain) defaultplugins.DataResyncDSL {
	key := l2.BridgeDomainKey(val.Name)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// BDFIB add Bridge Domain to the RESYNC request
func (dsl *DataResyncDSL) BDFIB(val *l2.FibTableEntries_FibTableEntry) defaultplugins.DataResyncDSL {
	key := l2.FibKey(val.BridgeDomain, val.PhysAddress)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// XConnect adds Cross Connect to the RESYNC request
func (dsl *DataResyncDSL) XConnect(val *l2.XConnectPairs_XConnectPair) defaultplugins.DataResyncDSL {
	key := l2.XConnectKey(val.ReceiveInterface)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// StaticRoute adss L3 Static Route to the RESYNC request
func (dsl *DataResyncDSL) StaticRoute(val *l3.StaticRoutes_Route) defaultplugins.DataResyncDSL {
	_, dstAddr, _ := net.ParseCIDR(val.DstIpAddr)
	key := l3.RouteKey(val.VrfId, dstAddr, val.NextHopAddr)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// ACL adds Access Control List to the RESYNC request
func (dsl *DataResyncDSL) ACL(val *acl.AccessLists_Acl) defaultplugins.DataResyncDSL {
	key := acl.Key(val.AclName)
	dsl.txn.Put(key, val)
	dsl.txnKeys = append(dsl.txnKeys, key)

	return dsl
}

// AppendKeys is a helper function that fils the keySet with iterator values
func appendKeys(keys *keySet, it keyval.ProtoKeyIterator) {
	for {
		k, _, stop := it.GetNext()
		if stop {
			break
		}

		(*keys)[k] = nil
	}
}

// KeySet is a helper type that reuse the map keys to store the set vales. The values of the map are nil.
type keySet map[string] /*key*/ interface{} /*nil*/

// Send propagates the request to the plugins. It deletes obsolete keys if listKeys function is not null.
// The listkeys() function is used to list all current keys.
func (dsl *DataResyncDSL) Send() defaultplugins.Reply {

	for dsl.listKeys != nil {
		toBeDeleted := keySet{}

		// fill all known keys of one VPP:

		keys, err := dsl.listKeys(intf.InterfaceKeyPrefix())
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l2.BridgeDomainKeyPrefix())
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l2.XConnectKeyPrefix())
		if err != nil {
			break
		}
		appendKeys(&toBeDeleted, keys)
		keys, err = dsl.listKeys(l3.RouteKeyPrefix())
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
