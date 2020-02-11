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

package vpp1908

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/ip"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/vpe"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

func TestDumpVrfTables(t *testing.T) {
	ctx := vppmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()
	vthandler := NewVrfTableVppHandler(ctx.MockChannel, logrus.DefaultLogger())

	ctx.MockVpp.MockReply(
		&ip.IPTableDetails{
			Table: ip.IPTable{
				TableID: 1,
				Name:    []byte("table3"),
				IsIP6:   0,
			},
		},
		&ip.IPTableDetails{
			Table: ip.IPTable{
				TableID: 2,
				Name:    []byte("table3"),
				IsIP6:   0,
			},
		},
		&ip.IPTableDetails{
			Table: ip.IPTable{
				TableID: 3,
				Name:    []byte("table2"),
				IsIP6:   1,
			},
		},
	)
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(
		&ip.IPRouteDetails{
			Route: ip.IPRoute{
				TableID: 2,
				Paths:   []ip.FibPath{{SwIfIndex: 5}},
			},
		})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	vrfTables, err := vthandler.DumpVrfTables()
	Expect(err).To(Succeed())
	Expect(vrfTables).To(HaveLen(3))
	if vrfTables[0].Label == "table2" {
		Expect(vrfTables[1]).To(Equal(&l3.VrfTable{Id: 3, Protocol: l3.VrfTable_IPV4, Label: "table3"}))
		Expect(vrfTables[0]).To(Equal(&l3.VrfTable{Id: 2, Protocol: l3.VrfTable_IPV4, Label: "table2"}))
	} else {
		Expect(vrfTables[0]).To(Equal(&l3.VrfTable{Id: 1, Protocol: l3.VrfTable_IPV4, Label: "table3"}))
		Expect(vrfTables[1]).To(Equal(&l3.VrfTable{Id: 2, Protocol: l3.VrfTable_IPV4, Label: "table3"}))
	}
	Expect(vrfTables[2]).To(Equal(&l3.VrfTable{Id: 3, Protocol: l3.VrfTable_IPV6, Label: "table2"}))
}
