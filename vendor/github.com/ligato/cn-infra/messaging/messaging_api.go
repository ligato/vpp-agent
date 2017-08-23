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
	NewSyncPublisher(topic string) ProtoPublisher
	NewAsyncPublisher(topic string, successClb func(ProtoMessage), errorClb func(err ProtoMessageErr)) ProtoPublisher
	NewWatcher(subscriberName string) ProtoWatcher
}

// ProtoPublisher allows to publish a message of type proto.Message into messaging system.
type ProtoPublisher interface {
	datasync.KeyProtoValWriter
}

// ProtoWatcher allows to subscribe for receiving of messages published to given topics.
type ProtoWatcher interface {
	Watch(msgCallback func(ProtoMessage), topics ...string) error
	StopWatch(topic string) error
}

// ProtoMessage defines functions for inspection of a message receive from messaging system.
type ProtoMessage interface {
	keyval.ProtoKvPair
	GetTopic() string
}

// ProtoMessageErr represents a message that was not published successfully to a messaging system.
type ProtoMessageErr interface {
	ProtoMessage
	Error() error
}
