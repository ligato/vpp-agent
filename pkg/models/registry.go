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

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

var (
	// DefaultRegistry represents a global registry for local models (models known in compile time)
	DefaultRegistry Registry = NewRegistry()

	debugRegister = strings.Contains(os.Getenv("DEBUG_MODELS"), "register")
)

// LocalRegistry defines model registry for managing registered local models. Local models are locally compiled into
// the program binary and hence some additional information in compare to remote models, i.e. go type.
type LocalRegistry struct {
	registeredModelsByGoType    map[reflect.Type]*LocallyKnownModel
	registeredModelsByProtoName map[string]*LocallyKnownModel
	modelNames                  map[string]*LocallyKnownModel
	ordered                     []reflect.Type
	proxied                     *RemoteRegistry
}

// NewRegistry returns initialized Registry.
func NewRegistry() *LocalRegistry {
	return &LocalRegistry{
		registeredModelsByGoType:    make(map[reflect.Type]*LocallyKnownModel),
		registeredModelsByProtoName: make(map[string]*LocallyKnownModel),
		modelNames:                  make(map[string]*LocallyKnownModel),
		proxied:                     NewRemoteRegistry(),
	}
}

// GetModel returns registered model for the given model name
// or error if model is not found.
func (r *LocalRegistry) GetModel(name string) (KnownModel, error) {
	model, found := r.modelNames[name]
	if !found {
		if model, err := r.proxied.GetModel(name); err == nil {
			return model, nil
		}
		return &LocallyKnownModel{}, fmt.Errorf("no model registered for name %v", name)
	}
	return model, nil
}

// GetModelFor returns registered model for the given proto message.
func (r *LocalRegistry) GetModelFor(x interface{}) (KnownModel, error) {
	// find model by Go type
	t := reflect.TypeOf(x)
	model, found := r.registeredModelsByGoType[t]
	if !found {
		// check remotely retrieved models registered in local registry
		if proxModel, err := r.proxied.GetModelFor(x); err == nil {
			return proxModel, nil
		}

		// find model by Proto name
		// (useful when using dynamically generated config instead of configurator.Config => go type of proto
		// messages is in such case always dynamicpb.Message and never the go type of registered (generated)
		// proto message)
		if len(r.registeredModelsByProtoName) == 0 && len(r.registeredModelsByGoType) > 0 {
			r.lazyInitRegisteredTypesByProtoName()
		}
		var protoName string
		if pb, ok := x.(protoreflect.ProtoMessage); ok {
			protoName = string(pb.ProtoReflect().Descriptor().FullName())
		} else if v1, ok := x.(proto.Message); ok {
			protoName = string(v1.ProtoReflect().Descriptor().FullName())
		}
		if protoName != "" {
			if model, found = r.registeredModelsByProtoName[protoName]; found {
				return model, nil
			}
		}

		// find model by checking proto options
		if model = r.checkProtoOptions(x); model == nil {
			return &LocallyKnownModel{}, fmt.Errorf("no model registered for type %v", t)
		}
	}
	return model, nil
}

// lazyInitRegisteredTypesByProtoName performs lazy initialization of registeredModelsByProtoName. The reason
// why initialization can't happen while registration (call of func Register(...)) is that some proto reflect
// functionality is not available during this time. The registration happens as variable initialization, but
// the reflection is initialized in init() func and that happens after variable initialization.
//
// Alternative solution would be to change when the models are registered (VPP-Agent have it like described
// above and 3rd party model are probably copying the same behaviour). So to not break anything, the lazy
// initialization seems like the best solution for now.
func (r *LocalRegistry) lazyInitRegisteredTypesByProtoName() {
	for _, model := range r.registeredModelsByGoType {
		r.registeredModelsByProtoName[model.ProtoName()] = model // ProtoName() == ProtoReflect().Descriptor().FullName()
	}
}

// GetModelForKey returns registered model for the given key or error.
func (r *LocalRegistry) GetModelForKey(key string) (KnownModel, error) {
	for _, model := range r.registeredModelsByGoType {
		if model.IsKeyValid(key) {
			return model, nil
		}
	}
	if model, err := r.proxied.GetModelForKey(key); err == nil {
		return model, nil
	}
	return &LocallyKnownModel{}, fmt.Errorf("no registered model matches for key %v", key)
}

// RegisteredModels returns all registered modules.
func (r *LocalRegistry) RegisteredModels() []KnownModel {
	var models []KnownModel
	for _, typ := range r.ordered {
		models = append(models, r.registeredModelsByGoType[typ])
	}
	models = append(models, r.proxied.RegisteredModels()...)
	return models
}

// MessageTypeRegistry creates new message type registry from registered proto messages
func (r *LocalRegistry) MessageTypeRegistry() *protoregistry.Types {
	typeRegistry := new(protoregistry.Types)
	for _, model := range r.modelNames {
		typeRegistry.RegisterMessage(dynamicpb.NewMessageType(model.proto.ProtoReflect().Descriptor()))
	}
	proxiedTypes := r.proxied.MessageTypeRegistry()
	proxiedTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
		typeRegistry.RegisterMessage(mt)
		return true
	})
	return typeRegistry
}

// Register registers either a protobuf message known at compile-time together with the given model specification,
// or a remote model represented by an instance of ModelInfo obtained via KnownModels RPC from MetaService.
// While the former case is prevalent, the latter option is useful for scenarios with multiple agents and configuration
// requests being proxied from one to another (remote model registered into LocalRegistry may act as a proxy for the
// agent from which it was learned).
// If spec.Class is unset then it defaults to 'config'.
func (r *LocalRegistry) Register(x interface{}, spec Spec, opts ...ModelOption) (KnownModel, error) {
	// check if the model was learned remotely
	if modelInfo, isProxied := x.(*ModelInfo); isProxied {
		// check for collision with local models
		mName := ToSpec(modelInfo.Spec).ModelName()
		if _, duplicate := r.modelNames[mName]; duplicate {
			return nil, fmt.Errorf("model %v is already known locally and cannot be proxied", mName)
		}
		return r.proxied.Register(x, spec, opts...)
	}

	goType := reflect.TypeOf(x)
	// Check go type duplicate registration
	if m, ok := r.registeredModelsByGoType[goType]; ok {
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
	if _, err := r.proxied.GetModel(spec.ModelName()); err == nil {
		return nil, fmt.Errorf("model name %q is already proxied", spec.ModelName())
	}

	model := &LocallyKnownModel{
		spec:   spec,
		goType: goType,
	}

	if pb, ok := x.(protoreflect.ProtoMessage); ok {
		model.proto = pb
	} else if v1, ok := x.(proto.Message); ok {
		model.proto = v1
	}

	// Use GetName as fallback for generating name
	if _, ok := x.(named); ok {
		model.nameFunc = func(obj interface{}) (s string, e error) {
			// handling dynamic messages (they don't implement named interface)
			if dynMessage, ok := obj.(*dynamicpb.Message); ok {
				obj, e = DynamicLocallyKnownMessageToGeneratedMessage(dynMessage)
				if e != nil {
					return "", e
				}
			}
			// handling other proto message
			return obj.(named).GetName(), nil
		}
		model.nameTemplate = namedTemplate
	}

	// Apply custom options
	for _, opt := range opts {
		opt(&model.modelOptions)
	}

	r.registeredModelsByGoType[goType] = model
	r.modelNames[model.Name()] = model
	r.ordered = append(r.ordered, goType)

	if debugRegister {
		fmt.Printf("- model %s registered: %+v\n", model.Name(), model)
	}
	return model, nil
}
