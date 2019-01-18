//  Copyright (c) 2018 Cisco and/or its affiliates.
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
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

var (
	// UpdatesChannelSize defines size for the channel
	// used for sending updates through single writer
	UpdatesChannelSize = 100

	// DefaultSafeUpdateTimeout is used as default timeout
	// for waiting on result from updater.
	DefaultSafeUpdateTimeout = time.Second
)

const minimumTimeout = time.Second / 100

type updateTx struct {
	updates []*update
	done    chan *result
}

type update struct {
	key   []byte
	value []byte
}

type result struct {
	prevValue []byte
	err       error
}

func (c *Client) safeUpdate(updates ...*update) (prevVal []byte, err error) {
	tx := &updateTx{
		updates: updates,
		done:    make(chan *result, 1),
	}
	timeoutDur := DefaultSafeUpdateTimeout
	if timeoutDur < minimumTimeout {
		timeoutDur = minimumTimeout
	}
	// send update to channel
	select {
	case c.updateChan <- tx:
		// update sent ok
	case <-time.After(timeoutDur):
		return nil, errors.New("bolt: updates full")
	}
	// wait for result
	select {
	case r := <-tx.done:
		if r == nil {
			return nil, errors.New("bolt: update failed")
		}
		return r.prevValue, r.err
	case <-time.After(timeoutDur):
		return nil, errors.New("bolt: update timeout")
	}
}

func (c *Client) startUpdater() {
	defer c.wg.Done()
	for {
		select {
		case utx := <-c.updateChan:
			r := &result{}
			r.err = c.db.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket(rootBucket)
				if len(utx.updates) == 1 {
					u := utx.updates[0]
					prev := bucket.Get(u.key)
					if prev != nil {
						r.prevValue = append([]byte(nil), prev...) // value needs to be copied
					} else if u.value == nil {
						return fmt.Errorf("bolt: key %q does not exist", u.key)
					}
				}
				for _, u := range utx.updates {
					var err error
					if u.value == nil {
						err = bucket.Delete(u.key)
					} else {
						err = bucket.Put(u.key, u.value)
					}
					if err != nil {
						return err
					}
				}
				return nil
			})
			utx.done <- r
		case <-c.quit:
			return
		}
	}
}
