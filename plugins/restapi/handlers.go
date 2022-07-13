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

//go:generate go-bindata-assetfs -pkg restapi -o bindata.go ./templates/...

package restapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	yaml2 "github.com/ghodss/yaml"
	"github.com/goccy/go-yaml"
	"github.com/unrolled/render"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/pluginpb"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/version"
	"go.ligato.io/vpp-agent/v3/plugins/configurator"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	kvscheduler "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator/contextdecorator"
	"go.ligato.io/vpp-agent/v3/plugins/restapi/jsonschema/converter"
	"go.ligato.io/vpp-agent/v3/plugins/restapi/resturl"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const (
	// URLFieldNamingParamName is URL parameter name for JSON schema http handler's setting
	// to output field names using proto/json/both names for fields
	URLFieldNamingParamName = "fieldnames"
	// OnlyProtoFieldNames is URL parameter value for JSON schema http handler to use only proto names as field names
	OnlyProtoFieldNames = "onlyproto"
	// OnlyJSONFieldNames is URL parameter value for JSON schema http handler to use only JSON names as field names
	OnlyJSONFieldNames = "onlyjson"

	// URLReplaceParamName is URL parameter name for modifying NB configuration PUT behaviour to act as whole
	// configuration replacer instead of config updater (fullresync vs update). It has the same effect as replace
	// parameter for agentctl config update.
	// Examples how to use full resync:
	// <VPP-Agent IP address>:9191/configuration?replace
	// <VPP-Agent IP address>:9191/configuration?replace=true
	URLReplaceParamName = "replace"

	// YamlContentType is http header content type for YAML content
	YamlContentType = "application/yaml"

	internalErrorLogPrefix = "500 Internal server error: "
)

var (
	// ErrHandlerUnavailable represents error returned when particular
	// handler is not available
	ErrHandlerUnavailable = errors.New("handler is not available")
)

func (p *Plugin) registerInfoHandlers() {
	p.HTTPHandlers.RegisterHTTPHandler(resturl.Version, p.versionHandler, GET)
	p.HTTPHandlers.RegisterHTTPHandler(resturl.JSONSchema, p.jsonSchemaHandler, GET)
}

func (p *Plugin) registerNBConfigurationHandlers() {
	p.HTTPHandlers.RegisterHTTPHandler(resturl.Validate, p.validationHandler, POST)
	p.HTTPHandlers.RegisterHTTPHandler(resturl.Configuration, p.configurationGetHandler, GET)
	p.HTTPHandlers.RegisterHTTPHandler(resturl.Configuration, p.configurationUpdateHandler, PUT)
}

// Registers ABF REST handler
func (p *Plugin) registerABFHandler() {
	p.registerHTTPHandler(resturl.ABF, GET, func() (interface{}, error) {
		if p.abfHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.abfHandler.DumpABFPolicy()
	})
}

// Registers access list REST handlers
func (p *Plugin) registerACLHandlers() {
	// GET IP ACLs
	p.registerHTTPHandler(resturl.ACLIP, GET, func() (interface{}, error) {
		if p.aclHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.aclHandler.DumpACL()
	})
	// GET MACIP ACLs
	p.registerHTTPHandler(resturl.ACLMACIP, GET, func() (interface{}, error) {
		if p.aclHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.aclHandler.DumpMACIPACL()
	})
}

// Registers interface REST handlers
func (p *Plugin) registerInterfaceHandlers() {
	// GET all interfaces
	p.registerHTTPHandler(resturl.Interface, GET, func() (interface{}, error) {
		return p.ifHandler.DumpInterfaces(context.TODO())
	})
	// GET loopback interfaces
	p.registerHTTPHandler(resturl.Loopback, GET, func() (interface{}, error) {
		return p.ifHandler.DumpInterfacesByType(context.TODO(), interfaces.Interface_SOFTWARE_LOOPBACK)
	})
	// GET ethernet interfaces
	p.registerHTTPHandler(resturl.Ethernet, GET, func() (interface{}, error) {
		return p.ifHandler.DumpInterfacesByType(context.TODO(), interfaces.Interface_DPDK)
	})
	// GET memif interfaces
	p.registerHTTPHandler(resturl.Memif, GET, func() (interface{}, error) {
		return p.ifHandler.DumpInterfacesByType(context.TODO(), interfaces.Interface_MEMIF)
	})
	// GET tap interfaces
	p.registerHTTPHandler(resturl.Tap, GET, func() (interface{}, error) {
		return p.ifHandler.DumpInterfacesByType(context.TODO(), interfaces.Interface_TAP)
	})
	// GET af-packet interfaces
	p.registerHTTPHandler(resturl.AfPacket, GET, func() (interface{}, error) {
		return p.ifHandler.DumpInterfacesByType(context.TODO(), interfaces.Interface_AF_PACKET)
	})
	// GET VxLAN interfaces
	p.registerHTTPHandler(resturl.VxLan, GET, func() (interface{}, error) {
		return p.ifHandler.DumpInterfacesByType(context.TODO(), interfaces.Interface_VXLAN_TUNNEL)
	})
}

