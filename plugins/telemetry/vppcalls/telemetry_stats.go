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

package vppcalls

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	govppapi "go.fd.io/govpp/api"
)

// TelemetryStats is an implementation of TelemetryVppAPI that uses
// VPP stats API to retrieve the telemetry data.
type TelemetryStats struct {
	stats govppapi.StatsProvider
	// telemetry API helps with reading Memory/Threads data
	// (i.e. those who are not a part of the stats API or are not
	// implemented yet)
	telemetryAPI TelemetryVppAPI

	sysStats  govppapi.SystemStats
	ifStats   govppapi.InterfaceStats
	nodeStats govppapi.NodeStats
	errStats  govppapi.ErrorStats
	bufStats  govppapi.BufferStats
}

func NewTelemetryVppStats(stats govppapi.StatsProvider, teleApi TelemetryVppAPI) *TelemetryStats {
	return &TelemetryStats{
		stats:        stats,
		telemetryAPI: teleApi,
	}
}

// GetSystemStats retrieves system stats.
func (h *TelemetryStats) GetSystemStats(context.Context) (*govppapi.SystemStats, error) {
	err := h.stats.GetSystemStats(&h.sysStats)
	if err != nil {
		return nil, err
	}
	return &h.sysStats, nil
}

// GetMemory retrieves `show memory` info.
// todo switch to stats when memory data will be implemented
func (h *TelemetryStats) GetMemory(ctx context.Context) (*MemoryInfo, error) {
	if h.telemetryAPI == nil {
		return nil, fmt.Errorf("`GetMemory` unavailable, telemetry handler was not provided")
	}
	return h.telemetryAPI.GetMemory(ctx)
}

// GetThreads retrieves `show threads` info.
// todo switch to stats when threads data will be available
func (h *TelemetryStats) GetThreads(ctx context.Context) (*ThreadsInfo, error) {
	if h.telemetryAPI == nil {
		return nil, fmt.Errorf("`GetThreads` unavailable, telemetry handler was not provided")
	}
	return h.telemetryAPI.GetThreads(ctx)
}

// GetInterfaceStats retrieves interface stats.
func (h *TelemetryStats) GetInterfaceStats(context.Context) (*govppapi.InterfaceStats, error) {
	err := h.stats.GetInterfaceStats(&h.ifStats)
	if err != nil {
		return nil, err
	}
	return &h.ifStats, nil
}

// GetNodeCounters retrieves node counters info.
func (h *TelemetryStats) GetNodeCounters(ctx context.Context) (*NodeCounterInfo, error) {
	err := h.stats.GetErrorStats(&h.errStats)
	if err != nil {
		return nil, err
	}

	var counters []NodeCounter
	for _, c := range h.errStats.Errors {
		node, reason := SplitErrorName(c.CounterName)
		var valSum uint64 = 0
		// c.Values are per worker counters
		for _, val := range c.Values {
			valSum += val
		}
		counters = append(counters, NodeCounter{
			Value: valSum,
			Node:  node,
			Name:  reason,
		})
	}

	info := &NodeCounterInfo{
		Counters: counters,
	}

	return info, nil
}

// GetRuntimeInfo retrieves how runtime info.
func (h *TelemetryStats) GetRuntimeInfo(ctx context.Context) (*RuntimeInfo, error) {
	err := h.stats.GetNodeStats(&h.nodeStats)
	if err != nil {
		return nil, err
	}

	thread := RuntimeThread{
		Name: "ALL",
	}
	for _, node := range h.nodeStats.Nodes {
		vpc := 0.0
		if node.Vectors != 0 && node.Calls != 0 {
			vpc = float64(node.Vectors) / float64(node.Calls)
		}
		thread.Items = append(thread.Items, RuntimeItem{
			Index:          uint(node.NodeIndex),
			Name:           node.NodeName,
			Calls:          node.Calls,
			Vectors:        node.Vectors,
			Suspends:       node.Suspends,
			Clocks:         float64(node.Clocks),
			VectorsPerCall: vpc,
		})
	}

	threads := []RuntimeThread{
		thread,
	}

	return &RuntimeInfo{
		Threads: threads,
	}, nil
}

// GetBuffersInfo retrieves buffers info from VPP.
func (h *TelemetryStats) GetBuffersInfo(ctx context.Context) (*BuffersInfo, error) {
	err := h.stats.GetBufferStats(&h.bufStats)
	if err != nil {
		return nil, err
	}

	var items []BuffersItem

	for _, c := range h.bufStats.Buffer {
		items = append(items, BuffersItem{
			Name:  c.PoolName,
			Alloc: uint64(c.Used),
			Free:  uint64(c.Available),
			// Cached:  c.Cached,
		})
	}

	info := &BuffersInfo{
		Items: items,
	}

	return info, nil
}

var (
	errorNameLikeMemifRe   = regexp.MustCompile(`^[A-Za-z0-9-]+([0-9]+\/[0-9]+|pg\/stream)`)
	errorNameLikeGigabitRe = regexp.MustCompile(`^[A-Za-z0-9]+[0-9a-f]+(\/[0-9a-f]+){2}`)
)

func SplitErrorName(str string) (node, reason string) {
	parts := strings.Split(str, "/")
	switch len(parts) {
	case 1:
		return parts[0], ""
	case 2:
		return parts[0], parts[1]
	case 3:
		if strings.Contains(parts[1], " ") {
			return parts[0], strings.Join(parts[1:], "/")
		}
		if errorNameLikeMemifRe.MatchString(str) {
			return strings.Join(parts[:2], "/"), parts[2]
		}
	default:
		if strings.Contains(parts[2], " ") {
			return strings.Join(parts[:2], "/"), strings.Join(parts[2:], "/")
		}
		if errorNameLikeGigabitRe.MatchString(str) {
			return strings.Join(parts[:3], "/"), strings.Join(parts[3:], "/")
		}
	}
	return strings.Join(parts[:len(parts)-1], "/"), parts[len(parts)-1]
}
