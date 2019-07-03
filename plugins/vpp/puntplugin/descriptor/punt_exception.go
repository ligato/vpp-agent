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
	"strings"

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
	RegisterSocketFn func(register bool, toHost *punt.Exception, socketPath string)

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
	if oldPunt.Reason != newPunt.Reason {
		return false
	}

	// if the socket path contains '!' as prefix we return false
	// to force scheduler to recreate (register) punt socket
	if strings.HasPrefix(oldPunt.SocketPath, "!") {
		return false
	}

	return true
}

// Validate validates VPP punt configuration.
func (d *PuntExceptionDescriptor) Validate(key string, puntCfg *punt.Exception) error {
	// validate reason
	if puntCfg.GetReason() == "" {
		return ErrPuntExceptionWithoutReason
	}

	if puntCfg.SocketPath == "" {
		return kvs.NewInvalidValueError(ErrPuntWithoutSocketPath, "socket_path")
	}

	return nil
}

// Create adds new punt to host entry or registers new punt to unix domain socket.
func (d *PuntExceptionDescriptor) Create(key string, punt *punt.Exception) (interface{}, error) {
	// register punt exception
	pathname, err := d.puntHandler.AddPuntException(punt)
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
func (d *PuntExceptionDescriptor) Delete(key string, punt *punt.Exception, metadata interface{}) error {
	// check if the socketpath contains '!' as prefix from retrieve
	p := punt
	if strings.HasPrefix(p.SocketPath, "!") {
		p = &(*punt)
		p.SocketPath = strings.TrimPrefix(p.SocketPath, "!")
	}

	// delete punt exception
	err := d.puntHandler.DeletePuntException(punt)
	if err != nil {
		d.log.Error(err)
		return err
	}

	if d.RegisterSocketFn != nil {
		d.RegisterSocketFn(false, punt, "")
	}

	return nil
}

// Retrieve returns all configured VPP punt exception entries.
func (d *PuntExceptionDescriptor) Retrieve(correlate []adapter.PuntExceptionKVWithMetadata) (retrieved []adapter.PuntExceptionKVWithMetadata, err error) {
	// Dump punt exceptions
	punts, err := d.puntHandler.DumpExceptions()
	if err == vppcalls.ErrUnsupported {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// for all dumped punts that were not yet registered and for which
	// the VPP socket is unknown we prepend '!' as prefix
	// to allow descriptor to recognize this in equivalent
	// and force recreation or make it possible to delete it
	for _, p := range punts {
		if p.Exception.SocketPath == "" && p.SocketPath != "" {
			p.Exception.SocketPath = "!" + p.SocketPath
		}
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
