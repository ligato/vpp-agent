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

package descriptor

import (
	"net"
	"reflect"
	"strings"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/pkg/errors"
	scheduler "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/srplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"
)

const (
	// LocalSIDDescriptorName is the name of the descriptor for VPP LocalSIDs
	LocalSIDDescriptorName = "vpp-sr-localsid"

	// dependency labels
	localsidOutgoingInterfaceDep = "sr-localsid-outgoing-interface-exists"
	localsidIncomingInterfaceDep = "sr-localsid-incoming-interface-exists"
	localsidInstallationVRFDep   = "sr-localsid-installation-vrf-table-exists"
	localsidLookupVRFDep         = "sr-localsid-routing-lookup-vrf-table-exists"
)

// LocalSIDDescriptor teaches KVScheduler how to configure VPP LocalSIDs.
type LocalSIDDescriptor struct {
	// dependencies
	log       logging.Logger
	srHandler vppcalls.SRv6VppAPI
}

// NewLocalSIDDescriptor creates a new instance of the LocalSID descriptor.
func NewLocalSIDDescriptor(srHandler vppcalls.SRv6VppAPI, log logging.PluginLogger) *scheduler.KVDescriptor {
	ctx := &LocalSIDDescriptor{
		log:       log.NewLogger("localsid-descriptor"),
		srHandler: srHandler,
	}

	typedDescr := &adapter.LocalSIDDescriptor{
		Name:            LocalSIDDescriptorName,
		NBKeyPrefix:     srv6.ModelLocalSID.KeyPrefix(),
		ValueTypeName:   srv6.ModelLocalSID.ProtoName(),
		KeySelector:     srv6.ModelLocalSID.IsKeyValid,
		KeyLabel:        srv6.ModelLocalSID.StripKeyPrefix,
		ValueComparator: ctx.EquivalentLocalSIDs,
		Validate:        ctx.Validate,
		Create:          ctx.Create,
		Delete:          ctx.Delete,
		Dependencies:    ctx.Dependencies,
	}
	return adapter.NewLocalSIDDescriptor(typedDescr)
}

// EquivalentLocalSIDs determines whether 2 localSIDs are logically equal. This comparison takes into consideration also
// semantics that couldn't be modeled into proto models (i.e. SID is IPv6 address and not only string)
func (d *LocalSIDDescriptor) EquivalentLocalSIDs(key string, oldLocalSID, newLocalSID *srv6.LocalSID) bool {
	return oldLocalSID.InstallationVrfId == newLocalSID.InstallationVrfId &&
		equivalentSIDs(oldLocalSID.Sid, newLocalSID.Sid) &&
		d.equivalentEndFunctions(oldLocalSID.EndFunction, newLocalSID.EndFunction)
}

