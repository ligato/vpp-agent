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
	"github.com/go-errors/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/plugins/vpp/model/bfd"
	"github.com/ligato/vpp-agent/plugins/vpp/model/l4"
	"github.com/ligato/vpp-agent/plugins/vpp/model/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/model/stn"

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
	"github.com/ligato/cn-infra/servicelabel"
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
	dumpVppSvc   GetVppSvc
	notifSvc     NotificationSvc
}

// Deps - dependencies of Plugin
type Deps struct {
	infra.PluginDeps
	GRPCServer   grpc.Server
	Brokers      map[string]keyval.KvProtoPlugin
	ServiceLabel servicelabel.ReaderAPI
	GoVppmux     govppmux.TraceAPI
	VPP          vpp.API
	Linux        linux.API
}

// Config groups configurations fields.
type Config struct {
	Broker string `json:"persistence-db"`
}

// ChangeVppSvc forwards GRPC request to the localclient.
type ChangeVppSvc struct {
	log logging.Logger
	pb  keyval.ProtoBroker
}

// ResyncVppSvc forwards GRPC request to the localclient.
type ResyncVppSvc struct {
	log logging.Logger
	pb  keyval.ProtoBroker
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
	p.dumpVppSvc = GetVppSvc{
		log: p.Log.NewLogger("dumpVppSvc"),
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
			rpc.RegisterDataDumpServiceServer(grpcServer, &p.dumpVppSvc)
		}
	}

	// Set grpc interface notification function for VPP plugin
	p.VPP.SetGRPCNotificationService(p.notifSvc.updateNotifications)

	// Get DB broker to persis files
	var wasErr error
	defer func() {
		// TODO workaround prevents crash if persistent db is defined but the required plugin is disabled because
		// of the missing config file (attempt to create a new broker from nil proto wrapper object, fix in cn-infra)
		if r := recover(); r != nil {
			if broker, err := p.getBrokerFromConfig(); err != nil {
				wasErr = err
			} else if broker != nil {
				protoBroker := broker.NewBroker(p.ServiceLabel.GetAgentPrefix())
				p.resyncVppSvc.pb, p.changeVppSvc.pb = protoBroker, protoBroker
			}
		} else {
			p.Log.Warnf("grpc plugin recovered from panic, make sure plugins for persistence are already loaded: %v",
				p.Brokers)
		}
	}()

	return wasErr
}

// Close does nothing.
func (p *Plugin) Close() error {
	return nil
}

// Put adds configuration data present in data request to the VPP/Linux
func (svc *ChangeVppSvc) Put(ctx context.Context, data *rpc.DataRequest) (*rpc.PutResponse, error) {
	request := localclient.DataChangeRequest("rpc").Put()
	dataSet := make(map[string]proto.Message)
	processPutRequest(data, request, dataSet)
	if svc.pb != nil {
		for key, value := range dataSet {
			if err := svc.pb.Put(key, value); err != nil {
				return &rpc.PutResponse{}, errors.Errorf("failed to persist (put) GRPC data: %v", err)
			}
		}
	}
	return &rpc.PutResponse{}, request.Send().ReceiveReply()
}

// Del removes configuration data present in data request from the VPP/linux
func (svc *ChangeVppSvc) Del(ctx context.Context, data *rpc.DataRequest) (*rpc.DelResponse, error) {
	request := localclient.DataChangeRequest("rpc").Delete()
	dataSet := make(map[string]proto.Message)
	processDelRequest(data, request, dataSet)
	if svc.pb != nil {
		for key := range dataSet {
			if _, err := svc.pb.Delete(key); err != nil {
				return &rpc.DelResponse{}, errors.Errorf("failed to persist (delete) GRPC data: %v", err)
			}
		}
	}
	return &rpc.DelResponse{}, request.Send().ReceiveReply()
}

// Resync creates a resync request which adds data tp the VPP/linux
func (svc *ResyncVppSvc) Resync(ctx context.Context, data *rpc.DataRequest) (*rpc.ResyncResponse, error) {
	request := localclient.DataResyncRequest("rpc")
	dataSet := make(map[string]proto.Message)
	processResyncRequest(data, request, dataSet)
	if svc.pb != nil {
		for key, value := range dataSet {
			if err := svc.pb.Put(key, value); err != nil {
				return &rpc.ResyncResponse{}, errors.Errorf("failed to persist (resync) GRPC data: %v", err)
			}
		}
	}
	return &rpc.ResyncResponse{}, request.Send().ReceiveReply()
}

