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
	"context"

	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/api/models"
)

// Local is global client for direct local access.
var Local = NewClient(&txnFactory{local.DefaultRegistry})

type ProtoTxnFactory interface {
	NewTxn(resync bool) keyval.ProtoTxn
}

type client struct {
	txnFactory ProtoTxnFactory
}

// NewClient returns new instance that uses given registry for data propagation.
func NewClient(factory ProtoTxnFactory) SyncClient {
	return &client{factory}
}

/*func (c *client) SyncRequest(ctx context.Context) SyncRequest {

	return &request{txn: c.txnFactory.NewTxn(true)}
}*/

// ResyncRequest returns new resync request.
func (c *client) ResyncRequest() SyncRequest {
	return &request{txn: c.txnFactory.NewTxn(true)}
}

// ChangeRequest return new change request.
func (c *client) ChangeRequest() SyncRequest {
	return &request{txn: c.txnFactory.NewTxn(false)}
}

type request struct {
	txn keyval.ProtoTxn
	err error
}

/*// Put adds the given model data to the transaction.
func (r *request) Put(items ...models.ProtoModel) {
	r.Update(items...)
}*/

// Update adds update for the given model data to the transaction.
func (r *request) Update(items ...models.ProtoModel) {
	if r.err != nil {
		return
	}
	for _, item := range items {
		key, err := models.GetKey(item)
		if err != nil {
			r.err = err
			return
		}
		r.txn.Put(key, item)
	}
}

// Delete adds delete for the given model keys to the transaction.
func (r *request) Delete(items ...models.ProtoModel) {
	if r.err != nil {
		return
	}
	for _, item := range items {
		key, err := models.GetKey(item)
		if err != nil {
			r.err = err
			return
		}
		r.txn.Delete(key)
	}
}

// Send commits the transaction with all data.
func (r *request) Send(ctx context.Context) error {
	if r.err != nil {
		return r.err
	}
	return r.txn.Commit()
}

type txnFactory struct {
	registry *syncbase.Registry
}

func (p *txnFactory) NewTxn(resync bool) keyval.ProtoTxn {
	if resync {
		return local.NewProtoTxn(p.registry.PropagateResync)
	} else {
		return local.NewProtoTxn(p.registry.PropagateChanges)
	}
}
