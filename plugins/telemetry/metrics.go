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

package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"

	"go.ligato.io/vpp-agent/v3/pkg/version"
)

var (
	currentVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "ligato",
		Subsystem: "build",
		Name:      "info",
		Help:      "Which version is running. 1 for 'agent_version' label with current version.",
	},
		[]string{"version", "revision", "build_date", "built_by"},
	)
	processMetrics = prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
		Namespace: "ligato",
	})
)

func init() {
	prometheus.MustRegister(currentVersion)
	prometheus.MustRegister(processMetrics)

	ver, rev, date := version.Data()
	currentVersion.WithLabelValues(ver, rev, date, version.BuiltBy()).Set(1)
}
