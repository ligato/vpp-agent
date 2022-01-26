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
	"testing"

	yaml2 "github.com/ghodss/yaml"
	"github.com/go-errors/errors"
	"github.com/goccy/go-yaml"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	testmodel "go.ligato.io/vpp-agent/v3/pkg/models/testdata/proto"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	vpp_srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"
)

// TestYamlCompatibility test dynamically generated all-in-one configuration proto message to be compatible
// with its hardcoded counterpart(configurator.Config). The compatibility refers to the ability to use the same
// yaml config file to set the configuration.
func TestYamlCompatibility(t *testing.T) {
	RegisterTestingT(t)

	// fill hardcoded Config with configuration
	ifaces := []*interfaces.Interface{memIFRed, memIFBlack, loop1, vppTap1}
	config := &configurator.Config{
		VppConfig: &vpp.ConfigData{
			Interfaces: ifaces,
			Srv6Global: srv6Global,
		},
	}
	// TODO add more configuration to hardcoded version of configuration so it can cover all configuration
	//  possibilities

	// create construction input for dynamic config from locally registered models (only with class "config")
	// (for remote models use generic client's KnownModels, example of this is in agentctl yaml config
	// update (commands.runConfigUpdate))
	var knownModels []*models.ModelInfo
	for _, model := range models.RegisteredModels() {
		if model.Spec().Class == "config" {
			knownModels = append(knownModels, &models.ModelInfo{
				ModelDetail:       model.ModelDetail(),
				MessageDescriptor: model.NewInstance().ProtoReflect().Descriptor(),
			})
		}
	}

	// create dynamic config
	// Note: for revealing dynamic structure use fmt.Println(client.ExportDynamicConfigStructure(dynConfig))
	dynConfig, err := client.NewDynamicConfig(knownModels)
	Expect(err).ShouldNot(HaveOccurred(), "can't create dynamic config")

	// Hardcoded Config filled with data -> YAML -> JSON -> load into empty dynamic Config -> YAML
	yamlFromHardcodedConfig, err := toYAML(config) // should be the same output as agentctl config get
	Expect(err).ShouldNot(HaveOccurred(), "can't export hardcoded config as yaml (initial export)")
	bj, err := yaml2.YAMLToJSON([]byte(yamlFromHardcodedConfig))
	Expect(err).ShouldNot(HaveOccurred(), "can't convert yaml (from hardcoded config) to json")
	Expect(protojson.Unmarshal(bj, dynConfig)).To(Succeed(),
		"can't marshal json data (from hardcoded config) to dynamic config")
	yamlFromDynConfig, err := toYAML(dynConfig)
	Expect(err).ShouldNot(HaveOccurred(), "can't export hardcoded config as yaml")

	// final compare of YAML from hardcoded and dynamic config
	Expect(yamlFromDynConfig).To(BeEquivalentTo(yamlFromHardcodedConfig))
}

// TestDynamicConfigWithThirdPartyModel tests whether 3rd party model (= model not in hardcoded configurator.Config)
// data can be loaded into dynamic config from yaml form
func TestDynamicConfigWithThirdPartyModel(t *testing.T) {
	RegisterTestingT(t)
	yaml := `modelConfig:
    Basic_list:
    - name: testName1
    - name: testName2
`
	models.Register(&testmodel.Basic{}, models.Spec{
		Module:  "model",
		Type:    "basic",
		Version: "v1",
	}, models.WithNameTemplate("{{.Name}}")) // contains Name template => as repeated field in dynamic config
	// Note: no Name template (and no GetName() in generated proto message) => as optional field in
	// dynamic config (without "_list" suffix and  list to single reference in yaml)

	// create construction input for dynamic config
	var knownModels []*models.ModelInfo
	for _, model := range models.RegisteredModels() {
		if model.Spec().Class == "config" && model.Spec().Module == "model" {
			knownModels = append(knownModels, &models.ModelInfo{
				ModelDetail:       model.ModelDetail(),
				MessageDescriptor: model.NewInstance().ProtoReflect().Descriptor(),
			})
		}
	}

	// create dynamic config
	// Note: for revealing dynamic sctructure use fmt.Println(client.ExportDynamicConfigStructure(dynConfig))
	dynConfig, err := client.NewDynamicConfig(knownModels)
	Expect(err).ShouldNot(HaveOccurred(), "can't create dynamic config")

	// test loading of 3rd party model data into dynamic config
	bj2, err := yaml2.YAMLToJSON([]byte(yaml))
	Expect(err).ShouldNot(HaveOccurred(), "can't convert yaml to json")
	Expect(protojson.Unmarshal(bj2, dynConfig)).To(Succeed(),
		"can't marshal json data to dynamic config")
}

func toYAML(data interface{}) (string, error) {
	out, err := encodeJson(data, "")
	if err != nil {
		return "", errors.Errorf("can't encode to JSON due to: %v", err)
	}
	bb, err := jsonToYaml(out)
	if err != nil {
		return "", errors.Errorf("can't convert json to yaml due to: %v", err)
	}
	return string(bb), nil
}

func encodeJson(data interface{}, ident string) ([]byte, error) {
	if msg, ok := data.(proto.Message); ok {
		m := protojson.MarshalOptions{
			Indent: ident,
		}
		b, err := m.Marshal(msg)
		if err != nil {
			return nil, errors.Errorf("can't marshal proto message to json due to: %v", err)
		}
		return b, nil
	}
	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	encoder.SetIndent("", ident)
	if err := encoder.Encode(data); err != nil {
		return nil, errors.Errorf("can't marshal data to json due to: %v", err)
	}
	return b.Bytes(), nil
}

func jsonToYaml(j []byte) ([]byte, error) {
	var jsonObj interface{}
	err := yaml.UnmarshalWithOptions(j, &jsonObj, yaml.UseOrderedMap())
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(jsonObj)
}

// test configuration
var (
	memIFRed = &interfaces.Interface{
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
	memIFBlack = &interfaces.Interface{
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
	loop1 = &interfaces.Interface{
		Name:        "loop-test-1",
		Type:        interfaces.Interface_SOFTWARE_LOOPBACK,
		Enabled:     true,
		Mtu:         1500,
		IpAddresses: []string{"10.10.1.1/24"},
	}
	vppTap1 = &interfaces.Interface{
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
	srv6Global = &vpp_srv6.SRv6Global{
		EncapSourceAddress: "10.1.1.1",
	}
)