func (d *LocalSIDDescriptor) equivalentEndFunctions(ef1, ef2 interface{}) bool {
	if ef1 == nil || ef2 == nil {
		return ef1 == ef2
	}
	if reflect.TypeOf(ef1) != reflect.TypeOf(ef2) {
		return false
	}
	switch ef1typed := ef1.(type) {
	case *srv6.LocalSID_BaseEndFunction:
		return true
	case *srv6.LocalSID_EndFunctionX:
		return ef1typed.EndFunctionX.Psp == ef2.(*srv6.LocalSID_EndFunctionX).EndFunctionX.Psp &&
			equivalentIPv6(ef1typed.EndFunctionX.NextHop, ef2.(*srv6.LocalSID_EndFunctionX).EndFunctionX.NextHop) &&
			equivalentTrimmedLowered(ef1typed.EndFunctionX.OutgoingInterface, ef2.(*srv6.LocalSID_EndFunctionX).EndFunctionX.OutgoingInterface)
	case *srv6.LocalSID_EndFunctionT:
		return ef1typed.EndFunctionT.Psp == ef2.(*srv6.LocalSID_EndFunctionT).EndFunctionT.Psp &&
			ef1typed.EndFunctionT.VrfId == ef2.(*srv6.LocalSID_EndFunctionT).EndFunctionT.VrfId
	case *srv6.LocalSID_EndFunctionDx2:
		return ef1typed.EndFunctionDx2.VlanTag == ef2.(*srv6.LocalSID_EndFunctionDx2).EndFunctionDx2.VlanTag &&
			equivalentTrimmedLowered(ef1typed.EndFunctionDx2.OutgoingInterface, ef2.(*srv6.LocalSID_EndFunctionDx2).EndFunctionDx2.OutgoingInterface)
	case *srv6.LocalSID_EndFunctionDx4:
		return equivalentIPv4(ef1typed.EndFunctionDx4.NextHop, ef2.(*srv6.LocalSID_EndFunctionDx4).EndFunctionDx4.NextHop) &&
			equivalentTrimmedLowered(ef1typed.EndFunctionDx4.OutgoingInterface, ef2.(*srv6.LocalSID_EndFunctionDx4).EndFunctionDx4.OutgoingInterface)
	case *srv6.LocalSID_EndFunctionDx6:
		return equivalentIPv4(ef1typed.EndFunctionDx6.NextHop, ef2.(*srv6.LocalSID_EndFunctionDx6).EndFunctionDx6.NextHop) &&
			equivalentTrimmedLowered(ef1typed.EndFunctionDx6.OutgoingInterface, ef2.(*srv6.LocalSID_EndFunctionDx6).EndFunctionDx6.OutgoingInterface)
	case *srv6.LocalSID_EndFunctionDt4:
		return ef1typed.EndFunctionDt4.VrfId == ef2.(*srv6.LocalSID_EndFunctionDt4).EndFunctionDt4.VrfId
	case *srv6.LocalSID_EndFunctionDt6:
		return ef1typed.EndFunctionDt6.VrfId == ef2.(*srv6.LocalSID_EndFunctionDt6).EndFunctionDt6.VrfId
	case *srv6.LocalSID_EndFunctionAd:
		return equivalentTrimmedLowered(ef1typed.EndFunctionAd.OutgoingInterface, ef2.(*srv6.LocalSID_EndFunctionAd).EndFunctionAd.OutgoingInterface) &&
			equivalentTrimmedLowered(ef1typed.EndFunctionAd.IncomingInterface, ef2.(*srv6.LocalSID_EndFunctionAd).EndFunctionAd.IncomingInterface) &&
			(equivalentIPv4(ef1typed.EndFunctionAd.L3ServiceAddress, ef2.(*srv6.LocalSID_EndFunctionAd).EndFunctionAd.L3ServiceAddress) || // l3 ipv4 service
				equivalentIPv6(ef1typed.EndFunctionAd.L3ServiceAddress, ef2.(*srv6.LocalSID_EndFunctionAd).EndFunctionAd.L3ServiceAddress) || // l3 ipv6 service
				(strings.TrimSpace(ef1typed.EndFunctionAd.L3ServiceAddress) == "" && strings.TrimSpace(ef2.(*srv6.LocalSID_EndFunctionAd).EndFunctionAd.L3ServiceAddress) == "")) // l2 service
	default:
		d.log.Warn("EquivalentSteering found unknown end function type (%T). Using general reflect.DeepEqual for it.", ef1)
		return reflect.DeepEqual(ef1, ef2) // unknown end function type
	}
}

