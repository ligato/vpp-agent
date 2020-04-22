// Copyright (c) 2020 Cisco and/or its affiliates.
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
	"errors"

	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ipfixplugin/vppcalls"
	ipfix "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/ipfix"
)

const (
	// IPFIXDescriptorName is the name of the descriptor for
	// VPP IP Flow Information eXport (IPFIX) configuration.
	IPFIXDescriptorName = "vpp-ipfix"
)

// Validation errors:
var (
	// ErrColAddrNotDefined returned when collector address in confiugration is empty string.
	ErrColAddrNotDefined = errors.New("address of a collector was not provided")
	// ErrSrcAddrNotDefined returned when source address in confiugration is empty string.
	ErrSrcAddrNotDefined = errors.New("address of a source was not provided")
	// ErrTooBigMTU informs about the maximum value for Path MTU.
	ErrTooBigMTU = errors.New("too big value, maximum is 1450")
	// ErrTooSmlMTU informs about the minimum value for Path MTU.
	ErrTooSmlMTU = errors.New("too small value, minimum is 68")
)

// IPFIXDescriptor configures IPFIX for VPP.
type IPFIXDescriptor struct {
	ipfixHandler vppcalls.IpfixVppAPI
	log          logging.Logger
}

// NewIPFIXDescriptor creates a new instance of IPFIXDescriptor.
func NewIPFIXDescriptor(ipfixHandler vppcalls.IpfixVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {
	ctx := &IPFIXDescriptor{
		ipfixHandler: ipfixHandler,
		log:          log.NewLogger("ipfix-descriptor"),
	}
	typedDescr := &adapter.IPFIXDescriptor{
		Name:            IPFIXDescriptorName,
		NBKeyPrefix:     ipfix.ModelIPFIX.KeyPrefix(),
		ValueTypeName:   ipfix.ModelIPFIX.ProtoName(),
		KeySelector:     ipfix.ModelIPFIX.IsKeyValid,
		KeyLabel:        ipfix.ModelIPFIX.StripKeyPrefix,
		ValueComparator: ctx.EquivalentIPFIX,
		Validate:        ctx.Validate,
		Create:          ctx.Create,
		Delete:          ctx.Delete,
		Retrieve:        ctx.Retrieve,
		Update:          ctx.Update,
	}
	return adapter.NewIPFIXDescriptor(typedDescr)
}

// EquivalentIPFIX returns true if two IPFIX configurations are equal.
func (d *IPFIXDescriptor) EquivalentIPFIX(key string, oldValue, newValue *ipfix.IPFIX) bool {
	if oldValue.GetCollector().GetAddress() != newValue.GetCollector().GetAddress() {
		return false
	}

	if oldValue.GetSourceAddress() != newValue.GetSourceAddress() {
		return false
	}

	oldPort := oldValue.GetCollector().GetPort()
	newPort := newValue.GetCollector().GetPort()
	if oldPort != newPort {
		defaultPort := uint32(4739)
		oldIsNotDefault := oldPort != 0 && oldPort != defaultPort
		newIsNotDefault := newPort != 0 && newPort != defaultPort

		if oldIsNotDefault || newIsNotDefault {
			return false
		}
	}

	if oldValue.GetVrfId() != newValue.GetVrfId() {
		return false
	}

	oldMTU := oldValue.GetPathMtu()
	newMTU := newValue.GetPathMtu()
	if oldMTU != newMTU {
		defaultMTU := uint32(512)
		oldIsNotDefault := oldMTU != 0 && oldMTU != defaultMTU
		newIsNotDefault := newMTU != 0 && newMTU != defaultMTU

		if oldIsNotDefault || newIsNotDefault {
			return false
		}
	}

	oldInterval := oldValue.GetTemplateInterval()
	newInterval := newValue.GetTemplateInterval()
	if oldInterval != newInterval {
		defaultInterval := uint32(20)
		oldIsNotDefault := oldInterval != 0 && oldInterval != defaultInterval
		newIsNotDefault := newInterval != 0 && newInterval != defaultInterval

		if oldIsNotDefault || newIsNotDefault {
			return false
		}
	}

	return true
}

// Validate does basic check of VPP IPFIX configuration.
func (d *IPFIXDescriptor) Validate(key string, value *ipfix.IPFIX) error {
	if value.GetCollector().GetAddress() == "" {
		return kvs.NewInvalidValueError(ErrColAddrNotDefined, "collector.address")
	}

	if value.GetSourceAddress() == "" {
		return kvs.NewInvalidValueError(ErrSrcAddrNotDefined, "source_address")
	}

	if mtu := value.GetPathMtu(); mtu == 0 {
		// That's okay. No worries. VPP will use the default Path MTU value.
	} else if mtu > vppcalls.MaxPathMTU {
		return kvs.NewInvalidValueError(ErrTooBigMTU, "path_mtu")
	} else if mtu < vppcalls.MinPathMTU {
		return kvs.NewInvalidValueError(ErrTooSmlMTU, "path_mtu")
	}

	return nil
}

// Create calls Update method, because IPFIX configuration is always there and can not be created.
func (d *IPFIXDescriptor) Create(key string, val *ipfix.IPFIX) (metadata interface{}, err error) {
	_, err = d.Update(key, nil, val, nil)
	return
}

// Update sets VPP IPFIX configuration.
func (d *IPFIXDescriptor) Update(key string, oldVal, newVal *ipfix.IPFIX, oldMetadata interface{}) (newMetadata interface{}, err error) {
	err = d.ipfixHandler.SetExporter(newVal)
	return
}

// Delete does nothing, because there are neither ability
// nor reasons to delete VPP IPFIX configuration.
// You can only configure exporting in a way you want to.
func (d *IPFIXDescriptor) Delete(key string, val *ipfix.IPFIX, metadata interface{}) (err error) {
	return nil
}

// Retrieve returns configuration of IP Flow Infromation eXporter.
func (d *IPFIXDescriptor) Retrieve(correlate []adapter.IPFIXKVWithMetadata) (retrieved []adapter.IPFIXKVWithMetadata, err error) {
	ipfixes, err := d.ipfixHandler.DumpExporters()
	if err != nil {
		return nil, err
	}

	for _, e := range ipfixes {
		retrieved = append(retrieved, adapter.IPFIXKVWithMetadata{
			Key:    models.Key(e),
			Value:  e,
			Origin: kvs.FromSB,
		})
	}

	return retrieved, nil
}
