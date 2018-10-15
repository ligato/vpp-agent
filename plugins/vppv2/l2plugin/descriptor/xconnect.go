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
	"strings"

	"github.com/go-errors/errors"
	"github.com/gogo/protobuf/proto"

	"github.com/ligato/cn-infra/logging"
	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"

	"github.com/ligato/vpp-agent/plugins/vppv2/l2plugin/descriptor/adapter"
	vpp_ifdescriptor "github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/l2plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/l2"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/interfaces"
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
		Name:               XConnectDescriptorName,
		KeySelector:        d.IsXConnectKey,
		ValueTypeName:      proto.MessageName(&l2.XConnectPair{}),
		NBKeyPrefix:        l2.XConnectPrefix,
		Add:                d.Add,
		Delete:             d.Delete,
		ModifyWithRecreate: d.ModifyWithRecreate,
		IsRetriableFailure: d.IsRetriableFailure,
		Dependencies:       d.Dependencies,
		Dump:               d.Dump,
		DumpDependencies:   []string{vpp_ifdescriptor.InterfaceDescriptorName},
	}
}

// IsXConnectKey returns true if the key is identifying VPP xConnect configuration.
func (d *XConnectDescriptor) IsXConnectKey(key string) bool {
	return strings.HasPrefix(key, l2.XConnectPrefix)
}

// IsRetriableFailure returns <false> for errors related to invalid configuration.
func (d *XConnectDescriptor) IsRetriableFailure(err error) bool {
	nonRetriable := []error{
		ErrXConnectWithoutInterface,
	}
	for _, nonRetriableErr := range nonRetriable {
		if err == nonRetriableErr {
			return false
		}
	}
	return true
}

// Add adds new xConnect pair.
func (d *XConnectDescriptor) Add(key string, xc *l2.XConnectPair) (metadata interface{}, err error) {
	// validate the configuration first
	err = d.validateXConnectConfig(xc)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

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

// ModifyWithRecreate always returns true - xConnect pairs are modified via re-creation.
func (d *XConnectDescriptor) ModifyWithRecreate(key string, oldXC, newXC *l2.XConnectPair, metadata interface{}) bool {
	return true
}

// Dependencies lists both Rx and Tx interface as dependencies.
func (d *XConnectDescriptor) Dependencies(key string, xc *l2.XConnectPair) []scheduler.Dependency {
	return []scheduler.Dependency{
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

// Dump returns all configured VPP xConnect pairs.
func (d *XConnectDescriptor) Dump(correlate []adapter.XConnectKVWithMetadata) (dump []adapter.XConnectKVWithMetadata, err error) {
	xConnectPairs, err := d.xcHandler.DumpXConnectPairs()
	if err != nil {
		d.log.Error(err)
		return dump, err
	}
	for _, xc := range xConnectPairs {
		dump = append(dump, adapter.XConnectKVWithMetadata{
			Key:      l2.XConnectKey(xc.Xc.ReceiveInterface),
			Value:    xc.Xc,
			Origin:   scheduler.FromNB,
		})
	}

	d.log.Debugf("Dumping xConnect pairs: %v", dump)
	return dump, nil
}

// validateXConnectConfig validates VPP xConnect pair configuration.
func (d *XConnectDescriptor) validateXConnectConfig(xc *l2.XConnectPair) error {
	if xc.ReceiveInterface == "" || xc.TransmitInterface == "" {
		return ErrXConnectWithoutInterface
	}
	return nil
}