//  Copyright (c) 2018 Cisco and/or its affiliates.
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
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// IPScanNeighborDescriptorName is the name of the descriptor.
	IPScanNeighborDescriptorName = "vpp-ip-scan-neighbor"
)

// IPScanNeighborDescriptor teaches KVScheduler how to configure VPP proxy ARPs.
type IPScanNeighborDescriptor struct {
	log                   logging.Logger
	ipNeigh               vppcalls.IPNeighVppAPI
	scheduler             kvs.KVScheduler
	defaultIPScanNeighbor *l3.IPScanNeighbor
}

// NewIPScanNeighborDescriptor creates a new instance of the IPScanNeighborDescriptor.
func NewIPScanNeighborDescriptor(
	scheduler kvs.KVScheduler,
	ipNeighHandler vppcalls.IPNeighVppAPI,
	log logging.PluginLogger,
) *kvs.KVDescriptor {

	ctx := &IPScanNeighborDescriptor{
		scheduler:             scheduler,
		ipNeigh:               ipNeighHandler,
		log:                   log.NewLogger("ip-scan-neigh-descriptor"),
		defaultIPScanNeighbor: ipNeighHandler.DefaultIPScanNeighbor(),
	}

	typedDescr := &adapter.IPScanNeighborDescriptor{
		Name:            IPScanNeighborDescriptorName,
		NBKeyPrefix:     l3.ModelIPScanNeighbor.KeyPrefix(),
		ValueTypeName:   l3.ModelIPScanNeighbor.ProtoName(),
		KeySelector:     l3.ModelIPScanNeighbor.IsKeyValid,
		ValueComparator: ctx.EquivalentIPScanNeighbors,
		Create:          ctx.Create,
		Update:          ctx.Update,
		Delete:          ctx.Delete,
		Retrieve:        ctx.Retrieve,
		// TODO: define validation method
	}
	return adapter.NewIPScanNeighborDescriptor(typedDescr)
}

// EquivalentIPScanNeighbors compares the IP Scan Neighbor values.
func (d *IPScanNeighborDescriptor) EquivalentIPScanNeighbors(key string, oldValue, newValue *l3.IPScanNeighbor) bool {
	return proto.Equal(d.withDefaults(oldValue), d.withDefaults(newValue))
}

// Create adds VPP IP Scan Neighbor.
func (d *IPScanNeighborDescriptor) Create(key string, value *l3.IPScanNeighbor) (metadata interface{}, err error) {
	return d.Update(key, d.defaultIPScanNeighbor, value, nil)
}

// Delete deletes VPP IP Scan Neighbor.
func (d *IPScanNeighborDescriptor) Delete(key string, value *l3.IPScanNeighbor, metadata interface{}) error {
	_, err := d.Update(key, value, d.defaultIPScanNeighbor, metadata)
	if errors.Is(err, vppcalls.ErrIPNeighborNotImplemented) {
		d.log.Debug("SetIPScanNeighbor failed:", err)
		return nil
	}
	return err
}

// Update modifies VPP IP Scan Neighbor.
func (d *IPScanNeighborDescriptor) Update(key string, oldValue, newValue *l3.IPScanNeighbor, oldMetadata interface{}) (newMetadata interface{}, err error) {
	if err := d.ipNeigh.SetIPScanNeighbor(newValue); err != nil {
		return nil, err
	}
	return nil, nil
}

// Retrieve returns current VPP IP Scan Neighbor configuration.
func (d *IPScanNeighborDescriptor) Retrieve(correlate []adapter.IPScanNeighborKVWithMetadata) (
	retrieved []adapter.IPScanNeighborKVWithMetadata, err error,
) {
	// Retrieve VPP configuration
	ipNeigh, err := d.ipNeigh.GetIPScanNeighbor()
	if errors.Is(err, vppcalls.ErrIPNeighborNotImplemented) {
		d.log.Debug("GetIPScanNeighbor failed:", err)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	d.fillDefaults(ipNeigh)

	var origin = kvs.FromNB
	if proto.Equal(ipNeigh, d.defaultIPScanNeighbor) {
		origin = kvs.FromSB
	}

	retrieved = append(retrieved, adapter.IPScanNeighborKVWithMetadata{
		Key:    models.Key(ipNeigh),
		Value:  ipNeigh,
		Origin: origin,
	})

	return retrieved, nil
}
func (d *IPScanNeighborDescriptor) withDefaults(orig *l3.IPScanNeighbor) *l3.IPScanNeighbor {
	var (
		def = d.defaultIPScanNeighbor
		val = *orig
	)
	if def == nil {
		return &val
	}
	if val.ScanInterval == 0 {
		val.ScanInterval = def.GetScanInterval()
	}
	if val.MaxProcTime == 0 {
		val.MaxProcTime = def.GetMaxProcTime()
	}
	if val.MaxUpdate == 0 {
		val.MaxUpdate = def.GetMaxUpdate()
	}
	if val.ScanIntDelay == 0 {
		val.ScanIntDelay = def.GetScanIntDelay()
	}
	if val.StaleThreshold == 0 {
		val.StaleThreshold = def.GetStaleThreshold()
	}
	return &val
}

func (d *IPScanNeighborDescriptor) fillDefaults(orig *l3.IPScanNeighbor) {
	var (
		def = d.defaultIPScanNeighbor
		val = orig
	)
	if def == nil {
		return
	}
	if val.ScanInterval == 0 {
		val.ScanInterval = def.GetScanInterval()
	}
	if val.MaxProcTime == 0 {
		val.MaxProcTime = def.GetMaxProcTime()
	}
	if val.MaxUpdate == 0 {
		val.MaxUpdate = def.GetMaxUpdate()
	}
	if val.ScanIntDelay == 0 {
		val.ScanIntDelay = def.GetScanIntDelay()
	}
	if val.StaleThreshold == 0 {
		val.StaleThreshold = def.GetStaleThreshold()
	}
}
