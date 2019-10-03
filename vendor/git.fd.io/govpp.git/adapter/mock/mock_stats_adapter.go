// Copyright (c) 2019 Cisco and/or its affiliates.
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

// Package mock is an alternative VPP stats adapter aimed for unit/integration testing where the
// actual communication with VPP is not demanded.

package mock

import (
	"git.fd.io/govpp.git/adapter"
)

// implements StatsAPI
var _ adapter.StatsAPI = (*StatsAdapter)(nil)

// StatsAdapter simulates VPP stats socket from which stats can be read
type StatsAdapter struct {
	entries []adapter.StatEntry
	dir     *adapter.StatDir
}

// NewStatsAdapter returns a new mock stats adapter.
func NewStatsAdapter() *StatsAdapter {
	return &StatsAdapter{}
}

// Connect mocks client connection to the stats API.
func (a *StatsAdapter) Connect() error {
	return nil
}

// Disconnect mocks client connection termination.
func (a *StatsAdapter) Disconnect() error {
	return nil
}

// ListStats mocks name listing for all stats.
func (a *StatsAdapter) ListStats(patterns ...string) ([]string, error) {
	var statNames []string
	for _, stat := range a.entries {
		statNames = append(statNames, string(stat.Name))
	}
	return statNames, nil
}

// DumpStats mocks all stat entries dump.
func (a *StatsAdapter) DumpStats(patterns ...string) ([]adapter.StatEntry, error) {
	return a.entries, nil
}

func (a *StatsAdapter) PrepareDir(prefixes ...string) (*adapter.StatDir, error) {
	return a.dir, nil
}

func (a *StatsAdapter) UpdateDir(dir *adapter.StatDir) error {
	*dir = *a.dir
	return nil
}

// MockStats sets mocked stat entries to be returned by DumpStats.
func (a *StatsAdapter) MockStats(stats []adapter.StatEntry) {
	a.entries = stats
}

// MockStats sets mocked stat dir to be returned by PrepareDir.
func (a *StatsAdapter) MockDir(dir *adapter.StatDir) {
	a.dir = dir
}
