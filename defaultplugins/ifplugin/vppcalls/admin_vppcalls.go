package vppcalls

import (
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/interfaces"
)

// InterfaceAdminDown calls binary API SwInterfaceSetFlagsReply with AdminUpDown=0
func InterfaceAdminDown(ifIdx uint32, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &interfaces.SwInterfaceSetFlags{}
	req.SwIfIndex = ifIdx
	req.AdminUpDown = 0

	reply := &interfaces.SwInterfaceSetFlagsReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Setting of interface flags returned %d", reply.Retval)
	}
	return nil

}

// InterfaceAdminUp calls binary API SwInterfaceSetFlagsReply with AdminUpDown=1
func InterfaceAdminUp(ifIdx uint32, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &interfaces.SwInterfaceSetFlags{}
	req.SwIfIndex = ifIdx
	req.AdminUpDown = 1

	reply := &interfaces.SwInterfaceSetFlagsReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Setting of interface flags returned %d", reply.Retval)
	}
	return nil

}