// Registers NAT REST handlers
func (p *Plugin) registerNATHandlers() {
	// GET NAT global config
	p.registerHTTPHandler(resturl.NatGlobal, GET, func() (interface{}, error) {
		if p.natHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.natHandler.Nat44GlobalConfigDump(false)
	})
	// GET DNAT config
	p.registerHTTPHandler(resturl.NatDNat, GET, func() (interface{}, error) {
		if p.natHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.natHandler.DNat44Dump()
	})
	// GET NAT interfaces
	p.registerHTTPHandler(resturl.NatInterfaces, GET, func() (interface{}, error) {
		if p.natHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.natHandler.Nat44InterfacesDump()
	})
	// GET NAT address pools
	p.registerHTTPHandler(resturl.NatAddressPools, GET, func() (interface{}, error) {
		if p.natHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.natHandler.Nat44AddressPoolsDump()
	})
}

// Registers L2 plugin REST handlers
func (p *Plugin) registerL2Handlers() {
	// GET bridge domains
	p.registerHTTPHandler(resturl.Bd, GET, func() (interface{}, error) {
		if p.l2Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l2Handler.DumpBridgeDomains()
	})
	// GET FIB entries
	p.registerHTTPHandler(resturl.Fib, GET, func() (interface{}, error) {
		if p.l2Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l2Handler.DumpL2FIBs()
	})
	// GET cross connects
	p.registerHTTPHandler(resturl.Xc, GET, func() (interface{}, error) {
		if p.l2Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l2Handler.DumpXConnectPairs()
	})
}

// Registers L3 plugin REST handlers
func (p *Plugin) registerL3Handlers() {
	// GET ARP entries
	p.registerHTTPHandler(resturl.Arps, GET, func() (interface{}, error) {
		if p.l3Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l3Handler.DumpArpEntries()
	})
	// GET proxy ARP interfaces
	p.registerHTTPHandler(resturl.PArpIfs, GET, func() (interface{}, error) {
		if p.l3Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l3Handler.DumpProxyArpInterfaces()
	})
	// GET proxy ARP ranges
	p.registerHTTPHandler(resturl.PArpRngs, GET, func() (interface{}, error) {
		if p.l3Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l3Handler.DumpProxyArpRanges()
	})
	// GET static routes
	p.registerHTTPHandler(resturl.Routes, GET, func() (interface{}, error) {
		if p.l3Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l3Handler.DumpRoutes()
	})
	// GET scan ip neighbor setup
	p.registerHTTPHandler(resturl.IPScanNeigh, GET, func() (interface{}, error) {
		if p.l3Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l3Handler.GetIPScanNeighbor()
	})
	// GET vrrp entries
	p.registerHTTPHandler(resturl.Vrrps, GET, func() (interface{}, error) {
		if p.l3Handler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.l3Handler.DumpVrrpEntries()
	})
}

// Registers IPSec plugin REST handlers
func (p *Plugin) registerIPSecHandlers() {
	// GET IPSec SPD entries
	p.registerHTTPHandler(resturl.SPDs, GET, func() (interface{}, error) {
		if p.ipSecHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.ipSecHandler.DumpIPSecSPD()
	})
	// GET IPSec SP entries
	p.registerHTTPHandler(resturl.SPs, GET, func() (interface{}, error) {
		if p.ipSecHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.ipSecHandler.DumpIPSecSP()
	})
	// GET IPSec SA entries
	p.registerHTTPHandler(resturl.SAs, GET, func() (interface{}, error) {
		if p.ipSecHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.ipSecHandler.DumpIPSecSA()
	})
}

// Registers punt plugin REST handlers
func (p *Plugin) registerPuntHandlers() {
	// GET punt registered socket entries
	p.registerHTTPHandler(resturl.PuntSocket, GET, func() (interface{}, error) {
		if p.puntHandler == nil {
			return nil, ErrHandlerUnavailable
		}
		return p.puntHandler.DumpRegisteredPuntSockets()
	})
}

