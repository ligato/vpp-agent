package vppcalls

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/af_packet"
	intf "github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
)

// AddAfPacketInterface calls AfPacketCreate VPP binary API.
func AddAfPacketInterface(afPacketIntf *intf.Interfaces_Interface_Afpacket, vppChan *govppapi.Channel) (swIndex uint32, err error) {
	// prepare the message
	req := &af_packet.AfPacketCreate{}

	req.HostIfName = []byte(afPacketIntf.HostIfName)
	req.UseRandomHwAddr = 1

	reply := &af_packet.AfPacketCreateReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return 0, err
	}

	if 0 != reply.Retval {
		return 0, fmt.Errorf("Add af_packet interface returned %d", reply.Retval)
	}
	return reply.SwIfIndex, nil
}

// DeleteAfPacketInterface calls AfPacketDelete VPP binary API.
func DeleteAfPacketInterface(afPacketIntf *intf.Interfaces_Interface_Afpacket, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &af_packet.AfPacketDelete{}
	req.HostIfName = []byte(afPacketIntf.HostIfName)

	reply := &af_packet.AfPacketDeleteReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("Deleting of af_packet interface returned %d", reply.Retval)
	}
	return nil
}
