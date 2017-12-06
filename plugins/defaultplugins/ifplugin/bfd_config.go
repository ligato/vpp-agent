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

package ifplugin

//go:generate protoc --proto_path=../common/model/bfd --gogo_out=../common/model/bfd ../common/model/bfd/bfd.proto

import (
	"fmt"
	"net"
	"strings"

	"strconv"
	"time"

	"strconv"
	"time"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	bfd_api "github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// BFDConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/bfd/bfd.proto"
// and stored in ETCD under the key "/vnf-agent/{agent-label}/vpp/config/v1/bfd/".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type BFDConfigurator struct {
	Log logging.Logger

	GoVppmux     govppmux.API
	SwIfIndexes  ifaceidx.SwIfIndex
	ServiceLabel servicelabel.ReaderAPI
	BfdIDSeq     uint32
	Stopwatch    *measure.Stopwatch // timer used to measure and store time
	// Base mappings
	bfdSessionsIndexes   idxvpp.NameToIdxRW
	bfdKeysIndexes       idxvpp.NameToIdxRW
	bfdEchoFunctionIndex idxvpp.NameToIdxRW
	// Auxiliary mappings
	bfdRemovedAuthIndex idxvpp.NameToIdxRW
	vppChannel          *govppapi.Channel
}

// Init members and channels
func (plugin *BFDConfigurator) Init(bfdSessionIndexes idxvpp.NameToIdxRW, bfdKeyIndexes idxvpp.NameToIdxRW, bfdEchoFunctionIndex idxvpp.NameToIdxRW,
	bfdRemovedAuthIndex idxvpp.NameToIdxRW) (err error) {
	plugin.Log.Infof("Initializing BFD configurator")
	plugin.bfdSessionsIndexes = bfdSessionIndexes
	plugin.bfdKeysIndexes = bfdKeyIndexes
	plugin.bfdEchoFunctionIndex = bfdEchoFunctionIndex
	plugin.bfdRemovedAuthIndex = bfdRemovedAuthIndex

	plugin.vppChannel, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	err = vppcalls.CheckMsgCompatibilityForBfd(plugin.vppChannel)
	if err != nil {
		plugin.Log.Error(err)
		return err
	}

	return nil
}

// Close GOVPP channel
func (plugin *BFDConfigurator) Close() error {
	return safeclose.Close(plugin.vppChannel)
}

