//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package bolt

import (
	"bytes"
	"context"
	"strings"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
)

func (c *Client) bumpWatchers(we *watchEvent) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, w := range c.watchers {
		if strings.HasPrefix(we.Key, w.prefix) {
			w.watchCh <- we
		}
	}
}

type watchResp struct {
	typ              datasync.Op
	key              string
	value, prevValue []byte
	rev              int64
}

// GetChangeType returns "Put" for BytesWatchPutResp.
func (resp *watchResp) GetChangeType() datasync.Op {
	return resp.typ
}

// GetKey returns the key that the value has been inserted under.
func (resp *watchResp) GetKey() string {
	return resp.key
}

// GetValue returns the value that has been inserted.
func (resp *watchResp) GetValue() []byte {
	return resp.value
}

// GetPrevValue returns the previous value that has been inserted.
func (resp *watchResp) GetPrevValue() []byte {
	return resp.prevValue
}

// GetRevision returns the revision associated with the 'put' operation.
func (resp *watchResp) GetRevision() int64 {
	return resp.rev
}

func (c *Client) watch(resp func(watchResp keyval.BytesWatchResp), closeCh chan string, prefix string) error {
	boltLogger.Debug("watch:", prefix)

	ctx, cancel := context.WithCancel(context.Background())

	recvChan := c.watchPrefix(ctx, prefix)

	go func(regPrefix string) {
		defer cancel()
		for {
			select {
			case ev, ok := <-recvChan:
				if !ok {
					boltLogger.WithField("prefix", prefix).
						Debug("Watch recv chan was closed")
					return
				}
				if c.cfg.FilterDupNotifs && bytes.Equal(ev.Value, ev.PrevValue) {
					continue
				}
				r := &watchResp{
					typ:       ev.Type,
					key:       ev.Key,
					value:     ev.Value,
					prevValue: ev.PrevValue,
					rev:       ev.Revision,
				}
				resp(r)
			case closeVal, ok := <-closeCh:
				if !ok || closeVal == regPrefix {
					boltLogger.WithField("prefix", prefix).
						Debug("Watch ended")
					return
				}
			}
		}
	}(prefix)

	return nil
}

type watchEvent struct {
	Type      datasync.Op
	Key       string
	Value     []byte
	PrevValue []byte
	Revision  int64
}

type prefixWatcher struct {
	prefix  string
	watchCh chan *watchEvent
}

func (c *Client) watchPrefix(ctx context.Context, prefix string) <-chan *watchEvent {
	boltLogger.Debug("watchPrefix:", prefix)

	ch := make(chan *watchEvent, 1)

	c.mu.Lock()
	index := len(c.watchers)
	c.watchers = append(c.watchers, &prefixWatcher{
		prefix:  prefix,
		watchCh: ch,
	})
	c.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
			c.mu.Lock()
			if len(c.watchers) == index+1 {
				c.watchers = c.watchers[:index]
			} else {
				c.watchers = append(c.watchers[:index], c.watchers[index+1:]...)
			}
			c.mu.Unlock()
		}
	}()

	return ch
}
