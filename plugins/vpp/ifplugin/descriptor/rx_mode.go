// Copyright (c) 2018 Cisco and/or its affiliates.
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

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/pkg/errors"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/vppcalls"
)

const (
	// RxModeDescriptorName is the name of the descriptor for the unnumbered
	// config-subsection of VPP interfaces.
	RxModeDescriptorName = "vpp-interface-rx-mode"

	// dependency labels
	linkIsUpDep = "interface-link-is-UP"
)

// A list of non-retriable errors:
var (
	// ErrUnsupportedRxMode is returned when the given interface type does not support the chosen
	// RX mode.
	ErrUnsupportedRxMode = errors.New("unsupported RX Mode")

	// ErrUndefinedRxMode is returned when the Rx mode is not defined.
	ErrUndefinedRxMode = errors.New("undefined RX Mode")

	// ErrUnsupportedRxMode is returned when Rx mode has multiple definitions for the same queue.
	ErrRedefinedRxMode = errors.New("redefined RX Mode")
)

// RxModeDescriptor configures Rx mode for VPP interface queues.
type RxModeDescriptor struct {
	log       logging.Logger
	ifHandler vppcalls.InterfaceVppAPI
	ifIndex   ifaceidx.IfaceMetadataIndex
}

// NewRxModeDescriptor creates a new instance of RxModeDescriptor.
func NewRxModeDescriptor(ifHandler vppcalls.InterfaceVppAPI, ifIndex ifaceidx.IfaceMetadataIndex,
	log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &RxModeDescriptor{
		ifHandler: ifHandler,
		ifIndex:   ifIndex,
		log:       log.NewLogger("rx-mode-descriptor"),
	}

	typedDescr := &adapter.RxModeDescriptor{
		Name:            RxModeDescriptorName,
		KeySelector:     ctx.IsInterfaceRxModeKey,
		// proto message Interface is only used as container for RxMode
		ValueTypeName: proto.MessageName(&interfaces.Interface{}),
		ValueComparator: ctx.EquivalentRxMode,
		Validate:      ctx.Validate,
		Create:        ctx.Create,
		Update:        ctx.Update,
		Delete:        ctx.Delete,
		Dependencies:  ctx.Dependencies,
	}

	return adapter.NewRxModeDescriptor(typedDescr)
}

// IsInterfaceRxModeKey returns true if the key is identifying RxMode configuration.
func (d *RxModeDescriptor) IsInterfaceRxModeKey(key string) bool {
	_, isValid := interfaces.ParseRxModesKey(key)
	return isValid
}

// EquivalentRxMode compares Rx modes for equivalency.
func (d *RxModeDescriptor) EquivalentRxMode(key string, oldIntf, newIntf *interfaces.Interface) bool {
	/* Note: default Rx mode cannot be dumped - compare only if these are two NB
	   configurations (not a refreshed, i.e. dumped, value).
	*/
	oldDefMode := getDefaultRxMode(oldIntf)
	newDefMode := getDefaultRxMode(newIntf)
	if oldDefMode != interfaces.Interface_RxMode_UNKNOWN &&
		newDefMode != interfaces.Interface_RxMode_UNKNOWN {
		if oldDefMode != newDefMode {
			return false
		}
	}
	// compare queue-specific RX modes
	for _, rxMode := range oldIntf.GetRxModes() {
		if rxMode.DefaultMode {
			continue
		}
		oldMode := normalizeRxMode(rxMode.Mode, oldIntf)
		newMode := getQueueRxMode(rxMode.Queue, newIntf)
		if oldMode != newMode {
			return false
		}
	}
	for _, rxMode := range newIntf.GetRxModes() {
		if rxMode.DefaultMode {
			continue
		}
		newMode := normalizeRxMode(rxMode.Mode, newIntf)
		oldMode := getQueueRxMode(rxMode.Queue, oldIntf)
		if oldMode != newMode {
			return false
		}
	}
	return true
}

// Validate validates Rx mode configuration.
func (d *RxModeDescriptor) Validate(key string, ifaceWithRxMode *interfaces.Interface) error {
	for i, rxMode1 := range ifaceWithRxMode.GetRxModes() {
		if rxMode1.Mode == interfaces.Interface_RxMode_UNKNOWN {
			if rxMode1.DefaultMode {
				return kvs.NewInvalidValueError(ErrUndefinedRxMode,"rx_mode[default]")
			}
			return kvs.NewInvalidValueError(ErrUndefinedRxMode,
				fmt.Sprintf("rx_mode[.queue=%d]", rxMode1.Queue))
		}
		for j := i + 1; j < len(ifaceWithRxMode.GetRxModes()); j++ {
			rxMode2 := ifaceWithRxMode.GetRxModes()[j]
			if rxMode1.DefaultMode != rxMode2.DefaultMode {
				continue
			}
			if rxMode1.DefaultMode {
				return kvs.NewInvalidValueError(ErrRedefinedRxMode,"rx_mode[default]")
			}
			if rxMode1.Queue == rxMode2.Queue {
				return kvs.NewInvalidValueError(ErrRedefinedRxMode,
					fmt.Sprintf("rx_mode[.queue=%d]", rxMode1.Queue))
			}
		}
	}

	if ifaceWithRxMode.GetType() == interfaces.Interface_DPDK {
		for _, rxMode := range ifaceWithRxMode.GetRxModes() {
			mode := normalizeRxMode(rxMode.Mode, ifaceWithRxMode)
			if mode != interfaces.Interface_RxMode_POLLING {
				if rxMode.DefaultMode {
					return kvs.NewInvalidValueError(ErrUnsupportedRxMode,
						"rx_mode[default]")
				}
				return kvs.NewInvalidValueError(ErrUnsupportedRxMode,
					fmt.Sprintf("rx_mode[.queue=%d]", rxMode.Queue))
			}
		}
	}
	return nil
}

