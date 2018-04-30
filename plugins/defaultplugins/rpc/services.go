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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l4"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/rpc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/stn"
	linuxIf "github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/interfaces"
	linuxL3 "github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/l3"
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

	// VPP

	if err := aclRequest(data.AccessLists, request); err != nil {
		wasErr = err
	}
	if err := vppInterfaceRequest(data.Interfaces, request); err != nil {
		wasErr = err
	}
	if err := bfdSessionRequest(data.BfdSessions, request); err != nil {
		wasErr = err
	}
	if err := bfdAuthenticationRequest(data.BfdAuthKeys, request); err != nil {
		wasErr = err
	}
	if err := bfdEchoFunctionRequest(data.BfdEchoFunction, request); err != nil {
		wasErr = err
	}
	if err := bridgeDomainRequest(data.BridgeDomains, request); err != nil {
		wasErr = err
	}
	if err := fibRequest(data.FIBs, request); err != nil {
		wasErr = err
	}
	if err := xConnectRequest(data.XCons, request); err != nil {
		wasErr = err
	}
	if err := vppRouteRequest(data.StaticRoutes, request); err != nil {
		wasErr = err
	}
	if err := vppArpRequest(data.ArpEntries, request); err != nil {
		wasErr = err
	}
	if err := proxyArpIfRequest(data.ProxyArpInterfaces, request); err != nil {
		wasErr = err
	}
	if err := proxyArpRngRequest(data.ProxyArpRanges, request); err != nil {
		wasErr = err
	}
	if err := l4Request(data.L4Feature, request); err != nil {
		wasErr = err
	}
	if err := appNsRequest(data.ApplicationNamespaces, request); err != nil {
		wasErr = err
	}
	if err := stnRequest(data.StnRules, request); err != nil {
		wasErr = err
	}
	if err := natGlobalRequest(data.NatGlobal, request); err != nil {
		wasErr = err
	}
	if err := dnatRequest(data.DNATs, request); err != nil {
		wasErr = err
	}

	// Linux

	if err := linuxInterfaceRequest(data.LinuxInterfaces, request); err != nil {
		wasErr = err
	}
	if err := linuxARPRequest(data.LinuxArpEntries, request); err != nil {
		wasErr = err
	}
	if err := linuxRouteRequest(data.LinuxRoutes, request); err != nil {
		wasErr = err
	}

	return wasErr
}

/* VPP requests (defaultplugins) */

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

