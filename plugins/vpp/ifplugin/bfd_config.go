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

//go:generate protoc --proto_path=../model/bfd --gogo_out=../model/bfd ../model/bfd/bfd.proto

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/model/bfd"
)

// BFDConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of BFDs as modelled by the proto file "../model/bfd/bfd.proto"
// and stored in ETCD under the key "/vnf-agent/{agent-label}/vpp/config/v1/bfd/".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type BFDConfigurator struct {
	log logging.Logger

	ifIndexes ifaceidx.SwIfIndex
	bfdIDSeq  uint32
	stopwatch *measure.Stopwatch // timer used to measure and store time
	// Base mappings
	sessionsIndexes   idxvpp.NameToIdxRW
	keysIndexes       idxvpp.NameToIdxRW
	echoFunctionIndex idxvpp.NameToIdxRW

	vppChan govppapi.Channel

	// VPP API handler
	bfdHandler vppcalls.BfdVppAPI
}

// Init members and channels
func (plugin *BFDConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API, swIfIndexes ifaceidx.SwIfIndex,
	enableStopwatch bool) (err error) {
	// Logger
	plugin.log = logger.NewLogger("-bfd-conf")
	plugin.log.Infof("Initializing BFD configurator")

	// Configurator-wide stopwatch instance
	if enableStopwatch {
		plugin.stopwatch = measure.NewStopwatch("BFD-configurator", plugin.log)
	}

	// Mappings
	plugin.ifIndexes = swIfIndexes
	plugin.sessionsIndexes = nametoidx.NewNameToIdx(plugin.log, "bfd_session_indexes", nil)
	plugin.keysIndexes = nametoidx.NewNameToIdx(plugin.log, "bfd_auth_keys_indexes", nil)
	plugin.echoFunctionIndex = nametoidx.NewNameToIdx(plugin.log, "bfd_echo_function_index", nil)

	// VPP channel
	plugin.vppChan, err = goVppMux.NewAPIChannel()
	if err != nil {
		return err
	}

	// VPP API handler
	plugin.bfdHandler = vppcalls.NewBfdVppHandler(plugin.vppChan, plugin.ifIndexes, plugin.log, plugin.stopwatch)

	return nil
}

// Close GOVPP channel
func (plugin *BFDConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (plugin *BFDConfigurator) clearMapping() {
	plugin.sessionsIndexes.Clear()
	plugin.keysIndexes.Clear()
	plugin.echoFunctionIndex.Clear()
}

// GetBfdSessionIndexes gives access to BFD session indexes
func (plugin *BFDConfigurator) GetBfdSessionIndexes() idxvpp.NameToIdxRW {
	return plugin.sessionsIndexes
}

// GetBfdKeyIndexes gives access to BFD key indexes
func (plugin *BFDConfigurator) GetBfdKeyIndexes() idxvpp.NameToIdxRW {
	return plugin.keysIndexes
}

// GetBfdEchoFunctionIndexes gives access to BFD echo function indexes
func (plugin *BFDConfigurator) GetBfdEchoFunctionIndexes() idxvpp.NameToIdxRW {
	return plugin.echoFunctionIndex
}

// ConfigureBfdSession configures bfd session (including authentication if exists). Provided interface has to contain
// ip address defined in BFD as source
func (plugin *BFDConfigurator) ConfigureBfdSession(bfdInput *bfd.SingleHopBFD_Session) error {
	plugin.log.Infof("Configuring BFD session for interface %v", bfdInput.Interface)

	// Verify interface presence
	ifIdx, ifMeta, found := plugin.ifIndexes.LookupIdx(bfdInput.Interface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdInput.Interface)
	}

	// Check whether BFD contains source IP address
	if ifMeta == nil {
		return fmt.Errorf("unable to get IP address data from interface %v", bfdInput.Interface)
	}
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

	// Call vpp api
	err := plugin.bfdHandler.AddBfdUDPSession(bfdInput, ifIdx, plugin.keysIndexes)
	if err != nil {
		return fmt.Errorf("error while configuring BFD for interface %v", bfdInput.Interface)
	}

	plugin.sessionsIndexes.RegisterName(bfdInput.Interface, plugin.bfdIDSeq, nil)
	plugin.log.Debugf("BFD session with interface %v registered. Idx: %v", bfdInput.Interface, plugin.bfdIDSeq)
	plugin.bfdIDSeq++

	plugin.log.Infof("BFD session for interface %v configured ", bfdInput.Interface)

	return nil
}

