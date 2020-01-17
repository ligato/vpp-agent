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

package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/proto"
)

var tmplFuncs = template.FuncMap{
	"json":  jsonTmpl,
	"yaml":  yamlTmpl,
	"proto": protoTmpl,
	"epoch": epochTmpl,
	"ago":   agoTmpl,
}

func formatAsTemplate(w io.Writer, format string, data interface{}) error {
	t := template.New("format")
	t.Funcs(tmplFuncs)

	if format == "json" {
		format = "{{json .}}"
	} else if format == "yaml" {
		format = "{{yaml .}}"
	} else if format == "proto" {
		format = "{{proto .}}"
	}

	if _, err := t.Parse(format); err != nil {
		return fmt.Errorf("parsing format template failed: %v", err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return fmt.Errorf("executing format template failed: %v", err)
	}

	_, err := b.WriteTo(w)
	return err
}

func yamlTmpl(data interface{}) string {
	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	if err := encoder.Encode(data); err != nil {
		panic(err)
	}
	bb, err := yaml.JSONToYAML(b.Bytes())
	if err != nil {
		panic(err)
	}
	return string(bb)
}

func jsonTmpl(data interface{}) string {
	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		panic(err)
	}
	return b.String()
}

func protoTmpl(data interface{}) string {
	pb, ok := data.(proto.Message)
	if !ok {
		panic(fmt.Sprintf("%T is not a proto message", data))
	}
	var b bytes.Buffer
	m := proto.TextMarshaler{}
	if err := m.Marshal(&b, pb); err != nil {
		panic(err)
	}
	return b.String()
}

func epochTmpl(s int64) time.Time {
	return time.Unix(s, 0)
}

func agoTmpl(t time.Time) time.Duration {
	return time.Since(t).Round(time.Second)
}
