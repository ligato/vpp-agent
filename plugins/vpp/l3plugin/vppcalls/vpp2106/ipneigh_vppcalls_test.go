//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package vpp2106

/*
import (
	"testing"

	. "github.com/onsi/gomega"

	vpp_vpe "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)


func TestGetIPScanNeighbor(t *testing.T) {
	tests := []struct {
		name     string
		cliReply string
		expected l3.IPScanNeighbor
	}{
		{
			name: "default",
			cliReply: `ip4:
  limit:50000, age:0, recycle:0
ip6:
  limit:50000, age:0, recycle:0
`,
			expected: l3.IPScanNeighbor{
				Mode:      l3.IPScanNeighbor_DISABLED,
				MaxUpdate: 50000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := vppmock.SetupTestCtx(t)
			defer ctx.TeardownTestCtx()

			ctx.MockVpp.MockReply(&vpp_vpe.CliInbandReply{
				Reply: test.cliReply,
			})

			handler := NewIPNeighVppHandler(ctx.MockChannel, nil)

			ipNeigh, err := handler.GetIPScanNeighbor()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ipNeigh.Mode).To(Equal(test.expected.Mode))
			Expect(ipNeigh.ScanInterval).To(Equal(test.expected.ScanInterval))
			Expect(ipNeigh.ScanIntDelay).To(Equal(test.expected.ScanIntDelay))
			Expect(ipNeigh.StaleThreshold).To(Equal(test.expected.StaleThreshold))
			Expect(ipNeigh.MaxUpdate).To(Equal(test.expected.MaxUpdate))
			Expect(ipNeigh.MaxProcTime).To(Equal(test.expected.MaxProcTime))
		})
	}
}
*/
