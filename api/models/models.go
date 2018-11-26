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

// ProtoModel represents proto.Message that returns model key.
type ProtoModel interface {
	proto.Message
	ModelID() string
}

// Spec represents model specification for registering models.
type Spec struct {
	Version   string
	Module    string
	Class     string
	Kind      string
	protoName string
}

// KeyPrefix returns key prefix used for storing model in KV stores.
func (s Spec) KeyPrefix() string {
	return fmt.Sprintf("%s/%s/%s/%s/", s.Module, s.Class, s.Version, s.Kind)
}

// Key is a shorthand for the GetKey for avoid error checking.
func Key(m ProtoModel) string {
	key, _ := GetKey(m)
	return key
}

// GetKey returns complete key for gived model,
// including key prefix defined by model specification.
// It returns error if given model is not registered.
func GetKey(m ProtoModel) (string, error) {
	protoName := proto.MessageName(m)
	spec := registeredModels[protoName]
	if spec == nil {
		return "", fmt.Errorf("model %s is not registered", protoName)
	}
	key := spec.KeyPrefix() + m.ModelID()
	return key, nil
}

// GetModelSpec returns registered model specification for given model.
func GetModelSpec(m ProtoModel) (*Spec, error) {
	protoName := proto.MessageName(m)
	spec := registeredModels[protoName]
	if spec == nil {
		return nil, fmt.Errorf("model %s is not registered", protoName)
	}
	return spec, nil
}

var registeredModels = make(map[string]*Spec)

// Register registers given protobuf with model specification.
func Register(pb proto.Message, spec Spec, fn ...interface{}) {
	protoName := proto.MessageName(pb)
	if _, ok := registeredModels[protoName]; ok {
		panic(fmt.Sprintf("duplicate model registered: %s", protoName))
	} else if !strings.HasPrefix(spec.Version, "v") {
		panic(fmt.Sprintf("version for model %s does not start with 'v': %q", protoName, spec.Version))
	} else if spec.Class != "config" && spec.Class != "status" {
		panic(fmt.Sprintf("class for model %s is invalid: %q", protoName, spec.Class))
	} else if len(spec.Kind) == 0 {
		panic(fmt.Sprintf("kind for model %s is empty", protoName))
		//} else if spec.IDfunc == nil {
		//	panic(fmt.Sprintf("IDFunc for model %s is undefined", protoName))
	}
	registeredModels[protoName] = &spec
	fmt.Printf("- registered model %q: %+v\n", protoName, registeredModels[protoName])
}