// Validate validates VPP LocalSIDs.
func (d *LocalSIDDescriptor) Validate(key string, localSID *srv6.LocalSID) error {
	// checking basic attributes
	_, err := ParseIPv6(localSID.GetSid())
	if err != nil {
		return scheduler.NewInvalidValueError(errors.Errorf("failed to parse local sid %s, should be a valid ipv6 address: %v", localSID.GetSid(), err), "sid")
	}
	if localSID.GetInstallationVrfId() < 0 {
		return scheduler.NewInvalidValueError(errors.Errorf("installation vrf id can't be lower than zero, input value %v", localSID.GetInstallationVrfId()), "installationVrfId")
	}

	// checking end functions
	switch ef := localSID.EndFunction.(type) {
	case *srv6.LocalSID_BaseEndFunction:
	case *srv6.LocalSID_EndFunctionX:
		_, err := ParseIPv6(ef.EndFunctionX.NextHop)
		if err != nil {
			return scheduler.NewInvalidValueError(errors.Errorf("failed to parse next hop %s, should be a valid ipv6 address: %v", ef.EndFunctionX.NextHop, err), "EndFunctionX.NextHop")
		}
	case *srv6.LocalSID_EndFunctionT:
	case *srv6.LocalSID_EndFunctionDx2:
	case *srv6.LocalSID_EndFunctionDx4:
		_, err := ParseIPv4(ef.EndFunctionDx4.NextHop)
		if err != nil {
			return scheduler.NewInvalidValueError(errors.Errorf("failed to parse next hop %s, should be a valid ipv4 address: %v", ef.EndFunctionDx4.NextHop, err), "EndFunctionDX4.NextHop")
		}
	case *srv6.LocalSID_EndFunctionDx6:
		_, err := ParseIPv6(ef.EndFunctionDx6.NextHop)
		if err != nil {
			return scheduler.NewInvalidValueError(errors.Errorf("failed to parse next hop %s, should be a valid ipv6 address: %v", ef.EndFunctionDx6.NextHop, err), "EndFunctionDX6.NextHop")
		}
	case *srv6.LocalSID_EndFunctionDt4:
	case *srv6.LocalSID_EndFunctionDt6:
	case *srv6.LocalSID_EndFunctionAd:
		if strings.TrimSpace(ef.EndFunctionAd.L3ServiceAddress) == "" {
			return nil // l2 service
		}
		// l3 service
		ip := net.ParseIP(ef.EndFunctionAd.L3ServiceAddress)
		if ip == nil {
			return scheduler.NewInvalidValueError(errors.Errorf("failed to parse service address %s, should be a valid ip address(ipv4 or ipv6) or empty(case of l2 service): %v", ef.EndFunctionAd.L3ServiceAddress, err), "EndFunctionAD.L3ServiceAddress")
		}
	case nil:
		return scheduler.NewInvalidValueError(errors.New("end function must be provided"), "endfunction")
	default:
		return scheduler.NewInvalidValueError(errors.Errorf("end function has unexpected model link type %T", ef), "endfunction")
	}

	return nil
}

// Create creates new Local SID into VPP using VPP's binary api
func (d *LocalSIDDescriptor) Create(key string, value *srv6.LocalSID) (metadata interface{}, err error) {
	if err := d.srHandler.AddLocalSid(value); err != nil {
		return nil, errors.Errorf("failed to add local sid %s: %v", value.GetSid(), err)
	}
	return nil, nil
}

// Delete removes Local SID from VPP using VPP's binary api
func (d *LocalSIDDescriptor) Delete(key string, value *srv6.LocalSID, metadata interface{}) error {
	if err := d.srHandler.DeleteLocalSid(value); err != nil {
		return errors.Errorf("failed to delete local sid %s: %v", value.GetSid(), err)
	}
	return nil
}

// Dependencies for LocalSIDs are represented by interface (interface in up state)
func (d *LocalSIDDescriptor) Dependencies(key string, localSID *srv6.LocalSID) (dependencies []scheduler.Dependency) {
	dependencies = append(dependencies, scheduler.Dependency{
		Label: localsidInstallationVRFDep,
		Key:   vpp_l3.VrfTableKey(localSID.InstallationVrfId, vpp_l3.VrfTable_IPV6),
	})

	switch ef := localSID.EndFunction.(type) {
	case *srv6.LocalSID_EndFunctionT:
		if ef.EndFunctionT.VrfId != 0 { // VRF 0 is in VPP by default
			dependencies = append(dependencies, scheduler.Dependency{
				Label: localsidLookupVRFDep,
				Key:   vpp_l3.VrfTableKey(ef.EndFunctionT.VrfId, vpp_l3.VrfTable_IPV6), // T refers to IPv6 VRF table
			})
		}
	case *srv6.LocalSID_EndFunctionX:
		dependencies = append(dependencies, scheduler.Dependency{
			Label: localsidOutgoingInterfaceDep,
			Key:   interfaces.InterfaceKey(ef.EndFunctionX.OutgoingInterface),
		})
	case *srv6.LocalSID_EndFunctionDx2:
		dependencies = append(dependencies, scheduler.Dependency{
			Label: localsidOutgoingInterfaceDep,
			Key:   interfaces.InterfaceKey(ef.EndFunctionDx2.OutgoingInterface),
		})
	case *srv6.LocalSID_EndFunctionDx4:
		dependencies = append(dependencies, scheduler.Dependency{
			Label: localsidOutgoingInterfaceDep,
			Key:   interfaces.InterfaceKey(ef.EndFunctionDx4.OutgoingInterface),
		})
	case *srv6.LocalSID_EndFunctionDx6:
		dependencies = append(dependencies, scheduler.Dependency{
			Label: localsidOutgoingInterfaceDep,
			Key:   interfaces.InterfaceKey(ef.EndFunctionDx6.OutgoingInterface),
		})
	case *srv6.LocalSID_EndFunctionDt4:
		if ef.EndFunctionDt4.VrfId != 0 { // VRF 0 is in VPP by default
			dependencies = append(dependencies, scheduler.Dependency{
				Label: localsidLookupVRFDep,
				Key:   vpp_l3.VrfTableKey(ef.EndFunctionDt4.VrfId, vpp_l3.VrfTable_IPV4), // we want ipv4 VRF because DT4
			})
		}
	case *srv6.LocalSID_EndFunctionDt6:
		if ef.EndFunctionDt6.VrfId != 0 { // VRF 0 is in VPP by default
			dependencies = append(dependencies, scheduler.Dependency{
				Label: localsidLookupVRFDep,
				Key:   vpp_l3.VrfTableKey(ef.EndFunctionDt6.VrfId, vpp_l3.VrfTable_IPV6), // we want ipv6 VRF because DT6
			})
		}
	case *srv6.LocalSID_EndFunctionAd:
		dependencies = append(dependencies, scheduler.Dependency{
			Label: localsidOutgoingInterfaceDep,
			Key:   interfaces.InterfaceKey(ef.EndFunctionAd.OutgoingInterface),
		})
		dependencies = append(dependencies, scheduler.Dependency{
			Label: localsidIncomingInterfaceDep,
			Key:   interfaces.InterfaceKey(ef.EndFunctionAd.IncomingInterface),
		})
	}

	return dependencies
}

