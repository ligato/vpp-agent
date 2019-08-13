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

package api

// SystemStats represents global system statistics.
type SystemStats struct {
	VectorRate     float64
	InputRate      float64
	LastUpdate     float64
	LastStatsClear float64
	Heartbeat      float64
}

// NodeStats represents per node statistics.
type NodeStats struct {
	Nodes []NodeCounters
}

// NodeCounters represents node counters.
type NodeCounters struct {
	NodeIndex uint32
	NodeName  string // requires VPP 19.04+

	Clocks   uint64
	Vectors  uint64
	Calls    uint64
	Suspends uint64
}

// InterfaceStats represents per interface statistics.
type InterfaceStats struct {
	Interfaces []InterfaceCounters
}

// InterfaceCounters represents interface counters.
type InterfaceCounters struct {
	InterfaceIndex uint32
	InterfaceName  string // requires VPP 19.04+

	RxPackets uint64
	RxBytes   uint64
	RxErrors  uint64
	TxPackets uint64
	TxBytes   uint64
	TxErrors  uint64

	RxUnicast     [2]uint64 // packets[0], bytes[1]
	RxMulticast   [2]uint64 // packets[0], bytes[1]
	RxBroadcast   [2]uint64 // packets[0], bytes[1]
	TxUnicastMiss [2]uint64 // packets[0], bytes[1]
	TxMulticast   [2]uint64 // packets[0], bytes[1]
	TxBroadcast   [2]uint64 // packets[0], bytes[1]

	Drops   uint64
	Punts   uint64
	IP4     uint64
	IP6     uint64
	RxNoBuf uint64
	RxMiss  uint64
}

// ErrorStats represents statistics per error counter.
type ErrorStats struct {
	Errors []ErrorCounter
}

// ErrorCounter represents error counter.
type ErrorCounter struct {
	CounterName string
	Value       uint64
}

// BufferStats represents statistics per buffer pool.
type BufferStats struct {
	Buffer map[string]BufferPool
}

// BufferPool represents buffer pool.
type BufferPool struct {
	PoolName  string
	Cached    float64
	Used      float64
	Available float64
}

// StatsProvider provides the methods for getting statistics.
type StatsProvider interface {
	GetSystemStats() (*SystemStats, error)
	GetNodeStats() (*NodeStats, error)
	GetInterfaceStats() (*InterfaceStats, error)
	GetErrorStats(names ...string) (*ErrorStats, error)
	GetBufferStats() (*BufferStats, error)
}
