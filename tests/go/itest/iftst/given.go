package iftst

import (
	"reflect"
	"strings"

	govppmock "git.fd.io/govpp.git/adapter/mock"
	"git.fd.io/govpp.git/adapter/mock/binapi"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/acl"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/af_packet"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/bfd"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/ipsec"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/memif"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/session"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/stats"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/tap"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/tapv2"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vxlan"
)

var swIfIndexSeq uint32

// RepliesSuccess replies with success binary API message.
func RepliesSuccess(vppMock *govppmock.VppAdapter) {
	vppMock.RegisterBinAPITypes(acl.Types)
	vppMock.RegisterBinAPITypes(af_packet.Types)
	vppMock.RegisterBinAPITypes(bfd.Types)
	vppMock.RegisterBinAPITypes(interfaces.Types)
	vppMock.RegisterBinAPITypes(ipsec.Types)
	vppMock.RegisterBinAPITypes(memif.Types)
	vppMock.RegisterBinAPITypes(session.Types)
	vppMock.RegisterBinAPITypes(stats.Types)
	vppMock.RegisterBinAPITypes(tap.Types)
	vppMock.RegisterBinAPITypes(tapv2.Types)
	vppMock.RegisterBinAPITypes(vpe.Types)
	vppMock.RegisterBinAPITypes(vxlan.Types)

	vppMock.MockReplyHandler(VppMockHandler(vppMock))
}

// VppMockHandler returns reply handler for mock adapter
func VppMockHandler(vppMock *govppmock.VppAdapter) govppmock.ReplyHandler {
	return func(request govppmock.MessageDTO) (reply []byte, msgID uint16, prepared bool) {
		reqName, found := vppMock.GetMsgNameByID(request.MsgID)
		if !found {
			logrus.DefaultLogger().Error("Not existing req msg name for MsgID=", request.MsgID)
			return reply, 0, false
		}
		logrus.DefaultLogger().Debug("MockReplyHandler ", request.MsgID, " ", reqName)

		//TODO refactor this to several funcs
		if strings.HasSuffix(reqName, "_dump") {
			// Do not reply to the dump message and reply to the following control_ping.
		} else {
			if replyMsg, msgID, ok := vppMock.ReplyFor(reqName); ok {
				val := reflect.ValueOf(replyMsg)
				valType := val.Type()
				if binapi.HasSwIfIdx(valType) {
					swIfIndexSeq++
					logrus.DefaultLogger().Debug("Succ default reply for ", reqName, " ", msgID, " sw_if_idx=", swIfIndexSeq)
					binapi.SetSwIfIdx(val, swIfIndexSeq)
				} else {
					logrus.DefaultLogger().Debug("Succ default reply for ", reqName, " ", msgID)
				}

				reply, err := vppMock.ReplyBytes(request, replyMsg)
				if err == nil {
					return reply, msgID, true
				}
				logrus.DefaultLogger().Error("Error creating bytes ", err)
			} else {
				logrus.DefaultLogger().Warn("No default reply for ", reqName, ", ", request.MsgID)
			}
		}

		return reply, 0, false
	}
}
