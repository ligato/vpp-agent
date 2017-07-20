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

	log "github.com/ligato/cn-infra/logging/logrus"
)

// Adapter implements datasync.TransportAdapter but allows optionally implement these WatchData / PublishData
type Adapter struct {
	Watcher   datasync.Watcher
	Publisher datasync.Publisher
}

// WatchData using Kafka Watcher Topic Watcher
func (adapter *Adapter) WatchData(resyncName string, changeChan chan datasync.ChangeEvent,
	resyncChan chan datasync.ResyncEvent, keyPrefixes ...string) (datasync.WatchDataRegistration, error) {

	if adapter.Watcher != nil {
		return adapter.Watcher.WatchData(resyncName, changeChan, resyncChan, keyPrefixes...)
	}
	log.Debug("Watcher is nil")

	return nil, nil
}

// PublishData using Kafka Watcher Topic Publisher
func (adapter *Adapter) PublishData(key string, data proto.Message) error {
	if adapter.Publisher != nil {
		return adapter.Publisher.PublishData(key, data)
	}

	return nil
}