// Registers linux interface plugin REST handlers
func (p *Plugin) registerLinuxInterfaceHandlers() {
	// GET linux interfaces
	p.registerHTTPHandler(resturl.LinuxInterface, GET, func() (interface{}, error) {
		return p.linuxIfHandler.DumpInterfaces()
	})
	// GET linux interface stats
	p.registerHTTPHandler(resturl.LinuxInterfaceStats, GET, func() (interface{}, error) {
		return p.linuxIfHandler.DumpInterfaceStats()
	})
}

// Registers linux L3 plugin REST handlers
func (p *Plugin) registerLinuxL3Handlers() {
	// GET linux routes
	p.registerHTTPHandler(resturl.LinuxRoutes, GET, func() (interface{}, error) {
		return p.linuxL3Handler.DumpRoutes()
	})
	// GET linux ARPs
	p.registerHTTPHandler(resturl.LinuxArps, GET, func() (interface{}, error) {
		return p.linuxL3Handler.DumpARPEntries()
	})
}

// Registers Telemetry handler
func (p *Plugin) registerTelemetryHandlers() {
	p.HTTPHandlers.RegisterHTTPHandler(resturl.Telemetry, p.telemetryHandler, GET)
	p.HTTPHandlers.RegisterHTTPHandler(resturl.TMemory, p.telemetryMemoryHandler, GET)
	p.HTTPHandlers.RegisterHTTPHandler(resturl.TRuntime, p.telemetryRuntimeHandler, GET)
	p.HTTPHandlers.RegisterHTTPHandler(resturl.TNodeCount, p.telemetryNodeCountHandler, GET)
}

func (p *Plugin) registerStatsHandler() {
	p.HTTPHandlers.RegisterHTTPHandler(resturl.ConfiguratorStats, p.configuratorStatsHandler, GET)
}

// Registers index page
func (p *Plugin) registerIndexHandlers() {
	r := render.New(render.Options{
		Directory:  "templates",
		Asset:      Asset,
		AssetNames: AssetNames,
	})
	handlerFunc := func(formatter *render.Render) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {

			p.Log.Debugf("%v - %s %q", req.RemoteAddr, req.Method, req.URL)
			p.logError(r.HTML(w, http.StatusOK, "index", p.index))
		}
	}
	p.HTTPHandlers.RegisterHTTPHandler("/", handlerFunc, GET)
}

// registerHTTPHandler is common register method for all handlers
func (p *Plugin) registerHTTPHandler(key, method string, f func() (interface{}, error)) {
	handlerFunc := func(formatter *render.Render) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			p.govppmux.Lock()
			defer p.govppmux.Unlock()

			res, err := f()
			if err != nil {
				errMsg := fmt.Sprintf("500 Internal server error: request failed: %v\n", err)
				p.Log.Error(errMsg)
				p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
				return
			}
			p.Deps.Log.Debugf("Rest uri: %s, data: %v", key, res)
			p.logError(formatter.JSON(w, http.StatusOK, res))
		}
	}
	p.HTTPHandlers.RegisterHTTPHandler(key, handlerFunc, method)
}

// jsonSchemaHandler returns JSON schema of VPP-Agent configuration.
// This handler also accepts URL query parameters changing the exported field names of proto messages. By default,
// proto message fields are exported twice in JSON scheme. Once with proto name and once with JSON name. This should
// allow to use any of the 2 forms in JSON/YAML configuration when used JSON schema for validation. However,
// this behaviour can be modified by URLFieldNamingParamName URL query parameter, that force to export only
// proto named fields (OnlyProtoFieldNames URL query parameter value) or JSON named fields (OnlyJSONFieldNames
// URL query parameter value).
func (p *Plugin) jsonSchemaHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		res, err := buildJsonSchema(req.URL.Query())
		if err != nil {
			if res != nil {
				errMsg := fmt.Sprintf("failed generate JSON schema: %v (%v)\n", res.Error, err)
				p.Log.Error(internalErrorLogPrefix + errMsg)
				p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
				return
			}
			p.internalError("", err, w, formatter)
			return
		}

		// extract json schema
		// (protoc_plugin.CodeGeneratorResponse could have cut the file content into multiple pieces
		// for performance optimization (due to godoc), but we know that all pieces are only one file
		// due to requesting one file -> join all content together)
		var sb strings.Builder
		for _, file := range res.File {
			sb.WriteString(file.GetContent())
		}

		// writing response
		// (jsonschema is in raw form (string) and non of the available format renders supports raw data output
		// with customizable content type setting in header -> custom handling)
		w.Header().Set(render.ContentType, render.ContentJSON+"; charset=UTF-8")
		w.Write([]byte(sb.String())) // will also call WriteHeader(http.StatusOK) automatically
	}
}

