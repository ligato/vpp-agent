// Copyright (c) 2020 Pantheon.tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	yaml2 "github.com/ghodss/yaml"
	"github.com/goccy/go-yaml"
	. "github.com/onsi/gomega"
	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TODO clean up test from development-helping/debug stuff

// TestYamlCompatibility test dynamically generated all-in-one configuration proto message to be compatible
// with its hardcoded counterpart(configurator.Config). The compatibility refers to the ability to use the same
// yaml config file to set the configuration.
func TestYamlCompatibility(t *testing.T) {
	RegisterTestingT(t)

	memIFRed := &interfaces.Interface{
		Name:        "red",
		Type:        interfaces.Interface_MEMIF,
		IpAddresses: []string{"100.0.0.1/24"},
		Mtu:         9200,
		Enabled:     true,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             1,
				Master:         false,
				SocketFilename: "/var/run/memif_k8s-master.sock",
			},
		},
	}
	memIFBlack := &interfaces.Interface{
		Name:        "black",
		Type:        interfaces.Interface_MEMIF,
		IpAddresses: []string{"192.168.20.1/24"},
		Mtu:         9200,
		Enabled:     true,
		Link: &interfaces.Interface_Memif{
			Memif: &interfaces.MemifLink{
				Id:             2,
				Master:         false,
				SocketFilename: "/var/run/memif_k8s-master.sock",
			},
		},
	}
	loop1 := &interfaces.Interface{
		Name:        "loop-test-1",
		Type:        interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		Mtu:         1500,
		IpAddresses: []string{"10.10.1.1/24"},
	}
	vppTap1 := &interfaces.Interface{
		Name:        "vpp-tap1",
		Type:        interfaces.Interface_TAP,
		Enabled:     true,
		IpAddresses: []string{"10.10.10.1/24"},
		Link: &interfaces.Interface_Tap{
			Tap: &interfaces.TapLink{
				Version:        2,
				ToMicroservice: "test-microservice1",
			},
		},
	}
	// TODO add more configuration to hardcoded version of configuration so it can cover all configuration
	//  possibilities
	ifaces := []*interfaces.Interface{memIFRed, memIFBlack, loop1, vppTap1}
	configRoot := configurator.GetResponse{ // using fake Config root to get "config" root element in json/yaml (mimicking agentctl config yaml handling)
		Config: &configurator.Config{
			VppConfig: &vpp.ConfigData{
				Interfaces: ifaces,
			},
		},
	}
	var buf bytes.Buffer
	if err := formatAsTemplate(&buf, "yaml", configRoot); err != nil {
		t.Fatalf("can't export hardcoded config as yaml due to: %v", err)
	}
	yaml := buf.String()

	//client.NewDynamicConfig()
	models, err := client.LocalClient.KnownModels("config")
	if err != nil {
		t.Fatalf("can't retrieve known models due to: %v", err)
	}
	config, err := client.NewDynamicConfig(models)
	if err != nil {
		t.Fatalf("can't create dynamic config due to: %v", err)
	}
	b := []byte(yaml)
	//var update = &configurator.Config{}
	bj, err := yaml2.YAMLToJSON(b)
	if err != nil {
		fmt.Print(err) //TODO
	}
	//msg := proto.MessageV2(allConfigMsg)
	fields := config.ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fmt.Println(fields.Get(i).Name())
	}

	err = protojson.Unmarshal(bj, config)
	if err != nil {
		fmt.Print(err) //TODO
	}

	var buf2 bytes.Buffer
	if err := formatAsTemplate(&buf2, "yaml", config); err != nil {
		t.Fatalf("can't export hardcoded config as yaml due to: %v", err)
	}
	yaml2 := buf2.String()

	Expect(yaml2).To(BeEquivalentTo(yaml))

	fmt.Println(yaml)

	client.DynamicConfigExport(config)

	//opts := cmp.Options{
	//	cmp.Comparer(func (x,y protocmp.Message) bool {
	//		fullName1 := x.Descriptor().FullName()
	//		fullName2 := y.Descriptor().FullName()
	//		if (fullName1 == "ligato.configurator.GetResponse" && fullName2 == "ligato.configurator.Config") ||
	//			(fullName2 == "ligato.configurator.GetResponse" && fullName1 == "ligato.configurator.Config") {
	//			return cmp.Equal(x.ProtoReflect().Get(), , protocmp.Transform())
	//		}
	//	}),
	//	cmp.Transformer("ignoreRootConfigTypes", func(msg protocmp.Message) protocmp.Message {
	//		fn := msg.Descriptor().FullName()
	//		a := fn == "ligato.configurator.GetResponse"
	//		fmt.Print(a)
	//		if msg["@type"] == "ligato.configurator.GetResponse" {
	//			newMsg := make(protocmp.Message)
	//			for k, v := range msg {
	//				newMsg[k] = v
	//			}
	//			newMsg["@type"] = "ligato.configurator.Config"
	//			return newMsg
	//		}
	//		return msg
	//	}),
	//}
	//if diff := cmp.Diff(configRoot, config, protocmp.Transform(), opts); diff != "" {
	//	t.Errorf("Merge mismatch (-want +got):\n%s", diff)
	//}

}

func formatAsTemplate(w io.Writer, format string, data interface{}) error {
	var b bytes.Buffer
	switch strings.ToLower(format) {
	case "json":
		//b.WriteString(jsonTmpl(data))
	case "yaml", "yml":
		b.WriteString(yamlTmpl(data))
	case "proto": // TODO clean up help functions
		//b.WriteString(protoTmpl(data))
		//default:
		//	t := template.New("format")
		//	t.Funcs(tmplFuncs)
		//	if _, err := t.Parse(format); err != nil {
		//		return fmt.Errorf("parsing format template failed: %v", err)
		//	}
		//	if err := t.Execute(&b, data); err != nil {
		//		return fmt.Errorf("executing format template failed: %v", err)
		//	}
	}
	_, err := b.WriteTo(w)
	return err
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
