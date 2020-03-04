//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package e2e

import (
	"context"
	"io"
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	linux_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
	linux_ns "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
	vpp_interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestTelemetryStatsPoller(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		msName     = "microservice1"
		fullMsName = msNamePrefix + msName
		srcTapName = "vpp_span_src"
		dstTapName = "vpp_span_dst"
	)

	srcTap := &vpp_interfaces.Interface{
		Name:    srcTapName,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		IpAddresses: []string{
			"10.10.1.2/24",
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version: 2,
			},
		},
	}
	srcLinuxTap := &linux_interfaces.Interface{
		Name:    "linux_span_tap1",
		Type:    linux_interfaces.Interface_TAP_TO_VPP,
		Enabled: true,
		IpAddresses: []string{
			"10.10.1.1/24",
		},
		HostIfName: "linux_span_tap1",
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: srcTapName,
			},
		},
	}

	dstTap := &vpp_interfaces.Interface{
		Name:    dstTapName,
		Type:    vpp_interfaces.Interface_TAP,
		Enabled: true,
		IpAddresses: []string{
			"10.20.1.2/24",
		},
		Link: &vpp_interfaces.Interface_Tap{
			Tap: &vpp_interfaces.TapLink{
				Version:        2,
				ToMicroservice: fullMsName,
			},
		},
	}
	dstLinuxTap := &linux_interfaces.Interface{
		Name:    "linux_span_tap2",
		Type:    linux_interfaces.Interface_TAP_TO_VPP,
		Enabled: true,
		IpAddresses: []string{
			"10.20.1.1/24",
		},
		HostIfName: "linux_span_tap2",
		Link: &linux_interfaces.Interface_Tap{
			Tap: &linux_interfaces.TapLink{
				VppTapIfName: dstTapName,
			},
		},
		Namespace: &linux_ns.NetNamespace{
			Type:      linux_ns.NetNamespace_MICROSERVICE,
			Reference: fullMsName,
		},
	}

	spanRx := &vpp_interfaces.Span{
		InterfaceFrom: srcTapName,
		InterfaceTo:   dstTapName,
		Direction:     vpp_interfaces.Span_RX,
	}

	ctx.startMicroservice(msName)

	req := ctx.grpcClient.ChangeRequest()
	err := req.Update(dstTap, dstLinuxTap, spanRx).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	Eventually(ctx.getValueStateClb(dstTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))

	Expect(ctx.getValueState(spanRx)).To(Equal(kvscheduler.ValueState_PENDING))

	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(srcTap, srcLinuxTap).Send(context.Background())
	Expect(err).ToNot(HaveOccurred())

	Eventually(ctx.getValueStateClb(srcTap)).Should(Equal(kvscheduler.ValueState_CONFIGURED))

	Expect(ctx.getValueState(spanRx)).To(Equal(kvscheduler.ValueState_CONFIGURED))

	Expect(ctx.pingFromVPP("10.20.1.1")).To(Succeed())
	Expect(ctx.pingFromMs(msName, "10.20.1.2")).To(Succeed())

	pollerClient := configurator.NewStatsPollerServiceClient(ctx.grpcConn)

	t.Run("periodSec=0", func(tt *testing.T) {
		RegisterTestingT(tt)

		stream, err := pollerClient.PollStats(context.Background(), &configurator.PollStatsRequest{
			PeriodSec: 0,
		})
		Expect(err).ToNot(HaveOccurred())
		maxSeq := uint32(0)
		n := 0
		for {
			stats, err := stream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				tt.Fatal("recv error:", err)
			}
			tt.Logf("stats: %+v", stats)
			n++
			if stats.GetPollSeq() > maxSeq {
				maxSeq = stats.GetPollSeq()
			}
		}
		Expect(n).To(BeEquivalentTo(3))
		Expect(maxSeq).To(BeEquivalentTo(0))
	})

	t.Run("numPolls=1", func(tt *testing.T) {
		RegisterTestingT(tt)

		stream, err := pollerClient.PollStats(context.Background(), &configurator.PollStatsRequest{
			NumPolls:  1,
			PeriodSec: 1,
		})
		Expect(err).ToNot(HaveOccurred())
		maxSeq := uint32(0)
		n := 0
		for {
			stats, err := stream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				tt.Fatal("recv error:", err)
			}
			tt.Logf("stats: %+v", stats)
			n++
			if stats.GetPollSeq() > maxSeq {
				maxSeq = stats.GetPollSeq()
			}
		}
		Expect(n).To(BeEquivalentTo(3))
		Expect(maxSeq).To(BeEquivalentTo(1))
	})

	t.Run("numPolls=2", func(tt *testing.T) {
		RegisterTestingT(tt)

		stream, err := pollerClient.PollStats(context.Background(), &configurator.PollStatsRequest{
			NumPolls: 2,
		})
		Expect(err).ToNot(HaveOccurred())
		_, err = stream.Recv()
		Expect(err).To(HaveOccurred())
	})

}