func buildJsonSchema(query url.Values) (*pluginpb.CodeGeneratorResponse, error) {
	logging.Debugf("=======================================================")
	logging.Debugf(" BUILDING JSON SCHEMA ")
	logging.Debugf("=======================================================")

	// create FileDescriptorProto for dynamic Config holding all VPP-Agent configuration
	knownModels, err := client.LocalClient.KnownModels("config") // locally registered models
	if err != nil {
		return nil, fmt.Errorf("can't get registered models: %w", err)
	}
	config, err := client.NewDynamicConfig(knownModels)
	if err != nil {
		return nil, fmt.Errorf("can't create dynamic config: %w", err)
	}
	dynConfigFileDescProto := protodesc.ToFileDescriptorProto(config.ProtoReflect().Descriptor().ParentFile())

	// create list of all FileDescriptorProtos (imports should be before converted proto file -> dynConfig is last)
	fileDescriptorProtos := allFileDescriptorProtos(knownModels)
	fileDescriptorProtos = append(fileDescriptorProtos, dynConfigFileDescProto)

	// creating input for protoc's plugin (code extracted in plugins/restapi/jsonschema) that can convert
	// FileDescriptorProtos to JSONSchema
	params := []string{
		"messages=[Dynamic_config]",      // targeting only the main config message (proto file has also other messages)
		"disallow_additional_properties", // additional unknown json fields makes configuration applying fail
	}
	fieldNamesConverterParam := "proto_and_json_fieldnames" // create proto and json named fields by default
	if fieldNames, found := query[URLFieldNamingParamName]; found && len(fieldNames) > 0 {
		// converting REST API request params to 3rd party tool params
		switch fieldNames[0] {
		case OnlyProtoFieldNames:
			fieldNamesConverterParam = ""
		case OnlyJSONFieldNames:
			fieldNamesConverterParam = "json_fieldnames"
		}
	}
	if fieldNamesConverterParam != "" {
		params = append(params, fieldNamesConverterParam)
	}
	paramsStr := strings.Join(params, ",")
	cgReq := &pluginpb.CodeGeneratorRequest{
		ProtoFile:       fileDescriptorProtos,
		FileToGenerate:  []string{dynConfigFileDescProto.GetName()},
		Parameter:       &paramsStr,
		CompilerVersion: nil, // compiler version is not need in this protoc plugin
	}
	cgReqMarshalled, err := proto.Marshal(cgReq)
	if err != nil {
		return nil, fmt.Errorf("can't proto marshal CodeGeneratorRequest: %w", err)
	}

	logging.Debugf("-------------------------------------------------------")
	logging.Debugf(" CONVERTING SCHEMA ")
	logging.Debugf("-------------------------------------------------------")

	// use JSON schema converter and handle error cases
	logging.Debug("Processing code generator request")
	protoConverter := converter.New(logrus.DefaultLogger().Logger)
	res, err := protoConverter.ConvertFrom(bytes.NewReader(cgReqMarshalled))
	if err != nil {
		if res == nil {
			// p.internalError("failed to read registered model configuration input", err, w, formatter)
			return nil, fmt.Errorf("failed to read registered model configuration input: %w", err)
		}
		return res, err
	}

	return res, nil
}

// allImports retrieves all imports from given FileDescriptor including transitive imports (import
// duplication can occur)
func allImports(fileDesc protoreflect.FileDescriptor) []protoreflect.FileDescriptor {
	result := make([]protoreflect.FileDescriptor, 0)
	imports := fileDesc.Imports()
	for i := 0; i < imports.Len(); i++ {
		currentImport := imports.Get(i).FileDescriptor
		result = append(result, currentImport)
		result = append(result, allImports(currentImport)...)
	}
	return result
}

// allFileDescriptorProtos retrieves all FileDescriptorProtos related to given models (including
// all imported proto files)
func allFileDescriptorProtos(knownModels []*client.ModelInfo) []*descriptorpb.FileDescriptorProto {
	// extract all FileDescriptors for given known models (including direct and transitive file imports)
	fileDescriptors := make(map[string]protoreflect.FileDescriptor) // using map for deduplication
	for _, knownModel := range knownModels {
		protoFile := knownModel.MessageDescriptor.ParentFile()
		fileDescriptors[protoFile.Path()] = protoFile
		for _, importProtoFile := range allImports(protoFile) {
			fileDescriptors[importProtoFile.Path()] = importProtoFile
		}
	}

	// convert retrieved FileDescriptors to FileDescriptorProtos
	fileDescriptorProtos := make([]*descriptorpb.FileDescriptorProto, 0, len(knownModels))
	for _, fileDescriptor := range fileDescriptors {
		fileDescriptorProtos = append(fileDescriptorProtos, protodesc.ToFileDescriptorProto(fileDescriptor))
	}
	return fileDescriptorProtos
}

