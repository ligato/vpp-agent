// Copyright (c) 2017 Cisco and/or its affiliates.
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

package probe

import (
	"net/http"

	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/unrolled/render"
)

const (
	defaultPluginName string = "PROMETHEUS"

	// DefaultMetricsPath default Prometheus metrics URL
	DefaultMetricsPath string = "/metrics"
	// DefaultHealthPath default Prometheus health metrics URL
	DefaultHealthPath string = "/health"

	// Namespace namespace to use for Prometheus health metrics
	Namespace string = ""
	// Subsystem subsystem to use for Prometheus health metrics
	Subsystem string = ""
	// ServiceLabel label for service field
	ServiceLabel string = "service"
	// DependencyLabel label for dependency field
	DependencyLabel string = "dependency"
	// BuildVersionLabel label for build version field
	BuildVersionLabel string = "build_version"
	// BuildDateLabel label for build date field
	BuildDateLabel string = "build_date"

	// ServiceHealthName name of service health metric
	ServiceHealthName string = "service_health"

	// ServiceHealthHelp help text for service health metric
	// Adapt Ligato status code for now.
	// TODO: Consolidate with that from the "Common Container Telemetry" proposal.
	// ServiceHealthHelp    string = "The health of the ServiceLabel 0 = INIT, 1 = UP, 2 = DOWN, 3 = OUTAGE"
	ServiceHealthHelp string = "The health of the ServiceLabel 0 = INIT, 1 = OK, 2 = ERROR"

	// DependencyHealthName name of dependency health metric
	DependencyHealthName string = "service_dependency_health"

	// DependencyHealthHelp help text for dependency health metric
	// Adapt Ligato status code for now.
	// TODO: Consolidate with that from the "Common Container Telemetry" proposal.
	// DependencyHealthHelp string = "The health of the DependencyLabel 0 = INIT, 1 = UP, 2 = DOWN, 3 = OUTAGE"
	DependencyHealthHelp string = "The health of the DependencyLabel 0 = INIT, 1 = OK, 2 = ERROR"

	// ServiceInfoName name of service info metric
	ServiceInfoName string = "service_info"
	// ServiceInfoHelp help text for service info metric
	ServiceInfoHelp string = "Build info for the service.  Value is always 1, build info is in the tags."
)

// PrometheusPlugin struct holds all plugin-related data.
type PrometheusPlugin struct {
	Deps
	healthRegistry *prometheus.Registry
}

// Init may create a new (custom) instance of HTTP if the injected instance uses
// different HTTP port than requested.
func (p *PrometheusPlugin) Init() (err error) {

	p.healthRegistry = prometheus.NewRegistry()

	p.registerGauge(
		Namespace,
		Subsystem,
		ServiceHealthName,
		ServiceHealthHelp,
		prometheus.Labels{ServiceLabel: p.getServiceLabel()},
		p.getServiceHealth,
	)

	agentStatus := p.StatusCheck.GetAgentStatus()
	p.registerGauge(
		Namespace,
		Subsystem,
		ServiceInfoName,
		ServiceInfoHelp,
		prometheus.Labels{
			ServiceLabel:      p.getServiceLabel(),
			BuildVersionLabel: agentStatus.BuildVersion,
			BuildDateLabel:    agentStatus.BuildDate},
		func() float64 { return 1 },
	)

	return nil
}

// AfterInit registers HTTP handlers.
func (p *PrometheusPlugin) AfterInit() error {
	if p.HTTP != nil {
		if p.StatusCheck != nil {
			p.Log.Info("Starting Prometheus metrics handlers")
			p.HTTP.RegisterHTTPHandler(DefaultMetricsPath, p.metricsHandler, "GET")
			p.HTTP.RegisterHTTPHandler(DefaultHealthPath, p.healthMetricsHandler, "GET")
			p.Log.Infof("Serving %s on port %d", DefaultMetricsPath, p.HTTP.GetPort())
			p.Log.Infof("Serving %s on port %d", DefaultHealthPath, p.HTTP.GetPort())
		} else {
			p.Log.Info("Unable to register Prometheus metrics handlers, StatusCheck is nil")
		}
	} else {
		p.Log.Info("Unable to register Prometheus metrics handlers, HTTP is nil")
	}

	//TODO: Need improvement - instead of the exposing the map directly need to use in-memory mapping
	if p.PluginStatusCheck != nil {
		allPluginStatusMap := p.PluginStatusCheck.GetAllPluginStatus()
		for k, v := range allPluginStatusMap {
			p.Log.Infof("k=%v, v=%v, state=%v", k, v, v.State)
			p.registerGauge(
				Namespace,
				Subsystem,
				DependencyHealthName,
				DependencyHealthHelp,
				prometheus.Labels{
					ServiceLabel:    p.getServiceLabel(),
					DependencyLabel: k,
				},
				p.getDependencyHealth(k, v),
			)
		}
	} else {
		p.Log.Error("PluginStatusCheck is nil")
	}

	/*if p.PluginStatusCheck != nil {
		if p.PluginStatusCheck.GetPluginStatusMap() != nil {
			pluginStatusIdx := p.PluginStatusCheck.GetPluginStatusMap()
			allPluginNames := pluginStatusIdx.GetMapping().ListAllNames()
			for _, v := range allPluginNames {
				p.registerGauge(
					Namespace,
					Subsystem,
					DependencyHealthName,
					DependencyHealthHelp,
					prometheus.Labels{
						ServiceLabel:    agentName,
						DependencyLabel: v,
					},
					func() float64 {
						p.Log.Infof("DependencyHealth for Plugin %v", v)
						pluginStatus, ok := pluginStatusIdx.GetValue(v)
						if ok {
							p.Log.Infof("DependencyHealth: %v", float64(pluginStatus.State))
							return float64(pluginStatus.State)
						} else {
							p.Log.Info("DependencyHealth not found")
							return float64(-1)
						}
					},
				)
			}
		} else {
			p.Log.Error("Plugin map is nil")
		}
	} else {
		p.Log.Error("PluginStatusCheck is nil")
	}*/

	return nil
}