// BFD session request forwards multiple BFD sessions to the localclient
func bfdSessionRequest(data []*bfd.SingleHopBFD_Session, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, sessionItem := range data {
			r.BfdSession(sessionItem)
		}
	case linux.DeleteDSL:
		for _, sessionItem := range data {
			r.BfdSession(sessionItem.Interface)
		}
	case linux.DataResyncDSL:
		for _, sessionItem := range data {
			r.BfdSession(sessionItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// BFD authentication key request forwards multiple authentication keys to the localclient
func bfdAuthenticationRequest(data []*bfd.SingleHopBFD_Key, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, keyItem := range data {
			r.BfdAuthKeys(keyItem)
		}
	case linux.DeleteDSL:
		for _, keyItem := range data {
			r.BfdAuthKeys(keyItem.Name)
		}
	case linux.DataResyncDSL:
		for _, keyItem := range data {
			r.BfdAuthKeys(keyItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// BFD echo function request forwards configuration to the localclient
func bfdEchoFunctionRequest(data *bfd.SingleHopBFD_EchoFunction, request interface{}) error {
	if data == nil {
		return nil
	}
	switch r := request.(type) {
	case linux.PutDSL:
		r.BfdEchoFunction(data)
	case linux.DeleteDSL:
		r.BfdEchoFunction(data.Name)
	case linux.DataResyncDSL:
		r.BfdEchoFunction(data)
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

// FIB request forwards multiple FIB enties to the localclient
func fibRequest(data []*l2.FibTable_FibEntry, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, fibItem := range data {
			r.BDFIB(fibItem)
		}
	case linux.DeleteDSL:
		for _, fibItem := range data {
			r.BDFIB(fibItem.BridgeDomain, fibItem.PhysAddress)
		}
	case linux.DataResyncDSL:
		for _, fibItem := range data {
			r.BDFIB(fibItem)
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
func vppRouteRequest(data []*l3.StaticRoutes_Route, request interface{}) error {
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

// VPP ARP request forwards multiple ARPs to the localclient
func vppArpRequest(data []*l3.ArpTable_ArpEntry, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, arpItem := range data {
			r.Arp(arpItem)
		}
	case linux.DeleteDSL:
		for _, arpItem := range data {
			r.Arp(arpItem.Interface, arpItem.IpAddress)
		}
	case linux.DataResyncDSL:
		for _, arpItem := range data {
			r.Arp(arpItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// Proxy ARP interface request forwards multiple Proxy ARP interfaces to the localclient
func proxyArpIfRequest(data []*l3.ProxyArpInterfaces_InterfaceList, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, paiItem := range data {
			r.ProxyArpInterfaces(paiItem)
		}
	case linux.DeleteDSL:
		for _, paiItem := range data {
			r.ProxyArpInterfaces(paiItem.Label)
		}
	case linux.DataResyncDSL:
		for _, paiItem := range data {
			r.ProxyArpInterfaces(paiItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// Proxy ARP range request forwards multiple proxy ARP ranges to the localclient
func proxyArpRngRequest(data []*l3.ProxyArpRanges_RangeList, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, parItem := range data {
			r.ProxyArpRanges(parItem)
		}
	case linux.DeleteDSL:
		for _, parItem := range data {
			r.ProxyArpRanges(parItem.Label)
		}
	case linux.DataResyncDSL:
		for _, parItem := range data {
			r.ProxyArpRanges(parItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// L4 features request forwards configuration to the localclient
func l4Request(data *l4.L4Features, request interface{}) error {
	if data == nil {
		return nil
	}
	switch r := request.(type) {
	case linux.PutDSL:
		r.L4Features(data)
	case linux.DeleteDSL:
		r.L4Features()
	case linux.DataResyncDSL:
		r.L4Features(data)
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// Application namespace request forwards application namespaces to the localclient
func appNsRequest(data []*l4.AppNamespaces_AppNamespace, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, anItem := range data {
			r.AppNamespace(anItem)
		}
	case linux.DeleteDSL:
		for _, anItem := range data {
			r.AppNamespace(anItem.NamespaceId)
		}
	case linux.DataResyncDSL:
		for _, anItem := range data {
			r.AppNamespace(anItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// STN request forwards multiple STNs to the localclient
func stnRequest(data []*stn.STN_Rule, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, stnItem := range data {
			r.StnRule(stnItem)
		}
	case linux.DeleteDSL:
		for _, stnItem := range data {
			r.StnRule(stnItem.Interface)
		}
	case linux.DataResyncDSL:
		for _, stnItem := range data {
			r.StnRule(stnItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// NAT global request forwards configuration to the localclient
func natGlobalRequest(data *nat.Nat44Global, request interface{}) error {
	if data == nil {
		return nil
	}
	switch r := request.(type) {
	case linux.PutDSL:
		r.NAT44Global(data)
	case linux.DeleteDSL:
		r.NAT44Global()
	case linux.DataResyncDSL:
		r.NAT44Global(data)
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// DNAT request forwards multiple DNAT configurations to the localclient
func dnatRequest(data []*nat.Nat44DNat_DNatConfig, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, natItem := range data {
			r.NAT44DNat(natItem)
		}
	case linux.DeleteDSL:
		for _, natItem := range data {
			r.NAT44DNat(natItem.Label)
		}
	case linux.DataResyncDSL:
		for _, natItem := range data {
			r.NAT44DNat(natItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

/* Linux requests (linuxplugin) */

// Linux interface request forwards multiple linux interfaces to the localclient
func linuxInterfaceRequest(data []*linuxIf.LinuxInterfaces_Interface, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, ifItem := range data {
			r.LinuxInterface(ifItem)
		}
	case linux.DeleteDSL:
		for _, ifItem := range data {
			r.LinuxInterface(ifItem.Name)
		}
	case linux.DataResyncDSL:
		for _, ifItem := range data {
			r.LinuxInterface(ifItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// Linux ARP request forwards multiple linux ARPs to the localclient
func linuxARPRequest(data []*linuxL3.LinuxStaticArpEntries_ArpEntry, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, arpItem := range data {
			r.LinuxArpEntry(arpItem)
		}
	case linux.DeleteDSL:
		for _, arpItem := range data {
			r.LinuxArpEntry(arpItem.Name)
		}
	case linux.DataResyncDSL:
		for _, arpItem := range data {
			r.LinuxArpEntry(arpItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}

// Linux Route request forwards multiple linux Routes to the localclient
func linuxRouteRequest(data []*linuxL3.LinuxStaticRoutes_Route, request interface{}) error {
	switch r := request.(type) {
	case linux.PutDSL:
		for _, rtItem := range data {
			r.LinuxRoute(rtItem)
		}
	case linux.DeleteDSL:
		for _, rtItem := range data {
			r.LinuxRoute(rtItem.Name)
		}
	case linux.DataResyncDSL:
		for _, rtItem := range data {
			r.LinuxRoute(rtItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}
	return nil
}
