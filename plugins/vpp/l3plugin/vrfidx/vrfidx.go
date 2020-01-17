// Copyright (c) 2019 Cisco and/or its affiliates.
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

package vrfidx

import (
	"time"

	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"

	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"

	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
)

// VRFMetadataIndex provides read-only access to mapping with VPP VRF
// metadata. It extends from NameToIndex.
type VRFMetadataIndex interface {
	// LookupByName retrieves a previously stored metadata of VRF
	// identified by <label>. If there is no VRF associated with the give
	// label in the mapping, the <exists> is returned as *false* and <metadata>
	// as *nil*.
	LookupByName(name string) (metadata *VRFMetadata, exists bool)

	// LookupByVRFIndex retrieves a previously stored VRF identified in
	// VPP by the given <index>.
	// If there is no VRF associated with the given index, <exists> is returned
	// as *false* with <name> and <metadata> both set to empty values.
	LookupByVRFIndex(index uint32) (name string, metadata *VRFMetadata, exists bool)

	// ListAllVRFs returns slice of labels of all VRFs in the mapping.
	ListAllVRFs() (names []string)

	// ListAllVrfMetadata returns a list of VRF metadata - ID/Proto pairs as
	// read from VRF metadata.
	ListAllVrfMetadata() (idList []*VRFMetadata)

	// WatchVRFs allows to subscribe to watch for changes in the VRF mapping.
	WatchVRFs(subscriber string, channel chan<- VRFMetadataDto)
}

// VRFMetadataIndexRW provides read-write access to mapping with VRF
// metadata.
type VRFMetadataIndexRW interface {
	VRFMetadataIndex
	idxmap.NamedMappingRW
}

// VRFMetadata collects metadata for VPP VRF used in secondary lookups.
type VRFMetadata struct {
	Index    uint32
	Protocol l3.VrfTable_Protocol
}

// GetIndex returns VRF index.
func (vrfm *VRFMetadata) GetIndex() uint32 {
	return vrfm.Index
}

// GetProtocol returns VRF IP protocol.
func (vrfm *VRFMetadata) GetProtocol() l3.VrfTable_Protocol {
	return vrfm.Protocol
}

// VRFMetadataDto represents an item sent through watch channel in VRFMetadataIndex.
// In contrast to NamedMappingGenericEvent, it contains typed VRF metadata.
type VRFMetadataDto struct {
	idxmap.NamedMappingEvent
	Metadata *VRFMetadata
}

// vrfMetadataIndex is type-safe implementation of mapping between VRF
// label and metadata of type *VRFMeta.
type vrfMetadataIndex struct {
	idxmap.NamedMappingRW /* embeds */

	log         logging.Logger
	nameToIndex idxvpp.NameToIndex /* contains */
}

// NewVRFIndex creates a new instance implementing VRFMetadataIndexRW.
func NewVRFIndex(logger logging.Logger, title string) VRFMetadataIndexRW {
	mapping := idxvpp.NewNameToIndex(logger, title, indexMetadata)
	return &vrfMetadataIndex{
		NamedMappingRW: mapping,
		log:            logger,
		nameToIndex:    mapping,
	}
}

// LookupByName retrieves a previously stored metadata of VRF
// identified by <label>. If there is no VRF associated with the given
// name in the mapping, the <exists> is returned as *false* and <metadata>
// as *nil*.
func (m *vrfMetadataIndex) LookupByName(name string) (metadata *VRFMetadata, exists bool) {
	meta, found := m.GetValue(name)
	if found {
		if typedMeta, ok := meta.(*VRFMetadata); ok {
			return typedMeta, found
		}
	}
	return nil, false
}

// LookupByVRFIndex retrieves a previously stored VRF identified in
// VPP by the given/ <index>.
// If there is no VRF associated with the given index, <exists> is returned
// as *false* with <name> and <metadata> both set to empty values.
func (m *vrfMetadataIndex) LookupByVRFIndex(swIfIndex uint32) (name string, metadata *VRFMetadata, exists bool) {
	var item idxvpp.WithIndex
	name, item, exists = m.nameToIndex.LookupByIndex(swIfIndex)
	if exists {
		var isVrfMeta bool
		metadata, isVrfMeta = item.(*VRFMetadata)
		if !isVrfMeta {
			exists = false
		}
	}
	return
}

// ListAllVRFs returns slice of labels of all VRFs in the mapping.
func (m *vrfMetadataIndex) ListAllVRFs() (names []string) {
	return m.ListAllNames()
}

// ListAllVrfMetadata returns a list of VRF metadata - ID/Proto pairs as
// read from VRF metadata.
func (m *vrfMetadataIndex) ListAllVrfMetadata() (metaList []*VRFMetadata) {
	for _, vrf := range m.ListAllNames() {
		vrfMeta, ok := m.LookupByName(vrf)
		if vrfMeta == nil || !ok {
			continue
		}
		metaList = append(metaList, vrfMeta)
	}
	return
}

// WatchVRFs allows to subscribe to watch for changes in the VRF mapping.
func (m *vrfMetadataIndex) WatchVRFs(subscriber string, channel chan<- VRFMetadataDto) {
	watcher := func(dto idxmap.NamedMappingGenericEvent) {
		typedMeta, ok := dto.Value.(*VRFMetadata)
		if !ok {
			return
		}
		msg := VRFMetadataDto{
			NamedMappingEvent: dto.NamedMappingEvent,
			Metadata:          typedMeta,
		}
		timeout := idxmap.DefaultNotifTimeout
		select {
		case channel <- msg:
			// OK
		case <-time.After(timeout):
			m.log.Warnf("Unable to deliver VRF watch notification after %v, channel is full", timeout)
		}
	}
	if err := m.Watch(subscriber, watcher); err != nil {
		m.log.Error(err)
	}
}

// indexMetadata is an index function used for VRF metadata.
func indexMetadata(metaData interface{}) map[string][]string {
	indexes := make(map[string][]string)

	ifMeta, ok := metaData.(*VRFMetadata)
	if !ok || ifMeta == nil {
		return indexes
	}

	return indexes
}
