// Copyright (c) 2019 Bell Canada, Pantheon Technologies and/or its affiliates.
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

// Package vpp1901 contains wrappers over VPP (version 19.01) binary APIs to simplify their usage
package vpp1901

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/ligato/cn-infra/logging"
	nbint "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/interfaces"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/sr"
)

// Constants for behavior function hardcoded into VPP (there can be also custom behavior functions implemented as VPP plugins)
// Constants are taken from VPP's vnet/srv6/sr.h (names are modified to Golang from original C form in VPP code)
const (
	BehaviorEnd    uint8 = iota + 1 // Behavior of simple endpoint
	BehaviorX                       // Behavior of endpoint with Layer-3 cross-connect
	BehaviorT                       // Behavior of endpoint with specific IPv6 table lookup
	BehaviorDfirst                  // Unused. Separator in between regular and D
	BehaviorDX2                     // Behavior of endpoint with decapulation and Layer-2 cross-connect (or DX2 with egress VLAN rewrite when VLAN notzero - not supported this variant yet)
	BehaviorDX6                     // Behavior of endpoint with decapsulation and IPv6 cross-connect
	BehaviorDX4                     // Behavior of endpoint with decapsulation and IPv4 cross-connect
	BehaviorDT6                     // Behavior of endpoint with decapsulation and specific IPv6 table lookup
	BehaviorDT4                     // Behavior of endpoint with decapsulation and specific IPv4 table lookup
	BehaviorLast                    // seems unused, note in VPP: "Must always be the last one"
)

// Constants for steering type
// Constants are taken from VPP's vnet/srv6/sr.h (names are modified to Golang from original C form in VPP code)
const (
	SteerTypeL2   uint8 = 2
	SteerTypeIPv4 uint8 = 4
	SteerTypeIPv6 uint8 = 6
)

// Constants for operation of SR policy modify binary API method
const (
	AddSRList            uint8 = iota + 1 // Add SR List to an existing SR policy
	DeleteSRList                          // Delete SR List from an existing SR policy
	ModifyWeightOfSRList                  // Modify the weight of an existing SR List
)

// AddLocalSid adds local sid given by <sidAddr> and <localSID> into VPP
func (h *SRv6VppHandler) AddLocalSid(sidAddr net.IP, localSID *srv6.LocalSID) error {
	return h.addDelLocalSid(false, sidAddr, localSID)
}

// DeleteLocalSid delets local sid given by <sidAddr> in VPP
func (h *SRv6VppHandler) DeleteLocalSid(sidAddr net.IP) error {
	return h.addDelLocalSid(true, sidAddr, nil)
}

func (h *SRv6VppHandler) addDelLocalSid(deletion bool, sidAddr net.IP, localSID *srv6.LocalSID) error {
	h.log.WithFields(logging.Fields{"localSID": sidAddr, "delete": deletion, "FIB table ID": h.fibTableID(localSID), "end function": h.endFunction(localSID)}).
		Debug("Adding/deleting Local SID", sidAddr)
	if !deletion && localSID.GetEndFunction_AD() != nil {
		return h.addSRProxy(sidAddr, localSID)
	}
	req := &sr.SrLocalsidAddDel{
		IsDel:    boolToUint(deletion),
		Localsid: sr.Srv6Sid{Addr: []byte(sidAddr)},
	}
	if !deletion {
		req.FibTable = localSID.FibTableId // where to install localsid entry //TODO meaning change? see proto model TODO
		if err := h.writeEndFunction(req, sidAddr, localSID); err != nil {
			return err
		}
	}
	reply := &sr.SrLocalsidAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	h.log.WithFields(logging.Fields{"localSID": sidAddr, "delete": deletion, "FIB table ID": h.fibTableID(localSID), "end function": h.endFunction(localSID)}).
		Debug("Added/deleted Local SID ", sidAddr)

	return nil
}

