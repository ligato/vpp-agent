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
	"fmt"

	linuxIf "github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
	linuxL3 "github.com/ligato/vpp-agent/plugins/linux/model/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l2"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l3"

	"git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linux"
	"github.com/ligato/vpp-agent/plugins/vpp"
	"github.com/ligato/vpp-agent/plugins/vpp/model/acl"

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/vpp-agent/clientv1/linux"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"
	iflinuxcalls "github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	aclvppcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	ipsecvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
	l2vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	l3vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	l4vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l4plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/model/rpc"
	"golang.org/x/net/context"
)

// Plugin registers VPP GRPC services in *grpc.Server.
type Plugin struct {
	Deps

	// Channels
	vppChan  api.Channel
	dumpChan api.Channel

	// Services
	changeVppSvc ChangeVppSvc
	resyncVppSvc ResyncVppSvc
	getVppSvc    GetVppSvc
	notifSvc     NotificationSvc
}

// Deps - dependencies of Plugin
type Deps struct {
	infra.PluginDeps
	GRPCServer grpc.Server
	GoVppmux   govppmux.TraceAPI
	VPP        vpp.API
	Linux      linux.API
}

// ChangeVppSvc forwards GRPC request to the localclient.
type ChangeVppSvc struct {
	log logging.Logger
}

// ResyncVppSvc forwards GRPC request to the localclient.
type ResyncVppSvc struct {
	log logging.Logger
}

// GetVppSvc uses VPP/Linux plugin handlers to read VPP configuration
type GetVppSvc struct {
	log logging.Logger
	// VPP Handlers
	aclHandler   aclvppcalls.ACLVppRead
	ifHandler    ifvppcalls.IfVppRead
	bfdHandler   ifvppcalls.BfdVppRead
	natHandler   ifvppcalls.NatVppRead
	stnHandler   ifvppcalls.StnVppRead
	ipSecHandler ipsecvppcalls.IPSecVPPRead
	bdHandler    l2vppcalls.BridgeDomainVppRead
	fibHandler   l2vppcalls.FibVppRead
	xcHandler    l2vppcalls.XConnectVppRead
	arpHandler   l3vppcalls.ArpVppRead
	pArpHandler  l3vppcalls.ProxyArpVppRead
	rtHandler    l3vppcalls.RouteVppRead
	l4Handler    l4vppcalls.L4VppRead
	// Linux handlers
	linuxIfHandler iflinuxcalls.NetlinkAPI
	linuxL3Handler l3linuxcalls.NetlinkAPI
}

// Init sets plugin child loggers for changeVppSvc & resyncVppSvc.
func (p *Plugin) Init() (err error) {
	// VPP channels
	if p.vppChan, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	if p.dumpChan, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}

	// Data change
	p.changeVppSvc.log = p.Log.NewLogger("changeVppSvc")
	// Data resync
	p.resyncVppSvc.log = p.Log.NewLogger("resyncVppSvc")
	// Data get
	p.getVppSvc = GetVppSvc{
		log: p.Log.NewLogger("getVppSvc"),
	}
	p.initHandlers()
	// Notification service (represents GRPC client)
	p.notifSvc.log = p.Log.NewLogger("notifSvc")

	// Register all GRPC services if server is available. Register needs to be done
	// before 'ListenAndServe' is called in GRPC plugin
	grpcServer := p.GRPCServer.GetServer()
	if grpcServer != nil {
		rpc.RegisterDataChangeServiceServer(grpcServer, &p.changeVppSvc)
		rpc.RegisterDataResyncServiceServer(grpcServer, &p.resyncVppSvc)
		rpc.RegisterNotificationServiceServer(grpcServer, &p.notifSvc)
		if p.VPP != nil && p.Linux != nil {
			rpc.RegisterDataGetServiceServer(grpcServer, &p.getVppSvc)
		}
	}

	// Set grpc interface notification function for VPP plugin
	p.VPP.SetGRPCNotificationService(p.notifSvc.updateNotifications)

	return nil
}

// Close does nothing.
func (p *Plugin) Close() error {
	return nil
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

// GetAcls reads IP/MACIP access lists and returns them as an *AclResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetAcls(ctx context.Context, request *rpc.GetRequest) (*rpc.AclResponse, error) {
	var acls []*acl.AccessLists_Acl
	ipACLs, err := svc.aclHandler.DumpIPACL(nil)
	if err != nil {
		return nil, err
	}
	macIPACLs, err := svc.aclHandler.DumpMACIPACL(nil)
	if err != nil {
		return nil, err
	}
	for _, aclDetails := range ipACLs {
		acls = append(acls, aclDetails.ACL)
	}
	for _, aclDetails := range macIPACLs {
		acls = append(acls, aclDetails.ACL)
	}

	return &rpc.AclResponse{AccessLists: acls}, nil
}

