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

package abfidx

import (
	"go.ligato.io/cn-infra/v2/idxmap"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
	abf "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/abf"
)

// ABFMetadataIndex provides read-only access to mapping between ABF indexes (generated in the ABF plugin)
// and ABF names.
type ABFMetadataIndex interface {
	// LookupByName looks up previously stored item identified by name in the mapping.
	LookupByName(name string) (metadata *ABFMetadata, exists bool)

	// LookupByIndex looks up previously stored item identified by index in the mapping.
	LookupByIndex(idx uint32) (name string, metadata *ABFMetadata, exists bool)
}

// ABFMetadataIndexRW is mapping between ABF indexes (generated in the ABF plugin) and ABF names.
type ABFMetadataIndexRW interface {
	ABFMetadataIndex
	idxmap.NamedMappingRW
}

// ABFMetadata represents metadata for ABF.
type ABFMetadata struct {
	Index    uint32
	Attached []*abf.ABF_AttachedInterface
}

// Attached is helper struct for metadata (ABF attached interface).
type Attached struct {
	Name     string
	IsIPv6   bool
	Priority uint32
}

// GetIndex returns index of the ABF.
func (m *ABFMetadata) GetIndex() uint32 {
	return m.Index
}

// ABFMetadataDto represents an item sent through watch channel in abfIndex.
type ABFMetadataDto struct {
	idxmap.NamedMappingEvent
	Metadata *ABFMetadata
}

type abfMetadataIndex struct {
	idxmap.NamedMappingRW

	log         logging.Logger
	nameToIndex idxvpp.NameToIndex
}

// NewABFIndex creates new instance of abfMetadataIndex.
func NewABFIndex(logger logging.Logger, title string) ABFMetadataIndexRW {
	mapping := idxvpp.NewNameToIndex(logger, title, indexAbfMetadata)
	return &abfMetadataIndex{
		NamedMappingRW: mapping,
		log:            logger,
		nameToIndex:    mapping,
	}
}

// LookupByName looks up previously stored item identified by index in mapping.
func (abfIdx *abfMetadataIndex) LookupByName(name string) (metadata *ABFMetadata, exists bool) {
	meta, found := abfIdx.GetValue(name)
	if found {
		if typedMeta, ok := meta.(*ABFMetadata); ok {
			return typedMeta, found
		}
	}
	return nil, false
}

// LookupByIndex looks up previously stored item identified by name in mapping.
func (abfIdx *abfMetadataIndex) LookupByIndex(idx uint32) (name string, metadata *ABFMetadata, exists bool) {
	var item idxvpp.WithIndex
	name, item, exists = abfIdx.nameToIndex.LookupByIndex(idx)
	if exists {
		var isIfaceMeta bool
		metadata, isIfaceMeta = item.(*ABFMetadata)
		if !isIfaceMeta {
			exists = false
		}
	}
	return
}

// indexMetadata is an index function used for ABF metadata.
func indexAbfMetadata(metaData interface{}) map[string][]string {
	indexes := make(map[string][]string)

	ifMeta, ok := metaData.(*ABFMetadata)
	if !ok || ifMeta == nil {
		return indexes
	}

	return indexes
}