// ModifyBfdSession modifies BFD session fields. Source and destination IP address for old and new config has to be the
// same. Authentication is NOT changed here, BFD modify bin api call does not support that
func (plugin *BFDConfigurator) ModifyBfdSession(oldBfdInput *bfd.SingleHopBFD_Session, newBfdInput *bfd.SingleHopBFD_Session) error {
	plugin.log.Infof("Modifying BFD session for interface %v", newBfdInput.Interface)

	// Verify interface presence
	ifIdx, ifMeta, found := plugin.ifIndexes.LookupIdx(newBfdInput.Interface)
	if !found {
		return fmt.Errorf("interface %v does not exist", newBfdInput.Interface)
	}

	// Check whether BFD contains source IP address
	if ifMeta == nil {
		return fmt.Errorf("unable to get IP address data from interface %v", newBfdInput.Interface)
	}
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

	// Find old BFD session
	_, _, found = plugin.sessionsIndexes.LookupIdx(oldBfdInput.Interface)
	if !found {
		plugin.log.Printf("Previous BFD session does not exist, creating a new one for interface %v", newBfdInput.Interface)
		err := plugin.bfdHandler.AddBfdUDPSession(newBfdInput, ifIdx, plugin.keysIndexes)
		if err != nil {
			return err
		}
		plugin.sessionsIndexes.RegisterName(newBfdInput.Interface, plugin.bfdIDSeq, nil)
		plugin.bfdIDSeq++
	} else {
		// Compare source and destination addresses which cannot change if BFD session is modified
		// todo new BFD input should be compared to BFD data on the vpp, not the last change (old BFD data)
		if oldBfdInput.SourceAddress != newBfdInput.SourceAddress || oldBfdInput.DestinationAddress != newBfdInput.DestinationAddress {
			return fmt.Errorf("unable to modify BFD session, adresses does not match. Odl session source: %v, dest: %v, new session source: %v, dest: %v",
				oldBfdInput.SourceAddress, oldBfdInput.DestinationAddress, newBfdInput.SourceAddress, newBfdInput.DestinationAddress)
		}
		err := plugin.bfdHandler.ModifyBfdUDPSession(newBfdInput, plugin.ifIndexes)
		if err != nil {
			return err
		}
	}

	plugin.log.Infof("BFD session for interface %v modified ", newBfdInput.Interface)

	return nil
}

// DeleteBfdSession removes BFD session
func (plugin *BFDConfigurator) DeleteBfdSession(bfdInput *bfd.SingleHopBFD_Session) error {
	plugin.log.Info("Deleting BFD session")

	ifIndex, _, found := plugin.ifIndexes.LookupIdx(bfdInput.Interface)
	if !found {
		return fmt.Errorf("cannot remove BFD session, interface %s not found", bfdInput.Interface)
	}

	err := plugin.bfdHandler.DeleteBfdUDPSession(ifIndex, bfdInput.SourceAddress, bfdInput.DestinationAddress)
	if err != nil {
		return fmt.Errorf("error while deleting BFD for interface %v", bfdInput.Interface)
	}

	plugin.sessionsIndexes.UnregisterName(bfdInput.Interface)
	plugin.log.Debugf("BFD session with interface %v unregistered", bfdInput.Interface)

	return nil
}

