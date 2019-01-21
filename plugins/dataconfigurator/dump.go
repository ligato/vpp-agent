package dataconfigurator

import (
	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"

	rpc "github.com/ligato/vpp-agent/api/dataconfigurator"
	"github.com/ligato/vpp-agent/api/models/vpp/acl"
	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	"github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	"github.com/ligato/vpp-agent/api/models/vpp/l2"
	"github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/api/models/vpp/punt"
	iflinuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/linuxcalls"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/l3plugin/linuxcalls"
	aclvppcalls "github.com/ligato/vpp-agent/plugins/vppv2/aclplugin/vppcalls"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/vppcalls"
	ipsecvppcalls "github.com/ligato/vpp-agent/plugins/vppv2/ipsecplugin/vppcalls"
	l2vppcalls "github.com/ligato/vpp-agent/plugins/vppv2/l2plugin/vppcalls"
	l3vppcalls "github.com/ligato/vpp-agent/plugins/vppv2/l3plugin/vppcalls"
	natvppcalls "github.com/ligato/vpp-agent/plugins/vppv2/natplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/puntplugin/vppcalls"
)

type dumpService struct {
	log logging.Logger

	// VPP Handlers
	aclHandler   aclvppcalls.ACLVppRead
	ifHandler    ifvppcalls.IfVppRead
	natHandler   natvppcalls.NatVppRead
	bdHandler    l2vppcalls.BridgeDomainVppRead
	fibHandler   l2vppcalls.FIBVppRead
	xcHandler    l2vppcalls.XConnectVppRead
	arpHandler   l3vppcalls.ArpVppRead
	pArpHandler  l3vppcalls.ProxyArpVppRead
	rtHandler    l3vppcalls.RouteVppRead
	ipsecHandler ipsecvppcalls.IPSecVPPRead
	puntHandler  vppcalls.PuntVPPRead
	// Linux handlers
	linuxIfHandler iflinuxcalls.NetlinkAPIRead
	linuxL3Handler l3linuxcalls.NetlinkAPIRead
}

func (svc *dumpService) Dump(context.Context, *rpc.DumpRequest) (*rpc.DumpResponse, error) {
	state := newData()

	state.VppData.Interfaces, _ = svc.DumpInterfaces()
	state.VppData.Acls, _ = svc.DumpAcls()
	state.VppData.IpsecSpds, _ = svc.DumpIPSecSPDs()
	state.VppData.IpsecSas, _ = svc.DumpIPSecSAs()
	state.VppData.BridgeDomains, _ = svc.DumpBDs()
	state.VppData.Routes, _ = svc.DumpRoutes()
	state.VppData.Arps, _ = svc.DumpARPs()
	state.VppData.Fibs, _ = svc.DumpFIBs()
	state.VppData.XconnectPairs, _ = svc.DumpXConnects()
	state.VppData.PuntTohosts, _ = svc.DumpPunt()

	// FIXME: linux interface handler should return known proto instead of netlink
	// state.LinuxData.Interfaces, _ = svc.DumpLinuxInterfaces()

	return &rpc.DumpResponse{State: state}, nil
}

// DumpAcls reads IP/MACIP access lists and returns them as an *AclResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpAcls() ([]*vpp_acl.Acl, error) {
	var acls []*vpp_acl.Acl
	ipACLs, err := svc.aclHandler.DumpACL()
	if err != nil {
		return nil, err
	}
	macIPACLs, err := svc.aclHandler.DumpMACIPACL()
	if err != nil {
		return nil, err
	}
	for _, aclDetails := range ipACLs {
		acls = append(acls, aclDetails.ACL)
	}
	for _, aclDetails := range macIPACLs {
		acls = append(acls, aclDetails.ACL)
	}

	return acls, nil
}

// DumpInterfaces reads interfaces and returns them as an *InterfaceResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpInterfaces() ([]*vpp_interfaces.Interface, error) {
	var ifs []*vpp_interfaces.Interface
	ifDetails, err := svc.ifHandler.DumpInterfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifDetails {
		ifs = append(ifs, iface.Interface)
	}

	return ifs, nil
}

// DumpIPSecSPDs reads IPSec SPD and returns them as an *IPSecSPDResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpIPSecSPDs() ([]*vpp_ipsec.SecurityPolicyDatabase, error) {
	var spds []*vpp_ipsec.SecurityPolicyDatabase
	spdDetails, err := svc.ipsecHandler.DumpIPSecSPD()
	if err != nil {
		return nil, err
	}
	for _, spd := range spdDetails {
		spds = append(spds, spd.Spd)
	}

	return spds, nil
}

