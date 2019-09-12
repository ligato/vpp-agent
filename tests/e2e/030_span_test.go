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

package e2e

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"testing"

	. "github.com/onsi/gomega"

	linux_interfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	linux_ns "github.com/ligato/vpp-agent/api/models/linux/namespace"
	vpp_interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

func TestSpan(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const (
		msName     = "microservice1"
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
				ToMicroservice: msNamePrefix + msName,
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
			Reference: msNamePrefix + msName,
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
	Expect(err).To(BeNil())

	Eventually(ctx.getValueStateClb(dstTap), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED))

	// While one of the end of span is missing, its state must be `PENDING`
	// in this case, source (`from`) interface is missing
	Expect(ctx.getValueState(spanRx)).To(Equal(kvs.ValueState_PENDING))

	req = ctx.grpcClient.ChangeRequest()
	err = req.Update(srcTap, srcLinuxTap).Send(context.Background())
	Expect(err).To(BeNil())

	Eventually(ctx.getValueStateClb(srcTap), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED))

	// When both ends for span are configured, span must be `CONFIGURED` also
	Expect(ctx.getValueState(spanRx)).To(Equal(kvs.ValueState_CONFIGURED))

	ctx.stopMicroservice(msName)
	Eventually(ctx.getValueStateClb(dstTap), msUpdateTimeout).Should(Equal(kvs.ValueState_PENDING))

	// While one of the end of span is missing, its state must be `PENDING`
	// in this case, destination (`to`) interface is missing
	Expect(ctx.getValueState(spanRx)).To(Equal(kvs.ValueState_PENDING))

	// Check output of `show span` command
	var stdout bytes.Buffer
	cmd := exec.Command("vppctl", "show", "int", "span")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output := stdout.String()
	space := regexp.MustCompile(`\s+`)
	s := space.ReplaceAllString(output, " ")
	Expect(s).To(Equal(""))

	// Get back our container and configure destination interface again
	ctx.startMicroservice(msName)

	Eventually(ctx.getValueStateClb(dstTap), msUpdateTimeout).Should(Equal(kvs.ValueState_CONFIGURED))

	// Everything must be ready for span to be `CONFIGURED`
	Expect(ctx.getValueState(spanRx)).To(Equal(kvs.ValueState_CONFIGURED))

	// Check output of `show span` command
	stdout.Reset()
	cmd = exec.Command("vppctl", "show", "int", "span")
	cmd.Stdout = &stdout
	err = cmd.Run()
	Expect(err).To(BeNil())
	output = stdout.String()
	s = space.ReplaceAllString(output, " ")
	Expect(s).To(Equal("Source Destination Device L2 tap1 tap0 ( rx) ( none) "))
}
