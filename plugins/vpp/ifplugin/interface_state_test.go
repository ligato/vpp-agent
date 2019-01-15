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

package ifplugin_test

import (
	"net"
	"testing"
	"time"

	"git.fd.io/govpp.git/adapter"

	govppmock "git.fd.io/govpp.git/adapter/mock"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux/mock"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	intf "github.com/ligato/vpp-agent/plugins/vpp/model/interfaces"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
)

func testPluginDataInitialization(t *testing.T) (*mock.GoVPPMux, ifaceidx.SwIfIndexRW, *ifplugin.InterfaceStateUpdater,
	chan govppapi.Message, chan *intf.InterfaceNotification) {
	RegisterTestingT(t)

	// Initialize notification channel
	notifChan := make(chan govppapi.Message, 100)

	// Initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "interface_state_test", ifaceidx.IndexMetadata)
	index := ifaceidx.NewSwIfIndex(nameToIdx)
	names := nameToIdx.ListNames()
	Expect(names).To(BeEmpty())

	// Create publish state function
	publishChan := make(chan *intf.InterfaceNotification, 100)
	publishIfState := func(notification *intf.InterfaceNotification) {
		t.Logf("Received notification change %v", notification)
		publishChan <- notification
	}

	// Create context
	ctx, _ := context.WithCancel(context.Background())

	// Create VPP connection
	mockCtx := &vppcallmock.TestCtx{
		MockVpp:   govppmock.NewVppAdapter(),
		MockStats: govppmock.NewStatsAdapter(),
	}

	goVppMux, err := mock.NewMockGoVPPMux(mockCtx)
	Expect(err).To(BeNil())

	// Prepare Init VPP replies
	mockCtx.MockVpp.MockReply(&interfaces.WantInterfaceEventsReply{})

	// Create plugin logger
	pluginLogger := logging.ForPlugin("testname")

	// Test initialization
	ifPlugin := &ifplugin.InterfaceStateUpdater{}
	err = ifPlugin.Init(ctx, pluginLogger, goVppMux, index, notifChan, publishIfState)
	Expect(err).To(BeNil())
	err = ifPlugin.AfterInit()
	Expect(err).To(BeNil())

	return goVppMux, index, ifPlugin, notifChan, publishChan
}

func testPluginDataTeardown(plugin *ifplugin.InterfaceStateUpdater, goVPPMux *mock.GoVPPMux) {
	goVPPMux.Close()
	Expect(plugin.Close()).To(BeNil())
	logging.DefaultRegistry.ClearRegistry()
}

// Test UPDOWN notification
func TestInterfaceStateUpdaterUpDownNotif(t *testing.T) {
	goVppMux, index, ifPlugin, notifChan, publishChan := testPluginDataInitialization(t)
	defer testPluginDataTeardown(ifPlugin, goVppMux)

	// Register name
	index.RegisterName("test", 0, &intf.Interfaces_Interface{
		Name:        "test",
		Enabled:     true,
		Type:        intf.InterfaceType_MEMORY_INTERFACE,
		IpAddresses: []string{"192.168.0.1/24"},
	})

	// Test notifications
	notifChan <- &interfaces.SwInterfaceEvent{
		PID:         0,
		SwIfIndex:   0,
		AdminUpDown: 1,
		LinkUpDown:  1,
		Deleted:     0,
	}

	var notif *intf.InterfaceNotification

	Eventually(publishChan).Should(Receive(&notif))
	Expect(notif.Type).To(Equal(intf.InterfaceNotification_UPDOWN))
	Expect(notif.State.AdminStatus).Should(BeEquivalentTo(intf.InterfacesState_Interface_UP))
}