// versionHandler returns version of Agent.
func (p *Plugin) versionHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ver := types.Version{
			App:       version.App(),
			Version:   version.Version(),
			GitCommit: version.GitCommit(),
			GitBranch: version.GitBranch(),
			BuildUser: version.BuildUser(),
			BuildHost: version.BuildHost(),
			BuildTime: version.BuildTime(),
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		}
		p.logError(formatter.JSON(w, http.StatusOK, ver))
	}
}

// validationHandler validates yaml configuration for VPP-Agent. This is the same configuration as used
// in agentctl configuration get/update.
func (p *Plugin) validationHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// reading input data (yaml-formatted dynamic config containing all VPP-Agent configuration)
		yamlBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			p.internalError("can't read request body", err, w, formatter)
			return
		}

		// get empty dynamic Config able to hold all VPP-Agent configuration
		knownModels, err := client.LocalClient.KnownModels("config") // locally registered models
		if err != nil {
			p.internalError("can't get registered models", err, w, formatter)
			return
		}
		config, err := client.NewDynamicConfig(knownModels)
		if err != nil {
			p.internalError("can't create dynamic config", err, w, formatter)
			return
		}

		// filling dynamically created config with data from request body
		// (=syntax check of data + prepare for further processing)
		bj, err := yaml2.YAMLToJSON(yamlBytes)
		if err != nil {
			p.internalError("can't convert yaml configuration "+
				"from request body to JSON", err, w, formatter)
			return
		}
		err = protojson.Unmarshal(bj, config)
		if err != nil {
			p.internalError("can't unmarshall string input data "+
				"into dynamically created config", err, w, formatter)
			return
		}

		// extracting proto messages from dynamically created config structure
		configMessages, err := client.DynamicConfigExport(config)
		if err != nil {
			p.internalError("can't extract single proto message "+
				"from one dynamic config to validate them per proto message", err, w, formatter)
			return
		}

		// run Descriptor validators on config messages
		err = p.KVScheduler.ValidateSemantically(configMessages)
		if err != nil {
			if validationErrors, ok := err.(*kvscheduler.InvalidMessagesError); ok {
				convertedValidationErrors := p.ConvertValidationErrorOutput(validationErrors, knownModels, config)
				p.logError(formatter.JSON(w, http.StatusBadRequest, convertedValidationErrors))
				return
			}
			p.internalError("can't validate data", err, w, formatter)
			return
		}
		p.logError(formatter.JSON(w, http.StatusOK, struct{}{}))
	}
}