// addSRProxy adds local sid with SR-proxy end function (End.AD). This functionality has no binary API in VPP, therefore
// CLI commands are used (VPE binary API that calls VPP's CLI).
func (h *SRv6VppHandler) addSRProxy(sidAddr net.IP, localSID *srv6.LocalSID) error {
	// get VPP-internal names of IN and OUT interfaces
	names, err := h.interfaceNameMapping()
	if err != nil {
		return fmt.Errorf("can't convert interface names from etcd to VPP-internal interface names:%v", err)
	}
	outInterface, found := names[localSID.GetEndFunction_AD().OutgoingInterface]
	if !found {
		return fmt.Errorf("can't find VPP-internal name for interface %v (name in etcd)", localSID.GetEndFunction_AD().OutgoingInterface)
	}
	inInterface, found := names[localSID.GetEndFunction_AD().IncomingInterface]
	if !found {
		return fmt.Errorf("can't find VPP-internal name for interface %v (name in etcd)", localSID.GetEndFunction_AD().IncomingInterface)
	}

	// add SR-proxy using VPP CLI
	data, err := h.RunCli(fmt.Sprintf("sr localsid address %v behavior end.ad nh %v oif %v iif %v", sidAddr, localSID.GetEndFunction_AD().ServiceAddress, outInterface, inInterface))
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) > 0 {
		return fmt.Errorf("addition of dynamic segment routing proxy failed by returning nonblank space text in CLI: %v", string(data))
	}
	return nil
}

// interfaceNameMapping dumps from VPP internal names of interfaces and uses them to produce mapping from ligato interface names to vpp internal names.
func (h *SRv6VppHandler) interfaceNameMapping() (map[string]string, error) {
	mapping := make(map[string]string)
	reqCtx := h.callsChannel.SendMultiRequest(&interfaces.SwInterfaceDump{})

	for {
		// get next interface info
		ifDetails := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(ifDetails)
		if stop {
			break // Break from the loop.
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump interface: %v", err)
		}

		// extract and compute names
		ligatoName := string(bytes.SplitN(ifDetails.Tag, []byte{0x00}, 2)[0])
		vppInternalName := string(bytes.SplitN(ifDetails.InterfaceName, []byte{0x00}, 2)[0])
		if ifDetails.SupSwIfIndex == ifDetails.SwIfIndex && // no subinterface (subinterface are not DPDK)
			guessInterfaceType(string(ifDetails.InterfaceName)) == nbint.Interface_DPDK { // fill name for physical interfaces (they are mostly without tag)
			ligatoName = vppInternalName
		}

		mapping[ligatoName] = vppInternalName
	}
	return mapping, nil
}

func (h *SRv6VppHandler) fibTableID(localSID *srv6.LocalSID) string {
	if localSID != nil {
		return string(localSID.FibTableId)
	}
	return "<nil>"
}

func (h *SRv6VppHandler) endFunction(localSID *srv6.LocalSID) string {
	switch ef := localSID.GetEndFunction().(type) {
	case *srv6.LocalSID_BaseEndFunction:
		return fmt.Sprintf("End{psp: %v}", ef.BaseEndFunction.Psp)
	case *srv6.LocalSID_EndFunction_X:
		return fmt.Sprintf("X{psp: %v, OutgoingInterface: %v, NextHop: %v}", ef.EndFunction_X.Psp, ef.EndFunction_X.OutgoingInterface, ef.EndFunction_X.NextHop)
	case *srv6.LocalSID_EndFunction_T:
		return fmt.Sprintf("T{psp: %v}", ef.EndFunction_T.Psp)
	case *srv6.LocalSID_EndFunction_DX2:
		return fmt.Sprintf("DX2{VlanTag: %v, OutgoingInterface: %v, NextHop: %v}", ef.EndFunction_DX2.VlanTag, ef.EndFunction_DX2.OutgoingInterface, ef.EndFunction_DX2.NextHop)
	case *srv6.LocalSID_EndFunction_DX4:
		return fmt.Sprintf("DX4{OutgoingInterface: %v, NextHop: %v}", ef.EndFunction_DX4.OutgoingInterface, ef.EndFunction_DX4.NextHop)
	case *srv6.LocalSID_EndFunction_DX6:
		return fmt.Sprintf("DX6{OutgoingInterface: %v, NextHop: %v}", ef.EndFunction_DX6.OutgoingInterface, ef.EndFunction_DX6.NextHop)
	case *srv6.LocalSID_EndFunction_DT4:
		return fmt.Sprint("DT4")
	case *srv6.LocalSID_EndFunction_DT6:
		return fmt.Sprint("DT6")
	case *srv6.LocalSID_EndFunction_AD:
		return fmt.Sprintf("AD{ServiceAddress: %v, OutgoingInterface: %v, IncomingInterface: %v}", ef.EndFunction_AD.ServiceAddress, ef.EndFunction_AD.OutgoingInterface, ef.EndFunction_AD.IncomingInterface)
	case nil:
		return "<nil>"
	default:
		return "unknown end function"
	}
}

