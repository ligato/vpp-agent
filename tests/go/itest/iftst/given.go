package iftst

import (
	//"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin"
	"reflect"

	"strings"

	govppmock "git.fd.io/govpp.git/adapter/mock"
	"git.fd.io/govpp.git/adapter/mock/binapi"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/af_packet"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/bfd"
	interfaces_bin "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/memif"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/tap"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/vpe"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/vxlan"
)

var swIfIndexSeq uint32

// RepliesSuccess replies with success binary API message.
func RepliesSuccess(vppMock *govppmock.VppAdapter) {
	vppMock.RegisterBinAPITypes(interfaces_bin.Types)
	vppMock.RegisterBinAPITypes(memif.Types)
	vppMock.RegisterBinAPITypes(tap.Types)
	vppMock.RegisterBinAPITypes(af_packet.Types)
	vppMock.RegisterBinAPITypes(vpe.Types)
	vppMock.RegisterBinAPITypes(vxlan.Types)
	vppMock.RegisterBinAPITypes(bfd.Types)

	vppMock.MockReplyHandler(func(request govppmock.MessageDTO) (reply []byte, msgID uint16, prepared bool) {
		reqName, found := vppMock.GetMsgNameByID(request.MsgID)
		if !found {
			logroot.StandardLogger().Error("Not existing req msg name for MsgID=", request.MsgID)
			return reply, 0, false
		}
		logroot.StandardLogger().Debug("MockReplyHandler ", request.MsgID, " ", reqName)

		//TODO refactor this to several funcs
		if strings.HasSuffix(reqName, "_dump") {
			// Do not reply to the dump message and reply to the following control_ping.
		} else {
			if replyMsg, msgID, ok := vppMock.ReplyFor(reqName); ok {
				val := reflect.ValueOf(replyMsg)
				valType := val.Type()
				if binapi.HasSwIfIdx(valType) {
					swIfIndexSeq++
					logroot.StandardLogger().Debug("Succ default reply for ", reqName, " ", msgID, " sw_if_idx=", swIfIndexSeq)
					binapi.SetSwIfIdx(val, swIfIndexSeq)
				} else {
					logroot.StandardLogger().Debug("Succ default reply for ", reqName, " ", msgID)
				}

				reply, err := vppMock.ReplyBytes(request, replyMsg)
				if err == nil {
					return reply, msgID, true
				}
				logroot.StandardLogger().Error("Error creating bytes ", err)
			} else {
				logroot.StandardLogger().Info("No default reply for ", reqName, ", ", request.MsgID)
			}
		}

		return reply, 0, false
	})
}
