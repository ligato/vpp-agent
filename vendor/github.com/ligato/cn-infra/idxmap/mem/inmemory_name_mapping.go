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

package mem

import (
	"fmt"
	"sync"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
)

// mappingItem represents single item stored in mapping.
type mappingItem struct {
	// name identifies item in the mapping (primary index).
	name string
	// stored data
	value interface{}
	// indexed contains fields extracted from value (secondary indexes).
	// Extracted field can be used as lookup criteria.
	indexed map[string][]string
}

// memNamedMapping is an in-memory implementation of idxmap.NamedMappingRW.
type memNamedMapping struct {
	logging.Logger
	access    sync.RWMutex
	nameToIdx map[string]*mappingItem
	// createIndexes is function that computes secondary indexes for a given item.
	createIndexes func(interface{}) map[string][]string
	// indexes is a register of secondary indexes
	indexes map[string] /* index name */ map[string] /* index value */ *nameSet
	// subscribers to whom notifications are delivered
	subscribers map[core.PluginName]func(idxmap.NamedMappingGenericEvent)
	owner       core.PluginName
	title       string
}

// NewNamedMapping creates a new instance of the in-memory implementation
// of idxmap.NamedMappingRW
// An index function that builds secondary indexes for an item can be defined
// and passed as <indexFunction>.
func NewNamedMapping(logger logging.Logger, owner core.PluginName, title string,
	indexFunction func(interface{}) map[string][]string) idxmap.NamedMappingRW {
	mem := memNamedMapping{}
	mem.Logger = logger
	mem.nameToIdx = map[string]*mappingItem{}
	mem.indexes = map[string]map[string]*nameSet{}
	mem.createIndexes = indexFunction
	mem.subscribers = map[core.PluginName]func(idxmap.NamedMappingGenericEvent){}
	mem.owner = owner
	mem.title = title
	return &mem
}

// Put adds an item to the mapping associated with the <name>.
// If there is an already stored item with that name, it gets overwritten.
func (mem *memNamedMapping) Put(name string, value interface{}) {
	mem.access.Lock()
	defer mem.access.Unlock()

	mem.putNameToIdx(name, value)

	mem.publishToChannel(name, value)

}

// Delete removes an item associated with the given <name> from the mapping.
func (mem *memNamedMapping) Delete(name string) (value interface{}, found bool) {
	mem.access.Lock()
	defer mem.access.Unlock()

	item, found := mem.nameToIdx[name]
	if found {
		value = item.value

		mem.publishDelToChannel(name, item.value) //TODO improve and not send nil
		mem.removeNameIdx(name)
	}

	return value, found
}

// GetRegistryTitle returns the title assigned to the registry.
func (mem *memNamedMapping) GetRegistryTitle() string {
	return mem.title
}

// GetValue looks up an item in the mapping by <name> (primary index).
func (mem *memNamedMapping) GetValue(name string) (value interface{}, exists bool) {
	mem.access.RLock()
	defer mem.access.RUnlock()

	item, found := mem.nameToIdx[name]
	if found {
		return item.value, found
	}
	return
}

// ListAllNames returns all names in the mapping.
func (mem *memNamedMapping) ListAllNames() (names []string) {
	mem.access.RLock()
	defer mem.access.RUnlock()

	var ret []string

	for name := range mem.nameToIdx {
		ret = append(ret, name)
	}

	return ret
}

// ListNames looks up the items by secondary indexes. It returns all
// names matching the selection.
func (mem *memNamedMapping) ListNames(field string, value string) []string {
	mem.access.RLock()
	defer mem.access.RUnlock()

	ix, found := mem.indexes[field]
	if !found {
		return nil
	}
	set, found := ix[value]

	if !found {
		return nil
	}

	return set.content()
}

// Watch allows to subscribe for tracking changes in the mapping.
// When an item is added or removed, the given <callback> is triggered.
func (mem *memNamedMapping) Watch(subscriber core.PluginName, callback func(idxmap.NamedMappingGenericEvent)) error {
	mem.Debug("Watch ", subscriber)
	mem.access.Lock()
	defer mem.access.Unlock()

	_, found := mem.subscribers[subscriber]
	if !found {
		mem.subscribers[subscriber] = callback
	} else {
		return fmt.Errorf("Already registered channel per subscriber ")
	}
	return nil
}

func (mem *memNamedMapping) updateIndexes(item *mappingItem, name string) {
	if mem.createIndexes == nil {
		return
	}
	mem.removeIndexes(item, name)

	item.indexed = mem.createIndexes(item.value)
	for key, vals := range item.indexed {
		ix, keyExists := mem.indexes[key]
		if !keyExists {
			ix = map[string]*nameSet{}
			mem.indexes[key] = ix
		}
		for _, v := range vals {
			set, found := ix[v]
			if !found {
				set = newIndexSet()
				ix[v] = set
			}
			set.add(name)
		}
	}

}

func (mem *memNamedMapping) removeIndexes(item *mappingItem, name string) {
	for key, vals := range item.indexed {
		ix, found := mem.indexes[key]
		if !found {
			continue
		}
		for _, v := range vals {
			set, found := ix[v]
			if found {
				set.remove(name)
			}
		}
	}
}

func (mem *memNamedMapping) removeNameIdx(name string) {
	item := mem.nameToIdx[name]
	mem.removeIndexes(item, name)
	delete(mem.nameToIdx, name)
}

func (mem *memNamedMapping) putNameToIdx(name string, metadata interface{}) {
	oldItem, found := mem.nameToIdx[name]
	if found {
		mem.removeIndexes(oldItem, name)
	}

	item := &mappingItem{name, metadata, map[string][]string{}}
	mem.nameToIdx[name] = item
	mem.updateIndexes(item, name)
}

func (mem *memNamedMapping) publishToChannel(name string, value interface{}) {
	for subscriber, clb := range mem.subscribers {
		if clb != nil {
			dto := idxmap.NamedMappingGenericEvent{NamedMappingEvent: idxmap.NamedMappingEvent{
				Owner:         mem.owner,
				RegistryTitle: mem.title,
				Name:          name,
				Del:           false},
				Value: value,
			}
			mem.Debug("publish write to ", subscriber, dto)
			clb(dto)
		}
	}
}

func (mem *memNamedMapping) publishDelToChannel(name string, value interface{}) {
	for subscriber, clb := range mem.subscribers {
		if clb != nil {
			dto := idxmap.NamedMappingGenericEvent{NamedMappingEvent: idxmap.NamedMappingEvent{
				Owner:         mem.owner,
				RegistryTitle: mem.title,
				Name:          name,
				Del:           true},
				Value: value,
			}
			mem.Debug("publish del to ", subscriber, dto)
			clb(dto)
		}
	}
}

// nameSet is a simple implementation of a set holding names of type string
type nameSet struct {
	set map[string]interface{}
}

func newIndexSet() *nameSet {
	return &nameSet{set: map[string]interface{}{}}
}

func (s *nameSet) add(val string) {
	s.set[val] = nil
}

func (s *nameSet) remove(val string) {
	delete(s.set, val)
}

func (s *nameSet) contains(val string) bool {
	_, found := s.set[val]
	return found
}

func (s *nameSet) content() []string {
	var res []string
	for i := range s.set {
		res = append(res, i)
	}
	return res
}
