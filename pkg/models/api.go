// Copyright (c) 2020 Pantheon.tech
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

package models

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

// ModelInfo represents model information retrieved using meta service
type ModelInfo struct {
	*generic.ModelDetail

	// MessageDescriptor is the proto message descriptor of the message represented by this ModelInfo struct
	MessageDescriptor protoreflect.MessageDescriptor
}

// Registry defines model registry for managing registered models
type Registry interface {
	// GetModel returns registered model for the given model name
	// or error if model is not found.
	GetModel(name string) (KnownModel, error)

	// GetModelFor returns registered model for the given proto message.
	GetModelFor(x interface{}) (KnownModel, error)

	// GetModelForKey returns registered model for the given key or error.
	GetModelForKey(key string) (KnownModel, error)

	// MessageTypeRegistry creates new message type registry from registered proto messages
	MessageTypeRegistry() *protoregistry.Types

	// RegisteredModels returns all registered modules.
	RegisteredModels() []KnownModel

	// Register registers either a protobuf message known at compile-time together
	// with the given model specification (for LocalRegistry),
	// or a remote model represented by an instance of ModelInfo obtained via KnownModels RPC from MetaService
	// (for RemoteRegistry or also for LocalRegistry but most likely just proxied to a remote agent).
	// If spec.Class is unset, then it defaults to 'config'.
	Register(x interface{}, spec Spec, opts ...ModelOption) (KnownModel, error)
}

// KnownModel represents a registered model
type KnownModel interface {
	// Spec returns model specification for the model.
	Spec() *Spec

	// ModelDetail returns descriptor for the model.
	ModelDetail() *generic.ModelDetail

	// NewInstance creates new instance value for model type.
	NewInstance() proto.Message

	// ProtoName returns proto message name registered with the model.
	ProtoName() string

	// ProtoFile returns proto file name for the model.
	ProtoFile() string

	// NameTemplate returns name template for the model.
	NameTemplate() string

	// GoType returns go type for the model.
	GoType() string

	// LocalGoType returns reflect go type for the model. The reflect type can be retrieved only
	// for locally registered model that provide locally known go types. The remotely retrieved model
	// can't provide reflect type so if known model information is retrieved remotely, this method
	// will return nil.
	LocalGoType() reflect.Type

	// PkgPath returns package import path for the model definition.
	PkgPath() string

	// Name returns name for the model.
	Name() string

	// KeyPrefix returns key prefix for the model.
	KeyPrefix() string

	// ParseKey parses the given key and returns item name
	// or returns empty name and valid as false if the key is not valid.
	ParseKey(key string) (name string, valid bool)

	// IsKeyValid returns true if given key is valid for this model.
	IsKeyValid(key string) bool

	// StripKeyPrefix returns key with prefix stripped.
	StripKeyPrefix(key string) string

	// InstanceName computes message name for given proto message using name template (if present).
	InstanceName(x interface{}) (string, error)
}
