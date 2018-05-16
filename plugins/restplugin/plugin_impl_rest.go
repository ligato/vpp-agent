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

package restplugin

import (
	"fmt"
	"time"

	"github.com/ligato/cn-infra/flavors/local"
	prom "github.com/ligato/cn-infra/rpc/prometheus"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	swIndexVarName = "swindex"
)

// RESTAPIPlugin - registers VPP REST API Plugin
type RESTAPIPlugin struct {
	Deps

	gaugeVecs    map[string]*prometheus.GaugeVec
	runtimeStats map[string]*runtimeStats
	indexItems   []indexItem
}

// Deps - dependencies of RESTAPIPlugin
type Deps struct {
	local.PluginInfraDeps
	HTTPHandlers rest.HTTPHandlers
	GoVppmux     govppmux.API
	Prometheus   prom.API
}

type indexItem struct {
	Name string
	Path string
}

// Init - initializes the RESTAPIPlugin
func (plugin *RESTAPIPlugin) Init() (err error) {
	plugin.indexItems = []indexItem{
		{Name: "Interfaces", Path: "/interfaces"},
		{Name: "Bridge domains", Path: "/bridgedomains"},
		{Name: "L2Fibs", Path: "/l2fibs"},
		{Name: "XConnectorPairs", Path: "/xconnectpairs"},
		{Name: "Static routes", Path: "/staticroutes"},
		{Name: "ACL IP", Path: "/acl/ip"},
		{Name: "Telemetry", Path: "/telemetry"},
	}

	plugin.setupPrometheus()

	return nil
}

// AfterInit - used to register HTTP handlers
func (plugin *RESTAPIPlugin) AfterInit() (err error) {
	plugin.Log.Debug("REST API Plugin is up and running")

	plugin.HTTPHandlers.RegisterHTTPHandler("/interfaces", plugin.interfacesGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/bridgedomains", plugin.bridgeDomainsGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/bridgedomainids", plugin.bridgeDomainIdsGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/l2fibs", plugin.fibTableEntriesGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/xconnectpairs", plugin.xconnectPairsGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/staticroutes", plugin.staticRoutesGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler(fmt.Sprintf("/acl/interface/{%s:[0-9]+}", swIndexVarName),
		plugin.interfaceACLGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/ip", plugin.ipACLPostHandler, "POST")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/ip", plugin.ipACLGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/acl/ip/example", plugin.exampleACLGetHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/command", plugin.commandHandler, "POST")
	plugin.HTTPHandlers.RegisterHTTPHandler("/telemetry", plugin.telemetryHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/telemetry/memory", plugin.telemetryMemoryHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/telemetry/runtime", plugin.telemetryRuntimeHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/telemetry/nodecount", plugin.telemetryNodeCountHandler, "GET")
	plugin.HTTPHandlers.RegisterHTTPHandler("/", plugin.indexHandler, "GET")

	return nil
}

// Close - used to clean up resources used by RESTAPIPlugin
func (plugin *RESTAPIPlugin) Close() (err error) {
	return nil
}

const (
	callsMetric          = "calls"
	vectorsMetric        = "vectors"
	suspendsMetric       = "suspends"
	clocksMetric         = "clocks"
	vectorsPerCallMetric = "vectorsPerCall"

	agentLabel       = "agent"
	runtimeItemLabel = "runtimeItem"
)

func (plugin *RESTAPIPlugin) setupPrometheus() error {
	plugin.gaugeVecs = make(map[string]*prometheus.GaugeVec)
	plugin.runtimeStats = make(map[string]*runtimeStats)

	for _, metric := range [][2]string{
		{callsMetric, "Number of calls"},
		{vectorsMetric, "Number of vectors"},
		{suspendsMetric, "Number of suspends"},
		{clocksMetric, "Number of clocks"},
		{vectorsPerCallMetric, "Number of vectors per call"},
	} {
		name := metric[0]
		plugin.gaugeVecs[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: metric[1],
			ConstLabels: prometheus.Labels{
				agentLabel: plugin.ServiceLabel.GetAgentLabel(),
			},
		}, []string{runtimeItemLabel})

	}

	// register created vectors to prometheus
	for name, metric := range plugin.gaugeVecs {
		if err := plugin.Prometheus.Register(prom.DefaultRegistry, metric); err != nil {
			plugin.Log.Errorf("failed to register %v metric: %v", name, err)
			return err
		}
	}

	ch, err := plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		plugin.Log.Errorf("Error creating channel: %v", err)
		return err
	}

	go func() {
		defer ch.Close()
		for {
			runtimeInfo, err := vppcalls.GetRuntimeInfo(ch)
			if err != nil {
				plugin.Log.Errorf("Sending command failed: %v", err)
				return
			}

			for _, item := range runtimeInfo.Items {
				stats, ok := plugin.runtimeStats[item.Name]
				if !ok {
					stats = &runtimeStats{
						itemName: item.Name,
						metrics:  map[string]prometheus.Gauge{},
					}

					// add gauges with corresponding labels into vectors
					for k, vec := range plugin.gaugeVecs {
						stats.metrics[k], err = vec.GetMetricWith(prometheus.Labels{
							runtimeItemLabel: item.Name,
						})
						if err != nil {
							plugin.Log.Error(err)
						}
					}
				}

				stats.metrics[callsMetric].Set(float64(item.Calls))
				stats.metrics[vectorsMetric].Set(float64(item.Vectors))
				stats.metrics[suspendsMetric].Set(float64(item.Suspends))
				stats.metrics[clocksMetric].Set(item.Clocks)
				stats.metrics[vectorsPerCallMetric].Set(item.VectorsCall)
			}
			time.Sleep(time.Second * 5)
		}
	}()

	return nil
}

type runtimeStats struct {
	itemName string
	metrics  map[string]prometheus.Gauge
}
