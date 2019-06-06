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

	"github.com/ligato/cn-infra/logging/logrus"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/ip"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1908/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/vppcallmock"
	. "github.com/onsi/gomega"
)

func TestDumpVrfTables(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()
	vthandler := NewVrfTableVppHandler(ctx.MockChannel, logrus.DefaultLogger())

	ctx.MockVpp.MockReply(
		&ip.IPFibDetails{
			TableID:   3,
			TableName: []byte("table3"),
			Path:      []ip.FibPath{{SwIfIndex: 2}, {SwIfIndex: 4}},
		},
		&ip.IPFibDetails{
			TableID:   3,
			TableName: []byte("table3"),
			Path:      []ip.FibPath{{SwIfIndex: 5}},
		},
		&ip.IPFibDetails{
			TableID:   2,
			TableName: []byte("table2"),
			Path:      []ip.FibPath{{SwIfIndex: 5}},
		})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	ctx.MockVpp.MockReply(
		&ip.IP6FibDetails{
			TableID:   2,
			TableName: []byte("table2"),
			Path:      []ip.FibPath{{SwIfIndex: 5}},
		})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})

	vrfTables, err := vthandler.DumpVrfTables()
	Expect(err).To(Succeed())
	Expect(vrfTables).To(HaveLen(3))
	if vrfTables[0].Label == "table2" {
		Expect(vrfTables[1]).To(Equal(&l3.VrfTable{Id: 3, Protocol: l3.VrfTable_IPV4, Label: "table3"}))
		Expect(vrfTables[0]).To(Equal(&l3.VrfTable{Id: 2, Protocol: l3.VrfTable_IPV4, Label: "table2"}))
	} else {
		Expect(vrfTables[0]).To(Equal(&l3.VrfTable{Id: 3, Protocol: l3.VrfTable_IPV4, Label: "table3"}))
		Expect(vrfTables[1]).To(Equal(&l3.VrfTable{Id: 2, Protocol: l3.VrfTable_IPV4, Label: "table2"}))
	}
	Expect(vrfTables[2]).To(Equal(&l3.VrfTable{Id: 2, Protocol: l3.VrfTable_IPV6, Label: "table2"}))
}
