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

package client

import (
	"context"

	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

// ModelInfo is just retyped models.ModelInfo for backward compatibility purpose
// Deprecated: use models.ModelInfo instead
type ModelInfo = models.ModelInfo

type StateItem = generic.StateItem
type ConfigItem = generic.ConfigItem

type UpdateItem struct {
	Message proto.Message
	Labels  map[string]string
}

// GetValue exists so that UpdateItem satisfies datasync.LazyValue interface.
func (item UpdateItem) GetValue(out proto.Message) error {
	if item.Message != nil {
		proto.Merge(out, item.Message)
	}
	return nil
}

func (item UpdateItem) GetLabels() map[string]string {
	return item.Labels
}

type UpdateResult struct {
	Key    string
	Status *generic.ItemStatus
}

// If (Ids|Labels) is nil that means no filtering for (Ids|Labels)
// But if both are not nil then an error is returned
// (because of ambiguity in what should the result be filtered by).
// If for a given label key the corresponding value is "" then items are
// only matched using the key.
type Filter struct {
	Ids    []*generic.Item_ID
	Labels map[string]string
}

// ConfigClient ...
// Deprecated: use GenericClient instead
type ConfigClient = GenericClient

// GenericClient is the client-side interface for generic handler.
type GenericClient interface {
	// KnownModels retrieves list of known modules.
	KnownModels(class string) ([]*ModelInfo, error)

	// ChangeRequest returns transaction for changing config.
	ChangeRequest() ChangeRequest

	// NewItemChange() return transaction with label support for changing config.
	NewItemChange() ItemChange

	// ResyncConfig overwrites existing config.
	ResyncConfig(items ...proto.Message) error

	// GetConfig retrieves current config into dsts.
	// TODO: return as list of config items
	GetConfig(dsts ...interface{}) error

	// GetFilteredConfig retrieves current config into dsts according to the provided filter.
	GetFilteredConfig(filter Filter, dsts ...interface{}) error

	// GetItems returns list of all current ConfigItems.
	GetItems(ctx context.Context) ([]*ConfigItem, error)

	UpdateItems(ctx context.Context, items []UpdateItem, resync bool) ([]*UpdateResult, error)

	DeleteItems(ctx context.Context, items []UpdateItem) ([]*UpdateResult, error)

	// DumpState dumps actual running state.
	DumpState() ([]*StateItem, error)
}

// ChangeRequest is interface for config change request.
type ChangeRequest interface {
	// Update appends updates for given items to the request.
	Update(items ...proto.Message) ChangeRequest

	// Delete appends deletes for given items to the request.
	Delete(items ...proto.Message) ChangeRequest

	// Send sends the request.
	Send(ctx context.Context) error
}

// ItemChange is interface for update item change request (so it supports item labels as well).
type ItemChange interface {
	// Update appends updates for given items to the batch change request.
	Update(items ...UpdateItem) ItemChange

	// Delete appends deletes for given items to the batch change request.
	Delete(items ...UpdateItem) ItemChange

	// Send sends the batch change request.
	Send(ctx context.Context) error
}
