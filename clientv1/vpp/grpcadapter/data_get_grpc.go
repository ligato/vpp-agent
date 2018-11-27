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

package grpcadapter

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/clientv1/vpp"
	linuxIf "github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
	linuxL3 "github.com/ligato/vpp-agent/plugins/linux/model/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/model/rpc"
)

// NewDataGetDSL is a constructor
func NewDataGetDSL(client rpc.DataGetServiceClient) *DataGetDSL {
	return &DataGetDSL{
		client: client,
	}
}

// DataGetDSL is used to conveniently assign all the data that are needed for the Data read.
// This is an implementation of Domain Specific Language (DSL) for a change of the VPP configuration.
type DataGetDSL struct {
	client rpc.DataGetServiceClient
	get    []proto.Message
}

// GetDSL allows to read the configuration of default plugins based on grpc requests.
type GetDSL struct {
	parent *DataGetDSL
}

// Get enables reading Interface/BD...
func (dsl *DataGetDSL) Get() vppclient.GetDSL {
	return &GetDSL{dsl}
}

// ACLs adds a request to read an existing VPP access lists
func (dsl *GetDSL) ACLs() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &acl.AccessLists_Acl{})
	return dsl
}

// Interfaces adds a request to read an existing VPP interfaces
func (dsl *GetDSL) Interfaces() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &interfaces.Interfaces_Interface{})
	return dsl
}

// IPSecSPDs adds a request to read an existing IPSec SPDs
func (dsl *GetDSL) IPSecSPDs() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &ipsec.SecurityPolicyDatabases_SPD{})
	return dsl
}

// IPSecSAs adds a request to read an existing IPSec SAs
func (dsl *GetDSL) IPSecSAs() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &ipsec.SecurityAssociations_SA{})
	return dsl
}

// IPSecTunnels adds a request to read an existing IPSec tunnels
func (dsl *GetDSL) IPSecTunnels() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &ipsec.TunnelInterfaces_Tunnel{})
	return dsl
}

// BDs adds a request to read an existing bridge domains
func (dsl *GetDSL) BDs() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &l2.BridgeDomains_BridgeDomain{})
	return dsl
}

// FIBs adds a request to read an existing FIBs
func (dsl *GetDSL) FIBs() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &l2.FibTable_FibEntry{})
	return dsl
}

// XConnects adds a request to read an existing cross connects
func (dsl *GetDSL) XConnects() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &l2.XConnectPairs_XConnectPair{})
	return dsl
}

// Routes adds a request to read an existing VPP routes
func (dsl *GetDSL) Routes() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &l3.StaticRoutes_Route{})
	return dsl
}

// ARPs adds a request to read an existing VPP ARPs
func (dsl *GetDSL) ARPs() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &l3.ArpTable_ArpEntry{})
	return dsl
}

// LinuxInterfaces adds a request to read an existing linux interfaces
func (dsl *GetDSL) LinuxInterfaces() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &linuxIf.LinuxInterfaces_Interface{})
	return dsl
}

// LinuxARPs adds a request to read an existing linux ARPs
func (dsl *GetDSL) LinuxARPs() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &linuxL3.LinuxStaticArpEntries_ArpEntry{})
	return dsl
}

// LinuxRoutes adds a request to read an existing linux routes
func (dsl *GetDSL) LinuxRoutes() vppclient.GetDSL {
	dsl.parent.get = append(dsl.parent.get, &linuxL3.LinuxStaticRoutes_Route{})
	return dsl
}

// Send propagates request
func (dsl *GetDSL) Send() vppclient.GetReply {
	return dsl.parent.Send()
}

