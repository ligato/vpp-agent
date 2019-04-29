package configurator

import (
	"github.com/go-errors/errors"
	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"

	rpc "github.com/ligato/vpp-agent/api/configurator"
	vpp_acl "github.com/ligato/vpp-agent/api/models/vpp/acl"
	vpp_interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	vpp_ipsec "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	vpp_l2 "github.com/ligato/vpp-agent/api/models/vpp/l2"
	vpp_l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	vpp_punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	iflinuxcalls "github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	l3linuxcalls "github.com/ligato/vpp-agent/plugins/linux/l3plugin/linuxcalls"
	aclvppcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	ipsecvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
	l2vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	l3vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	natvppcalls "github.com/ligato/vpp-agent/plugins/vpp/natplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

type dumpService struct {
	log logging.Logger

	// VPP Handlers
	aclHandler   aclvppcalls.ACLVppRead
	ifHandler    ifvppcalls.InterfaceVppRead
	natHandler   natvppcalls.NatVppRead
	l2Handler    l2vppcalls.L2VppAPI
	l3Handler    l3vppcalls.L3VppAPI
	ipsecHandler ipsecvppcalls.IPSecVPPRead
	puntHandler  vppcalls.PuntVPPRead
	// Linux handlers
	linuxIfHandler iflinuxcalls.NetlinkAPIRead
	linuxL3Handler l3linuxcalls.NetlinkAPIRead
}

func (svc *dumpService) Dump(context.Context, *rpc.DumpRequest) (*rpc.DumpResponse, error) {
	defer trackOperation("Dump")()

	svc.log.Debugf("Received Dump request..")

	dump := newConfig()

	var err error

	dump.VppConfig.Interfaces, err = svc.DumpInterfaces()
	if err != nil {
		svc.log.Errorf("DumpInterfaces failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Acls, err = svc.DumpAcls()
	if err != nil {
		svc.log.Errorf("DumpAcls failed: %v", err)
		return nil, err
	}
	dump.VppConfig.IpsecSpds, err = svc.DumpIPSecSPDs()
	if err != nil {
		svc.log.Errorf("DumpIPSecSPDs failed: %v", err)
		return nil, err
	}
	dump.VppConfig.IpsecSas, err = svc.DumpIPSecSAs()
	if err != nil {
		svc.log.Errorf("DumpIPSecSAs failed: %v", err)
		return nil, err
	}
	dump.VppConfig.BridgeDomains, err = svc.DumpBDs()
	if err != nil {
		svc.log.Errorf("DumpBDs failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Routes, err = svc.DumpRoutes()
	if err != nil {
		svc.log.Errorf("DumpRoutes failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Arps, err = svc.DumpARPs()
	if err != nil {
		svc.log.Errorf("DumpARPs failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Fibs, err = svc.DumpFIBs()
	if err != nil {
		svc.log.Errorf("DumpFIBs failed: %v", err)
		return nil, err
	}
	dump.VppConfig.XconnectPairs, err = svc.DumpXConnects()
	if err != nil {
		svc.log.Errorf("DumpXConnects failed: %v", err)
		return nil, err
	}
	dump.VppConfig.PuntTohosts, err = svc.DumpPunt()
	if err != nil {
		svc.log.Errorf("DumpPunt failed: %v", err)
		return nil, err
	}

	// FIXME: linux interface handler should return known proto instead of netlink
	// state.LinuxData.Interfaces, _ = svc.DumpLinuxInterfaces()

	return &rpc.DumpResponse{Dump: dump}, nil
}

// DumpAcls reads IP/MACIP access lists and returns them as an *AclResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpAcls() ([]*vpp_acl.ACL, error) {
	var acls []*vpp_acl.ACL
	if svc.aclHandler == nil {
		return nil, errors.New("aclHandler is not available")
	}
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
	if svc.ipsecHandler == nil {
		return nil, errors.New("ipsecHandler is not available")
	}
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
	if svc.ipsecHandler == nil {
		return nil, errors.New("ipsecHandler is not available")
	}
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
	bdDetails, err := svc.l2Handler.DumpBridgeDomains()
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
	fibDetails, err := svc.l2Handler.DumpL2FIBs()
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
	xcDetails, err := svc.l2Handler.DumpXConnectPairs()
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
	rtDetails, err := svc.l3Handler.DumpRoutes()
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
	arpDetails, err := svc.l3Handler.DumpArpEntries()
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
	if svc.puntHandler == nil {
		return nil, errors.New("puntHandler is not available")
	}
	dump, err := svc.puntHandler.DumpRegisteredPuntSockets()
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
