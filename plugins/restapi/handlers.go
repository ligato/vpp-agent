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
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	protoc_plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/unrolled/render"

	"go.ligato.io/cn-infra/v2/logging/logrus"
	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	"go.ligato.io/vpp-agent/v3/pkg/version"
	"go.ligato.io/vpp-agent/v3/plugins/configurator"
	"go.ligato.io/vpp-agent/v3/plugins/restapi/jsonschema/converter"
	"go.ligato.io/vpp-agent/v3/plugins/restapi/resturl"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const internalErrorLogPrefix = "500 Internal server error: "

var (
	// ErrHandlerUnavailable represents error returned when particular
	// handler is not available
	ErrHandlerUnavailable = errors.New("Handler is not available")
)

func (p *Plugin) registerInfoHandlers() {
	p.HTTPHandlers.RegisterHTTPHandler(resturl.Version, p.versionHandler, GET)
	p.HTTPHandlers.RegisterHTTPHandler(resturl.JSONSchema, p.jsonSchemaHandler, GET)
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
func (p *Plugin) jsonSchemaHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// create FileDescriptorProto for dynamic Config holding all VPP-Agent configuration
		knownModels, err := client.LocalClient.KnownModels("config") // locally registered models
		if err != nil {
			errMsg := fmt.Sprintf("can't get registered models: %v\n", err)
			p.Log.Error(internalErrorLogPrefix + errMsg)
			p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
			return
		}
		config, err := client.NewDynamicConfig(knownModels)
		if err != nil {
			errMsg := fmt.Sprintf("can't create dynamic config due to: %v\n", err)
			p.Log.Error(internalErrorLogPrefix + errMsg)
			p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
			return
		}
		dynConfigFileDescProto := protodesc.ToFileDescriptorProto(config.ProtoReflect().Descriptor().ParentFile())

		// create list of all FileDescriptorProtos (imports should be before converted proto file -> dynConfig is last)
		fileDescriptorProtos := allFileDescriptorProtos(knownModels)
		fileDescriptorProtos = append(fileDescriptorProtos, dynConfigFileDescProto)

		// creating input for protoc's plugin (code extracted in plugins/restapi/jsonschema) that can convert
		// FileDescriptorProtos to JSONSchema
		params := "messages=[Dynamic_config]" + // targeting only the main config message (proto file has also other messages)
			",disallow_additional_properties" + // additional unknown json fields makes configuration applying fail
			",proto_and_json_fieldnames" // allow also proto names for fields
		cgReq := &protoc_plugin.CodeGeneratorRequest{
			ProtoFile:       fileDescriptorProtos,
			FileToGenerate:  []string{dynConfigFileDescProto.GetName()},
			Parameter:       &params,
			CompilerVersion: nil, // compiler version is not need in this protoc plugin
		}
		cgReqMarshalled, err := proto.Marshal(cgReq)
		if err != nil {
			errMsg := fmt.Sprintf("can't proto marshal CodeGeneratorRequest: %v\n", err)
			p.Log.Error(internalErrorLogPrefix + errMsg)
			p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
			return
		}

		// use JSON schema converter and handle error cases
		p.Log.Debug("Processing code generator request")
		protoConverter := converter.New(logrus.DefaultLogger().StandardLogger())
		res, err := protoConverter.ConvertFrom(bytes.NewReader(cgReqMarshalled))
		if err != nil {
			if res == nil {
				errMsg := fmt.Sprintf("failed to read registered model configuration input: %v\n", err)
				p.Log.Error(internalErrorLogPrefix + errMsg)
				p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
				return
			}
			errMsg := fmt.Sprintf("failed generate JSON schema: %v (%v)\n", res.Error, err)
			p.Log.Error(internalErrorLogPrefix + errMsg)
			p.logError(formatter.JSON(w, http.StatusInternalServerError, errMsg))
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
		version := types.Version{
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
		p.logError(formatter.JSON(w, http.StatusOK, version))
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
