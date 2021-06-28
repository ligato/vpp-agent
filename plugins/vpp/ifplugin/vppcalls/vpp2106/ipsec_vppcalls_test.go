// Copyright (c) 2021 Cisco and/or its affiliates.
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

package vpp2106_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	vpp_ipsec "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ipsec"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/tunnel_types"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestAddIPSecTunnelInterface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecItfCreateReply{
		SwIfIndex: 2,
	})

	ipSecLink := &ifs.IPSecLink{
		TunnelMode: ifs.IPSecLink_POINT_TO_POINT,
	}
	index, err := ifHandler.AddIPSecTunnelInterface(ctx.Context, "if1", ipSecLink)
	Expect(err).To(BeNil())
	Expect(index).To(Equal(uint32(2)))

	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ipsec.IpsecItfCreate)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).ToNot(BeNil())

	itf := vpp_ipsec.IpsecItf{
		Mode: tunnel_types.TunnelMode(ifs.IPSecLink_POINT_TO_POINT),
	}

	Expect(vppMsg.Itf).To(Equal(itf))
}

func TestAddIPSecTunnelInterfaceError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecItfCreateReply{
		SwIfIndex: 2,
		Retval:    9,
	})

	index, err := ifHandler.AddIPSecTunnelInterface(ctx.Context, "if1", &ifs.IPSecLink{})
	Expect(err).ToNot(BeNil())
	Expect(index).To(Equal(uint32(0)))
}

func TestDeleteIPSecTunnelInterface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecItfDeleteReply{})

	err := ifHandler.DeleteIPSecTunnelInterface(ctx.Context, "if1", 2, &ifs.IPSecLink{})
	Expect(err).To(BeNil())

	vppMsg, ok := ctx.MockChannel.Msg.(*vpp_ipsec.IpsecItfDelete)
	Expect(ok).To(BeTrue())
	Expect(vppMsg).ToNot(BeNil())

	Expect(vppMsg.SwIfIndex).To(Equal(interface_types.InterfaceIndex(2)))
}

func TestDeleteIPSecTunnelInterfaceError(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&vpp_ipsec.IpsecItfDeleteReply{
		Retval: 9,
	})

	err := ifHandler.DeleteIPSecTunnelInterface(ctx.Context, "if1", 2, &ifs.IPSecLink{})
	Expect(err).ToNot(BeNil())
}
