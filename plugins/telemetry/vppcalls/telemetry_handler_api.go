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

	govppapi "go.fd.io/govpp/api"
	log "go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
)

var (
	// FallbackToCli defines whether should telemetry handler
	// fallback to parsing stats from CLI output.
	FallbackToCli = false
)

// TelemetryVppAPI provides API for retrieving telemetry data from VPP.
type TelemetryVppAPI interface {
	GetSystemStats(context.Context) (*govppapi.SystemStats, error)
	GetMemory(context.Context) (*MemoryInfo, error)
	GetNodeCounters(context.Context) (*NodeCounterInfo, error)
	GetRuntimeInfo(context.Context) (*RuntimeInfo, error)
	GetBuffersInfo(context.Context) (*BuffersInfo, error)
	GetInterfaceStats(context.Context) (*govppapi.InterfaceStats, error)
	GetThreads(ctx context.Context) (*ThreadsInfo, error)
}

// MemoryInfo contains memory thread info.
type MemoryInfo struct {
	Threads []MemoryThread `json:"threads"`
}

// GetThreads is safe getter for threads,
func (i *MemoryInfo) GetThreads() []MemoryThread {
	if i == nil {
		return nil
	}
	return i.Threads
}

// MemoryThread represents single thread memory counters
type MemoryThread struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Size            uint64 `json:"size"`
	Pages           uint64 `json:"pages"`
	PageSize        uint64 `json:"page_size"`
	Total           uint64 `json:"total"`
	Used            uint64 `json:"used"`
	Free            uint64 `json:"free"`
	Trimmable       uint64 `json:"trimmable"`
	FreeChunks      uint64 `json:"free_chunks"`
	FreeFastbinBlks uint64 `json:"free_fastbin_blks"`
	MaxTotalAlloc   uint64 `json:"max_total_allocated"`
}

// NodeCounterInfo contains node counters info.
type NodeCounterInfo struct {
	Counters []NodeCounter `json:"counters"`
}

// GetCounters is safe getter for counters,
func (i *NodeCounterInfo) GetCounters() []NodeCounter {
	if i == nil {
		return nil
	}
	return i.Counters
}

// NodeCounter represents single node counter
type NodeCounter struct {
	Value uint64 `json:"value"`
	Node  string `json:"node"`
	Name  string `json:"name"`
}

// RuntimeInfo contains values returned from 'show runtime'
type RuntimeInfo struct {
	Threads []RuntimeThread `json:"threads"`
}

// GetThreads is safe getter for threads.
func (i *RuntimeInfo) GetThreads() []RuntimeThread {
	if i == nil {
		return nil
	}
	return i.Threads
}

// RuntimeThread represents single runtime thread
type RuntimeThread struct {
	ID                  uint          `json:"id"`
	Name                string        `json:"name"`
	Time                float64       `json:"time"`
	AvgVectorsPerNode   float64       `json:"avg_vectors_per_node"`
	LastMainLoops       uint64        `json:"last_main_loops"`
	VectorsPerMainLoop  float64       `json:"vectors_per_main_loop"`
	VectorLengthPerNode float64       `json:"vector_length_per_node"`
	VectorRatesIn       float64       `json:"vector_rates_in"`
	VectorRatesOut      float64       `json:"vector_rates_out"`
	VectorRatesDrop     float64       `json:"vector_rates_drop"`
	VectorRatesPunt     float64       `json:"vector_rates_punt"`
	Items               []RuntimeItem `json:"items"`
}

// RuntimeItem represents single runtime item
type RuntimeItem struct {
	Index          uint    `json:"index"`
	Name           string  `json:"name"`
	State          string  `json:"state"`
	Calls          uint64  `json:"calls"`
	Vectors        uint64  `json:"vectors"`
	Suspends       uint64  `json:"suspends"`
	Clocks         float64 `json:"clocks"`
	VectorsPerCall float64 `json:"vectors_per_call"`
}

// BuffersInfo contains values returned from 'show buffers'
type BuffersInfo struct {
	Items []BuffersItem `json:"items"`
}

// GetItems is safe getter for items,
func (i *BuffersInfo) GetItems() []BuffersItem {
	if i == nil {
		return nil
	}
	return i.Items
}

// BuffersItem represents single buffers item
type BuffersItem struct {
	ThreadID uint   `json:"thread_id"`
	Name     string `json:"name"`
	Index    uint   `json:"index"`
	Size     uint64 `json:"size"`
	Alloc    uint64 `json:"alloc"`
	Free     uint64 `json:"free"`
	NumAlloc uint64 `json:"num_alloc"`
	NumFree  uint64 `json:"num_free"`
}

// ThreadsInfo contains values returned form `show threads`
type ThreadsInfo struct {
	Items []ThreadsItem
}

// GetItems is safe getter for thread items
func (i *ThreadsInfo) GetItems() []ThreadsItem {
	if i == nil {
		return nil
	}
	return i.Items
}

// ThreadsItem represents single threads item
type ThreadsItem struct {
	Name      string `json:"name"`
	ID        uint32 `json:"id"`
	Type      string `json:"type"`
	PID       uint32 `json:"pid"`
	CPUID     uint32 `json:"cpuid"`
	Core      uint32 `json:"core"`
	CPUSocket uint32 `json:"cpu_socket"`
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "telemetry",
	HandlerAPI: (*TelemetryVppAPI)(nil),
})

type NewHandlerFunc func(vpp.Client) TelemetryVppAPI

// AddHandlerVersion registers vppcalls Handler for the given version.
func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			return c.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			return h(c)
		},
	})
}

// NewHandler returns the telemetry handler preferring the VPP stats API
// with CLI/binary API handler injected to retrieve data not included in
// stats. In case the stats API is not available, CLI handler is returned.
func NewHandler(c vpp.Client) (TelemetryVppAPI, error) {
	var compatibleHandler TelemetryVppAPI = nil
	v, err := Handler.GetCompatibleVersion(c)
	if err != nil {
		log.Warnf("compatible handler unavailable: %v", err)
	} else {
		compatibleHandler = v.NewHandler(c).(TelemetryVppAPI)
	}
	// Prefer the VPP stats API (even without the handler)
	if stats := c.Stats(); stats != nil {
		return NewTelemetryVppStats(stats, compatibleHandler), nil
	}
	if err != nil {
		return nil, err
	}
	return compatibleHandler, nil
}

// CompatibleTelemetryHandler returns the telemetry handler respecting
// VPP version. It returns the stats API handler when available, or
// fallbacks to the CLI when requested.
func CompatibleTelemetryHandler(c vpp.Client) TelemetryVppAPI {
	var compatibleHandler TelemetryVppAPI = nil
	v := Handler.FindCompatibleVersion(c)
	if v != nil {
		compatibleHandler = v.NewHandler(c).(TelemetryVppAPI)
	}
	if FallbackToCli && v != nil {
		log.Info("falling back to parsing CLI output for telemetry")
		return v.NewHandler(c).(TelemetryVppAPI)
	}
	if stats := c.Stats(); stats != nil {
		if v == nil {
			log.Warn("handler unavailable, functionality limited")
		}
		return NewTelemetryVppStats(stats, compatibleHandler)
	}
	// no compatible version found
	log.Warnf("stats connection not available for telemetry")
	return nil
}
