// Copyright (c) 2019 Cisco and/or its affiliates.
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

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/ligato/vpp-agent/api/genericmanager"
	"github.com/ligato/vpp-agent/pkg/models"

	_ "github.com/ligato/vpp-agent/api/models/linux"
	_ "github.com/ligato/vpp-agent/api/models/vpp"
	//linux_if_keys "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	//linux_l3_keys "github.com/ligato/vpp-agent/api/models/linux/l3"
	//if_keys "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
	//ipsec_keys "github.com/ligato/vpp-agent/api/models/vpp/ipsec"
	//l2_keys "github.com/ligato/vpp-agent/api/models/vpp/l2"
	//l3_keys "github.com/ligato/vpp-agent/api/models/vpp/l3"
	//nat_keys "github.com/ligato/vpp-agent/api/models/vpp/nat"
	//punt_keys "github.com/ligato/vpp-agent/api/models/vpp/punt"
	//stn_keys "github.com/ligato/vpp-agent/api/models/vpp/stn"
)

const (
	// target file path
	path = "docs/KeyOverview.md"

	// key high-level plugin prefixes
	vppPrefix   = "config/vpp/"
	linuxPrefix = "config/linux/"

	// link prefixes
	linkPrefix = "https://github.com/ligato/vpp-agent/blob/master/api/models"
	vnfPrefix  = "/vnf-agent/<ms-label>/config/"
)

const (
	header = `GENERATED FILE, DO NOT EDIT BY HAND

This page is an overview of all keys supported for the VPP-Agent

# Key overview

- [VPP keys](#vpp)
- [Linux keys](#linux)

Parts of the key in ` + "`<>`" + ` must be set with the same value as in a model. The microservice label is set to ` + "`vpp1`" + ` in every mentioned key, but if different value is used, it needs to be replaced in the key as well.

Link in key title redirects to the associated proto definition.

`
	vppTitle = `### <a name="vpp">VPP keys</a>

`
	linuxTitle = `### <a name="linux">Linux keys</a>

`
)

type File struct {
	file *os.File
}

