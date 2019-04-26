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
	"strings"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
)

type watchEvent struct {
	Type      datasync.Op
	Key       string
	Value     []byte
	PrevValue []byte
	Revision  int64
}

type watchers map[chan string]*watcher // close channel -> watcher

type watcher struct {
	prefixes    []watchPrefix
	watchCh     chan *watchEvent
	closeCh     chan string
	prefixRegCh chan watchPrefix // for registration of new key prefixes to watch
}

type watchResp struct {
	typ              datasync.Op
	key              string
	value, prevValue []byte
	rev              int64
}

type watchPrefix struct {
	prefix string
	cb     watchCallback
}

type watchCallback func(watchResp keyval.BytesWatchResp)

func (c *Client) bumpWatchers(we *watchEvent) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cfg.FilterDupNotifs && bytes.Equal(we.Value, we.PrevValue) {
		return
	}
	for _, w := range c.watchers {
		w.watchCh <- we
	}
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

// Watch watches given list of key prefixes.
func (c *Client) Watch(resp func(watchResp keyval.BytesWatchResp), closeCh chan string, prefixes ...string) error {
	boltLogger.Debugf("watch: %q", prefixes)

	c.mu.Lock()
	defer c.mu.Unlock()

	w, exists := c.watchers[closeCh]
	if exists {
		// this close channel is already in use
		for _, prefix := range prefixes {
			w.prefixRegCh <- watchPrefix{
				prefix: prefix,
				cb:     resp,
			}
		}
		return nil
	}

	// create and register new watcher
	w = &watcher{
		prefixes:    make([]watchPrefix, len(prefixes)),
		watchCh:     make(chan *watchEvent, 10),
		closeCh:     closeCh,
		prefixRegCh: make(chan watchPrefix, 10),
	}
	for i, prefix := range prefixes {
		w.prefixes[i] = watchPrefix{
			prefix: prefix,
			cb:     resp,
		}
	}
	c.watchers[closeCh] = w

	go func() {
		w.watch()
		// un-register when done
		c.mu.Lock()
		delete(c.watchers, closeCh)
		c.mu.Unlock()
	}()
	return nil
}

func (w *watcher) watch() {
	for {
		select {
		case ev, ok := <-w.watchCh:
			if !ok {
				boltLogger.WithField("prefixes", w.prefixes).
					Debug("Watch channel was closed")
				return
			}
			var cb watchCallback
			for _, wp := range w.prefixes {
				if strings.HasPrefix(ev.Key, wp.prefix) {
					cb = wp.cb
					break
				}
			}
			if cb == nil {
				// key not watched by this watcher
				continue
			}
			r := &watchResp{
				typ:       ev.Type,
				key:       ev.Key,
				value:     ev.Value,
				prevValue: ev.PrevValue,
				rev:       ev.Revision,
			}
			cb(r)

		case regPrefix, ok := <- w.prefixRegCh:
			if !ok {
				boltLogger.WithField("prefixes", w.prefixes).
					Debug("Prefix-registration channel was closed")
				return
			}
			w.prefixes = append(w.prefixes, regPrefix)
			boltLogger.WithField("prefixes", w.prefixes).Debug(
				"The set of watched prefixes was extended")

		case closeVal, ok := <-w.closeCh:
			if !ok {
				boltLogger.WithField("prefixes", w.prefixes).Debug("Watch ended")
				return
			}
			for i, wp := range w.prefixes {
				if wp.prefix == closeVal {
					boltLogger.WithField("prefix", wp.prefix).Debug("Watch ended")
					w.prefixes[i] = w.prefixes[len(w.prefixes)-1]
					w.prefixes = w.prefixes[:len(w.prefixes)-1]
					break
				}
			}
		}
	}
}