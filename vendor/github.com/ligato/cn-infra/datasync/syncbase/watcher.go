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
	"sync"

	"errors"
	"strings"
	"time"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/logroot"
)

// NewRegistry creates reusable registry of subscriptions for a particular datasync plugin.
func NewRegistry() *Registry {
	return &Registry{subscriptions: map[string]*Subscription{}, access: sync.Mutex{}, lastRev: NewLatestRev()}
}

// Registry of subscriptions and latest revisions.
// This structure contains extracted reusable code among various datasync implementations.
// By having this code datasync plugins do not need to repeat code related management of subscriptions.
type Registry struct {
	subscriptions map[string]*Subscription
	access        sync.Mutex
	lastRev       *PrevRevisions
}

// WatchDataReg implements interface datasync.WatchDataRegistration
type WatchDataReg struct {
	ResyncName string
	adapter    *Registry
	CloseChan  chan interface{}
}

// Close stops watching of particular KeyPrefixes.
func (reg *WatchDataReg) Close() error {
	reg.adapter.access.Lock()
	defer reg.adapter.access.Unlock()

	delete(reg.adapter.subscriptions, reg.ResyncName)

	reg.CloseChan <- nil

	return nil
}

// Subscription TODO
type Subscription struct {
	ResyncName  string
	ChangeChan  chan datasync.ChangeEvent
	ResyncChan  chan datasync.ResyncEvent
	KeyPrefixes []string
}

// WatchDataBase just appends channels
func (adapter *Registry) WatchDataBase(resyncName string, changeChan chan datasync.ChangeEvent,
	resyncChan chan datasync.ResyncEvent, keyPrefixes ...string) (*WatchDataReg, error) {

	adapter.access.Lock()
	defer adapter.access.Unlock()

	if _, found := adapter.subscriptions[resyncName]; found {
		return nil, errors.New("Already watching " + resyncName)
	}

	reg := &WatchDataReg{ResyncName: resyncName, adapter: adapter, CloseChan: make(chan interface{}, 1)}
	adapter.subscriptions[resyncName] = &Subscription{
		resyncName, changeChan,
		resyncChan, keyPrefixes,
	}

	return reg, nil
}

// Watch just appends channels
func (adapter *Registry) Watch(resyncName string, changeChan chan datasync.ChangeEvent,
	resyncChan chan datasync.ResyncEvent, keyPrefixes ...string) (datasync.WatchRegistration, error) {
	return adapter.WatchDataBase(resyncName, changeChan, resyncChan, keyPrefixes...)
}

// Subscriptions returns the current subscriptions.
func (adapter *Registry) Subscriptions() map[string]*Subscription {
	return adapter.subscriptions
}

// LastRev is just getter
func (adapter *Registry) LastRev() *PrevRevisions {
	return adapter.lastRev
}

// PropagateChanges fills registered channels with the data
func (adapter *Registry) PropagateChanges(txData map[string] /*key*/ datasync.ChangeValue) error {
	events := []func(done chan error){}

	for _, sub := range adapter.subscriptions {
		for _, prefix := range sub.KeyPrefixes {
			for key, val := range txData {
				if strings.HasPrefix(key, prefix) {
					var prev datasync.LazyValueWithRev
					var curRev int64
					if datasync.Delete == val.GetChangeType() {
						_, prev = adapter.lastRev.Del(key)
						if prev != nil {
							curRev = prev.GetRevision() + 1
						}
					} else {
						_, prev, curRev = adapter.lastRev.Put(key, val)
					}

					events = append(events,
						func(sub *Subscription, key string, val datasync.ChangeValue) func(done chan error) {
							return func(done chan error) {
								sub.ChangeChan <- &ChangeEvent{key, val.GetChangeType(),
									val, curRev, prev, NewDoneChannel(done)}
							}
						}(sub, key, val))
				}
			}
		}
	}

	done := make(chan error, 1)
	go AggregateDone(events, done)

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-time.After(5 * time.Second):
		logroot.StandardLogger().Warn("Timeout of aggregated change callback")
	}

	return nil
}

// PropagateResync fills registered channels with the data
func (adapter *Registry) PropagateResync(txData map[ /*key*/ string]datasync.ChangeValue) error {
	for _, sub := range adapter.subscriptions {
		resyncEv := NewResyncEventDB(map[string] /*keyPrefix*/ datasync.KeyValIterator{})
		for _, prefix := range sub.KeyPrefixes {
			kvs := []datasync.KeyVal{}
			for key, val := range txData {
				if strings.HasPrefix(key, prefix) {
					adapter.lastRev.PutWithRevision(key, val)
					kvs = append(kvs, &KeyVal{key, val, val.GetRevision()})
				}
			}
			resyncEv.its[prefix] = NewKVIterator(kvs)
		}
		sub.ResyncChan <- resyncEv //TODO default and/or timeout
	}

	return nil
}
