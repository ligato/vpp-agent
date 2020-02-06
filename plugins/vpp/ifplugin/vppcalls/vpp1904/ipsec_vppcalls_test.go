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

package vpp1904_test

import (
	"encoding/hex"
	"testing"

	. "github.com/onsi/gomega"

	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/ipsec"
)

func TestAddIPSecTunnelInterface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&ipsec.IpsecTunnelIfAddDelReply{
		SwIfIndex: 2,
	})

	index, err := ifHandler.AddIPSecTunnelInterface(ctx.Context, "if1", &vpp_interfaces.IPSecLink{
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
	})
	Expect(err).To(BeNil())
	Expect(index).To(Equal(uint32(2)))

	vppMsg, ok := ctx.MockChannel.Msg.(*ipsec.IpsecTunnelIfAddDel)
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
	Expect(vppMsg.LocalIP.Af).To(Equal(ipsec.AddressFamily(0)))
	Expect(vppMsg.LocalIP.Un).To(BeEquivalentTo(ipsec.AddressUnion{XXX_UnionData: [16]byte{10, 10, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}))
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
	ctx.MockVpp.MockReply(&ipsec.IpsecTunnelIfAddDelReply{
		SwIfIndex: 2,
		Retval:    9,
	})

	index, err := ifHandler.AddIPSecTunnelInterface(ctx.Context, "if1", &vpp_interfaces.IPSecLink{
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
	ctx.MockVpp.MockReply(&ipsec.IpsecTunnelIfAddDelReply{
		SwIfIndex: 2,
	})

	err := ifHandler.DeleteIPSecTunnelInterface(ctx.Context, "if1", &vpp_interfaces.IPSecLink{
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
	ctx.MockVpp.MockReply(&ipsec.IpsecTunnelIfAddDelReply{
		SwIfIndex: 2,
		Retval:    9,
	})

	err := ifHandler.DeleteIPSecTunnelInterface(ctx.Context, "if1", &vpp_interfaces.IPSecLink{
		Esn:            true,
		LocalIp:        "10.10.0.1",
		LocalCryptoKey: "4a506a794f574265564551694d653768",
	})
	Expect(err).ToNot(BeNil())
}
