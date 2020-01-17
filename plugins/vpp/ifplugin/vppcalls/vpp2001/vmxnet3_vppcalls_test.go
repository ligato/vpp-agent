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

package vpp2001_test

import (
	"testing"

	. "github.com/onsi/gomega"
	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	vpp_vmxnet3 "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/vmxnet3"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestAddVmxNet3Interface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vmxnet3.Vmxnet3CreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	swIfIdx, err := ifHandler.AddVmxNet3("vmxnet3-face/be/1c/4", &ifs.VmxNet3Link{
		EnableElog: true,
		RxqSize:    2048,
		TxqSize:    512,
	})
	Expect(err).To(BeNil())
	Expect(swIfIdx).To(BeEquivalentTo(1))
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_vmxnet3.Vmxnet3Create)
		if ok {
			Expect(vppMsg.PciAddr).To(BeEquivalentTo(2629761742))
			Expect(vppMsg.EnableElog).To(BeEquivalentTo(1))
			Expect(vppMsg.RxqSize).To(BeEquivalentTo(2048))
			Expect(vppMsg.TxqSize).To(BeEquivalentTo(512))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestAddVmxNet3InterfacePCIErr(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vmxnet3.Vmxnet3CreateReply{
		SwIfIndex: 1,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	// Name in incorrect format
	_, err := ifHandler.AddVmxNet3("vmxnet3-a/14/19", nil)
	Expect(err).ToNot(BeNil())
}

func TestAddVmxNet3InterfaceRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vmxnet3.Vmxnet3CreateReply{
		Retval: 1,
	})

	_, err := ifHandler.AddVmxNet3("vmxnet3-a/14/19/1e", nil)
	Expect(err).ToNot(BeNil())
}

func TestDelVmxNet3Interface(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vmxnet3.Vmxnet3DeleteReply{
		Retval: 0,
	})
	ctx.MockVpp.MockReply(&vpp_ifs.SwInterfaceTagAddDelReply{})

	err := ifHandler.DeleteVmxNet3("vmxnet3-a/14/19/1e", 1)
	Expect(err).To(BeNil())
	var msgCheck bool
	for _, msg := range ctx.MockChannel.Msgs {
		vppMsg, ok := msg.(*vpp_vmxnet3.Vmxnet3Delete)
		if ok {
			Expect(vppMsg.SwIfIndex).To(BeEquivalentTo(1))
			msgCheck = true
		}
	}
	Expect(msgCheck).To(BeTrue())
}

func TestDelVmxNet3InterfaceRetval(t *testing.T) {
	ctx, ifHandler := ifTestSetup(t)
	defer ctx.TeardownTestCtx()

	ctx.MockVpp.MockReply(&vpp_vmxnet3.Vmxnet3DeleteReply{
		Retval: 1,
	})

	err := ifHandler.DeleteVmxNet3("vmxnet3-a/14/19/1e", 1)
	Expect(err).ToNot(BeNil())
}
