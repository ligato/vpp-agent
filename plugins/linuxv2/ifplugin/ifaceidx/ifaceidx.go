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

package ifaceidx

import (
	"strconv"
	"time"

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/idxmap"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/idxmap/mem"

	nsmodel "github.com/ligato/vpp-agent/plugins/linuxv2/model/namespace"
)

// LinuxIfMetadataIndex provides read-only access to mapping with Linux interface
// metadata. It extends from NameToIndex.
type LinuxIfMetadataIndex interface {
	// LookupByName retrieves a previously stored metadata of interface
	// identified by logical <name>. If there is no interface associated with
	// the given name in the mapping, the <exists> is returned as *false* and
	// <metadata> as *nil*.
	LookupByName(name string) (metadata *LinuxIfMetadata, exists bool)

	// LookupByLinuxIfIndex retrieves a previously stored interface identified in
	// Linux by the given <linuxIfIndex> inside the given <namespace>.
	// If there is no interface associated with the given index, <exists> is returned
	// as *false* with <name> and <metadata> both set to empty values.
	LookupByLinuxIfIndex(linuxIfIndex int, namespace string) (name string, metadata *LinuxIfMetadata, exists bool)

	// LookupByLinuxIfName retrieves a previously stored interface identified in
	// Linux by the given <linuxIfName> inside the given <namespace>.
	// If there is no interface associated with the given name, <exists> is returned
	// as *false* with <name> and <metadata> both set to empty values.
	LookupByLinuxIfName(linuxIfName string, namespace string) (name string, metadata *LinuxIfMetadata, exists bool)

	// ListAllInterfaces returns slice of names of all interfaces in the mapping.
	ListAllInterfaces() (names []string)

	// WatchInterfaces allows to subscribe to watch for changes in the mapping
	// of interface metadata.
	WatchInterfaces(subscriber infra.PluginName, channel chan<- LinuxIfMetadataIndexDto)
}

// LinuxIfMetadataIndexRW provides read-write access to mapping with interface
// metadata.
type LinuxIfMetadataIndexRW interface {
	LinuxIfMetadataIndex
	idxmap.NamedMappingRW
}

// LinuxIfMetadata collects metadata for Linux interface used in secondary lookups.
type LinuxIfMetadata struct {
	LinuxIfIndex int
	HostIfName   string
	Namespace    *nsmodel.Namespace
}

// LinuxIfMetadataIndexDto represents an item sent through watch channel in LinuxIfMetadataIndex.
// In contrast to NamedMappingGenericEvent, it contains typed interface metadata.
type LinuxIfMetadataIndexDto struct {
	idxmap.NamedMappingEvent
	Metadata *LinuxIfMetadata
}

// linuxIfMetadataIndex is type-safe implementation of mapping between interface
// name and metadata of type *LinuxIfMetadata.
type linuxIfMetadataIndex struct {
	idxmap.NamedMappingRW /* embeds */
	log logging.Logger
}

const (
	// linuxIfNameKeyPrefix is used as prefix (appended with namespace) for secondary key
	// used to search interface by host name.
	linuxIfNameKeyPrefix = "linux-if-name/"

	// linuxIfIndexKeyPrefix is used as prefix (appended with namespace) for secondary key
	// used to search interface by host index.
	linuxIfIndexKeyPrefix = "linux-if-index/"
)

// NewLinuxIfIndex creates a new instance implementing LinuxIfMetadataIndexRW.
func NewLinuxIfIndex(logger logging.Logger, title string) LinuxIfMetadataIndexRW {
	return &linuxIfMetadataIndex{
		NamedMappingRW: mem.NewNamedMapping(logger, title, indexMetadata),
	}
}

