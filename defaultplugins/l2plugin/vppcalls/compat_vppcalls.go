package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	l2ba "github.com/ligato/vpp-agent/defaultplugins/l2plugin/bin_api/l2"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/bin_api/vpe"
)

// CheckMsgCompatibilityForBridgeDomains checks if CRSs are compatible with VPP in runtime
func CheckMsgCompatibilityForBridgeDomains(vppChan *govppapi.Channel) error {
	msgs := []govppapi.Message{
		&l2ba.BridgeDomainAddDel{},
		&l2ba.BridgeDomainAddDelReply{},
		&l2ba.L2fibAddDel{},
		&l2ba.L2fibAddDelReply{},
		&vpe.BdIPMacAddDel{},
		&vpe.BdIPMacAddDelReply{},
		&vpe.SwInterfaceSetL2Bridge{},
		&vpe.SwInterfaceSetL2BridgeReply{},
	}
	err := vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		log.Error(err)
	}
	return err
}

// CheckMsgCompatibilityForL2FIB checks if CRSs are compatible with VPP in runtime
func CheckMsgCompatibilityForL2FIB(vppChan *govppapi.Channel) error {
	msgs := []govppapi.Message{
		&l2ba.BridgeDomainDump{},
		&l2ba.BridgeDomainDetails{},
		&l2ba.L2FibTableDump{},
		&l2ba.L2FibTableDetails{},
		&l2ba.L2fibAddDel{},
		&l2ba.L2fibAddDelReply{},
	}
	err := vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		log.Error(err)
	}
	return err
}

// CheckMsgCompatibilityForL2XConnect checks if CRSs are compatible with VPP in runtime
func CheckMsgCompatibilityForL2XConnect(vppChan *govppapi.Channel) error {
	msgs := []govppapi.Message{
		&l2ba.L2XconnectDump{},
		&l2ba.L2XconnectDetails{},
		&vpe.SwInterfaceSetL2Xconnect{},
		&vpe.SwInterfaceSetL2XconnectReply{},
	}
	err := vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		log.Error(err)
	}
	return err
}
