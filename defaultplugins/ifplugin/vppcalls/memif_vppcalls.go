package vppcalls

import (
	"fmt"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/memif"
	intf "github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
)

// AddMemifInterface calls MemifCreate bin API
func AddMemifInterface(memIntf *intf.Interfaces_Interface_Memif, vppChan *govppapi.Channel) (swIndex uint32, err error) {
	// prepare the message
	req := &memif.MemifCreate{}

	req.ID = memIntf.Id
	if memIntf.Master {
		req.Role = 0
	} else {
		req.Role = 1
	}
	req.Mode = uint8(memIntf.Mode)
	req.Secret = []byte(memIntf.Secret)
	req.SocketFilename = []byte(memIntf.SocketFilename)
	req.BufferSize = uint16(memIntf.BufferSize)
	req.RingSize = memIntf.RingSize
	req.RxQueues = uint8(memIntf.RxQueues)
	req.TxQueues = uint8(memIntf.TxQueues)

	/* TODO: temporary fix, waiting for https://gerrit.fd.io/r/#/c/7266/ */
	if req.RxQueues == 0 {
		req.RxQueues = 1
	}
	if req.TxQueues == 0 {
		req.TxQueues = 1
	}

	reply := &memif.MemifCreateReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)

	if err != nil {
		return 0, err
	}

	if 0 != reply.Retval {
		return 0, fmt.Errorf("Add memif interface returned %d", reply.Retval)
	}

	return reply.SwIfIndex, nil

}

// DeleteMemifInterface calls MemifDelete bin API
func DeleteMemifInterface(idx uint32, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &memif.MemifDelete{}
	req.SwIfIndex = idx

	reply := &memif.MemifDeleteReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Deleting of interface returned %d", reply.Retval)
	}
	return nil

}
