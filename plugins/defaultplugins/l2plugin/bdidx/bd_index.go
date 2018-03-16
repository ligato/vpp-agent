// Copyright (c) 2017 Cisco and/or its affiliates.
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

package bdidx

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
)

// BDIndex provides read-only access to mapping between indices (used internally in VPP) and Bridge Domain names.
type BDIndex interface {
	// GetMapping returns internal read-only mapping with metadata of type interface{}.
	GetMapping() idxvpp.NameToIdxRW

	// LookupIdx looks up previously stored item identified by index in mapping.
	LookupIdx(name string) (idx uint32, metadata *l2.BridgeDomains_BridgeDomain, exists bool)

	// LookupName looks up previously stored item identified by name in mapping.
	LookupName(idx uint32) (name string, metadata *l2.BridgeDomains_BridgeDomain, exists bool)

	// LookupBdForInterface looks up for bridge domain the interface belongs to
	LookupBdForInterface(ifName string) (bdIdx uint32, metadata *l2.BridgeDomains_BridgeDomain, bvi bool, exists bool)

	// WatchNameToIdx allows to subscribe for watching changes in bdIndex mapping
	WatchNameToIdx(subscriber core.PluginName, pluginChannel chan ChangeDto)
}

// BDIndexRW is mapping between indices (used internally in VPP) and Bridge Domain names.
type BDIndexRW interface {
	BDIndex

	// RegisterName adds new item into name-to-index mapping.
	RegisterName(name string, idx uint32, metadata *l2.BridgeDomains_BridgeDomain)

	// UnregisterName removes an item identified by name from mapping.
	UnregisterName(name string) (idx uint32, metadata *l2.BridgeDomains_BridgeDomain, exists bool)

	// UpdateMetadata updates metadata in existing bridge domain entry.
	UpdateMetadata(name string, metadata *l2.BridgeDomains_BridgeDomain) (success bool)
}

// bdIndex is type-safe implementation of mapping between Software interface index
// and interface name. It holds as well metadata of type *InterfaceMeta.
type bdIndex struct {
	mapping idxvpp.NameToIdxRW
}

// ChangeDto represents an item sent through watch channel in bdIndex.
// In contrast to NameToIdxDto, it contains typed metadata.
type ChangeDto struct {
	idxvpp.NameToIdxDtoWithoutMeta
	Metadata *l2.BridgeDomains_BridgeDomain
}

const (
	ifaceNameIndexKey = "ipAddrKey" //TODO interfaces in the bridge domain
)

// NewBDIndex creates new instance of bdIndex.
func NewBDIndex(mapping idxvpp.NameToIdxRW) BDIndexRW {
	return &bdIndex{mapping: mapping}
}

// GetMapping returns internal read-only mapping. It is used in tests to inspect the content of the bdIndex.
func (bdi *bdIndex) GetMapping() idxvpp.NameToIdxRW {
	return bdi.mapping
}

// RegisterName adds new item into name-to-index mapping.
func (bdi *bdIndex) RegisterName(name string, idx uint32, ifMeta *l2.BridgeDomains_BridgeDomain) {
	bdi.mapping.RegisterName(name, idx, ifMeta)
}

// IndexMetadata creates indices for metadata. Index for IPAddress will be created.
func IndexMetadata(metaData interface{}) map[string][]string {
	indexes := map[string][]string{}
	ifMeta, ok := metaData.(*l2.BridgeDomains_BridgeDomain)
	if !ok || ifMeta == nil {
		return indexes
	}

	ifacenNames := []string{}
	for _, bdIface := range ifMeta.Interfaces {
		if bdIface != nil {
			ifacenNames = append(ifacenNames, bdIface.Name)
		}
	}
	indexes[ifaceNameIndexKey] = ifacenNames

	return indexes
}

// UnregisterName removes an item identified by name from mapping.
func (bdi *bdIndex) UnregisterName(name string) (idx uint32, metadata *l2.BridgeDomains_BridgeDomain, exists bool) {
	idx, meta, exists := bdi.mapping.UnregisterName(name)
	return idx, bdi.castMetadata(meta), exists
}

// UpdateMetadata updates metadata in existing bridge domain entry.
func (bdi *bdIndex) UpdateMetadata(name string, metadata *l2.BridgeDomains_BridgeDomain) (success bool) {
	return bdi.mapping.UpdateMetadata(name, metadata)
}

// LookupIdx looks up previously stored item identified by index in mapping.
func (bdi *bdIndex) LookupIdx(name string) (idx uint32, metadata *l2.BridgeDomains_BridgeDomain, exists bool) {
	idx, meta, exists := bdi.mapping.LookupIdx(name)
	if exists {
		metadata = bdi.castMetadata(meta)
	}
	return idx, metadata, exists
}

// LookupName looks up previously stored item identified by name in mapping.
func (bdi *bdIndex) LookupName(idx uint32) (name string, metadata *l2.BridgeDomains_BridgeDomain, exists bool) {
	name, meta, exists := bdi.mapping.LookupName(idx)
	if exists {
		metadata = bdi.castMetadata(meta)
	}
	return name, metadata, exists
}

// LookupBdForInterface returns a bridge domain which contains provided interface
func (bdi *bdIndex) LookupBdForInterface(ifName string) (bdIdx uint32, bd *l2.BridgeDomains_BridgeDomain, bvi bool, exists bool) {
	bdNames := bdi.mapping.ListNames()
	for _, bdName := range bdNames {
		bdIdx, meta, exists := bdi.mapping.LookupIdx(bdName)
		if exists && meta != nil {
			bd = bdi.castMetadata(meta)
			if bd != nil {
				for _, iface := range bd.Interfaces {
					if iface.Name == ifName {
						return bdIdx, bd, iface.BridgedVirtualInterface, true
					}
				}
			}
		}
	}

	return bdIdx, nil, bvi, false
}

func (bdi *bdIndex) castMetadata(meta interface{}) *l2.BridgeDomains_BridgeDomain {
	ifMeta, ok := meta.(*l2.BridgeDomains_BridgeDomain)
	if !ok {
		return nil
	}
	return ifMeta
}

// WatchNameToIdx allows to subscribe for watching changes in bdIndex mapping.
func (bdi *bdIndex) WatchNameToIdx(subscriber core.PluginName, pluginChannel chan ChangeDto) {
	ch := make(chan idxvpp.NameToIdxDto)
	bdi.mapping.Watch(subscriber, nametoidx.ToChan(ch))
	go func() {
		for c := range ch {
			pluginChannel <- ChangeDto{
				NameToIdxDtoWithoutMeta: c.NameToIdxDtoWithoutMeta,
				Metadata:                bdi.castMetadata(c.Metadata),
			}

		}
	}()
}