// LookupByName retrieves a previously stored metadata of interface
// identified by logical <name>. If there is no interface associated with
// the give/ name in the mapping, the <exists> is returned as *false* and
// <metadata> as *nil*.
func (ifmx *linuxIfMetadataIndex) LookupByName(name string) (metadata *LinuxIfMetadata, exists bool) {
	meta, found := ifmx.GetValue(name)
	if found {
		if typedMeta, ok := meta.(*LinuxIfMetadata); ok {
			return typedMeta, found
		}
	}
	return nil, false
}

// LookupByLinuxIfIndex retrieves a previously stored interface identified in
// Linux by the given <linuxIfIndex> inside the given <namespace>.
// If there is no interface associated with the given index, <exists> is returned
// as *false* with <name> and <metadata> both set to empty values.
func (ifmx *linuxIfMetadataIndex) LookupByLinuxIfIndex(linuxIfIndex int, namespace string) (name string, metadata *LinuxIfMetadata, exists bool) {
	return ifmx.lookupBySecondaryKey(linuxIfIndexKeyPrefix, namespace, strconv.FormatInt(int64(linuxIfIndex), 10))
}

// LookupByLinuxIfName retrieves a previously stored interface identified in
// Linux by the given <linuxIfName> inside the given <namespace>.
// If there is no interface associated with the given name, <exists> is returned
// as *false* with <name> and <metadata> both set to empty values.
func (ifmx *linuxIfMetadataIndex) LookupByLinuxIfName(linuxIfName string, namespace string) (name string, metadata *LinuxIfMetadata, exists bool) {
	return ifmx.lookupBySecondaryKey(linuxIfNameKeyPrefix, namespace, linuxIfName)
}

// lookupBySecondaryKey performs lookup by a secondary key within a single
// namespace that should yield one or zero matches.
func (ifmx *linuxIfMetadataIndex) lookupBySecondaryKey(keyPrefix, namespace, value string) (name string, metadata *LinuxIfMetadata, exists bool) {
	if namespace == "" {
		namespace = nsmodel.DefaultNamespaceName
	}
	res := ifmx.ListNames(keyPrefix + namespace, value)
	if len(res) != 1 {
		return
	}
	untypedMeta, found := ifmx.GetValue(res[0])
	if found {
		if ifMeta, ok := untypedMeta.(*LinuxIfMetadata); ok {
			return res[0], ifMeta, found
		}
	}
	return
}

// ListAllInterfaces returns slice of names of all interfaces in the mapping.
func (ifmx *linuxIfMetadataIndex) ListAllInterfaces() (names []string) {
	return ifmx.ListAllNames()
}

// WatchInterfaces allows to subscribe to watch for changes in the mapping
// if interface metadata.
func (ifmx *linuxIfMetadataIndex) WatchInterfaces(subscriber infra.PluginName, channel chan<- LinuxIfMetadataIndexDto) {
	watcher := func(dto idxmap.NamedMappingGenericEvent) {
		typedMeta, ok := dto.Value.(*LinuxIfMetadata)
		if !ok {
			return
		}
		msg := LinuxIfMetadataIndexDto{
			NamedMappingEvent: dto.NamedMappingEvent,
			Metadata:          typedMeta,
		}
		select {
		case channel <- msg:
		case <-time.After(idxmap.DefaultNotifTimeout):
			ifmx.log.Warn("Unable to deliver notification")
		}
	}
	ifmx.Watch(subscriber, watcher)
}

// indexMetadata is an index function used for interface metadata.
func indexMetadata(metaData interface{}) map[string][]string {
	indexes := make(map[string][]string)

	ifMeta, ok := metaData.(*LinuxIfMetadata)
	if !ok || ifMeta == nil {
		return indexes
	}

	ns := nsmodel.DefaultNamespaceName
	if ifMeta.Namespace != nil {
		ns = ifMeta.Namespace.Name
	}

	indexes[linuxIfIndexKeyPrefix + ns] = []string{strconv.FormatInt(int64(ifMeta.LinuxIfIndex), 10)}
	indexes[linuxIfNameKeyPrefix + ns] = []string{ifMeta.HostIfName}
	return indexes
}