// Test simple counter notification
func TestInterfaceStateUpdaterVnetSimpleCounterNotif(t *testing.T) {
	goVPPMux, index, ifPlugin, notifChan, publishChan := testPluginDataInitialization(t)
	defer testPluginDataTeardown(ifPlugin, goVPPMux)

	// Register name
	index.RegisterName("test", 0, &intf.Interfaces_Interface{
		Name:        "test",
		Enabled:     true,
		Type:        intf.InterfaceType_MEMORY_INTERFACE,
		IpAddresses: []string{"192.168.0.1/24"},
	})

	// Test stats
	var stats []*adapter.StatEntry
	stats = append(stats,
		&adapter.StatEntry{
			Name: "/if/drops",
			Type: adapter.StatType(2),
			Data: adapter.SimpleCounterStat{{32768}},
		},
		&adapter.StatEntry{
			Name: "/if/punt",
			Type: adapter.StatType(2),
			Data: adapter.SimpleCounterStat{{32769}},
		},
		&adapter.StatEntry{
			Name: "/if/ip4",
			Type: adapter.StatType(2),
			Data: adapter.SimpleCounterStat{{32770}},
		},
		&adapter.StatEntry{
			Name: "/if/ip6",
			Type: adapter.StatType(2),
			Data: adapter.SimpleCounterStat{{32771}},
		},
		&adapter.StatEntry{
			Name: "/if/rx-no-buf",
			Type: adapter.StatType(2),
			Data: adapter.SimpleCounterStat{{32772}},
		},
		&adapter.StatEntry{
			Name: "/if/rx-miss",
			Type: adapter.StatType(2),
			Data: adapter.SimpleCounterStat{{32773}},
		},
		&adapter.StatEntry{
			Name: "/if/rx-error",
			Type: adapter.StatType(2),
			Data: adapter.SimpleCounterStat{{32774}},
		},
		&adapter.StatEntry{
			Name: "/if/tx-error",
			Type: adapter.StatType(2),
			Data: adapter.SimpleCounterStat{{32775}},
		})
	err := goVPPMux.MockStats(stats)
	Expect(err).To(BeNil())

	// Send interface event notification to propagate update from counter to publish channel
	notifChan <- &interfaces.SwInterfaceEvent{
		PID:         0,
		SwIfIndex:   0,
		AdminUpDown: 1,
		LinkUpDown:  1,
		Deleted:     0,
	}

	// Stats are read periodically every second, let's give it some time to update the notification
	time.Sleep(2* time.Second)

	var notif *intf.InterfaceNotification
	Eventually(publishChan).Should(Receive(&notif))
	Expect(notif.Type).To(Equal(intf.InterfaceNotification_UPDOWN))
	Expect(notif.State.AdminStatus).Should(BeEquivalentTo(intf.InterfacesState_Interface_UP))
	Expect(notif.State.Statistics.DropPackets).Should(BeEquivalentTo(32768))
	Expect(notif.State.Statistics.PuntPackets).Should(BeEquivalentTo(32769))
	Expect(notif.State.Statistics.Ipv4Packets).Should(BeEquivalentTo(32770))
	Expect(notif.State.Statistics.Ipv6Packets).Should(BeEquivalentTo(32771))
	Expect(notif.State.Statistics.InNobufPackets).Should(BeEquivalentTo(32772))
	Expect(notif.State.Statistics.InMissPackets).Should(BeEquivalentTo(32773))
	Expect(notif.State.Statistics.InErrorPackets).Should(BeEquivalentTo(32774))
	Expect(notif.State.Statistics.OutErrorPackets).Should(BeEquivalentTo(32775))
}

// Test VnetIntCombined notification
func TestInterfaceStateUpdaterVnetIntCombinedNotif(t *testing.T) {
	goVPPMux, index, ifPlugin, notifChan, publishChan := testPluginDataInitialization(t)
	defer testPluginDataTeardown(ifPlugin, goVPPMux)

	// Register name
	index.RegisterName("test0", 0, &intf.Interfaces_Interface{
		Name:        "test0",
		Enabled:     true,
		Type:        intf.InterfaceType_MEMORY_INTERFACE,
		IpAddresses: []string{"192.168.0.1/24"},
	})

	index.RegisterName("test1", 1, &intf.Interfaces_Interface{
		Name:        "test1",
		Enabled:     true,
		Type:        intf.InterfaceType_MEMORY_INTERFACE,
		IpAddresses: []string{"192.168.0.2/24"},
	})

	// Test stats
	var stats []*adapter.StatEntry
	stats = append(stats,
		&adapter.StatEntry{
			Name: "/if/tx",
			Type: adapter.StatType(3),
			Data: adapter.CombinedCounterStat{{adapter.CombinedCounter{
				Packets: 3000,
				Bytes:   8000,
			}}},
		})
	err := goVPPMux.MockStats(stats)
	Expect(err).To(BeNil())

	// Send interface event notification to propagate update from counter to publish channel
	notifChan <- &interfaces.SwInterfaceEvent{
		PID:         0,
		SwIfIndex:   0,
		AdminUpDown: 1,
		LinkUpDown:  1,
		Deleted:     0,
	}

	// Stats are read periodically every second, let's give it some time to update the notification
	time.Sleep(2 * time.Second)

	var notif *intf.InterfaceNotification

	Eventually(publishChan).Should(Receive(&notif))
	Expect(notif.Type).To(Equal(intf.InterfaceNotification_UPDOWN))
	Expect(notif.State.Statistics.OutPackets).Should(BeEquivalentTo(3000))
	Expect(notif.State.Statistics.OutBytes).Should(BeEquivalentTo(8000))
}

