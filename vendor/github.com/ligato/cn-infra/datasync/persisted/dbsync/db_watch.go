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
	"github.com/ligato/cn-infra/db"
	"github.com/ligato/cn-infra/db/keyval"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/resync"
	resync_types "github.com/ligato/cn-infra/datasync/resync/resyncevent"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"time"
)

// WatchBrokerKeys implements go routines on top of Change & Resync channels
type watchBrokerKeys struct {
	resyncName string
	changeChan chan datasync.ChangeEvent
	resyncChan chan datasync.ResyncEvent
	prefixes   []string

	adapater *Adapter
}

// WatchAndResyncBrokerKeys calls etcdmux.Watch() & resync.Register()
// This creates go routines for each tuple changeChan+resyncChan.
func watchAndResyncBrokerKeys(resyncName string, changeChan chan datasync.ChangeEvent, resyncChan chan datasync.ResyncEvent,
	adapter *Adapter, keyPrefixes ...string) (*watchBrokerKeys, error) {

	var wasError error
	watchCh := make(chan keyval.BytesWatchResp)
	resyncReg := resync.Register(resyncName)

	keys := &watchBrokerKeys{
		resyncName: resyncName,
		changeChan: changeChan,
		resyncChan: resyncChan,
		adapater:   adapter,
		prefixes:   keyPrefixes}

	if resyncReg != nil {
		go keys.watchResync(resyncReg)
	}
	if changeChan != nil {
		go keys.watchChanges(watchCh)

		err := adapter.dbW.Watch(watchCh, keys.prefixes...)
		if err != nil {
			wasError = err
		}

	}
	return keys, wasError
}

func (keys *watchBrokerKeys) watchChanges(watchCh chan keyval.BytesWatchResp) {
	for x := range watchCh {
		var prev datasync.LazyValue
		if db.Delete == x.GetChangeType() {
			_, prev = keys.adapater.base.LastRev().Del(x.GetKey())
		} else {
			_, prev = keys.adapater.base.LastRev().PutWithRevision(x.GetKey(),
				syncbase.NewKeyValBytes(x.GetKey(), x.GetValue(), x.GetRevision()))
		}

		ch := NewChangeWatchResp(x, prev)

		log.Debug("dbAdapter x:", x)
		log.Debug("dbAdapter ch:", *ch)

		keys.changeChan <- ch
		// TODO NICE-to-HAVE publish the err using the transport asynchronously
	}
}

// resyncReg.StatusChan == Started => resync
func (keys *watchBrokerKeys) watchResync(resyncReg resync_types.Registration) {
	for resyncStatus := range resyncReg.StatusChan() {
		if resyncStatus.ResyncStatus() == resync_types.Started {
			err := keys.resync()
			if err != nil {
				log.Error("error getting resync data ", err) //we are not able to propagate it somewhere else
				// TODO NICE-to-HAVE publish the err using the transport asynchronously
			}
		}
		resyncStatus.Ack()
	}
}

// Resync fills the resyncChan with most recent snapshot (db.ListValues)
func (keys *watchBrokerKeys) resync() error {
	its := map[string] /*keyPrefix*/ datasync.KeyValIterator{}
	for _, keyPrefix := range keys.prefixes {
		it, err := keys.adapater.db.ListValues(keyPrefix)
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
		log.Warn("Timeout of resync callback")
	}

	return nil
}

// String returns resyncName
func (keys *watchBrokerKeys) String() string {
	return keys.resyncName
}
