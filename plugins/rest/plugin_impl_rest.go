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

package rest

import (
	"fmt"
	"sync"

	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/rest/resturl"
	"github.com/ligato/vpp-agent/plugins/vpp"
	aclvppcalls "github.com/ligato/vpp-agent/plugins/vpp/aclplugin/vppcalls"
	ifvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	ipsecvppcalls "github.com/ligato/vpp-agent/plugins/vpp/ipsecplugin/vppcalls"
	l2vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l2plugin/vppcalls"
	l3vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	l4vppcalls "github.com/ligato/vpp-agent/plugins/vpp/l4plugin/vppcalls"
)

// REST api methods
const (
	GET  = "GET"
	POST = "POST"
)

// Plugin registers Rest Plugin
type Plugin struct {
	Deps

	// Partial index items
	aclIndexItems       []indexItem
	ifIndexItems        []indexItem
	ipSecIndexItems     []indexItem
	l2IndexItems        []indexItem
	l3IndexItems        []indexItem
	l4IndexItems        []indexItem
	telemetryIndexItems []indexItem
	commonIndexItems    []indexItem

	// Channels
	vppChan  api.Channel
	dumpChan api.Channel

	// Handlers
	aclHandler   aclvppcalls.AclVppRead
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

	sync.Mutex
}

// Deps represents dependencies of Rest Plugin
type Deps struct {
	infra.PluginDeps
	HTTPHandlers rest.HTTPHandlers
	GoVppmux     govppmux.API
	VPP          vpp.API
}

type indexItem struct {
	Name string
	Path string
}

// Init initializes the Rest Plugin
func (plugin *Plugin) Init() (err error) {
	// Check VPP dependency
	if plugin.VPP == nil {
		return fmt.Errorf("REST plugin requires VPP plugin API")
	}
	// VPP channels
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	if plugin.dumpChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return err
	}
	// Indexes
	ifIndexes := plugin.VPP.GetSwIfIndexes()
	bdIndexes := plugin.VPP.GetBDIndexes()
	spdIndexes := plugin.VPP.GetIPSecSPDIndexes()

	// Initialize handlers
	if plugin.aclHandler, err = aclvppcalls.NewAclVppHandler(plugin.vppChan, plugin.dumpChan, nil); err != nil {
		return err
	}
	if plugin.ifHandler, err = ifvppcalls.NewIfVppHandler(plugin.vppChan, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.bfdHandler, err = ifvppcalls.NewBfdVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.natHandler, err = ifvppcalls.NewNatVppHandler(plugin.vppChan, plugin.dumpChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.stnHandler, err = ifvppcalls.NewStnVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.ipSecHandler, err = ipsecvppcalls.NewIPsecVppHandler(plugin.vppChan, ifIndexes, spdIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.bdHandler, err = l2vppcalls.NewBridgeDomainVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.fibHandler, err = l2vppcalls.NewFibVppHandler(plugin.vppChan, plugin.dumpChan, make(chan *l2vppcalls.FibLogicalReq),
		ifIndexes, bdIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.xcHandler, err = l2vppcalls.NewXConnectVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.arpHandler, err = l3vppcalls.NewArpVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.pArpHandler, err = l3vppcalls.NewProxyArpVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.rtHandler, err = l3vppcalls.NewRouteVppHandler(plugin.vppChan, ifIndexes, plugin.Log, nil); err != nil {
		return err
	}
	if plugin.l4Handler, err = l4vppcalls.NewL4VppHandler(plugin.vppChan, plugin.Log, nil); err != nil {
		return err
	}

	// Fill index item lists
	plugin.aclIndexItems = []indexItem{
		{Name: "ACL IP", Path: resturl.AclIP},
		{Name: "ACL MACIP", Path: resturl.AclMACIP},
	}
	plugin.ifIndexItems = []indexItem{
		{Name: "Interfaces", Path: resturl.Interface},
		{Name: "Loopback interfaces", Path: resturl.Loopback},
		{Name: "Ethernet interfaces", Path: resturl.Ethernet},
		{Name: "Memif interfaces", Path: resturl.Memif},
		{Name: "Tap interfaces", Path: resturl.Tap},
		{Name: "VxLAN interfaces", Path: resturl.VxLan},
		{Name: "Af-packet nterfaces", Path: resturl.AfPacket},
	}
	plugin.ipSecIndexItems = []indexItem{
		{Name: "IPSec SPD", Path: resturl.IPSecSpd},
		{Name: "IPSec SA", Path: resturl.IPSecSa},
		{Name: "IPSec Tunnel interfaces", Path: resturl.IPSecTnIf},
	}
	plugin.l2IndexItems = []indexItem{
		{Name: "Bridge domains", Path: resturl.Bd},
		{Name: "Bridge domain IDs", Path: resturl.BdId},
		{Name: "L2Fibs", Path: resturl.Fib},
		{Name: "XConnectorPairs", Path: resturl.Xc},
	}
	plugin.l3IndexItems = []indexItem{
		{Name: "ARPs", Path: resturl.Arps},
		{Name: "Proxy ARPs", Path: resturl.ProxyArps},
		{Name: "Proxy ARP interfaces", Path: resturl.PArpIfs},
		{Name: "Proxy ARP ranges", Path: resturl.PArpRngs},
		{Name: "Static routes", Path: resturl.Routes},
	}
	plugin.l4IndexItems = []indexItem{
		{Name: "L4 sessions", Path: resturl.Sessions},
	}
	plugin.telemetryIndexItems = []indexItem{
		{Name: "Telemetry", Path: resturl.Telemetry},
		{Name: "Telemetry memory", Path: resturl.TMemory},
		{Name: "Telemetry runtime", Path: resturl.TRuntime},
		{Name: "Telemetry node count", Path: resturl.TNodeCount},
	}
	plugin.commonIndexItems = []indexItem{
		{Name: "CLI command", Path: resturl.Command},
		{Name: "Index page", Path: resturl.Index},
		{Name: "Index page for ACL plugin", Path: resturl.IndexAcl},
		{Name: "Index page for interface plugin", Path: resturl.IndexIf},
		{Name: "Index page for IPSec plugin", Path: resturl.IndexIPSec},
		{Name: "Index page for L2 plugin", Path: resturl.IndexL2},
		{Name: "Index page for L3 plugin", Path: resturl.IndexL3},
		{Name: "Index page for L4 plugin", Path: resturl.IndexL4},
		{Name: "Index page for telemetry", Path: resturl.IndexTel},
		{Name: "Index page for common commands", Path: resturl.IndexComm},
	}

	return nil
}

// AfterInit is used to register HTTP handlers
func (plugin *Plugin) AfterInit() (err error) {
	plugin.Log.Debug("REST API Plugin is up and running")

	plugin.registerAccessListHandlers()
	plugin.registerInterfaceHandlers()
	plugin.registerBfdHandlers()
	plugin.registerNatHandlers()
	plugin.registerStnHandlers()
	plugin.registerIPSecHandlers()
	plugin.registerL2Handlers()
	plugin.registerL3Handlers()
	plugin.registerL4Handlers()
	plugin.registerTelemetryHandlers()
	plugin.registerCommandHandler()
	plugin.registerIndexHandlers()

	return nil
}

// Close is used to clean up resources used by Plugin
func (plugin *Plugin) Close() (err error) {
	return safeclose.Close(plugin.vppChan, plugin.dumpChan)
}
