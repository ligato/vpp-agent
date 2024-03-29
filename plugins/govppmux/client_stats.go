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

package govppmux

import (
	"go.fd.io/govpp/adapter"
	govppapi "go.fd.io/govpp/api"
)

// ListStats returns all stats names
func (p *Plugin) ListStats(prefixes ...string) ([]adapter.StatIdentifier, error) {
	if p.statsAdapter == nil {
		return nil, nil
	}
	return p.statsAdapter.ListStats(prefixes...)
}

// DumpStats returns all stats with name, type and value
func (p *Plugin) DumpStats(prefixes ...string) ([]adapter.StatEntry, error) {
	if p.statsAdapter == nil {
		return nil, nil
	}
	return p.statsAdapter.DumpStats(prefixes...)
}

// GetSystemStats retrieves system statistics of the connected VPP instance like Vector rate, Input rate, etc.
func (p *Plugin) GetSystemStats(stats *govppapi.SystemStats) error {
	if p.statsConn == nil {
		return nil
	}
	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	return p.statsConn.GetSystemStats(stats)
}

// GetNodeStats retrieves a list of Node VPP counters (vectors, clocks, ...)
func (p *Plugin) GetNodeStats(stats *govppapi.NodeStats) error {
	if p.statsConn == nil {
		return nil
	}
	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	return p.statsConn.GetNodeStats(stats)
}

// GetInterfaceStats retrieves all counters related to the VPP interfaces
func (p *Plugin) GetInterfaceStats(stats *govppapi.InterfaceStats) error {
	if p.statsConn == nil {
		return nil
	}
	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	return p.statsConn.GetInterfaceStats(stats)
}

// GetErrorStats retrieves VPP error counters
func (p *Plugin) GetErrorStats(stats *govppapi.ErrorStats) error {
	if p.statsConn == nil {
		return nil
	}
	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	return p.statsConn.GetErrorStats(stats)
}

// GetBufferStats retrieves VPP error counters
func (p *Plugin) GetBufferStats(stats *govppapi.BufferStats) error {
	if p.statsConn == nil {
		return nil
	}
	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	return p.statsConn.GetBufferStats(stats)
}

// GetMemoryStats retrieves VPP memory info
func (p *Plugin) GetMemoryStats(stats *govppapi.MemoryStats) error {
	if p.statsConn == nil {
		return nil
	}
	p.statsMu.Lock()
	defer p.statsMu.Unlock()
	return p.statsConn.GetMemoryStats(stats)
}
