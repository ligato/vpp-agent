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
	govppapi "git.fd.io/govpp.git/api"
	log "github.com/ligato/cn-infra/logging"
)

// MemoryInfo contains values returned from 'show memory'
type MemoryInfo struct {
	Threads []MemoryThread `json:"threads"`
}

// MemoryThread represents single thread memory counters
type MemoryThread struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Objects   uint64 `json:"objects"`
	Used      uint64 `json:"used"`
	Total     uint64 `json:"total"`
	Free      uint64 `json:"free"`
	Reclaimed uint64 `json:"reclaimed"`
	Overhead  uint64 `json:"overhead"`
	Capacity  uint64 `json:"capacity"`
}

// NodeCounterInfo contains values returned from 'show node counters'
type NodeCounterInfo struct {
	Counters []NodeCounter `json:"counters"`
}

// NodeCounter represents single node counter
type NodeCounter struct {
	Count  uint64 `json:"count"`
	Node   string `json:"node"`
	Reason string `json:"reason"`
}

// RuntimeInfo contains values returned from 'show runtime'
type RuntimeInfo struct {
	Threads []RuntimeThread `json:"threads"`
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

type TelemetryVppAPI interface {
	GetMemory() (*MemoryInfo, error)
	GetNodeCounters() (*NodeCounterInfo, error)
	GetRuntimeInfo() (*RuntimeInfo, error)
	GetBuffersInfo() (*BuffersInfo, error)
}

var Versions = map[string]HandlerVersion{}

type HandlerVersion struct {
	Msgs []govppapi.Message
	New  func(govppapi.Channel) TelemetryVppAPI
}

func CompatibleTelemetryHandler(ch govppapi.Channel) TelemetryVppAPI {
	for ver, h := range Versions {
		if err := ch.CheckCompatiblity(h.Msgs...); err != nil {
			log.Debugf("version %s not compatible", ver)
			continue
		}
		log.Debug("found compatible version:", ver)
		return h.New(ch)
	}
	panic("no compatible version available")
}
