// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vppcalls

import (
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/idxvpp"
	bfd_api "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/bin_api/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"net"
	"time"
)

// AddBfdUDPSession adds new BFD session with authentication if available
func AddBfdUDPSession(bfdSession *bfd.SingleHopBFD_Session, swIfIndexes ifaceidx.SwIfIndex, bfdKeyIndexes idxvpp.NameToIdx,
	log logging.Logger, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// BfdUDPAdd time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Verify interface presence
	ifIdx, _, found := swIfIndexes.LookupIdx(bfdSession.Interface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdSession.Interface)
	}

	// Prepare the message
	req := &bfd_api.BfdUDPAdd{}

	// Base fields
	req.SwIfIndex = ifIdx
	req.DesiredMinTx = bfdSession.DesiredMinTxInterval
	req.RequiredMinRx = bfdSession.RequiredMinRxInterval
	req.DetectMult = uint8(bfdSession.DetectMultiplier)
	// IP
	isLocalIpv6, err := addrs.IsIPv6(bfdSession.SourceAddress)
	if err != nil {
		return err
	}
	isPeerIpv6, err := addrs.IsIPv6(bfdSession.DestinationAddress)
	if err != nil {
		return err
	}
	if isLocalIpv6 && isPeerIpv6 {
		req.IsIpv6 = 1
		req.LocalAddr = net.ParseIP(bfdSession.SourceAddress).To16()
		req.PeerAddr = net.ParseIP(bfdSession.DestinationAddress).To16()
	} else if !isLocalIpv6 && !isPeerIpv6 {
		req.IsIpv6 = 0
		req.LocalAddr = net.ParseIP(bfdSession.SourceAddress).To4()
		req.PeerAddr = net.ParseIP(bfdSession.DestinationAddress).To4()
	} else {
		return fmt.Errorf("different IP versions or missing IP address. Local: %v, Peer: %v",
			bfdSession.SourceAddress, bfdSession.DestinationAddress)
	}
	// Authentication
	if bfdSession.Authentication != nil {
		keyID := string(bfdSession.Authentication.KeyId)
		log.Infof("Setting up authentication with index %v", keyID)
		_, _, found := bfdKeyIndexes.LookupIdx(keyID)
		if found {
			req.IsAuthenticated = 1
			req.BfdKeyID = uint8(bfdSession.Authentication.KeyId)
			req.ConfKeyID = bfdSession.Authentication.AdvertisedKeyId
		} else {
			log.Infof("Authentication key %v not found", bfdSession.Authentication.KeyId)
			req.IsAuthenticated = 0
		}
	} else {
		req.IsAuthenticated = 0
	}

	reply := &bfd_api.BfdUDPAddReply{}
	err = vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("add BFD UDP session interface returned %d", reply.Retval)
	}

	return nil
}

// AddBfdUDPSessionFromDetails adds new BFD session with authentication if available
func AddBfdUDPSessionFromDetails(bfdSession *bfd_api.BfdUDPSessionDetails, bfdKeyIndexes idxvpp.NameToIdx, log logging.Logger,
	vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) error {
	// BfdUDPAdd time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message
	req := &bfd_api.BfdUDPAdd{}

	// Base fields
	req.SwIfIndex = bfdSession.SwIfIndex
	req.DesiredMinTx = bfdSession.DesiredMinTx
	req.RequiredMinRx = bfdSession.RequiredMinRx
	req.LocalAddr = bfdSession.LocalAddr
	req.PeerAddr = bfdSession.PeerAddr
	req.DetectMult = bfdSession.DetectMult
	req.IsIpv6 = bfdSession.IsIpv6
	// Authentication
	if bfdSession.IsAuthenticated != 0 {
		keyID := string(bfdSession.BfdKeyID)
		log.Infof("Setting up authentication with index %v", keyID)
		_, _, found := bfdKeyIndexes.LookupIdx(keyID)
		if found {
			req.IsAuthenticated = 1
			req.BfdKeyID = bfdSession.BfdKeyID
			req.ConfKeyID = bfdSession.ConfKeyID
		} else {
			log.Infof("Authentication key %v not found", bfdSession.BfdKeyID)
			req.IsAuthenticated = 0
		}
	} else {
		req.IsAuthenticated = 0
	}

	reply := &bfd_api.BfdUDPAddReply{}
	err := vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("add BFD UDP session interface returned %d", reply.Retval)
	}

	return nil
}

// ModifyBfdUDPSession modifies existing BFD session excluding authentication which cannot be changed this way
func ModifyBfdUDPSession(bfdSession *bfd.SingleHopBFD_Session, swIfIndexes ifaceidx.SwIfIndex, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) (err error) {
	// BfdUDPMod time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Find interface
	ifIdx, _, found := swIfIndexes.LookupIdx(bfdSession.Interface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdSession.Interface)
	}

	// Prepare the message
	req := &bfd_api.BfdUDPMod{}

	// Base fields
	req.SwIfIndex = ifIdx
	req.DesiredMinTx = bfdSession.DesiredMinTxInterval
	req.RequiredMinRx = bfdSession.RequiredMinRxInterval
	req.DetectMult = uint8(bfdSession.DetectMultiplier)
	// IP
	isLocalIpv6, err := addrs.IsIPv6(bfdSession.SourceAddress)
	if err != nil {
		return err
	}
	isPeerIpv6, err := addrs.IsIPv6(bfdSession.DestinationAddress)
	if err != nil {
		return err
	}
	if isLocalIpv6 && isPeerIpv6 {
		req.IsIpv6 = 1
		req.LocalAddr = net.ParseIP(bfdSession.SourceAddress).To16()
		req.PeerAddr = net.ParseIP(bfdSession.DestinationAddress).To16()
	} else if !isLocalIpv6 && !isPeerIpv6 {
		req.IsIpv6 = 0
		req.LocalAddr = net.ParseIP(bfdSession.SourceAddress).To4()
		req.PeerAddr = net.ParseIP(bfdSession.DestinationAddress).To4()
	} else {
		return fmt.Errorf("different IP versions or missing IP address. Local: %v, Peer: %v",
			bfdSession.SourceAddress, bfdSession.DestinationAddress)
	}

	reply := &bfd_api.BfdUDPModReply{}
	err = vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("update BFD UDP session interface returned %d", reply.Retval)
	}
	return nil
}

