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
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/api/models"
)

// Local is global client for direct local access.
var Local = NewClient(&txnFactory{local.DefaultRegistry})

type client struct {
	txnFactory ProtoTxnFactory
}

// NewClient returns new instance that uses given registry for data propagation.
func NewClient(factory ProtoTxnFactory) ConfiguratorClient {
	return &client{factory}
}

func (c *client) ListModules() (map[string][]models.Model, error) {
	modules := make(map[string][]models.Model)
	for _, model := range models.RegisteredModels() {
		modules[model.Module] = append(modules[model.Module], *model)
	}
	return modules, nil
}

func (c *client) SetConfig(resync bool) SetConfigRequest {
	return &setConfigRequest{txn: c.txnFactory.NewTxn(resync)}
}

type setConfigRequest struct {
	txn keyval.ProtoTxn
	err error
}

func (r *setConfigRequest) Update(items ...models.ProtoItem) {
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

func (r *setConfigRequest) Delete(items ...models.ProtoItem) {
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

func (r *setConfigRequest) Send(ctx context.Context) error {
	if r.err != nil {
		return r.err
	}
	return r.txn.Commit()
}
