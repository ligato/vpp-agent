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

package ifaceidx

import (
	"fmt"
	"strings"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/cacheutil"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	linux_ifaces "github.com/ligato/vpp-agent/plugins/linuxplugin/model/interfaces"
)

const ipAddressKey = "ipAddrKey"

// Cache the VETH interfaces of a particular agent by watching transport. If change appears, it is registered in
// idx map
func Cache(watcher datasync.Watcher, caller core.PluginName) idxvpp.NameToIdxRW {
	resyncName := fmt.Sprintf("linux-iface-cache-%s-%s", caller, watcher)
	linuxIfIdx := nametoidx.NewNameToIdx(logroot.Logger(), caller, resyncName, IndexMetadata)

	helper := cacheutil.CacheHelper{
		Prefix:        linux_ifaces.InterfaceKeyPrefix(),
		IDX:           linuxIfIdx,
		DataPrototype: &linux_ifaces.LinuxInterfaces_Interface{Name: "aaa"},
		ParseName:     ParseNameFromKey,
	}

	go helper.DoWatching(resyncName, watcher)

	return linuxIfIdx
}

// IndexMetadata creates indexes for metadata. Index for IPAddress will be created
func IndexMetadata(metaData interface{}) map[string][]string {
	indexes := map[string][]string{}
	ifMeta, ok := metaData.(*linux_ifaces.LinuxInterfaces_Interface)
	if !ok || ifMeta == nil {
		return indexes
	}

	ip := ifMeta.IpAddresses
	if ip != nil {
		indexes[ipAddressKey] = ip
	}
	return indexes
}

// ParseNameFromKey returns suffix of the key (name)
func ParseNameFromKey(key string) (name string, err error) {
	lastSlashPos := strings.LastIndex(key, "/")
	if lastSlashPos > 0 && lastSlashPos < len(key)-1 {
		return key[lastSlashPos+1:], nil
	}

	return key, fmt.Errorf("Incorrect format of the key %s", key)
}
