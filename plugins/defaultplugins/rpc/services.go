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

//go:generate protoc --proto_path=../common/model/rpc --proto_path=$GOPATH/src --gogo_out=plugins=grpc:../common/model/rpc ../common/model/rpc/rpc.proto

package rpc

import (
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/rpc"
	"golang.org/x/net/context"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
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
	rpc.RegisterDataChangeServiceServer(grpcServer, &plugin.changeVppSvc)
	//rpc.RegisterChangeConfigServiceServer(grpcServer, &plugin.changeVppSvc)
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

// Put adds configuration data present in request to the VPP/Linux
func (svc *ChangeVppSvc) Put(ctx context.Context, request *rpc.DataRequest) (*rpc.PutResponse, error) {
	err := svc.putOrDelRequest(ctx, request, false)
	return &rpc.PutResponse{}, err
}

// Del removes configuration data present in request from the VPP/linux
func (svc *ChangeVppSvc) Del(ctx context.Context, request *rpc.DataRequest) (*rpc.DelResponse, error) {
	err := svc.putOrDelRequest(ctx, request, true)
	return &rpc.DelResponse{}, err
}

// Common method which puts or deletes data of every configuration type separately
func (svc *ChangeVppSvc) putOrDelRequest(ctx context.Context, request *rpc.DataRequest, isDelete bool) error {
	var wasErr error
	if err := aclRequest(request.ACLs, isDelete); err != nil {
		wasErr = err
	}
	if err := vppInterfaceRequest(request.Interfaces, isDelete); err != nil {
		wasErr = err
	}
	if err := bridgeDomainRequest(request.BDs, isDelete); err != nil {
		wasErr = err
	}
	if err := xConnectRequest(request.XCons, isDelete); err != nil {
		wasErr = err
	}
	if err := staticRouteRequest(request.StaticRoutes, isDelete); err != nil {
		wasErr = err
	}

	return wasErr
}

// ACL request forwards multiple access lists to the localclient
func aclRequest(data []*acl.AccessLists_Acl, delete bool) error {
	changeRequest := localclient.DataChangeRequest("rpc")
	if delete {
		delChangeRequest := changeRequest.Delete()
		for _, aclItem := range data {
			delChangeRequest.ACL(aclItem.AclName)
		}
	} else {
		putChangeRequest := changeRequest.Put()
		for _, aclItem := range data {
			putChangeRequest.ACL(aclItem)
		}
	}
	return changeRequest.Send().ReceiveReply()
}

// VPP interface request forwards multiple interfaces to the localclient
func vppInterfaceRequest(data []*interfaces.Interfaces_Interface, delete bool) error {
	changeRequest := localclient.DataChangeRequest("rpc")
	if delete {
		delChangeRequest := changeRequest.Delete()
		for _, ifItem := range data {
			delChangeRequest.VppInterface(ifItem.Name)
		}
	} else {
		putChangeRequest := changeRequest.Put()
		for _, ifItem := range data {
			putChangeRequest.VppInterface(ifItem)
		}
	}
	return changeRequest.Send().ReceiveReply()
}

// BD request forwards multiple bridge domains to the localclient
func bridgeDomainRequest(data []*l2.BridgeDomains_BridgeDomain, delete bool) error {
	changeRequest := localclient.DataChangeRequest("rpc")
	if delete {
		delChangeRequest := changeRequest.Delete()
		for _, bdItem := range data {
			delChangeRequest.BD(bdItem.Name)
		}
	} else {
		putChangeRequest := changeRequest.Put()
		for _, bdItem := range data {
			putChangeRequest.BD(bdItem)
		}
	}
	return changeRequest.Send().ReceiveReply()
}

// XConnect request forwards multiple cross connects to the localclient
func xConnectRequest(data []*l2.XConnectPairs_XConnectPair, delete bool) error {
	changeRequest := localclient.DataChangeRequest("rpc")
	if delete {
		delChangeRequest := changeRequest.Delete()
		for _, xcItem := range data {
			delChangeRequest.XConnect(xcItem.ReceiveInterface)
		}
	} else {
		putChangeRequest := changeRequest.Put()
		for _, xcItem := range data {
			putChangeRequest.XConnect(xcItem)
		}
	}
	return changeRequest.Send().ReceiveReply()
}

// VPP static route request forwards multiple static routes to the localclient
func staticRouteRequest(data []*l3.StaticRoutes_Route, delete bool) error {
	changeRequest := localclient.DataChangeRequest("rpc")
	if delete {
		delChangeRequest := changeRequest.Delete()
		for _, routeItem := range data {
			delChangeRequest.StaticRoute(routeItem.VrfId, routeItem.DstIpAddr, routeItem.NextHopAddr)
		}
	} else {
		putChangeRequest := changeRequest.Put()
		for _, routeItem := range data {
			putChangeRequest.StaticRoute(routeItem)
		}
	}
	return changeRequest.Send().ReceiveReply()
}

// ResyncConfig fills data resync request of defaultvppplugin configuration
// , i.e. forwards the input to the localclient.
func (svc *ResyncVppSvc) ResyncConfig(ctx context.Context, request *rpc.ResyncConfigRequest) (
	*rpc.ResyncConfigResponse, error) {

	localReq := localclient.DataResyncRequest("rpc")

	if request.Interfaces != nil {
		for _, intf := range request.Interfaces.Interfaces {
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
		for _, accessList := range request.ACLs.Acls {
			localReq.ACL(accessList)
		}
	}
	if request.StaticRoutes != nil {
		for _, route := range request.StaticRoutes.Routes {
			localReq.StaticRoute(route)
		}
	}

	err := localReq.Send().ReceiveReply()
	return &rpc.ResyncConfigResponse{}, err
}
