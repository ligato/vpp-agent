// Copyright (c) 2019 Cisco and/or its affiliates.
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

package vpp2001_test

import (
	"encoding/hex"
	"net"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ip_types"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/ipsec"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestAddIPSecTunnelInterface(t *testing.T) {
	var ipv4Addr [16]byte

	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecTunnelIfAddDelReply{
		SwIfIndex: 2,
	})

	ipSecLink := &ifs.IPSecLink{
		Esn:             true,
		AntiReplay:      true,
		LocalIp:         "10.10.0.1",
		RemoteIp:        "10.10.0.2",
		LocalSpi:        1500,
		RemoteSpi:       2000,
		CryptoAlg:       9,
		LocalCryptoKey:  "4a506a794f574265564551694d653768",
		RemoteCryptoKey: "9a506a794f574265564551694d653456",
		IntegAlg:        4,
		LocalIntegKey:   "3a506a794f574265564551694d653769",
		RemoteIntegKey:  "8a506a794f574265564551694d653457",
		EnableUdpEncap:  true,
	}
	index, err := ifHandler.AddIPSecTunnelInterface(ctx.Context, "if1", ipSecLink)
	Expect(err).To(BeNil())
	Expect(index).To(Equal(uint32(2)))

	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ipsec.IpsecTunnelIfAddDel)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).ToNot(BeNil())
	localCryptoKey, err := hex.DecodeString("4a506a794f574265564551694d653768")
	Expect(err).To(BeNil())
	remoteCryptoKey, err := hex.DecodeString("9a506a794f574265564551694d653456")
	Expect(err).To(BeNil())
	localIntegKey, err := hex.DecodeString("3a506a794f574265564551694d653769")
	Expect(err).To(BeNil())
	remoteIntegKey, err := hex.DecodeString("8a506a794f574265564551694d653457")
	Expect(err).To(BeNil())

	Expect(vppMsg.Esn).To(Equal(uint8(1)))
	Expect(vppMsg.IsAdd).To(Equal(uint8(1)))
	Expect(vppMsg.AntiReplay).To(Equal(uint8(1)))
	Expect(vppMsg.LocalIP.Af).To(Equal(ip_types.ADDRESS_IP4))
	copy(ipv4Addr[:], net.ParseIP(ipSecLink.LocalIp)[12:])
	Expect(vppMsg.LocalIP.Un).To(BeEquivalentTo(vpp_ipsec.AddressUnion{XXX_UnionData: ipv4Addr}))
	Expect(vppMsg.LocalSpi).To(Equal(uint32(1500)))
	Expect(vppMsg.RemoteSpi).To(Equal(uint32(2000)))
	Expect(vppMsg.CryptoAlg).To(Equal(uint8(9)))
	Expect(vppMsg.LocalCryptoKey).To(BeEquivalentTo(localCryptoKey))
	Expect(vppMsg.LocalCryptoKeyLen).To(Equal(uint8(16)))
	Expect(vppMsg.RemoteCryptoKey).To(BeEquivalentTo(remoteCryptoKey))
	Expect(vppMsg.RemoteCryptoKeyLen).To(Equal(uint8(16)))
	Expect(vppMsg.IntegAlg).To(Equal(uint8(4)))
	Expect(vppMsg.LocalIntegKey).To(BeEquivalentTo(localIntegKey))
	Expect(vppMsg.LocalIntegKeyLen).To(Equal(uint8(16)))
	Expect(vppMsg.RemoteIntegKey).To(BeEquivalentTo(remoteIntegKey))
	Expect(vppMsg.RemoteIntegKeyLen).To(Equal(uint8(16)))
	Expect(vppMsg.UDPEncap).To(Equal(uint8(1)))
}

func TestAddIPSecTunnelInterfaceError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecTunnelIfAddDelReply{
		SwIfIndex: 2,
		Retval:    9,
	})

	index, err := ifHandler.AddIPSecTunnelInterface(ctx.Context, "if1", &ifs.IPSecLink{
		Esn:            true,
		LocalIp:        "10.10.0.1",
		LocalCryptoKey: "4a506a794f574265564551694d653768",
	})
	Expect(err).ToNot(BeNil())
	Expect(index).To(Equal(uint32(0)))
}

func TestDeleteIPSecTunnelInterface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecTunnelIfAddDelReply{
		SwIfIndex: 2,
	})

	err := ifHandler.DeleteIPSecTunnelInterface(ctx.Context, "if1", &ifs.IPSecLink{
		Esn:             true,
		LocalIp:         "10.10.0.1",
		RemoteIp:        "10.10.0.2",
		LocalCryptoKey:  "4a506a794f574265564551694d653768",
		RemoteCryptoKey: "9a506a794f574265564551694d653456",
	})

	Expect(err).To(BeNil())
}

func TestDeleteIPSecTunnelInterfaceError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecTunnelIfAddDelReply{
		SwIfIndex: 2,
		Retval:    9,
	})

	err := ifHandler.DeleteIPSecTunnelInterface(ctx.Context, "if1", &ifs.IPSecLink{
		Esn:            true,
		LocalIp:        "10.10.0.1",
		LocalCryptoKey: "4a506a794f574265564551694d653768",
	})
	Expect(err).ToNot(BeNil())
}
