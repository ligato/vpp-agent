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

package vpp1901

import (
	"testing"

	"github.com/ligato/vpp-agent/api/models/vpp/interfaces"
)

func TestCreateSubif(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()
	ctx.MockVpp.MockReply(&sub.CreateSubif{
		SwIfIndex: 2,
	})

	err := ifHandler.DeleteIPSecTunnelInterface("if1", &vpp_interfaces.IPSecLink{
		Esn:             true,
		LocalIp:         "10.10.0.1",
		RemoteIp:        "10.10.0.2",
		LocalCryptoKey:  "4a506a794f574265564551694d653768",
		RemoteCryptoKey: "9a506a794f574265564551694d653456",
	})

	Expect(err).To(BeNil())
}
package vpp1901

import (
"testing"

. "github.com/onsi/gomega"

"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/interfaces"
ifModel "github.com/ligato/vpp-agent/api/models/vpp/interfaces"

)

func TestCreateSubif(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&interfaces.CreateSubif{
		SwIfIndex: 3,
	})

	subif, _ := ifHandler.
		err := ifHandler.CreateSubif(10, 25)

	Expect(err).To(BeNil())
	vppMsg, ok := ctx.MockChannel.Msg.(*interfaces.CreateSubif)
	Expect(ok).To(BeTrue())
	Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
	Expect(vppMsg.VlanID).To(BeEquivalentTo(mac))
}