// DumpIPSecSAs reads IPSec SA and returns them as an *IPSecSAResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpIPSecSAs() ([]*vpp_ipsec.SecurityAssociation, error) {
	var sas []*vpp_ipsec.SecurityAssociation
	saDetails, err := svc.ipsecHandler.DumpIPSecSA()
	if err != nil {
		return nil, err
	}
	for _, sa := range saDetails {
		sas = append(sas, sa.Sa)
	}

	return sas, nil
}

// DumpIPSecTunnels reads IPSec tunnels and returns them as an *IPSecTunnelResponse. If reading ends up with error,
// only error is send back in response
/*func (svc *dumpService) DumpIPSecTunnels() (*rpc.IPSecTunnelResponse, error) {
	var tuns []*vpp_ipsec.
	tunDetails, err := svc.ipSecHandler.DumpIPSecTunnelInterfaces()
	if err != nil {
		return nil, err
	}
	for _, tun := range tunDetails {
		tuns = append(tuns, tun.Tunnel)
	}

	return &rpc.IPSecTunnelResponse{Tunnels: tuns}, nil
}*/

// DumpBDs reads bridge domains and returns them as an *BDResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpBDs() ([]*vpp_l2.BridgeDomain, error) {
	var bds []*vpp_l2.BridgeDomain
	bdDetails, err := svc.bdHandler.DumpBridgeDomains()
	if err != nil {
		return nil, err
	}
	for _, bd := range bdDetails {
		bds = append(bds, bd.Bd)
	}

	return bds, nil
}

// DumpFIBs reads FIBs and returns them as an *FibResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpFIBs() ([]*vpp_l2.FIBEntry, error) {
	var fibs []*vpp_l2.FIBEntry
	fibDetails, err := svc.fibHandler.DumpL2FIBs()
	if err != nil {
		return nil, err
	}
	for _, fib := range fibDetails {
		fibs = append(fibs, fib.Fib)
	}

	return fibs, nil
}

// DumpXConnects reads cross connects and returns them as an *XcResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpXConnects() ([]*vpp_l2.XConnectPair, error) {
	var xcs []*vpp_l2.XConnectPair
	xcDetails, err := svc.xcHandler.DumpXConnectPairs()
	if err != nil {
		return nil, err
	}
	for _, xc := range xcDetails {
		xcs = append(xcs, xc.Xc)
	}

	return xcs, nil
}

// DumpRoutes reads VPP routes and returns them as an *RoutesResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpRoutes() ([]*vpp_l3.Route, error) {
	var routes []*vpp_l3.Route
	rtDetails, err := svc.rtHandler.DumpRoutes()
	if err != nil {
		return nil, err
	}
	for _, rt := range rtDetails {
		routes = append(routes, rt.Route)
	}

	return routes, nil
}

// DumpARPs reads VPP ARPs and returns them as an *ARPsResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpARPs() ([]*vpp_l3.ARPEntry, error) {
	var arps []*vpp_l3.ARPEntry
	arpDetails, err := svc.arpHandler.DumpArpEntries()
	if err != nil {
		return nil, err
	}
	for _, arp := range arpDetails {
		arps = append(arps, arp.Arp)
	}

	return arps, nil
}

// DumpPunt reads VPP Punt socket registrations and returns them as an *PuntResponse.
func (svc *dumpService) DumpPunt() (punts []*vpp_punt.ToHost, err error) {
	dump, err := svc.puntHandler.DumpPuntRegisteredSockets()
	if err != nil {
		return nil, err
	}
	for _, puntDetails := range dump {
		punts = append(punts, puntDetails.PuntData)
	}

	return punts, nil
}

// DumpLinuxInterfaces reads linux interfaces and returns them as an *LinuxInterfaceResponse. If reading ends up with error,
// only error is send back in response
/*func (svc *dumpService) DumpLinuxInterfaces() ([]*linux_interfaces.Interface, error) {
	var linuxIfs []*linux_interfaces.Interface
	ifDetails, err := svc.linuxIfHandler.GetLinkList()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifDetails {
		linuxIfs = append(linuxIfs, )
	}

	return linuxIfs, nil
}

// DumpLinuxARPs reads linux ARPs and returns them as an *LinuxARPsResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpLinuxARPs(ctx context.Context, request *rpc.DumpRequest) (*rpc.LinuxARPsResponse, error) {
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
func (svc *dumpService) DumpLinuxRoutes(ctx context.Context, request *rpc.DumpRequest) (*rpc.LinuxRoutesResponse, error) {
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
*/
