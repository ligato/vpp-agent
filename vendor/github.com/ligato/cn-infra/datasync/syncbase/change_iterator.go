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

package syncbase

import (
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db"
)

// NewChangeIterator is a constructor
func NewChangeIterator(data []*Change) *ChangeIterator {
	return &ChangeIterator{data: data}
}

// ChangeIterator is a simple in memory implementation of data.Iterator
type ChangeIterator struct {
	data  []*Change
	index int
}

// GetNext TODO
func (it *ChangeIterator) GetNext() (kv datasync.KeyVal, changeType db.PutDel, allReceived bool) {
	if it.index >= len(it.data) {
		return nil, db.Put, true
	}

	ret := it.data[it.index]
	it.index++
	return ret, ret.changeType, false
}

// NewChange is a constructor
func NewChange(key string, value proto.Message, rev int64, changeType db.PutDel) *Change {
	return &Change{changeType, &KeyVal{key, &lazyProto{value}, rev}}
}

// NewChangeBytes is a constructor
func NewChangeBytes(key string, value []byte, rev int64, changeType db.PutDel) *Change {
	return &Change{changeType, &KeyValBytes{key, value, rev}}
}

// Change represents a single Key-value pair plus changeType
type Change struct {
	changeType db.PutDel
	datasync.KeyVal
}

// GetChangeType returns type of the change.
func (kv *Change) GetChangeType() db.PutDel {
	return kv.changeType
}
