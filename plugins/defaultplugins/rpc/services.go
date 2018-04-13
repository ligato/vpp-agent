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
func (svc *ChangeVppSvc) PutInterfaces(ctx context.Context, request *interfaces.Interfaces) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqPut := localReq.Put()
	for _, intf := range request.Interface {
		localReqPut.VppInterface(intf)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelInterfaces deletes one or multiple interfaces by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelInterfaces(ctx context.Context, request *rpc.DelNamesRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqDel := localReq.Delete()
	for _, intfName := range request.Name {
		localReqDel.VppInterface(intfName)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// PutBDs creates or updates one or multiple BDs
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutBDs(ctx context.Context, request *l2.BridgeDomains) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqPut := localReq.Put()
	for _, bd := range request.BridgeDomains {
		localReqPut.BD(bd)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelBDs deletes one or multiple BDs by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelBDs(ctx context.Context, request *rpc.DelNamesRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqDel := localReq.Delete()
	for _, bdName := range request.Name {
		localReqDel.BD(bdName)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// PutXCons creates or updates one or multiple Cross Connects
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutXCons(ctx context.Context, request *l2.XConnectPairs) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqPut := localReq.Put()
	for _, xcon := range request.XConnectPairs {
		localReqPut.XConnect(xcon)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelXCons deletes one or multiple Cross Connects by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelXCons(ctx context.Context, request *rpc.DelNamesRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqDel := localReq.Delete()
	for _, rxIfaceName := range request.Name {
		localReqDel.XConnect(rxIfaceName)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// PutACLs creates or updates one or multiple ACLs
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutACLs(ctx context.Context, request *acl.AccessLists) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqPut := localReq.Put()
	for _, acl := range request.Acl {
		localReqPut.ACL(acl)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelACLs deletes one or multiple ACLs by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelACLs(ctx context.Context, request *rpc.DelNamesRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqDel := localReq.Delete()
	for _, aclName := range request.Name {
		localReqDel.ACL(aclName)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// PutStaticRoutes creates or updates one or multiple ACLs
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) PutStaticRoutes(ctx context.Context, request *l3.StaticRoutes) (
	*rpc.PutResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqPut := localReq.Put()
	for _, route := range request.Route {
		localReqPut.StaticRoute(route)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// DelStaticRoutes deletes one or multiple ACLs by their unique names
// (forwards the input to the localclient).
func (svc *ChangeVppSvc) DelStaticRoutes(ctx context.Context, request *rpc.DelStaticRoutesRequest) (*rpc.DelResponse, error) {
	localReq := localclient.DataChangeRequest("rpc")
	localReqDel := localReq.Delete()
	for _, route := range request.Route {
		localReqDel.StaticRoute(route.VRF, route.DstAddr, route.NextHopAddr)
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// ResyncConfig fills data resync request of defaultvppplugin configuration
// , i.e. forwards the input to the localclient.
func (svc *ResyncVppSvc) ResyncConfig(ctx context.Context, request *rpc.ResyncConfigRequest) (
	*rpc.ResyncConfigResponse, error) {

	localReq := localclient.DataResyncRequest("rpc")

	if request.Interfaces != nil {
		for _, intf := range request.Interfaces.Interface {
			localReq.VppInterface(intf)
		}
	}
	if request.BDs != nil {
		for _, bd := range request.BDs.BridgeDomains {
			localReq.BD(bd)
		}
	}
	if request.XCons != nil {
		for _, xcon := range request.XCons.XConnectPairs {
			localReq.XConnect(xcon)
		}
	}
	if request.ACLs != nil {
		for _, accessList := range request.ACLs.Acl {
			localReq.ACL(accessList)
		}
	}
	if request.StaticRoutes != nil {
		for _, route := range request.StaticRoutes.Route {
			localReq.StaticRoute(route)
		}
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.ResyncConfigResponse{}, err
}
