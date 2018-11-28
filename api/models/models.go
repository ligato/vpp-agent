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

type ProtoModel = proto.Message

// ProtoModel represents proto.Message that returns model key.
/*type ProtoModel interface {
	proto.Message
	ModelID() string
}*/

// Spec represents model specification for registering models.
type Spec struct {
	Module  string
	Class   string
	Version string
	Kind    string
	TmplID  string

	protoName string
	keyPrefix string
	idTmpl    *template.Template
}

func (s Spec) ParseKey(key string) (id string, valid bool) {
	trim := strings.TrimPrefix(key, s.KeyPrefix())
	if trim != key && trim != "" {
		return trim, true
	}
	return "", false
}

func (s Spec) IsKeyValid(key string) bool {
	trim := strings.TrimPrefix(key, s.keyPrefix)
	if trim != key && trim != "" {
		// TODO: validate name?
		return true
	}
	return false
}

// StripKeyPrefix returns key with prefix stripped.
func (s Spec) StripKeyPrefix(key string) string {
	trim := strings.TrimPrefix(key, s.KeyPrefix())
	if trim != key && trim != "" {
		return trim
	}
	return key
}

// KeyPrefix returns key prefix used for storing model in KV stores.
func (s Spec) KeyPrefix() string {
	if s.keyPrefix == "" {
		s.keyPrefix = s.prepareKeyPrefix()
	}
	return s.keyPrefix
}

// ProtoName returns proto message name of the model.
func (s Spec) ProtoName() string {
	return s.protoName
}

var prefixTmpl = template.Must(template.New("keyPrefix").Parse(
	`{{.Module}}/{{.Class}}/{{.Version}}/{{.Kind}}/`,
))

func (s Spec) prepareKeyPrefix() string {
	var str strings.Builder
	if err := prefixTmpl.Execute(&str, s); err != nil {
		panic(err)
	}
	return str.String()
}

func (s Spec) ModelID(m ProtoModel) string {
	protoName := proto.MessageName(m)
	if protoName != s.protoName {
		return "<wrong-model>"
	}
	var str strings.Builder
	if err := s.idTmpl.Execute(&str, m); err != nil {
		return "<template-failed>"
	}
	return str.String()
}

/*func (s Spec) RawID(m map[string]string) string {
	var str strings.Builder
	if err := s.idTmpl.Execute(&str, m); err != nil {
		return "<template-failed>"
	}
	return str.String()
}*/

/*func ParseKey(key string) map[string]string {
	parts := map[string]string{}
	return parts
}*/

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

// KeyPrefix is a shorthand for the GetKey for avoid error checking.
func KeyPrefix(m ProtoModel) string {
	key, err := GetKey(m)
	if err != nil {
		panic(err)
	}
	return key
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
	return GetProtoSpec(proto.MessageName(m))
}

func GetProtoSpec(protoName string) (Spec, error) {
	spec := registeredSpecs[protoName]
	if spec == nil {
		return Spec{}, fmt.Errorf("model %s is not registered", protoName)
	}
	return *spec, nil
}

/*func ProtoKey(protoName string, m map[string]string) string {
	spec, err := GetProtoSpec(protoName)
	if err != nil {
		return ""
	}

	var id strings.Builder
	if err := spec.idTmpl.Execute(&id, m); err != nil {
		panic(err)
	}

	return spec.KeyPrefix() + id.String()
}*/

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
	registeredSpecs = make(map[string]*Spec)
	keyPrefixes     = make(map[string]string)
)

// GetRegistered returns all registered model specs.
func GetRegistered() map[string]Spec {
	m := make(map[string]Spec)
	for k, v := range registeredSpecs {
		m[k] = *v
	}
	return m
}

// Register registers given protobuf with model specification.
func Register(pb proto.Message, spec Spec) {
	if spec.Class == "" {
		spec.Class = "config"
	}
	protoName := proto.MessageName(pb)
	if _, ok := registeredSpecs[protoName]; ok {
		panic(fmt.Sprintf("duplicate model registered: %s", protoName))
	} else if !strings.HasPrefix(spec.Version, "v") {
		panic(fmt.Sprintf("version for model %s does not start with 'v': %q", protoName, spec.Version))
	} else if spec.Class != "config" && spec.Class != "status" {
		panic(fmt.Sprintf("class for model %s is invalid: %q", protoName, spec.Class))
	} else if len(spec.Kind) == 0 {
		panic(fmt.Sprintf("kind for model %s is empty", protoName))
	} else if spec.TmplID == "" {
		panic(fmt.Sprintf("TmplID for model %s is empty", protoName))
	}
	spec.protoName = protoName
	spec.keyPrefix = spec.prepareKeyPrefix()
	if pn, ok := keyPrefixes[spec.keyPrefix]; ok {
		panic(fmt.Sprintf("key prefix %q already used for: %s", spec.keyPrefix, pn))
	}
	keyPrefixes[spec.keyPrefix] = protoName
	spec.idTmpl = template.Must(template.New("TmplID").Funcs(funcMap).Parse(spec.TmplID))
	if debugRegister {
		fmt.Printf("- registered model %q: %+v\n", protoName, spec)
	}
	registeredSpecs[protoName] = &spec
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
