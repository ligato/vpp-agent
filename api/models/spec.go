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
	"regexp"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/api"
)

var debugRegister = strings.Contains(os.Getenv("DEBUG_MODELS"), "register")

// Spec represents model specification.
type Spec struct {
	Module   string
	Type     string
	Version  string
	Class    string
	IDFormat string

	protoName string
	keyPrefix string
	idTmpl    *template.Template
}

// ProtoName returns proto message name of the model.
func (s Spec) ProtoName() string {
	return s.protoName
}

// KeyPrefix returns key prefix used for storing model in KV stores.
func (s Spec) KeyPrefix() string {
	return s.keyPrefix
}

// ParseKey parses the given key and returns model ID or
// returns valid as false if the key is not valid.
func (s Spec) ParseKey(key string) (id string, valid bool) {
	trim := strings.TrimPrefix(key, s.KeyPrefix())
	if trim != key && trim != "" {
		// TODO: validate name?
		return trim, true
	}
	return "", false
}

// IsKeyValid returns true if give key is valid and matches this model Spec.
func (s Spec) IsKeyValid(key string) bool {
	_, valid := s.ParseKey(key)
	return valid

}

// StripKeyPrefix returns key with prefix stripped.
func (s Spec) StripKeyPrefix(key string) string {
	trim := strings.TrimPrefix(key, s.KeyPrefix())
	if trim != key && trim != "" {
		return trim
	}
	return key
}

var (
	registeredSpecs = make(map[string]*Spec)
	keyPrefixes     = make(map[string]string)
)

// RegisteredModels returns all registered modules.
func RegisteredModels() (models []*api.Model) {
	for _, s := range registeredSpecs {
		models = append(models, &api.Model{
			Module:  s.Module,
			Type:    s.Type,
			Version: s.Version,
			Meta: map[string]string{
				"id-format":  s.IDFormat,
				"proto-name": s.protoName,
				"key-prefix": s.keyPrefix,
			},
		})
	}
	return
}

var (
	validType   = regexp.MustCompile(`^[a-z_0-9-]+$`)
	validModule = regexp.MustCompile(`^[a-z_0-9-/]+$`)
)

const keyPrefix = `{{.Class}}/{{.Version}}/{{.Module}}/{{.Type}}/`

// Register registers given protobuf with model specification.
func Register(pb proto.Message, spec Spec) {
	spec.protoName = proto.MessageName(pb)

	if _, ok := registeredSpecs[spec.protoName]; ok {
		panic(fmt.Sprintf("duplicate model registered: %s", spec.protoName))
	}
	if !validModule.MatchString(spec.Module) {
		panic(fmt.Sprintf("module for model %s is invalid", spec.protoName))
	}
	if !validType.MatchString(spec.Type) {
		panic(fmt.Sprintf("name for model %s is invalid", spec.protoName))
	}
	if !strings.HasPrefix(spec.Version, "v") {
		panic(fmt.Sprintf("version for model %s is invalid", spec.protoName))
	}
	if spec.Class == "" {
		panic(fmt.Sprintf("class for model %s is empty", spec.protoName))
	}
	if spec.IDFormat == "" {
		panic(fmt.Sprintf("ID format for model %s is empty", spec.protoName))
	}
	spec.idTmpl = template.Must(template.New("ID").Funcs(funcMap).Parse(spec.IDFormat))

	prefixTmpl := template.Must(template.New("keyPrefix").Parse(keyPrefix))
	var prefix strings.Builder
	if err := prefixTmpl.Execute(&prefix, spec); err != nil {
		panic(err)
	}
	spec.keyPrefix = prefix.String()
	if pn, ok := keyPrefixes[spec.keyPrefix]; ok {
		panic(fmt.Sprintf("key prefix %q already used by: %s", spec.keyPrefix, pn))
	}
	keyPrefixes[spec.keyPrefix] = spec.protoName

	if debugRegister {
		fmt.Printf("- registered model: %-40v\t%q\n", spec, spec.KeyPrefix())
	}
	registeredSpecs[spec.protoName] = &spec
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