// DumpAcls reads IP/MACIP access lists and returns them as an *AclResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpAcls(ctx context.Context, request *rpc.DumpRequest) (*rpc.AclResponse, error) {
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

// DumpInterfaces reads interfaces and returns them as an *InterfaceResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpInterfaces(ctx context.Context, request *rpc.DumpRequest) (*rpc.InterfaceResponse, error) {
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

// DumpIPSecSPDs reads IPSec SPD and returns them as an *IPSecSPDResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpIPSecSPDs(ctx context.Context, request *rpc.DumpRequest) (*rpc.IPSecSPDResponse, error) {
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

// DumpIPSecSAs reads IPSec SA and returns them as an *IPSecSAResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpIPSecSAs(ctx context.Context, request *rpc.DumpRequest) (*rpc.IPSecSAResponse, error) {
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

// DumpIPSecTunnels reads IPSec tunnels and returns them as an *IPSecTunnelResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpIPSecTunnels(ctx context.Context, request *rpc.DumpRequest) (*rpc.IPSecTunnelResponse, error) {
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

// DumpBDs reads bridge domains and returns them as an *BDResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpBDs(ctx context.Context, request *rpc.DumpRequest) (*rpc.BDResponse, error) {
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

// DumpFIBs reads FIBs and returns them as an *FibResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpFIBs(ctx context.Context, request *rpc.DumpRequest) (*rpc.FibResponse, error) {
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

// DumpXConnects reads cross connects and returns them as an *XcResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpXConnects(ctx context.Context, request *rpc.DumpRequest) (*rpc.XcResponse, error) {
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

// DumpRoutes reads VPP routes and returns them as an *RoutesResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpRoutes(ctx context.Context, request *rpc.DumpRequest) (*rpc.RoutesResponse, error) {
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

// DumpARPs reads VPP ARPs and returns them as an *ARPsResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpARPs(ctx context.Context, request *rpc.DumpRequest) (*rpc.ARPsResponse, error) {
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

// DumpLinuxInterfaces reads linux interfaces and returns them as an *LinuxInterfaceResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpLinuxInterfaces(ctx context.Context, request *rpc.DumpRequest) (*rpc.LinuxInterfaceResponse, error) {
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

// DumpLinuxARPs reads linux ARPs and returns them as an *LinuxARPsResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpLinuxARPs(ctx context.Context, request *rpc.DumpRequest) (*rpc.LinuxARPsResponse, error) {
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