func (h *SRv6VppHandler) writeEndFunction(req *sr.SrLocalsidAddDel, sidAddr net.IP, localSID *srv6.LocalSID) error {
	switch ef := localSID.EndFunction.(type) {
	case *srv6.LocalSID_BaseEndFunction:
		req.Behavior = BehaviorEnd
		req.EndPsp = boolToUint(ef.BaseEndFunction.Psp)
	case *srv6.LocalSID_EndFunction_X:
		req.Behavior = BehaviorX
		req.EndPsp = boolToUint(ef.EndFunction_X.Psp)
		ifMeta, exists := h.ifIndexes.LookupByName(ef.EndFunction_X.OutgoingInterface)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", ef.EndFunction_X.OutgoingInterface)
		}
		req.SwIfIndex = ifMeta.SwIfIndex
		nhAddr, err := parseIPv6(ef.EndFunction_X.NextHop) // parses also ipv4 addresses but into ipv6 address form
		if err != nil {
			return err
		}
		if nhAddr4 := nhAddr.To4(); nhAddr4 != nil { // ipv4 address in ipv6 address form?
			req.NhAddr4 = nhAddr4
		} else {
			req.NhAddr6 = []byte(nhAddr)
		}
	case *srv6.LocalSID_EndFunction_T:
		req.Behavior = BehaviorT
		req.EndPsp = boolToUint(ef.EndFunction_T.Psp)
	case *srv6.LocalSID_EndFunction_DX2:
		req.Behavior = BehaviorDX2
		req.VlanIndex = ef.EndFunction_DX2.VlanTag
		ifMeta, exists := h.ifIndexes.LookupByName(ef.EndFunction_DX2.OutgoingInterface)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", ef.EndFunction_DX2.OutgoingInterface)
		}
		req.SwIfIndex = ifMeta.SwIfIndex
		nhAddr, err := parseIPv6(ef.EndFunction_DX2.NextHop) // parses also ipv4 addresses but into ipv6 address form
		if err != nil {
			return err
		}
		if nhAddr4 := nhAddr.To4(); nhAddr4 != nil { // ipv4 address in ipv6 address form?
			req.NhAddr4 = nhAddr4
		} else {
			req.NhAddr6 = []byte(nhAddr)
		}
	case *srv6.LocalSID_EndFunction_DX4:
		req.Behavior = BehaviorDX4
		ifMeta, exists := h.ifIndexes.LookupByName(ef.EndFunction_DX4.OutgoingInterface)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", ef.EndFunction_DX4.OutgoingInterface)
		}
		req.SwIfIndex = ifMeta.SwIfIndex
		nhAddr, err := parseIPv6(ef.EndFunction_DX4.NextHop) // parses also IPv4
		if err != nil {
			return err
		}
		nhAddr4 := nhAddr.To4()
		if nhAddr4 == nil {
			return fmt.Errorf("next hop of DX4 end function (%v) is not valid IPv4 address", ef.EndFunction_DX4.NextHop)
		}
		req.NhAddr4 = []byte(nhAddr4)
	case *srv6.LocalSID_EndFunction_DX6:
		req.Behavior = BehaviorDX6
		ifMeta, exists := h.ifIndexes.LookupByName(ef.EndFunction_DX6.OutgoingInterface)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", ef.EndFunction_DX6.OutgoingInterface)
		}
		req.SwIfIndex = ifMeta.SwIfIndex
		nhAddr6, err := parseIPv6(ef.EndFunction_DX6.NextHop)
		if err != nil {
			return err
		}
		req.NhAddr6 = []byte(nhAddr6)
	case *srv6.LocalSID_EndFunction_DT4:
		req.Behavior = BehaviorDT4
	case *srv6.LocalSID_EndFunction_DT6:
		req.Behavior = BehaviorDT6
	case nil:
		return fmt.Errorf("End function not set. Please configure end function for local SID %v ", sidAddr)
	default:
		return fmt.Errorf("unknown end function (model link type %T)", ef) // EndFunction_AD is handled elsewhere
	}

	return nil
}

