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

package orchestrator

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync/local"
	"go.ligato.io/cn-infra/v2/rpc/grpc"

	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
)

// DefaultPlugin is default instance of Plugin
var DefaultPlugin = *NewPlugin()

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Plugin {
	p := &Plugin{}

	p.PluginName = "orchestrator"
	p.GRPC = &grpc.DefaultPlugin
	p.KVScheduler = &kvscheduler.DefaultPlugin
	p.Watcher = local.DefaultRegistry
	p.reflection = true

	for _, o := range opts {
		o(p)
	}
	p.PluginDeps.Setup()

	return p
}

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Plugin)

func UseReflection(enabled bool) Option {
	return func(p *Plugin) {
		p.reflection = enabled
	}
}

func EnabledGrpcMetrics() {
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc.UsePromMetrics(grpc_prometheus.DefaultServerMetrics)(&grpc.DefaultPlugin)
}
