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
	"sync"

	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/protobuf/proto"
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
	modelsByGoType    map[reflect.Type]*knownModel
	modelsByProtoName map[string]*knownModel
	modelsByName      map[string]*knownModel
}

// NewRegistry returns initialized Registry.
func NewRegistry() *LocalRegistry {
	return &LocalRegistry{
		modelsByGoType:    make(map[reflect.Type]*knownModel),
		modelsByProtoName: make(map[string]*knownModel),
		modelsByName:      make(map[string]*knownModel),
	}
}

// GetModel returns registered model for the given model name
// or error if model is not found.
func (r *LocalRegistry) GetModel(name string) (KnownModel, error) {
	model, found := r.modelsByName[name]
	if !found {
		return &knownModel{}, fmt.Errorf("no model registered for name %v", name)
	}
	return model, nil
}

// GetModelFor returns registered model for the given proto message.
func (r *LocalRegistry) GetModelFor(x any) (KnownModel, error) {
	if len(r.modelsByGoType) > len(r.modelsByProtoName) {
		r.lazyInitRegisteredTypesByProtoName()
	}
	msg, ok := x.(proto.Message)
	if !ok {
		return &knownModel{}, fmt.Errorf("can't get model: %v is not a proto message", x)
	}
	msgName := string(msg.ProtoReflect().Descriptor().FullName())
	model, found := r.modelsByProtoName[msgName]
	if !found {
		return &knownModel{}, fmt.Errorf("no model registered for proto message %s", msgName)
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
	for _, model := range r.modelsByGoType {
		r.modelsByProtoName[model.ProtoName()] = model // ProtoName() == ProtoReflect().Descriptor().FullName()
	}
}

// GetModelForKey returns registered model for the given key or error.
func (r *LocalRegistry) GetModelForKey(key string) (KnownModel, error) {
	for _, model := range r.modelsByProtoName {
		if model.IsKeyValid(key) {
			return model, nil
		}
	}
	return &knownModel{}, fmt.Errorf("no registered model matches for key %v", key)
}

// RegisteredModels returns all registered models.
func (r *LocalRegistry) RegisteredModels() []KnownModel {
	if len(r.modelsByGoType) > len(r.modelsByProtoName) {
		r.lazyInitRegisteredTypesByProtoName()
	}
	var models []KnownModel
	for _, model := range r.modelsByProtoName {
		models = append(models, model)
	}
	return models
}

// MessageTypeRegistry creates new message type registry from registered proto messages
func (r *LocalRegistry) MessageTypeRegistry() *protoregistry.Types {
	typeRegistry := new(protoregistry.Types)
	for _, model := range r.modelsByName {
		err := typeRegistry.RegisterMessage(dynamicpb.NewMessageType(model.pb.ProtoReflect().Descriptor()))
		if err != nil {
			logging.Warn("registering message %v for local registry failed: %v", model, err)
		}
	}
	return typeRegistry
}

// Register registers either a protobuf message known at compile-time together with the given model specification,
// or a remote model represented by an instance of ModelInfo obtained via KnownModels RPC from MetaService.
// While the former case is prevalent, the latter option is useful for scenarios with multiple agents and configuration
// requests being proxied from one to another (remote model registered into LocalRegistry may act as a proxy for the
// agent from which it was learned).
// If spec.Class is unset then it defaults to 'config'.
func (r *LocalRegistry) RegisterMsg(msg proto.Message) (KnownModel, error) {
	var goType reflect.Type
	// if the message is not dynamic store the Go type information
	if _, ok := msg.(*dynamicpb.Message); !ok {
		goType = reflect.TypeOf(msg)
		// Check go type duplicate registration
		if m, ok := r.modelsByGoType[goType]; ok {
			return nil, fmt.Errorf("go type %v already registered for model %v", goType, m.Name())
		}
	}

	desc := msg.ProtoReflect().Descriptor()
	protoName := string(desc.FullName())
	spec, err := SpecFromProtoDesc(desc)
	if err != nil {
		return nil, fmt.Errorf("model registration failed: %w", err)
	}
	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("spec validation for %s failed: %v", goType, err)
	}

	// Check model name collisions
	if model, ok := r.modelsByProtoName[protoName]; ok {
		return nil, fmt.Errorf("proto name %s already used by model %s", protoName, model.Name())
	}
	if model, ok := r.modelsByName[spec.ModelName()]; ok {
		return nil, fmt.Errorf("model name %q already used by model %s", spec.ModelName(), model.Name())
	}

	model := &knownModel{
		pb:     msg,
		spec:   spec,
		goType: goType,
	}

	// Apply custom options
	for _, opt := range optsFromProtoDesc(desc) {
		opt(&model.modelOptions)
	}

	r.modelsByGoType[goType] = model
	r.modelsByProtoName[string(desc.FullName())] = model
	r.modelsByName[model.Name()] = model

	if debugRegister {
		fmt.Printf("- model %s registered: %+v\n", model.Name(), model)
	}
	return model, nil
}

func (r *LocalRegistry) Register(x interface{}, spec Spec, opts ...ModelOption) (KnownModel, error) {
	msg, ok := x.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't register a non-proto message model")
	}
	goType := reflect.TypeOf(x)
	// Check go type duplicate registration
	if m, ok := r.modelsByGoType[goType]; ok {
		return nil, fmt.Errorf("go type %v already registered for model %v", goType, m.Name())
	}

	s := spec.Normalize()
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("spec validation for %s failed: %v", goType, err)
	}

	// Check model name collisions
	// if model, ok := r.modelsByProtoName[protoName]; ok {
	// 	return nil, fmt.Errorf("proto name %s already used by model %s", protoName, model.Name())
	// }
	if model, ok := r.modelsByName[s.ModelName()]; ok {
		return nil, fmt.Errorf("model name %q already used by model %s", spec.ModelName(), model.Name())
	}

	model := &knownModel{
		spec:         s,
		goType:       goType,
		pb:           msg,
		modelOptions: defaultOptions(x),
	}

	// Apply custom options
	for _, opt := range opts {
		opt(&model.modelOptions)
	}

	r.modelsByGoType[goType] = model
	// r.modelsByProtoName[string(model.pb.ProtoReflect().Descriptor().FullName())] = model
	r.modelsByName[model.Name()] = model

	if debugRegister {
		fmt.Printf("- model %s registered: %+v\n", model.Name(), model)
	}
	return model, nil
}

type SourceBroadcast[T any] struct {
	*Broadcast[T]
	S chan T
}

func NewSourceBroadcast[T any]() *SourceBroadcast[T] {
	s := make(chan T)
	return &SourceBroadcast[T]{
		S:         s,
		Broadcast: NewBroadcast(s),
	}
}

type Broadcast[T any] struct {
	mu          sync.RWMutex
	source      <-chan T
	subscribers []chan<- T
}

func NewBroadcast[T any](source <-chan T) *Broadcast[T] {
	b := &Broadcast[T]{
		source: source,
	}
	go b.serve()
	return b
}

func (b *Broadcast[T]) Subscribe() <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan T, 1)
	b.subscribers = append(b.subscribers, ch)
	return ch
}

func (b *Broadcast[T]) serve() {
	for val := range b.source {
		b.broadcast(val)
	}
	b.close()
}

func (b *Broadcast[T]) broadcast(val T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, sub := range b.subscribers {
		sub <- val
	}
}

func (b *Broadcast[T]) close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, sub := range b.subscribers {
		close(sub)
	}
}
