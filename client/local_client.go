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

package client

import (
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/vpp-agent/api/models"
)

// Local is global client for direct local access.
var Local = NewClient(local.DefaultRegistry)

type client struct {
	registry *syncbase.Registry
}

// NewClient returns new instance that uses given registry for data propagation.
func NewClient(registry *syncbase.Registry) SyncClient {
	return &client{registry}
}

// ResyncRequest returns new resync request.
func (c *client) ResyncRequest() ResyncRequest {
	return &request{local.NewProtoTxn(c.registry.PropagateResync)}
}

// ChangeRequest return new change request.
func (c *client) ChangeRequest() ChangeRequest {
	return &request{local.NewProtoTxn(c.registry.PropagateChanges)}
}

type request struct {
	txn *local.ProtoTxn
}

// Update adds update for the given model data to the transaction.
func (r *request) Update(models ...models.ProtoModel) {
	for _, m := range models {
		r.txn.Put(m.ModelKey(), m)
	}
}

// Delete adds delete for the given model keys to the transaction.
func (r *request) Delete(keys ...string) {
	for _, key := range keys {
		r.txn.Delete(key)
	}
}

// Send commits the transaction with all data.
func (r *request) Send() error {
	return r.txn.Commit()
}
