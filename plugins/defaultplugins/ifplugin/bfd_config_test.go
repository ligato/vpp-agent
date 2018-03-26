// Copyright (c) 2018 Cisco and/or its affiliates.
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

package ifplugin

import (
	"testing"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	bfd_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

/* BFD Sessions */

// Configure BFD session without interface
func TestBfdConfiguratorConfigureSessionNoInterfaceError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Data
	data := getTestBfdSession(ifNames[0], ipAddresses[0])
	// Test configure BFD session without interface
	err = plugin.ConfigureBfdSession(data)
	Expect(err).ToNot(BeNil())
}

// Configure BFD session while interface metadata is missing
func TestBfdConfiguratorConfigureSessionNoInterfaceMetaError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Data
	data := getTestBfdSession(ifNames[0], ipAddresses[0])
	// Register
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	// Test configure BFD session
	err = plugin.ConfigureBfdSession(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.bfdSessionsIndexes.LookupIdx(data.Interface)
	Expect(found).To(BeFalse())
}

// Configure BFD session while source IP does not match with interface IP
func TestBfdConfiguratorConfigureSessionSrcDoNotMatch(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Data
	data := getTestBfdSession(ifNames[0], ipAddresses[1])
	// Register
	var addresses []string
	swIfIndices.RegisterName(ifNames[0], 1, getTestInterface(append(addresses, netAddresses[0])))
	// Test configure BFD session
	err = plugin.ConfigureBfdSession(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.bfdSessionsIndexes.LookupIdx(data.Interface)
	Expect(found).To(BeFalse())
}

// Configure BFD session
func TestBfdConfiguratorConfigureSession(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})
	// Data
	data := getTestBfdSession(ifNames[0], ipAddresses[1])
	// Register
	var addresses []string
	swIfIndices.RegisterName(ifNames[0], 1, getTestInterface(append(addresses, netAddresses[1])))
	// Test configure BFD session
	err = plugin.ConfigureBfdSession(data)
	Expect(err).To(BeNil())
	_, meta, found := plugin.bfdSessionsIndexes.LookupIdx(data.Interface)
	Expect(found).To(BeTrue())
	Expect(meta).To(BeNil())
}

// Modify BFD session without interface
func TestBfdConfiguratorModifySessionNoInterfaceError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Data
	oldData := getTestBfdSession(ifNames[0], ipAddresses[0])
	newData := getTestBfdSession(ifNames[1], ipAddresses[1])
	// Register
	var addresses []string
	swIfIndices.RegisterName(ifNames[0], 1, getTestInterface(append(addresses, netAddresses[0])))
	// Test modify BFD session
	err = plugin.ModifyBfdSession(oldData, newData)
	Expect(err).ToNot(BeNil())
}

// Modify BFD session without interface metadata
func TestBfdConfiguratorModifySessionNoInterfaceMeta(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Data
	oldData := getTestBfdSession(ifNames[0], ipAddresses[0])
	newData := getTestBfdSession(ifNames[1], ipAddresses[1])
	// Register
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	// Test modify BFD session
	err = plugin.ModifyBfdSession(oldData, newData)
	Expect(err).ToNot(BeNil())
}

// Modify BFD session where source IP does not match
func TestBfdConfiguratorModifySessionSrcDoNotMatchError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Data
	oldData := getTestBfdSession(ifNames[0], ipAddresses[0])
	newData := getTestBfdSession(ifNames[0], ipAddresses[1])
	// Register
	plugin.bfdSessionsIndexes.RegisterName(oldData.Interface, 1, nil)
	var addresses []string
	swIfIndices.RegisterName(ifNames[0], 1, getTestInterface(append(addresses, netAddresses[2])))
	// Test modify BFD session
	err = plugin.ModifyBfdSession(oldData, newData)
	Expect(err).ToNot(BeNil())
}

// Modify BFD session without previous data
func TestBfdConfiguratorModifySessionNoPrevious(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPAddReply{})
	// Data
	oldData := getTestBfdSession(ifNames[0], ipAddresses[0])
	newData := getTestBfdSession(ifNames[1], ipAddresses[1])
	// Register
	var addresses []string
	swIfIndices.RegisterName(ifNames[1], 1, getTestInterface(append(addresses, netAddresses[1])))
	// Test modify BFD session
	err = plugin.ModifyBfdSession(oldData, newData)
	Expect(err).To(BeNil())
	_, meta, found := plugin.bfdSessionsIndexes.LookupIdx(newData.Interface)
	Expect(found).To(BeTrue())
	Expect(meta).To(BeNil())
}

