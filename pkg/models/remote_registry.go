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
	"fmt"

	"github.com/go-errors/errors"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// RemoteRegistry defines model registry for managing registered remote models. The remote model have no
// included compiled code in program binary so only information available are from remote sources
// (i.e. generic.Client's known models)
type RemoteRegistry struct {
	modelByName map[string]*RemotelyKnownModel
}

// NewRemoteRegistry returns initialized RemoteRegistry.
func NewRemoteRegistry() *RemoteRegistry {
	return &RemoteRegistry{
		modelByName: make(map[string]*RemotelyKnownModel),
	}
}

// GetModel returns registered model for the given model name
// or error if model is not found.
func (r *RemoteRegistry) GetModel(name string) (KnownModel, error) {
	model, found := r.modelByName[name]
	if !found {
		return &RemotelyKnownModel{}, fmt.Errorf("no remote model registered for name %v", name)
	}
	return model, nil
}

// GetModelFor returns registered model for the given proto message.
func (r *RemoteRegistry) GetModelFor(x interface{}) (KnownModel, error) {
	messageDesc := protoMessageOf(x).ProtoReflect().Descriptor()
	messageFullName := string(messageDesc.FullName())
	var foundModel *RemotelyKnownModel
	for _, model := range r.modelByName {
		if model.ProtoName() == messageFullName {
			foundModel = model
			break
		}
	}
	if foundModel == nil {
		return nil, errors.Errorf("can't find remote model for message %v "+
			"(All remote models by model names: %#v)", messageFullName, r.modelByName)
	}
	return foundModel, nil
}

// GetModelForKey returns registered model for the given key or error.
func (r *RemoteRegistry) GetModelForKey(key string) (KnownModel, error) {
	for _, model := range r.modelByName {
		if model.IsKeyValid(key) {
			return model, nil
		}
	}
	return &RemotelyKnownModel{}, fmt.Errorf("no registered remote model matches for key %v", key)
}

// RegisteredModels returns all registered modules.
func (r *RemoteRegistry) RegisteredModels() []KnownModel {
	var models []KnownModel
	for _, model := range r.modelByName {
		models = append(models, model)
	}
	return models
}

// MessageTypeRegistry creates new message type registry from registered proto messages
func (r *RemoteRegistry) MessageTypeRegistry() *protoregistry.Types {
	typeRegistry := new(protoregistry.Types)
	for _, model := range r.modelByName {
		if err := typeRegistry.RegisterMessage(dynamicpb.NewMessageType(model.model.MessageDescriptor)); err != nil {
			logging.Warn("registering message %v for remote registry failed: %v", model, err)
		}
	}
	return typeRegistry
}

// Register registers remote model ModelInfo (given as interface{} for common register interface flexibility).
// The given spec and options are already in ModelInfo and therefore these input arguments are ignored.
func (r *RemoteRegistry) Register(model interface{}, spec Spec, opts ...ModelOption) (KnownModel, error) {
	if model == nil {
		return nil, fmt.Errorf("can't register nil model")
	}
	modelInfo, ok := model.(*ModelInfo)
	if !ok {
		return nil, fmt.Errorf("can't register model that is not *ModelInfo (input type %T)", model)
	}
	if modelInfo.MessageDescriptor == nil {
		return nil, fmt.Errorf("can't register model with nil message descriptor")
	}

	// Check model spec
	if modelInfo.Spec.Class == "" {
		// spec with undefined class fallbacks to config
		modelInfo.Spec.Class = "config"
	}
	if modelInfo.Spec.Version == "" {
		modelInfo.Spec.Version = "v0"
	}

	if err := ToSpec(modelInfo.Spec).Validate(); err != nil {
		return nil, fmt.Errorf("spec validation for %s failed: %v", modelInfo.ProtoName, err)
	}

	// Check model name collisions
	if pn, ok := r.modelByName[ToSpec(modelInfo.Spec).ModelName()]; ok {
		return nil, fmt.Errorf("model name %q already used by %s", spec.ModelName(), pn.ProtoName())
	}

	// create RemotelyKnownModel and register it
	remoteModel := &RemotelyKnownModel{
		model: modelInfo,
	}
	r.modelByName[ToSpec(remoteModel.model.Spec).ModelName()] = remoteModel

	if debugRegister {
		fmt.Printf("- remote model %s registered: %+v\n", remoteModel.Name(), remoteModel)
	}
	return remoteModel, nil
}
