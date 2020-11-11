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
	"strings"
	"text/template"
	"time"

	"github.com/goccy/go-yaml"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var tmplFuncs = template.FuncMap{
	"json":       jsonTmpl,
	"yaml":       yamlTmpl,
	"proto":      protoTmpl,
	"protomulti": protoTmplMulti,
	"epoch":      epochTmpl,
	"ago":        agoTmpl,
	"dur":        shortHumanDuration,
	"prefix":     prefixTmpl,
}

func formatAsTemplate(w io.Writer, format string, data interface{}) error {
	var b bytes.Buffer
	switch strings.ToLower(format) {
	case "json":
		b.WriteString(jsonTmpl(data))
	case "yaml", "yml":
		b.WriteString(yamlTmpl(data))
	case "proto":
		b.WriteString(protoTmpl(data))
	default:
		t := template.New("format")
		t.Funcs(tmplFuncs)
		if _, err := t.Parse(format); err != nil {
			return fmt.Errorf("parsing format template failed: %v", err)
		}
		if err := t.Execute(&b, data); err != nil {
			return fmt.Errorf("executing format template failed: %v", err)
		}
	}
	_, err := b.WriteTo(w)
	return err
}

func jsonTmpl(data interface{}) string {
	b := encodeJson(data, "  ")
	return string(b)
}

func yamlTmpl(data interface{}) string {
	out := encodeJson(data, "")
	bb, err := jsonToYaml(out)
	if err != nil {
		panic(err)
	}
	return string(bb)
}

func encodeJson(data interface{}, ident string) []byte {
	if msg, ok := data.(proto.Message); ok {
		m := protojson.MarshalOptions{
			Indent: ident,
		}
		b, err := m.Marshal(msg)
		if err != nil {
			panic(err)
		}
		return b
	}
	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	encoder.SetIndent("", ident)
	if err := encoder.Encode(data); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func jsonToYaml(j []byte) ([]byte, error) {
	var jsonObj interface{}
	err := yaml.UnmarshalWithOptions(j, &jsonObj, yaml.UseOrderedMap())
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(jsonObj)
}

func protoTmpl(data interface{}) string {
	pb, ok := data.(proto.Message)
	if !ok {
		panic(fmt.Sprintf("%T is not a proto message", data))
	}
	out, err := prototext.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func protoTmplMulti(data interface{}) string {
	pb, ok := data.(proto.Message)
	if !ok {
		panic(fmt.Sprintf("%T is not a proto message", data))
	}
	out, err := prototext.MarshalOptions{Multiline: true}.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func epochTmpl(s int64) time.Time {
	return time.Unix(s, 0)
}

func agoTmpl(t time.Time) time.Duration {
	return time.Since(t).Round(time.Second)
}

func shortHumanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < -1 {
		return fmt.Sprintf("<invalid>")
	} else if seconds < 0 {
		return fmt.Sprintf("0s")
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}

func prefixTmpl(s string, prefix string) string {
	ps := strings.TrimRight(s, "\n")
	ps = strings.ReplaceAll(ps, "\n", "\n"+prefix)
	if strings.HasSuffix(s, "\n") {
		ps += "\n"
	}
	return prefix + ps
}
