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

package models

import (
	"path"

	"github.com/golang/protobuf/proto"
)

// Register registers model in DefaultRegistry.
func Register(pb proto.Message, spec Spec, opts ...ModelOption) *KnownModel {
	model, err := DefaultRegistry.Register(pb, spec, opts...)
	if err != nil {
		panic(err)
	}
	return model
}

// RegisteredModels returns models registered in the DefaultRegistry.
func RegisteredModels() []KnownModel {
	return DefaultRegistry.RegisteredModels()
}

// GetModel returns registered model for given model name.
func GetModel(name string) (KnownModel, error) {
	return DefaultRegistry.GetModel(name)
}

// GetModelFor returns model registered in DefaultRegistry for given proto message.
func GetModelFor(x proto.Message) (KnownModel, error) {
	return DefaultRegistry.GetModelFor(x)
}

// GetModelForKey returns model registered in DefaultRegistry which matches key.
func GetModelForKey(key string) (KnownModel, error) {
	return DefaultRegistry.GetModelForKey(key)
}

// Key is a helper for the GetKey which panics on errors.
func Key(x proto.Message) string {
	key, err := GetKey(x)
	if err != nil {
		panic(err)
	}
	return key
}

// Name is a helper for the GetName which panics on errors.
func Name(x proto.Message) string {
	name, err := GetName(x)
	if err != nil {
		panic(err)
	}
	return name
}

// GetKey returns complete key for gived model,
// including key prefix defined by model specification.
// It returns error if given model is not registered.
func GetKey(x proto.Message) (string, error) {
	model, err := GetModelFor(x)
	if err != nil {
		return "", err
	}
	name, err := model.instanceName(x)
	if err != nil {
		return "", err
	}
	key := path.Join(model.KeyPrefix(), name)
	return key, nil
}

// GetName returns instance name for given model.
// It returns error if given model is not registered.
func GetName(x proto.Message) (string, error) {
	model, err := GetModelFor(x)
	if err != nil {
		return "", err
	}
	name, err := model.instanceName(x)
	if err != nil {
		return "", err
	}
	return name, nil
}
