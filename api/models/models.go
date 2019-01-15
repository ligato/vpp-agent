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
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
)

// ID is a shorthand for the GetID for avoid error checking.
func ID(m proto.Message) string {
	id, err := GetID(m)
	if err != nil {
		panic(err)
	}
	return id
}

// Key is a shorthand for the GetKey for avoid error checking.
func Key(m proto.Message) string {
	key, err := GetKey(m)
	if err != nil {
		panic(err)
	}
	return key
}

// KeyPrefix is a shorthand for the GetKeyPrefix for avoid error checking.
func KeyPrefix(m proto.Message) string {
	prefix, err := GetKeyPrefix(m)
	if err != nil {
		panic(err)
	}
	return prefix
}

// ModelSpec returns registered model specification for given item.
func ModelSpec(m proto.Message) Spec {
	spec, err := GetModelSpec(m)
	if err != nil {
		panic(err)
	}
	return spec
}

// GetID
func GetID(m proto.Message) (string, error) {
	spec, err := GetModelSpec(m)
	if err != nil {
		return "", err
	}
	var str strings.Builder
	if err := spec.idTmpl.Execute(&str, m); err != nil {
		return "", err
	}
	return str.String(), nil
}

// GetKey returns complete key for gived model,
// including key prefix defined by model specification.
// It returns error if given model is not registered.
func GetKey(m proto.Message) (string, error) {
	spec, err := GetModelSpec(m)
	if err != nil {
		return "", err
	}
	var id strings.Builder
	if err := spec.idTmpl.Execute(&id, m); err != nil {
		panic(err)
	}
	key := spec.KeyPrefix() + id.String()
	return key, nil
}

// GetKeyPrefix returns key prefix for gived model.
// It returns error if given model is not registered.
func GetKeyPrefix(m proto.Message) (string, error) {
	spec, err := GetModelSpec(m)
	if err != nil {
		return "", err
	}
	return spec.KeyPrefix(), nil
}

// GetModelSpec returns registered model specification for given model.
func GetModelSpec(m proto.Message) (Spec, error) {
	protoName := proto.MessageName(m)
	spec := registeredSpecs[protoName]
	if spec == nil {
		return Spec{}, fmt.Errorf("model %s is not registered", protoName)
	}
	return *spec, nil
}
