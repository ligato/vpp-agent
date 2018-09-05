// Copyright (c) 2018 Cisco and/or its affiliates.
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

package registry

import (
	"container/list"

	. "github.com/ligato/cn-infra/kvscheduler/api"
)

const (
	// maxKeyCacheSize is the maximum number of key->descriptor entries the registry
	// will cache.
	maxKeyCacheSize = 500
)

// registry is an implementation of Registry for descriptors.
type registry struct {
	descriptors     map[string]KVDescriptor  // descriptor name -> descriptor
	keyToCacheEntry map[string]*list.Element // key -> cache entry
	keyCache        *list.List               // doubly linked list of cached entries key->descriptor
}

// cacheEntry encapsulates data for one entry in registry.keyCache
type cacheEntry struct {
	key        string
	descriptor KVDescriptor
}

// NewRegistry creates a new instance of registry.
func NewRegistry() Registry {
	return &registry{
		descriptors:     make(map[string]KVDescriptor),
		keyToCacheEntry: make(map[string]*list.Element),
		keyCache:        list.New(),
	}
}

// RegisterDescriptor add new descriptor into the registry.
func (reg *registry) RegisterDescriptor(descriptor KVDescriptor) {
	reg.descriptors[descriptor.GetName()] = descriptor
}

// GetAllDescriptors returns all registered descriptors.
func (reg *registry) GetAllDescriptors() (descriptors []KVDescriptor) {
	// order topologically respecting dependencies.
	deps := make(map[string][]string)
	for _, descriptor := range reg.descriptors {
		descriptors = append(descriptors, descriptor)
		deps[descriptor.GetName()] = descriptor.DumpDependencies()
	}
	return topologicalOrder(descriptors, deps)
}

// GetDescriptor returns descriptor with the given name.
func (reg *registry) GetDescriptor(name string) KVDescriptor {
	descriptor, has := reg.descriptors[name]
	if !has {
		return nil
	}
	return descriptor
}

// GetDescriptorForKey returns descriptor handling the given key.
func (reg *registry) GetDescriptorForKey(key string) KVDescriptor {
	elem, cached := reg.keyToCacheEntry[key]
	if cached {
		// get descriptor from the cache
		entry := elem.Value.(*cacheEntry)
		reg.keyCache.MoveToFront(elem)
		return entry.descriptor
	}
	if reg.keyCache.Len() == maxKeyCacheSize {
		// the cache is full => remove the last used key
		toRemove := reg.keyCache.Back()
		toRemoveKey := toRemove.Value.(*cacheEntry).key
		delete(reg.keyToCacheEntry, toRemoveKey)
		reg.keyCache.Remove(toRemove)
	}
	// find the descriptor
	var keyDescriptor KVDescriptor
	for _, descriptor := range reg.descriptors {
		if descriptor.KeySelector(key) {
			keyDescriptor = descriptor
			break
		}
	}
	// add entry to cache
	entry := &cacheEntry{key: key, descriptor: keyDescriptor}
	elem = reg.keyCache.PushFront(entry)
	reg.keyToCacheEntry[key] = elem
	return keyDescriptor
}

// topologicalOrder orders descriptors topologically by Kahn's algorithm to respect
// the dump dependencies.
func topologicalOrder(descriptors []KVDescriptor, deps map[string][]string) (sorted []KVDescriptor) {
	// move descriptors from the list to a map
	descMap := make(map[string]KVDescriptor)
	for _, desc := range descriptors {
		descMap[desc.GetName()] = desc
	}
	for len(descMap) > 0 {
		// find first node without any dependencies within the remaining set
		oneDown := false
		for descName, desc := range descMap {
			descDeps := deps[descName]
			if len(descDeps) == 0 {
				sorted = append(sorted, desc)
				delete(descMap, descName)
				oneDown = true
				// remove the dependency edges going to this descriptor
				for descName2, descDeps2 := range deps {
					deps[descName2] = removeFromSlice(descDeps2, descName)
				}
				break
			}
		}
		if !oneDown {
			panic("Dependency cycle!")
		}
	}
	return sorted
}

func removeFromSlice(slice []string, itemToRemove string) []string {
	var filtered []string
	for _, item := range slice {
		if item == itemToRemove {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}
