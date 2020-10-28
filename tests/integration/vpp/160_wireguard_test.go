//  Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package vpp

import (
	"fmt"
	"testing"

	"go.ligato.io/cn-infra/v2/logging/logrus"

	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	ifplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin"
	wgplugin_vppcalls "go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	wg "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/wireguard"
)

type testEntry struct {
	name       string
	wgInt      *interfaces.WireguardLink
	wgInt2     *interfaces.WireguardLink
	peer       *wg.Peer
	peer2      *wg.Peer
	shouldFail bool
}

func TestWireguard(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	release := ctx.versionInfo.Release()
	if release < "20.09" {
		t.Skipf("Wireguard: skipped for VPP < 20.09 (%s)", release)
	}

	ifHandler := ifplugin_vppcalls.CompatibleInterfaceVppHandler(ctx.vppClient, logrus.NewLogger("test"))
	ifIndexes := ifaceidx.NewIfaceIndex(logrus.NewLogger("test"), "test-ifidx")

	tests := []testEntry{
		{
			name: "Create Wireguard tunnel (IPv4)",
			wgInt: &interfaces.WireguardLink{
				PrivateKey: "gIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
				Port:       12312,
				SrcAddr:    "10.10.0.1",
			},
			shouldFail: false,
		},
		{
			name: "Create Wireguard tunnel with invalid privateKey",
			wgInt: &interfaces.WireguardLink{
				PrivateKey: "d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
				Port:       12312,
				SrcAddr:    "10.10.0.1",
			},
			shouldFail: true,
		},
		{
			name: "Create Wireguard 2 tunnels (IPv4)",
			wgInt: &interfaces.WireguardLink{
				PrivateKey: "gIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
				Port:       12322,
				SrcAddr:    "10.10.0.1",
			},
			wgInt2: &interfaces.WireguardLink{
				PrivateKey: "qDUQL8I5RNMWfbi3qgFs237FYD+SyZTj5g0Ix3qRvGs=",
				Port:       12323,
				SrcAddr:    "10.11.0.1",
			},
			shouldFail: false,
		},
		{
			name: "Create Wireguard tunnel with peer",
			wgInt: &interfaces.WireguardLink{
				PrivateKey: "gIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
				Port:       12332,
				SrcAddr:    "10.10.0.1",
			},
			peer: &wg.Peer{
				PublicKey:           "dIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
				WgIfName:            "wg3",
				Port:                12314,
				PersistentKeepalive: 10,
				Endpoint:            "10.10.0.2",
				Flags:               0,
				AllowedIps:          []string{"10.10.0.0/24"},
			},
			shouldFail: false,
		},
		{
			name: "Create Wireguard tunnel with 2 itfs and 2 peers",
			wgInt: &interfaces.WireguardLink{
				PrivateKey: "gIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
				Port:       12342,
				SrcAddr:    "10.10.0.1",
			},
			wgInt2: &interfaces.WireguardLink{
				PrivateKey: "qDUQL8I5RNMWfbi3qgFs237FYD+SyZTj5g0Ix3qRvGs=",
				Port:       12343,
				SrcAddr:    "10.11.0.1",
			},
			peer: &wg.Peer{
				PublicKey:           "dIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
				WgIfName:            "wg4",
				Port:                12314,
				PersistentKeepalive: 10,
				Endpoint:            "10.10.0.2",
				Flags:               0,
				AllowedIps:          []string{"10.10.0.0/24"},
			},
			peer2: &wg.Peer{
				PublicKey:           "33GyVvUQalLCscTfN8TxtTp/ixtSWg55PhHy0aWABHQ=",
				WgIfName:            "wg4-2",
				Port:                12314,
				PersistentKeepalive: 10,
				Endpoint:            "10.11.0.2",
				Flags:               0,
				AllowedIps:          []string{"10.11.0.0/24"},
			},
			shouldFail: false,
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ifIndexes.Clear()

			ifName := fmt.Sprintf("wg%d", i)
			ifIdx, err := ifHandler.AddWireguardTunnel(ifName, test.wgInt)

			ifIndexes.Put(ifName, &ifaceidx.IfaceMetadata{
				SwIfIndex: ifIdx,
			})

			if err != nil {
				if test.shouldFail {
					return
				}
				t.Fatalf("create Wireguard tunnel failed: %v\n", err)
			} else {
				if test.shouldFail && test.wgInt2 == nil {
					t.Fatal("create Wireguard tunnel must fail, but it's not")
				}
			}

			var (
				ifName2 string
				ifIdx2  uint32
			)
			if test.wgInt2 != nil {
				ifName2 := fmt.Sprintf("wg%d-2", i)
				ifIdx2, err = ifHandler.AddWireguardTunnel(ifName2, test.wgInt2)
				ifIndexes.Put(ifName2, &ifaceidx.IfaceMetadata{
					SwIfIndex: ifIdx2,
				})

				if err != nil {
					if test.shouldFail {
						return
					}
					t.Fatalf("create Wireguard tunnel failed: %v\n", err)
				} else {
					if test.shouldFail {
						t.Fatal("create Wireguard tunnel must fail, but it's not")
					}
				}
			}

			ifaces, err := ifHandler.DumpInterfaces(ctx.Ctx)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}
			iface, ok := ifaces[ifIdx]
			if !ok {
				t.Fatalf("Wireguard interface was not found in dump")
			}
			if test.wgInt2 != nil {
				_, ok := ifaces[ifIdx2]
				if !ok {
					t.Fatalf("Wireguard interface2 was not found in dump")
				}
			}

			err = peersTest(&test, ifIndexes, ctx)
			if err != nil {
				t.Fatalf("Peers failed: %v", err)
			}

			if iface.Interface.GetType() != interfaces.Interface_WIREGUARD_TUNNEL {
				t.Fatalf("Interface is not an Wireguard tunnel")
			}

			wgLink := iface.Interface.GetWireguard()
			if test.wgInt.SrcAddr != wgLink.SrcAddr {
				t.Fatalf("expected source address <%s>, got: <%s>", test.wgInt.SrcAddr, wgLink.SrcAddr)
			}

			err = ifHandler.DeleteWireguardTunnel(ifName, ifIdx)
			if err != nil {
				t.Fatalf("delete Wireguard tunnel failed: %v\n", err)
			}
			if test.wgInt2 != nil {
				err = ifHandler.DeleteWireguardTunnel(ifName2, ifIdx2)
				if err != nil {
					t.Fatalf("delete Wireguard tunnel failed: %v\n", err)
				}
			}

			ifaces, err = ifHandler.DumpInterfaces(ctx.Ctx)
			if err != nil {
				t.Fatalf("dumping interfaces failed: %v", err)
			}

			if _, ok := ifaces[ifIdx]; ok {
				t.Fatalf("Wireguard interface was found in dump after removing")
			}
			if test.wgInt2 != nil {
				if _, ok := ifaces[ifIdx2]; ok {
					t.Fatalf("Wireguard interface2 was found in dump after removing")
				}
			}
		})
	}
}