func (d *LocalSIDDescriptor) isIPv4RouteKey(key string) bool {
	isIPv6, err := isRouteDstIpv6(key)
	if err != nil {
		d.log.Debug("Can't determine whether key %v is for ipv4 route or not due to: %v", key, err)
		return false // it fails also in route creation (vpp_calls) and it is before needed vrf creation
	}
	return !isIPv6
}

func (d *LocalSIDDescriptor) isIPv6RouteKey(key string) bool {
	isIPv6, err := isRouteDstIpv6(key)
	if err != nil {
		d.log.Debug("Can't determine whether key %v is for ipv6 route or not due to: %v", key, err)
		return false // it fails also in route creation (vpp_calls) and it is before needed vrf creation
	}
	return isIPv6
}

// ParseIPv6 parses string <str> to IPv6 address (including IPv4 address converted to IPv6 address)
func ParseIPv6(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, errors.Errorf(" %q is not ip address", str)
	}
	ipv6 := ip.To16()
	if ipv6 == nil {
		return nil, errors.Errorf(" %q is not ipv6 address", str)
	}
	return ipv6, nil
}

// ParseIPv4 parses string <str> to IPv4 address
func ParseIPv4(str string) (net.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return nil, errors.Errorf(" %q is not ip address", str)
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil, errors.Errorf(" %q is not ipv4 address", str)
	}
	return ipv4, nil
}

func isRouteDstIpv6(key string) (bool, error) {
	_, _, dstNet, _, isRouteKey := vpp_l3.ParseRouteKey(key)
	if !isRouteKey {
		return false, errors.Errorf("Key %v is not route key", key)
	}
	_, isIPv6, err := addrs.ParseIPWithPrefix(dstNet)
	return isIPv6, err
}

func equivalentSIDs(sid1, sid2 string) bool {
	return equivalentIPv6(sid1, sid2)
}

func equivalentIPv6(ip1Str, ip2str string) bool {
	ip1, err1 := ParseIPv6(ip1Str)
	ip2, err2 := ParseIPv6(ip2str)
	if err1 != nil || err2 != nil { // one of values is invalid, but that will handle validator -> compare by strings
		return equivalentTrimmedLowered(ip1Str, ip2str)
	}
	return ip1.Equal(ip2) // form doesn't matter, are they representig the same IP value ?
}

func equivalentIPv4(ip1str, ip2str string) bool {
	ip1, err1 := ParseIPv4(ip1str)
	ip2, err2 := ParseIPv4(ip2str)
	if err1 != nil || err2 != nil { // one of values is invalid, but that will handle validator -> compare by strings
		return equivalentTrimmedLowered(ip1str, ip2str)
	}
	return ip1.Equal(ip2) // form doesn't matter, are they representig the same IP value ?
}

func equivalentTrimmedLowered(str1, str2 string) bool {
	return strings.TrimSpace(strings.ToLower(str1)) == strings.TrimSpace(strings.ToLower(str2))
}
