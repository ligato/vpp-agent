// Copyright (c) 2018 Bell Canada, Pantheon Technologies and/or its affiliates.
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
	"strings"
	"time"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/sr"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/srv6"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
)

// Constants for behavior function hardcoded into VPP (there can be also custom behavior functions implemented as VPP plugins)
// Constants are taken from VPP's vnet/srv6/sr.h (names are modified to Golang from original C form in VPP code)
const (
	BehaviorEnd    uint8 = iota + 1 //Behavior of simple endpoint
	BehaviorX                       //Behavior of endpoint with Layer-3 cross-connect
	BehaviorT                       //Behavior of endpoint with specific IPv6 table lookup
	BehaviorDfirst                  //Unused. Separator in between regular and D
	BehaviorDX2                     //Behavior of endpoint with decapulation and Layer-2 cross-connect (or DX2 with egress VLAN rewrite when VLAN notzero - not supported this variant yet)
	BehaviorDX6                     //Behavior of endpoint with decapsulation and IPv6 cross-connect
	BehaviorDX4                     //Behavior of endpoint with decapsulation and IPv4 cross-connect
	BehaviorDT6                     //Behavior of endpoint with decapsulation and specific IPv6 table lookup
	BehaviorDT4                     //Behavior of endpoint with decapsulation and specific IPv4 table lookup
	BehaviorLast                    //seems unused, note in VPP: "Must always be the last one"
)

// Constants for steering type
// Constants are taken from VPP's vnet/srv6/sr.h (names are modified to Golang from original C form in VPP code)
const (
	SteerTypeL2 uint8 = iota*2 + 2
	SteerTypeIPv4
	SteerTypeIPv6
)

// Constants for operation of SR policy modify binary API method
const (
	AddSRList            uint8 = iota + 1 //Add SR List to an existing SR policy
	DeleteSRList                          //Delete SR List from an existing SR policy
	ModifyWeightOfSRList                  //Modify the weight of an existing SR List
)

