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

package puntidx

import (
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/vpp/model/punt"
)

// PuntIndex provides read-only access to mapping between punt registrations and punt names
type PuntIndex interface {
	// GetMapping returns internal read-only mapping with metadata of type interface{}.
	GetMapping() idxvpp.NameToIdxRW

	// LookupIdx looks up previously stored item identified by index in mapping.
	LookupIdx(name string) (idx uint32, metadata *PuntMetadata, exists bool)

	// LookupName looks up previously stored item identified by name in mapping.
	LookupName(idx uint32) (name string, metadata *PuntMetadata, exists bool)
}

// PuntIndexRW is mapping between software punt indexes and names.
type PuntIndexRW interface {
	PuntIndex

	// RegisterName adds new item into name-to-index mapping.
	RegisterName(name string, idx uint32, puntMeta *PuntMetadata)

	// UnregisterName removes an item identified by name from mapping
	UnregisterName(name string) (idx uint32, metadata *PuntMetadata, exists bool)

	// UpdateMetadata updates metadata in existing punt entry.
	UpdateMetadata(name string, metadata *PuntMetadata) (success bool)

	// Clear removes all punt entries from the mapping.
	Clear()
}

// PuntIdx is type-safe implementation of mapping between punt indexes and names.
type PuntIdx struct {
	mapping idxvpp.NameToIdxRW
}

// PuntMetadata are custom metadata of the punt configuration
type PuntMetadata struct {
	Punt       *punt.Punt
	SocketPath []byte
}

// NewPuntIndex creates new instance of PuntIndexRW.
func NewPuntIndex(mapping idxvpp.NameToIdxRW) PuntIndexRW {
	return &PuntIdx{mapping: mapping}
}

// GetMapping returns internal read-only mapping. It is used in tests to inspect the content of the PuntIdx.
func (p *PuntIdx) GetMapping() idxvpp.NameToIdxRW {
	return p.mapping
}

// LookupIdx looks up previously stored item identified by index in mapping.
func (p *PuntIdx) LookupIdx(name string) (idx uint32, metadata *PuntMetadata, exists bool) {
	idx, meta, exists := p.mapping.LookupIdx(name)
	if exists {
		metadata = p.castMetadata(meta)
	}
	return idx, metadata, exists
}

// LookupName looks up previously stored item identified by name in mapping.
func (p *PuntIdx) LookupName(idx uint32) (name string, metadata *PuntMetadata, exists bool) {
	name, meta, exists := p.mapping.LookupName(idx)
	if exists {
		metadata = p.castMetadata(meta)
	}
	return name, metadata, exists
}

// RegisterName adds new item into name-to-index mapping.
func (p *PuntIdx) RegisterName(name string, idx uint32, ifMeta *PuntMetadata) {
	p.mapping.RegisterName(name, idx, ifMeta)
}

// UnregisterName removes an item identified by name from mapping
func (p *PuntIdx) UnregisterName(name string) (idx uint32, metadata *PuntMetadata, exists bool) {
	idx, meta, exists := p.mapping.UnregisterName(name)
	return idx, p.castMetadata(meta), exists
}

// UpdateMetadata updates metadata in existing punt entry.
func (p *PuntIdx) UpdateMetadata(name string, metadata *PuntMetadata) (success bool) {
	return p.mapping.UpdateMetadata(name, metadata)
}

// Clear removes all punt entries from the cache.
func (p *PuntIdx) Clear() {
	p.mapping.Clear()
}

func (p *PuntIdx) castMetadata(meta interface{}) *PuntMetadata {
	if puntMeta, ok := meta.(*PuntMetadata); ok {
		return puntMeta
	}

	return nil
}