// DumpLinuxRoutes reads linux routes and returns them as an *LinuxRoutesResponse. If reading ends up with error,
// only error is send back in response
func (svc *GetVppSvc) DumpLinuxRoutes(ctx context.Context, request *rpc.DumpRequest) (*rpc.LinuxRoutesResponse, error) {
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

func processPutRequest(reqData *rpc.DataRequest, req linuxclient.PutDSL, dataSet map[string]proto.Message) {
	for _, aclItem := range reqData.AccessLists {
		req.ACL(aclItem)
		dataSet[acl.Key(aclItem.AclName)] = aclItem
	}
	for _, ifItem := range reqData.Interfaces {
		req.VppInterface(ifItem)
		dataSet[interfaces.InterfaceKey(ifItem.Name)] = ifItem
	}
	for _, spdItem := range reqData.SPDs {
		req.VppIPSecSPD(spdItem)
		dataSet[ipsec.SPDKey(spdItem.Name)] = spdItem
	}
	for _, saItem := range reqData.SAs {
		req.VppIPSecSA(saItem)
		dataSet[ipsec.SAKey(saItem.Name)] = saItem
	}
	for _, tunItem := range reqData.Tunnels {
		req.VppIPSecTunnel(tunItem)
		dataSet[ipsec.TunnelKey(tunItem.Name)] = tunItem
	}
	for _, sessionItem := range reqData.BfdSessions {
		req.BfdSession(sessionItem)
		dataSet[bfd.SessionKey(sessionItem.Interface)] = sessionItem
	}
	for _, keyItem := range reqData.BfdAuthKeys {
		req.BfdAuthKeys(keyItem)
		dataSet[bfd.AuthKeysKey(keyItem.Name)] = keyItem
	}
	if reqData.BfdEchoFunction != nil {
		req.BfdEchoFunction(reqData.BfdEchoFunction)
		dataSet[bfd.EchoFunctionKey(reqData.BfdEchoFunction.Name)] = reqData.BfdEchoFunction
	}
	for _, bdItem := range reqData.BridgeDomains {
		req.BD(bdItem)
		dataSet[l2.BridgeDomainKey(bdItem.Name)] = bdItem
	}
	for _, fibItem := range reqData.FIBs {
		req.BDFIB(fibItem)
		dataSet[l2.FibKey(fibItem.BridgeDomain, fibItem.PhysAddress)] = fibItem
	}
	for _, xcItem := range reqData.XCons {
		req.XConnect(xcItem)
		dataSet[l2.XConnectKey(xcItem.ReceiveInterface)] = xcItem
	}
	for _, rtItem := range reqData.StaticRoutes {
		req.StaticRoute(rtItem)
		dataSet[l3.RouteKey(rtItem.VrfId, rtItem.DstIpAddr, rtItem.NextHopAddr)] = rtItem
	}
	for _, arpItem := range reqData.ArpEntries {
		req.Arp(arpItem)
		dataSet[l3.ArpEntryKey(arpItem.Interface, arpItem.IpAddress)] = arpItem
	}
	for _, paiItem := range reqData.ProxyArpInterfaces {
		req.ProxyArpInterfaces(paiItem)
		dataSet[l3.ProxyArpInterfaceKey(paiItem.Label)] = paiItem
	}
	for _, parItem := range reqData.ProxyArpRanges {
		req.ProxyArpRanges(parItem)
		dataSet[l3.ProxyArpRangeKey(parItem.Label)] = parItem
	}
	if reqData.L4Feature != nil {
		req.L4Features(reqData.L4Feature)
		dataSet[l4.FeatureKey()] = reqData.L4Feature
	}
	for _, anItem := range reqData.ApplicationNamespaces {
		req.AppNamespace(anItem)
		dataSet[l4.AppNamespacesKey(anItem.NamespaceId)] = anItem
	}
	for _, stnItem := range reqData.StnRules {
		req.StnRule(stnItem)
		dataSet[stn.Key(stnItem.RuleName)] = stnItem
	}
	if reqData.NatGlobal != nil {
		req.NAT44Global(reqData.NatGlobal)
		dataSet[nat.Prefix+nat.GlobalPrefix] = reqData.NatGlobal
	}
	for _, natItem := range reqData.DNATs {
		req.NAT44DNat(natItem)
		dataSet[nat.DNatKey(natItem.Label)] = natItem
	}
	for _, ifItem := range reqData.LinuxInterfaces {
		req.LinuxInterface(ifItem)
		dataSet[linuxIf.InterfaceKey(ifItem.Name)] = ifItem
	}
	for _, arpItem := range reqData.LinuxArpEntries {
		req.LinuxArpEntry(arpItem)
		dataSet[linuxL3.StaticRouteKey(arpItem.Name)] = arpItem
	}
	for _, rtItem := range reqData.LinuxRoutes {
		req.LinuxRoute(rtItem)
		dataSet[linuxL3.StaticArpKey(rtItem.Name)] = rtItem
	}
}

func processDelRequest(reqData *rpc.DataRequest, req linuxclient.DeleteDSL, dataSet map[string]proto.Message) {
	for _, aclItem := range reqData.AccessLists {
		req.ACL(aclItem.AclName)
		dataSet[acl.Key(aclItem.AclName)] = aclItem
	}
	for _, ifItem := range reqData.Interfaces {
		req.VppInterface(ifItem.Name)
		dataSet[interfaces.InterfaceKey(ifItem.Name)] = ifItem
	}
	for _, spdItem := range reqData.SPDs {
		req.VppIPSecSPD(spdItem.Name)
		dataSet[ipsec.SPDKey(spdItem.Name)] = spdItem
	}
	for _, saItem := range reqData.SAs {
		req.VppIPSecSA(saItem.Name)
		dataSet[ipsec.SAKey(saItem.Name)] = saItem
	}
	for _, tunItem := range reqData.Tunnels {
		req.VppIPSecTunnel(tunItem.Name)
		dataSet[ipsec.TunnelKey(tunItem.Name)] = tunItem
	}
	for _, sessionItem := range reqData.BfdSessions {
		req.BfdSession(sessionItem.Interface)
		dataSet[bfd.SessionKey(sessionItem.Interface)] = sessionItem
	}
	for _, keyItem := range reqData.BfdAuthKeys {
		req.BfdAuthKeys(keyItem.Name)
		dataSet[bfd.AuthKeysKey(keyItem.Name)] = keyItem
	}
	if reqData.BfdEchoFunction != nil {
		req.BfdEchoFunction(reqData.BfdEchoFunction.Name)
		dataSet[bfd.EchoFunctionKey(reqData.BfdEchoFunction.Name)] = reqData.BfdEchoFunction
	}
	for _, bdItem := range reqData.BridgeDomains {
		req.BD(bdItem.Name)
		dataSet[l2.BridgeDomainKey(bdItem.Name)] = bdItem
	}
	for _, fibItem := range reqData.FIBs {
		req.BDFIB(fibItem.BridgeDomain, fibItem.PhysAddress)
		dataSet[l2.FibKey(fibItem.BridgeDomain, fibItem.PhysAddress)] = fibItem
	}
	for _, xcItem := range reqData.XCons {
		req.XConnect(xcItem.ReceiveInterface)
		dataSet[l2.XConnectKey(xcItem.ReceiveInterface)] = xcItem
	}
	for _, rtItem := range reqData.StaticRoutes {
		req.StaticRoute(rtItem.VrfId, rtItem.DstIpAddr, rtItem.NextHopAddr)
		dataSet[l3.RouteKey(rtItem.VrfId, rtItem.DstIpAddr, rtItem.NextHopAddr)] = rtItem
	}
	for _, arpItem := range reqData.ArpEntries {
		req.Arp(arpItem.Interface, arpItem.IpAddress)
		dataSet[l3.ArpEntryKey(arpItem.Interface, arpItem.IpAddress)] = arpItem
	}
	for _, paiItem := range reqData.ProxyArpInterfaces {
		req.ProxyArpInterfaces(paiItem.Label)
		dataSet[l3.ProxyArpInterfaceKey(paiItem.Label)] = paiItem
	}
	for _, parItem := range reqData.ProxyArpRanges {
		req.ProxyArpRanges(parItem.Label)
		dataSet[l3.ProxyArpRangeKey(parItem.Label)] = parItem
	}
	if reqData.L4Feature != nil {
		req.L4Features()
		dataSet[l4.FeatureKey()] = reqData.L4Feature
	}
	for _, anItem := range reqData.ApplicationNamespaces {
		req.AppNamespace(anItem.NamespaceId)
		dataSet[l4.AppNamespacesKey(anItem.NamespaceId)] = anItem
	}
	for _, stnItem := range reqData.StnRules {
		req.StnRule(stnItem.RuleName)
		dataSet[stn.Key(stnItem.RuleName)] = stnItem
	}
	if reqData.NatGlobal != nil {
		req.NAT44Global()
		dataSet[nat.Prefix+nat.GlobalPrefix] = reqData.NatGlobal
	}
	for _, natItem := range reqData.DNATs {
		req.NAT44DNat(natItem.Label)
		dataSet[nat.DNatKey(natItem.Label)] = natItem
	}
	for _, ifItem := range reqData.LinuxInterfaces {
		req.LinuxInterface(ifItem.Name)
		dataSet[linuxIf.InterfaceKey(ifItem.Name)] = ifItem
	}
	for _, arpItem := range reqData.LinuxArpEntries {
		req.LinuxArpEntry(arpItem.Name)
		dataSet[linuxL3.StaticRouteKey(arpItem.Name)] = arpItem
	}
	for _, rtItem := range reqData.LinuxRoutes {
		req.LinuxRoute(rtItem.Name)
		dataSet[linuxL3.StaticArpKey(rtItem.Name)] = rtItem
	}
}

func processResyncRequest(reqData *rpc.DataRequest, req linuxclient.DataResyncDSL, dataSet map[string]proto.Message) {
	for _, aclItem := range reqData.AccessLists {
		req.ACL(aclItem)
		dataSet[acl.Key(aclItem.AclName)] = aclItem
	}
	for _, ifItem := range reqData.Interfaces {
		req.VppInterface(ifItem)
		dataSet[interfaces.InterfaceKey(ifItem.Name)] = ifItem
	}
	for _, spdItem := range reqData.SPDs {
		req.VppIPSecSPD(spdItem)
		dataSet[ipsec.SPDKey(spdItem.Name)] = spdItem
	}
	for _, saItem := range reqData.SAs {
		req.VppIPSecSA(saItem)
		dataSet[ipsec.SAKey(saItem.Name)] = saItem
	}
	for _, tunItem := range reqData.Tunnels {
		req.VppIPSecTunnel(tunItem)
		dataSet[ipsec.TunnelKey(tunItem.Name)] = tunItem
	}
	for _, sessionItem := range reqData.BfdSessions {
		req.BfdSession(sessionItem)
		dataSet[bfd.SessionKey(sessionItem.Interface)] = sessionItem
	}
	for _, keyItem := range reqData.BfdAuthKeys {
		req.BfdAuthKeys(keyItem)
		dataSet[bfd.AuthKeysKey(keyItem.Name)] = keyItem
	}
	if reqData.BfdEchoFunction != nil {
		req.BfdEchoFunction(reqData.BfdEchoFunction)
		dataSet[bfd.EchoFunctionKey(reqData.BfdEchoFunction.Name)] = reqData.BfdEchoFunction
	}
	for _, bdItem := range reqData.BridgeDomains {
		req.BD(bdItem)
		dataSet[l2.BridgeDomainKey(bdItem.Name)] = bdItem
	}
	for _, fibItem := range reqData.FIBs {
		req.BDFIB(fibItem)
		dataSet[l2.FibKey(fibItem.BridgeDomain, fibItem.PhysAddress)] = fibItem
	}
	for _, xcItem := range reqData.XCons {
		req.XConnect(xcItem)
		dataSet[l2.XConnectKey(xcItem.ReceiveInterface)] = xcItem
	}
	for _, rtItem := range reqData.StaticRoutes {
		req.StaticRoute(rtItem)
		dataSet[l3.RouteKey(rtItem.VrfId, rtItem.DstIpAddr, rtItem.NextHopAddr)] = rtItem
	}
	for _, arpItem := range reqData.ArpEntries {
		req.Arp(arpItem)
		dataSet[l3.ArpEntryKey(arpItem.Interface, arpItem.IpAddress)] = arpItem
	}
	for _, paiItem := range reqData.ProxyArpInterfaces {
		req.ProxyArpInterfaces(paiItem)
		dataSet[l3.ProxyArpInterfaceKey(paiItem.Label)] = paiItem
	}
	for _, parItem := range reqData.ProxyArpRanges {
		req.ProxyArpRanges(parItem)
		dataSet[l3.ProxyArpRangeKey(parItem.Label)] = parItem
	}
	if reqData.L4Feature != nil {
		req.L4Features(reqData.L4Feature)
		dataSet[l4.FeatureKey()] = reqData.L4Feature
	}
	for _, anItem := range reqData.ApplicationNamespaces {
		req.AppNamespace(anItem)
		dataSet[l4.AppNamespacesKey(anItem.NamespaceId)] = anItem
	}
	for _, stnItem := range reqData.StnRules {
		req.StnRule(stnItem)
		dataSet[stn.Key(stnItem.RuleName)] = stnItem
	}
	if reqData.NatGlobal != nil {
		req.NAT44Global(reqData.NatGlobal)
		dataSet[nat.Prefix+nat.GlobalPrefix] = reqData.NatGlobal
	}
	for _, natItem := range reqData.DNATs {
		req.NAT44DNat(natItem)
		dataSet[nat.DNatKey(natItem.Label)] = natItem
	}
	for _, ifItem := range reqData.LinuxInterfaces {
		req.LinuxInterface(ifItem)
		dataSet[linuxIf.InterfaceKey(ifItem.Name)] = ifItem
	}
	for _, arpItem := range reqData.LinuxArpEntries {
		req.LinuxArpEntry(arpItem)
		dataSet[linuxL3.StaticRouteKey(arpItem.Name)] = arpItem
	}
	for _, rtItem := range reqData.LinuxRoutes {
		req.LinuxRoute(rtItem)
		dataSet[linuxL3.StaticArpKey(rtItem.Name)] = rtItem
	}
}

// helper method initializes all VPP/Linux plugin handlers
func (p *Plugin) initHandlers() {
	// VPP Indexes
	ifIndexes := p.VPP.GetSwIfIndexes()
	bdIndexes := p.VPP.GetBDIndexes()
	spdIndexes := p.VPP.GetIPSecSPDIndexes()
	// Initialize VPP handlers
	p.dumpVppSvc.aclHandler = aclvppcalls.NewACLVppHandler(p.vppChan, p.dumpChan)
	p.dumpVppSvc.ifHandler = ifvppcalls.NewIfVppHandler(p.vppChan, p.Log)
	p.dumpVppSvc.bfdHandler = ifvppcalls.NewBfdVppHandler(p.vppChan, ifIndexes, p.Log)
	p.dumpVppSvc.natHandler = ifvppcalls.NewNatVppHandler(p.vppChan, p.dumpChan, ifIndexes, p.Log)
	p.dumpVppSvc.stnHandler = ifvppcalls.NewStnVppHandler(p.vppChan, ifIndexes, p.Log)
	p.dumpVppSvc.ipSecHandler = ipsecvppcalls.NewIPsecVppHandler(p.vppChan, ifIndexes, spdIndexes, p.Log)
	p.dumpVppSvc.bdHandler = l2vppcalls.NewBridgeDomainVppHandler(p.vppChan, ifIndexes, p.Log)
	p.dumpVppSvc.fibHandler = l2vppcalls.NewFibVppHandler(p.vppChan, p.dumpChan, ifIndexes, bdIndexes, p.Log)
	p.dumpVppSvc.xcHandler = l2vppcalls.NewXConnectVppHandler(p.vppChan, ifIndexes, p.Log)
	p.dumpVppSvc.arpHandler = l3vppcalls.NewArpVppHandler(p.vppChan, ifIndexes, p.Log)
	p.dumpVppSvc.pArpHandler = l3vppcalls.NewProxyArpVppHandler(p.vppChan, ifIndexes, p.Log)
	p.dumpVppSvc.rtHandler = l3vppcalls.NewRouteVppHandler(p.vppChan, ifIndexes, p.Log)
	p.dumpVppSvc.l4Handler = l4vppcalls.NewL4VppHandler(p.vppChan, p.Log)
	// Linux indexes and handlers
	if p.Linux != nil && !p.Linux.IsDisabled() {
		linuxIfIndexes := p.Linux.GetLinuxIfIndexes()
		linuxArpIndexes := p.Linux.GetLinuxARPIndexes()
		linuxRtIndexes := p.Linux.GetLinuxRouteIndexes()
		// Initialize Linux handlers
		linuxNsHandler := p.Linux.GetNamespaceHandler()
		p.dumpVppSvc.linuxIfHandler = iflinuxcalls.NewNetLinkHandler(linuxNsHandler, linuxIfIndexes, p.Log)
		p.dumpVppSvc.linuxL3Handler = l3linuxcalls.NewNetLinkHandler(linuxNsHandler, linuxIfIndexes, linuxArpIndexes, linuxRtIndexes, p.Log)
	}
}

// Returns database broker defined in config file. The broker also has to be available in map of proto writers
func (p *Plugin) getBrokerFromConfig() (keyval.KvProtoPlugin, error) {
	config := &Config{}

	found, err := p.Cfg.LoadValue(config)
	if err != nil {
		return nil, err
	} else if !found {
		p.Log.Debug("rpc-plugin config not found")
		return nil, nil
	}
	p.Log.Debugf("rpc-plugin config found: %+v", config)

	if config != nil && config.Broker != "" {
		return p.Brokers[config.Broker], nil
	}
	return nil, err
}
