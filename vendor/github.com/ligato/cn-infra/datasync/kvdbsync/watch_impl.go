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

package kvdbsync

import (
	"time"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging/logroot"
)

// WatchBrokerKeys implements go routines on top of Change & Resync channels.
type watchBrokerKeys struct {
	resyncReg  resync.Registration
	changeChan chan datasync.ChangeEvent
	resyncChan chan datasync.ResyncEvent
	prefixes   []string
	adapter    *watcher
}

type watcher struct {
	db   keyval.ProtoBroker
	dbW  keyval.ProtoWatcher
	base *syncbase.Registry
}

// WatchAndResyncBrokerKeys calls keyval watcher Watch() & resync Register().
// This creates go routines for each tuple changeChan + resyncChan.
func watchAndResyncBrokerKeys(resyncReg resync.Registration, changeChan chan datasync.ChangeEvent, resyncChan chan datasync.ResyncEvent,
	closeChan chan string, adapter *watcher, keyPrefixes ...string) (keys *watchBrokerKeys, err error) {
	keys = &watchBrokerKeys{
		resyncReg:  resyncReg,
		changeChan: changeChan,
		resyncChan: resyncChan,
		adapter:    adapter,
		prefixes:   keyPrefixes}

	if resyncReg != nil {
		go keys.watchResync(resyncReg)
	}
	if changeChan != nil {
		err = keys.adapter.dbW.Watch(keys.watchChanges, closeChan, keys.prefixes...)
	}
	return keys, err
}

func (keys *watchBrokerKeys) watchChanges(x keyval.ProtoWatchResp) {
	var prev datasync.LazyValue
	if datasync.Delete == x.GetChangeType() {
		_, prev = keys.adapter.base.LastRev().Del(x.GetKey())
	} else {
		_, prev = keys.adapter.base.LastRev().PutWithRevision(x.GetKey(),
			syncbase.NewKeyVal(x.GetKey(), x, x.GetRevision()))
	}

	ch := NewChangeWatchResp(x, prev)

	logroot.StandardLogger().Debug("dbAdapter x:", x)
	logroot.StandardLogger().Debug("dbAdapter ch:", *ch)

	keys.changeChan <- ch
	// TODO NICE-to-HAVE publish the err using the transport asynchronously
}

// resyncReg.StatusChan == Started => resync
func (keys *watchBrokerKeys) watchResync(resyncReg resync.Registration) {
	for resyncStatus := range resyncReg.StatusChan() {
		if resyncStatus.ResyncStatus() == resync.Started {
			err := keys.resync()
			if err != nil {
				logroot.StandardLogger().Error("error getting resync data ", err) // We are not able to propagate it somewhere else.
				// TODO NICE-to-HAVE publish the err using the transport asynchronously
			}
		}
		resyncStatus.Ack()
	}
}

// Resync fills the resyncChan with the most recent snapshot (db.ListValues).
func (keys *watchBrokerKeys) resync() error {
	its := map[string] /*keyPrefix*/ datasync.KeyValIterator{}
	for _, keyPrefix := range keys.prefixes {
		it, err := keys.adapter.db.ListValues(keyPrefix)
		if err != nil {
			return err
		}
		its[keyPrefix] = NewIterator(it)
	}

	resyncEvent := syncbase.NewResyncEventDB(its)
	keys.resyncChan <- resyncEvent

	select {
	case err := <-resyncEvent.DoneChan:
		if err != nil {
			return err
		}
	case <-time.After(4 * time.Second):
		logroot.StandardLogger().Warn("Timeout of resync callback")
	}
	return nil
}

// String returns resyncName.
func (keys *watchBrokerKeys) String() string {
	return keys.resyncReg.String()
}
