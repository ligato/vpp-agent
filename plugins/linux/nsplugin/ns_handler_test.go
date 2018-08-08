// Copyright (c) 2018 Cisco and/or its affiliates.
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

package nsplugin_test

import (
	"testing"

	"net"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/ligato/vpp-agent/tests/linuxmock"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"
	"github.com/ligato/cn-infra/utils/safeclose"
)

/* Linux namespace handler init and close */

// Test init function
func TestNsHandlerInit(t *testing.T) {
	plugin, ifHandler, sysHandler, msChan, ifNotif := nsHandlerTestSetup(t)
	defer nsHandlerTestTeardown(plugin, ifHandler, sysHandler, msChan, ifNotif)

	// Base fields
	Expect(plugin).ToNot(BeNil())
	Expect(plugin.GetMicroserviceByLabel()).ToNot(BeNil())
	Expect(plugin.GetMicroserviceByLabel()).To(HaveLen(0))
	Expect(plugin.GetMicroserviceByID()).ToNot(BeNil())
	Expect(plugin.GetMicroserviceByID()).To(HaveLen(0))

	// todo test microservice tracker
}

/* Namespace handler Test Setup */

func TestSetInterfaceNamespace(t *testing.T) {
	plugin, ifHandler, sysHandler, msChan, ifNotif := nsHandlerTestSetup(t)
	defer nsHandlerTestTeardown(plugin, ifHandler, sysHandler, msChan, ifNotif)

	// IP address list
	var ipAddresses []netlink.Addr
	ipAddresses = append(ipAddresses,
		netlink.Addr{IPNet: getIPNetAddress("10.0.0.1/24")},
		netlink.Addr{IPNet: getIPNetAddress("172.168.0.1/24")},
		netlink.Addr{IPNet: getIPNetAddress("192.168.0.1/24")},
		// Link local address which should be skipped
		netlink.Addr{IPNet: getIPNetAddress("fe80::883f:c3ff:fe9e:fba/64")})
	ifHandler.When("GetLinkByName").ThenReturn(&netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{
			Name:  "if1",
			Flags: net.FlagUp,
		},
	})
	ifHandler.When("GetAddressList").ThenReturn(ipAddresses)
	sysHandler.When("LinkSetNsFd").ThenReturn()

	// Context and namespace
	ctx := nsplugin.NewNamespaceMgmtCtx()
	ns := &interfaces.LinuxInterfaces_Interface_Namespace{
		Type: interfaces.LinuxInterfaces_Interface_Namespace_NAMED_NS,
	}

	err := plugin.SetInterfaceNamespace(ctx, "if1", ns)
	Expect(err).To(BeNil())

	// Check calls to ensure that only required IP addresses were configured
	num, calls := ifHandler.GetCallsFor("AddInterfaceIP")
	Expect(num).To(Equal(3))
	Expect(calls).ToNot(BeNil())
	for callIdx, call := range calls {
		ifName := call[0].(string)
		Expect(ifName).To(Equal("if1"))
		ipAdd := call[1].(*net.IPNet)
		if callIdx == 1 {
			Expect(ipAdd.String()).To(Equal("10.0.0.1/24"))
		}
		if callIdx == 2 {
			Expect(ipAdd.String()).To(Equal("172.168.0.1/24"))
		}
		if callIdx == 3 {
			Expect(ipAdd.String()).To(Equal("192.168.0.1/24"))
		}
	}
}

func TestSetInterfaceNamespaceIPv6(t *testing.T) {
	plugin, ifHandler, sysHandler, msChan, ifNotif := nsHandlerTestSetup(t)
	defer nsHandlerTestTeardown(plugin, ifHandler, sysHandler, msChan, ifNotif)

	// IP address list
	var ipAddresses []netlink.Addr
	ipAddresses = append(ipAddresses,
		netlink.Addr{IPNet: getIPNetAddress("10.0.0.1/24")},
		netlink.Addr{IPNet: getIPNetAddress("172.168.0.1/24")},
		// Link local address should not be skipped if there is another non-link-local IPv6
		netlink.Addr{IPNet: getIPNetAddress("fe80::883f:c3ff:fe9e:fba/64")},
		netlink.Addr{IPNet: getIPNetAddress("ad48::42:e8ff:feb1:e976/64")})
	ifHandler.When("GetLinkByName").ThenReturn(&netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{
			Name:  "if1",
			Flags: net.FlagUp,
		},
	})
	ifHandler.When("GetAddressList").ThenReturn(ipAddresses)
	sysHandler.When("LinkSetNsFd").ThenReturn()

	// Context and namespace
	ctx := nsplugin.NewNamespaceMgmtCtx()
	ns := &interfaces.LinuxInterfaces_Interface_Namespace{
		Type: interfaces.LinuxInterfaces_Interface_Namespace_NAMED_NS,
	}

	err := plugin.SetInterfaceNamespace(ctx, "if1", ns)
	Expect(err).To(BeNil())

	// Check calls to ensure that only required IP addresses were configured
	num, calls := ifHandler.GetCallsFor("AddInterfaceIP")
	Expect(num).To(Equal(4))
	Expect(calls).ToNot(BeNil())
	for callIdx, call := range calls {
		ifName := call[0].(string)
		Expect(ifName).To(Equal("if1"))
		ipAdd := call[1].(*net.IPNet)
		if callIdx == 1 {
			Expect(ipAdd.String()).To(Equal("10.0.0.1/24"))
		}
		if callIdx == 2 {
			Expect(ipAdd.String()).To(Equal("172.168.0.1/24"))
		}
		if callIdx == 3 {
			Expect(ipAdd.String()).To(Equal("fe80::883f:c3ff:fe9e:fba/64"))
		}
		if callIdx == 4 {
			Expect(ipAdd.String()).To(Equal("ad48::42:e8ff:feb1:e976/64"))
		}
	}
}

func nsHandlerTestSetup(t *testing.T) (*nsplugin.NsHandler, *linuxmock.IfNetlinkHandlerMock, *linuxmock.SystemMock,
	chan *nsplugin.MicroserviceCtx, chan *nsplugin.MicroserviceEvent) {
	RegisterTestingT(t)

	// Loggers
	pluginLog := logging.ForPlugin("linux-ns-handler-log")
	pluginLog.SetLevel(logging.DebugLevel)
	// Handlers
	ifHandler := linuxmock.NewIfNetlinkHandlerMock()
	sysHandler := linuxmock.NewSystemMock()
	// Channels
	msChan := make(chan *nsplugin.MicroserviceCtx)
	ifNotif := make(chan *nsplugin.MicroserviceEvent)
	// Configurator
	plugin := &nsplugin.NsHandler{}
	err := plugin.Init(pluginLog, ifHandler, sysHandler, msChan, ifNotif)
	Expect(err).To(BeNil())

	return plugin, ifHandler, sysHandler, msChan, ifNotif
}

func nsHandlerTestTeardown(plugin *nsplugin.NsHandler, ifHnadler *linuxmock.IfNetlinkHandlerMock, sysHnadler *linuxmock.SystemMock,
	msChan chan *nsplugin.MicroserviceCtx, msEventChan chan *nsplugin.MicroserviceEvent) {
	Expect(plugin.Close()).To(Succeed())
	err := safeclose.Close(ifHnadler, sysHnadler, msChan, msEventChan)
	Expect(err).To(BeNil())
	logging.DefaultRegistry.ClearRegistry()
}

func getIPNetAddress(ipAddr string) *net.IPNet {
	ip, ipNet, err := net.ParseCIDR(ipAddr)
	ipNet.IP = ip
	Expect(err).To(BeNil())
	return ipNet
}
