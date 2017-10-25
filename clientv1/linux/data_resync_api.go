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

package linux

import (
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/model/interfaces"

	vpp_clientv1 "github.com/ligato/vpp-agent/clientv1/defaultplugins"
	vpp_acl "github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	vpp_intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	vpp_l2 "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	vpp_l3 "github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	vpp_bfd "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
)

// DataResyncDSL defines the Domain Specific Language (DSL) for data RESYNC
// of both Linux and VPP configuration.
// Use this interface to make your implementation independent of the local
// and any remote client.
// Each method (apart from Send) returns the receiver, allowing the calls
// to be chained together conveniently in a single statement.
type DataResyncDSL interface {
	// LinuxInterface adds Linux interface to the RESYNC request.
	LinuxInterface(intf *interfaces.LinuxInterfaces_Interface) DataResyncDSL
	// LinuxInterface adds Linux ARP entry to the RESYNC request.
	LinuxArpEntry(intf *l3.LinuxStaticArpEntries_ArpEntry) DataResyncDSL
	// LinuxInterface adds Linux route to the RESYNC request.
	LinuxRoute(intf *l3.LinuxStaticRoutes_Route) DataResyncDSL

	// VppInterface adds VPP interface to the RESYNC request.
	VppInterface(intf *vpp_intf.Interfaces_Interface) DataResyncDSL
	// BfdSession adds VPP bidirectional forwarding detection session
	// to the RESYNC request.
	BfdSession(val *vpp_bfd.SingleHopBFD_Session) DataResyncDSL
	// BfdAuthKeys adds VPP bidirectional forwarding detection key to the RESYNC
	// request.
	BfdAuthKeys(val *vpp_bfd.SingleHopBFD_Key) DataResyncDSL
	// BfdEchoFunction adds VPP bidirectional forwarding detection echo function
	// to the RESYNC request.
	BfdEchoFunction(val *vpp_bfd.SingleHopBFD_EchoFunction) DataResyncDSL
	// BD adds VPP Bridge Domain to the RESYNC request.
	BD(bd *vpp_l2.BridgeDomains_BridgeDomain) DataResyncDSL
	// BDFIB adds VPP L2 FIB to the RESYNC request.
	BDFIB(fib *vpp_l2.FibTableEntries_FibTableEntry) DataResyncDSL
	// XConnect adds VPP Cross Connect to the RESYNC request.
	XConnect(xcon *vpp_l2.XConnectPairs_XConnectPair) DataResyncDSL
	// StaticRoute adds VPP L3 Static Route to the RESYNC request.
	StaticRoute(staticRoute *vpp_l3.StaticRoutes_Route) DataResyncDSL
	// ACL adds VPP Access Control List to the RESYNC request.
	ACL(acl *vpp_acl.AccessLists_Acl) DataResyncDSL

	// Send propagates the RESYNC request to the plugins.
	Send() vpp_clientv1.Reply
}