// GetInterfaces reads interfaces and returns them as an *InterfaceResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetInterfaces(ctx context.Context, request *rpc.GetRequest) (*rpc.InterfaceResponse, error) {
	var ifs []*interfaces.Interfaces_Interface
	ifDetails, err := svc.ifHandler.DumpInterfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifDetails {
		ifs = append(ifs, iface.Interface)
	}

	return &rpc.InterfaceResponse{Interfaces: ifs}, nil
}

// GetIPSecSPDs reads IPSec SPD and returns them as an *IPSecSPDResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetIPSecSPDs(ctx context.Context, request *rpc.GetRequest) (*rpc.IPSecSPDResponse, error) {
	var spds []*ipsec.SecurityPolicyDatabases_SPD
	spdDetails, err := svc.ipSecHandler.DumpIPSecSPD()
	if err != nil {
		return nil, err
	}
	for _, spd := range spdDetails {
		spds = append(spds, spd.Spd)
	}

	return &rpc.IPSecSPDResponse{SPDs: spds}, nil
}

// GetIPSecSAs reads IPSec SA and returns them as an *IPSecSAResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetIPSecSAs(ctx context.Context, request *rpc.GetRequest) (*rpc.IPSecSAResponse, error) {
	var sas []*ipsec.SecurityAssociations_SA
	saDetails, err := svc.ipSecHandler.DumpIPSecSA()
	if err != nil {
		return nil, err
	}
	for _, sa := range saDetails {
		sas = append(sas, sa.Sa)
	}

	return &rpc.IPSecSAResponse{SAa: sas}, nil
}

// GetIPSecTunnels reads IPSec tunnels and returns them as an *IPSecTunnelResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetIPSecTunnels(ctx context.Context, request *rpc.GetRequest) (*rpc.IPSecTunnelResponse, error) {
	var tuns []*ipsec.TunnelInterfaces_Tunnel
	tunDetails, err := svc.ipSecHandler.DumpIPSecTunnelInterfaces()
	if err != nil {
		return nil, err
	}
	for _, tun := range tunDetails {
		tuns = append(tuns, tun.Tunnel)
	}

	return &rpc.IPSecTunnelResponse{Tunnels: tuns}, nil
}

// GetBDs reads bridge domains and returns them as an *BDResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetBDs(ctx context.Context, request *rpc.GetRequest) (*rpc.BDResponse, error) {
	var bds []*l2.BridgeDomains_BridgeDomain
	bdDetails, err := svc.bdHandler.DumpBridgeDomains()
	if err != nil {
		return nil, err
	}
	for _, bd := range bdDetails {
		bds = append(bds, bd.Bd)
	}

	return &rpc.BDResponse{BridgeDomains: bds}, nil
}

// GetFIBs reads FIBs and returns them as an *FibResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetFIBs(ctx context.Context, request *rpc.GetRequest) (*rpc.FibResponse, error) {
	var fibs []*l2.FibTable_FibEntry
	fibDetails, err := svc.fibHandler.DumpFIBTableEntries()
	if err != nil {
		return nil, err
	}
	for _, fib := range fibDetails {
		fibs = append(fibs, fib.Fib)
	}

	return &rpc.FibResponse{FIBs: fibs}, nil
}

// GetXConnects reads cross connects and returns them as an *XcResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetXConnects(ctx context.Context, request *rpc.GetRequest) (*rpc.XcResponse, error) {
	var xcs []*l2.XConnectPairs_XConnectPair
	xcDetails, err := svc.xcHandler.DumpXConnectPairs()
	if err != nil {
		return nil, err
	}
	for _, xc := range xcDetails {
		xcs = append(xcs, xc.Xc)
	}

	return &rpc.XcResponse{XCons: xcs}, nil
}

