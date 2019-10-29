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

package vpp2001_324

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
	vpp_ip "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/ip"
	vpp_vpe "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/vpe"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/vppcallmock"
	l3 "go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp/l3"
)

func TestDumpVrfTables(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()
	vthandler := NewVrfTableVppHandler(ctx.MockChannel, logrus.DefaultLogger())

	ctx.MockVpp.MockReply(
		&vpp_ip.IPTableDetails{
			Table: vpp_ip.IPTable{
				TableID: 1,
				Name:    []byte("table3"),
				IsIP6:   0,
			},
		},
		&vpp_ip.IPTableDetails{
			Table: vpp_ip.IPTable{
				TableID: 2,
				Name:    []byte("table3"),
				IsIP6:   0,
			},
		},
		&vpp_ip.IPTableDetails{
			Table: vpp_ip.IPTable{
				TableID: 3,
				Name:    []byte("table2"),
				IsIP6:   1,
			},
		},
	)
	ctx.MockVpp.MockReply(&vpp_vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(
		&vpp_ip.IPRouteDetails{
			Route: vpp_ip.IPRoute{
				TableID: 2,
				Paths:   []vpp_ip.FibPath{{SwIfIndex: 5}},
			},
		})
	ctx.MockVpp.MockReply(&vpp_vpe.ControlPingReply{})

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
