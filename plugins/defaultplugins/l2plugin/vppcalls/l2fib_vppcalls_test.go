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
	"testing"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/logrus"
	l2ba "github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bin_api/l2"
	. "github.com/onsi/gomega"
)

var vppReqChan = make(chan *govppapi.VppRequest)
var vppRepChan = make(chan *govppapi.VppReply)

type mockedMessageDecoder struct {
}

func (mockedMsgDecoder *mockedMessageDecoder) DecodeMsg(data []byte, msg govppapi.Message) error {
	return nil
}

var testDataInFib = []struct {
	mac    string
	bdID   uint32
	ifIdx  uint32
	bvi    bool
	static bool
}{
	{"FF:FF:FF:FF:FF:FF", 5, 55, true, true},
	{"FF:FF:FF:FF:FF:FF", 5, 55, false, true},
	{"FF:FF:FF:FF:FF:FF", 5, 55, true, false},
	{"FF:FF:FF:FF:FF:FF", 5, 55, false, false},
}

var createTestDatasOutFib = []*l2ba.L2fibAddDel{
	{BdID: 5, IsAdd: 1, SwIfIndex: 55, BviMac: 1, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 1},
	{BdID: 5, IsAdd: 1, SwIfIndex: 55, BviMac: 0, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 1},
	{BdID: 5, IsAdd: 1, SwIfIndex: 55, BviMac: 1, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 0},
	{BdID: 5, IsAdd: 1, SwIfIndex: 55, BviMac: 0, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, StaticMac: 0},
}

var deleteTestDataOutFib = &l2ba.L2fibAddDel{
	BdID: 5, IsAdd: 0, SwIfIndex: 55, Mac: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
}

func TestAdd(t *testing.T) {
	RegisterTestingT(t)
	l2FibVppCalls := NewL2FibVppCalls(&govppapi.Channel{ReqChan: vppReqChan, ReplyChan: vppRepChan,
		MsgDecoder: &mockedMessageDecoder{}}, nil)

	for idx := 0; idx < len(testDataInFib); idx++ {
		go l2FibVppCalls.Add(testDataInFib[idx].mac, testDataInFib[idx].bdID, testDataInFib[idx].ifIdx,
			testDataInFib[idx].bvi, testDataInFib[idx].static, nil, logrus.DefaultLogger())
		vppRequest := <-vppReqChan
		l2fibAddDel := vppRequest.Message.(*l2ba.L2fibAddDel)
		Expect(l2fibAddDel).To(Equal(createTestDatasOutFib[idx]))
	}
	Expect(l2FibVppCalls.waitingForReply.Len()).To(Equal(4))
}

func TestDelete(t *testing.T) {
	RegisterTestingT(t)
	l2FibVppCalls := NewL2FibVppCalls(&govppapi.Channel{ReqChan: vppReqChan, ReplyChan: vppRepChan,
		MsgDecoder: &mockedMessageDecoder{}}, nil)

	for idx := 0; idx < len(testDataInFib); idx++ {
		go l2FibVppCalls.Delete(testDataInFib[idx].mac, testDataInFib[idx].bdID, testDataInFib[idx].ifIdx,
			nil, logrus.DefaultLogger())
		vppRequest := <-vppReqChan
		l2fibAddDel := vppRequest.Message.(*l2ba.L2fibAddDel)
		Expect(l2fibAddDel).To(Equal(deleteTestDataOutFib))
	}
	Expect(l2FibVppCalls.waitingForReply.Len()).To(Equal(4))
}

//var counter uint8 = 0
//func dummyCallback(err error) {
//	counter++
//}

/**
currently commented because method WatchFIBReplies still run despite closing channel
*/
//TestWatchFIBReplies tests WatchFIBReplies method
//func TestWatchFIBReplies(t *testing.T) {
//	RegisterTestingT(t)
//	vppReq := &FibLogicalReq{Delete: false, BDIdx: 4, BVI: false, MAC: "FF:FF:FF:FF:FF:FF", Static: false,
//		SwIfIdx: 45, callback: dummyCallback}
//	go func() {
//		vppRepChan <- &govppapi.VppReply{MessageID: 4}
//		close(vppReqChan)
//	}()
//	l2FibVppCalls.waitingForReply.PushFront(vppReq)
//	l2FibVppCalls.WatchFIBReplies(logrus.DefaultLogger())
//}