// GetRoutes reads VPP routes and returns them as an *RoutesResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetRoutes(ctx context.Context, request *rpc.GetRequest) (*rpc.RoutesResponse, error) {
	var routes []*l3.StaticRoutes_Route
	rtDetails, err := svc.rtHandler.DumpStaticRoutes()
	if err != nil {
		return nil, err
	}
	for _, rt := range rtDetails {
		routes = append(routes, rt.Route)
	}

	return &rpc.RoutesResponse{StaticRoutes: routes}, nil
}

// GetARPs reads VPP ARPs and returns them as an *ARPsResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetARPs(ctx context.Context, request *rpc.GetRequest) (*rpc.ARPsResponse, error) {
	var arps []*l3.ArpTable_ArpEntry
	arpDetails, err := svc.arpHandler.DumpArpEntries()
	if err != nil {
		return nil, err
	}
	for _, arp := range arpDetails {
		arps = append(arps, arp.Arp)
	}

	return &rpc.ARPsResponse{ArpEntries: arps}, nil
}

// GetLinuxInterfaces reads linux interfaces and returns them as an *LinuxInterfaceResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetLinuxInterfaces(ctx context.Context, request *rpc.GetRequest) (*rpc.LinuxInterfaceResponse, error) {
	var linuxIfs []*linuxIf.LinuxInterfaces_Interface
	ifDetails, err := svc.linuxIfHandler.DumpInterfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifDetails {
		linuxIfs = append(linuxIfs, iface.Interface)
	}

	return &rpc.LinuxInterfaceResponse{LinuxInterfaces: linuxIfs}, nil
}

// GetLinuxARPs reads linux ARPs and returns them as an *LinuxARPsResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetLinuxARPs(ctx context.Context, request *rpc.GetRequest) (*rpc.LinuxARPsResponse, error) {
	var linuxArps []*linuxL3.LinuxStaticArpEntries_ArpEntry
	arpDetails, err := svc.linuxL3Handler.DumpArpEntries()
	if err != nil {
		return nil, err
	}
	for _, arp := range arpDetails {
		linuxArps = append(linuxArps, arp.Arp)
	}

	return &rpc.LinuxARPsResponse{LinuxArpEntries: linuxArps}, nil
}

// GetLinuxRoutes reads linux routes and returns them as an *LinuxRoutesResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) GetLinuxRoutes(ctx context.Context, request *rpc.GetRequest) (*rpc.LinuxRoutesResponse, error) {
	var linuxRoutes []*linuxL3.LinuxStaticRoutes_Route
	rtDetails, err := svc.linuxL3Handler.DumpRoutes()
	if err != nil {
		return nil, err
	}
	for _, rt := range rtDetails {
		linuxRoutes = append(linuxRoutes, rt.Route)
	}

	return &rpc.LinuxRoutesResponse{LinuxRoutes: linuxRoutes}, nil
}

