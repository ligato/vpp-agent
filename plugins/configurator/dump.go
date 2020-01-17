//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package configurator

import (
	"context"
	"errors"

	"github.com/ligato/cn-infra/logging"

	iflinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	l3linuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/l3plugin/linuxcalls"
	abfvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/abfplugin/vppcalls"
	aclvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/aclplugin/vppcalls"
	ifvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	ipsecvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ipsecplugin/vppcalls"
	l2vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	l3vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	natvppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/natplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/puntplugin/vppcalls"
	rpc "go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/linux/l3"
	vpp_abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
	vpp_acl "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/acl"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipsec"
	vpp_l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	vpp_nat "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/nat"
	vpp_punt "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/punt"
)

type dumpService struct {
	log logging.Logger

	// VPP Handlers

	// core
	ifHandler    ifvppcalls.InterfaceVppRead
	l2Handler    l2vppcalls.L2VppAPI
	l3Handler    l3vppcalls.L3VppAPI
	ipsecHandler ipsecvppcalls.IPSecVPPRead
	// plugins
	aclHandler  aclvppcalls.ACLVppRead
	abfHandler  abfvppcalls.ABFVppRead
	natHandler  natvppcalls.NatVppRead
	puntHandler vppcalls.PuntVPPRead

	// Linux handlers
	linuxIfHandler iflinuxcalls.NetlinkAPIRead
	linuxL3Handler l3linuxcalls.NetlinkAPIRead
}

