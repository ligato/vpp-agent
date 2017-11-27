// Copyright (c) 2017 Cisco and/or its affiliates.
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

package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/vppcalls/test/impl"
	"testing"
	"github.com/stretchr/testify/assert"
)

const (
	dummyBridgeDomain uint32 = 4
	dummyMACAddress = "FF:FF:FF:FF:FF:FF"
	dummyIPAddress = "192.168.4.4"
)


//TestVppAddArpTerminationTableEntry tests VppAddArpTerminationTableEntry
func TestVppAddArpTerminationTableEntry(t *testing.T) {
	mockedChannel := &impl.MockedChannel{Channel:govppapi.Channel{}}
	VppAddArpTerminationTableEntry(dummyBridgeDomain, dummyMACAddress, dummyIPAddress, &impl.MockedLogger{},
		mockedChannel,nil)

	var vppMsg l2ba.BdIPMacAddDel
	switch t := mockedChannel.Msg.(type) {
	case *l2ba.BdIPMacAddDel:
		vppMsg = *t
	}

	assert.NotNil(t, vppMsg)

	//check values which will be send to VPP
	assert.Equal(t, dummyBridgeDomain, vppMsg.BdID)
	assert.Equal(t, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, vppMsg.MacAddress)
	assert.Equal(t, []byte{192, 168, 4, 4}, vppMsg.IPAddress)
	assert.Equal(t, uint8 (0), vppMsg.IsIpv6)
	assert.Equal(t, uint8 (1), vppMsg.IsAdd)
}

// TestVppRemoveArpTerminationTableEntry tests VppRemoveArpTerminationTableEntry method
func TestVppRemoveArpTerminationTableEntry(t *testing.T) {
	mockedChannel := &impl.MockedChannel{Channel:govppapi.Channel{}}
	VppAddArpTerminationTableEntry(dummyBridgeDomain, dummyMACAddress, dummyIPAddress, &impl.MockedLogger{},
		mockedChannel,nil)

}
