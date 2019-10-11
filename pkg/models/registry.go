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

package models

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
)

var (
	// DefaultRegistry represents a global registry for models.
	DefaultRegistry = NewRegistry()

	debugRegister = strings.Contains(os.Getenv("DEBUG_MODELS"), "register")
)

// Registry defines model registry for managing registered models.
type Registry struct {
	registeredTypes map[reflect.Type]*RegisteredModel
	modelNames      map[string]*RegisteredModel
}

// NewRegistry returns initialized Registry.
func NewRegistry() *Registry {
	return &Registry{
		registeredTypes: make(map[reflect.Type]*RegisteredModel),
		modelNames:      make(map[string]*RegisteredModel),
	}
}

// GetModel returns registered model for the given model path
// or error if model is not found.
func (r *Registry) GetModel(path string) (RegisteredModel, error) {
	model, found := r.modelNames[path]
	if !found {
		return RegisteredModel{}, fmt.Errorf("no model registered for path %v", path)
	}
	return *model, nil
}

// GetModelFor returns registered model for the given proto message.
func (r *Registry) GetModelFor(x proto.Message) (RegisteredModel, error) {
	t := reflect.TypeOf(x)
	model, found := r.registeredTypes[t]
	if !found {
		return RegisteredModel{}, fmt.Errorf("no model registered for type %v", t)
	}
	return *model, nil
}

// GetModelForKey returns registered model for the given key or error.
func (r *Registry) GetModelForKey(key string) (RegisteredModel, error) {
	for _, model := range r.registeredTypes {
		if model.IsKeyValid(key) {
			return *model, nil
		}
	}
	return RegisteredModel{}, fmt.Errorf("no registered model matches for key %v", key)
}

// RegisteredModels returns all registered modules.
func (r *Registry) RegisteredModels() []RegisteredModel {
	var models []RegisteredModel
	for _, model := range r.registeredTypes {
		models = append(models, *model)
	}
	return models
}

// Register registers a protobuf message with given model specification.
// If spec.Class is unset empty it defaults to 'config'.
func (r *Registry) Register(pb proto.Message, spec Spec, opts ...ModelOption) (*RegisteredModel, error) {
	goType := reflect.TypeOf(pb)

	// Check go type duplicate registration
	if m, ok := r.registeredTypes[goType]; ok {
		return nil, fmt.Errorf("go type %v already registered for model %v", goType, m.Name())
	}

	// Check model spec
	if spec.Class == "" {
		// spec with undefined class fallbacks to config
		spec.Class = "config"
	}
	if spec.Version == "" {
		spec.Version = "v0"
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("spec validation for %s failed: %v", goType, err)
	}

	// Check model name collisions
	if pn, ok := r.modelNames[spec.ModelName()]; ok {
		return nil, fmt.Errorf("model name %q already used by %s", spec.ModelName(), pn.goType)
	}

	model := &RegisteredModel{
		spec:      spec,
		goType:    goType,
		protoName: proto.MessageName(pb),
	}

	// Use GetName as fallback for generating name
	if _, ok := pb.(named); ok {
		model.nameFunc = func(obj interface{}) (s string, e error) {
			return obj.(named).GetName(), nil
		}
	}

	// Apply custom options
	for _, opt := range opts {
		opt(&model.modelOptions)
	}

	r.registeredTypes[goType] = model
	r.modelNames[model.Name()] = model

	if debugRegister {
		fmt.Printf("- model %s registered: %+v\n", model.Name(), model)
	}
	return model, nil
}