// ConvertValidationErrorOutput converts kvscheduler.ValidateSemantically(...) output to REST API output
func (p *Plugin) ConvertValidationErrorOutput(validationErrors *kvscheduler.InvalidMessagesError, knownModels []*models.ModelInfo, config *dynamicpb.Message) []interface{} {
	// create helper mapping
	nameToModel := make(map[protoreflect.FullName]*models.ModelInfo)
	for _, knownModel := range knownModels {
		nameToModel[knownModel.MessageDescriptor.FullName()] = knownModel
	}

	// define types for REST API output (could use map, but struct hold field ordering within each validation error)
	type singleConfig struct {
		Path  string `json:"path"`
		Error string `json:"error"`
	}
	type repeatedConfig struct {
		Path            string `json:"path"`
		Error           string `json:"error"`
		ErrorConfigPart string `json:"error_config_part"`
	}
	type singleConfigDerivedValue struct {
		Path                   string `json:"path"`
		Error                  string `json:"error"`
		ErrorDerivedConfigPart string `json:"error_derived_config_part"`
	}
	type repeatedConfigDerivedValue struct {
		Path                   string `json:"path"`
		Error                  string `json:"error"`
		ErrorDerivedConfigPart string `json:"error_derived_config_part"`
		ErrorConfigPart        string `json:"error_config_part"`
	}

	// convert each validation error to REST API output (data filled structs defined above)
	convertedValidationErrors := make([]interface{}, 0, len(validationErrors.MessageErrors()))
	for _, messageError := range validationErrors.MessageErrors() {
		// get yaml names of messages/fields on path to configuration with error
		nonDerivedMessage := messageError.Message()
		if messageError.ParentMessage() != nil {
			nonDerivedMessage = messageError.ParentMessage()
		}
		messageModel := nameToModel[nonDerivedMessage.ProtoReflect().Descriptor().FullName()]
		groupFieldName := client.DynamicConfigGroupFieldNaming(messageModel)
		modelFieldProtoName, modelFieldName := client.DynamicConfigKnownModelFieldNaming(messageModel)
		invalidMessageFields := messageError.InvalidFields()
		invalidMessageFieldsStr := invalidMessageFields[0]
		if invalidMessageFieldsStr == "" {
			invalidMessageFieldsStr = "<unknown field>"
		}
		if len(invalidMessageFields) > 1 {
			invalidMessageFieldsStr = fmt.Sprintf("[%s]", strings.Join(invalidMessageFields, ","))
		}

		// attempt to guess yaml field by name from KVDescriptor.Validate (there is no enforcing of correct field name)
		if len(invalidMessageFields) == 1 { // guessing only for single field references
			// disassemble field reference (can refer to inner message field), guess the yaml name for each
			// segment and assemble the path again
			fieldPath := strings.Split(invalidMessageFieldsStr, ".")
			messageDesc := messageError.Message().ProtoReflect().Descriptor()
			for i := range fieldPath {
				// find current field path segment in proto message fields
				fieldDesc := messageDesc.Fields().ByName(protoreflect.Name(fieldPath[i]))
				if fieldDesc == nil {
					fieldDesc = messageDesc.Fields().ByJSONName(fieldPath[i])
				}
				if fieldDesc == nil {
					break // name guessing failed -> can't continue and replace other field path segments
				}

				// replacing messageError name with name used in yaml
				fieldPath[i] = fieldDesc.JSONName()

				// updating message descriptor as we move through field path
				messageDesc = fieldDesc.Message()
			}
			invalidMessageFieldsStr = strings.Join(fieldPath, ".")
		}

		// compute cardinality of field (in configGroup) referring to configuration with error
		cardinality := protoreflect.Optional
		if configGroupField := config.ProtoReflect().Descriptor().Fields().
			ByName(protoreflect.Name(groupFieldName)); configGroupField != nil {
			modelField := configGroupField.Message().Fields().ByName(protoreflect.Name(modelFieldProtoName))
			if modelField != nil {
				cardinality = modelField.Cardinality()
			}
		}

		// compute string representation of derived value configuration (yaml is preferred even when there is
		// no direct yaml configuration for derived value)
		var parentConfigPart string
		if messageError.ParentMessage() != nil {
			parentConfigPart = prototext.Format(messageError.ParentMessage())
			json, err := protojson.Marshal(messageError.ParentMessage())
			if err == nil {
				parentConfigPart = string(json)
				b, err := yaml2.JSONToYAML(json)
				if err == nil {
					parentConfigPart = string(b)
				}
			}
		}

		// compute again the string representation of error configuration (yaml is preferred)
		// (no original reference to REST API string is remembered -> computing it from proto message)
		configPart := prototext.Format(messageError.Message())
		json, err := protojson.Marshal(messageError.Message())
		if err == nil {
			configPart = string(json)
			b, err := yaml2.JSONToYAML(json)
			if err == nil {
				configPart = string(b)
			}
		}

		// fill correct struct for REST API output
		var convertedValidationError interface{}
		if cardinality == protoreflect.Repeated {
			if parentConfigPart == "" {
				convertedValidationError = repeatedConfig{
					Path: fmt.Sprintf("%s.%s*.%s",
						groupFieldName, modelFieldName, invalidMessageFieldsStr),
					Error:           messageError.ValidationError().Error(),
					ErrorConfigPart: configPart,
				}
			} else { // problem in derived values
				convertedValidationError = repeatedConfigDerivedValue{
					Path: fmt.Sprintf("%s.%s*.[derivedConfiguration].%s",
						groupFieldName, modelFieldName, invalidMessageFieldsStr),
					Error:                  messageError.ValidationError().Error(),
					ErrorConfigPart:        parentConfigPart,
					ErrorDerivedConfigPart: configPart,
				}
			}
		} else {
			if parentConfigPart == "" {
				convertedValidationError = singleConfig{
					Path:  fmt.Sprintf("%s.%s.%s", groupFieldName, modelFieldName, invalidMessageFieldsStr),
					Error: messageError.ValidationError().Error(),
				}
			} else { // problem in derived values
				convertedValidationError = singleConfigDerivedValue{
					Path: fmt.Sprintf("%s.%s.[derivedConfiguration].%s",
						groupFieldName, modelFieldName, invalidMessageFieldsStr),
					Error:                  messageError.ValidationError().Error(),
					ErrorDerivedConfigPart: configPart,
				}
			}
		}

		convertedValidationErrors = append(convertedValidationErrors, convertedValidationError)
	}
	return convertedValidationErrors
}

