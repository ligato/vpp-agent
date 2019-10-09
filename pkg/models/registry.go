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

	api "github.com/ligato/vpp-agent/api/genericmanager"
)

var (
	// DefaultRegistry represents a global registry for models.
	DefaultRegistry = NewRegistry()

	debugRegister = strings.Contains(os.Getenv("DEBUG_MODELS"), "register")
)

// Registry defines model registry for managing registered models.
type Registry struct {
	name             string
	registeredModels map[reflect.Type]*Model
	modelPaths       map[string]*Model
}

// NewRegistry returns initialized Registry.
func NewRegistry() *Registry {
	return &Registry{
		registeredModels: make(map[reflect.Type]*Model),
		modelPaths:       make(map[string]*Model),
	}
}

// GetModel returns registered model for the given model path
// or error if model is not found.
func (r *Registry) GetModel(path string) (Model, error) {
	model, found := r.modelPaths[path]
	if !found {
		return Model{}, fmt.Errorf("no model registered for path %v", path)
	}
	return *model, nil
}

// GetModelFor returns registered model for the given proto message.
func (r *Registry) GetModelFor(x proto.Message) (Model, error) {
	t := reflect.TypeOf(x)
	model, found := r.registeredModels[t]
	if !found {
		return Model{}, fmt.Errorf("no model registered for type %v", t)
	}
	return *model, nil
}

// GetModelForKey returns registered model for the given key or error.
func (r *Registry) GetModelForKey(key string) (Model, error) {
	for _, model := range r.registeredModels {
		if model.IsKeyValid(key) {
			return *model, nil
		}
	}
	return Model{}, fmt.Errorf("no registered model matches for key %v", key)
}

// RegisteredModels returns all registered modules.
func (r *Registry) RegisteredModels() []*api.ModelInfo {
	var models []*api.ModelInfo
	for _, s := range r.registeredModels {
		models = append(models, &api.ModelInfo{
			Model: &api.Model{
				Module:  s.Module,
				Type:    s.Type,
				Version: s.Version,
			},
			Info: map[string]string{
				"nameTemplate": s.nameTemplate,
				"protoName":    s.ProtoName(),
				"modelPath":    s.Path(),
				"keyPrefix":    s.KeyPrefix(),
			},
		})
	}
	return models
}

// Register registers a protobuf message with given model specification.
func (r *Registry) Register(pb proto.Message, spec Spec, opts ...ModelOption) (*Model, error) {
	t := reflect.TypeOf(pb)

	model := &Model{
		modelSpec: modelSpec{spec},
		goType:    t,
		protoName: proto.MessageName(pb),
	}

	// Check duplicate registration
	if _, ok := r.registeredModels[t]; ok {
		return nil, fmt.Errorf("proto message %q already registered", model.protoName)
	}

	// Check proto message name
	if model.protoName == "" {
		// We do not want to panic anymore, because model might be registered in same package.
		//panic(fmt.Sprintf("empty proto message name for type: %T\n\n\tPlease ensure your .proto file contains: 'option (gogoproto.messagename_all) = true'", pb))
		//fmt.Printf("empty proto message name for type: %T\n\n\tPlease ensure your .proto file contains: 'option (gogoproto.messagename_all) = true'", pb)
	}

	// Validate model spec
	if !validModule.MatchString(spec.Module) {
		return nil, fmt.Errorf("module for model %s is invalid", model.protoName)
	}
	if !validType.MatchString(spec.Type) {
		return nil, fmt.Errorf("model type for %s is invalid", model.protoName)
	}
	if !strings.HasPrefix(spec.Version, "v") {
		return nil, fmt.Errorf("model version for %s is invalid", model.protoName)
	}

	// Generate keys & paths
	model.modelPath = buildModelPath(spec.Version, spec.Module, spec.Type)
	if pn, ok := r.modelPaths[model.modelPath]; ok {
		return nil, fmt.Errorf("path prefix %q already used by: %s", model.modelPath, pn.modelPath)
	}

	modulePath := strings.Replace(spec.Module, ".", "/", -1)
	model.keyPrefix = fmt.Sprintf("config/%s/%s/%s/", modulePath, spec.Version, spec.Type)

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

	if model.nameFunc == nil {
		model.keyPrefix = strings.TrimSuffix(model.keyPrefix, "/")
	}

	r.registeredModels[t] = model
	r.modelPaths[model.modelPath] = model

	if debugRegister {
		fmt.Printf("- registered model: %+v\t%q\n", model, model.modelPath)
	}

	return model, nil
}