// SRv6Calls is API boundary for vppcall package access, introduced to properly test code dependent on vppcalls package
type SRv6Calls interface {
	// AddLocalSid adds local sid given by <sidAddr> and <localSID> into VPP
	AddLocalSid(sidAddr net.IP, localSID *srv6.LocalSID, swIfIndex ifaceidx.SwIfIndex, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
	// DeleteLocalSid delets local sid given by <sidAddr> in VPP
	DeleteLocalSid(sidAddr net.IP, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
	//SetEncapsSourceAddress sets for SRv6 in VPP the source address used for encapsulated packet
	SetEncapsSourceAddress(address string, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
	//AddPolicy adds SRv6 policy given by identified <bindingSid>,initial segment for policy <policySegment> and other policy settings in <policy>
	AddPolicy(bindingSid net.IP, policy *srv6.Policy, policySegment *srv6.PolicySegment, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
	//DeletePolicy deletes SRv6 policy given by binding SID <bindingSid>
	DeletePolicy(bindingSid net.IP, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
	//AddPolicySegment adds segment <policySegment> to SRv6 policy <policy> that has policy BSID <bindingSid>
	AddPolicySegment(bindingSid net.IP, policy *srv6.Policy, policySegment *srv6.PolicySegment, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
	//DeletePolicySegment removes segment <policySegment> (with segment index <segmentIndex>) from SRv6 policy <policy> that has policy BSID <bindingSid>
	DeletePolicySegment(bindingSid net.IP, policy *srv6.Policy, policySegment *srv6.PolicySegment, segmentIndex uint32, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
	//AddSteering sets in VPP steering into SRv6 policy.
	AddSteering(steering *srv6.Steering, swIfIndex ifaceidx.SwIfIndex, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
	//RemoveSteering removes in VPP steering into SRv6 policy.
	RemoveSteering(steering *srv6.Steering, swIfIndex ifaceidx.SwIfIndex, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error
}

type srv6Calls struct{}

// NewSRv6Calls creates implementation of SRv6Calls interface
func NewSRv6Calls() SRv6Calls {
	return &srv6Calls{}
}

// AddLocalSid adds local sid given by <sidAddr> and <localSID> into VPP
func (calls *srv6Calls) AddLocalSid(sidAddr net.IP, localSID *srv6.LocalSID, swIfIndex ifaceidx.SwIfIndex,
	log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	return calls.addDelLocalSid(false, sidAddr, localSID, swIfIndex, log, vppChan, stopwatch)
}

// DeleteLocalSid delets local sid given by <sidAddr> in VPP
func (calls *srv6Calls) DeleteLocalSid(sidAddr net.IP, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	return calls.addDelLocalSid(true, sidAddr, nil, nil, log, vppChan, stopwatch)
}

func (calls *srv6Calls) addDelLocalSid(deletion bool, sidAddr net.IP, localSID *srv6.LocalSID, swIfIndex ifaceidx.SwIfIndex,
	log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.WithFields(logging.Fields{"localSID": sidAddr, "delete": deletion, "FIB table ID": calls.fibTableID(localSID), "end function": calls.endFunction(localSID)}).
		Debug("Adding/deleting Local SID", sidAddr)
	defer func(t time.Time) {
		stopwatch.TimeLog(sr.SrLocalsidAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &sr.SrLocalsidAddDel{
		IsDel:        boolToInt(deletion),
		LocalsidAddr: []byte(sidAddr),
	}
	if !deletion {
		req.FibTable = localSID.FibTableID //where to install localsid entry
		if err := calls.writeEndFunction(req, sidAddr, localSID, swIfIndex); err != nil {
			return err
		}
	}
	reply := &sr.SrLocalsidAddDelReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"localSID": sidAddr, "delete": deletion, "FIB table ID": calls.fibTableID(localSID), "end function": calls.endFunction(localSID)}).
		Debug("Added/deleted Local SID ", sidAddr)

	return nil
}

func (calls *srv6Calls) fibTableID(localSID *srv6.LocalSID) string {
	if localSID != nil {
		return string(localSID.FibTableID)
	}
	return "<nil>"
}

func (calls *srv6Calls) endFunction(localSID *srv6.LocalSID) string {
	if localSID == nil {
		return "<nil>"
	} else if localSID.BaseEndFunction != nil {
		return fmt.Sprintf("End{psp: %v}", localSID.BaseEndFunction.Psp)
	} else if localSID.EndFunctionX != nil {
		return fmt.Sprintf("X{psp: %v, OutgoingInterface: %v, NextHop: %v}", localSID.EndFunctionX.Psp, localSID.EndFunctionX.OutgoingInterface, localSID.EndFunctionX.NextHop)
	} else if localSID.EndFunctionT != nil {
		return fmt.Sprintf("T{psp: %v}", localSID.EndFunctionT.Psp)
	} else if localSID.EndFunctionDX2 != nil {
		return fmt.Sprintf("DX2{VlanTag: %v, OutgoingInterface: %v, NextHop: %v}", localSID.EndFunctionDX2.VlanTag, localSID.EndFunctionDX2.OutgoingInterface, localSID.EndFunctionDX2.NextHop)
	} else if localSID.EndFunctionDX4 != nil {
		return fmt.Sprintf("DX4{OutgoingInterface: %v, NextHop: %v}", localSID.EndFunctionDX4.OutgoingInterface, localSID.EndFunctionDX4.NextHop)
	} else if localSID.EndFunctionDX6 != nil {
		return fmt.Sprintf("DX6{OutgoingInterface: %v, NextHop: %v}", localSID.EndFunctionDX6.OutgoingInterface, localSID.EndFunctionDX6.NextHop)
	} else if localSID.EndFunctionDT4 != nil {
		return fmt.Sprint("DT4")
	} else if localSID.EndFunctionDT6 != nil {
		return fmt.Sprint("DT6")
	}
	return "unknown end function"
}

func (calls *srv6Calls) writeEndFunction(req *sr.SrLocalsidAddDel, sidAddr net.IP, localSID *srv6.LocalSID, swIfIndex ifaceidx.SwIfIndex) error {
	if localSID.BaseEndFunction != nil {
		req.Behavior = BehaviorEnd
		req.EndPsp = boolToInt(localSID.BaseEndFunction.Psp)
	} else if localSID.EndFunctionX != nil {
		req.Behavior = BehaviorX
		req.EndPsp = boolToInt(localSID.EndFunctionX.Psp)
		interfaceSwIndex, _, exists := swIfIndex.LookupIdx(localSID.EndFunctionX.OutgoingInterface)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", localSID.EndFunctionX.OutgoingInterface)
		}
		req.SwIfIndex = interfaceSwIndex
		nhAddr, err := parseIPv6(localSID.EndFunctionX.NextHop)
		if err != nil {
			return err
		}
		req.NhAddr = []byte(nhAddr)
	} else if localSID.EndFunctionT != nil {
		req.Behavior = BehaviorT
		req.EndPsp = boolToInt(localSID.EndFunctionT.Psp)
	} else if localSID.EndFunctionDX2 != nil {
		req.Behavior = BehaviorDX2
		req.VlanIndex = localSID.EndFunctionDX2.VlanTag
		interfaceSwIndex, _, exists := swIfIndex.LookupIdx(localSID.EndFunctionDX2.OutgoingInterface)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", localSID.EndFunctionDX2.OutgoingInterface)
		}
		req.SwIfIndex = interfaceSwIndex
		nhAddr, err := parseIPv6(localSID.EndFunctionDX2.NextHop)
		if err != nil {
			return err
		}
		req.NhAddr = []byte(nhAddr)
	} else if localSID.EndFunctionDX4 != nil {
		req.Behavior = BehaviorDX4
		interfaceSwIndex, _, exists := swIfIndex.LookupIdx(localSID.EndFunctionDX4.OutgoingInterface)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", localSID.EndFunctionDX4.OutgoingInterface)
		}
		req.SwIfIndex = interfaceSwIndex
		nhAddr, err := parseIPv6(localSID.EndFunctionDX4.NextHop) //parses also IPv4
		if err != nil {
			return err
		}
		if nhAddr.To4() == nil {
			return fmt.Errorf("next hop of DX4 end function (%v) is not valid IPv4 address", localSID.EndFunctionDX4.NextHop)
		}
		req.NhAddr = []byte(nhAddr)
	} else if localSID.EndFunctionDX6 != nil {
		req.Behavior = BehaviorDX6
		interfaceSwIndex, _, exists := swIfIndex.LookupIdx(localSID.EndFunctionDX6.OutgoingInterface)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", localSID.EndFunctionDX6.OutgoingInterface)
		}
		req.SwIfIndex = interfaceSwIndex
		nhAddr, err := parseIPv6(localSID.EndFunctionDX6.NextHop)
		if err != nil {
			return err
		}
		req.NhAddr = []byte(nhAddr)
	} else if localSID.EndFunctionDT4 != nil {
		req.Behavior = BehaviorDT4
	} else if localSID.EndFunctionDT6 != nil {
		req.Behavior = BehaviorDT6
	} else {
		return fmt.Errorf("End function not set. Please configure end function for local SID %v ", sidAddr)
	}
	return nil
}

//SetEncapsSourceAddress sets for SRv6 in VPP the source address used for encapsulated packet
func (calls *srv6Calls) SetEncapsSourceAddress(address string, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.Debugf("Configuring encapsulation source address to address %v", address)
	defer func(t time.Time) {
		stopwatch.TimeLog(sr.SrSetEncapSource{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	ipAddress, err := parseIPv6(address)
	if err != nil {
		return err
	}
	req := &sr.SrSetEncapSource{
		EncapsSource: []byte(ipAddress),
	}
	reply := &sr.SrSetEncapSourceReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"Encapsulation source address": address}).
		Debug("Encapsulation source address configured.")

	return nil
}

//AddPolicy adds SRv6 policy given by identified <bindingSid>,initial segment for policy <policySegment> and other policy settings in <policy>
func (calls *srv6Calls) AddPolicy(bindingSid net.IP, policy *srv6.Policy, policySegment *srv6.PolicySegment,
	log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.Debugf("Adding SR policy with binding SID %v and list of next SIDs %v", bindingSid, policySegment.Segments)
	defer func(t time.Time) {
		stopwatch.TimeLog(sr.SrPolicyAdd{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	segmentsCount, segments, err := calls.convertNextSidList(policySegment.Segments)
	if err != nil {
		return err
	}
	req := &sr.SrPolicyAdd{
		BsidAddr:  []byte(bindingSid),
		Weight:    policySegment.Weight,
		NSegments: segmentsCount,
		Segments:  segments,
		IsEncap:   boolToInt(policy.SrhEncapsulation),
		Type:      boolToInt(policy.SprayBehaviour),
		FibTable:  policy.FibTableID,
	}
	reply := &sr.SrPolicyAddReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"binding SID": bindingSid, "list of next SIDs": policySegment.Segments}).
		Debug("SR policy added")

	return nil
}

//DeletePolicy deletes SRv6 policy given by binding SID <bindingSid>
func (calls *srv6Calls) DeletePolicy(bindingSid net.IP, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.Debugf("Deleting SR policy with binding SID %v ", bindingSid)
	defer func(t time.Time) {
		stopwatch.TimeLog(sr.SrPolicyDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &sr.SrPolicyDel{
		BsidAddr: []byte(bindingSid), //TODO add ability to define policy also by index (SrPolicyIndex)
	}
	reply := &sr.SrPolicyDelReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"binding SID": bindingSid}).
		Debug("SR policy deleted")

	return nil
}

//AddPolicySegment adds segment <policySegment> to SRv6 policy <policy> that has policy BSID <bindingSid>
func (calls *srv6Calls) AddPolicySegment(bindingSid net.IP, policy *srv6.Policy, policySegment *srv6.PolicySegment,
	log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.Debugf("Adding segment %v to SR policy with binding SID %v", policySegment.Segments, bindingSid)
	err := calls.modPolicy(AddSRList, bindingSid, policy, policySegment, 0, log, vppChan, stopwatch)
	if err == nil {
		log.WithFields(logging.Fields{"binding SID": bindingSid, "list of next SIDs": policySegment.Segments}).
			Debug("SR policy modified(added another segment list)")
	}
	return err
}

//DeletePolicySegment removes segment <policySegment> (with segment index <segmentIndex>) from SRv6 policy <policy> that has policy BSID <bindingSid>
func (calls *srv6Calls) DeletePolicySegment(bindingSid net.IP, policy *srv6.Policy, policySegment *srv6.PolicySegment,
	segmentIndex uint32, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	log.Debugf("Removing segment %v (index %v) from SR policy with binding SID %v", policySegment.Segments, segmentIndex, bindingSid)
	err := calls.modPolicy(DeleteSRList, bindingSid, policy, policySegment, segmentIndex, log, vppChan, stopwatch)
	if err == nil {
		log.WithFields(logging.Fields{"binding SID": bindingSid, "list of next SIDs": policySegment.Segments, "segmentIndex": segmentIndex}).
			Debug("SR policy modified(removed segment list)")
	}
	return err
}

func (calls *srv6Calls) modPolicy(operation uint8, bindingSid net.IP, policy *srv6.Policy, policySegment *srv6.PolicySegment,
	segmentIndex uint32, log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(sr.SrPolicyMod{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	segmentsCount, segments, err := calls.convertNextSidList(policySegment.Segments)
	if err != nil {
		return err
	}
	req := &sr.SrPolicyMod{
		BsidAddr:  []byte(bindingSid), //TODO add ability to define policy also by index (SrPolicyIndex)
		Operation: operation,
		Weight:    policySegment.Weight,
		NSegments: segmentsCount,
		Segments:  segments,
		FibTable:  policy.FibTableID,
	}
	if operation == DeleteSRList || operation == ModifyWeightOfSRList {
		req.SlIndex = segmentIndex
	}

	reply := &sr.SrPolicyModReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}
	return nil
}

func (calls *srv6Calls) convertNextSidList(nextSidList []string) (uint8, []sr.IPv6type, error) {
	segments := make([]sr.IPv6type, 0)
	for _, sid := range nextSidList {
		// parse to IPv6 address
		parserSid, err := parseIPv6(sid)
		if err != nil {
			return 0, []sr.IPv6type{}, err
		}

		// add sid to segment list
		ipv6Segment := sr.IPv6type{}
		copy(ipv6Segment.Value[:], parserSid) //sr.IPv6type.Value = [16]byte
		segments = append(segments, ipv6Segment)
	}
	return uint8(len(nextSidList)), segments, nil
}

//AddSteering sets in VPP steering into SRv6 policy.
func (calls *srv6Calls) AddSteering(steering *srv6.Steering, swIfIndex ifaceidx.SwIfIndex, log logging.Logger,
	vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	return calls.addDelSteering(false, steering, swIfIndex, log, vppChan, stopwatch)
}

//RemoveSteering removes in VPP steering into SRv6 policy.
func (calls *srv6Calls) RemoveSteering(steering *srv6.Steering, swIfIndex ifaceidx.SwIfIndex, log logging.Logger,
	vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	return calls.addDelSteering(true, steering, swIfIndex, log, vppChan, stopwatch)
}

func (calls *srv6Calls) addDelSteering(delete bool, steering *srv6.Steering, swIfIndex ifaceidx.SwIfIndex,
	log logging.Logger, vppChan VPPChannel, stopwatch *measure.Stopwatch) error {
	defer func(t time.Time) {
		stopwatch.TimeLog(sr.SrSteeringAddDel{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	// defining operation strings for logging
	operationProgressing, operationFinished := "Adding", "Added"
	if delete {
		operationProgressing, operationFinished = "Removing", "Removed"
	}

	// logging info about operation with steering
	if steering.L3Traffic != nil {
		log.Debugf("%v steering for l3 traffic with destination %v to SR policy (binding SID %v, policy index %v)",
			operationProgressing, steering.L3Traffic.PrefixAddress, steering.PolicyBSID, steering.PolicyIndex)
	} else {
		log.Debugf("%v steering for l2 traffic from interface %v to SR policy (binding SID %v, policy index %v)",
			operationProgressing, steering.L2Traffic.InterfaceName, steering.PolicyBSID, steering.PolicyIndex)
	}

	// converting policy reference
	bsidAddr := make([]byte, 0)
	if len(strings.Trim(steering.PolicyBSID, " ")) > 0 {
		bsid, err := parseIPv6(steering.PolicyBSID)
		if err != nil {
			return fmt.Errorf("can't parse binding SID %q to IP address: %v ", steering.PolicyBSID, err)
		}
		bsidAddr = []byte(bsid)
	}

	// converting target traffic info
	var prefixAddr []byte
	steerType := SteerTypeIPv6
	tableID := uint32(0)
	maskWidth := uint32(0)
	intIndex := uint32(0)
	if steering.L3Traffic != nil {
		ip, ipnet, err := net.ParseCIDR(steering.L3Traffic.PrefixAddress)
		if err != nil {
			return fmt.Errorf("can't parse ip prefix %q: %v", steering.L3Traffic.PrefixAddress, err)
		}
		if ip.To4() != nil { //IPv4 address
			steerType = SteerTypeIPv4
		}
		tableID = steering.L3Traffic.FibTableID
		prefixAddr = []byte(ip.To16())
		ms, _ := ipnet.Mask.Size()
		maskWidth = uint32(ms)
	} else if steering.L2Traffic != nil {
		steerType = SteerTypeL2
		interfaceSwIndex, _, exists := swIfIndex.LookupIdx(steering.L2Traffic.InterfaceName)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", steering.L2Traffic.InterfaceName)
		}
		intIndex = interfaceSwIndex
	}
	req := &sr.SrSteeringAddDel{
		IsDel:         boolToInt(delete),
		TableID:       tableID,
		BsidAddr:      bsidAddr,             //policy (to which we want to steer routing into) identified by binding sid (alternativelly it can be used SRPolicyIndex)
		SrPolicyIndex: steering.PolicyIndex, //policy (to which we want to steer routing into) identified by policy index (alternativelly it can be used BsidAddr)
		TrafficType:   steerType,            // type of traffic to steer
		PrefixAddr:    prefixAddr,           // destination prefix address (L3 traffic type only)
		MaskWidth:     maskWidth,            // destination ip prefix mask (L3 traffic type only)
		SwIfIndex:     intIndex,             // incoming interface (L2 traffic type only)
	}
	reply := &sr.SrSteeringAddDelReply{}

	if err := vppChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	log.WithFields(logging.Fields{"steer type": steerType, "L3 prefix address bytes": prefixAddr,
		"L2 interface index": intIndex, "policy binding SID": bsidAddr, "policy index": steering.PolicyIndex}).
		Debugf("%v steering to SR policy ", operationFinished)

	return nil
}

func boolToInt(input bool) uint8 {
	if input {
		return uint8(1)
	}
	return uint8(0)
}

// parseIPv6 parses string <str> to IPv6 address (including IPv4 address converted to IPv6 address)
func parseIPv6(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, fmt.Errorf(" %q is not ip address", str)
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, fmt.Errorf(" %q is not ipv6 address", str)
	}
	return ipv6, nil
}
