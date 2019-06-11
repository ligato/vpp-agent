//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package descriptor

import (
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"

	punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

const (
	// PuntExceptionDescriptorName is the name of the descriptor for the VPP punt exception
	PuntExceptionDescriptorName = "vpp-punt-exception"
)

// A list of non-retriable errors:
var (
	// ErrPuntExceptionWithoutReason is returned when VPP punt exception has undefined reason.
	ErrPuntExceptionWithoutReason = errors.New("VPP punt exception defined without reason")
)

// PuntExceptionDescriptor teaches KVScheduler how to configure VPP putn exception.
type PuntExceptionDescriptor struct {
	// dependencies
	log         logging.Logger
	puntHandler vppcalls.PuntVppAPI
}

// NewPuntExceptionDescriptor creates a new instance of the punt exception.
func NewPuntExceptionDescriptor(puntHandler vppcalls.PuntVppAPI, log logging.LoggerFactory) *PuntExceptionDescriptor {
	return &PuntExceptionDescriptor{
		log:         log.NewLogger("punt-exception-descriptor"),
		puntHandler: puntHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *PuntExceptionDescriptor) GetDescriptor() *adapter.PuntExceptionDescriptor {
	return &adapter.PuntExceptionDescriptor{
		Name:            PuntExceptionDescriptorName,
		NBKeyPrefix:     punt.ModelException.KeyPrefix(),
		ValueTypeName:   punt.ModelException.ProtoName(),
		KeySelector:     punt.ModelException.IsKeyValid,
		KeyLabel:        punt.ModelException.StripKeyPrefix,
		ValueComparator: d.EquivalentPuntException,
		Validate:        d.Validate,
		Create:          d.Create,
		Delete:          d.Delete,
		Retrieve:        d.Retrieve,
	}
}

// EquivalentPuntException is case-insensitive comparison function for punt.Exception.
func (d *PuntExceptionDescriptor) EquivalentPuntException(key string, oldPunt, newPunt *punt.Exception) bool {
	return proto.Equal(oldPunt, newPunt)
}

// Validate validates VPP punt configuration.
func (d *PuntExceptionDescriptor) Validate(key string, puntCfg *punt.Exception) error {
	// validate reason
	if puntCfg.GetReason() == "" {
		return ErrPuntExceptionWithoutReason
	}

	return nil
}

// Create adds new punt to host entry or registers new punt to unix domain socket.
func (d *PuntExceptionDescriptor) Create(key string, punt *punt.Exception) (interface{}, error) {
	// add punt exception
	err := d.puntHandler.AddPuntException(punt)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	return nil, nil
}

// Delete removes VPP punt configuration.
func (d *PuntExceptionDescriptor) Delete(key string, punt *punt.Exception, metadata interface{}) error {
	// delete punt exception
	err := d.puntHandler.DeletePuntException(punt)
	if err != nil {
		d.log.Error(err)
		return err
	}

	return nil
}

// Retrieve returns all configured VPP punt exception entries.
func (d *PuntExceptionDescriptor) Retrieve(correlate []adapter.PuntExceptionKVWithMetadata) (retrieved []adapter.PuntExceptionKVWithMetadata, err error) {
	// Dump punt exceptions
	punts, err := d.puntHandler.DumpExceptions()
	if err != nil {
		return nil, err
	}

	for _, p := range punts {
		retrieved = append(retrieved, adapter.PuntExceptionKVWithMetadata{
			Key:    models.Key(p.Exception),
			Value:  p.Exception,
			Origin: kvs.FromNB,
		})
	}

	return retrieved, nil
}