// ConfigureBfdAuthKey crates new authentication key which can be used for BFD session
func (plugin *BFDConfigurator) ConfigureBfdAuthKey(bfdAuthKey *bfd.SingleHopBFD_Key) error {
	plugin.log.Infof("Configuring BFD authentication key with ID %v", bfdAuthKey.Id)

	err := plugin.bfdHandler.SetBfdUDPAuthenticationKey(bfdAuthKey)
	if err != nil {
		return fmt.Errorf("error while setting up BFD auth key with ID %v", bfdAuthKey.Id)
	}

	authKeyIDAsString := AuthKeyIdentifier(bfdAuthKey.Id)
	plugin.keysIndexes.RegisterName(authKeyIDAsString, plugin.bfdIDSeq, nil)
	plugin.log.Debugf("BFD authentication key with id %v registered. Idx: %v", bfdAuthKey.Id, plugin.bfdIDSeq)
	plugin.bfdIDSeq++

	plugin.log.Infof("BFD authentication key with ID %v configured", bfdAuthKey.Id)

	return nil
}

// ModifyBfdAuthKey modifies auth key fields. Key which is assigned to one or more BFD session cannot be modified
func (plugin *BFDConfigurator) ModifyBfdAuthKey(oldInput *bfd.SingleHopBFD_Key, newInput *bfd.SingleHopBFD_Key) error {
	plugin.log.Infof("Modifying BFD auth key for ID %d", oldInput.Id)

	// Check that this auth key is not used in any session
	sessionList, err := plugin.bfdHandler.DumpBfdUDPSessionsWithID(newInput.Id)
	if err != nil {
		return fmt.Errorf("error while verifying authentication key usage. Id: %d: %v", oldInput.Id, err)
	}
	if sessionList != nil && len(sessionList.Session) != 0 {
		// Authentication Key is used and cannot be removed directly
		for _, bfds := range sessionList.Session {
			sourceAddr := net.HardwareAddr(bfds.SourceAddress).String()
			destAddr := net.HardwareAddr(bfds.DestinationAddress).String()
			ifIdx, _, found := plugin.ifIndexes.LookupIdx(bfds.Interface)
			if !found {
				plugin.log.Warnf("Modify BFD auth key: interface index for %s not found", bfds.Interface)
			}
			err := plugin.bfdHandler.DeleteBfdUDPSession(ifIdx, sourceAddr, destAddr)
			if err != nil {
				return err
			}
		}
		plugin.log.Debugf("%v session(s) temporary removed", len(sessionList.Session))
	}

	err = plugin.bfdHandler.DeleteBfdUDPAuthenticationKey(oldInput)
	if err != nil {
		return fmt.Errorf("error while removing BFD auth key with ID %d: %v", oldInput.Id, err)
	}
	err = plugin.bfdHandler.SetBfdUDPAuthenticationKey(newInput)
	if err != nil {
		return fmt.Errorf("error while setting up BFD auth key with ID %d: %v", oldInput.Id, err)
	}

	// Recreate BFD sessions if necessary
	if sessionList != nil && len(sessionList.Session) != 0 {
		for _, bfdSession := range sessionList.Session {
			ifIdx, _, found := plugin.ifIndexes.LookupIdx(bfdSession.Interface)
			if !found {
				plugin.log.Warnf("Modify BFD auth key: interface index for %s not found", bfdSession.Interface)
			}
			err := plugin.bfdHandler.AddBfdUDPSession(bfdSession, ifIdx, plugin.keysIndexes)
			if err != nil {
				return err
			}
		}
		plugin.log.Debugf("%v session(s) recreated", len(sessionList.Session))
	}

	return nil
}

