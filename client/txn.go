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
	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/api/models"
)

// Transaction prepares request data.
type Transaction interface {
	// ProtoItem returns model with given ID from the request items.
	// If the found is true the model with such ID is found
	// and if the model is nil the item represents delete.
	Item(id string) (model proto.Message, found bool)

	// Items returns map of items defined for the request,
	// where key represents model ID and nil value represents delete.
	Items() map[string]proto.Message
}

type Txn struct {
	items map[string]proto.Message
}

func NewTxn() *Txn {
	return &Txn{
		items: make(map[string]proto.Message),
	}
}

func (t *Txn) Add(model proto.Message) {
	t.items[models.Key(model)] = model
}

func (t *Txn) Remove(model proto.Message) {
	delete(t.items, models.Key(model))
}

func (t *Txn) Set(model proto.Message) {
	t.items[models.Key(model)] = model
}

func (t *Txn) SetDelete(model proto.Message) {
	t.items[models.Key(model)] = nil
}

func (t *Txn) Item(id string) (model proto.Message, found bool) {
	item, ok := t.items[id]
	return item, ok
}

func (t *Txn) Items() map[string]proto.Message {
	return t.items
}
