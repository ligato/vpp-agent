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
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
)

// BDIndex provides read-only access to mapping between indexes (used internally in VPP) and Bridge Domain names.
type BDIndex interface {
	// GetMapping returns internal read-only mapping with metadata of type interface{}.
	GetMapping() idxvpp.NameToIdxRW

	// LookupIdx looks up previously stored item identified by index in mapping.
	LookupIdx(name string) (idx uint32, metadata *l2.BridgeDomains_BridgeDomain, exists bool)

	// LookupName looks up previously stored item identified by name in mapping.
	LookupName(idx uint32) (name string, metadata *l2.BridgeDomains_BridgeDomain, exists bool)

	// LookupNameByIP returns names of items that contains given IP address in metadata
	LookupNameByIfaceName(ifaceName string) []string

	// WatchNameToIdx allows to subscribe for watching changes in bdIndex mapping
	WatchNameToIdx(subscriber core.PluginName, pluginChannel chan ChangeDto)
}

// BDIndexRW is mapping between indexes (used internally in VPP) and Bridge Domain names.
type BDIndexRW interface {
	BDIndex

	// RegisterName adds new item into name-to-index mapping.
	RegisterName(name string, idx uint32, ifMeta *l2.BridgeDomains_BridgeDomain)

	// UnregisterName removes an item identified by name from mapping
	UnregisterName(name string) (idx uint32, metadata *l2.BridgeDomains_BridgeDomain, exists bool)
}

// bdIndex is type-safe implementation of mapping between Software interface index
// and interface name. It holds as well metadata of type *InterfaceMeta.
type bdIndex struct {
	mapping idxvpp.NameToIdxRW
}

// ChangeDto represents an item sent through watch channel in bdIndex.
// In contrast to NameToIdxDto it contains typed metadata.
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
func (swi *bdIndex) GetMapping() idxvpp.NameToIdxRW {
	return swi.mapping
}

// RegisterName adds new item into name-to-index mapping.
func (swi *bdIndex) RegisterName(name string, idx uint32, ifMeta *l2.BridgeDomains_BridgeDomain) {
	swi.mapping.RegisterName(name, idx, ifMeta)
}

// IndexMetadata creates indexes for metadata. Index for IPAddress will be created
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

// UnregisterName removes an item identified by name from mapping
func (swi *bdIndex) UnregisterName(name string) (idx uint32, metadata *l2.BridgeDomains_BridgeDomain, exists bool) {
	idx, meta, exists := swi.mapping.UnregisterName(name)
	return idx, swi.castMetadata(meta), exists
}

// LookupIdx looks up previously stored item identified by index in mapping.
func (swi *bdIndex) LookupIdx(name string) (idx uint32, metadata *l2.BridgeDomains_BridgeDomain, exists bool) {
	idx, meta, exists := swi.mapping.LookupIdx(name)
	if exists {
		metadata = swi.castMetadata(meta)
	}
	return idx, metadata, exists
}

// LookupName looks up previously stored item identified by name in mapping.
func (swi *bdIndex) LookupName(idx uint32) (name string, metadata *l2.BridgeDomains_BridgeDomain, exists bool) {
	name, meta, exists := swi.mapping.LookupName(idx)
	if exists {
		metadata = swi.castMetadata(meta)
	}
	return name, metadata, exists
}

// LookupNameByIP returns names of items that contains given IP address in metadata
func (swi *bdIndex) LookupNameByIfaceName(ifaceName string) []string {
	return swi.mapping.LookupNameByMetadata(ifaceNameIndexKey, ifaceName)
}

func (swi *bdIndex) castMetadata(meta interface{}) *l2.BridgeDomains_BridgeDomain {
	ifMeta, ok := meta.(*l2.BridgeDomains_BridgeDomain)
	if !ok {
		return nil
	}
	return ifMeta
}

// WatchNameToIdx allows to subscribe for watching changes in bdIndex mapping
func (swi *bdIndex) WatchNameToIdx(subscriber core.PluginName, pluginChannel chan ChangeDto) {
	ch := make(chan idxvpp.NameToIdxDto)
	swi.mapping.Watch(subscriber, nametoidx.ToChan(ch))
	go func() {
		for c := range ch {
			pluginChannel <- ChangeDto{
				NameToIdxDtoWithoutMeta: c.NameToIdxDtoWithoutMeta,
				Metadata:                swi.castMetadata(c.Metadata),
			}

		}
	}()
}
