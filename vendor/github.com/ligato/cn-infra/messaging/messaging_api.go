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

package messaging

import (
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
)

// Mux defines API for the plugins that use access to kafka brokers.
type Mux interface {
	// Creates new Kafka synchronous publisher sending messages to given topic. Partitioner has to be set to 'hash' (default)
	// or 'random' scheme, otherwise an error is thrown
	NewSyncPublisher(topic string) (ProtoPublisher, error)

	// Creates new Kafka synchronous publisher sending messages to given topic and partition. Partitioner has to be
	// set to 'manual' scheme, otherwise an error is thrown
	NewSyncPublisherToPartition(topic string, partition int32) (ProtoPublisher, error)

	// Creates new Kafka asynchronous publisher sending messages to given topic. Partitioner has to be set to 'hash' (default)
	// or 'random' scheme, otherwise an error is thrown
	NewAsyncPublisher(topic string, successClb func(ProtoMessage), errorClb func(err ProtoMessageErr)) (ProtoPublisher, error)

	// Creates new Kafka asynchronous publisher sending messages to given topic and partition. Partitioner has to be
	// set to 'manual' scheme, otherwise an error is thrown
	NewAsyncPublisherToPartition(topic string, partition int32,
		successClb func(ProtoMessage), errorClb func(err ProtoMessageErr)) (ProtoPublisher, error)

	// Initializes new watcher which can start/stop watching on topic, eventually partition and offset
	NewWatcher(subscriberName string) ProtoWatcher
}

// ProtoPublisher allows to publish a message of type proto.Message into messaging system.
type ProtoPublisher interface {
	datasync.KeyProtoValWriter
}

// ProtoWatcher allows to subscribe for receiving of messages published to given topics.
type ProtoWatcher interface {
	// Watch given topic. Returns error if 'manual' partitioner scheme is chosen
	Watch(msgCallback func(ProtoMessage), topics ...string) error

	// Stop watching on topic. Return error if topic is not subscribed
	StopWatch(topic string) error

	// Watch given topic, partition and offset. Offset is the oldest message index consumed, all previously written
	// messages are ignored. Manual partitioner must be set, otherwise error is thrown
	WatchPartition(msgCallback func(ProtoMessage), topic string, partition int32, offset int64) error

	// Stop watching on topic/partition/offset. Return error if such a combination is not subscribed
	StopWatchPartition(topic string, partition int32, offset int64) error
}

// ProtoMessage defines functions for inspection of a message receive from messaging system.
type ProtoMessage interface {
	keyval.ProtoKvPair
	GetTopic() string
	GetPartition() int32
	GetOffset() int64
}

// ProtoMessageErr represents a message that was not published successfully to a messaging system.
type ProtoMessageErr interface {
	ProtoMessage
	Error() error
}
