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

	govppapi "git.fd.io/govpp.git/api"
	log "go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
)

var (
	// FallbackToCli defines wether should telemetry handler
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
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	Objects   uint64 `json:"objects"`
	Used      uint64 `json:"used"`
	Total     uint64 `json:"total"`
	Free      uint64 `json:"free"`
	Reclaimed uint64 `json:"reclaimed"`
	Overhead  uint64 `json:"overhead"`
	Pages     uint64 `json:"pages"`
	PageSize  uint64 `json:"page_size"`
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

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "telemetry",
	HandlerAPI: (*TelemetryVppAPI)(nil),
})

type NewHandlerFunc func(govppapi.Channel) TelemetryVppAPI

// AddHandlerVersion registers vppcalls Handler for the given version.
func AddHandlerVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			return c.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return h(ch)
		},
	})
}

func NewHandler(c vpp.Client) (TelemetryVppAPI, error) {
	// Prefer using VPP stats API.
	if stats := c.Stats(); stats != nil {
		return NewTelemetryVppStats(stats), nil
	}
	v, err := Handler.GetCompatibleVersion(c)
	if err != nil {
		return nil, err
	}
	ch, err := c.NewAPIChannel()
	if err != nil {
		return nil, err
	}
	return v.New.(NewHandlerFunc)(ch), nil
}

func CompatibleTelemetryHandler(c vpp.Client) TelemetryVppAPI {
	// Prefer using VPP stats API.
	if stats := c.Stats(); stats != nil {
		return NewTelemetryVppStats(stats)
	}
	if FallbackToCli {
		if v := Handler.FindCompatibleVersion(c); v != nil {
			log.Info("falling back to parsing CLI output for telemetry")
			return v.NewHandler(c).(TelemetryVppAPI)
		}
		// no compatible version found
	}
	log.Warnf("stats connection not available for telemetry")
	return nil
}
