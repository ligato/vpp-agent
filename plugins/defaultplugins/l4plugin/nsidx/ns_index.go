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

package nsidx

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/model/l4"
)

// AppNsIndex provides read-only access to mapping between indexes (used internally in VPP) and AppNamespace indexes.
type AppNsIndex interface {
	// GetMapping returns internal read-only mapping with metadata of type interface{}.
	GetMapping() idxvpp.NameToIdxRW

	// LookupIdx looks up previously stored item identified by index in mapping.
	LookupIdx(name string) (idx uint32, metadata *l4.AppNamespaces_AppNamespace, exists bool)

	// LookupName looks up previously stored item identified by name in mapping.
	LookupName(idx uint32) (name string, metadata *l4.AppNamespaces_AppNamespace, exists bool)

	// LookupNamesByInterface returns names of items that contains given IP address in metadata
	LookupNamesByInterface(ifName string) []*l4.AppNamespaces_AppNamespace

	// WatchNameToIdx allows to subscribe for watching changes in appNsIndex mapping
	WatchNameToIdx(subscriber core.PluginName, pluginChannel chan ChangeDto)
}

// AppNsIndexRW is mapping between indexes (used internally in VPP) and AppNamespace indexes.
type AppNsIndexRW interface {
	AppNsIndex

	// RegisterName adds new item into name-to-index mapping.
	RegisterName(name string, idx uint32, metadata *l4.AppNamespaces_AppNamespace)

	// UnregisterName removes an item identified by name from mapping
	UnregisterName(name string) (idx uint32, metadata *l4.AppNamespaces_AppNamespace, exists bool)
}

// appNsIndex is type-safe implementation of mapping between AppNamespace index
// and its name. It holds as well metadata of type *AppNsMeta.
type appNsIndex struct {
	mapping idxvpp.NameToIdxRW
}

// ChangeDto represents an item sent through watch channel in appNsIndex.
// In contrast to NameToIdxDto it contains typed metadata.
type ChangeDto struct {
	idxvpp.NameToIdxDtoWithoutMeta
	Metadata *l4.AppNamespaces_AppNamespace
}

const (
	appNsNameIndexKey = "appNsKey"
)

// NewAppNsIndex creates new instance of appNsIndex.
func NewAppNsIndex(mapping idxvpp.NameToIdxRW) AppNsIndexRW {
	return &appNsIndex{mapping: mapping}
}

// GetMapping returns internal read-only mapping. It is used in tests to inspect the content of the appNsIndex.
func (swi *appNsIndex) GetMapping() idxvpp.NameToIdxRW {
	return swi.mapping
}

// RegisterName adds new item into name-to-index mapping.
func (swi *appNsIndex) RegisterName(name string, idx uint32, appNsMeta *l4.AppNamespaces_AppNamespace) {
	swi.mapping.RegisterName(name, idx, appNsMeta)
}

// UnregisterName removes an item identified by name from mapping
func (swi *appNsIndex) UnregisterName(name string) (idx uint32, metadata *l4.AppNamespaces_AppNamespace, exists bool) {
	idx, meta, exists := swi.mapping.UnregisterName(name)
	return idx, swi.castMetadata(meta), exists
}

// LookupIdx looks up previously stored item identified by index in mapping.
func (swi *appNsIndex) LookupIdx(name string) (idx uint32, metadata *l4.AppNamespaces_AppNamespace, exists bool) {
	idx, meta, exists := swi.mapping.LookupIdx(name)
	if exists {
		metadata = swi.castMetadata(meta)
	}
	return idx, metadata, exists
}

// LookupName looks up previously stored item identified by name in mapping.
func (swi *appNsIndex) LookupName(idx uint32) (name string, metadata *l4.AppNamespaces_AppNamespace, exists bool) {
	name, meta, exists := swi.mapping.LookupName(idx)
	if exists {
		metadata = swi.castMetadata(meta)
	}
	return name, metadata, exists
}

// LookupNamesByInterface returns all names related to the provided interface
func (swi *appNsIndex) LookupNamesByInterface(ifName string) []*l4.AppNamespaces_AppNamespace {
	var match []*l4.AppNamespaces_AppNamespace
	for _, name := range swi.mapping.ListNames() {
		_, meta, found := swi.LookupIdx(name)
		if found && meta != nil && meta.Interface == ifName {
			match = append(match, meta)
		}
	}
	return match
}

// WatchNameToIdx allows to subscribe for watching changes in appNsIndex mapping
func (swi *appNsIndex) WatchNameToIdx(subscriber core.PluginName, pluginChannel chan ChangeDto) {
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

func (swi *appNsIndex) castMetadata(meta interface{}) *l4.AppNamespaces_AppNamespace {
	appNsMeta, ok := meta.(*l4.AppNamespaces_AppNamespace)
	if !ok {
		return nil
	}
	return appNsMeta
}

func (swi *appNsIndex) castIfMetadata(meta interface{}) string {
	ifMeta, ok := meta.(string)
	if !ok {
		return ""
	}
	return ifMeta
}
