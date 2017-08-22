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

package msg

import (
	"strings"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/syncbase"
)

// DataMsgRequestToKVs TODO
func DataMsgRequestToKVs(req *DataMsgRequest, keyPrefixes []string) map[string] /*keyPrefix*/ []datasync.KeyVal {
	ret := map[string] /*keyPrefix*/ []datasync.KeyVal{}
	for _, dataResync := range req.GetDataResyncs() {
		for _, keyPrefix := range keyPrefixes {
			if strings.HasPrefix(dataResync.Key, keyPrefix) {
				kvs, found := ret[keyPrefix]
				kv := syncbase.NewKeyValBytes(dataResync.Key, dataResync.Content, 0) /*TODO prev value*/
				if !found {
					kvs = []datasync.KeyVal{kv}
				} else {
					kvs = append(kvs, kv)
				}
				ret[keyPrefix] = kvs
			}
		}
	}

	return ret
}