// configurationGetHandler returns NB configuration of VPP-Agent in yaml format as used by agentctl.
func (p *Plugin) configurationGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// create dynamically config that can hold all locally known models (to use only configurator.Config is
		// not enough as VPP-Agent could be used as library and additional model could be registered and
		// these models are unknown for configurator.Config)
		knownModels, err := client.LocalClient.KnownModels("config")
		if err != nil {
			p.internalError("failed to get registered models", err, w, formatter)
			return
		}
		config, err := client.NewDynamicConfig(knownModels)
		if err != nil {
			p.internalError("failed to create empty "+
				"all-config proto message dynamically", err, w, formatter)
			return
		}

		// retrieve data into config
		if err := client.LocalClient.GetConfig(config); err != nil {
			p.internalError("failed to retrieve all configuration "+
				"into dynamic all-config proto message", err, w, formatter)
			return
		}

		// convert data-filled config into yaml
		jsonBytes, err := protojson.Marshal(config)
		if err != nil {
			p.internalError("failed to convert retrieved configuration "+
				"to intermediate json output", err, w, formatter)
			return
		}
		var yamlObj interface{}
		if err := yaml.UnmarshalWithOptions(jsonBytes, &yamlObj, yaml.UseOrderedMap()); err != nil {
			p.internalError("failed to unmarshall intermediate json formatted "+
				"retrieved configuration to yaml object", err, w, formatter)
			return
		}
		yamlBytes, err := yaml.Marshal(yamlObj)
		if err != nil {
			p.internalError("failed to marshal retrieved configuration to yaml output", err, w, formatter)
			return
		}

		// writing response (no YAML support in formatters -> custom handling)
		w.Header().Set(render.ContentType, YamlContentType+"; charset=UTF-8")
		w.Write(yamlBytes) // will also call WriteHeader(http.StatusOK) automatically
	}
}

func (p *Plugin) internalError(additionalErrorMsgPrefix string, err error, w http.ResponseWriter,
	formatter *render.Render) {
	errMsg := fmt.Sprintf("%s: %v\n", additionalErrorMsgPrefix, err)
	p.Log.Error(internalErrorLogPrefix + errMsg)
	p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
}

// configurationUpdateHandler creates/updates NB configuration of VPP-Agent. The input configuration should be
// in yaml format as used by agentctl.
func (p *Plugin) configurationUpdateHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// create dynamically config that can hold input yaml configuration
		knownModels, err := client.LocalClient.KnownModels("config")
		if err != nil {
			p.internalError("failed to get registered models", err, w, formatter)
			return
		}
		config, err := client.NewDynamicConfig(knownModels)
		if err != nil {
			p.internalError("can't create all-config proto message dynamically", err, w, formatter)
			return
		}

		// reading input data (yaml-formatted dynamic config containing all VPP-Agent configuration)
		yamlBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			p.internalError("can't read request body", err, w, formatter)
			return
		}

		// filling dynamically created config with data
		bj, err := yaml2.YAMLToJSON(yamlBytes)
		if err != nil {
			p.internalError("converting yaml input to json failed", err, w, formatter)
			return
		}
		err = protojson.Unmarshal(bj, config)
		if err != nil {
			p.internalError("can't unmarshall input yaml data "+
				"into dynamically created config", err, w, formatter)
			return
		}

		// extracting proto messages from dynamically created config structure
		// (further processing needs single proto messages and not one big hierarchical config)
		configMessages, err := client.DynamicConfigExport(config)
		if err != nil {
			p.internalError("can't extract single configuration proto messages "+
				"from one big configuration proto message", err, w, formatter)
			return
		}

		// convert config messages to input for p.Dispatcher.PushData(...)
		var configKVPairs []orchestrator.KeyVal
		for _, configMessage := range configMessages {
			// convert config message from dynamic to statically-generated proto message (if possible)
			// (this is needed for later processing of message - generated KVDescriptor adapters cast
			// to statically-generated proto message and fail with dynamicpb.Message proto messages)
			dynamicMessage, ok := configMessage.(*dynamicpb.Message)
			if !ok { // should not happen, but checking anyway
				errMsg := fmt.Sprintf("proto message is expected to be "+
					"dynamicpb.Message (message=%s)\n", configMessage)
				p.Log.Error(internalErrorLogPrefix + errMsg)
				p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
				return
			}
			model, err := models.GetModelFor(dynamicMessage)
			if err != nil {
				errMsg := fmt.Sprintf("can't get model for dynamic message "+
					"due to: %v (message=%v)", err, dynamicMessage)
				p.Log.Error(internalErrorLogPrefix + errMsg)
				p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
				return
			}
			var message proto.Message
			if _, isRemoteModel := model.(*models.RemotelyKnownModel); isRemoteModel {
				// message is retrieved from localclient but it has remotely known model => it is the proxy
				// models in local model registry => can't convert it to generated message due to unknown
				// generated message go type (to use reflection to create it), however the processing of proxy
				// models is different so it might no need type casting fix at all -> using the only thing
				// available, the dynamic message
				message = dynamicMessage
			} else { // message has locally known model -> using generated proto message
				message, err = models.DynamicLocallyKnownMessageToGeneratedMessage(dynamicMessage)
				if err != nil {
					errMsg := fmt.Sprintf("can't convert dynamic message to statically generated message "+
						"due to: %v (dynamic message=%v)", err, dynamicMessage)
					p.Log.Error(internalErrorLogPrefix + errMsg)
					p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
					return
				}
			}

			// extract model key
			key, err := models.GetKey(message)
			if err != nil {
				errMsg := fmt.Sprintf("can't get model key for dynamic message "+
					"due to: %v (dynamic message=%v)", err, dynamicMessage)
				p.Log.Error(internalErrorLogPrefix + errMsg)
				p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
			}

			// build key-value pair structure
			configKVPairs = append(configKVPairs, orchestrator.KeyVal{
				Key: key,
				Val: message,
			})
		}

		// create context for data push
		ctx := context.Background()
		// // FullResync
		if _, found := req.URL.Query()[URLReplaceParamName]; found {
			ctx = kvs.WithResync(ctx, kvs.FullResync, true)
		}
		// // Note: using "grpc" data source so that 'agentctl update --replace' can also work with this data
		// // ('agentctl update' can change data also from non-grpc data sources, but
		// // 'agentctl update --replace' (=resync) can't)
		ctx = contextdecorator.DataSrcContext(ctx, "grpc")

		// config data pushed into VPP-Agent
		_, err = p.Dispatcher.PushData(ctx, configKVPairs, nil)
		if err != nil {
			p.internalError("can't push data into vpp-agent", err, w, formatter)
			return
		}

		p.logError(formatter.JSON(w, http.StatusOK, struct{}{}))
	}
}

