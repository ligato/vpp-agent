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
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/ligato/cn-infra/db/keyval"
)

type op struct {
	value []byte
	del   bool
}

// Txn allows to group operations into the transaction. Transaction executes multiple operations
// in a more efficient way in contrast to executing them one by one.
type Txn struct {
	pool   *redis.Pool
	ops    map[string] /*key*/ *op
	prefix string
}

func (tx *Txn) addPrefix(key string) string {
	return tx.prefix + key
}

// Put adds a new 'put' operation to a previously created transaction.
// If the key does not exist in the data store, a new key-value item
// will be added to the data store. If key exists in the data store,
// the existing value will be overwritten with the value from this
// operation.
func (tx *Txn) Put(key string, value []byte) keyval.BytesTxn {
	tx.ops[tx.addPrefix(key)] = &op{value, false}
	return tx
}

// Delete adds a new 'delete' operation to a previously created
// transaction.
func (tx *Txn) Delete(key string) keyval.BytesTxn {
	tx.ops[tx.addPrefix(key)] = &op{nil, true}
	return tx
}

// Commit commits all operations in a transaction to the data store.
// Commit is atomic - either all operations in the transaction are
// committed to the data store, or none of them.
// TODO: use Redis MULTI transactional command instead?
func (tx *Txn) Commit() (err error) {
	toBeDeleted := []interface{}{}
	msetArgs := []interface{}{}
	for key, op := range tx.ops {
		if op.del {
			toBeDeleted = append(toBeDeleted, key)
		} else {
			msetArgs = append(msetArgs, key)
			msetArgs = append(msetArgs, string(op.value))
		}
	}

	conn := tx.pool.Get()
	defer conn.Close()

	if len(toBeDeleted) > 0 {
		_, err = conn.Do("DEL", toBeDeleted...)
		if err != nil {
			return fmt.Errorf("Do(DEL) failed: %s", err)
		}
	}
	if len(msetArgs) > 0 {
		_, err = conn.Do("MSET", msetArgs...)
		if err != nil {
			return fmt.Errorf("Do(MSET) failed: %s", err)
		}
	}
	return nil
}
