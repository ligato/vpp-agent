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

package datasync

import (
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/db"
	"io"
)

// TransportAdapter is an high-level abstraction of a transport to a remote
// data back end ETCD/Kafka/Rest/GRPC, used by Agent plugins to access data
// in a uniform & consistent way.
type TransportAdapter interface {
	Watcher
	Publisher
}

// Watcher is used by plugin to subscribe to both data change events and
// data resync events. Multiple keys can be specified, the caller will
// be subscribed to events on each key.
type Watcher interface {
	// WatchData using ETCD or any other data transport
	WatchData(resyncName string, changeChan chan ChangeEvent, resyncChan chan ResyncEvent,
		keyPrefixes ...string) (WatchDataRegistration, error)
}

// Publisher allows plugins to push their data changes to a data store.
type Publisher interface {
	// PublishData to ETCD or any other data transport (from other Agent Plugins)
	PublishData(key string, data proto.Message) error
}

// WatchDataRegistration is a facade that avoids importing the io.Closer package
// into Agent plugin implementations.
type WatchDataRegistration interface {
	io.Closer
}

// LazyValue defines value that is unmarshalled into proto message on demand.
// The reason for defining interface with only one method is primary to unify interfaces in this package
type LazyValue interface {
	// GetValue gets the current in the data change event.
	// The caller must provide an address of a proto message buffer
	// for each value.
	// returns:
	// - revision associated with the latest change in the key-value pair
	// - error if value argument can not be properly filled
	GetValue(value proto.Message) error
}

// LazyValueWithRev defines value that is unmarshalled into proto message on demand with a revision.
// The reason for defining interface with only one method is primary to unify interfaces in this package
type LazyValueWithRev interface {
	LazyValue

	// GetRevision gets revision of current value
	GetRevision() (rev int64)
}

// ChangeValue represents single propagated change.
type ChangeValue interface {
	GetChangeType() db.PutDel
	LazyValueWithRev
}

// ChangeEvent is used as the data type for the change channel
// (see the VPP Standard Plugins API). A data change event contains
// a key identifying where the change happened and two values for
// data stored under that key: the value *before* the change (previous
// value) and the value *after* the change (current value).
type ChangeEvent interface {
	CallbackResult

	// GetKey returns the key in the data change event key-value tuple
	GetKey() string

	ChangeValue

	// GetPrevValue gets previous value in the data change event.
	// The caller must provide an address of a proto message buffer
	// for each value.
	// returns:
	// - prevValueExist flag is set to 'true' if prevValue was filled
	// - error if value argument can not be properly filled
	GetPrevValue(prevValue proto.Message) (prevValueExist bool, err error)
}

// ResyncEvent is used as the data type for the resync channel
// (see the ifplugin API)
type ResyncEvent interface {
	CallbackResult

	GetValues() map[string] /*keyPrefix*/ KeyValIterator
}

// KeyVal represents a single key-value pair
type KeyVal interface {
	// GetKey returns the key of the pair
	GetKey() string

	LazyValueWithRev
}

// KeyValIterator is an iterator for KeyVals
type KeyValIterator interface {
	// GetNext retrieves the next value from the iterator context.  The retrieved
	// value is unmarshaled into the provided argument. The allReceived flag is
	// set to true on the last KeyVal pair in the context.
	GetNext() (kv KeyVal, allReceived bool)
}

// CallbackResult can be used by an event receiver to indicate to the event producer
// whether an operation was successful (error is nil) or unsuccessful (error is
// not nil)
//
// DoneMethod is reused later. There are at least two implementations DoneChannel, DoneCallback
type CallbackResult interface {
	// Done allows plugins that are processing data change/resync to send feedback
	// If there was no error the Done(nil) needs to be called. Use the noError=nil
	// definition for better readability, for example:
	//     Done(noError).
	Done(error)
}