// Generates a file with all vpp-agent key templates
func main() {
	file, err := os.Create(path)
	var buffer bytes.Buffer

	// store available models into vpp and linux categories
	var vppModels, linuxModels []*genericmanager.ModelInfo
	for _, model := range models.RegisteredModels() {
		if strings.HasPrefix(model.Info["keyPrefix"], vppPrefix) {
			vppModels = append(vppModels, model)
		}
		if strings.HasPrefix(model.Info["keyPrefix"], linuxPrefix) {
			linuxModels = append(linuxModels, model)
		}
	}

	// to have consistent order, sort models according to their names
	vppModels, linuxModels = sortModels(vppModels), sortModels(linuxModels)

	// write header, vpp title and vpp models
	write(header+vppTitle, &buffer)
	write(generateForModels(vppModels), &buffer)

	// add linux title and models
	write(linuxTitle, &buffer)
	write(generateForModels(linuxModels), &buffer)

	// store configuration to file
	_, err = io.Copy(file, strings.NewReader(buffer.String()))
	if err != nil {
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}

// Generate entry for every model from the list in format:
// **[ModelName:](https://link/to/proto)**
// ```
// /model/key
// ```
func generateForModels(models []*genericmanager.ModelInfo) (data string) {
	for _, model := range models {
		// link to proto path
		var protoPath string
		moduleParts := strings.Split(model.GetModel().GetModule(), ".")

		// file name is derived from proto file name if defined, or from type
		var fileName string
		if model.GetModel().GetProtoFileName() != "" {
			fileName = model.GetModel().GetProtoFileName()
		} else {
			fileName = model.GetModel().GetType()
		}
		// if module is single value, use type as proto package
		if len(moduleParts) == 1 {
			protoPath = fmt.Sprintf("%s/%s/%s/%s.proto", linkPrefix, moduleParts[0], model.GetModel().GetType(), fileName)
		}
		// if module also contains package name, us it
		if len(moduleParts) == 2 {
			protoPath = fmt.Sprintf("%s/%s/%s/%s.proto", linkPrefix, moduleParts[0], moduleParts[1], fileName)
		}

		data += fmt.Sprintf("**[%s:](%s)**\n```\n%s\n```\n\n", prune(model.Info["protoName"]), protoPath, parseModelKey(model))
	}

	return data
}

// Sort models alphabetically, based on model proto names. All names are considered
// lower-cased for this purpose.
func sortModels(models []*genericmanager.ModelInfo) []*genericmanager.ModelInfo {
	var sorted []*genericmanager.ModelInfo
ModelLoop:
	for _, model := range models {
		// first entry can be just stored
		if len(sorted) == 0 {
			sorted = append(sorted, model)
			continue
		}
		// loops through known models and looks where the current one should be put. It compares ASCII values
		// of both names and places the model to respective place.
		for i, sortedModel := range sorted {
			if func(n1, n2 string) bool {
				n1, n2 = strings.ToLower(n1), strings.ToLower(n2)
				for i, rune1 := range n1 {
					ascii1 := int(rune1)
					if len(n2) < i+1 {
						return false
					}
					ascii2 := int(n2[i])
					if ascii1 > ascii2 {
						return false
					} else if ascii1 < ascii2 {
						return true
					} else {
						continue
					}
				}
				return false
			}(prune(model.Info["protoName"]), prune(sortedModel.Info["protoName"])) {
				sorted = append(sorted, nil)
				copy(sorted[i+1:], sorted[i:])
				sorted[i] = model
				continue ModelLoop
			}
			if len(sorted) == i+1 {
				sorted = append(sorted, model)
			}
		}
	}

	return sorted
}


// Parses model path and template to reconstruct key. Template is parsed as following:
//   "{{.BridgeDomain}}/mac/{{.PhysAddress}}" => /<bridge-domain>/mac/<phys-address>
func parseModelKey(model *genericmanager.ModelInfo) string {
	// prepare parts of the key leveraged from template if exists
	var templateParts []string
	parts := strings.Split(model.Info["nameTemplate"], "/")
	for _, part := range parts {
		var word string
		var lever bool
		for i, letter := range part {
			// special case when verbs are used, skip them
			if i == 0 && letter == '%' {
				break
			}
			// if part is a plain text, just copy the whole word
			if i == 0 && letter != '{' {
				templateParts = append(templateParts, part)
				break
			}
			// detect end of the template directive
			if letter == '}' || letter == ' ' {
				lever = false
				if word != "" {
					// at this point, the word is in format '.SomeName' and will be formatted to '<some-name>'
					word = strings.Replace(word, ".", "", 1)
					var formatted string
					for j, wordLetter := range word {
						if j != 0 && unicode.IsUpper(wordLetter) {
							formatted += "-"
						}
						formatted += string(wordLetter)
					}
					templateParts = append(templateParts, strings.ToLower("<" + formatted + ">"))
					break
				}
			}
			// dot marks beginning of the template directive and "switches" the lever so all characters
			// after that are counted
			if letter == '.' || lever {
				lever = true
				word += string(letter)
			}
		}
	}

	// add path common prefix, model path and template parts if exist
	modelPath := strings.Replace(model.Info["modelPath"], ".", "/", -1)
	// TODO workarounds for some specific keys because of the inconsistencies, will be removed when fixed
	if !strings.Contains(modelPath, "linux") && strings.Contains(modelPath, "l3/") {
		modelPath = strings.Replace(modelPath, "l3/", "", 1)
	}
	if strings.Contains(modelPath, "punt/") {
		modelPath = strings.Replace(modelPath, "punt/", "", 1)
	}

	// reconstruct the key
	key := vnfPrefix + modelPath
	for _, part := range templateParts {
		key += "/" + part
	}

	return key
}

func prune(protoName string) string {
	parts := strings.Split(protoName, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func write(data string, buffer *bytes.Buffer) {
	if _, err := buffer.WriteString(data); err != nil {
		panic(err)
	}
}
