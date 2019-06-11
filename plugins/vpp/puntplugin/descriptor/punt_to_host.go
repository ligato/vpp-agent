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
	"errors"

	"github.com/ligato/cn-infra/logging"
	punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

const (
	// PuntToHostDescriptorName is the name of the descriptor for the VPP punt to host/socket
	PuntToHostDescriptorName = "vpp-punt-to-host"
)

// A list of non-retriable errors:
var (
	// ErrPuntWithoutL3Protocol is returned when VPP punt has undefined L3 protocol.
	ErrPuntWithoutL3Protocol = errors.New("VPP punt defined without L3 protocol")

	// ErrPuntWithoutL4Protocol is returned when VPP punt has undefined L4 protocol.
	ErrPuntWithoutL4Protocol = errors.New("VPP punt defined without L4 protocol")

	// ErrPuntWithoutPort is returned when VPP punt has undefined port.
	ErrPuntWithoutPort = errors.New("VPP punt defined without port")
)

// PuntToHostDescriptor teaches KVScheduler how to configure VPP punt to host or unix domain socket.
type PuntToHostDescriptor struct {
	RegisterSocketFn func(register bool, toHost *punt.ToHost, socketPath string)

	// dependencies
	log         logging.Logger
	puntHandler vppcalls.PuntVppAPI
}

// NewPuntToHostDescriptor creates a new instance of the punt to host descriptor.
func NewPuntToHostDescriptor(puntHandler vppcalls.PuntVppAPI, log logging.LoggerFactory) *PuntToHostDescriptor {
	return &PuntToHostDescriptor{
		log:         log.NewLogger("punt-to-host-descriptor"),
		puntHandler: puntHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *PuntToHostDescriptor) GetDescriptor() *adapter.PuntToHostDescriptor {
	return &adapter.PuntToHostDescriptor{
		Name:            PuntToHostDescriptorName,
		NBKeyPrefix:     punt.ModelToHost.KeyPrefix(),
		ValueTypeName:   punt.ModelToHost.ProtoName(),
		KeySelector:     punt.ModelToHost.IsKeyValid,
		KeyLabel:        punt.ModelToHost.StripKeyPrefix,
		ValueComparator: d.EquivalentPuntToHost,
		Validate:        d.Validate,
		Create:          d.Create,
		Delete:          d.Delete,
		Retrieve:        d.Retrieve,
	}
}

// EquivalentPuntToHost is case-insensitive comparison function for punt.ToHost.
func (d *PuntToHostDescriptor) EquivalentPuntToHost(key string, oldPunt, newPunt *punt.ToHost) bool {
	if oldPunt.L3Protocol != newPunt.L3Protocol ||
		oldPunt.L4Protocol != newPunt.L4Protocol ||
		oldPunt.Port != newPunt.Port {
		return false
	}
	return true
}

// Validate validates VPP punt configuration.
func (d *PuntToHostDescriptor) Validate(key string, puntCfg *punt.ToHost) error {
	// validate L3 protocol
	switch puntCfg.L3Protocol {
	case punt.L3Protocol_IPv4:
	case punt.L3Protocol_IPv6:
	case punt.L3Protocol_ALL:
	default:
		return kvs.NewInvalidValueError(ErrPuntWithoutL3Protocol, "l3_protocol")
	}

	// validate L4 protocol
	switch puntCfg.L4Protocol {
	case punt.L4Protocol_TCP:
	case punt.L4Protocol_UDP:
	default:
		return kvs.NewInvalidValueError(ErrPuntWithoutL4Protocol, "l4_protocol")
	}

	if puntCfg.Port == 0 {
		return kvs.NewInvalidValueError(ErrPuntWithoutPort, "port")
	}

	return nil
}

// Create adds new punt to host entry or registers new punt to unix domain socket.
func (d *PuntToHostDescriptor) Create(key string, punt *punt.ToHost) (interface{}, error) {
	// add punt to host
	if punt.SocketPath == "" {
		if err := d.puntHandler.AddPunt(punt); err != nil {
			d.log.Error(err)
			return nil, err
		}
		return nil, nil
	}

	// register punt to socket
	pathname, err := d.puntHandler.RegisterPuntSocket(punt)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	if d.RegisterSocketFn != nil {
		d.RegisterSocketFn(true, punt, pathname)
	}

	return nil, nil
}

// Delete removes VPP punt configuration.
func (d *PuntToHostDescriptor) Delete(key string, punt *punt.ToHost, metadata interface{}) error {
	if punt.SocketPath == "" {
		if err := d.puntHandler.DeletePunt(punt); err != nil {
			d.log.Error(err)
			return err
		}
	}

	// deregister punt to socket
	if err := d.puntHandler.DeregisterPuntSocket(punt); err != nil {
		d.log.Error(err)
		return err
	}

	if d.RegisterSocketFn != nil {
		d.RegisterSocketFn(false, punt, "")
	}

	return nil
}

// Retrieve returns all configured VPP punt to host entries.
func (d *PuntToHostDescriptor) Retrieve(correlate []adapter.PuntToHostKVWithMetadata) (retrieved []adapter.PuntToHostKVWithMetadata, err error) {
	// TODO dump for punt and punt socket register missing in api
	d.log.Info("Dump punt/socket register is not supported by the VPP")

	socks, err := d.puntHandler.DumpRegisteredPuntSockets()
	if err != nil {
		return nil, err
	}

	for _, punt := range socks {
		retrieved = append(retrieved, adapter.PuntToHostKVWithMetadata{
			Key:    models.Key(punt.PuntData),
			Value:  punt.PuntData,
			Origin: kvs.FromNB,
		})
	}

	return retrieved, nil
}
