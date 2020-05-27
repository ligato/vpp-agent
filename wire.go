//  Copyright (c) 2020 Cisco and/or its affiliates.
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

//+build wireinject

package vppagent

import (
	"context"

	"github.com/google/wire"
	cninfra "go.ligato.io/cn-infra/v2"
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/health/probe"
	"go.ligato.io/cn-infra/v2/health/statuscheck"
	"go.ligato.io/cn-infra/v2/rpc/grpc"
	"go.ligato.io/cn-infra/v2/rpc/prometheus"
	"go.ligato.io/cn-infra/v2/servicelabel"
)

//go:generate wire

func InjectDefaultVPPAgent(ctx context.Context) (a Agent, c func(), e error) {
	wire.Build(
		cninfra.WireDefaultAll,

		cninfra.CoreProviders,
		cninfra.ServerProviders,

		cninfra.WireDefaultConfig,
		cninfra.WireLogManager,
		cninfra.WirePrometheusProbe,

		resync.WireDefault,

		WireKVScheduler,
		WireOrchestrator,

		// Dataplane components with plugins
		WireDefaultNetAlloc,
		WireDefaultLinux,
		WireDefaultVPP,

		// Dataplane-related components
		WireGoVppMux,
		WireConfigurator,
		WireRestAPI,
		WireTelemetry,

		wire.Struct(new(Agent), "*"),
	)
	return
}

func InjectAgent(ctx context.Context, conf config.Config,
	core cninfra.Base, server cninfra.Server) (a Agent, c func(), e error) {
	wire.Build(
		//cninfra.InjectDefaultCore,
		//cninfra.InjectDefaultServer,

		/*cninfra.ProvideServiceLabelReaderAPI,
		cninfra.ProvideStatusCheckStatusReader,
		cninfra.ProvideStatusCheckPluginStatusWriter,
		cninfra.ProvideRestHTTPHandlers,
		cninfra.ProvideGrpcServer,*/

		wire.FieldsOf(new(cninfra.Base), "ServiceLabel"),
		wire.Bind(new(servicelabel.ReaderAPI), new(*servicelabel.Plugin)),

		wire.FieldsOf(new(cninfra.Base), "StatusCheck"),
		wire.Bind(new(statuscheck.StatusReader), new(*statuscheck.Plugin)),
		wire.Bind(new(statuscheck.PluginStatusWriter), new(*statuscheck.Plugin)),

		// Injecting HTTP & GRPC server as deps

		// 1. using helper function which returns proper API interface:
		cninfra.ProvideRestHTTPHandlers,
		//cninfra.GrpcServerProvider,

		// 2. or by manually extracting GRPC field and binding it to interface:
		wire.FieldsOf(new(cninfra.Server), "GRPC"),
		wire.Bind(new(grpc.Server), new(*grpc.Plugin)),
		//wire.FieldsOf(new(cninfra.Server), "HTTP"),
		//wire.Bind(new(rest.HTTPHandlers), new(*rest.Plugin)),

		//wire.FieldsOf(new(cninfra.Core), "LogRegistry"),
		//logmanager.WireDefault,
		cninfra.WireLogManager,

		resync.WireDefault,
		probe.WireDefault,
		prometheus.WireDefault,
		//cninfra.WirePrometheusProbe,

		WireKVScheduler,
		WireOrchestrator,

		// Dataplane components with plugins
		WireDefaultNetAlloc,
		WireDefaultLinux,
		WireDefaultVPP,

		// Dataplane related components
		WireGoVppMux,
		WireConfigurator,
		WireRestAPI,
		WireTelemetry,
		//wire.Bind(new(telemetry.InterfaceIndexProvider), new(*ifplugin.IfPlugin)),

		wire.Struct(new(Agent), "*"),
	)
	return
}
