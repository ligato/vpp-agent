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
	"net"
	"strings"
	"testing"

	"git.fd.io/govpp.git/adapter/mock"
	"git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	bfd_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

var bfdKeyNames = []string{"key1", "key2"}
var secret = "bfd-key-secret"

/* BFD configurator init and close */

// Test init function
func TestBfdConfiguratorInit(t *testing.T) {
	RegisterTestingT(t)
	connection, err := core.Connect(&mock.VppAdapter{})
	Expect(err).To(BeNil())
	plugin := &BFDConfigurator{
		Log:      logrus.DefaultLogger(),
		GoVppmux: connection,
	}
	bfdSessionsIndexes := nametoidx.NewNameToIdx(plugin.Log, "bfds-test", "bfds", nil)
	bfdKeysIndexes := nametoidx.NewNameToIdx(plugin.Log, "bfdk-test", "bfdk", nil)
	bfdEchoFunctionIndex := nametoidx.NewNameToIdx(plugin.Log, "echo-test", "echo", nil)
	bfdRemovedAuthIndex := nametoidx.NewNameToIdx(plugin.Log, "bfdr-test", "bfdr", nil)
	err = plugin.Init(bfdSessionsIndexes, bfdKeysIndexes, bfdEchoFunctionIndex, bfdRemovedAuthIndex)
	Expect(err).To(BeNil())
	Expect(plugin.vppChan).ToNot(BeNil())
	Expect(plugin.bfdSessionsIndexes).ToNot(BeNil())
	Expect(plugin.bfdKeysIndexes).ToNot(BeNil())
	Expect(plugin.bfdEchoFunctionIndex).ToNot(BeNil())
	Expect(plugin.bfdRemovedAuthIndex).ToNot(BeNil())
	err = plugin.Close()
	Expect(err).To(BeNil())
	connection.Disconnect()
}

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

// BFD session dump
func TestBfdConfiguratorDumpBfdSessions(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSessionDetails{
		SwIfIndex: 1,
		LocalAddr: net.ParseIP(ipAddresses[0]).To4(),
		PeerAddr:  net.ParseIP(ipAddresses[1]).To4(),
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	// Register
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	// Test bfd session dump
	sessions, err := plugin.DumpBfdSessions()
	Expect(err).To(BeNil())
	Expect(sessions).To(HaveLen(1))
	Expect(sessions[0].Interface).To(BeEquivalentTo(ifNames[0]))
	Expect(sessions[0].SourceAddress).To(BeEquivalentTo(ipAddresses[0]))
	Expect(sessions[0].DestinationAddress).To(BeEquivalentTo(ipAddresses[1]))

}

// Configure BFD authentication key
func TestBfdConfiguratorSetAuthKey(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{})
	// Data
	data := getTestBfdAuthKey(bfdKeyNames[0], secret, 1, 1, bfd.SingleHopBFD_Key_KEYED_SHA1)
	// Test key configuration
	err = plugin.ConfigureBfdAuthKey(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.bfdKeysIndexes.LookupIdx(authKeyIdentifier(data.Id))
	Expect(found).To(BeTrue())
}

// Configure BFD authentication key with error return value
func TestBfdConfiguratorSetAuthKeyError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{
		Retval: 1,
	})
	// Data
	data := getTestBfdAuthKey(bfdKeyNames[0], secret, 1, 1, bfd.SingleHopBFD_Key_KEYED_SHA1)
	// Test key configuration
	err = plugin.ConfigureBfdAuthKey(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.bfdKeysIndexes.LookupIdx(authKeyIdentifier(data.Id))
	Expect(found).To(BeFalse())
}

// Modify BFD authentication key which is not used in any session
func TestBfdConfiguratorModifyUnusedAuthKey(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})       // Session dump
	ctx.MockVpp.MockReply(&bfd_api.BfdAuthDelKeyReply{}) // Authentication key delete/create
	ctx.MockVpp.MockReply(&bfd_api.BfdAuthSetKeyReply{})
	// Data
	oldData := getTestBfdAuthKey(bfdKeyNames[0], secret, 1, 1, bfd.SingleHopBFD_Key_KEYED_SHA1)
	newData := getTestBfdAuthKey(bfdKeyNames[0], secret, 1, 1, bfd.SingleHopBFD_Key_METICULOUS_KEYED_SHA1)
	// Register
	plugin.bfdKeysIndexes.RegisterName(authKeyIdentifier(oldData.Id), 1, nil)
	// Test key modification
	err = plugin.ModifyBfdAuthKey(oldData, newData)
	Expect(err).To(BeNil())
}

