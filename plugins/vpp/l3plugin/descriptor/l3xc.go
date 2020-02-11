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
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// L3XCDescriptorName is the name of the descriptor.
	L3XCDescriptorName = "vpp-l3xc"

	// dependency labels
	l3xcTargetInterfaceDep = "target-interface-exists"
	l3xcPathInterfaceDep   = "outgoing-interface-exists"
)

// L3XCDescriptor teaches KVScheduler how to configure VPP L3XCs.
type L3XCDescriptor struct {
	log         logging.Logger
	l3xcHandler vppcalls.L3XCVppAPI
	ifIndexes   ifaceidx.IfaceMetadataIndex
}

// NewL3XCDescriptor creates a new instance of the L3XCDescriptor.
func NewL3XCDescriptor(l3xcHandler vppcalls.L3XCVppAPI, ifIndexes ifaceidx.IfaceMetadataIndex,
	log logging.PluginLogger,
) *kvs.KVDescriptor {

	ctx := &L3XCDescriptor{
		ifIndexes:   ifIndexes,
		l3xcHandler: l3xcHandler,
		log:         log.NewLogger("l3xc-descriptor"),
	}

	typedDescr := &adapter.L3XCDescriptor{
		Name:                 L3XCDescriptorName,
		NBKeyPrefix:          l3.ModelL3XC.KeyPrefix(),
		ValueTypeName:        l3.ModelL3XC.ProtoName(),
		KeySelector:          l3.ModelL3XC.IsKeyValid,
		KeyLabel:             l3.ModelL3XC.StripKeyPrefix,
		ValueComparator:      ctx.EquivalentL3XCs,
		Validate:             ctx.Validate,
		Create:               ctx.Create,
		Update:               ctx.Update,
		Delete:               ctx.Delete,
		Retrieve:             ctx.Retrieve,
		Dependencies:         ctx.Dependencies,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewL3XCDescriptor(typedDescr)
}

// EquivalentL3XCs is comparison function for L3XC entries.
func (d *L3XCDescriptor) EquivalentL3XCs(key string, oldL3XC, newL3XC *l3.L3XConnect) bool {
	return proto.Equal(oldL3XC, newL3XC)
}

// Validate returns if given l3xc is valid.
func (d *L3XCDescriptor) Validate(key string, l3xc *l3.L3XConnect) error {
	if l3xc.Interface == "" {
		return errors.Errorf("no interface defined")
	}
	if len(l3xc.Paths) == 0 {
		return errors.Errorf("no paths defined")
	}
	return nil
}

// Dependencies lists dependencies for a VPP L3XC entry.
func (d *L3XCDescriptor) Dependencies(key string, l3xc *l3.L3XConnect) (deps []kvs.Dependency) {
	// the outgoing interface must exist
	if l3xc.Interface != "" {
		deps = append(deps, kvs.Dependency{
			Label: l3xcTargetInterfaceDep,
			Key:   interfaces.InterfaceKey(l3xc.Interface),
		})
	}
	for _, path := range l3xc.Paths {
		deps = append(deps, kvs.Dependency{
			Label: l3xcPathInterfaceDep,
			Key:   interfaces.InterfaceKey(path.OutgoingInterface),
		})
	}
	return deps
}

// Create adds VPP L3XC entry.
func (d *L3XCDescriptor) Create(key string, l3xc *l3.L3XConnect) (interface{}, error) {
	return d.update(key, l3xc)
}

// Update updates VPP L3XC entry.
func (d *L3XCDescriptor) Update(key string, oldL3XC, newL3XC *l3.L3XConnect, oldMeta interface{}) (interface{}, error) {
	return d.update(key, newL3XC)
}

func (d *L3XCDescriptor) update(key string, l3xc *l3.L3XConnect) (interface{}, error) {
	ctx := context.TODO()

	var swIfIndex uint32
	if strings.HasPrefix(l3xc.Interface, "MISSING-") {
		idx := strings.TrimPrefix(l3xc.Interface, "MISSING-")
		x, _ := strconv.ParseUint(idx, 10, 32)
		swIfIndex = uint32(x)
	} else {
		meta, found := d.ifIndexes.LookupByName(l3xc.Interface)
		if !found {
			return nil, errors.Errorf("interface %s not found", l3xc.Interface)
		}
		swIfIndex = meta.SwIfIndex
	}

	paths := make([]vppcalls.Path, len(l3xc.Paths))
	for i, p := range l3xc.Paths {
		pmeta, found := d.ifIndexes.LookupByName(p.OutgoingInterface)
		if !found {
			return nil, errors.Errorf("interface %s from path #%d not found", p.OutgoingInterface, i)
		}
		paths[i] = vppcalls.Path{
			SwIfIndex:  pmeta.SwIfIndex,
			Weight:     uint8(p.Weight),
			Preference: uint8(p.Preference),
			NextHop:    net.ParseIP(p.NextHopAddr),
		}
	}

	if err := d.l3xcHandler.UpdateL3XC(ctx, &vppcalls.L3XC{
		SwIfIndex: swIfIndex,
		IsIPv6:    l3xc.Protocol == l3.L3XConnect_IPV6,
		Paths:     paths,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// Delete removes VPP L3XC entry.
func (d *L3XCDescriptor) Delete(key string, l3xc *l3.L3XConnect, metadata interface{}) error {
	ctx := context.TODO()

	var swIfIndex uint32
	if strings.HasPrefix(l3xc.Interface, "MISSING-") {
		idx := strings.TrimPrefix(l3xc.Interface, "MISSING-")
		x, _ := strconv.ParseUint(idx, 10, 32)
		swIfIndex = uint32(x)
	} else {
		meta, found := d.ifIndexes.LookupByName(l3xc.Interface)
		if !found {
			return errors.Errorf("interface %s not found", l3xc.Interface)
		}
		swIfIndex = meta.SwIfIndex
	}
	isIPv6 := l3xc.Protocol == l3.L3XConnect_IPV6

	if err := d.l3xcHandler.DeleteL3XC(ctx, swIfIndex, isIPv6); err != nil {
		return err
	}

	return nil
}

// Retrieve returns all L3XC entries associated with interfaces managed by this agent.
func (d *L3XCDescriptor) Retrieve(correlate []adapter.L3XCKVWithMetadata) (
	retrieved []adapter.L3XCKVWithMetadata, err error,
) {
	ctx := context.TODO()

	l3xcEntries, err := d.l3xcHandler.DumpAllL3XC(ctx)
	if err != nil {
		return nil, errors.Errorf("dumping VPP L3XCs failed: %v", err)
	}

	for _, l3xc := range l3xcEntries {
		ifName, _, exists := d.ifIndexes.LookupBySwIfIndex(l3xc.SwIfIndex)
		if !exists {
			ifName = fmt.Sprintf("MISSING-%d", l3xc.SwIfIndex)
			d.log.Warnf("L3XC dump: interface index %d not found", l3xc.SwIfIndex)

		}
		ipProto := l3.L3XConnect_IPV4
		if l3xc.IsIPv6 {
			ipProto = l3.L3XConnect_IPV6
		}
		paths := make([]*l3.L3XConnect_Path, len(l3xc.Paths))
		for i, p := range l3xc.Paths {
			ifNamePath, _, exists := d.ifIndexes.LookupBySwIfIndex(p.SwIfIndex)
			if !exists {
				ifNamePath = fmt.Sprintf("MISSING-%d", p.SwIfIndex)
				d.log.Warnf("L3XC dump: interface index %d for path #%d not found", p.SwIfIndex, i)
			}
			paths[i] = &l3.L3XConnect_Path{
				OutgoingInterface: ifNamePath,
				NextHopAddr:       p.NextHop.String(),
				Weight:            uint32(p.Weight),
				Preference:        uint32(p.Preference),
			}
		}
		value := &l3.L3XConnect{
			Interface: ifName,
			Protocol:  ipProto,
			Paths:     paths,
		}
		retrieved = append(retrieved, adapter.L3XCKVWithMetadata{
			Key:    models.Key(value),
			Value:  value,
			Origin: kvs.FromNB,
		})
	}

	return retrieved, nil
}
