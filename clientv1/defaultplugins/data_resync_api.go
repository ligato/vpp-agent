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

package defaultplugins

import (
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
)

// DataResyncDSL is used to conveniently assign all the data that are needed for the RESYNC.
// Use this interface to make you implementation independent of local or particular remote client.
// This interface defines the Domain Specific Language (DSL) for data RESYNC of the VPP configuration.
type DataResyncDSL interface {
	// Interface add interface to the RESYNC request
	Interface(intf *interfaces.Interfaces_Interface) DataResyncDSL
	// BD add Bridge Domain to the RESYNC request
	BD(bd *l2.BridgeDomains_BridgeDomain) DataResyncDSL
	// BDFIB add L2 Forwarding Information Base
	BDFIB(fib *l2.FibTableEntries_FibTableEntry) DataResyncDSL
	// XConnect adds Cross Connect to the RESYNC request
	XConnect(xcon *l2.XConnectPairs_XConnectPair) DataResyncDSL
	// StaticRoute adds L3 Static Route to the RESYNC request
	StaticRoute(staticRoute *l3.StaticRoutes_Route) DataResyncDSL
	// ACL adds Access Control List to the RESYNC request
	ACL(acl *acl.AccessLists_Acl) DataResyncDSL
	// Send propagates the request to the plugins
	Send() Reply
}