// Dump implements Dump method for Configurator
func (svc *dumpService) Dump(ctx context.Context, req *rpc.DumpRequest) (*rpc.DumpResponse, error) {
	defer trackOperation("Dump")()

	svc.log.Debugf("Received Dump request..")

	dump := newConfig()

	var err error

	// core
	dump.VppConfig.Interfaces, err = svc.DumpInterfaces(ctx)
	if err != nil {
		svc.log.Errorf("DumpInterfaces failed: %v", err)
		return nil, err
	}
	dump.VppConfig.BridgeDomains, err = svc.DumpBDs()
	if err != nil {
		svc.log.Errorf("DumpBDs failed: %v", err)
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

	// plugins
	dump.VppConfig.Acls, err = svc.DumpACLs()
	if err != nil {
		svc.log.Errorf("DumpACLs failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Abfs, err = svc.DumpABFs()
	if err != nil {
		svc.log.Errorf("DumpABFs failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Nat44Global, err = svc.DumpNAT44Global()
	if err != nil {
		svc.log.Errorf("DumpNAT44Global failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Dnat44S, err = svc.DumpDNAT44s()
	if err != nil {
		svc.log.Errorf("DumpDNAT44s failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Nat44Interfaces, err = svc.DumpNAT44Interfaces()
	if err != nil {
		svc.log.Errorf("DumpNAT44Interfaces failed: %v", err)
		return nil, err
	}
	dump.VppConfig.Nat44Pools, err = svc.DumpNAT44AddressPools()
	if err != nil {
		svc.log.Errorf("DumpNAT44AddressPools failed: %v", err)
		return nil, err
	}
	dump.VppConfig.PuntTohosts, err = svc.DumpPunt()
	if err != nil {
		svc.log.Errorf("DumpPunt failed: %v", err)
		return nil, err
	}
	dump.VppConfig.PuntExceptions, err = svc.DumpPuntExceptions()
	if err != nil {
		svc.log.Errorf("DumpPuntExceptions failed: %v", err)
		return nil, err
	}

	dump.LinuxConfig.Interfaces, err = svc.DumpLinuxInterfaces()
	if err != nil {
		svc.log.Errorf("DumpLinuxInterfaces failed: %v", err)
		return nil, err
	}

	dump.LinuxConfig.ArpEntries, err = svc.DumpLinuxARPs()
	if err != nil {
		svc.log.Errorf("DumpLinuxARPs failed: %v", err)
		return nil, err
	}

	dump.LinuxConfig.Routes, err = svc.DumpLinuxRoutes()
	if err != nil {
		svc.log.Errorf("DumpLinuxRoutes failed: %v", err)
		return nil, err
	}

	return &rpc.DumpResponse{Dump: dump}, nil
}

// DumpInterfaces reads interfaces and returns them as an *InterfaceResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpInterfaces(ctx context.Context) (ifs []*vpp_interfaces.Interface, err error) {
	if svc.ifHandler == nil {
		// handler is not available
		return nil, nil
	}

	ifDetails, err := svc.ifHandler.DumpInterfaces(ctx)
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
func (svc *dumpService) DumpIPSecSPDs() (spds []*vpp_ipsec.SecurityPolicyDatabase, err error) {
	if svc.ipsecHandler == nil {
		// handler is not available
		return nil, nil
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
func (svc *dumpService) DumpIPSecSAs() (sas []*vpp_ipsec.SecurityAssociation, err error) {
	if svc.ipsecHandler == nil {
		// handler is not available
		return nil, nil
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

// DumpBDs reads bridge domains and returns them as an *BDResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpBDs() (bds []*vpp_l2.BridgeDomain, err error) {
	if svc.l2Handler == nil {
		// handler is not available
		return nil, nil
	}

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
func (svc *dumpService) DumpFIBs() (fibs []*vpp_l2.FIBEntry, err error) {
	if svc.l2Handler == nil {
		// handler is not available
		return nil, nil
	}

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
func (svc *dumpService) DumpXConnects() (xcs []*vpp_l2.XConnectPair, err error) {
	if svc.l2Handler == nil {
		// handler is not available
		return nil, nil
	}

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
func (svc *dumpService) DumpRoutes() (routes []*vpp_l3.Route, err error) {
	if svc.l3Handler == nil {
		// handler is not available
		return nil, nil
	}

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
func (svc *dumpService) DumpARPs() (arps []*vpp_l3.ARPEntry, err error) {
	if svc.l3Handler == nil {
		// handler is not available
		return nil, nil
	}

	arpDetails, err := svc.l3Handler.DumpArpEntries()
	if err != nil {
		return nil, err
	}
	for _, arp := range arpDetails {
		arps = append(arps, arp.Arp)
	}
	return arps, nil
}

// DumpACLs reads IP/MACIP access lists and returns them as an *AclResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpACLs() (acls []*vpp_acl.ACL, err error) {
	if svc.aclHandler == nil {
		// handler is not available
		return nil, nil
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

// DumpABFs reads the ACL-based forwarding and returns data read as an *AbfResponse. If the reading ends up with
// an error, only the error is send back in the response
func (svc *dumpService) DumpABFs() (abfs []*vpp_abf.ABF, err error) {
	if svc.abfHandler == nil {
		// handler is not available
		return nil, nil
	}

	abfPolicy, err := svc.abfHandler.DumpABFPolicy()
	if err != nil {
		return nil, err
	}
	for _, abfDetails := range abfPolicy {
		abfs = append(abfs, abfDetails.ABF)
	}
	return abfs, nil
}

// DumpNAT44Global dumps NAT44Global
func (svc *dumpService) DumpNAT44Global() (glob *vpp_nat.Nat44Global, err error) {
	if svc.natHandler == nil {
		// handler is not available
		return nil, nil
	}

	glob, err = svc.natHandler.Nat44GlobalConfigDump(false)
	if err != nil {
		return nil, err
	}
	return glob, nil
}

// DumpDNAT44s dumps DNat44
func (svc *dumpService) DumpDNAT44s() (dnats []*vpp_nat.DNat44, err error) {
	if svc.natHandler == nil {
		// handler is not available
		return nil, nil
	}

	dnats, err = svc.natHandler.DNat44Dump()
	if err != nil {
		return nil, err
	}
	return dnats, nil
}

// DumpNAT44Interfaces dumps NAT44Interfaces
func (svc *dumpService) DumpNAT44Interfaces() (natIfs []*vpp_nat.Nat44Interface, err error) {
	if svc.natHandler == nil {
		// handler is not available
		return nil, nil
	}

	natIfs, err = svc.natHandler.Nat44InterfacesDump()
	if err != nil {
		return nil, err
	}
	return natIfs, nil
}

// DumpNAT44AddressPools dumps NAT44AddressPools
func (svc *dumpService) DumpNAT44AddressPools() (natPools []*vpp_nat.Nat44AddressPool, err error) {
	if svc.natHandler == nil {
		// handler is not available
		return nil, nil
	}

	natPools, err = svc.natHandler.Nat44AddressPoolsDump()
	if err != nil {
		return nil, err
	}
	return natPools, nil
}

// DumpPunt reads VPP Punt socket registrations and returns them as an *PuntResponse.
func (svc *dumpService) DumpPunt() (punts []*vpp_punt.ToHost, err error) {
	if svc.puntHandler == nil {
		// handler is not available
		return nil, nil
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

// DumpPuntExceptions reads VPP Punt exceptions and returns them as an *PuntResponse.
func (svc *dumpService) DumpPuntExceptions() (punts []*vpp_punt.Exception, err error) {
	if svc.puntHandler == nil {
		return nil, errors.New("puntHandler is not available")
	}
	dump, err := svc.puntHandler.DumpExceptions()
	if err != nil {
		return nil, err
	}
	for _, puntDetails := range dump {
		punts = append(punts, puntDetails.Exception)
	}

	return punts, nil
}

// DumpLinuxInterfaces reads linux interfaces and returns them as an *LinuxInterfaceResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpLinuxInterfaces() (linuxIfs []*linux_interfaces.Interface, err error) {
	if svc.linuxIfHandler == nil {
		return nil, errors.New("linuxIfHandler is not available")
	}

	ifDetails, err := svc.linuxIfHandler.DumpInterfaces()
	if err != nil {
		return nil, err
	}
	for _, ifDetail := range ifDetails {
		linuxIfs = append(linuxIfs, ifDetail.Interface)
	}

	return linuxIfs, nil
}

// DumpLinuxARPs reads linux ARPs and returns them as an *LinuxARPsResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpLinuxARPs() (linuxARPs []*linux_l3.ARPEntry, err error) {
	if svc.linuxL3Handler == nil {
		return nil, errors.New("linuxL3Handler is not available")
	}

	arpDetails, err := svc.linuxL3Handler.DumpARPEntries()
	if err != nil {
		return nil, err
	}
	for _, arpDetail := range arpDetails {
		linuxARPs = append(linuxARPs, arpDetail.ARP)
	}

	return linuxARPs, nil
}

// DumpLinuxRoutes reads linux routes and returns them as an *LinuxRoutesResponse. If reading ends up with error,
// only error is send back in response
func (svc *dumpService) DumpLinuxRoutes() (linuxRoutes []*linux_l3.Route, err error) {
	if svc.linuxL3Handler == nil {
		return nil, errors.New("linuxL3Handler is not available")
	}

	rtDetails, err := svc.linuxL3Handler.DumpRoutes()
	if err != nil {
		return nil, err
	}
	for _, rt := range rtDetails {
		linuxRoutes = append(linuxRoutes, rt.Route)
	}

	return linuxRoutes, nil
}
