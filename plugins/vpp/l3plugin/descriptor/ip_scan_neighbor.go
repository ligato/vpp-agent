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
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v2/pkg/models"
	kvs "go.ligato.io/vpp-agent/v2/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/l3"
)

const (
	// IPScanNeighborDescriptorName is the name of the descriptor.
	IPScanNeighborDescriptorName = "vpp-ip-scan-neighbor"
)

const (
	defaultScanInterval   = 1
	defaultMaxProcTime    = 20
	defaultMaxUpdate      = 10
	defaultScanIntDelay   = 1
	defaultStaleThreshold = 4
)

var defaultIPScanNeighbor = &l3.IPScanNeighbor{
	Mode:           l3.IPScanNeighbor_DISABLED,
	ScanInterval:   defaultScanInterval,
	MaxProcTime:    defaultMaxProcTime,
	MaxUpdate:      defaultMaxUpdate,
	ScanIntDelay:   defaultScanIntDelay,
	StaleThreshold: defaultStaleThreshold,
}

// IPScanNeighborDescriptor teaches KVScheduler how to configure VPP proxy ARPs.
type IPScanNeighborDescriptor struct {
	log       logging.Logger
	ipNeigh   vppcalls.IPNeighVppAPI
	scheduler kvs.KVScheduler
}

// NewIPScanNeighborDescriptor creates a new instance of the IPScanNeighborDescriptor.
func NewIPScanNeighborDescriptor(scheduler kvs.KVScheduler,
	proxyArpHandler vppcalls.IPNeighVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &IPScanNeighborDescriptor{
		scheduler: scheduler,
		ipNeigh:   proxyArpHandler,
		log:       log.NewLogger("ip-scan-neigh-descriptor"),
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
	}
	return adapter.NewIPScanNeighborDescriptor(typedDescr)
}

// EquivalentIPScanNeighbors compares the IP Scan Neighbor values.
func (d *IPScanNeighborDescriptor) EquivalentIPScanNeighbors(key string, oldValue, newValue *l3.IPScanNeighbor) bool {
	return proto.Equal(withDefaults(oldValue), withDefaults(newValue))
}

// Create adds VPP IP Scan Neighbor.
func (d *IPScanNeighborDescriptor) Create(key string, value *l3.IPScanNeighbor) (metadata interface{}, err error) {
	return d.Update(key, defaultIPScanNeighbor, value, nil)
}

// Delete deletes VPP IP Scan Neighbor.
func (d *IPScanNeighborDescriptor) Delete(key string, value *l3.IPScanNeighbor, metadata interface{}) error {
	_, err := d.Update(key, value, defaultIPScanNeighbor, metadata)
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
	if err != nil {
		return nil, err
	}
	fillDefaults(ipNeigh)

	var origin = kvs.FromNB
	if proto.Equal(ipNeigh, defaultIPScanNeighbor) {
		origin = kvs.FromSB
	}

	retrieved = append(retrieved, adapter.IPScanNeighborKVWithMetadata{
		Key:    models.Key(ipNeigh),
		Value:  ipNeigh,
		Origin: origin,
	})

	return retrieved, nil
}
func withDefaults(orig *l3.IPScanNeighbor) *l3.IPScanNeighbor {
	var val = *orig
	if val.ScanInterval == 0 {
		val.ScanInterval = defaultScanInterval
	}
	if val.MaxProcTime == 0 {
		val.MaxProcTime = defaultMaxProcTime
	}
	if val.MaxUpdate == 0 {
		val.MaxUpdate = defaultMaxUpdate
	}
	if val.ScanIntDelay == 0 {
		val.ScanIntDelay = defaultScanIntDelay
	}
	if val.StaleThreshold == 0 {
		val.StaleThreshold = defaultStaleThreshold
	}
	return &val
}

func fillDefaults(orig *l3.IPScanNeighbor) {
	var val = orig
	if val.ScanInterval == 0 {
		val.ScanInterval = defaultScanInterval
	}
	if val.MaxProcTime == 0 {
		val.MaxProcTime = defaultMaxProcTime
	}
	if val.MaxUpdate == 0 {
		val.MaxUpdate = defaultMaxUpdate
	}
	if val.ScanIntDelay == 0 {
		val.ScanIntDelay = defaultScanIntDelay
	}
	if val.StaleThreshold == 0 {
		val.StaleThreshold = defaultStaleThreshold
	}
}