// Test SwInterfaceDetails notification
func TestInterfaceStateUpdaterSwInterfaceDetailsNotif(t *testing.T) {
	goVPPMux, index, ifPlugin, notifChan, publishChan := testPluginDataInitialization(t)
	defer testPluginDataTeardown(ifPlugin, goVPPMux)

	// Register name
	index.RegisterName("test", 0, &intf.Interfaces_Interface{
		Name:        "test",
		Enabled:     true,
		Type:        intf.InterfaceType_MEMORY_INTERFACE,
		IpAddresses: []string{"192.168.0.1/24"},
	})

	// Test notifications
	hwAddr1Parse, err := net.ParseMAC("01:23:45:67:89:ab")
	Expect(err).To(BeNil())

	notifChan <- &interfaces.SwInterfaceDetails{
		InterfaceName:   []byte("if0"),
		AdminUpDown:     1,    // adm up
		LinkUpDown:      0,    // oper down
		LinkMtu:         9216, // Default MTU
		L2Address:       hwAddr1Parse,
		L2AddressLength: uint32(len(hwAddr1Parse)),
		LinkSpeed:       2, // 100MB, full duplex
	}

	var notif *intf.InterfaceNotification

	Eventually(publishChan).Should(Receive(&notif))
	Expect(notif.Type).To(Equal(intf.InterfaceNotification_UNKNOWN))
	Expect(notif.State.AdminStatus).To(Equal(intf.InterfacesState_Interface_UP))
	Expect(notif.State.OperStatus).To(Equal(intf.InterfacesState_Interface_DOWN))
	Expect(notif.State.InternalName).To(Equal("if0"))
	Expect(notif.State.Mtu).To(BeEquivalentTo(9216))
	Expect(notif.State.PhysAddress).To(Equal("01:23:45:67:89:ab"))
	Expect(notif.State.Duplex).To(Equal(intf.InterfacesState_Interface_FULL))
	Expect(notif.State.Speed).To(BeEquivalentTo(100 * 1000000))
}

// Test deleted notification
func TestInterfaceStateUpdaterIfStateDeleted(t *testing.T) {
	goVPPMux, index, ifPlugin, notifChan, publishChan := testPluginDataInitialization(t)
	defer testPluginDataTeardown(ifPlugin, goVPPMux)

	// Register name
	index.RegisterName("test", 0, &intf.Interfaces_Interface{
		Name:        "test",
		Enabled:     true,
		Type:        intf.InterfaceType_MEMORY_INTERFACE,
		IpAddresses: []string{"192.168.0.1/24"},
	})

	// Test notifications
	notifChan <- &interfaces.SwInterfaceEvent{
		PID:         0,
		SwIfIndex:   0,
		AdminUpDown: 1,
		LinkUpDown:  1,
		Deleted:     0,
	}

	var notif *intf.InterfaceNotification

	Eventually(publishChan).Should(Receive(&notif))
	Expect(notif.Type).To(Equal(intf.InterfaceNotification_UPDOWN))
	Expect(notif.State.AdminStatus).Should(BeEquivalentTo(intf.InterfacesState_Interface_UP))

	// Unregister name
	index.UnregisterName("test")

	Eventually(publishChan).Should(Receive(&notif))
	Expect(notif.Type).To(Equal(intf.InterfaceNotification_UNKNOWN))
	Expect(notif.State.AdminStatus).Should(BeEquivalentTo(intf.InterfacesState_Interface_DELETED))
	Expect(notif.State.OperStatus).Should(BeEquivalentTo(intf.InterfacesState_Interface_DELETED))
}
