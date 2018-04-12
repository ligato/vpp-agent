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

//go:generate protoc --proto_path=../common/model/rpc --proto_path=$GOPATH/src --go_out=plugins=grpc:../common/model/rpc ../common/model/rpc/rpc.proto

package rpc

import (
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/rpc"
	"golang.org/x/net/context"
)

// GRPCSvcPlugin registers VPP GRPC services in *grpc.Server.
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

// Init sets plugin child loggers for changeVppSvc & resyncVppSvc.
func (plugin *GRPCSvcPlugin) Init() error {
	plugin.changeVppSvc.Log = plugin.Deps.Log.NewLogger("changeVppSvc")
	plugin.resyncVppSvc.Log = plugin.Deps.Log.NewLogger("resyncVppSvc")

	return nil
}

// AfterInit registers all GRPC services in vppscv package
// (be sure that defaultvppplugins are totally initialized).
func (plugin *GRPCSvcPlugin) AfterInit() error {
	grpcServer := plugin.Deps.GRPC.Server()
	rpc.RegisterChangeConfigServiceServer(grpcServer, &plugin.changeVppSvc)
	rpc.RegisterResyncConfigServiceServer(grpcServer, &plugin.resyncVppSvc)

	return nil
}

// Close does nothing.
func (plugin *GRPCSvcPlugin) Close() error {
	return nil
}

// ChangeVppSvc forwards GRPC request to the localclient.
type ChangeVppSvc struct {
	Log logging.Logger
}

// ResyncVppSvc forwards GRPC request to the localclient.
type ResyncVppSvc struct {
	Log logging.Logger
}

// PutInterfaces creates or updates one or multiple interfaces
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutInterface(ctx context.Context, request *interfaces.Interfaces_Interface) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	localReqPut.VppInterface(request)

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelInterfaces deletes one or multiple interfaces by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelInterface(ctx context.Context, request *rpc.DelNameRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	localReqDel.VppInterface(request.Name)

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// PutBDs creates or updates one or multiple BDs
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutBD(ctx context.Context, request *l2.BridgeDomains_BridgeDomain) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	localReqPut.BD(request)

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelBDs deletes one or multiple BDs by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelBD(ctx context.Context, request *rpc.DelNameRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	localReqDel.BD(request.Name)

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// PutXCons creates or updates one or multiple Cross Connects
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutXCon(ctx context.Context, request *l2.XConnectPairs_XConnectPair) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	localReqPut.XConnect(request)

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelXCons deletes one or multiple Cross Connects by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelXCon(ctx context.Context, request *rpc.DelNameRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	localReqDel.XConnect(request.Name)

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// PutACLs creates or updates one or multiple ACLs
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutACL(ctx context.Context, request *acl.AccessLists_Acl) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	localReqPut.ACL(request)

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelACLs deletes one or multiple ACLs by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelACL(ctx context.Context, request *rpc.DelNameRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	localReqDel.ACL(request.Name)

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// PutStaticRoutes creates or updates one or multiple ACLs
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutStaticRoute(ctx context.Context, request *l3.StaticRoutes_Route) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqPut := localReq.Put()
	localReqPut.StaticRoute(request)

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelStaticRoutes deletes one or multiple ACLs by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelStaticRoute(ctx context.Context, request *rpc.DelStaticRouteRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("vppsvc")
	localReqDel := localReq.Delete()
	localReqDel.StaticRoute(request.VRF, request.DstAddr, request.NextHopAddr)

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// ResyncConfig fills data resync request of defaultvppplugin configuration
// , i.e. forwards the input to the localclient.
func (svc *ResyncVppSvc) ResyncConfig(ctx context.Context, request *rpc.ResyncConfigRequest) (
	*rpc.ResyncConfigResponse, error) {

	localReq := localclient.DataResyncRequest("vppsvc")

	if request.Interfaces != nil && len(request.Interfaces.Interface) > 0 {
		for _, intf := range request.Interfaces.Interface {
			localReq.VppInterface(intf)
		}
	}
	if request.BDs != nil && len(request.BDs.BridgeDomains) > 0 {
		for _, bd := range request.BDs.BridgeDomains {
			localReq.BD(bd)
		}
	}
	if request.XCons != nil && len(request.XCons.XConnectPairs) > 0 {
		for _, xcon := range request.XCons.XConnectPairs {
			localReq.XConnect(xcon)
		}
	}
	if request.ACLs != nil && len(request.ACLs.Acl) > 0 {
		for _, accessList := range request.ACLs.Acl {
			localReq.ACL(accessList)
		}
	}
	if request.StaticRoutes != nil && len(request.StaticRoutes.Route) > 0 {
		for _, route := range request.StaticRoutes.Route {
			localReq.StaticRoute(route)
		}
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.ResyncConfigResponse{}, err
}
