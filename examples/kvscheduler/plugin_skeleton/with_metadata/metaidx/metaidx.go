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

package metaidx

import (
	"time"

	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
)

// SkeletonMetadataIndex provides read-only access to mapping between skeleton values
// and their metadata.
type SkeletonMetadataIndex interface {
	// LookupName looks up previously stored item identified by name in the mapping.
	LookupByName(name string) (metadata *SkeletonMetadata, exists bool)

	// TODO: define additional secondary lookups here ...

	// WatchSkeletonMetadata allows to watch for changes in the mapping.
	WatchSkeletonMetadata(subscriber string, channel chan<- SkeletonMetadataDto)
}

// SkeletonMetadataIndexRW is a mapping between skeleton values and their metadata.
type SkeletonMetadataIndexRW interface {
	SkeletonMetadataIndex
	idxmap.NamedMappingRW
}

// SkeletonMetadata represents metadata for skeleton value.
type SkeletonMetadata struct {
	// TODO: put metadata attributes here ...
}

// SkeletonMetadataDto represents an item sent through a watch channel.
type SkeletonMetadataDto struct {
	idxmap.NamedMappingEvent
	Metadata *SkeletonMetadata
}

// skeletonMetadataIndex is an implementation of SkeletonMetadataIndexRW.
type skeletonMetadataIndex struct {
	idxmap.NamedMappingRW
	log logging.Logger
}

// NewSkeletonIndex creates new instance of skeletonMetadataIndex.
func NewSkeletonIndex(logger logging.Logger, title string) SkeletonMetadataIndexRW {
	mapping := idxvpp.NewNameToIndex(logger, title, indexMetadata)
	return &skeletonMetadataIndex{
		NamedMappingRW: mapping,
		log:            logger,
	}
}

// LookupName looks up previously stored item identified by name in mapping.
func (idx *skeletonMetadataIndex) LookupByName(name string) (metadata *SkeletonMetadata, exists bool) {
	meta, found := idx.GetValue(name)
	if found {
		if typedMeta, ok := meta.(*SkeletonMetadata); ok {
			return typedMeta, found
		}
	}
	return nil, false
}

// TODO: add secondary lookup functions here...

// WatchSkeletonMetadata allows to watch for changes in the mapping.
func (idx *skeletonMetadataIndex) WatchSkeletonMetadata(subscriber string, channel chan<- SkeletonMetadataDto) {
	watcher := func(dto idxmap.NamedMappingGenericEvent) {
		typedMeta, ok := dto.Value.(*SkeletonMetadata)
		if !ok {
			return
		}
		msg := SkeletonMetadataDto{
			NamedMappingEvent: dto.NamedMappingEvent,
			Metadata:          typedMeta,
		}
		select {
		case channel <- msg:
		case <-time.After(idxmap.DefaultNotifTimeout):
			idx.log.Warn("Unable to deliver notification")
		}
	}
	if err := idx.Watch(subscriber, watcher); err != nil {
		idx.log.Error(err)
	}
}

// indexMetadata is an index function used for skeleton metadata.
func indexMetadata(metadata interface{}) map[string][]string {
	indexes := make(map[string][]string)

	// TODO: define secondary indexes here ...

	return indexes
}
