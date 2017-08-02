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

package dbsync

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/cn-infra/db/keyval"
)

// NewAdapter creates a new instance of Adapter.
func NewAdapter(name string, db keyval.BytesBroker, dbW keyval.BytesWatcher) *Adapter {
	return &Adapter{name, db, dbW, syncbase.NewWatcher()}
}

// Adapter is ETCD watcher/publisher in front of Agent Plugins
type Adapter struct {
	Name string
	db   keyval.BytesBroker
	dbW  keyval.BytesWatcher
	base *syncbase.Watcher
}

// WatchData using ETCD or any other default transport.
func (adapter *Adapter) WatchData(resyncName string, changeChan chan datasync.ChangeEvent,
	resyncChan chan datasync.ResyncEvent, keyPrefixes ...string) (datasync.WatchDataRegistration, error) {

	reg, err := adapter.base.WatchData(resyncName, changeChan, resyncChan, keyPrefixes...)
	if err != nil {
		return nil, err
	}

	_, err = watchAndResyncBrokerKeys(resyncName, changeChan, resyncChan, adapter, keyPrefixes...)
	if err != nil {
		return nil, err
	}

	return reg, err
}

// PublishData using ETCD or any other default transport
func (adapter *Adapter) PublishData(key string, data proto.Message) error {
	if data == nil {
		_, err := adapter.db.Delete(key)
		return err
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return adapter.db.Put(key, bytes)
}

// String return resyncName
func (adapter *Adapter) String() string {
	return adapter.Name
}
