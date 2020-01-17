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

	"github.com/golang/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

type ModelInfo = generic.ModelDetail

type StateItem = generic.StateItem

// ConfigClient defines the client-side interface for config.
type ConfigClient interface {
	// KnownModels retrieves list of known modules.
	KnownModels(class string) ([]*ModelInfo, error)

	// ChangeRequest returns transaction for changing config.
	ChangeRequest() ChangeRequest

	// ResyncConfig overwrites existing config.
	ResyncConfig(items ...proto.Message) error

	// GetConfig retrieves current config into dsts.
	// TODO: return as list of config items
	GetConfig(dsts ...interface{}) error

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
