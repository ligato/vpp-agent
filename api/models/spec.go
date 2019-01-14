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
	"text/template"

	"github.com/ligato/vpp-agent/api"
)

type Model = api.ModelSpec
type Item = api.Item

/*type Module = api.Module
type ModelSpec = api.ModelSpec*/

func (s Spec) ToModelSpec() Model {
	ref := strings.ToLower(s.protoName)
	ref = strings.Replace(ref, ".", "/", -1)

	return Model{
		Name:    s.Type,
		Version: s.Version,
		Module:  s.Module,
		Meta: map[string]string{
			"id-tmpl":    s.IdTemplate,
			"proto-name": s.protoName,
			"key-prefix": s.KeyPrefix(),
			"REF":        ref, //fmt.Sprintf("%s", ref),
		},
	}
}

// Spec represents model specification for registering models.
type Spec struct {
	Version    string
	Class      string
	Module     string
	Type       string
	IdTemplate string

	protoName string
	keyPrefix string
	idTmpl    *template.Template
}

func (s Spec) String() string {
	return fmt.Sprintf("%s.%s", s.protoName, s.Version)
}

// KeyPrefix returns key prefix used for storing model in KV stores.
func (s Spec) KeyPrefix() string {
	if s.keyPrefix == "" {
		s.keyPrefix = s.buildPrefix()
	}
	return s.keyPrefix
}

// ParseKey parses the given key and returns model ID or
// returns valid as false if the key is not valid.
func (s Spec) ParseKey(key string) (id string, valid bool) {
	trim := strings.TrimPrefix(key, s.KeyPrefix())
	if trim != key && trim != "" {
		return trim, true
	}
	return "", false
}

// IsKeyValid returns true if give key is valid and matches this model Spec.
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

// ProtoName returns proto message name of the model.
func (s Spec) ProtoName() string {
	return s.protoName
}

const prefixTemplate = `{{.Spec.Module}}/{{.Spec.Class}}/{{.Spec.Version}}/{{.Spec.Type}}/`

var prefixTmpl = template.Must(template.New("keyPrefix").Parse(prefixTemplate))

func (s Spec) buildPrefix() string {
	var str strings.Builder
	if err := prefixTmpl.Execute(&str, struct {
		Spec Spec
	}{s}); err != nil {
		panic(err)
	}
	return str.String()
}