// ConfigureBfdSession configures bfd session (including authentication if exists). Provided interface has to contain
// ip address defined in BFD as source
func (plugin *BFDConfigurator) ConfigureBfdSession(bfdInput *bfd.SingleHopBFD_Session) error {
	plugin.Log.Infof("Configuring BFD session for interface %v", bfdInput.Interface)

	// Verify interface presence
	ifIdx, ifMeta, found := plugin.SwIfIndexes.LookupIdx(bfdInput.Interface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdInput.Interface)
	}

	// Check whether BFD contains source IP address
	if ifMeta == nil {
		return fmt.Errorf("unable to get IP address data from interface %v", bfdInput.Interface)
	} else {
		var ipFound bool
		for _, ipAddr := range ifMeta.IpAddresses {
			// Remove suffix (BFD is not using it)
			ipWithMask := strings.Split(ipAddr, "/")
			if len(ipWithMask) == 0 {
				return fmt.Errorf("incorrect IP address format %v", ipAddr)
			}
			ipAddrWithoutMask := ipWithMask[0] // the first index is IP address
			if ipAddrWithoutMask == bfdInput.SourceAddress {
				ipFound = true
				break
			}
		}
		if !ipFound {
			return fmt.Errorf("interface %v does not contain address %v required for BFD session",
				bfdInput.Interface, bfdInput.SourceAddress)
		}
	}

	// Call vpp api
	err := vppcalls.AddBfdUDPSession(bfdInput, ifIdx, plugin.bfdKeysIndexes, plugin.Log, plugin.vppChannel,
		measure.GetTimeLog(bfd_api.BfdUDPAdd{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while configuring BFD for interface %v", bfdInput.Interface)
	}

	plugin.bfdSessionsIndexes.RegisterName(bfdInput.Interface, plugin.BfdIDSeq, nil)
	plugin.Log.Debugf("BFD session with interface %v registered. Idx: %v", bfdInput.Interface, plugin.BfdIDSeq)
	plugin.BfdIDSeq++

	plugin.Log.Printf("BFD session for interface %v configured ", bfdInput.Interface)

	return nil
}

// ModifyBfdSession modifies BFD session fields. Source and destination IP address for old and new config has to be the
// same. Authentication is NOT changed here, BFD modify bin api call does not support that
func (plugin *BFDConfigurator) ModifyBfdSession(oldBfdInput *bfd.SingleHopBFD_Session, newBfdInput *bfd.SingleHopBFD_Session) error {
	plugin.Log.Infof("Modifying BFD session for interface %v", newBfdInput.Interface)

	// Verify interface presence
	ifIdx, ifMeta, found := plugin.SwIfIndexes.LookupIdx(newBfdInput.Interface)
	if !found {
		return fmt.Errorf("interface %v does not exist", newBfdInput.Interface)
	}

	// Check whether BFD contains source IP address
	if ifMeta == nil {
		return fmt.Errorf("unable to get IP address data from interface %v", newBfdInput.Interface)
	} else {
		var ipFound bool
		for _, ipAddr := range ifMeta.IpAddresses {
			// Remove suffix
			ipWithMask := strings.Split(ipAddr, "/")
			if len(ipWithMask) == 0 {
				return fmt.Errorf("incorrect IP address format %v", ipAddr)
			}
			ipAddrWithoutMask := ipWithMask[0] // the first index is IP address
			if ipAddrWithoutMask == newBfdInput.SourceAddress {
				ipFound = true
				break
			}
		}
		if !ipFound {
			return fmt.Errorf("interface %v does not contain address %v required for modified BFD session",
				newBfdInput.Interface, newBfdInput.SourceAddress)
		}
	}

	// Find old BFD session
	_, _, found = plugin.bfdSessionsIndexes.LookupIdx(oldBfdInput.Interface)
	if !found {
		plugin.Log.Printf("Previous BFD session does not exist, creating a new one for interface %v", newBfdInput.Interface)
		err := vppcalls.AddBfdUDPSession(newBfdInput, ifIdx, plugin.bfdKeysIndexes, plugin.Log, plugin.vppChannel,
			measure.GetTimeLog(bfd_api.BfdUDPAdd{}, plugin.Stopwatch))
		if err != nil {
			return err
		}
	} else {
		// Compare source and destination addresses which cannot change if BFD session is modified
		// todo new BFD input should be compared to BFD data on the vpp, not the last change (old BFD data)
		if oldBfdInput.SourceAddress != newBfdInput.SourceAddress || oldBfdInput.DestinationAddress != newBfdInput.DestinationAddress {
			return fmt.Errorf("unable to modify BFD session, adresses does not match. Odl session source: %v, dest: %v, new session source: %v, dest: %v",
				oldBfdInput.SourceAddress, oldBfdInput.DestinationAddress, newBfdInput.SourceAddress, newBfdInput.DestinationAddress)
		}
		err := vppcalls.ModifyBfdUDPSession(newBfdInput, plugin.SwIfIndexes, plugin.vppChannel,
			measure.GetTimeLog(bfd_api.BfdUDPMod{}, plugin.Stopwatch))
		if err != nil {
			return err
		}
	}

	plugin.Log.Printf("BFD session for interface %v modified ", newBfdInput.Interface)

	return nil
}

// DeleteBfdSession removes BFD session
func (plugin *BFDConfigurator) DeleteBfdSession(bfdInput *bfd.SingleHopBFD_Session) error {
	plugin.Log.Info("Deleting BFD session")

	ifIndex, _, found := plugin.SwIfIndexes.LookupIdx(bfdInput.Interface)
	if !found {
		return nil
	}

	err := vppcalls.DeleteBfdUDPSession(ifIndex, bfdInput.SourceAddress, bfdInput.DestinationAddress, plugin.vppChannel,
		measure.GetTimeLog(bfd_api.BfdUDPDel{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while deleting BFD for interface %v", bfdInput.Interface)
	}

	plugin.bfdSessionsIndexes.UnregisterName(bfdInput.Interface)
	plugin.Log.Debugf("BFD session with interface %v unregistered", bfdInput.Interface)

	plugin.Log.Printf("BFD session for interface %v removed ", bfdInput.Interface)

	return nil
}

// DumpBfdSessions returns a list of all configured BFD sessions
func (plugin *BFDConfigurator) DumpBfdSessions() ([]*bfd.SingleHopBFD_Session, error) {
	bfdList, err := vppcalls.DumpBfdUDPSessions(plugin.vppChannel,
		measure.GetTimeLog(bfd_api.BfdUDPSessionDump{}, plugin.Stopwatch))

	if err != nil {
		return nil, err
	}
	var wasError error
	var bfdSessionList []*bfd.SingleHopBFD_Session
	for _, bfdItem := range bfdList {
		// find interface
		ifName, _, found := plugin.SwIfIndexes.LookupName(bfdItem.SwIfIndex)
		if !found {
			plugin.Log.Warnf("required interface %v not found for BFD", bfdItem.SwIfIndex)
		}

		// Prepare IPv4 IP addresses
		var dstAddr net.IP = bfdItem.PeerAddr[:4]
		var srcAddr net.IP = bfdItem.LocalAddr[:4]

		bfdSessionList = append(bfdSessionList, &bfd.SingleHopBFD_Session{
			Interface:             ifName,
			DestinationAddress:    dstAddr.To4().String(),
			SourceAddress:         srcAddr.To4().String(),
			Enabled:               true,
			DesiredMinTxInterval:  bfdItem.DesiredMinTx,
			RequiredMinRxInterval: bfdItem.RequiredMinRx,
			DetectMultiplier:      uint32(bfdItem.DetectMult),
			// todo dump authentication
		})
	}

	return bfdSessionList, wasError
}

// ConfigureBfdAuthKey crates new authentication key which can be used for BFD session
func (plugin *BFDConfigurator) ConfigureBfdAuthKey(bfdAuthKey *bfd.SingleHopBFD_Key) error {
	plugin.Log.Print("Setting up BFD authentication key with ID ", bfdAuthKey.Id)

	// Check whether this auth key was not recreated
	authKeyIndex := strconv.FormatUint(uint64(bfdAuthKey.Id), 10)
	_, _, found := plugin.bfdRemovedAuthIndex.LookupIdx(authKeyIndex)
	if found {
		plugin.bfdRemovedAuthIndex.UnregisterName(authKeyIndex)
		plugin.Log.Debugf("Authentication key with ID %v recreated", authKeyIndex)
		plugin.ModifyBfdAuthKey(bfdAuthKey, bfdAuthKey)
	}

	err := vppcalls.SetBfdUDPAuthenticationKey(bfdAuthKey, plugin.Log, plugin.vppChannel,
		measure.GetTimeLog(bfd_api.BfdAuthSetKey{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while setting up BFD auth key with ID %v", bfdAuthKey.Id)
	}

	authKeyIDAsString := strconv.FormatUint(uint64(bfdAuthKey.Id), 10)
	plugin.bfdKeysIndexes.RegisterName(authKeyIDAsString, plugin.BfdIDSeq, nil)
	plugin.Log.Debugf("BFD authentication key with id %v registered. Idx: %v", bfdAuthKey.Id, plugin.BfdIDSeq)
	plugin.BfdIDSeq++

	return nil
}

// ModifyBfdAuthKey modifies auth key fields. Key which is assigned to one or more BFD session cannot be modified
func (plugin *BFDConfigurator) ModifyBfdAuthKey(oldInput *bfd.SingleHopBFD_Key, newInput *bfd.SingleHopBFD_Key) error {
	plugin.Log.Print("Modifying BFD auth key for ID ", oldInput.Id)

	// Check whether this auth key was not recreated
	authKeyIndex := strconv.FormatUint(uint64(oldInput.Id), 10)
	_, _, found := plugin.bfdRemovedAuthIndex.LookupIdx(authKeyIndex)
	if found {
		plugin.bfdRemovedAuthIndex.UnregisterName(authKeyIndex)
		plugin.Log.Debugf("Authentication key with ID %v recreated", oldInput.Id)
	}
	// Check that this auth key is not used in any session
	sessionList, err := vppcalls.DumpBfdUDPSessionsWithID(newInput.Id, plugin.vppChannel,
		measure.GetTimeLog(bfd_api.BfdUDPSessionDump{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while verifying authentication key usage. Id: %v", oldInput.Id)
	}
	if len(sessionList) != 0 {
		// Authentication Key is used and cannot be removed directly
		for _, bfds := range sessionList {
			sourceAddr := net.HardwareAddr(bfds.LocalAddr).String()
			destAddr := net.HardwareAddr(bfds.PeerAddr).String()
			err := vppcalls.DeleteBfdUDPSession(bfds.SwIfIndex, sourceAddr, destAddr, plugin.vppChannel,
				measure.GetTimeLog(bfd_api.BfdUDPDel{}, plugin.Stopwatch))
			if err != nil {
				return err
			}
		}
		plugin.Log.Debugf("%v session(s) temporary removed", len(sessionList))
	}

	err = vppcalls.DeleteBfdUDPAuthenticationKey(oldInput, plugin.vppChannel, measure.GetTimeLog(bfd_api.BfdAuthDelKey{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while removing BFD auth key with ID %v", oldInput.Id)
	}
	err = vppcalls.SetBfdUDPAuthenticationKey(newInput, plugin.Log, plugin.vppChannel, measure.GetTimeLog(bfd_api.BfdAuthSetKey{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while setting up BFD auth key with ID %v", oldInput.Id)
	}

	plugin.LookupBfdKeys()

	// Recreate BFD sessions if necessary
	if len(sessionList) != 0 {
		for _, bfdSession := range sessionList {
			err := vppcalls.AddBfdUDPSessionFromDetails(bfdSession, plugin.bfdKeysIndexes, plugin.Log, plugin.vppChannel,
				measure.GetTimeLog(bfd_api.BfdUDPAdd{}, plugin.Stopwatch))
			if err != nil {
				return err
			}
		}
		plugin.Log.Debugf("%v session(s) recreated", len(sessionList))
	}

	return nil
}

// DeleteBfdAuthKey removes BFD authentication key but only if it is not used in any BFD session
func (plugin *BFDConfigurator) DeleteBfdAuthKey(bfdInput *bfd.SingleHopBFD_Key) error {
	plugin.Log.Info("Deleting BFD auth key")

	// Check that this auth key is not used in any session
	sessionList, err := vppcalls.DumpBfdUDPSessionsWithID(bfdInput.Id, plugin.vppChannel,
		measure.GetTimeLog(bfd_api.BfdUDPSessionDump{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while verifying authentication key usage. Id: %v", bfdInput.Id)
	}

	if len(sessionList) != 0 {
		// Authentication Key is used and cannot be removed directly
		for _, bfds := range sessionList {
			sourceAddr := net.IP(bfds.LocalAddr[0:4]).String()
			destAddr := net.IP(bfds.PeerAddr[0:4]).String()
			err := vppcalls.DeleteBfdUDPSession(bfds.SwIfIndex, sourceAddr, destAddr, plugin.vppChannel, nil)
			if err != nil {
				return err
			}
		}
		plugin.Log.Debugf("%v session(s) temporary removed", len(sessionList))
	}
	err = vppcalls.DeleteBfdUDPAuthenticationKey(bfdInput, plugin.vppChannel, nil)
	if err != nil {
		return fmt.Errorf("error while removing BFD auth key with ID %v", bfdInput.Id)
	}
	authKeyIDAsString := strconv.FormatUint(uint64(bfdInput.Id), 10)
	plugin.bfdKeysIndexes.UnregisterName(authKeyIDAsString)
	plugin.Log.Debugf("BFD authentication key with id %v unregistered", bfdInput.Id)
	// Recreate BFD sessions if necessary
	if len(sessionList) != 0 {
		for _, bfdSession := range sessionList {
			err := vppcalls.AddBfdUDPSessionFromDetails(bfdSession, plugin.bfdKeysIndexes, plugin.Log, plugin.vppChannel, nil)
			if err != nil {
				return err
			}
		}
		plugin.Log.Debugf("%v session(s) recreated", len(sessionList))
	}
	return nil
}

// ConfigureBfdEchoFunction is used to setup BFD Echo function on existing interface
func (plugin *BFDConfigurator) ConfigureBfdEchoFunction(bfdInput *bfd.SingleHopBFD_EchoFunction) error {
	plugin.Log.Print("Configuring BFD echo function for source interface ", bfdInput.EchoSourceInterface)

	// Verify interface presence
	_, _, found := plugin.SwIfIndexes.LookupIdx(bfdInput.EchoSourceInterface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdInput.EchoSourceInterface)
	}

	err := vppcalls.AddBfdEchoFunction(bfdInput, plugin.SwIfIndexes, plugin.vppChannel, measure.GetTimeLog(bfd_api.BfdUDPSetEchoSource{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while setting up BFD echo source with interface %v", bfdInput.EchoSourceInterface)
	}

	plugin.bfdEchoFunctionIndex.RegisterName(bfdInput.EchoSourceInterface, plugin.BfdIDSeq, nil)
	plugin.Log.Debugf("BFD echo function with interface %v registered. Idx: %v", bfdInput.EchoSourceInterface, plugin.BfdIDSeq)
	plugin.BfdIDSeq++

	return nil
}

// ModifyBfdEchoFunction handles echo function changes
func (plugin *BFDConfigurator) ModifyBfdEchoFunction(oldInput *bfd.SingleHopBFD_EchoFunction, newInput *bfd.SingleHopBFD_EchoFunction) error {
	plugin.Log.Debug("There is nothing to modify for BFD echo function")
	// NO-OP
	return nil
}

// DeleteBfdEchoFunction removes BFD echo function
func (plugin *BFDConfigurator) DeleteBfdEchoFunction(bfdInput *bfd.SingleHopBFD_EchoFunction) error {
	plugin.Log.Info("Deleting BFD echo function")

	err := vppcalls.DeleteBfdEchoFunction(plugin.vppChannel, measure.GetTimeLog(bfd_api.BfdUDPDelEchoSource{}, plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("error while removing BFD echo source with interface %v", bfdInput.EchoSourceInterface)
	}

	plugin.bfdEchoFunctionIndex.UnregisterName(bfdInput.EchoSourceInterface)
	plugin.Log.Debugf("BFD echo function with interface %v unregistered", bfdInput.EchoSourceInterface)

	return nil
}

// LookupBfdSessions looks up all BFD sessions and saves their name-to-index mapping
func (plugin *BFDConfigurator) LookupBfdSessions() error {
	start := time.Now()
	req := &bfd_api.BfdUDPSessionDump{}
	reqCtx := plugin.vppChannel.SendMultiRequest(req)

	for {
		msg := &bfd_api.BfdUDPSessionDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return err
		}

		// Store the name-to-index mapping if it does not exist yet
		name, _, found := plugin.SwIfIndexes.LookupName(msg.SwIfIndex)
		if !found {
			continue
		}
		_, _, found = plugin.bfdSessionsIndexes.LookupIdx(name)
		if !found {
			plugin.bfdEchoFunctionIndex.RegisterName(name, plugin.BfdIDSeq, nil)
			plugin.Log.Debugf("BFD session with interface registered. Idx: %v", plugin.BfdIDSeq)
			plugin.BfdIDSeq++
		}
	}

	// BfdUDPSessionDump time
	if plugin.Stopwatch != nil {
		timeLog := measure.GetTimeLog(bfd_api.BfdUDPSessionDump{}, plugin.Stopwatch)
		timeLog.LogTimeEntry(time.Since(start))
	}

	return nil
}

// LookupBfdKeys looks up all BFD auth keys and saves their name-to-index mapping
func (plugin *BFDConfigurator) LookupBfdKeys() error {
	start := time.Now()
	req := &bfd_api.BfdAuthKeysDump{}
	reqCtx := plugin.vppChannel.SendMultiRequest(req)

	for {
		msg := &bfd_api.BfdAuthKeysDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return err
		}

		// Store the name-to-index mapping if it does not exist yet
		keyID := strconv.FormatUint(uint64(msg.ConfKeyID), 10)
		_, _, found := plugin.bfdKeysIndexes.LookupIdx(keyID)
		if !found {
			plugin.bfdEchoFunctionIndex.RegisterName(keyID, plugin.BfdIDSeq, nil)
			plugin.Log.Debugf("BFD authentication key registered. Idx: %v", plugin.BfdIDSeq)
			plugin.BfdIDSeq++
		}
	}

	// BfdAuthKeysDump time
	if plugin.Stopwatch != nil {
		timeLog := measure.GetTimeLog(bfd_api.BfdAuthKeysDump{}, plugin.Stopwatch)
		timeLog.LogTimeEntry(time.Since(start))
	}

	return nil
}