// SetEncapsSourceAddress sets for SRv6 in VPP the source address used for encapsulated packet
func (h *SRv6VppHandler) SetEncapsSourceAddress(address string) error {
	h.log.Debugf("Configuring encapsulation source address to address %v", address)
	ipAddress, err := parseIPv6(address)
	if err != nil {
		return err
	}
	req := &sr.SrSetEncapSource{
		EncapsSource: []byte(ipAddress),
	}
	reply := &sr.SrSetEncapSourceReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	h.log.WithFields(logging.Fields{"Encapsulation source address": address}).
		Debug("Encapsulation source address configured.")

	return nil
}

// AddPolicy adds SRv6 policy given by identified <bindingSid>,initial segment list for policy <policySegmentList> and other policy settings in <policy>
func (h *SRv6VppHandler) AddPolicy(bindingSid net.IP, policy *srv6.Policy, segmentList *srv6.PolicySegmentList) error {
	h.log.Debugf("Adding SR policy with binding SID %v and list of next SIDs %v", bindingSid, segmentList.Segments)
	sids, err := h.convertPolicySegment(segmentList)
	if err != nil {
		return err
	}
	// Note: Weight in sr.SrPolicyAdd is leftover from API changes that moved weight into sr.Srv6SidList (it is weight of sid list not of the whole policy)
	req := &sr.SrPolicyAdd{
		BsidAddr: []byte(bindingSid),
		Sids:     *sids,
		IsEncap:  boolToUint(policy.SrhEncapsulation),
		Type:     boolToUint(policy.SprayBehaviour),
		FibTable: policy.FibTableId,
	}
	reply := &sr.SrPolicyAddReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	h.log.WithFields(logging.Fields{"binding SID": bindingSid, "list of next SIDs": segmentList.Segments}).
		Debug("SR policy added")

	return nil
}

// DeletePolicy deletes SRv6 policy given by binding SID <bindingSid>
func (h *SRv6VppHandler) DeletePolicy(bindingSid net.IP) error {
	h.log.Debugf("Deleting SR policy with binding SID %v ", bindingSid)
	req := &sr.SrPolicyDel{
		BsidAddr: sr.Srv6Sid{Addr: []byte(bindingSid)}, // TODO add ability to define policy also by index (SrPolicyIndex)
	}
	reply := &sr.SrPolicyDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	h.log.WithFields(logging.Fields{"binding SID": bindingSid}).Debug("SR policy deleted")

	return nil
}

// AddPolicySegmentList adds segment list <segmentList> to SRv6 policy <policy> that has policy BSID <bindingSid>
func (h *SRv6VppHandler) AddPolicySegmentList(bindingSid net.IP, policy *srv6.Policy, segmentList *srv6.PolicySegmentList) error {
	h.log.Debugf("Adding segment %v to SR policy with binding SID %v", segmentList.Segments, bindingSid)
	err := h.modPolicy(AddSRList, bindingSid, policy, segmentList, 0)
	if err == nil {
		h.log.WithFields(logging.Fields{"binding SID": bindingSid, "list of next SIDs": segmentList.Segments}).
			Debug("SR policy modified(added another segment list)")
	}
	return err
}

// DeletePolicySegmentList removes segment list <segmentList> (with segment list index <segmentListIndex>) from SRv6 policy <policy> that has policy BSID <bindingSid>
func (h *SRv6VppHandler) DeletePolicySegmentList(bindingSid net.IP, policy *srv6.Policy, segmentList *srv6.PolicySegmentList, segmentListIndex uint32) error {
	h.log.Debugf("Removing segment %v (index %v) from SR policy with binding SID %v", segmentList.Segments, segmentListIndex, bindingSid)
	err := h.modPolicy(DeleteSRList, bindingSid, policy, segmentList, segmentListIndex)
	if err == nil {
		h.log.WithFields(logging.Fields{"binding SID": bindingSid, "list of next SIDs": segmentList.Segments, "segmentListIndex": segmentListIndex}).
			Debug("SR policy modified(removed segment list)")
	}
	return err
}

