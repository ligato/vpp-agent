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

package redis

import (
	"bytes"
	"errors"
	"strings"

	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/ligato/cn-infra/db"
	"github.com/ligato/cn-infra/db/keyval"
)

const keySpaceEventPrefix = "__keyspace@*__:"

// BytesWatchPutResp is sent when new key-value pair has been inserted or the value is updated
type BytesWatchPutResp struct {
	key   string
	value []byte
	rev   int64 // TODO Does Redis data have revision?
}

// NewBytesWatchPutResp creates an instance of BytesWatchPutResp
func NewBytesWatchPutResp(key string, value []byte, revision int64) *BytesWatchPutResp {
	return &BytesWatchPutResp{key: key, value: value, rev: revision}
}

// GetChangeType returns "Put" for BytesWatchPutResp
func (resp *BytesWatchPutResp) GetChangeType() db.PutDel {
	return db.Put
}

// GetKey returns the key that has been inserted
func (resp *BytesWatchPutResp) GetKey() string {
	return resp.key
}

// GetValue returns the value that has been inserted
func (resp *BytesWatchPutResp) GetValue() []byte {
	return resp.value
}

// GetRevision returns the revision associated with create action
func (resp *BytesWatchPutResp) GetRevision() int64 {
	return resp.rev
}

// BytesWatchDelResp is sent when a key-value pair has been removed
type BytesWatchDelResp struct {
	key string
	rev int64 // TODO Does Redis data have revision?
}

// NewBytesWatchDelResp creates an instance of BytesWatchDelResp
func NewBytesWatchDelResp(key string, revision int64) *BytesWatchDelResp {
	return &BytesWatchDelResp{key: key, rev: revision}
}

// GetChangeType returns "Delete" for BytesWatchPutResp
func (resp *BytesWatchDelResp) GetChangeType() db.PutDel {
	return db.Delete
}

// GetKey returns the key that has been deleted
func (resp *BytesWatchDelResp) GetKey() string {
	return resp.key
}

// GetValue returns nil for BytesWatchDelResp
func (resp *BytesWatchDelResp) GetValue() []byte {
	return nil
}

// GetRevision returns the revision associated with the delete operation
func (resp *BytesWatchDelResp) GetRevision() int64 {
	return resp.rev
}

// Watch starts subscription for changes associated with the selected key. Watch events will be delivered to respChan.
// Subscription can be canceled by StopWatch call.
func (db *BytesConnectionRedis) Watch(respChan chan keyval.BytesWatchResp, keys ...string) error {
	return db.watch(respChan, db.closeCh, nil, keys...)
}

func (db *BytesConnectionRedis) watch(respChan chan<- keyval.BytesWatchResp,
	closeChan <-chan struct{}, trimPrefix func(key string) string, keys ...string) error {
	if db.closed {
		return fmt.Errorf("watch(%v) called on a closed broker", keys)
	}
	db.Debugf("watch(%v)", keys)
	var buf bytes.Buffer
	for _, k := range keys {
		err := db.watchPattern(respChan, k, db.closeCh, trimPrefix)
		if err != nil {
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(err.Error())
		}
	}
	if buf.Len() > 0 {
		return errors.New(buf.String())
	}
	return nil
}

func (db *BytesConnectionRedis) watchPattern(respChan chan<- keyval.BytesWatchResp, key string,
	closeChan <-chan struct{}, trimPrefix func(key string) string) error {
	pattern := keySpaceEventPrefix + wildcard(key)
	db.Debugf("PSubscribe %s\n", pattern)

	// Allocate 1 connection per watch...
	conn := db.pool.Get()
	pubSub := redis.PubSubConn{Conn: conn}
	err := pubSub.PSubscribe(pattern)
	if err != nil {
		pubSub.Close()
		db.Errorf("PSubscribe %s failed: %s", pattern, err)
		return err
	}
	go func() {
		defer func() { db.Debugf("Watcher on %s exited", pattern) }()
		for {
			val := pubSub.Receive()
			closing, err := db.handleChange(val, respChan, closeChan, trimPrefix)
			if err != nil && !db.closed {
				db.Error(err)
			}
			if closing {
				return
			}
		}
	}()
	go func() {
		_, active := <-closeChan
		if !active {
			db.Debugf("Received signal to close watcher on %s", pattern)
			err := pubSub.PUnsubscribe(pattern)
			if err != nil {
				db.Errorf("PUnsubscribe %s failed: %s", pattern, err)
			}
			pubSub.Close()
		}
	}()

	return nil
}

func (db *BytesConnectionRedis) handleChange(val interface{}, respChan chan<- keyval.BytesWatchResp,
	closeChan <-chan struct{}, trimPrefix func(key string) string) (close bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			// In case something like this happens:
			// panic: send on closed channel
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	switch n := val.(type) {
	case redis.Subscription:
		db.Debugf("Subscription: %s %s %d", n.Kind, n.Channel, n.Count)
		if n.Count == 0 {
			return true, nil
		}
	case redis.PMessage:
		db.Debugf("PMessage: %s %s %s", n.Pattern, n.Channel, n.Data)
		key := strings.Split(n.Channel, ":")[1]
		switch cmd := string(n.Data); cmd {
		case "set":
			// Ouch, keyspace event does not convey value.  Need to retrieve it.
			val, _, rev, err := db.GetValue(key)
			if err != nil {
				db.Errorf("GetValue(%s) failed with error %s", key, err)
			}
			if val == nil {
				db.Errorf("GetValue(%s) returned nil", key)
			}
			if trimPrefix != nil {
				key = trimPrefix(key)
			}
			respChan <- NewBytesWatchPutResp(key, val, rev)
		case "del", "expired":
			if trimPrefix != nil {
				key = trimPrefix(key)
			}
			respChan <- NewBytesWatchDelResp(key, 0)
		}
		//TODO NICE-to-HAVE no block here if buffer is overflown
	case redis.Message:
		// Not subscribing to this event type yet
		db.Debugf("Message: %s %s which I did not subscribe !", n.Channel, n.Data)
	case error:
		return true, n
	}

	return false, nil
}

// Watch starts subscription for changes associated with the selected key. Watch events will be delivered to respChan.
func (pdb *BytesBrokerWatcherRedis) Watch(respChan chan keyval.BytesWatchResp, keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = pdb.prefix + k
	}
	return pdb.delegate.watch(respChan, pdb.closeCh, pdb.trimPrefix, prefixedKeys...)
}
