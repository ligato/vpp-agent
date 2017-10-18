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

package rpc

import (
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"golang.org/x/net/context"
	"github.com/ligato/vpp-agent/flavors/rpc/model/vppsvc"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"net"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/flavors/local"
)

// GRPCSvcPlugin registers VPP GRPC services in *grpc.Server
type GRPCSvcPlugin struct {
	Deps         GRPCSvcPluginDeps
	changeVppSvc ChangeVppSvc
	resyncVppSvc ResyncVppSvc
}

// GRPCSvcPluginDeps - dependencies of GRPCSvcPlugin
type GRPCSvcPluginDeps struct {
	local.PluginLogDeps
	GRPC grpc.Server
}

// Init sets plugin child loggers for changeVppSvc & resyncVppSvc
func (plugin *GRPCSvcPlugin) Init() error {
	plugin.changeVppSvc.Log = plugin.Deps.Log.NewLogger("changeVppSvc")
	plugin.resyncVppSvc.Log = plugin.Deps.Log.NewLogger("resyncVppSvc")

	return nil
}

// AfterInit registers all GRPC servics in vppscv package
// (be sure that defaultvppplugins are totally initialized)
func (plugin *GRPCSvcPlugin) AfterInit() error {
	grpcServer := plugin.Deps.GRPC.Server()
	vppsvc.RegisterChangeConfigServiceServer(grpcServer, &plugin.changeVppSvc)
	vppsvc.RegisterResyncConfigServiceServer(grpcServer, &plugin.resyncVppSvc)

	return nil
}

// Close does nothing
func (plugin *GRPCSvcPlugin) Close() error {
	return nil
}

// ChangeVppSvc forward GRPC request to localclient
type ChangeVppSvc struct {
	Log logging.Logger
}

// ResyncVppSvc forward GRPC request to localclient
type ResyncVppSvc struct {
	Log logging.Logger
}

// PutInterfaces creates or updates one or multiple interfaces (forwards the input to localclient)
func (svc *ChangeVppSvc) PutInterfaces(ctx context.Context, request *interfaces.Interfaces) (
	*vppsvc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	for _, intf := range request.Interface {
		localReqPut.VppInterface(intf)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.PutResponse{}, err
}

// DelInterfaces one or multiple interfaces by their unique names 
// (forwards the input to localclient)
func (svc *ChangeVppSvc) DelInterfaces(ctx context.Context, request *vppsvc.DelNamesRequest) (*vppsvc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	for _, intfName := range request.Name {
		localReqDel.VppInterface(intfName)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.DelResponse{}, err
}

// PutBDs creates or updates one or multiple BDs (forwards the input to localclient)
// (forwards the input to localclient)
func (svc *ChangeVppSvc) PutBDs(ctx context.Context, request *l2.BridgeDomains) (
	*vppsvc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	for _, bd := range request.BridgeDomains {
		localReqPut.BD(bd)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.PutResponse{}, err
}

// DelBDs one or multiple BDs by their unique names (forwards the input to localclient)
// (forwards the input to localclient)
func (svc *ChangeVppSvc) DelBDs(ctx context.Context, request *vppsvc.DelNamesRequest) (*vppsvc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	for _, bdName := range request.Name {
		localReqDel.BD(bdName)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.DelResponse{}, err
}

// PutXCons creates or updates one or multiple Cross Connects (forwards the input to localclient)
// (forwards the input to localclient)
func (svc *ChangeVppSvc) PutXCons(ctx context.Context, request *l2.XConnectPairs) (
	*vppsvc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	for _, xcon := range request.XConnectPairs {
		localReqPut.XConnect(xcon)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.PutResponse{}, err
}

// DelXCons one or multiple Cross Connects by their unique names 
// (forwards the input to localclient)
func (svc *ChangeVppSvc) DelXCons(ctx context.Context, request *vppsvc.DelNamesRequest) (*vppsvc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	for _, rxIfaceName := range request.Name {
		localReqDel.XConnect(rxIfaceName)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.DelResponse{}, err
}

// PutACLs creates or updates one or multiple ACLs
// (forwards the input to localclient)
func (svc *ChangeVppSvc) PutACLs(ctx context.Context, request *acl.AccessLists) (
	*vppsvc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	for _, acl := range request.Acl {
		localReqPut.ACL(acl)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.PutResponse{}, err
}

// DelACLs one or multiple ACLs by their unique names
// (forwards the input to localclient)
func (svc *ChangeVppSvc) DelACLs(ctx context.Context, request *vppsvc.DelNamesRequest) (*vppsvc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	for _, aclName := range request.Name {
		localReqDel.ACL(aclName)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.DelResponse{}, err
}

// PutStaticRoutes creates or updates one or multiple ACLs
// (forwards the input to localclient)
func (svc *ChangeVppSvc) PutStaticRoutes(ctx context.Context, request *l3.StaticRoutes) (
	*vppsvc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	for _, route := range request.Route {
		localReqPut.StaticRoute(route)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.PutResponse{}, err
}

// DelStaticRoutes one or multiple ACLs by their unique names
// (forwards the input to localclient)
func (svc *ChangeVppSvc) DelStaticRoutes(ctx context.Context, request *vppsvc.DelStaticRoutesRequest) (*vppsvc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	for _, route := range request.Route {
		_, dst, err := net.ParseCIDR(route.DstAddr)
		if err != nil {
			localReqDel.StaticRoute(route.VRF, dst, net.ParseIP(route.NextHopAddr))
		} else {
			svc.Log.Error("error parsing static route ", route.DstAddr)
		}
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.DelResponse{}, err
}

// ResyncConfig full data resync request of defaultvppplugin configuration
// (forwards the input to localclient)
func (svc *ResyncVppSvc) ResyncConfig(ctx context.Context, request *vppsvc.ResyncConfigRequest) (
	*vppsvc.ResyncConfigResponse, error) {

	localReq := localclient.DataResyncRequest("vppsvc")
	for _, intf := range request.Interfaces.Interface {
		localReq.VppInterface(intf)
	}

	for _, bd := range request.BDs.BridgeDomains {
		localReq.BD(bd)
	}

	for _, xcon := range request.XCons.XConnectPairs {
		localReq.XConnect(xcon)
	}

	for _, acl := range request.ACLs.Acl {
		localReq.ACL(acl)
	}

	for _, route := range request.StaticRoutes.Route {
		localReq.StaticRoute(route)
	}

	err := localReq.Send().ReceiveReply()
	return &vppsvc.ResyncConfigResponse{}, err
}
