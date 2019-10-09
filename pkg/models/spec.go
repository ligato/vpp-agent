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
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"

	api "github.com/ligato/vpp-agent/api/genericmanager"
)

var (
	validModule = regexp.MustCompile(`^[-a-z0-9_]+(?:\.[-a-z0-9_]+)?$`)
	validType   = regexp.MustCompile(`^[-a-z0-9_]+(?:\.[-a-z0-9_]+)?$`)
)

// Model represents a registered model.
type Model struct {
	modelSpec

	goType    reflect.Type
	protoName string
	keyPrefix string
	modelPath string

	modelOptions
}

// Spec defines model specification used for registering model.
type Spec api.Model

type modelSpec struct {
	Spec
}

// NameFunc represents function which can name model instance.
type NameFunc func(obj interface{}) (string, error)

type modelOptions struct {
	nameTemplate string
	nameFunc     NameFunc
}

// ModelOption defines function type which sets model options.
type ModelOption func(*modelOptions)

// WithNameTemplate returns option for models which sets function
// for generating name of instances using custom template.
func WithNameTemplate(t string) ModelOption {
	return func(opts *modelOptions) {
		opts.nameFunc = NameTemplate(t)
		opts.nameTemplate = t
	}
}

// NewInstance creates new instance value for model type.
func (m Model) NewInstance() proto.Message {
	return reflect.New(m.goType.Elem()).Interface().(proto.Message)
}

// ProtoName returns proto message name registered with the model.
func (m Model) ProtoName() string {
	if m.protoName == "" {
		proto.MessageName(m.NewInstance())
	}
	return m.protoName
}

// Path returns path for the model.
func (m Model) Path() string {
	return m.modelPath
}

// KeyPrefix returns key prefix for the model.
func (m Model) KeyPrefix() string {
	return m.keyPrefix
}

// ParseKey parses the given key and returns item name
// or returns empty name and valid as false if the key is not valid.
func (m Model) ParseKey(key string) (name string, valid bool) {
	name = strings.TrimPrefix(key, m.keyPrefix)
	if name == key || (name == "" && m.nameFunc != nil) {
		name = strings.TrimPrefix(key, m.modelPath)
	}
	// key had the prefix and also either
	// non-empty name or no name template
	if name != key && (name != "" || m.nameFunc == nil) {
		// TODO: validate name?
		return name, true
	}
	return "", false
}

// IsKeyValid returns true if given key is valid for this model.
func (m Model) IsKeyValid(key string) bool {
	_, valid := m.ParseKey(key)
	return valid
}

// StripKeyPrefix returns key with prefix stripped.
func (m Model) StripKeyPrefix(key string) string {
	if name, valid := m.ParseKey(key); valid {
		return name
	}
	return key
}

func (m Model) name(x proto.Message) (string, error) {
	if m.nameFunc == nil {
		return "", nil
	}
	return m.nameFunc(x)
}

func buildModelPath(version, module, typ string) string {
	return fmt.Sprintf("%s.%s.%s", module, version, typ)
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
	"ip": func(s string) string {
		ip := net.ParseIP(s)
		if ip == nil {
			return "<invalid>"
		}
		return ip.String()
	},
	"protoip": func(s string) string {
		ip := net.ParseIP(s)
		if ip == nil {
			return "<invalid>"
		}

		if ip.To4() == nil {
			return "IPv6"
		}
		return "IPv4"
	},
	"ipnet": func(s string) map[string]interface{} {
		if strings.HasPrefix(s, "alloc:") {
			// reference to IP address allocated via netalloc
			return nil
		}
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			return map[string]interface{}{
				"IP":       "<invalid>",
				"MaskSize": 0,
				"AllocRef": "",
			}
		}
		maskSize, _ := ipNet.Mask.Size()
		return map[string]interface{}{
			"IP":       ipNet.IP.String(),
			"MaskSize": maskSize,
			"AllocRef": "",
		}
	},
}
