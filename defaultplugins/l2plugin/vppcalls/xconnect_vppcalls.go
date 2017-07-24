package vppcalls

import (
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/bin_api/vpe"
)

// VppSetL2XConnect creates xConnect between two existing interfaces
func VppSetL2XConnect(receiveIfaceIndex uint32, transmitIfaceIndex uint32, vppChan *govppapi.Channel) error {
	log.Debug("Setting up L2 xConnect pair for ", transmitIfaceIndex, receiveIfaceIndex)

	req := &vpe.SwInterfaceSetL2Xconnect{}
	req.TxSwIfIndex = transmitIfaceIndex
	req.RxSwIfIndex = receiveIfaceIndex
	req.Enable = 1

	reply := &vpe.SwInterfaceSetL2XconnectReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Creating xConnect returned %d", reply.Retval)
	}

	log.WithFields(log.Fields{"RxIface": receiveIfaceIndex, "TxIface": transmitIfaceIndex}).Debug("L2xConnect created.")
	return nil
}

// VppUnsetL2XConnect removes xConnect between two interfaces
func VppUnsetL2XConnect(receiveIfaceIndex uint32, transmitIfaceIndex uint32, vppChan *govppapi.Channel) error {
	log.Debug("Setting up L2 xConnect pair for ", transmitIfaceIndex, receiveIfaceIndex)

	req := &vpe.SwInterfaceSetL2Xconnect{}
	req.RxSwIfIndex = receiveIfaceIndex
	req.TxSwIfIndex = transmitIfaceIndex
	req.Enable = 0

	reply := &vpe.SwInterfaceSetL2XconnectReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Removing xConnect returned %d", reply.Retval)
	}

	log.WithFields(log.Fields{"RxIface": receiveIfaceIndex, "TxIface": transmitIfaceIndex}).Debug("L2xConnect removed.")
	return nil
}
