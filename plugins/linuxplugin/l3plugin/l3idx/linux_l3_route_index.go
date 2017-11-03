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

package l3idx

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/l3plugin/model/l3"
)

const hostRouteNameKey = "hostRouteName"

// LinuxRouteIndex provides read-only access to mapping between software route indexes and route names
type LinuxRouteIndex interface {
	// GetMapping returns internal read-only mapping with metadata of type interface{}.
	GetMapping() idxvpp.NameToIdxRW

	// LookupIdx looks up previously stored item identified by index in mapping.
	LookupIdx(name string) (idx uint32, metadata *l3.LinuxStaticRoutes_Route, exists bool)

	// LookupName looks up previously stored item identified by name in mapping.
	LookupName(idx uint32) (name string, metadata *l3.LinuxStaticRoutes_Route, exists bool)

	// LookupNameByHostIfName looks up the interface identified by the name used in HostOs
	LookupNameByHostIfName(hostIfName string) []string

	// WatchNameToIdx allows to subscribe for watching changes in linuxIfIndex mapping
	WatchNameToIdx(subscriber core.PluginName, pluginChannel chan LinuxRouteIndexDto)
}

// LinuxRouteIndexRW is mapping between software route indexes (used internally in VPP)
// and routes entry names.
type LinuxRouteIndexRW interface {
	LinuxRouteIndex

	// RegisterName adds new item into name-to-index mapping.
	RegisterName(name string, idx uint32, ifMeta *l3.LinuxStaticRoutes_Route)

	// UnregisterName removes an item identified by name from mapping
	UnregisterName(name string) (idx uint32, metadata *l3.LinuxStaticRoutes_Route, exists bool)
}

// LinuxRouteIndexDto represents an item sent through watch channel in LinuxRouteIndex.
// In contrast to NameToIdxDto it contains typed metadata.
type LinuxRouteIndexDto struct {
	idxvpp.NameToIdxDtoWithoutMeta
	Metadata *l3.LinuxStaticRoutes_Route
}

// linuxRouteIndex is type-safe implementation of mapping between Software ARP index
// and ARP name.
type linuxRouteIndex struct {
	mapping idxvpp.NameToIdxRW
}

// NewLinuxRouteIndex creates new instance of linuxRouteIndex.
func NewLinuxRouteIndex(mapping idxvpp.NameToIdxRW) LinuxRouteIndexRW {
	return &linuxRouteIndex{mapping: mapping}
}

// GetMapping returns internal read-only mapping. It is used in tests to inspect the content of the linuxArpIndex.
func (linuxRouteIndex *linuxRouteIndex) GetMapping() idxvpp.NameToIdxRW {
	return linuxRouteIndex.mapping
}

// LookupIdx looks up previously stored item identified by index in mapping.
func (linuxRouteIndex *linuxRouteIndex) LookupIdx(name string) (idx uint32, metadata *l3.LinuxStaticRoutes_Route, exists bool) {
	idx, meta, exists := linuxRouteIndex.mapping.LookupIdx(name)
	if exists {
		metadata = linuxRouteIndex.castMetadata(meta)
	}
	return idx, metadata, exists
}

// LookupName looks up previously stored item identified by name in mapping.
func (linuxRouteIndex *linuxRouteIndex) LookupName(idx uint32) (name string, metadata *l3.LinuxStaticRoutes_Route, exists bool) {
	name, meta, exists := linuxRouteIndex.mapping.LookupName(idx)
	if exists {
		metadata = linuxRouteIndex.castMetadata(meta)
	}
	return name, metadata, exists
}

// LookupNameByIP returns names of items that contains given IP address in metadata
func (linuxRouteIndex *linuxRouteIndex) LookupNameByHostIfName(hostARPName string) []string {
	return linuxRouteIndex.mapping.LookupNameByMetadata(hostARPNameKey, hostARPName)
}

// RegisterName adds new item into name-to-index mapping.
func (linuxRouteIndex *linuxRouteIndex) RegisterName(name string, idx uint32, ifMeta *l3.LinuxStaticRoutes_Route) {
	linuxRouteIndex.mapping.RegisterName(name, idx, ifMeta)
}

// UnregisterName removes an item identified by name from mapping
func (linuxRouteIndex *linuxRouteIndex) UnregisterName(name string) (idx uint32, metadata *l3.LinuxStaticRoutes_Route, exists bool) {
	idx, meta, exists := linuxRouteIndex.mapping.UnregisterName(name)
	return idx, linuxRouteIndex.castMetadata(meta), exists
}

// WatchNameToIdx allows to subscribe for watching changes in linuxIfIndex mapping
func (linuxRouteIndex *linuxRouteIndex) WatchNameToIdx(subscriber core.PluginName, pluginChannel chan LinuxRouteIndexDto) {
	ch := make(chan idxvpp.NameToIdxDto)
	linuxRouteIndex.mapping.Watch(subscriber, nametoidx.ToChan(ch))
	go func() {
		for c := range ch {
			pluginChannel <- LinuxRouteIndexDto{
				NameToIdxDtoWithoutMeta: c.NameToIdxDtoWithoutMeta,
				Metadata:                linuxRouteIndex.castMetadata(c.Metadata),
			}

		}
	}()
}

func (linuxRouteIndex *linuxRouteIndex) castMetadata(meta interface{}) *l3.LinuxStaticRoutes_Route {
	if ifMeta, ok := meta.(*l3.LinuxStaticRoutes_Route); ok {
		return ifMeta
	}

	return nil
}
