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

package idxvpp2

import (
	"strconv"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/idxmap/mem"
	"github.com/ligato/cn-infra/logging"
)

// WithSwIfIndex is interface that items with sw_if_index must implement to get
// indexed by SwIfIndex.
type WithSwIfIndex interface {
	// GetSwIfIndex should return sw_if_index assigned to the item.
	GetSwIfIndex() uint32
}

// SwIfIndex is the "user API" to the registry of VPP items indexed by sw_if_index.
// It provides read-only access intended for plugins that need to do the conversions
// between logical names from NB and VPP IDs.
type SwIfIndex interface {
	// LookupByName retrieves a previously stored item identified by
	// <name>. If there is no item associated with the give name in the mapping,
	// the <exists> is returned as *false* and <item> as *nil*.
	LookupByName(name string) (item WithSwIfIndex, exists bool)

	// LookupIdx retrieves a previously stored item identified in VPP by the given
	// <swIfIndex>.
	// If there is no item associated with the given index, <exists> is returned
	// as *false* with <name> and <item> both set to empty values.
	LookupBySwIfIndex(swIfIndex uint32) (name string, item WithSwIfIndex, exists bool)

	// WatchItems subscribes to receive notifications about the changes in the
	// mapping related to items with sw_if_index.
	WatchItems(subscriber core.PluginName, channel chan<- SwIfIndexDto)
}

// SwIfIndexRW is the "owner API" to the NameToIdx registry. Using this
// API the owner is able to add/update and delete associations between logical
// names and VPP items identified by sw_if_index.
type SwIfIndexRW interface {
	SwIfIndex
	idxmap.NamedMappingRW
}

// OnlySwIfIndex can be used to add items into SwIfIndex with sw_if_index
// as the only information associated with each item.
type OnlySwIfIndex struct {
	SwIfIndex uint32
}

// GetSwIfIndex returns sw_if_index assigned to the item.
func (idx *OnlySwIfIndex) GetSwIfIndex() uint32 {
	return idx.SwIfIndex
}

// SwIfIndexDto represents an item sent through watch channel in swIfIndex.
// In contrast to NamedMappingGenericEvent, it contains item casted to WithSwIfIndex.
type SwIfIndexDto struct {
	idxmap.NamedMappingEvent
	Item WithSwIfIndex
}

// swIfIndex implements NamedMapping for items with sw_if_index.
type swIfIndex struct {
	idxmap.NamedMappingRW
	log logging.Logger
}

const (
	// swIfIdxKey is a secondary index used to create association between
	// item name and sw_if_index from VPP.
	swIfIdxKey = "sw_if_index"
)

// NewSwIfIndex creates a new instance implementing SwIfIndexRW.
// User can optionally extend the secondary indexes through <indexFunction>.
func NewSwIfIndex(logger logging.Logger, title string,
	indexFunction func(interface{}) map[string][]string) SwIfIndexRW {
	return &swIfIndex{
		NamedMappingRW: mem.NewNamedMapping(logger, title,
			func(item interface{}) map[string][]string {
				idxs := internalIndexFunction(item)

				if indexFunction != nil {
					userIdxs := indexFunction(item)
					for k, v := range userIdxs {
						idxs[k] = v
					}
				}
				return idxs
			}),
	}
}

// LookupByName retrieves a previously stored item identified by
// <name>. If there is no item associated with the give name in the mapping,
// the <exists> is returned as *false* and <item> as *nil*.
func (swix *swIfIndex) LookupByName(name string) (item WithSwIfIndex, exists bool) {
	value, found := swix.GetValue(name)
	if found {
		if itemWithIndex, ok := value.(WithSwIfIndex); ok {
			return itemWithIndex, found
		}
	}
	return nil, false
}

// LookupIdx retrieves a previously stored item identified in VPP by the given
// <swIfIndex>.
// If there is no item associated with the given index, exists is returned
// as *false* with <name> and <item> both set to empty values.
func (swix *swIfIndex) LookupBySwIfIndex(swIfIndex uint32) (name string, item WithSwIfIndex, exists bool) {
	res := swix.ListNames(swIfIdxKey, strconv.FormatUint(uint64(swIfIndex), 10))
	if len(res) != 1 {
		return
	}
	value, found := swix.GetValue(res[0])
	if found {
		if itemWithIndex, ok := value.(WithSwIfIndex); ok {
			return res[0], itemWithIndex, found
		}
	}
	return
}

// WatchItems subscribes to receive notifications about the changes in the
// mapping.
func (swix *swIfIndex) WatchItems(subscriber core.PluginName, channel chan<- SwIfIndexDto) {
	watcher := func(dto idxmap.NamedMappingGenericEvent) {
		itemWithIndex, ok := dto.Value.(WithSwIfIndex)
		if !ok {
			return
		}
		msg := SwIfIndexDto{
			NamedMappingEvent: dto.NamedMappingEvent,
			Item:              itemWithIndex,
		}
		select {
		case channel <- msg:
		case <-time.After(idxmap.DefaultNotifTimeout):
			swix.log.Warn("Unable to deliver notification")
		}
	}
	swix.Watch(subscriber, watcher)
}

// internalIndexFunction is an index function used internally for sw_if_index.
func internalIndexFunction(item interface{}) map[string][]string {
	indexes := map[string][]string{}
	itemWithIndex, ok := item.(WithSwIfIndex)
	if !ok || itemWithIndex == nil {
		return indexes
	}

	indexes[swIfIdxKey] = []string{strconv.FormatUint(uint64(itemWithIndex.GetSwIfIndex()), 10)}
	return indexes
}
