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
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/ligato/cn-infra/logging"
	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/srplugin/vppcalls"
	"github.com/pkg/errors"
)

const (
	// SteeringDescriptorName is the name of the descriptor for VPP SRv6 steering
	SteeringDescriptorName = "vpp-sr-steering"

	// dependency labels
	policyExistsDep = "sr-policy-for-steering-exists"
)

// SteeringDescriptor teaches KVScheduler how to configure VPP SRv6 steering.
type SteeringDescriptor struct {
	// dependencies
	log       logging.Logger
	srHandler vppcalls.SRv6VppAPI
}

// NewSteeringDescriptor creates a new instance of the Srv6 steering descriptor.
func NewSteeringDescriptor(srHandler vppcalls.SRv6VppAPI, log logging.PluginLogger) *SteeringDescriptor {
	return &SteeringDescriptor{
		log:       log.NewLogger("steering-descriptor"),
		srHandler: srHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *SteeringDescriptor) GetDescriptor() *adapter.SteeringDescriptor {
	return &adapter.SteeringDescriptor{
		Name:            SteeringDescriptorName,
		NBKeyPrefix:     srv6.ModelSteering.KeyPrefix(),
		ValueTypeName:   srv6.ModelSteering.ProtoName(),
		KeySelector:     srv6.ModelSteering.IsKeyValid,
		KeyLabel:        srv6.ModelSteering.StripKeyPrefix,
		ValueComparator: d.EquivalentSteering,
		Validate:        d.Validate,
		Create:          d.Create,
		Delete:          d.Delete,
		Dependencies:    d.Dependencies,
	}
}

// EquivalentSteering determines whether 2 steerings are logically equal. This comparison takes into consideration also
// semantics that couldn't be modeled into proto models (i.e. SID is IPv6 address and not only string)
func (d *SteeringDescriptor) EquivalentSteering(key string, oldSteering, newSteering *srv6.Steering) bool {
	return strings.TrimSpace(oldSteering.Name) == strings.TrimSpace(newSteering.Name) &&
		d.equivalentPolicy(oldSteering.PolicyRef, newSteering.PolicyRef) &&
		d.equivalentTraffic(oldSteering.Traffic, newSteering.Traffic)
}

func (d *SteeringDescriptor) equivalentPolicy(policy1, policy2 interface{}) bool {
	if policy1 == nil || policy2 == nil {
		return policy1 == policy2
	}
	if reflect.TypeOf(policy1) != reflect.TypeOf(policy2) {
		return false
	}
	switch policy1typed := policy1.(type) {
	case *srv6.Steering_PolicyBsid:
		return equivalentSIDs(policy1typed.PolicyBsid, policy2.(*srv6.Steering_PolicyBsid).PolicyBsid)
	case *srv6.Steering_PolicyIndex:
		return policy1typed.PolicyIndex == policy2.(*srv6.Steering_PolicyIndex).PolicyIndex
	default:
		d.log.Warn("EquivalentSteering found unknown policy reference type (%T). Using general reflect.DeepEqual for it.", policy1)
		return reflect.DeepEqual(policy1, policy2) // unknown policies
	}
}

func (d *SteeringDescriptor) equivalentTraffic(traffic1, traffic2 interface{}) bool {
	if traffic1 == nil || traffic2 == nil {
		return traffic1 == traffic2
	}
	if reflect.TypeOf(traffic1) != reflect.TypeOf(traffic2) {
		return false
	}
	switch traffic1typed := traffic1.(type) {
	case *srv6.Steering_L3Traffic_:
		if traffic1typed.L3Traffic.FibTableId != traffic2.(*srv6.Steering_L3Traffic_).L3Traffic.FibTableId {
			return false
		}
		ip1, ipNet1, err1 := net.ParseCIDR(traffic1typed.L3Traffic.PrefixAddress)
		ip2, ipNet2, err2 := net.ParseCIDR(traffic2.(*srv6.Steering_L3Traffic_).L3Traffic.PrefixAddress)
		if err1 != nil || err2 != nil {
			return equivalentTrimmedLowered(traffic1typed.L3Traffic.PrefixAddress, traffic2.(*srv6.Steering_L3Traffic_).L3Traffic.PrefixAddress) // invalid one of prefixes, but still can equal on string basis
		}
		return ip1.Equal(ip2) && ipNet1.IP.Equal(ipNet2.IP) && ipNet1.Mask.String() == ipNet2.Mask.String()
	case *srv6.Steering_L2Traffic_:
		return equivalentTrimmedLowered(traffic1typed.L2Traffic.InterfaceName, traffic2.(*srv6.Steering_L2Traffic_).L2Traffic.InterfaceName)
	default:
		d.log.Warn("EquivalentSteering found unknown traffic type (%T). Using general reflect.DeepEqual for it.", traffic1)
		return reflect.DeepEqual(traffic1, traffic2) // unknown policies
	}
}

// Validate validates VPP SRv6 steerings.
func (d *SteeringDescriptor) Validate(key string, steering *srv6.Steering) error {
	// checking policy reference ("oneof" field)
	switch ref := steering.PolicyRef.(type) {
	case *srv6.Steering_PolicyBsid:
		_, err := ParseIPv6(ref.PolicyBsid)
		if err != nil {
			return scheduler.NewInvalidValueError(errors.Errorf("failed to parse steering's policy bsid %s, should be a valid ipv6 address: %v", ref.PolicyBsid, err), "policybsid")
		}
	case *srv6.Steering_PolicyIndex:
		if ref.PolicyIndex < 0 {
			return scheduler.NewInvalidValueError(errors.Errorf("policy index can't be negative number"), "policyindex")
		}
	case nil:
		return scheduler.NewInvalidValueError(errors.New("policy reference must be filled, either by policy bsid or policy index"), "PolicyRef(policybsid/policyindex)")
	default:
		return scheduler.NewInvalidValueError(errors.Errorf("policy reference has unexpected model link type %T", ref), "PolicyRef")
	}

	// checking traffic ("oneof" field)
	switch t := steering.Traffic.(type) {
	case *srv6.Steering_L2Traffic_:
		if strings.TrimSpace(t.L2Traffic.InterfaceName) == "" {
			return scheduler.NewInvalidValueError(errors.New("(non-empty) interface name must be given (L2 traffic definition)"), "InterfaceName")
		}
	case *srv6.Steering_L3Traffic_:
		if t.L3Traffic.FibTableId < 0 {
			return scheduler.NewInvalidValueError(errors.Errorf("fibtableid for l3 traffic can't be lower than zero, input value %v", t.L3Traffic.FibTableId), "fibtableid")
		}
		_, _, err := net.ParseCIDR(t.L3Traffic.PrefixAddress)
		if err != nil {
			return scheduler.NewInvalidValueError(errors.Errorf("steering l3 traffic prefix address %v can't be parsed as CIDR formatted prefix: %v", t.L3Traffic.PrefixAddress, err), "prefixaddress")
		}
	case nil:
		return scheduler.NewInvalidValueError(errors.New("steering must have filled traffic otherwise we don't know which traffic to steer into policy"), "traffic")
	default:
		return scheduler.NewInvalidValueError(errors.Errorf("steering's traffic reference has unexpected model link type %T", t), "traffic")
	}
	return nil
}

// Create adds new steering into VPP using VPP's binary api
func (d *SteeringDescriptor) Create(key string, steering *srv6.Steering) (metadata interface{}, err error) {
	if err := d.srHandler.AddSteering(steering); err != nil {
		name, valid := srv6.ModelSteering.ParseKey(key)
		if !valid {
			name = fmt.Sprintf("(key %v)", key)
		}
		return nil, errors.Errorf("failed to add steering %s: %v", name, err)
	}
	return nil, nil
}

// Delete removes steering from VPP using VPP's binary api
func (d *SteeringDescriptor) Delete(key string, steering *srv6.Steering, metadata interface{}) error {
	if err := d.srHandler.RemoveSteering(steering); err != nil {
		name, valid := srv6.ModelSteering.ParseKey(key)
		if !valid {
			name = fmt.Sprintf("(key %v)", key)
		}
		return errors.Errorf("failed to remove steering %s: %v", name, err)
	}
	return nil
}

// Dependencies defines dependencies of Steering descriptor
func (d *SteeringDescriptor) Dependencies(key string, steering *srv6.Steering) (dependencies []scheduler.Dependency) {
	//TODO support also policy identification by index for Dependency waiting (impl using derived value) and do proper robot test for it (vppcalls and NB support it already)
	if _, ok := steering.GetPolicyRef().(*srv6.Steering_PolicyBsid); !ok {
		d.log.Errorf("Non-BSID policy reference in steering is not fully supported (dependency waiting is not "+
			"implemented). Using empty dependencies for steering %+v", steering)
		return dependencies
	}

	dependencies = append(dependencies, scheduler.Dependency{
		Label: policyExistsDep,
		Key:   srv6.PolicyKey(steering.GetPolicyBsid()),
	})
	return dependencies
}
