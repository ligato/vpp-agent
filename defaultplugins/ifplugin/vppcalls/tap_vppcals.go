package vppcalls

import (
	"fmt"

	"errors"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/tap"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
)

// AddTapInterface calls TapConnect bin API
func AddTapInterface(tapIf *interfaces.Interfaces_Interface_Tap, vppChan *govppapi.Channel) (swIndex uint32, err error) {
	if tapIf == nil || tapIf.HostIfName == "" {
		return 0, errors.New("host interface name was not provided for the TAP interface")
	}

	// prepare the message
	req := &tap.TapConnect{}
	req.TapName = []byte(tapIf.HostIfName)

	req.UseRandomMac = 1

	reply := &tap.TapConnectReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return 0, err
	}

	if 0 != reply.Retval {
		return 0, fmt.Errorf("Add tap interface returned %d", reply.Retval)
	}
	return reply.SwIfIndex, nil
}

// DeleteTapInterface calls TapDelete bin API
func DeleteTapInterface(idx uint32, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &tap.TapDelete{}
	req.SwIfIndex = idx

	reply := &tap.TapDeleteReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("Deleting of interface returned %d", reply.Retval)
	}
	return nil
}