// Send propagates request
func (dsl *DataGetDSL) Send() vppclient.GetReply {
	ctx := context.Background()

	rd := &replyData{}

	for _, dataType := range dsl.get {
		request := &rpc.GetRequest{}

		switch dataType.(type) {
		case *acl.AccessLists_Acl:
			resp, err := dsl.client.GetAcls(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.acl = resp.AccessLists

		case *interfaces.Interfaces_Interface:
			resp, err := dsl.client.GetInterfaces(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.ifs = resp.Interfaces
		case *ipsec.SecurityPolicyDatabases_SPD:
			resp, err := dsl.client.GetIPSecSPDs(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.spds = resp.SPDs
		case *ipsec.SecurityAssociations_SA:
			resp, err := dsl.client.GetIPSecSAs(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.sas = resp.SAa
		case *ipsec.TunnelInterfaces_Tunnel:
			resp, err := dsl.client.GetIPSecTunnels(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.tuns = resp.Tunnels
		case *l2.BridgeDomains_BridgeDomain:
			resp, err := dsl.client.GetBDs(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.bds = resp.BridgeDomains
		case *l2.FibTable_FibEntry:
			resp, err := dsl.client.GetFIBs(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.fibs = resp.FIBs
		case *l2.XConnectPairs_XConnectPair:
			resp, err := dsl.client.GetXConnects(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.xcs = resp.XCons
		case *l3.ArpTable_ArpEntry:
			resp, err := dsl.client.GetARPs(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.arps = resp.ArpEntries
		case *l3.StaticRoutes_Route:
			resp, err := dsl.client.GetRoutes(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.routes = resp.StaticRoutes
		case *linuxIf.LinuxInterfaces_Interface:
			resp, err := dsl.client.GetLinuxInterfaces(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.linuxIfs = resp.LinuxInterfaces
		case *linuxL3.LinuxStaticArpEntries:
			resp, err := dsl.client.GetLinuxARPs(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.linuxArps = resp.LinuxArpEntries
		case *linuxL3.LinuxStaticRoutes_Route:
			resp, err := dsl.client.GetLinuxRoutes(ctx, request)
			if err != nil {
				return &GetReply{&replyData{err: err}}
			}
			rd.linuxRoutes = resp.LinuxRoutes
		}
	}

	return &GetReply{rd}
}

// GetReply enables waiting for the reply and getting result (data list/error).
type GetReply struct {
	rd *replyData
}

// replyData is helper struct implementing ReplyData interface and allows to read typed data from the reply
type replyData struct {
	err error

	acl         []*acl.AccessLists_Acl
	ifs         []*interfaces.Interfaces_Interface
	spds        []*ipsec.SecurityPolicyDatabases_SPD
	sas         []*ipsec.SecurityAssociations_SA
	tuns        []*ipsec.TunnelInterfaces_Tunnel
	bds         []*l2.BridgeDomains_BridgeDomain
	fibs        []*l2.FibTable_FibEntry
	xcs         []*l2.XConnectPairs_XConnectPair
	routes      []*l3.StaticRoutes_Route
	arps        []*l3.ArpTable_ArpEntry
	linuxIfs    []*linuxIf.LinuxInterfaces_Interface
	linuxArps   []*linuxL3.LinuxStaticArpEntries_ArpEntry
	linuxRoutes []*linuxL3.LinuxStaticRoutes_Route
}

// ReceiveReply returns all the data and error
func (reply *GetReply) ReceiveReply() (vppclient.ReplyData, error) {
	return reply.rd, reply.rd.err
}

// GetACLs returns all access lists from the reply
func (rd *replyData) GetACLs() []*acl.AccessLists_Acl {
	return rd.acl
}

// GetInterfaces returns all the interfaces from the reply
func (rd *replyData) GetInterfaces() []*interfaces.Interfaces_Interface {
	return rd.ifs
}

// GetIPSecSPDs returns all the IPSec SPDs from the reply
func (rd *replyData) GetIPSecSPDs() []*ipsec.SecurityPolicyDatabases_SPD {
	return rd.spds
}

// GetIPSecSAs returns all the IPSec SAa from the reply
func (rd *replyData) GetIPSecSAs() []*ipsec.SecurityAssociations_SA {
	return rd.sas
}

// GetBDs returns all the bridge domains from the reply
func (rd *replyData) GetBDs() []*l2.BridgeDomains_BridgeDomain {
	return rd.bds
}

// GetFIBs returns all the FIB entries from the reply
func (rd *replyData) GetFIBs() []*l2.FibTable_FibEntry {
	return rd.fibs
}

// GetXConnects returns all the XConnects from the reply
func (rd *replyData) GetXConnects() []*l2.XConnectPairs_XConnectPair {
	return rd.xcs
}

// GetARPs returns all the ARPs from the reply
func (rd *replyData) GetARPs() []*l3.ArpTable_ArpEntry {
	return rd.arps
}

// GetRoutes returns all the routes from the reply
func (rd *replyData) GetRoutes() []*l3.StaticRoutes_Route {
	return rd.routes
}

// GetLinuxInterfaces returns all the linux interfaces from the reply
func (rd *replyData) GetLinuxInterfaces() []*linuxIf.LinuxInterfaces_Interface {
	return rd.linuxIfs
}

// GetLinuxARPs returns all the linux ARPs from the reply
func (rd *replyData) GetLinuxARPs() []*linuxL3.LinuxStaticArpEntries_ArpEntry {
	return rd.linuxArps
}

// GetLinuxRoutes returns all the linux routes from the reply
func (rd *replyData) GetLinuxRoutes() []*linuxL3.LinuxStaticRoutes_Route {
	return rd.linuxRoutes
}