func peersTest(test *testEntry, ifIdx ifaceidx.IfaceMetadataIndex, ctx *TestCtx) (err error) {
	if test.peer == nil {
		return err
	}

	wgHandler := wgplugin_vppcalls.CompatibleWgVppHandler(ctx.vppClient, ifIdx, logrus.NewLogger("test"))
	if wgHandler == nil {
		return fmt.Errorf("no compatible wireguard handler")
	}
	peerIdx1, err := wgHandler.AddPeer(test.peer)

	if err != nil {
		if test.shouldFail {
			return
		}
		return err
	} else {
		if test.shouldFail && test.peer2 == nil {
			return fmt.Errorf("create peer must fail, but it's not")
		}
	}

	var (
		peerIdx2 uint32
	)
	if test.peer2 != nil {
		peerIdx2, err = wgHandler.AddPeer(test.peer2)

		if err != nil {
			if test.shouldFail {
				return
			}
			return err
		} else {
			if test.shouldFail {
				return fmt.Errorf("create peer must fail, but it's not")
			}
		}
	}

	peers, err := wgHandler.DumpWgPeers()
	if err != nil {
		return err
	}
	peer := peers[peerIdx1]

	if test.peer2 != nil {
		if len(peers) != 2 {
			return fmt.Errorf("Error peers dump")
		}
	} else {
		if len(peers) != 1 {
			return fmt.Errorf("Error peers dump")
		}
	}

	if test.peer.PublicKey != peer.PublicKey {
		return fmt.Errorf("expected source address <%s>, got: <%s>", test.peer.PublicKey, peer.PublicKey)
	}

	err = wgHandler.RemovePeer(peerIdx1)
	if err != nil {
		return fmt.Errorf("delete peer failed: %v\n", err)
	}
	if test.peer2 != nil {
		err = wgHandler.RemovePeer(peerIdx2)
		if err != nil {
			return fmt.Errorf("delete peer failed: %v\n", err)
		}
	}

	peers, err = wgHandler.DumpWgPeers()
	if err != nil {
		return err
	}

	if len(peers) != 0 {
		return fmt.Errorf("Error peers dump")
	}

	return
}
