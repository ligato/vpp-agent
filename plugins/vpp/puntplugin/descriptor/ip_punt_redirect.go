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

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"

	interfaces "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

const (
	// IPRedirectDescriptorName is the name of the descriptor for the VPP punt to host/socket
	IPRedirectDescriptorName = "vpp-punt-ipredirect"

	// dependency labels
	ipRedirectTxInterfaceDep = "tx-interface-exists"
	ipRedirectRxInterfaceDep = "rx-interface-exists"
)

// A list of non-retriable errors:
var (
	// ErrIPRedirectWithoutL3Protocol is returned when VPP IP redirect has undefined L3 protocol.
	ErrIPRedirectWithoutL3Protocol = errors.New("VPP IP punt redirect defined without L3 protocol")

	// ErrPuntWithoutL4Protocol is returned when VPP IP redirect has undefined L4 tx interface.
	ErrIPRedirectWithoutTxInterface = errors.New("VPP IP punt redirect defined without tx interface")

	// ErrIPRedirectWithoutNextHop is returned when VPP IP redirect has undefined next hop address.
	ErrIPRedirectWithoutNextHop = errors.New("VPP IP punt redirect defined without tx interface")
)

// IPRedirectDescriptor teaches KVScheduler how to configure VPP IP punt redirect.
type IPRedirectDescriptor struct {
	// dependencies
	log         logging.Logger
	puntHandler vppcalls.PuntVppAPI
}

// NewIPRedirectDescriptor creates a new instance of the punt to host descriptor.
func NewIPRedirectDescriptor(puntHandler vppcalls.PuntVppAPI, log logging.LoggerFactory) *IPRedirectDescriptor {
	return &IPRedirectDescriptor{
		log:         log.NewLogger("punt-ipredirect-descriptor"),
		puntHandler: puntHandler,
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *IPRedirectDescriptor) GetDescriptor() *adapter.IPPuntRedirectDescriptor {
	return &adapter.IPPuntRedirectDescriptor{
		Name:            IPRedirectDescriptorName,
		NBKeyPrefix:     punt.ModelIPRedirect.KeyPrefix(),
		ValueTypeName:   punt.ModelIPRedirect.ProtoName(),
		KeySelector:     punt.ModelIPRedirect.IsKeyValid,
		KeyLabel:        punt.ModelIPRedirect.StripKeyPrefix,
		ValueComparator: d.EquivalentIPRedirect,
		Validate:        d.Validate,
		Create:          d.Create,
		Delete:          d.Delete,
		Retrieve:        d.Retrieve,
		Dependencies:    d.Dependencies,
	}
}

// EquivalentIPRedirect is case-insensitive comparison function for punt.IpRedirect.
func (d *IPRedirectDescriptor) EquivalentIPRedirect(key string, oldIPRedirect, newIPRedirect *punt.IPRedirect) bool {
	// parameters compared by proto equal
	return proto.Equal(oldIPRedirect, newIPRedirect)
}

// Validate validates VPP punt configuration.
func (d *IPRedirectDescriptor) Validate(key string, redirect *punt.IPRedirect) error {
	// validate L3 protocol
	switch redirect.L3Protocol {
	case punt.L3Protocol_IPv4:
	case punt.L3Protocol_IPv6:
	case punt.L3Protocol_ALL:
	default:
		return kvs.NewInvalidValueError(ErrIPRedirectWithoutL3Protocol, "l3_protocol")
	}

	// validate tx interface
	if redirect.TxInterface == "" {
		return kvs.NewInvalidValueError(ErrIPRedirectWithoutTxInterface, "tx_interface")
	}

	// validate next hop
	if redirect.NextHop == "" {
		return kvs.NewInvalidValueError(ErrIPRedirectWithoutNextHop, "next_hop")
	}

	return nil
}

// Create adds new IP punt redirect entry.
func (d *IPRedirectDescriptor) Create(key string, redirect *punt.IPRedirect) (metadata interface{}, err error) {
	// add Punt to host/socket
	err = d.puntHandler.AddPuntRedirect(redirect)
	if err != nil {
		d.log.Error(err)
	}
	return nil, err
}

// Delete removes VPP IP punt redirect configuration.
func (d *IPRedirectDescriptor) Delete(key string, redirect *punt.IPRedirect, metadata interface{}) error {
	err := d.puntHandler.DeletePuntRedirect(redirect)
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// Retrieve returns all configured VPP punt to host entries.
func (d *IPRedirectDescriptor) Retrieve(correlate []adapter.IPPuntRedirectKVWithMetadata) (dump []adapter.IPPuntRedirectKVWithMetadata, err error) {
	punts, err := d.puntHandler.DumpPuntRedirect()
	if err == vppcalls.ErrUnsupported {
		return nil, nil
	} else if err != nil {
		d.log.Error(err)
		return nil, err
	}

	for _, p := range punts {
		dump = append(dump, adapter.IPPuntRedirectKVWithMetadata{
			Key:    models.Key(p),
			Value:  p,
			Origin: kvs.FromNB,
		})
	}

	return dump, nil
}

// Dependencies for IP punt redirect are represented by tx interface
func (d *IPRedirectDescriptor) Dependencies(key string, redirect *punt.IPRedirect) (dependencies []kvs.Dependency) {
	// TX interface
	dependencies = append(dependencies, kvs.Dependency{
		Label: ipRedirectTxInterfaceDep,
		Key:   interfaces.InterfaceKey(redirect.TxInterface),
	})
	// RX interface
	if redirect.RxInterface != "" {
		dependencies = append(dependencies, kvs.Dependency{
			Label: ipRedirectRxInterfaceDep,
			Key:   interfaces.InterfaceKey(redirect.RxInterface),
		})
	}
	return dependencies
}