// Common method which puts or deletes data of every configuration type separately
func processRequest(ctx context.Context, data *rpc.DataRequest, request interface{}) error {
	switch r := request.(type) {
	case linuxclient.PutDSL:
		for _, aclItem := range data.AccessLists {
			r.ACL(aclItem)
		}
		for _, ifItem := range data.Interfaces {
			r.VppInterface(ifItem)
		}
		for _, spdItem := range data.SPDs {
			r.VppIPSecSPD(spdItem)
		}
		for _, saItem := range data.SAs {
			r.VppIPSecSA(saItem)
		}
		for _, tunItem := range data.Tunnels {
			r.VppIPSecTunnel(tunItem)
		}
		for _, sessionItem := range data.BfdSessions {
			r.BfdSession(sessionItem)
		}
		for _, keyItem := range data.BfdAuthKeys {
			r.BfdAuthKeys(keyItem)
		}
		if data.BfdEchoFunction != nil {
			r.BfdEchoFunction(data.BfdEchoFunction)
		}
		for _, bdItem := range data.BridgeDomains {
			r.BD(bdItem)
		}
		for _, fibItem := range data.FIBs {
			r.BDFIB(fibItem)
		}
		for _, xcItem := range data.XCons {
			r.XConnect(xcItem)
		}
		for _, rtItem := range data.StaticRoutes {
			r.StaticRoute(rtItem)
		}
		for _, arpItem := range data.ArpEntries {
			r.Arp(arpItem)
		}
		for _, paiItem := range data.ProxyArpInterfaces {
			r.ProxyArpInterfaces(paiItem)
		}
		for _, parItem := range data.ProxyArpRanges {
			r.ProxyArpRanges(parItem)
		}
		if data.L4Feature != nil {
			r.L4Features(data.L4Feature)
		}
		for _, anItem := range data.ApplicationNamespaces {
			r.AppNamespace(anItem)
		}
		for _, stnItem := range data.StnRules {
			r.StnRule(stnItem)
		}
		if data.NatGlobal != nil {
			r.NAT44Global(data.NatGlobal)
		}
		for _, natItem := range data.DNATs {
			r.NAT44DNat(natItem)
		}
		for _, ifItem := range data.LinuxInterfaces {
			r.LinuxInterface(ifItem)
		}
		for _, arpItem := range data.LinuxArpEntries {
			r.LinuxArpEntry(arpItem)
		}
		for _, rtItem := range data.LinuxRoutes {
			r.LinuxRoute(rtItem)
		}
	case linuxclient.DeleteDSL:
		for _, aclItem := range data.AccessLists {
			r.ACL(aclItem.AclName)
		}
		for _, ifItem := range data.Interfaces {
			r.VppInterface(ifItem.Name)
		}
		for _, spdItem := range data.SPDs {
			r.VppIPSecSPD(spdItem.Name)
		}
		for _, saItem := range data.SAs {
			r.VppIPSecSA(saItem.Name)
		}
		for _, tunItem := range data.Tunnels {
			r.VppIPSecTunnel(tunItem.Name)
		}
		for _, sessionItem := range data.BfdSessions {
			r.BfdSession(sessionItem.Interface)
		}
		for _, keyItem := range data.BfdAuthKeys {
			r.BfdAuthKeys(keyItem.Name)
		}
		if data.BfdEchoFunction != nil {
			r.BfdEchoFunction(data.BfdEchoFunction.Name)
		}
		for _, bdItem := range data.BridgeDomains {
			r.BD(bdItem.Name)
		}
		for _, fibItem := range data.FIBs {
			r.BDFIB(fibItem.BridgeDomain, fibItem.PhysAddress)
		}
		for _, xcItem := range data.XCons {
			r.XConnect(xcItem.ReceiveInterface)
		}
		for _, rtItem := range data.StaticRoutes {
			r.StaticRoute(rtItem.VrfId, rtItem.DstIpAddr, rtItem.NextHopAddr)
		}
		for _, arpItem := range data.ArpEntries {
			r.Arp(arpItem.Interface, arpItem.IpAddress)
		}
		for _, paiItem := range data.ProxyArpInterfaces {
			r.ProxyArpInterfaces(paiItem.Label)
		}
		for _, parItem := range data.ProxyArpRanges {
			r.ProxyArpRanges(parItem.Label)
		}
		if data.L4Feature != nil {
			r.L4Features()
		}
		for _, anItem := range data.ApplicationNamespaces {
			r.AppNamespace(anItem.NamespaceId)
		}
		for _, stnItem := range data.StnRules {
			r.StnRule(stnItem.RuleName)
		}
		if data.NatGlobal != nil {
			r.NAT44Global()
		}
		for _, natItem := range data.DNATs {
			r.NAT44DNat(natItem.Label)
		}
		for _, ifItem := range data.LinuxInterfaces {
			r.LinuxInterface(ifItem.Name)
		}
		for _, arpItem := range data.LinuxArpEntries {
			r.LinuxArpEntry(arpItem.Name)
		}
		for _, rtItem := range data.LinuxRoutes {
			r.LinuxRoute(rtItem.Name)
		}
	case linuxclient.DataResyncDSL:
		for _, aclItem := range data.AccessLists {
			r.ACL(aclItem)
		}
		for _, ifItem := range data.Interfaces {
			r.VppInterface(ifItem)
		}
		for _, spdItem := range data.SPDs {
			r.VppIPSecSPD(spdItem)
		}
		for _, saItem := range data.SAs {
			r.VppIPSecSA(saItem)
		}
		for _, tunItem := range data.Tunnels {
			r.VppIPSecTunnel(tunItem)
		}
		for _, sessionItem := range data.BfdSessions {
			r.BfdSession(sessionItem)
		}
		for _, keyItem := range data.BfdAuthKeys {
			r.BfdAuthKeys(keyItem)
		}
		if data.BfdEchoFunction != nil {
			r.BfdEchoFunction(data.BfdEchoFunction)
		}
		for _, bdItem := range data.BridgeDomains {
			r.BD(bdItem)
		}
		for _, fibItem := range data.FIBs {
			r.BDFIB(fibItem)
		}
		for _, xcItem := range data.XCons {
			r.XConnect(xcItem)
		}
		for _, rtItem := range data.StaticRoutes {
			r.StaticRoute(rtItem)
		}
		for _, arpItem := range data.ArpEntries {
			r.Arp(arpItem)
		}
		for _, paiItem := range data.ProxyArpInterfaces {
			r.ProxyArpInterfaces(paiItem)
		}
		for _, parItem := range data.ProxyArpRanges {
			r.ProxyArpRanges(parItem)
		}
		if data.L4Feature != nil {
			r.L4Features(data.L4Feature)
		}
		for _, anItem := range data.ApplicationNamespaces {
			r.AppNamespace(anItem)
		}
		for _, stnItem := range data.StnRules {
			r.StnRule(stnItem)
		}
		if data.NatGlobal != nil {
			r.NAT44Global(data.NatGlobal)
		}
		for _, natItem := range data.DNATs {
			r.NAT44DNat(natItem)
		}
		for _, ifItem := range data.LinuxInterfaces {
			r.LinuxInterface(ifItem)
		}
		for _, arpItem := range data.LinuxArpEntries {
			r.LinuxArpEntry(arpItem)
		}
		for _, rtItem := range data.LinuxRoutes {
			r.LinuxRoute(rtItem)
		}
	default:
		return fmt.Errorf("unknown type of request: %v", r)
	}

	return nil
}