func (h *SRv6VppHandler) modPolicy(operation uint8, bindingSid net.IP, policy *srv6.Policy, segmentList *srv6.PolicySegmentList, segmentListIndex uint32) error {
	sids, err := h.convertPolicySegment(segmentList)
	if err != nil {
		return err
	}
	// Note: Weight in sr.SrPolicyMod is leftover from API changes that moved weight into sr.Srv6SidList (it is weight of sid list not of the whole policy)
	req := &sr.SrPolicyMod{
		BsidAddr:  []byte(bindingSid), // TODO add ability to define policy also by index (SrPolicyIndex)
		Operation: operation,
		Sids:      *sids,
		FibTable:  policy.FibTableId,
	}
	if operation == DeleteSRList || operation == ModifyWeightOfSRList {
		req.SlIndex = segmentListIndex
	}

	reply := &sr.SrPolicyModReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}
	return nil
}

func (h *SRv6VppHandler) convertPolicySegment(segmentList *srv6.PolicySegmentList) (*sr.Srv6SidList, error) {
	var segments []sr.Srv6Sid
	for _, sid := range segmentList.Segments {
		// parse to IPv6 address
		parserSid, err := parseIPv6(sid)
		if err != nil {
			return nil, err
		}

		// add sid to segment list
		ipv6Segment := sr.Srv6Sid{
			Addr: make([]byte, 16), // sr.Srv6Sid.Addr = [16]byte
		}
		copy(ipv6Segment.Addr, parserSid)
		segments = append(segments, ipv6Segment)
	}
	return &sr.Srv6SidList{
		NumSids: uint8(len(segments)),
		Sids:    segments,
		Weight:  segmentList.Weight,
	}, nil
}

// RetrievePolicySegmentIndex retrieves index of segment list <segmentList>
func (h *SRv6VppHandler) RetrievePolicySegmentIndex(segmentList *srv6.PolicySegmentList) (uint32, error) {
	// dump sr policies using VPP CLI
	data, err := h.RunCli("sh sr policies")
	if err != nil {
		return 0, fmt.Errorf("can't dump index data from VPP: %v", err)
	}

	// do necessary parsing to extract index of segment list
	dumpStr := strings.ToLower(string(data))
	policyHeader := fmt.Sprintf("bsid: %v", strings.ToLower(segmentList.GetPolicyBsid()))
	slRE := regexp.MustCompile(fmt.Sprintf("\\[(\\d+)\\].- < %s,[^:>]*> weight: %d", strings.ToLower(strings.Join(segmentList.Segments, ", ")), segmentList.Weight))
	for _, policyStr := range strings.Split(dumpStr, "-----------") {
		if strings.Contains(policyStr, policyHeader) {
			if match := slRE.FindStringSubmatch(policyStr); match != nil {
				parsed, err := strconv.ParseUint(match[1], 10, 32)
				if err != nil {
					return 0, fmt.Errorf("can't parse segment policy index %q", match[1])
				}
				return uint32(parsed), nil
			}
		}
	}
	return 0, fmt.Errorf("can't find index for policy segment list %v in dump %q", segmentList, dumpStr)
}

// AddSteering sets in VPP steering into SRv6 policy.
func (h *SRv6VppHandler) AddSteering(steering *srv6.Steering) error {
	return h.addDelSteering(false, steering)
}

// RemoveSteering removes in VPP steering into SRv6 policy.
func (h *SRv6VppHandler) RemoveSteering(steering *srv6.Steering) error {
	return h.addDelSteering(true, steering)
}

