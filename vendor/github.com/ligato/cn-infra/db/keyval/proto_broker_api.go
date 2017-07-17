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

package keyval

import (
	"github.com/golang/protobuf/proto"
)

// ProtoBroker is decorator that allows to read/write proto file modelled data.
// It marshals/unmarshals go structures to slice of bytes and vice versa behind the scenes.
type ProtoBroker interface {
	// Put puts single key-value pair into etcd
	Put(key string, data proto.Message, opts ...PutOption) error
	// NewTxn creates a transaction
	NewTxn() ProtoTxn
	// GetValue retrieves one item under the provided key. If the item exists it is unmarshaled into the reqObj.
	GetValue(key string, reqObj proto.Message) (found bool, revision int64, err error)
	// ListValues returns an iterator that enables to traverse all items stored under the provided key
	ListValues(key string) (ProtoKeyValIterator, error)
	// ListKeys is similar to the ListValues the difference is that values are not fetched
	ListKeys(prefix string) (ProtoKeyIterator, error)
	// Delete removes data stored under the key
	Delete(key string, opts ...DelOption) (existed bool, err error)
}

// ProtoKvPair group getter for single key-value pair
type ProtoKvPair interface {
	// GetKey returns the key of the pair
	GetKey() string
	// GetValue returns the value of the pair
	GetValue(proto.Message) error
}

// ProtoKeyIterator is an iterator returned by ListKeys call
type ProtoKeyIterator interface {
	// GetNext retrieves the following item from the context.
	GetNext() (key string, rev int64, allReceived bool)
}

// ProtoKeyVal represents a single key-value pair
type ProtoKeyVal interface {
	ProtoKvPair
	// GetRevision returns revision associated with the latest change in the key-value pair
	GetRevision() int64
}

// ProtoKeyValIterator is an iterator returned by ListValues call.
type ProtoKeyValIterator interface {
	// GetNext retrieves the following value from the context. GetValue is unmarshaled into the provided argument.
	GetNext() (kv ProtoKeyVal, allReceived bool)
}