// telemetryHandler - returns various telemetry data
func (p *Plugin) telemetryHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		type cmdOut struct {
			Command string
			Output  interface{}
		}
		var cmdOuts []cmdOut

		var runCmd = func(command string) {
			out, err := p.vpeHandler.RunCli(context.TODO(), command)
			if err != nil {
				errMsg := fmt.Sprintf("500 Internal server error: sending command failed: %v\n", err)
				p.Log.Error(errMsg)
				p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
				return
			}
			cmdOuts = append(cmdOuts, cmdOut{
				Command: command,
				Output:  out,
			})
		}

		runCmd("show node counters")
		runCmd("show runtime")
		runCmd("show buffers")
		runCmd("show memory")
		runCmd("show ip fib")
		runCmd("show ip6 fib")

		p.logError(formatter.JSON(w, http.StatusOK, cmdOuts))
	}
}

// telemetryMemoryHandler - returns various telemetry data
func (p *Plugin) telemetryMemoryHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		info, err := p.teleHandler.GetMemory(context.TODO())
		if err != nil {
			errMsg := fmt.Sprintf("500 Internal server error: sending command failed: %v\n", err)
			p.Log.Error(errMsg)
			p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
			return
		}

		p.logError(formatter.JSON(w, http.StatusOK, info))
	}
}

// telemetryHandler - returns various telemetry data
func (p *Plugin) telemetryRuntimeHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		runtimeInfo, err := p.teleHandler.GetRuntimeInfo(context.TODO())
		if err != nil {
			errMsg := fmt.Sprintf("500 Internal server error: sending command failed: %v\n", err)
			p.Log.Error(errMsg)
			p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
			return
		}

		p.logError(formatter.JSON(w, http.StatusOK, runtimeInfo))
	}
}

// telemetryHandler - returns various telemetry data
func (p *Plugin) telemetryNodeCountHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		nodeCounters, err := p.teleHandler.GetNodeCounters(context.TODO())
		if err != nil {
			errMsg := fmt.Sprintf("500 Internal server error: sending command failed: %v\n", err)
			p.Log.Error(errMsg)
			p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
			return
		}

		p.logError(formatter.JSON(w, http.StatusOK, nodeCounters))
	}
}

// configuratorStatsHandler - returns stats for Configurator
func (p *Plugin) configuratorStatsHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		stats := configurator.GetStats()
		if stats == nil {
			p.logError(formatter.JSON(w, http.StatusOK, "Configurator stats not available"))
			return
		}

		p.logError(formatter.JSON(w, http.StatusOK, stats))
	}
}

// logError logs non-nil errors from JSON formatter
func (p *Plugin) logError(err error) {
	if err != nil {
		p.Log.Error(err)
	}
}