// DeleteBfdUDPSession removes existing BFD session
func DeleteBfdUDPSession(ifIndex uint32, sourceAddres string, destAddres string, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) (err error) {
	// BfdUDPDel time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message
	req := &bfd_api.BfdUDPDel{}
	req.SwIfIndex = ifIndex
	req.LocalAddr = net.ParseIP(sourceAddres).To4()
	req.PeerAddr = net.ParseIP(destAddres).To4()
	req.IsIpv6 = 0

	reply := &bfd_api.BfdUDPDelReply{}
	err = vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("delete BFD UDP session interface returned %d", reply.Retval)
	}

	return nil
}

// DumpBfdUDPSessionsWithID returns a list of BFD session's metadata
func DumpBfdUDPSessionsWithID(authKeyIndex uint32, swIfIndexes ifaceidx.SwIfIndex, bfdSessionIndexes idxvpp.NameToIdx,
	vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) ([]*bfd_api.BfdUDPSessionDetails, error) {
	// BfdUDPSessionDump time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message
	req := &bfd_api.BfdUDPSessionDump{}
	reqCtx := vppChannel.SendMultiRequest(req)
	var sessionIfacesWithID []*bfd_api.BfdUDPSessionDetails
	for {
		msg := &bfd_api.BfdUDPSessionDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return sessionIfacesWithID, err
		}
		// Not interested in sessions without auth key
		if msg.IsAuthenticated == 0 {
			continue
		}
		// Get interface name used in session
		ifName, _, found := swIfIndexes.LookupName(msg.SwIfIndex)
		if !found {
			continue
		}
		// Verify session exists
		_, _, found = bfdSessionIndexes.LookupIdx(ifName)
		if !found {
			continue
		}
		if msg.BfdKeyID == uint8(authKeyIndex) {
			sessionIfacesWithID = append(sessionIfacesWithID, msg)
		}
	}

	return sessionIfacesWithID, nil
}

// SetBfdUDPAuthenticationKey creates configures new authentication key
func SetBfdUDPAuthenticationKey(bfdKey *bfd.SingleHopBFD_Key, log logging.Logger, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) (err error) {
	// BfdAuthSetKey time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Convert authentication according to RFC5880
	var authentication uint8
	if bfdKey.AuthenticationType == 0 {
		authentication = 4 // Keyed SHA1
	} else if bfdKey.AuthenticationType == 1 {
		authentication = 5 // Meticulous keyed SHA1
	} else {
		log.Warnf("Provided authentication type not supported, setting up SHA1")
		authentication = 4
	}

	// Prepare the message
	req := &bfd_api.BfdAuthSetKey{}
	req.ConfKeyID = bfdKey.Id
	req.AuthType = authentication
	req.Key = []byte(bfdKey.Secret)
	req.KeyLen = uint8(len(bfdKey.Secret))

	reply := &bfd_api.BfdAuthSetKeyReply{}
	err = vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("set BFD authentication key returned %d", reply.Retval)
	}

	return nil
}

// DeleteBfdUDPAuthenticationKey removes authentication key
func DeleteBfdUDPAuthenticationKey(bfdKey *bfd.SingleHopBFD_Key, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) (err error) {
	// BfdAuthDelKey time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message
	req := &bfd_api.BfdAuthDelKey{}
	req.ConfKeyID = bfdKey.Id

	reply := &bfd_api.BfdAuthDelKeyReply{}
	err = vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("delete BFD authentication key returned %d", reply.Retval)
	}

	return nil
}

// AddBfdEchoFunction sets up echo function  for interface
func AddBfdEchoFunction(bfdInput *bfd.SingleHopBFD_EchoFunction, swIfIndexes ifaceidx.SwIfIndex, vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) (err error) {
	// BfdUDPSetEchoSource time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Verify interface presence
	ifIdx, _, found := swIfIndexes.LookupIdx(bfdInput.EchoSourceInterface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdInput.EchoSourceInterface)
	}

	// Prepare the message
	req := &bfd_api.BfdUDPSetEchoSource{}
	req.SwIfIndex = ifIdx

	reply := &bfd_api.BfdUDPSetEchoSourceReply{}
	err = vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("set BFD echo source returned %d", reply.Retval)
	}
	return nil
}

// DeleteBfdEchoFunction removes echo function
func DeleteBfdEchoFunction(vppChannel *govppapi.Channel, timeLog measure.StopWatchEntry) (err error) {
	// BfdUDPDelEchoSource time measurement
	start := time.Now()
	defer func() {
		if timeLog != nil {
			timeLog.LogTimeEntry(time.Since(start))
		}
	}()

	// Prepare the message
	req := &bfd_api.BfdUDPDelEchoSource{}

	reply := &bfd_api.BfdUDPDelEchoSourceReply{}
	err = vppChannel.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		return err
	}

	if 0 != reply.Retval {
		return fmt.Errorf("delete BFD echo source returned %d", reply.Retval)
	}
	return nil
}
