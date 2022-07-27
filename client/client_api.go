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

// type ConfigItem struct {
// 	*generic.Item
// 	*generic.ItemStatus
// 	Value  proto.Message
// 	Labels map[string]string
// }

type UpdateItems struct {
	Msgs      []proto.Message
	Labels    map[string]string
	Overwrite bool
}

// Only set Ids or Labels (if both set )
type Filter struct {
	Ids    []generic.Item_ID
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

	// ResyncConfig overwrites existing config.
	ResyncConfig(items ...proto.Message) error

	// GetConfig retrieves current config into dsts.
	GetConfig(dsts ...interface{}) error

	// GetConfigItems returns current config as a list of config items
	GetConfigItems(filter Filter) ([]*generic.ConfigItem, error)

	UpdateConfigItems(items UpdateItems) (*generic.SetConfigResponse, error)

	DeleteConfigItems(items UpdateItems) (*generic.SetConfigResponse, error)

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