// helper method initializes all VPP/Linux plugin handlers
func (p *Plugin) initHandlers() {
	// VPP Indexes
	ifIndexes := p.VPP.GetSwIfIndexes()
	bdIndexes := p.VPP.GetBDIndexes()
	spdIndexes := p.VPP.GetIPSecSPDIndexes()
	// Initialize VPP handlers
	p.getVppSvc.aclHandler = aclvppcalls.NewACLVppHandler(p.vppChan, p.dumpChan)
	p.getVppSvc.ifHandler = ifvppcalls.NewIfVppHandler(p.vppChan, p.Log)
	p.getVppSvc.bfdHandler = ifvppcalls.NewBfdVppHandler(p.vppChan, ifIndexes, p.Log)
	p.getVppSvc.natHandler = ifvppcalls.NewNatVppHandler(p.vppChan, p.dumpChan, ifIndexes, p.Log)
	p.getVppSvc.stnHandler = ifvppcalls.NewStnVppHandler(p.vppChan, ifIndexes, p.Log)
	p.getVppSvc.ipSecHandler = ipsecvppcalls.NewIPsecVppHandler(p.vppChan, ifIndexes, spdIndexes, p.Log)
	p.getVppSvc.bdHandler = l2vppcalls.NewBridgeDomainVppHandler(p.vppChan, ifIndexes, p.Log)
	p.getVppSvc.fibHandler = l2vppcalls.NewFibVppHandler(p.vppChan, p.dumpChan, ifIndexes, bdIndexes, p.Log)
	p.getVppSvc.xcHandler = l2vppcalls.NewXConnectVppHandler(p.vppChan, ifIndexes, p.Log)
	p.getVppSvc.arpHandler = l3vppcalls.NewArpVppHandler(p.vppChan, ifIndexes, p.Log)
	p.getVppSvc.pArpHandler = l3vppcalls.NewProxyArpVppHandler(p.vppChan, ifIndexes, p.Log)
	p.getVppSvc.rtHandler = l3vppcalls.NewRouteVppHandler(p.vppChan, ifIndexes, p.Log)
	p.getVppSvc.l4Handler = l4vppcalls.NewL4VppHandler(p.vppChan, p.Log)
	// Linux indexes and handlers
	if p.Linux != nil && !p.Linux.IsDisabled() {
		linuxIfIndexes := p.Linux.GetLinuxIfIndexes()
		linuxArpIndexes := p.Linux.GetLinuxARPIndexes()
		linuxRtIndexes := p.Linux.GetLinuxRouteIndexes()
		// Initialize Linux handlers
		linuxNsHandler := p.Linux.GetNamespaceHandler()
		p.getVppSvc.linuxIfHandler = iflinuxcalls.NewNetLinkHandler(linuxNsHandler, linuxIfIndexes, p.Log)
		p.getVppSvc.linuxL3Handler = l3linuxcalls.NewNetLinkHandler(linuxNsHandler, linuxIfIndexes, linuxArpIndexes, linuxRtIndexes, p.Log)
	}
}
