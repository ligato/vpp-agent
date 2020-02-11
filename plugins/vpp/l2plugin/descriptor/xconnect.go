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
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	vpp_ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l2plugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l2 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l2"
)

const (
	// XConnectDescriptorName is the name of the descriptor for VPP xConnect pairs.
	XConnectDescriptorName = "vpp-xconnect"

	// dependency labels
	rxInterfaceDep = "rx-interface"
	txInterfaceDep = "tx-interface"
)

// A list of non-retriable errors:
var (
	// ErrXConnectWithoutInterface is returned when VPP xConnect has undefined
	// Rx or Tx interface.
	ErrXConnectWithoutInterface = errors.New("VPP xConnect defined without Rx/Tx interface")
)

// XConnectDescriptor teaches KVScheduler how to configure VPP xConnect pairs.
type XConnectDescriptor struct {
	// dependencies
	log       logging.Logger
	xcHandler vppcalls.XConnectVppAPI
}

// NewXConnectDescriptor creates a new instance of the xConnect descriptor.
func NewXConnectDescriptor(xcHandler vppcalls.XConnectVppAPI, log logging.PluginLogger) *XConnectDescriptor {

	return &XConnectDescriptor{
		xcHandler: xcHandler,
		log:       log.NewLogger("xconnect-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *XConnectDescriptor) GetDescriptor() *adapter.XConnectDescriptor {
	return &adapter.XConnectDescriptor{
		Name:                 XConnectDescriptorName,
		NBKeyPrefix:          l2.ModelXConnectPair.KeyPrefix(),
		ValueTypeName:        l2.ModelXConnectPair.ProtoName(),
		KeySelector:          l2.ModelXConnectPair.IsKeyValid,
		KeyLabel:             l2.ModelXConnectPair.StripKeyPrefix,
		Validate:             d.Validate,
		Create:               d.Create,
		Delete:               d.Delete,
		Retrieve:             d.Retrieve,
		Dependencies:         d.Dependencies,
		RetrieveDependencies: []string{vpp_ifdescriptor.InterfaceDescriptorName},
	}
}

// Validate validates VPP xConnect pair configuration.
func (d *XConnectDescriptor) Validate(key string, xc *l2.XConnectPair) error {
	if xc.ReceiveInterface == "" || xc.TransmitInterface == "" {
		return kvs.NewInvalidValueError(ErrXConnectWithoutInterface,
			"receive_interface", "transmit_interface")
	}
	return nil
}

// Create adds new xConnect pair.
func (d *XConnectDescriptor) Create(key string, xc *l2.XConnectPair) (metadata interface{}, err error) {
	// add xConnect pair
	err = d.xcHandler.AddL2XConnect(xc.ReceiveInterface, xc.TransmitInterface)
	if err != nil {
		d.log.Error(err)
	}
	return nil, err
}

// Delete removes existing xConnect pair.
func (d *XConnectDescriptor) Delete(key string, xc *l2.XConnectPair, metadata interface{}) error {
	err := d.xcHandler.DeleteL2XConnect(xc.ReceiveInterface, xc.TransmitInterface)
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// Retrieve returns all configured VPP xConnect pairs.
func (d *XConnectDescriptor) Retrieve(correlate []adapter.XConnectKVWithMetadata) (retrieved []adapter.XConnectKVWithMetadata, err error) {
	xConnectPairs, err := d.xcHandler.DumpXConnectPairs()
	if err != nil {
		d.log.Error(err)
		return retrieved, err
	}

	for _, xc := range xConnectPairs {
		retrieved = append(retrieved, adapter.XConnectKVWithMetadata{
			Key:    l2.XConnectKey(xc.Xc.ReceiveInterface),
			Value:  xc.Xc,
			Origin: kvs.FromNB,
		})
	}
	return retrieved, nil
}

// Dependencies lists both Rx and Tx interface as dependencies.
func (d *XConnectDescriptor) Dependencies(key string, xc *l2.XConnectPair) []kvs.Dependency {
	return []kvs.Dependency{
		{
			Label: rxInterfaceDep,
			Key:   interfaces.InterfaceKey(xc.ReceiveInterface),
		},
		{
			Label: txInterfaceDep,
			Key:   interfaces.InterfaceKey(xc.TransmitInterface),
		},
	}
}
