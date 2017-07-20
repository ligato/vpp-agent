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
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
)

// BytesPublisher allows to publish a message of type []bytes into messaging system.
type BytesPublisher interface {
	Publish(key string, data []byte) error
}

// ProtoPublisher allows to publish a message of type proto.Message into messaging system.
type ProtoPublisher interface {
	Publish(key string, data proto.Message) error
}

// BytesMessage defines functions for inspection of a message received from messaging system.
type BytesMessage interface {
	keyval.BytesKvPair
}

// ProtoMessage defines functions for inspection of a message receive from messaging system.
type ProtoMessage interface {
	keyval.ProtoKvPair
}