// DeleteBfdAuthKey removes BFD authentication key but only if it is not used in any BFD session
func (plugin *BFDConfigurator) DeleteBfdAuthKey(bfdInput *bfd.SingleHopBFD_Key) error {
	plugin.log.Info("Deleting BFD auth key")

	// Check that this auth key is not used in any session
	sessionList, err := plugin.bfdHandler.DumpBfdUDPSessionsWithID(bfdInput.Id)
	if err != nil {
		return fmt.Errorf("error while verifying authentication key usage. Id: %v", bfdInput.Id)
	}

	if sessionList != nil && len(sessionList.Session) != 0 {
		// Authentication Key is used and cannot be removed directly
		for _, bfds := range sessionList.Session {
			ifIdx, _, found := plugin.ifIndexes.LookupIdx(bfds.Interface)
			if !found {
				plugin.log.Warnf("Delete BFD auth key: interface index for %s not found", bfds.Interface)
			}
			err := plugin.bfdHandler.DeleteBfdUDPSession(ifIdx, bfds.SourceAddress, bfds.DestinationAddress)
			if err != nil {
				return err
			}
		}
		plugin.log.Debugf("%v session(s) temporary removed", len(sessionList.Session))
	}
	err = plugin.bfdHandler.DeleteBfdUDPAuthenticationKey(bfdInput)
	if err != nil {
		return fmt.Errorf("error while removing BFD auth key with ID %v", bfdInput.Id)
	}
	authKeyIDAsString := AuthKeyIdentifier(bfdInput.Id)
	plugin.keysIndexes.UnregisterName(authKeyIDAsString)
	plugin.log.Debugf("BFD authentication key with id %v unregistered", bfdInput.Id)
	// Recreate BFD sessions if necessary
	if sessionList != nil && len(sessionList.Session) != 0 {
		for _, bfdSession := range sessionList.Session {
			ifIdx, _, found := plugin.ifIndexes.LookupIdx(bfdSession.Interface)
			if !found {
				plugin.log.Warnf("Delete BFD auth key: interface index for %s not found", bfdSession.Interface)
			}
			err := plugin.bfdHandler.AddBfdUDPSession(bfdSession, ifIdx, plugin.keysIndexes)
			if err != nil {
				return err
			}
		}
		plugin.log.Debugf("%v session(s) recreated", len(sessionList.Session))
	}
	return nil
}

// ConfigureBfdEchoFunction is used to setup BFD Echo function on existing interface
func (plugin *BFDConfigurator) ConfigureBfdEchoFunction(bfdInput *bfd.SingleHopBFD_EchoFunction) error {
	plugin.log.Infof("Configuring BFD echo function for source interface %v", bfdInput.EchoSourceInterface)

	// Verify interface presence
	_, _, found := plugin.ifIndexes.LookupIdx(bfdInput.EchoSourceInterface)
	if !found {
		return fmt.Errorf("interface %v does not exist", bfdInput.EchoSourceInterface)
	}

	err := plugin.bfdHandler.AddBfdEchoFunction(bfdInput, plugin.ifIndexes)
	if err != nil {
		return fmt.Errorf("error while setting up BFD echo source with interface %v", bfdInput.EchoSourceInterface)
	}

	plugin.echoFunctionIndex.RegisterName(bfdInput.EchoSourceInterface, plugin.bfdIDSeq, nil)
	plugin.log.Debugf("BFD echo function with interface %v registered. Idx: %v", bfdInput.EchoSourceInterface, plugin.bfdIDSeq)
	plugin.bfdIDSeq++

	plugin.log.Infof("Echo source set to interface %v ", bfdInput.EchoSourceInterface)

	return nil
}

// ModifyBfdEchoFunction handles echo function changes
func (plugin *BFDConfigurator) ModifyBfdEchoFunction(oldInput *bfd.SingleHopBFD_EchoFunction, newInput *bfd.SingleHopBFD_EchoFunction) error {
	plugin.log.Debug("There is nothing to modify for BFD echo function")
	// NO-OP
	return nil
}

// DeleteBfdEchoFunction removes BFD echo function
func (plugin *BFDConfigurator) DeleteBfdEchoFunction(bfdInput *bfd.SingleHopBFD_EchoFunction) error {
	plugin.log.Info("Deleting BFD echo function")

	err := plugin.bfdHandler.DeleteBfdEchoFunction()
	if err != nil {
		return fmt.Errorf("error while removing BFD echo source with interface %v", bfdInput.EchoSourceInterface)
	}

	plugin.echoFunctionIndex.UnregisterName(bfdInput.EchoSourceInterface)
	plugin.log.Debugf("BFD echo function with interface %v unregistered", bfdInput.EchoSourceInterface)

	plugin.log.Info("Echo source unset")

	return nil
}

// Generates common identifier for authentication key
func AuthKeyIdentifier(id uint32) string {
	return strconv.Itoa(int(id))
}
