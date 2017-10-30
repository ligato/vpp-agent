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
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
)

// ProtoBroker is decorator that allows to read/write proto file modelled data.
// It marshals/unmarshals go structures to slice of bytes and vice versa behind
// the scenes.
type ProtoBroker interface {
	// Put puts single key-value pair into key value store.
	datasync.KeyProtoValWriter
	// NewTxn creates a transaction.
	NewTxn() ProtoTxn
	// GetValue retrieves one item under the provided <key>. If the item exists,
	// it is unmarshaled into the <reqObj>.
	GetValue(key string, reqObj proto.Message) (found bool, revision int64, err error)
	// ListValues returns an iterator that enables to traverse all items stored
	// under the provided <key>.
	ListValues(key string) (ProtoKeyValIterator, error)
	// ListKeys returns an iterator that allows to traverse all keys from data
	// store that share the given <prefix>
	ListKeys(prefix string) (ProtoKeyIterator, error)
	// Delete removes data stored under the <key>.
	Delete(key string, opts ...datasync.DelOption) (existed bool, err error)
}

// ProtoKvPair group getter for single key-value pair
type ProtoKvPair interface {
	datasync.LazyValue
	datasync.WithPrevValue
	datasync.WithKey
}

// ProtoKeyIterator is an iterator returned by ListKeys call
type ProtoKeyIterator interface {
	// GetNext retrieves the following item from the context.
	GetNext() (key string, rev int64, stop bool)
	// Closer is needed for closing the iterator (please check error returned by Close method)
	io.Closer
}

// ProtoKeyVal represents a single key-value pair
type ProtoKeyVal interface {
	ProtoKvPair
	datasync.WithRevision
}

// ProtoKeyValIterator is an iterator returned by ListValues call.
type ProtoKeyValIterator interface {
	// GetNext retrieves the following value from the context. GetValue is unmarshaled into the provided argument.
	GetNext() (kv ProtoKeyVal, stop bool)
	// Closer is needed for closing the iterator (please check error returned by Close method)
	io.Closer
}