// Modify BFD authentication key which is used in session todo control ping reply terminates mockvpp replies
func TestBfdConfiguratorModifyUsedAuthKey(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply handler
	ctx.MockVpp.RegisterBinAPITypes(bfd_api.Types)
	ctx.MockVpp.RegisterBinAPITypes(vpe.Types)
	ctx.MockVpp.MockReplyHandler(bfdVppMockHandler(ctx.MockVpp))
	// Data
	oldData := getTestBfdAuthKey(bfdKeyNames[0], secret, 1, 1, bfd.SingleHopBFD_Key_KEYED_SHA1)
	newData := getTestBfdAuthKey(bfdKeyNames[0], secret, 1, 1, bfd.SingleHopBFD_Key_METICULOUS_KEYED_SHA1)
	// Register
	plugin.bfdKeysIndexes.RegisterName(authKeyIdentifier(oldData.Id), 1, nil)
	// Test key modification
	err = plugin.ModifyBfdAuthKey(oldData, newData)
	Expect(err).To(BeNil())
}

// Delete BFD authentication key which is not used in any session
func TestBfdConfiguratorDeleteUnusedAuthKey(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})       // Session dump
	ctx.MockVpp.MockReply(&bfd_api.BfdAuthDelKeyReply{}) // Authentication key delete
	// Data
	data := getTestBfdAuthKey(bfdKeyNames[0], secret, 1, 1, bfd.SingleHopBFD_Key_KEYED_SHA1)
	// Register
	plugin.bfdKeysIndexes.RegisterName(authKeyIdentifier(data.Id), 1, nil)
	// Test key modification
	err = plugin.DeleteBfdAuthKey(data)
	Expect(err).To(BeNil())
}

// Delete BFD authentication key which is used in session todo control ping reply terminates mockvpp replies
func TestBfdConfiguratorDeleteUsedAuthKey(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply handler
	ctx.MockVpp.RegisterBinAPITypes(bfd_api.Types)
	ctx.MockVpp.RegisterBinAPITypes(vpe.Types)
	ctx.MockVpp.MockReplyHandler(bfdVppMockHandler(ctx.MockVpp))
	// Data
	data := getTestBfdAuthKey(bfdKeyNames[0], secret, 1, 1, bfd.SingleHopBFD_Key_KEYED_SHA1)
	// Register
	plugin.bfdKeysIndexes.RegisterName(authKeyIdentifier(data.Id), 1, nil)
	// Test key modification
	err = plugin.DeleteBfdAuthKey(data)
	Expect(err).To(BeNil())
}

// Dump BFD authentication key
func TestBfdConfiguratorDumpAuthKey(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdAuthKeysDetails{
		ConfKeyID: 1,
		AuthType:  4, // Means KEYED SHA1
	})
	ctx.MockVpp.MockReply(&bfd_api.BfdAuthKeysDetails{
		ConfKeyID: 2,
		AuthType:  1, // Any other number is METICULOUS KEYED SHA1
	})
	ctx.MockVpp.MockReply(&vpe.ControlPingReply{})
	// Test authentication key dump
	keys, err := plugin.DumpBFDAuthKeys()
	Expect(err).To(BeNil())
	Expect(keys).To(HaveLen(2))
	Expect(keys[0].AuthenticationType).To(BeEquivalentTo(bfd.SingleHopBFD_Key_KEYED_SHA1))
	Expect(keys[1].AuthenticationType).To(BeEquivalentTo(bfd.SingleHopBFD_Key_METICULOUS_KEYED_SHA1))
}