// Close shutdowns HTTP if a custom instance was created in Init().
func (p *PrometheusPlugin) Close() error {
	return nil
}

// metricsHandler handles Prometheus metrics collection.
func (p *PrometheusPlugin) metricsHandler(formatter *render.Render) http.HandlerFunc {
	return promhttp.Handler().ServeHTTP
}

// healthMetricsHandler handles custom health metrics for Prometheus.
func (p *PrometheusPlugin) healthMetricsHandler(formatter *render.Render) http.HandlerFunc {
	return promhttp.HandlerFor(p.healthRegistry, promhttp.HandlerOpts{}).ServeHTTP
}

// getServiceHealth returns agent health status
func (p *PrometheusPlugin) getServiceHealth() float64 {
	agentStatus := p.StatusCheck.GetAgentStatus()
	// Adapt Ligato status code for now.
	// TODO: Consolidate with that from the "Common Container Telemetry" proposal.
	health := float64(agentStatus.State)
	p.Log.Infof("ServiceHealth: %v", health)
	return health
}

// getDependencyHealth returns plugin health status
func (p *PrometheusPlugin) getDependencyHealth(pluginName string, pluginStatus *status.PluginStatus) func() float64 {
	p.Log.Infof("DependencyHealth for plugin %v: %v", pluginName, float64(pluginStatus.State))

	return func() float64 {
		health := float64(pluginStatus.State)
		depName := pluginName
		p.Log.Infof("Dependency Health %v: %v", depName, health)
		return health
	}
}

// registerGauge registers custom gauge with specific valueFunc to report status when invoked.
func (p *PrometheusPlugin) registerGauge(namespace string, subsystem string, name string, help string,
	labels prometheus.Labels, valueFunc func() float64) {
	gaugeName := name
	if subsystem != "" {
		gaugeName = subsystem + "_" + gaugeName
	}
	if namespace != "" {
		gaugeName = namespace + "_" + gaugeName
	}
	if err := p.healthRegistry.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			// Namespace, Subsystem, and Name are components of the fully-qualified
			// name of the Metric (created by joining these components with
			// "_"). Only Name is mandatory, the others merely help structuring the
			// name. Note that the fully-qualified name of the metric must be a
			// valid Prometheus metric name.
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,

			// Help provides information about this metric. Mandatory!
			//
			// Metrics with the same fully-qualified name must have the same Help
			// string.
			Help: help,

			// ConstLabels are used to attach fixed labels to this metric. Metrics
			// with the same fully-qualified name must have the same label names in
			// their ConstLabels.
			//
			// Note that in most cases, labels have a value that varies during the
			// lifetime of a process. Those labels are usually managed with a metric
			// vector collector (like CounterVec, GaugeVec, UntypedVec). ConstLabels
			// serve only special purposes. One is for the special case where the
			// value of a label does not change during the lifetime of a process,
			// e.g. if the revision of the running binary is put into a
			// label. Another, more advanced purpose is if more than one Collector
			// needs to collect Metrics with the same fully-qualified name. In that
			// case, those Metrics must differ in the values of their
			// ConstLabels. See the Collector examples.
			//
			// If the value of a label never changes (not even between binaries),
			// that label most likely should not be a label at all (but part of the
			// metric name).
			ConstLabels: labels,
		},
		valueFunc,
	)); err == nil {
		p.Log.Infof("GaugeFunc('%s') registered.", gaugeName)
	} else {
		p.Log.Errorf("GaugeFunc('%s') registration failed: %s", gaugeName, err)
	}
}

// String returns plugin name if it was injected, defaultPluginName otherwise.
func (p *PrometheusPlugin) String() string {
	if len(string(p.PluginName)) > 0 {
		return string(p.PluginName)
	}
	return defaultPluginName
}

func (p *PrometheusPlugin) getServiceLabel() string {
	serviceLabel := p.String()
	if p.Deps.ServiceLabel != nil {
		serviceLabel = p.Deps.ServiceLabel.GetAgentLabel()
	}
	return serviceLabel
}