// Modify BFD session different source addresses
func TestBfdConfiguratorModifySessionSrcAddrDiffError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})
	// Data
	oldData := getTestBfdSession(ifNames[0], ipAddresses[0])
	newData := getTestBfdSession(ifNames[0], ipAddresses[1])
	// Register
	plugin.bfdSessionsIndexes.RegisterName(oldData.Interface, 1, nil)
	var addresses []string
	swIfIndices.RegisterName(ifNames[0], 1, getTestInterface(append(addresses, netAddresses[0], netAddresses[1])))
	// Test modify BFD session
	err = plugin.ModifyBfdSession(oldData, newData)
	Expect(err).ToNot(BeNil())
}

// Modify BFD session
func TestBfdConfiguratorModifySession(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPModReply{})
	// Data
	oldData := getTestBfdSession(ifNames[0], ipAddresses[0])
	newData := getTestBfdSession(ifNames[0], ipAddresses[0])
	// Register
	plugin.bfdSessionsIndexes.RegisterName(oldData.Interface, 1, nil)
	var addresses []string
	swIfIndices.RegisterName(ifNames[0], 1, getTestInterface(append(addresses, netAddresses[0])))
	// Test modify BFD session
	err = plugin.ModifyBfdSession(oldData, newData)
	Expect(err).To(BeNil())
	_, meta, found := plugin.bfdSessionsIndexes.LookupIdx(newData.Interface)
	Expect(found).To(BeTrue())
	Expect(meta).To(BeNil())
	err = plugin.Close()
	Expect(err).To(BeNil())
}

// Test delete BFD session no interface
func TestBfdConfiguratorDeleteSessionNoInterfaceError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Data
	data := getTestBfdSession(ifNames[0], ipAddresses[0])
	// Register
	plugin.bfdSessionsIndexes.RegisterName(data.Interface, 1, nil)
	// Modify BFD session
	err = plugin.DeleteBfdSession(data)
	Expect(err).ToNot(BeNil())
}

// Test delete BFD session
func TestBfdConfiguratorDeleteSession(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPDelReply{})
	// Data
	data := getTestBfdSession(ifNames[0], ipAddresses[0])
	// Register
	plugin.bfdSessionsIndexes.RegisterName(data.Interface, 1, nil)
	var addresses []string
	swIfIndices.RegisterName(ifNames[0], 1, getTestInterface(append(addresses, netAddresses[0])))
	// Modify BFD session
	err = plugin.DeleteBfdSession(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.bfdSessionsIndexes.LookupIdx(data.Interface)
	Expect(found).To(BeFalse())
}

/* BFD Test Setup */

func bfdTestSetup(t *testing.T) (*vppcallmock.TestCtx, *BFDConfigurator, ifaceidx.SwIfIndexRW) {
	ctx := vppcallmock.SetupTestCtx(t)
	// Logger
	log := logrus.DefaultLogger()
	log.SetLevel(logging.DebugLevel)

	// Interface indices
	swIfIndices := ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(log, "bfd-configurator-test", "bfd", nil))

	return ctx, &BFDConfigurator{
		Log:                  log,
		SwIfIndexes:          swIfIndices,
		bfdSessionsIndexes:   nametoidx.NewNameToIdx(log, "bfds-test", "bfds", nil),
		bfdKeysIndexes:       nametoidx.NewNameToIdx(log, "bfdk-test", "bfdk", nil),
		bfdEchoFunctionIndex: nametoidx.NewNameToIdx(log, "echo-test", "echo", nil),
		bfdRemovedAuthIndex:  nametoidx.NewNameToIdx(log, "bfdr-test", "bfdr", nil),
		vppChan:              ctx.MockChannel,
	}, swIfIndices
}

func bfdTestTeardown(ctx *vppcallmock.TestCtx, plugin *BFDConfigurator) {
	ctx.TeardownTestCtx()
	err := plugin.Close()
	Expect(err).To(BeNil())
}

/* BFD Test Data */

func getTestBfdSession(ifName, srcAddr string) *bfd.SingleHopBFD_Session {
	return &bfd.SingleHopBFD_Session{
		Interface:          ifName,
		SourceAddress:      srcAddr,
		DestinationAddress: ipAddresses[4],
	}
}

func getTestInterface(ip []string) *interfaces.Interfaces_Interface {
	return &interfaces.Interfaces_Interface{
		Name:        ifNames[0],
		IpAddresses: ip,
	}
}
