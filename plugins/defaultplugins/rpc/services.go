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
	"fmt"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/vpp-agent/clientv1/linux"
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
	rpc.RegisterDataChangeServiceServer(grpcServer, &plugin.changeVppSvc)
	rpc.RegisterDataResyncServiceServer(grpcServer, &plugin.resyncVppSvc)

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

// Put adds configuration data present in data request to the VPP/Linux
func (svc *ChangeVppSvc) Put(ctx context.Context, data *rpc.DataRequest) (*rpc.PutResponse, error) {
	request := localclient.DataChangeRequest("rpc").Put()
	if err := processRequest(ctx, data, request); err != nil {
		return nil, err
	}
	err := request.Send().ReceiveReply()
	return &rpc.PutResponse{}, err
}

// Del removes configuration data present in data request from the VPP/linux
func (svc *ChangeVppSvc) Del(ctx context.Context, data *rpc.DataRequest) (*rpc.DelResponse, error) {
	request := localclient.DataChangeRequest("rpc").Delete()
	if err := processRequest(ctx, data, request); err != nil {
		return nil, err
	}
	err := request.Send().ReceiveReply()
	return &rpc.DelResponse{}, err
}

// Resync creates a resync request which adds data tp the VPP/linux
func (svc *ResyncVppSvc) Resync(ctx context.Context, data *rpc.DataRequest) (*rpc.ResyncResponse, error) {
	request := localclient.DataResyncRequest("rpc")
	if err := processRequest(ctx, data, request); err != nil {
		return nil, err
	}
	err := request.Send().ReceiveReply()
	return &rpc.ResyncResponse{}, err
}

// Common method which puts or deletes data of every configuration type separately
func processRequest(ctx context.Context, data *rpc.DataRequest, request interface{}) error {
	var wasErr error
	if err := aclRequest(data.ACLs, request); err != nil {
		wasErr = err
	}
	if err := vppInterfaceRequest(data.Interfaces, request); err != nil {
		wasErr = err
	}
	if err := bridgeDomainRequest(data.BDs, request); err != nil {
		wasErr = err
	}
	if err := xConnectRequest(data.XCons, request); err != nil {
		wasErr = err
	}
	if err := staticRouteRequest(data.StaticRoutes, request); err != nil {
		wasErr = err
	}

	return wasErr
}

// ACL request forwards multiple access lists to the localclient
func aclRequest(data []*acl.AccessLists_Acl, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, aclItem := range data {
			r.ACL(aclItem)
		}
	case linux.DeleteDSL:
		for _, aclItem := range data {
			r.ACL(aclItem.AclName)
		}
	case linux.DataResyncDSL:
		for _, aclItem := range data {
			r.ACL(aclItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// VPP interface request forwards multiple interfaces to the localclient
func vppInterfaceRequest(data []*interfaces.Interfaces_Interface, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, ifItem := range data {
			r.VppInterface(ifItem)
		}
	case linux.DeleteDSL:
		for _, ifItem := range data {
			r.VppInterface(ifItem.Name)
		}
	case linux.DataResyncDSL:
		for _, ifItem := range data {
			r.VppInterface(ifItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// BD request forwards multiple bridge domains to the localclient
func bridgeDomainRequest(data []*l2.BridgeDomains_BridgeDomain, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, bdItem := range data {
			r.BD(bdItem)
		}
	case linux.DeleteDSL:
		for _, bdItem := range data {
			r.BD(bdItem.Name)
		}
	case linux.DataResyncDSL:
		for _, bdItem := range data {
			r.BD(bdItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// XConnect request forwards multiple cross connects to the localclient
func xConnectRequest(data []*l2.XConnectPairs_XConnectPair, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, xcItem := range data {
			r.XConnect(xcItem)
		}
	case linux.DeleteDSL:
		for _, xcItem := range data {
			r.XConnect(xcItem.ReceiveInterface)
		}
	case linux.DataResyncDSL:
		for _, xcItem := range data {
			r.XConnect(xcItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// VPP static route request forwards multiple static routes to the localclient
func staticRouteRequest(data []*l3.StaticRoutes_Route, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, rtItem := range data {
			r.StaticRoute(rtItem)
		}
	case linux.DeleteDSL:
		for _, rtItem := range data {
			r.StaticRoute(rtItem.VrfId, rtItem.DstIpAddr, rtItem.NextHopAddr)
		}
	case linux.DataResyncDSL:
		for _, rtItem := range data {
			r.StaticRoute(rtItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}
