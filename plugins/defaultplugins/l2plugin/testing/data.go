// Package testing provides tools and input data for unit testing of the
// l2plugin. What remains to be defined are scenarios.
package testing

import (
	test_if "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/testing"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
)

var (
	// BDMemif100011ToMemif100012 is an example bridge domain configuration
	BDMemif100011ToMemif100012 = l2.BridgeDomains_BridgeDomain{
		Name:                "br2",
		Flood:               false,
		UnknownUnicastFlood: false,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /*means disable aging*/
		Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
			{
				Name: "memif1",
				BridgedVirtualInterface: true,
			}, {
				Name: "memif4",
				BridgedVirtualInterface: false,
			},
		},
	}

	// BDAfPacketVeth1VxlanVni5 is an example bridge domain configuration
	BDAfPacketVeth1VxlanVni5 = l2.BridgeDomains_BridgeDomain{
		Name:                "br1",
		Flood:               false,
		UnknownUnicastFlood: false,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0, /*means disable aging*/
		Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
			{
				Name: test_if.VxlanVni5.Name,
				BridgedVirtualInterface: false,
			}, {
				Name: test_if.AfPacketVeth1.Name,
				BridgedVirtualInterface: false,
			},
		},
	}

	// XConMemif100011ToMemif100012 is an example of cross connect configuration
	XConMemif100011ToMemif100012 = l2.XConnectPairs_XConnectPair{
		ReceiveInterface:  test_if.Memif100011Master.Name,
		TransmitInterface: test_if.Memif100011Master.Name,
	}
)

// SimpleBridgeDomain1XIfaceBuilder creates simple bridge domain with defined name and one interface
func SimpleBridgeDomain1XIfaceBuilder(name string, iface1 string, bvi1 bool) l2.BridgeDomains_BridgeDomain {
	return l2.BridgeDomains_BridgeDomain{
		Name:                name,
		Flood:               false,
		UnknownUnicastFlood: false,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0,
		Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
			{
				Name: iface1,
				BridgedVirtualInterface: bvi1,
			},
		},
	}
}

// SimpleBridgeDomain2XIfaceBuilder creates simple bridge domain with defined name and two interfaces
func SimpleBridgeDomain2XIfaceBuilder(name string, iface1 string, iface2 string, bvi1 bool, bvi2 bool) l2.BridgeDomains_BridgeDomain {
	return l2.BridgeDomains_BridgeDomain{
		Name:                name,
		Flood:               false,
		UnknownUnicastFlood: false,
		Forward:             true,
		Learn:               true,
		ArpTermination:      false,
		MacAge:              0,
		Interfaces: []*l2.BridgeDomains_BridgeDomain_Interfaces{
			{
				Name: iface1,
				BridgedVirtualInterface: bvi1,
			}, {
				Name: iface2,
				BridgedVirtualInterface: bvi2,
			},
		},
	}
}

// FIBBuilder builds FIB table entry
func FIBBuilder(mac string, bdName string, iface string, bvi bool) l2.FibTableEntries_FibTableEntry {
	return l2.FibTableEntries_FibTableEntry{
		PhysAddress:             mac,
		BridgeDomain:            bdName,
		OutgoingInterface:       iface,
		StaticConfig:            true,
		BridgedVirtualInterface: bvi,
	}
}

// XconnectBuilder prepares xConnect interface pair
func XconnectBuilder(rIface string, tIface string) l2.XConnectPairs_XConnectPair {
	return l2.XConnectPairs_XConnectPair{
		ReceiveInterface:  rIface,
		TransmitInterface: tIface,
	}
}
