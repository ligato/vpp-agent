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

	"github.com/boltdb/bolt"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
)

// BrokerWatcher uses Client to access the datastore.
// The connection can be shared among multiple BrokerWatcher.
// In case of accessing a particular subtree in Bolt only,
// BrokerWatcher allows defining a keyPrefix that is prepended
// to all keys in its methods in order to shorten keys used in arguments.
type BrokerWatcher struct {
	*Client
	prefix string
}

// NewBroker creates a new instance of a proxy that provides
// access to Bolt. The proxy will reuse the connection from Client.
// <prefix> will be prepended to the key argument in all calls from the created
// BrokerWatcher. To avoid using a prefix, pass keyval. Root constant as
// an argument.
func (c *Client) NewBroker(prefix string) keyval.BytesBroker {
	return &BrokerWatcher{
		Client: c,
		prefix: prefix,
	}
}

// NewWatcher creates a new instance of a proxy that provides
// access to Bolt. The proxy will reuse the connection from Client.
// <prefix> will be prepended to the key argument in all calls on created
// BrokerWatcher. To avoid using a prefix, pass keyval. Root constant as
// an argument.
func (c *Client) NewWatcher(prefix string) keyval.BytesWatcher {
	return &BrokerWatcher{
		Client: c,
		prefix: prefix,
	}
}

func (pdb *BrokerWatcher) prefixKey(key string) string {
	return pdb.prefix + key
}

// Put calls 'Put' function of the underlying Client.
// KeyPrefix defined in constructor is prepended to the key argument.
func (pdb *BrokerWatcher) Put(key string, data []byte, opts ...datasync.PutOption) error {
	return pdb.Client.Put(pdb.prefixKey(key), data, opts...)
}

// NewTxn creates a new transaction.
// KeyPrefix defined in constructor will be prepended to all key arguments
// in the transaction.
func (pdb *BrokerWatcher) NewTxn() keyval.BytesTxn {
	return pdb.Client.NewTxn()
}

// GetValue calls 'GetValue' function of the underlying Client.
// KeyPrefix defined in constructor is prepended to the key argument.
func (pdb *BrokerWatcher) GetValue(key string) (data []byte, found bool, revision int64, err error) {
	return pdb.Client.GetValue(pdb.prefixKey(key))
}

// Delete calls 'Delete' function of the underlying Client.
// KeyPrefix defined in constructor is prepended to the key argument.
func (pdb *BrokerWatcher) Delete(key string, opts ...datasync.DelOption) (existed bool, err error) {
	return pdb.Client.Delete(pdb.prefixKey(key), opts...)
}

// ListKeys calls 'ListKeys' function of the underlying Client.
// KeyPrefix defined in constructor is prepended to the argument.
func (pdb *BrokerWatcher) ListKeys(keyPrefix string) (keyval.BytesKeyIterator, error) {
	boltLogger.Debugf("ListKeys: %q [namespace=%s]", keyPrefix, pdb.prefix)

	var keys []string
	err := pdb.Client.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(rootBucket).Cursor()
		prefix := []byte(pdb.prefixKey(keyPrefix))
		boltLogger.Debugf("listing keys: %q", string(prefix))

		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			boltLogger.Debugf(" listing key: %q", string(k))
			keys = append(keys, string(k))
		}
		return nil
	})

	return &bytesKeyIterator{prefix: pdb.prefix, len: len(keys), keys: keys}, err
}

// ListValues calls 'ListValues' function of the underlying Client.
// KeyPrefix defined in constructor is prepended to the key argument.
// The prefix is removed from the keys of the returned values.
func (pdb *BrokerWatcher) ListValues(keyPrefix string) (keyval.BytesKeyValIterator, error) {
	boltLogger.Debugf("ListValues: %q [namespace=%s]", keyPrefix, pdb.prefix)

	var pairs []*kvPair
	err := pdb.Client.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(rootBucket).Cursor()
		prefix := []byte(pdb.prefixKey(keyPrefix))
		boltLogger.Debugf("listing vals: %q", string(prefix))

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			boltLogger.Debugf(" listing val: %q (len=%d)", string(k), len(v))

			pair := &kvPair{Key: string(k)}
			pair.Value = append([]byte(nil), v...) // value needs to be copied

			pairs = append(pairs, pair)
		}
		return nil
	})

	return &bytesKeyValIterator{prefix: pdb.prefix, pairs: pairs, len: len(pairs)}, err
}

// Watch starts subscription for changes associated with the selected <keys>.
// KeyPrefix defined in constructor is prepended to all <keys> in the argument
// list. The prefix is removed from the keys returned in watch events.
// Watch events will be delivered to <resp> callback.
func (pdb *BrokerWatcher) Watch(resp func(keyval.BytesWatchResp), closeChan chan string, keys ...string) error {
	var prefixedKeys []string
	for _, key := range keys {
		prefixedKeys = append(prefixedKeys, pdb.prefixKey(key))
	}
	return pdb.Client.Watch(func(origResp keyval.BytesWatchResp) {
		r := origResp.(*watchResp)
		r.key = strings.TrimPrefix(r.key, pdb.prefix)
		resp(r)
	}, closeChan, prefixedKeys...)
}