// Configure BFD echo function create/modify/delete
func TestBfdConfiguratorEchoFunction(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, swIfIndices := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Reply set
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPSetEchoSourceReply{})
	ctx.MockVpp.MockReply(&bfd_api.BfdUDPDelEchoSourceReply{})
	// Data
	data := getTestBfdEchoFunction(ifNames[0])
	//Registration
	swIfIndices.RegisterName(ifNames[0], 1, nil)
	// Test Echo function create
	err = plugin.ConfigureBfdEchoFunction(data)
	Expect(err).To(BeNil())
	_, _, found := plugin.bfdEchoFunctionIndex.LookupIdx(data.EchoSourceInterface)
	Expect(found).To(BeTrue())
	// Test Echo function modify
	err = plugin.ModifyBfdEchoFunction(data, data)
	Expect(err).To(BeNil())
	// Test echo function delete
	err = plugin.DeleteBfdEchoFunction(data)
	Expect(err).To(BeNil())
	_, _, found = plugin.bfdEchoFunctionIndex.LookupIdx(data.EchoSourceInterface)
	Expect(found).To(BeFalse())
}

// Configure BFD echo function create with non-existing interface
func TestBfdConfiguratorEchoFunctionNoInterfaceError(t *testing.T) {
	var err error
	// Setup
	ctx, plugin, _ := bfdTestSetup(t)
	defer bfdTestTeardown(ctx, plugin)
	// Data
	data := getTestBfdEchoFunction(ifNames[0])
	// Test Echo function create
	err = plugin.ConfigureBfdEchoFunction(data)
	Expect(err).ToNot(BeNil())
	_, _, found := plugin.bfdEchoFunctionIndex.LookupIdx(data.EchoSourceInterface)
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

func bfdVppMockHandler(vppMock *mock.VppAdapter) mock.ReplyHandler {
	var sendControlPing bool
	return func(request mock.MessageDTO) (reply []byte, msgID uint16, prepared bool) {
		logrus.DefaultLogger().Errorf("recived request %v", request.MsgName)
		if sendControlPing {
			sendControlPing = false
			data := &vpe.ControlPingReply{}
			reply, err := vppMock.ReplyBytes(request, data)
			Expect(err).To(BeNil())
			msgID, err := vppMock.GetMsgID(data.GetMessageName(), data.GetCrcString())
			Expect(err).To(BeNil())
			return reply, msgID, true
		}
		if strings.HasSuffix(request.MsgName, "_dump") {
			// Send control ping after first iteration
			sendControlPing = true
			data := &bfd_api.BfdUDPSessionDetails{
				SwIfIndex:       1,
				LocalAddr:       net.ParseIP(ipAddresses[0]).To4(),
				PeerAddr:        net.ParseIP(ipAddresses[1]).To4(),
				IsAuthenticated: 1,
				BfdKeyID:        1,
			}
			reply, err := vppMock.ReplyBytes(request, data)
			Expect(err).To(BeNil())
			msgID, err := vppMock.GetMsgID(data.GetMessageName(), data.GetCrcString())
			Expect(err).To(BeNil())
			return reply, msgID, true
		} else {
			if replyMsg, msgID, ok := vppMock.ReplyFor(request.MsgName); ok {
				reply, err := vppMock.ReplyBytes(request, replyMsg)
				Expect(err).To(BeNil())
				return reply, msgID, true
			}
		}

		return reply, 0, false
	}
}

/* BFD Test Data */

func getTestBfdSession(ifName, srcAddr string) *bfd.SingleHopBFD_Session {
	return &bfd.SingleHopBFD_Session{
		Interface:          ifName,
		SourceAddress:      srcAddr,
		DestinationAddress: ipAddresses[4],
	}
}

func getTestBfdAuthKey(name, secret string, keyIdx, id uint32, keyType bfd.SingleHopBFD_Key_AuthenticationType) *bfd.SingleHopBFD_Key {
	return &bfd.SingleHopBFD_Key{
		Name:               name,
		AuthKeyIndex:       keyIdx,
		Id:                 id,
		AuthenticationType: keyType,
		Secret:             secret,
	}
}

func getTestBfdEchoFunction(ifName string) *bfd.SingleHopBFD_EchoFunction {
	return &bfd.SingleHopBFD_EchoFunction{
		Name:                "echo",
		EchoSourceInterface: ifName,
	}
}

func getTestInterface(ip []string) *interfaces.Interfaces_Interface {
	return &interfaces.Interfaces_Interface{
		Name:        ifNames[0],
		IpAddresses: ip,
	}
}