func (h *SRv6VppHandler) addDelSteering(delete bool, steering *srv6.Steering) error {
	// defining operation strings for logging
	operationProgressing, operationFinished := "Adding", "Added"
	if delete {
		operationProgressing, operationFinished = "Removing", "Removed"
	}

	// logging info about operation with steering
	switch t := steering.Traffic.(type) {
	case *srv6.Steering_L3Traffic_:
		h.log.Debugf("%v steering for l3 traffic with destination %v to SR policy (binding SID %v, policy index %v)",
			operationProgressing, t.L3Traffic.PrefixAddress, steering.GetPolicyBsid(), steering.GetPolicyIndex())
	case *srv6.Steering_L2Traffic_:
		h.log.Debugf("%v steering for l2 traffic from interface %v to SR policy (binding SID %v, policy index %v)",
			operationProgressing, t.L2Traffic.InterfaceName, steering.GetPolicyBsid(), steering.GetPolicyIndex())
	}

	// converting policy reference
	var policyBSIDAddr []byte   // undefined reference
	var policyIndex = uint32(0) // undefined reference
	switch ref := steering.PolicyRef.(type) {
	case *srv6.Steering_PolicyBsid:
		bsid, err := parseIPv6(ref.PolicyBsid)
		if err != nil {
			return fmt.Errorf("can't parse binding SID %q to IP address: %v ", ref.PolicyBsid, err)
		}
		policyBSIDAddr = []byte(bsid)
	case *srv6.Steering_PolicyIndex:
		policyIndex = ref.PolicyIndex
	case nil:
		return fmt.Errorf("policy reference must be provided")
	default:
		return fmt.Errorf("unknown policy reference type (link type %+v)", ref)
	}

	// converting target traffic info
	var prefixAddr []byte
	steerType := SteerTypeIPv6
	tableID := uint32(0)
	maskWidth := uint32(0)
	intIndex := uint32(0)
	switch t := steering.Traffic.(type) {
	case *srv6.Steering_L3Traffic_:
		ip, ipnet, err := net.ParseCIDR(t.L3Traffic.PrefixAddress)
		if err != nil {
			return fmt.Errorf("can't parse ip prefix %q: %v", t.L3Traffic.PrefixAddress, err)
		}
		if ip.To4() != nil { // IPv4 address
			steerType = SteerTypeIPv4
		}
		tableID = t.L3Traffic.FibTableId
		prefixAddr = []byte(ip.To16())
		ms, _ := ipnet.Mask.Size()
		maskWidth = uint32(ms)
	case *srv6.Steering_L2Traffic_:
		steerType = SteerTypeL2
		ifMeta, exists := h.ifIndexes.LookupByName(t.L2Traffic.InterfaceName)
		if !exists {
			return fmt.Errorf("for interface %v doesn't exist sw index", t.L2Traffic.InterfaceName)
		}
		intIndex = ifMeta.SwIfIndex
	case nil:
		return fmt.Errorf("traffic type must be provided")
	default:
		return fmt.Errorf("unknown traffic type (link type %+v)", t)
	}
	req := &sr.SrSteeringAddDel{
		IsDel:         boolToUint(delete),
		TableID:       tableID,
		BsidAddr:      policyBSIDAddr, // policy (to which we want to steer routing into) identified by policy binding sid (alternativelly it can be used policy index)
		SrPolicyIndex: policyIndex,    // policy (to which we want to steer routing into) identified by policy index (alternativelly it can be used policy binding sid)
		TrafficType:   steerType,      // type of traffic to steer
		PrefixAddr:    prefixAddr,     // destination prefix address (L3 traffic type only)
		MaskWidth:     maskWidth,      // destination ip prefix mask (L3 traffic type only)
		SwIfIndex:     intIndex,       // incoming interface (L2 traffic type only)
	}
	reply := &sr.SrSteeringAddDelReply{}

	if err := h.callsChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}
	if reply.Retval != 0 {
		return fmt.Errorf("vpp call %q returned: %d", reply.GetMessageName(), reply.Retval)
	}

	h.log.WithFields(logging.Fields{"steer type": steerType, "L3 prefix address bytes": prefixAddr,
		"L2 interface index": intIndex, "policy binding SID": policyBSIDAddr, "policy index": policyIndex}).
		Debugf("%v steering to SR policy ", operationFinished)

	return nil
}

func boolToUint(input bool) uint8 {
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

// guessInterfaceType attempts to guess the correct interface type from its internal name (as given by VPP).
// This is required mainly for those interface types, that do not provide dump binary API,
// such as loopback of af_packet.
func guessInterfaceType(ifName string) nbint.Interface_Type {
	switch {
	case strings.HasPrefix(ifName, "loop"),
		strings.HasPrefix(ifName, "local"):
		return nbint.Interface_SOFTWARE_LOOPBACK
	case strings.HasPrefix(ifName, "memif"):
		return nbint.Interface_MEMIF
	case strings.HasPrefix(ifName, "tap"):
		return nbint.Interface_TAP
	case strings.HasPrefix(ifName, "host"):
		return nbint.Interface_AF_PACKET
	case strings.HasPrefix(ifName, "vxlan"):
		return nbint.Interface_VXLAN_TUNNEL
	case strings.HasPrefix(ifName, "ipsec"):
		return nbint.Interface_IPSEC_TUNNEL
	case strings.HasPrefix(ifName, "vmxnet3"):
		return nbint.Interface_VMXNET3_INTERFACE
	default:
		return nbint.Interface_DPDK
	}
}
