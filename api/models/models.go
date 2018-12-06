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
	"net"
	"os"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/proto"
)

var debugRegister = strings.Contains(os.Getenv("DEBUG_MODELS"), "register")

// ProtoModel represents main model type.
type ProtoModel = proto.Message

// ID is a shorthand for the GetID for avoid error checking.
func ID(m ProtoModel) string {
	id, err := GetID(m)
	if err != nil {
		panic(err)
	}
	return id
}

// Key is a shorthand for the GetKey for avoid error checking.
func Key(m ProtoModel) string {
	key, err := GetKey(m)
	if err != nil {
		panic(err)
	}
	return key
}

// KeyPrefix is a shorthand for the GetKeyPrefix for avoid error checking.
func KeyPrefix(m ProtoModel) string {
	prefix, err := GetKeyPrefix(m)
	if err != nil {
		panic(err)
	}
	return prefix
}

func GetID(m ProtoModel) (string, error) {
	spec, err := GetSpec(m)
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
func GetKey(m ProtoModel) (string, error) {
	spec, err := GetSpec(m)
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
func GetKeyPrefix(m ProtoModel) (string, error) {
	spec, err := GetSpec(m)
	if err != nil {
		return "", err
	}
	var id strings.Builder
	if err := spec.idTmpl.Execute(&id, m); err != nil {
		panic(err)
	}
	return spec.KeyPrefix(), nil
}

// MustSpec returns registered model specification for given model.
func MustSpec(m ProtoModel) Spec {
	spec, err := GetSpec(m)
	if err != nil {
		panic(err)
	}
	return spec
}

// GetSpec returns registered model specification for given model.
func GetSpec(m ProtoModel) (Spec, error) {
	protoName := proto.MessageName(m)
	spec := registeredSpecs[protoName]
	if spec == nil {
		return Spec{}, fmt.Errorf("model %s is not registered", protoName)
	}
	return *spec, nil
}

// StripKeyPrefix returns key with prefix stripped.
func StripKeyPrefix(s string) string {
	for _, spec := range registeredSpecs {
		if trim := strings.TrimPrefix(s, spec.KeyPrefix()); trim != s {
			return trim
		}
	}
	return s
}

var (
	moduleSpecs     = make(map[string][]string)
	registeredSpecs = make(map[string]*Spec)
	keyPrefixes     = make(map[string]string)
)

// GetRegisteredSpecs returns all registered model specs.
func GetRegisteredSpecs() map[string]Spec {
	m := make(map[string]Spec)
	for k, v := range registeredSpecs {
		m[k] = *v
	}
	return m
}

// GetRegisteredModules returns all registered modules.
func GetRegisteredModules() (modules []*Module) {
	for moduleName, protos := range moduleSpecs {
		var specs []*ModelSpec
		for _, protoName := range protos {
			modelSpec := registeredSpecs[protoName].ToModelSpec()
			specs = append(specs, &modelSpec)
		}
		modules = append(modules, &Module{
			Name:  moduleName,
			Specs: specs,
		})
	}
	return
}

// Register registers given protobuf with model specification.
func Register(pb proto.Message, spec Spec) {
	protoName := proto.MessageName(pb)

	if _, ok := registeredSpecs[protoName]; ok {
		panic(fmt.Sprintf("duplicate model registered: %s", protoName))
	} else if !strings.HasPrefix(spec.Version, "v") {
		panic(fmt.Sprintf("version for model %s does not start with 'v': %q", protoName, spec.Version))
	} else if spec.Class != "config" && spec.Class != "status" {
		panic(fmt.Sprintf("class for model %s is invalid: %q", protoName, spec.Class))
	} else if len(spec.Type) == 0 {
		panic(fmt.Sprintf("kind for model %s is empty", protoName))
	} else if spec.IdTemplate == "" {
		panic(fmt.Sprintf("TmplID for model %s is empty", protoName))
	}

	spec.protoName = protoName
	spec.keyPrefix = spec.buildPrefix()
	if pn, ok := keyPrefixes[spec.keyPrefix]; ok {
		panic(fmt.Sprintf("key prefix %q already used for: %s", spec.keyPrefix, pn))
	}
	keyPrefixes[spec.keyPrefix] = protoName
	spec.idTmpl = template.Must(template.New("TmplID").Funcs(funcMap).Parse(spec.IdTemplate))

	if debugRegister {
		fmt.Printf("- registered model: %v\t%q\n", spec, spec.KeyPrefix())
	}
	registeredSpecs[protoName] = &spec
	moduleSpecs[spec.Module] = append(moduleSpecs[spec.Module], protoName)
}

var funcMap = template.FuncMap{
	"ipnet": func(s string) map[string]interface{} {
		_, ipNet, _ := net.ParseCIDR(s)
		maskSize, _ := ipNet.Mask.Size()
		return map[string]interface{}{
			"IP":       ipNet.IP.String(),
			"MaskSize": maskSize,
		}
	},
}
