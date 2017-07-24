package vppcalls

import (
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/bin_api/vpe"
	"net"
	"strconv"
	"strings"
)

// VppAddArpTerminationTableEntry adds ARP termination entry
func VppAddArpTerminationTableEntry(bridgeDomainID uint32, mac string, ip string, vppChan *govppapi.Channel) error {
	log.Println("Adding arp termination entry")

	parsedMac, errMac := net.ParseMAC(mac)
	if errMac != nil {
		return fmt.Errorf("Error while parsing MAC address %v", mac)
	}

	// Convert ipv4 string to []byte
	ipv4Octets := strings.Split(ip, ".")
	var parsedIP []byte
	for _, ipv4Octet := range ipv4Octets {
		ipv4IntPart, err := strconv.ParseInt(ipv4Octet, 0, 32)
		if err != nil {
			return fmt.Errorf("Unable to parse ip address %s", ip)
		}
		parsedIP = append(parsedIP, byte(ipv4IntPart))
	}

	req := &vpe.BdIPMacAddDel{}
	req.BdID = bridgeDomainID
	req.IPAddress = parsedIP
	req.MacAddress = parsedMac
	req.IsIpv6 = 0
	req.IsAdd = 1

	reply := &vpe.BdIPMacAddDelReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Adding arp entry returned %d", reply.Retval)
	}
	log.WithFields(log.Fields{"Bridge domain": bridgeDomainID, "Mac": parsedMac, "Ip Address": ip}).Debug("Arp termination entry added.")

	return nil
}

// VppRemoveArpTerminationTableEntry removes ARP termination entry
func VppRemoveArpTerminationTableEntry(bdID uint32, mac string, ip string, vppChan *govppapi.Channel) error {
	log.Println("'Deleting' arp entry")

	parsedMac, errMac := net.ParseMAC(mac)
	if errMac != nil {
		return fmt.Errorf("Error while parsing MAC address %v", mac)
	}

	// Convert ipv4 string to []byte
	ipv4Octets := strings.Split(ip, ".")
	var parsedIP []byte
	for _, ipv4Octet := range ipv4Octets {
		ipv4IntPart, err := strconv.ParseInt(ipv4Octet, 0, 32)
		if err != nil {
			return fmt.Errorf("Unable to parse ip address %s", ip)
		}
		parsedIP = append(parsedIP, byte(ipv4IntPart))
	}

	req := &vpe.BdIPMacAddDel{}
	req.BdID = bdID
	req.MacAddress = parsedMac
	req.IPAddress = parsedIP
	req.IsAdd = 0

	reply := &vpe.BdIPMacAddDelReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}
	if 0 != reply.Retval {
		return fmt.Errorf("Deleting arp entry returned %d", reply.Retval)
	}
	log.WithFields(log.Fields{"bdID": bdID, "Mac": parsedMac, "Ip Address": ip}).Debug("Arp termination entry removed.")

	return nil
}
