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

package vppclient

import (
	linuxIf "github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
	linuxL3 "github.com/ligato/vpp-agent/plugins/linux/model/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
)

// DataGetDSL defines Domain Specific Language (DSL) for data read.
// of the VPP configuration.
// Use this interface to make your implementation independent of the local
// and any remote client.
// Every DSL statement (apart from Send) returns the receiver (possibly wrapped
// to change the scope of DSL), allowing the calls to be chained together
// conveniently in a single statement.
type DataGetDSL interface {
	// Get initiates a chained sequence of data read DSL statements, reading
	// existing configurable objects, e.g.:
	//     Get().Interfaces().BD() ... Send()
	// The set of available objects to be created or changed is defined by GetDSL.
	Get() GetDSL
}

// GetDSL is a subset of data read DSL statements, used to read existing
// VPP configuration.
type GetDSL interface {
	// ACLs adds a request to read VPP access lists.
	ACLs() GetDSL
	// Interfaces adds a request to read VPP interfaces.
	Interfaces() GetDSL
	// IPSecSPDs adds a request to read IPSec SPDs.
	IPSecSPDs() GetDSL
	// IPSecSAs adds a request to read IPSec SAs.
	IPSecSAs() GetDSL
	// IPSecTunnels adds a request to read IPSec tunnels.
	IPSecTunnels() GetDSL
	// BDs adds a request to read bridge domains.
	BDs() GetDSL
	// FIBs adds a request to read FIBs.
	FIBs() GetDSL
	// XConnects adds a request to read cross connects.
	XConnects() GetDSL
	// Routes adds a request to read routes.
	Routes() GetDSL
	// ARPs adds a request to read ARPs.
	ARPs() GetDSL
	// LinuxInterfaces adds a request to read linux interfaces.
	LinuxInterfaces() GetDSL
	// LinuxARPs adds a request to read linux ARPs.
	LinuxARPs() GetDSL
	// LinuxRoutes adds a request to read linux routes.
	LinuxRoutes() GetDSL

	// Send propagates requested changes to the plugins.
	Send() GetReply
}

// GetReply interface allows to wait for a reply to previously called Send() and
// extract the result from it (success/error).
type GetReply interface {
	// ReceiveReply waits for a reply to previously called Send() and returns
	// the result (data set or error).
	ReceiveReply() (ReplyData, error)
}

// ReplyData is helper interface for more convenient access to typed data
type ReplyData interface {
	// GetACLs returns all access lists from the reply
	GetACLs() []*acl.AccessLists_Acl
	// GetInterfaces returns all the interfaces from the reply
	GetInterfaces() []*interfaces.Interfaces_Interface
	// GetIPSecSPDs returns all the IPSec SPDs from the reply
	GetIPSecSPDs() []*ipsec.SecurityPolicyDatabases_SPD
	// GetIPSecSAs returns all the IPSec SAa from the reply
	GetIPSecSAs() []*ipsec.SecurityAssociations_SA
	// GetBDs returns all the bridge domains from the reply
	GetBDs() []*l2.BridgeDomains_BridgeDomain
	// GetFIBs returns all the FIB entries from the reply
	GetFIBs() []*l2.FibTable_FibEntry
	// GetXConnects returns all the XConnects from the reply
	GetXConnects() []*l2.XConnectPairs_XConnectPair
	// GetARPs returns all the ARPs from the reply
	GetARPs() []*l3.ArpTable_ArpEntry
	// GetRoutes returns all the routes from the reply
	GetRoutes() []*l3.StaticRoutes_Route
	// GetLinuxInterfaces returns all the linux interfaces from the reply
	GetLinuxInterfaces() []*linuxIf.LinuxInterfaces_Interface
	// GetLinuxARPs returns all the linux ARPs from the reply
	GetLinuxARPs() []*linuxL3.LinuxStaticArpEntries_ArpEntry
	// GetLinuxRoutes returns all the linux routes from the reply
	GetLinuxRoutes() []*linuxL3.LinuxStaticRoutes_Route
}
