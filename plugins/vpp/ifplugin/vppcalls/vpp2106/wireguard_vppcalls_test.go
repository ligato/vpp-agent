//  Copyright (c) 2021 Doc.ai and/or its affiliates.
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

package vpp2106_test

import (
	"encoding/base64"
	"testing"

	. "github.com/onsi/gomega"

	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	vpp_wg "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/wireguard"

	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestAddWgTunnelInterface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_wg.WireguardInterfaceCreateReply{
		SwIfIndex: 2,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	wgLink := &ifs.WireguardLink{
		PrivateKey:   "gIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=",
		Port:         12312,
		SrcAddr:      "10.0.0.1",
	}

	index, err := ifHandler.AddWireguardTunnel("wg1", wgLink)
	Expect(err).To(BeNil())
	Expect(index).To(Equal(uint32(2)))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_wg.WireguardInterfaceCreate)
		if ok {
			privKeyBin,_:= base64.StdEncoding.DecodeString("gIjXzrQfIFf80d0O8Hd2KhcfkKLRncc+8C70OjotIW8=")
			Expect(vppMsg.GenerateKey).To(BeEquivalentTo(false))
			Expect(vppMsg.Interface.PrivateKey).To(BeEquivalentTo(privKeyBin))
			Expect(vppMsg.Interface.Port).To(BeEquivalentTo(12312))
			Expect(vppMsg.Interface.SrcIP.Un.GetIP4()).To(BeEquivalentTo(ip_types.IP4Address{ 10, 0, 0, 1 }))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddWgTunnelInterfaceWithGenKey(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_wg.WireguardInterfaceCreateReply{
		SwIfIndex: 2,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	wgLink := &ifs.WireguardLink{
		Port:         12312,
		SrcAddr:      "10.0.0.1",
	}

	index, err := ifHandler.AddWireguardTunnel("wg1", wgLink)
	Expect(err).To(BeNil())
	Expect(index).To(Equal(uint32(2)))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_wg.WireguardInterfaceCreate)
		if ok {
			Expect(vppMsg.GenerateKey).To(BeEquivalentTo(true))
			Expect(vppMsg.Interface.Port).To(BeEquivalentTo(12312))
			Expect(vppMsg.Interface.SrcIP.Un.GetIP4()).To(BeEquivalentTo(ip_types.IP4Address{ 10, 0, 0, 1 }))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestDeleteWgTunnelInterface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_wg.WireguardInterfaceDeleteReply{})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DeleteWireguardTunnel("wg1", 2)

	Expect(err).To(BeNil())
}