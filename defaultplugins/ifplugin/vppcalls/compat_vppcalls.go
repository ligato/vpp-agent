package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/af_packet"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/bfd"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/interfaces"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/ip"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/memif"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/tap"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/vxlan"
)

// CheckMsgCompatibilityForInterface checks if interface CRSs are compatible with VPP in runtime
func CheckMsgCompatibilityForInterface(vppChan *govppapi.Channel) error {
	msgs := []govppapi.Message{
		&memif.MemifCreate{},
		&memif.MemifCreateReply{},
		&memif.MemifDelete{},
		&memif.MemifDeleteReply{},
		&memif.MemifDump{},
		&memif.MemifDetails{},

		&vxlan.VxlanAddDelTunnel{},
		&vxlan.VxlanAddDelTunnelReply{},
		&vxlan.VxlanTunnelDump{},
		&vxlan.VxlanTunnelDetails{},

		&af_packet.AfPacketCreate{},
		&af_packet.AfPacketCreateReply{},
		&af_packet.AfPacketDelete{},
		&af_packet.AfPacketDeleteReply{},

		&tap.TapConnect{},
		&tap.TapConnectReply{},
		&tap.TapDelete{},
		&tap.TapDeleteReply{},
		&tap.SwInterfaceTapDump{},
		&tap.SwInterfaceTapDetails{},

		&interfaces.SwInterfaceSetFlags{},
		&interfaces.SwInterfaceSetFlagsReply{},
		&interfaces.SwInterfaceAddDelAddress{},
		&interfaces.SwInterfaceAddDelAddressReply{},
		&interfaces.SwInterfaceSetMacAddress{},
		&interfaces.SwInterfaceSetMacAddressReply{},
		&interfaces.SwInterfaceDetails{},

		&ip.IPAddressDump{},
		&ip.IPAddressDetails{},
	}
	err := vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		log.Error(err)
	}
	return err
}

// CheckMsgCompatibilityForBfd checks if bfd CRSs are compatible with VPP in runtime
func CheckMsgCompatibilityForBfd(vppChan *govppapi.Channel) error {
	msgs := []govppapi.Message{
		&bfd.BfdUDPAdd{},
		&bfd.BfdUDPAddReply{},
		&bfd.BfdUDPMod{},
		&bfd.BfdUDPModReply{},
		&bfd.BfdUDPDel{},
		&bfd.BfdUDPDelReply{},
		&bfd.BfdAuthSetKey{},
		&bfd.BfdAuthSetKeyReply{},
		&bfd.BfdAuthDelKey{},
		&bfd.BfdAuthDelKeyReply{},
	}
	err := vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		log.Error(err)
	}
	return err
}
