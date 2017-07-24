package vppcalls

import (
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/vpe"
)

// AddLoopbackInterface calls CreateLoopback bin API
func AddLoopbackInterface(name string, vppChan *govppapi.Channel) (swIndex uint32, err error) {
	req := &vpe.CreateLoopback{}
	reply := &vpe.CreateLoopbackReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return 0, err
	}

	if 0 != reply.Retval {
		return 0, fmt.Errorf("Add loopback interface returned %d", reply.Retval)
	}
	return reply.SwIfIndex, nil
}

// DeleteLoopbackInterface calls DeleteLoopback bin API
func DeleteLoopbackInterface(idx uint32, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &vpe.DeleteLoopback{}
	req.SwIfIndex = idx

	reply := &vpe.DeleteLoopbackReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("Deleting of loopback interface returned %d", reply.Retval)
	}
	return nil
}
