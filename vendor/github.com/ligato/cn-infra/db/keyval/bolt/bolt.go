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
	"context"
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logrus"
)


var boltLogger = logrus.NewLogger("bolt")

func init() {
	if os.Getenv("DEBUG_BOLT_CLIENT") != "" {
		boltLogger.SetLevel(logging.DebugLevel)
	}
}

var rootBucket = []byte("root")

// Client serves as a client for Bolt KV storage and implements
// keyval.CoreBrokerWatcher interface.
type Client struct {
	db *bolt.DB

	cfg Config

	updateChan chan *updateTx

	mu       sync.RWMutex
	watchers watchers

	wg   sync.WaitGroup
	quit chan struct{}
}

// NewClient creates new client for Bolt using given config.
func NewClient(cfg *Config) (client *Client, err error) {
	db, err := bolt.Open(cfg.DbPath, cfg.FileMode, &bolt.Options{
		Timeout: cfg.LockTimeout,
	})
	if err != nil {
		return nil, err
	}
	boltLogger.Infof("bolt path: %v", db.Path())

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(rootBucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	c := &Client{
		db:         db,
		cfg:        *cfg,
		quit:       make(chan struct{}),
		updateChan: make(chan *updateTx, UpdatesChannelSize),
		watchers:   make(watchers),
	}

	c.wg.Add(1)
	go c.startUpdater()

	return c, nil
}

// Close closes Bolt database.
func (c *Client) Close() error {
	close(c.quit)
	c.wg.Wait()
	return c.db.Close()
}

// GetValue returns data for the given key
func (c *Client) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	boltLogger.Debugf("GetValue: %q", key)

	err = c.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(rootBucket).Get([]byte(key))
		if value == nil {
			return fmt.Errorf("value for key %q not found in bucket", key)
		}

		found = true
		data = append([]byte(nil), value...) // value needs to be copied

		return nil
	})
	return data, found, 0, err
}

// Put stores given data for the key
func (c *Client) Put(key string, data []byte, opts ...datasync.PutOption) (err error) {
	boltLogger.Debugf("Put: %q (len=%d)", key, len(data))

	prevVal, err := c.safeUpdate(&update{
		key:   []byte(key),
		value: data,
	})
	if err != nil {
		return err
	}

	c.bumpWatchers(&watchEvent{
		Key:       key,
		Value:     data,
		PrevValue: prevVal,
		Type:      datasync.Put,
	})

	return nil
}

// Delete deletes given key
func (c *Client) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	boltLogger.Debugf("Delete: %q", key)

	prevVal, err := c.safeUpdate(&update{
		key:   []byte(key),
		value: nil,
	})
	if err != nil {
		return false, err
	}
	existed = prevVal != nil

	c.bumpWatchers(&watchEvent{
		Key:       key,
		PrevValue: prevVal,
		Type:      datasync.Delete,
	})

	return existed, err
}

// ListKeys returns iterator with keys for given key prefix
func (c *Client) ListKeys(keyPrefix string) (keyval.BytesKeyIterator, error) {
	boltLogger.Debugf("ListKeys: %q", keyPrefix)

	var keys []string
	err := c.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(rootBucket).Cursor()
		prefix := []byte(keyPrefix)

		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			boltLogger.Debugf(" listing key: %q", string(k))
			keys = append(keys, string(k))
		}

		return nil
	})

	return &bytesKeyIterator{len: len(keys), keys: keys}, err
}

// ListValues returns iterator with key-value pairs for given key prefix
func (c *Client) ListValues(keyPrefix string) (keyval.BytesKeyValIterator, error) {
	boltLogger.Debugf("ListValues: %q", keyPrefix)

	var pairs []*kvPair
	err := c.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(rootBucket).Cursor()
		prefix := []byte(keyPrefix)

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			boltLogger.Debugf(" listing val: %q (len=%d)", string(k), len(v))

			pair := &kvPair{Key: string(k)}
			pair.Value = append([]byte(nil), v...) // value needs to be copied

			pairs = append(pairs, pair)
		}

		return nil
	})

	return &bytesKeyValIterator{len: len(pairs), pairs: pairs}, err
}

// NewTxn creates new transaction
func (c *Client) NewTxn() keyval.BytesTxn {
	return &txn{
		c: c,
	}
}

// Txn allows grouping operations into the transaction. Transaction executes
// multiple operations in a more efficient way in contrast to executing
// them one by one.
type txn struct {
	c       *Client
	updates []*update
}

// Put adds a new 'put' operation to a previously created transaction.
// If the <key> does not exist in the data store, a new key-value item
// will be added to the data store. If <key> exists in the data store,
// the existing value will be overwritten with the <value> from this
// operation.
func (t *txn) Put(key string, value []byte) keyval.BytesTxn {
	t.updates = append(t.updates, &update{
		key:   []byte(key),
		value: value,
	})
	return t
}

// Delete adds a new 'delete' operation to a previously created
// transaction. If <key> exists in the data store, the associated value
// will be removed.
func (t *txn) Delete(key string) keyval.BytesTxn {
	t.updates = append(t.updates, &update{
		key:   []byte(key),
		value: nil,
	})
	return t
}

// Commit commits all operations in a transaction to the data store.
// Commit is atomic - either all operations in the transaction are
// committed to the data store, or none of them.
func (t *txn) Commit(ctx context.Context) error {
	_, err := t.c.safeUpdate(t.updates...)
	return err
}
