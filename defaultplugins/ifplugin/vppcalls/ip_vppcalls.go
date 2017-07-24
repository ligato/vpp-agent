package vppcalls

import (
	"fmt"
	"net"

	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/bin_api/interfaces"
)

// AddInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=1
func AddInterfaceIP(ifIdx uint32, addr *net.IPNet, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &interfaces.SwInterfaceAddDelAddress{}
	req.SwIfIndex = ifIdx
	req.IsAdd = 1

	prefix, _ := addr.Mask.Size()
	req.AddressLength = byte(prefix)

	v6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	if v6 {
		req.Address = []byte(addr.IP.To16())
		req.IsIpv6 = 1
	} else {
		req.Address = []byte(addr.IP.To4())
		req.IsIpv6 = 0
	}

	log.Debug("add req: IsIpv6: ", req.IsIpv6, " len(req.Address)=", len(req.Address))

	reply := &interfaces.SwInterfaceAddDelAddressReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("Adding IP address returned %d", reply.Retval)
	}
	log.WithFields(log.Fields{"IPaddress": addr.IP, "mask": addr.Mask, "ifIdx": ifIdx}).Debug("IP address added.")
	return nil

}

// DelInterfaceIP calls SwInterfaceAddDelAddress bin API with IsAdd=00
func DelInterfaceIP(ifIdx uint32, addr *net.IPNet, vppChan *govppapi.Channel) error {
	// prepare the message
	req := &interfaces.SwInterfaceAddDelAddress{}
	req.SwIfIndex = ifIdx
	req.IsAdd = 0

	prefix, _ := addr.Mask.Size()
	req.AddressLength = byte(prefix)

	v6, err := addrs.IsIPv6(addr.IP.String())
	if err != nil {
		return err
	}
	if v6 {
		req.Address = []byte(addr.IP.To16())
		req.IsIpv6 = 1
	} else {
		req.Address = []byte(addr.IP.To4())
		req.IsIpv6 = 0
	}

	log.Debug("del req: IsIpv6: ", req.IsIpv6, " len(req.Address)=", len(req.Address))

	// send the message
	reply := &interfaces.SwInterfaceAddDelAddressReply{}
	err = vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("Removing IP address returned %d", reply.Retval)
	}
	log.WithFields(log.Fields{"IPaddress": addr.IP, "mask": addr.Mask, "ifIdx": ifIdx}).Debug("IP address removed.")
	return nil

}