// Create configures RxMode for a given interface.
// Please note the proto message Interface is only used as container for RxMode.
// Only interface name, type and Rx mode are set.
func (d *RxModeDescriptor) Create(key string, ifaceWithRxMode *interfaces.Interface) (metadata interface{}, err error) {
	err = d.configureRxMode(ifaceWithRxMode, kvs.TxnOperation_CREATE)
	return nil, err
}

// Update modifies Rx mode configuration.
func (d *RxModeDescriptor) Update(key string, _, ifaceWithRxMode *interfaces.Interface,
	oldMetadata interface{}) (newMetadata interface{}, err error) {

	err = d.configureRxMode(ifaceWithRxMode, kvs.TxnOperation_UPDATE)
	return nil, err
}

// Delete reverts back to the default rx mode configuration.
func (d *RxModeDescriptor) Delete(key string, ifaceWithRxMode *interfaces.Interface, metadata interface{}) error {
	return d.configureRxMode(ifaceWithRxMode, kvs.TxnOperation_DELETE)
}

// configureRxMode (re-)configures Rx mode for the interface.
func (d *RxModeDescriptor) configureRxMode(iface *interfaces.Interface, op kvs.TxnOperation) (err error) {

	ifMeta, found := d.ifIndex.LookupByName(iface.Name)
	if !found {
		err = errors.Errorf("failed to find interface %s", iface.Name)
		d.log.Error(err)
		return err
	}
	ifIdx := ifMeta.SwIfIndex

	defRxMode := getDefaultRxMode(iface)

	// first, revert back to default for all queues
	revertToDefault := op == kvs.TxnOperation_DELETE ||
		(op == kvs.TxnOperation_UPDATE && defRxMode == interfaces.Interface_RxMode_UNKNOWN)
	if revertToDefault {
		err = d.ifHandler.SetRxMode(ifIdx, &interfaces.Interface_RxMode{
			DefaultMode: true,
			Mode:        normalizeRxMode(interfaces.Interface_RxMode_DEFAULT, iface),
		})
		if err != nil {
			// treat error as warning here
			d.log.Warnf("failed to un-configure Rx-mode (%v) - most likely "+
				"the interface is already without a link", err)
			err = nil
		}
	}

	if op == kvs.TxnOperation_DELETE {
		return
	}

	// configure the requested default Rx mode
	if defRxMode != interfaces.Interface_RxMode_UNKNOWN {
		err = d.ifHandler.SetRxMode(ifIdx, &interfaces.Interface_RxMode{
			DefaultMode: true,
			Mode:        defRxMode,
		})
		if err != nil {
			err = errors.Errorf("failed to set default Rx-mode for interface %s: %v", iface.Name, err)
			d.log.Error(err)
			return err
		}
	}

	// configure per-queue RX mode
	for _, rxMode := range iface.GetRxModes() {
		if rxMode.DefaultMode || rxMode.Mode == defRxMode {
			continue
		}
		err = d.ifHandler.SetRxMode(ifIdx, rxMode)
		if err != nil {
			err = errors.Errorf("failed to set Rx-mode for queue %d of the interface %s: %v",
				rxMode.Queue, iface.Name, err)
			d.log.Error(err)
			return err
		}
	}

	return nil
}

// Dependencies informs scheduler that Rx mode configuration cannot be applied
// until the interface link is UP.
func (d *RxModeDescriptor) Dependencies(key string, ifaceWithRxMode *interfaces.Interface) (deps []kvs.Dependency) {
	return []kvs.Dependency{
		{
			Label: linkIsUpDep,
			Key:   interfaces.LinkStateKey(ifaceWithRxMode.Name, true),
		},
	}
}

// getDefaultRxMode reads default RX mode from the interface configuration.
func getDefaultRxMode(iface *interfaces.Interface) (rxMode interfaces.Interface_RxMode_Type) {
	for _, rxMode := range iface.GetRxModes() {
		if rxMode.DefaultMode {
			return normalizeRxMode(rxMode.Mode, iface)
		}
	}
	return interfaces.Interface_RxMode_UNKNOWN
}

// getQueueRxMode reads RX mode for the given queue from the interface configuration.
func getQueueRxMode(queue uint32, iface *interfaces.Interface) (mode interfaces.Interface_RxMode_Type) {
	for _, rxMode := range iface.GetRxModes() {
		if rxMode.DefaultMode {
			mode = rxMode.Mode
			continue // keep looking for a queue-specific RX mode
		}
		if rxMode.Queue == queue {
			mode = rxMode.Mode
			break
		}
	}
	return normalizeRxMode(mode, iface)
}

// normalizeRxMode resolves default/undefined Rx mode for specific interfaces.
func normalizeRxMode(mode interfaces.Interface_RxMode_Type, iface *interfaces.Interface) interfaces.Interface_RxMode_Type {
	if mode != interfaces.Interface_RxMode_DEFAULT {
		return mode
	}
	switch iface.GetType() {
	case interfaces.Interface_DPDK:
		return interfaces.Interface_RxMode_POLLING
	case interfaces.Interface_AF_PACKET:
		return interfaces.Interface_RxMode_INTERRUPT
	case interfaces.Interface_TAP:
		if iface.GetTap().GetVersion() == 2 {
			// TAP v2
			return interfaces.Interface_RxMode_INTERRUPT
		}
	}
	return mode
}
