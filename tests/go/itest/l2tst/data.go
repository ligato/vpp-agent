// Package l2tst provides tools and input data for unit testing of the
// l2plugin. What remains to be defined are scenarios.
package l2tst

import (
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/tests/go/itest/iftst"
)

var (
	// BDMemif100011ToMemif100012 is an example bridge domain configuration.
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

	// BDAfPacketVeth1VxlanVni5 is an example bridge domain configuration.
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
				Name: iftst.VxlanVni5.Name,
				BridgedVirtualInterface: false,
			}, {
				Name: iftst.AfPacketVeth1.Name,
				BridgedVirtualInterface: false,
			},
		},
	}

	// XConMemif100011ToMemif100012 is an example of cross connect configuration.
	XConMemif100011ToMemif100012 = l2.XConnectPairs_XConnectPair{
		ReceiveInterface:  iftst.Memif100011Master.Name,
		TransmitInterface: iftst.Memif100011Master.Name,
	}
)

// SimpleBridgeDomain1XIfaceBuilder creates a simple bridge domain with defined name and one interface.
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

// SimpleBridgeDomain2XIfaceBuilder creates a simple bridge domain with defined name and two interfaces.
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

// FIBBuilder builds FIB table entry.
func FIBBuilder(mac string, bdName string, iface string, bvi bool) l2.FibTableEntries_FibTableEntry {
	return l2.FibTableEntries_FibTableEntry{
		PhysAddress:             mac,
		BridgeDomain:            bdName,
		OutgoingInterface:       iface,
		StaticConfig:            true,
		BridgedVirtualInterface: bvi,
	}
}

// XconnectBuilder prepares xConnect interface pair.
func XconnectBuilder(rIface string, tIface string) l2.XConnectPairs_XConnectPair {
	return l2.XConnectPairs_XConnectPair{
		ReceiveInterface:  rIface,
		TransmitInterface: tIface,
	}
}
