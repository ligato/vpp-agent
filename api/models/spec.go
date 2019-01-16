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
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/proto"
)

const (
	// ConfigKeyPrefix is used as key prefix for config storage.
	ConfigKeyPrefix = `config/`
)

var (
	validType   = regexp.MustCompile(`^[-a-z0-9_./]+$`)
	validModule = regexp.MustCompile(`^[-a-z0-9_.]+$`)
)

type NameFunc func(obj interface{}) (string, error)

// Spec represents model specification.
type Spec struct {
	Module       string
	Type         string
	Version      string
	NameTemplate string
	NameFunc     NameFunc
}

type registeredModel struct {
	Spec

	protoName string
	keyPrefix string
	modelPath string
}

// ProtoName returns proto message name registered with the model.
func (m registeredModel) ProtoName() string {
	return m.protoName
}

// Path returns path for the model.
func (m registeredModel) Path() string {
	return m.modelPath
}

// KeyPrefix returns key prefix for the model.
func (m registeredModel) KeyPrefix() string {
	return m.keyPrefix
}

// ParseKey parses the given key and returns item name
// or returns empty name and valid as false if the key is not valid.
func (m registeredModel) ParseKey(key string) (name string, valid bool) {
	name = strings.TrimPrefix(key, m.keyPrefix)
	if name != key && name != "" {
		// TODO: validate name?
		return name, true
	}
	return "", false
}

// IsKeyValid returns true if given key is valid for this model.
func (m registeredModel) IsKeyValid(key string) bool {
	_, valid := m.ParseKey(key)
	return valid
}

// StripKeyPrefix returns key with prefix stripped.
func (m registeredModel) StripKeyPrefix(key string) string {
	trim := strings.TrimPrefix(key, m.keyPrefix)
	if trim != key && trim != "" {
		return trim
	}
	return key
}

var (
	registeredModels = make(map[string]registeredModel)
	modelPaths       = make(map[string]string)

	debugRegister = strings.Contains(os.Getenv("DEBUG_MODELS"), "register")
)

// RegisterProto registers given protobuf with model specification.
func RegisterProto(pb proto.Message, spec Spec) registeredModel {
	model := registeredModel{
		Spec:      spec,
		protoName: proto.MessageName(pb),
	}

	if _, ok := registeredModels[model.protoName]; ok {
		panic(fmt.Sprintf("proto message %q already registered", model.protoName))
	}
	if !validModule.MatchString(spec.Module) {
		panic(fmt.Sprintf("module for model %s is invalid", model.protoName))
	}
	if !validType.MatchString(spec.Type) {
		panic(fmt.Sprintf("model type for %s is invalid", model.protoName))
	}
	if !strings.HasPrefix(spec.Version, "v") {
		panic(fmt.Sprintf("model version for %s is invalid", model.protoName))
	}

	if spec.NameFunc == nil {
		if _, ok := pb.(named); ok {
			model.NameFunc = func(obj interface{}) (s string, e error) {
				return obj.(named).GetName(), nil
			}
		} else if spec.NameTemplate == "" {
			panic(fmt.Sprintf("name func for model %s is nil", model.protoName))
		} else {
			model.NameFunc = NameTemplate(spec.NameTemplate)
		}
	}
	/*if spec.NameTemplate == "" {
		panic(fmt.Sprintf("name template for model %s is empty", model.protoName))
	}*/

	model.modelPath = path.Join(spec.Module, spec.Type)
	if pn, ok := modelPaths[model.modelPath]; ok {
		panic(fmt.Sprintf("path prefix %q already used by: %s", model.modelPath, pn))
	}

	/*model.keyPrefix = strings.Join([]string{
		spec.Version, spec.Module, spec.Type,
	}, "/")*/
	//model.keyPrefix = fmt.Sprintf("%s/%s/%s/", spec.Module, spec.Version, spec.Type)

	model.keyPrefix = fmt.Sprintf("%s/%s/", spec.Module, spec.Type)

	if debugRegister {
		fmt.Printf("- registered model: %-40v\t%q\n", spec, model.modelPath)
	}

	registeredModels[model.protoName] = model
	modelPaths[model.modelPath] = model.protoName

	return model
}

type named interface {
	GetName() string
}

func NameTemplate(t string) NameFunc {
	tmpl := template.Must(
		template.New("name").Funcs(funcMap).Option("missingkey=error").Parse(t),
	)
	return func(obj interface{}) (string, error) {
		var s strings.Builder
		if err := tmpl.Execute(&s, obj); err != nil {
			return "", err
		}
		return s.String(), nil
	}
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
