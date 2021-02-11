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

package vpp2101_test

import (
	"encoding/base64"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/cn-infra/v2/logging/logrus"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/ip_types"
	vpp_wg "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2101/wireguard"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/vppcalls/vpp2101"
	wg "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/wireguard"
)

func TestVppAddPeer(t *testing.T) {
	ctx, wgHandler, ifIndex := wgTestSetup(t)
	defer ctx.TeardownTestCtx()

	ifIndex.Put("wg1", &ifaceidx.IfaceMetadata{SwIfIndex: 2})

	ctx.MockVpp.MockReply(&vpp_wg.WireguardPeerAddReply{
		PeerIndex: 0,
	})

	peer := &wg.Peer{
		PublicKey:           "dIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
		WgIfName:            "wg1",
		Port:                12314,
		PersistentKeepalive: 10,
		Endpoint:            "10.10.2.1",
		Flags:               0,
		AllowedIps:          []string{"10.10.0.0/24"},
	}

	index, err := wgHandler.AddPeer(peer)
	Expect(err).To(BeNil())
	Expect(index).To(Equal(uint32(0)))


	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_wg.WireguardPeerAdd)
	Expect(ok).To(BeTrue())

	pubKeyBin,_:= base64.StdEncoding.DecodeString("dIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=")
	Expect(vppMsg.Peer.PublicKey).To(BeEquivalentTo(pubKeyBin))
	Expect(vppMsg.Peer.SwIfIndex).To(BeEquivalentTo(2))
	Expect(vppMsg.Peer.Port).To(BeEquivalentTo(12314))
	Expect(vppMsg.Peer.PersistentKeepalive).To(BeEquivalentTo(10))
	Expect(vppMsg.Peer.TableID).To(BeEquivalentTo(0))
	Expect(vppMsg.Peer.Endpoint.Un.GetIP4()).To(BeEquivalentTo(ip_types.IP4Address{10, 10, 2, 1}))
	Expect(vppMsg.Peer.Flags).To(BeEquivalentTo(0))
	Expect(vppMsg.Peer.AllowedIps).To(BeEquivalentTo([]ip_types.Prefix{ip_types.Prefix{
		Address: ip_types.Address{
			Af: ip_types.ADDRESS_IP4,
			Un: ip_types.AddressUnionIP4(ip_types.IP4Address{10, 10, 0, 0}),
		},
		Len:     24,
	}}))
}

func TestVppRemovePeer(t *testing.T) {
	ctx, wgHandler, _ := wgTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_wg.WireguardPeerRemoveReply{})
	err := wgHandler.RemovePeer(0)

	Expect(err).ShouldNot(HaveOccurred())
}

func wgTestSetup(t *testing.T) (*vppmock.TestCtx, vppcalls.WgVppAPI, ifaceidx.IfaceMetadataIndexRW) {
	ctx := vppmock.SetupTestCtx(t)
	log := logrus.NewLogger("test-log")
	ifIndex := ifaceidx.NewIfaceIndex(log, "wg-test-ifidx")
	wgHandler := vpp2101.NewWgVppHandler(ctx.MockChannel, ifIndex, log)
	return ctx, wgHandler, ifIndex
}