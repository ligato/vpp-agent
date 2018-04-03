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
	"net"
	"time"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/idxvpp"
	bfd_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
)

// AddBfdUDPSession adds new BFD session with authentication if available.
func AddBfdUDPSession(bfdSess *bfd.SingleHopBFD_Session, ifIdx uint32, bfdKeyIndexes idxvpp.NameToIdx,
	log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdUDPAdd{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bfd_api.BfdUDPAdd{
		SwIfIndex:     ifIdx,
		DesiredMinTx:  bfdSess.DesiredMinTxInterval,
		RequiredMinRx: bfdSess.RequiredMinRxInterval,
		DetectMult:    uint8(bfdSess.DetectMultiplier),
	}

	isLocalIpv6, err := addrs.IsIPv6(bfdSess.SourceAddress)
	if err != nil {
		return err
	}
	isPeerIpv6, err := addrs.IsIPv6(bfdSess.DestinationAddress)
	if err != nil {
		return err
	}
	if isLocalIpv6 && isPeerIpv6 {
		req.IsIpv6 = 1
		req.LocalAddr = net.ParseIP(bfdSess.SourceAddress).To16()
		req.PeerAddr = net.ParseIP(bfdSess.DestinationAddress).To16()
	} else if !isLocalIpv6 && !isPeerIpv6 {
		req.IsIpv6 = 0
		req.LocalAddr = net.ParseIP(bfdSess.SourceAddress).To4()
		req.PeerAddr = net.ParseIP(bfdSess.DestinationAddress).To4()
	} else {
		return fmt.Errorf("different IP versions or missing IP address. Local: %v, Peer: %v",
			bfdSess.SourceAddress, bfdSess.DestinationAddress)
	}

	// Authentication
	if bfdSess.Authentication != nil {
		keyID := string(bfdSess.Authentication.KeyId)
		log.Infof("Setting up authentication with index %v", keyID)
		_, _, found := bfdKeyIndexes.LookupIdx(keyID)
		if found {
			req.IsAuthenticated = 1
			req.BfdKeyID = uint8(bfdSess.Authentication.KeyId)
			req.ConfKeyID = bfdSess.Authentication.AdvertisedKeyId
		} else {
			log.Infof("Authentication key %v not found", bfdSess.Authentication.KeyId)
			req.IsAuthenticated = 0
		}
	} else {
		req.IsAuthenticated = 0
	}

	reply := &bfd_api.BfdUDPAddReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// AddBfdUDPSessionFromDetails adds new BFD session with authentication if available.
func AddBfdUDPSessionFromDetails(bfdSess *bfd_api.BfdUDPSessionDetails, bfdKeyIndexes idxvpp.NameToIdx, log logging.Logger,
	vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdUDPAdd{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bfd_api.BfdUDPAdd{
		SwIfIndex:     bfdSess.SwIfIndex,
		DesiredMinTx:  bfdSess.DesiredMinTx,
		RequiredMinRx: bfdSess.RequiredMinRx,
		LocalAddr:     bfdSess.LocalAddr,
		PeerAddr:      bfdSess.PeerAddr,
		DetectMult:    bfdSess.DetectMult,
		IsIpv6:        bfdSess.IsIpv6,
	}

	// Authentication
	if bfdSess.IsAuthenticated != 0 {
		keyID := string(bfdSess.BfdKeyID)
		log.Infof("Setting up authentication with index %v", keyID)
		_, _, found := bfdKeyIndexes.LookupIdx(keyID)
		if found {
			req.IsAuthenticated = 1
			req.BfdKeyID = bfdSess.BfdKeyID
			req.ConfKeyID = bfdSess.ConfKeyID
		} else {
			log.Infof("Authentication key %v not found", bfdSess.BfdKeyID)
			req.IsAuthenticated = 0
		}
	} else {
		req.IsAuthenticated = 0
	}

	reply := &bfd_api.BfdUDPAddReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// ModifyBfdUDPSession modifies existing BFD session excluding authentication which cannot be changed this way.
func ModifyBfdUDPSession(bfdSess *bfd.SingleHopBFD_Session, swIfIndexes ifaceidx.SwIfIndex, vppChan VPPChannel, stopwatch *measure.Stopwatch) (err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdUDPMod{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Find the interface
	ifIdx, _, found := swIfIndexes.LookupIdx(bfdSess.Interface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdSess.Interface)
	}

	req := &bfd_api.BfdUDPMod{
		SwIfIndex:     ifIdx,
		DesiredMinTx:  bfdSess.DesiredMinTxInterval,
		RequiredMinRx: bfdSess.RequiredMinRxInterval,
		DetectMult:    uint8(bfdSess.DetectMultiplier),
	}

	isLocalIpv6, err := addrs.IsIPv6(bfdSess.SourceAddress)
	if err != nil {
		return err
	}
	isPeerIpv6, err := addrs.IsIPv6(bfdSess.DestinationAddress)
	if err != nil {
		return err
	}
	if isLocalIpv6 && isPeerIpv6 {
		req.IsIpv6 = 1
		req.LocalAddr = net.ParseIP(bfdSess.SourceAddress).To16()
		req.PeerAddr = net.ParseIP(bfdSess.DestinationAddress).To16()
	} else if !isLocalIpv6 && !isPeerIpv6 {
		req.IsIpv6 = 0
		req.LocalAddr = net.ParseIP(bfdSess.SourceAddress).To4()
		req.PeerAddr = net.ParseIP(bfdSess.DestinationAddress).To4()
	} else {
		return fmt.Errorf("different IP versions or missing IP address. Local: %v, Peer: %v",
			bfdSess.SourceAddress, bfdSess.DestinationAddress)
	}

	reply := &bfd_api.BfdUDPModReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// DeleteBfdUDPSession removes an existing BFD session.
func DeleteBfdUDPSession(ifIndex uint32, sourceAddress string, destAddress string, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdUDPDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bfd_api.BfdUDPDel{
		SwIfIndex: ifIndex,
		LocalAddr: net.ParseIP(sourceAddress).To4(),
		PeerAddr:  net.ParseIP(destAddress).To4(),
		IsIpv6:    0,
	}

	reply := &bfd_api.BfdUDPDelReply{}
	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// DumpBfdUDPSessions returns a list of BFD session's metadata
func DumpBfdUDPSessions(vppChan VPPChannel, stopwatch *measure.Stopwatch) ([]*bfd_api.BfdUDPSessionDetails, error) {
	return dumpBfdUDPSessionsWithID(false, 0, vppChan, stopwatch)
}

// DumpBfdUDPSessionsWithID returns a list of BFD session's metadata filtered according to provided authentication key
func DumpBfdUDPSessionsWithID(authKeyIndex uint32, vppChan VPPChannel, stopwatch *measure.Stopwatch) ([]*bfd_api.BfdUDPSessionDetails, error) {
	return dumpBfdUDPSessionsWithID(true, authKeyIndex, vppChan, stopwatch)
}

func dumpBfdUDPSessionsWithID(filterID bool, authKeyIndex uint32, vppChan VPPChannel, stopwatch *measure.Stopwatch) (sessions []*bfd_api.BfdUDPSessionDetails, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdUDPSessionDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bfd_api.BfdUDPSessionDump{}
	reqCtx := vppChan.SendMultiRequest(req)
	for {
		msg := &bfd_api.BfdUDPSessionDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return sessions, err
		}

		if filterID {
			// Not interested in sessions without auth key
			if msg.IsAuthenticated == 0 {
				continue
			}
			if msg.BfdKeyID == uint8(authKeyIndex) {
				sessions = append(sessions, msg)
			}
		} else {
			sessions = append(sessions, msg)
		}
	}

	return sessions, nil
}

// SetBfdUDPAuthenticationKey creates new authentication key.
func SetBfdUDPAuthenticationKey(bfdKey *bfd.SingleHopBFD_Key, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) (err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdAuthSetKey{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Convert authentication according to RFC5880.
	var authentication uint8
	if bfdKey.AuthenticationType == 0 {
		authentication = 4 // Keyed SHA1
	} else if bfdKey.AuthenticationType == 1 {
		authentication = 5 // Meticulous keyed SHA1
	} else {
		log.Warnf("Provided authentication type not supported, setting up SHA1")
		authentication = 4
	}

	req := &bfd_api.BfdAuthSetKey{
		ConfKeyID: bfdKey.Id,
		AuthType:  authentication,
		Key:       []byte(bfdKey.Secret),
		KeyLen:    uint8(len(bfdKey.Secret)),
	}

	reply := &bfd_api.BfdAuthSetKeyReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// DeleteBfdUDPAuthenticationKey removes the authentication key.
func DeleteBfdUDPAuthenticationKey(bfdKey *bfd.SingleHopBFD_Key, vppChan VPPChannel, stopwatch *measure.Stopwatch) (err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdAuthDelKey{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bfd_api.BfdAuthDelKey{
		ConfKeyID: bfdKey.Id,
	}

	reply := &bfd_api.BfdAuthDelKeyReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// DumpBfdKeys looks up all BFD auth keys and saves their name-to-index mapping
func DumpBfdKeys(vppChan VPPChannel, stopwatch *measure.Stopwatch) (keys []*bfd_api.BfdAuthKeysDetails, err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdAuthKeysDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bfd_api.BfdAuthKeysDump{}
	reqCtx := vppChan.SendMultiRequest(req)
	for {
		msg := &bfd_api.BfdAuthKeysDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}

		keys = append(keys, msg)
	}

	return keys, nil
}

// AddBfdEchoFunction sets up an echo function for the interface.
func AddBfdEchoFunction(bfdInput *bfd.SingleHopBFD_EchoFunction, swIfIndexes ifaceidx.SwIfIndex, vppChan VPPChannel, stopwatch *measure.Stopwatch) (err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdUDPSetEchoSource{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Verify the interface presence.
	ifIdx, _, found := swIfIndexes.LookupIdx(bfdInput.EchoSourceInterface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdInput.EchoSourceInterface)
	}

	req := &bfd_api.BfdUDPSetEchoSource{
		SwIfIndex: ifIdx,
	}

	reply := &bfd_api.BfdUDPSetEchoSourceReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}

// DeleteBfdEchoFunction removes an echo function.
func DeleteBfdEchoFunction(vppChan VPPChannel, stopwatch *measure.Stopwatch) (err error) {
	defer func(t time.Time) {
		stopwatch.TimeLog(bfd_api.BfdUDPDelEchoSource{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// Prepare the message.
	req := &bfd_api.BfdUDPDelEchoSource{}

	reply := &bfd_api.BfdUDPDelEchoSourceReply{}
	if err = vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return nil
}
