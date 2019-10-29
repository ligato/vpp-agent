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

package vpp2001_324

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	govppapi "git.fd.io/govpp.git/api"

	vpevppcalls "go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	"go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls/vpp2001_324"
	"go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/memclnt"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/vpe"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, memclnt.AllMessages()...)
	msgs = append(msgs, vpe.AllMessages()...)

	vppcalls.Versions["20.01_324"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(ch govppapi.Channel) vppcalls.TelemetryVppAPI {
			return NewTelemetryVppHandler(ch)
		},
	}
}

type TelemetryHandler struct {
	vpe vpevppcalls.VpeVppAPI
}

func NewTelemetryVppHandler(ch govppapi.Channel) *TelemetryHandler {
	return &TelemetryHandler{
		vpe: vpp2001_324.NewVpeHandler(ch),
	}
}

func (h *TelemetryHandler) GetSystemStats(context.Context) (*govppapi.SystemStats, error) {
	return nil, nil
}

var (
	// Regular expression to parse output from `show memory`
	memoryRe = regexp.MustCompile(
		`Thread\s+(\d+)\s+(\w+).?\s+` +
			`virtual memory start 0x[0-9a-f]+, size ([\dkmg\.]+), ([\dkmg\.]+) pages, page size ([\dkmg\.]+)\s+` +
			`(?:page information not available.*\s+)*` +
			`(?:(?:\s+(?:numa [\d]+|not mapped|unknown): [\dkmg\.]+ pages, [\dkmg\.]+\s+)*\s+)*` +
			`\s+total: ([\dkmgKMG\.]+), used: ([\dkmgKMG\.]+), free: ([\dkmgKMG\.]+), trimmable: ([\dkmgKMG\.]+)`,
	)
)

// GetMemory retrieves `show memory` info.
func (h *TelemetryHandler) GetMemory(ctx context.Context) (*vppcalls.MemoryInfo, error) {
	input, err := h.vpe.RunCli("show memory main-heap")
	if err != nil {
		return nil, err
	}

	threadMatches := memoryRe.FindAllStringSubmatch(input, -1)

	if len(threadMatches) == 0 && input != "" {
		return nil, fmt.Errorf("invalid memory input: %q", input)
	}

	var threads []vppcalls.MemoryThread
	for _, matches := range threadMatches {
		fields := matches[1:]
		if len(fields) != 9 {
			return nil, fmt.Errorf("invalid memory data %v for thread: %q", fields, matches[0])
		}
		id, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return nil, err
		}
		thread := &vppcalls.MemoryThread{
			ID:        uint(id),
			Name:      fields[1],
			Size:      strToUint64(fields[2]),
			Pages:     strToUint64(fields[3]),
			PageSize:  strToUint64(fields[4]),
			Total:     strToUint64(fields[5]),
			Used:      strToUint64(fields[6]),
			Free:      strToUint64(fields[7]),
			Reclaimed: strToUint64(fields[8]),
		}
		threads = append(threads, *thread)
	}

	info := &vppcalls.MemoryInfo{
		Threads: threads,
	}

	return info, nil
}

func (h *TelemetryHandler) GetInterfaceStats(context.Context) (*govppapi.InterfaceStats, error) {
	return nil, nil
}

var (
	// Regular expression to parse output from `show node counters`
	nodeCountersRe = regexp.MustCompile(`^\s+(\d+)\s+([\w-\/]+)\s+(.+)$`)
)

// GetNodeCounters retrieves node counters info.
func (h *TelemetryHandler) GetNodeCounters(ctx context.Context) (*vppcalls.NodeCounterInfo, error) {
	data, err := h.vpe.RunCli("show node counters")
	if err != nil {
		return nil, err
	}

	var counters []vppcalls.NodeCounter

	for i, line := range strings.Split(string(data), "\n") {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Check first line
		if i == 0 {
			fields := strings.Fields(line)
			// Verify header
			if len(fields) != 3 || fields[0] != "Count" {
				return nil, fmt.Errorf("invalid header for `show node counters` received: %q", line)
			}
			continue
		}

		// Parse lines using regexp
		matches := nodeCountersRe.FindStringSubmatch(line)
		if len(matches)-1 != 3 {
			return nil, fmt.Errorf("parsing failed for `show node counters` line: %q", line)
		}
		fields := matches[1:]

		counters = append(counters, vppcalls.NodeCounter{
			Value: strToUint64(fields[0]),
			Node:  fields[1],
			Name:  fields[2],
		})
	}

	info := &vppcalls.NodeCounterInfo{
		Counters: counters,
	}

	return info, nil
}

var (
	// Regular expression to parse output from `show runtime`
	runtimeRe = regexp.MustCompile(`(?:-+\n)?(?:Thread (\d+) (\w+)(?: \(lcore \d+\))?\n)?` +
		`Time ([0-9\.e-]+), average vectors/node ([0-9\.e-]+), last (\d+) main loops ([0-9\.e-]+) per node ([0-9\.e-]+)\s+` +
		`vector rates in ([0-9\.e-]+), out ([0-9\.e-]+), drop ([0-9\.e-]+), punt ([0-9\.e-]+)\n` +
		`\s+Name\s+State\s+Calls\s+Vectors\s+Suspends\s+Clocks\s+Vectors/Call\s+(?:Perf Ticks\s+)?` +
		`((?:[\w-:\.]+\s+\w+(?:[ -]\w+)*\s+\d+\s+\d+\s+\d+\s+[0-9\.e-]+\s+[0-9\.e-]+\s+)+)`)
	runtimeItemsRe = regexp.MustCompile(`([\w-:\.]+)\s+(\w+(?:[ -]\w+)*)\s+(\d+)\s+(\d+)\s+(\d+)\s+([0-9\.e-]+)\s+([0-9\.e-]+)\s+`)
)

// GetRuntimeInfo retrieves how runtime info.
func (h *TelemetryHandler) GetRuntimeInfo(ctx context.Context) (*vppcalls.RuntimeInfo, error) {
	input, err := h.vpe.RunCli("show runtime")
	if err != nil {
		return nil, err
	}

	threadMatches := runtimeRe.FindAllStringSubmatch(input, -1)

	if len(threadMatches) == 0 && input != "" {
		return nil, fmt.Errorf("invalid runtime input: %q", input)
	}

	var threads []vppcalls.RuntimeThread
	for _, matches := range threadMatches {
		fields := matches[1:]
		if len(fields) != 12 {
			return nil, fmt.Errorf("invalid runtime data for thread (len=%v): %q", len(fields), matches[0])
		}
		thread := vppcalls.RuntimeThread{
			ID:                  uint(strToUint64(fields[0])),
			Name:                fields[1],
			Time:                strToFloat64(fields[2]),
			AvgVectorsPerNode:   strToFloat64(fields[3]),
			LastMainLoops:       strToUint64(fields[4]),
			VectorsPerMainLoop:  strToFloat64(fields[5]),
			VectorLengthPerNode: strToFloat64(fields[6]),
			VectorRatesIn:       strToFloat64(fields[7]),
			VectorRatesOut:      strToFloat64(fields[8]),
			VectorRatesDrop:     strToFloat64(fields[9]),
			VectorRatesPunt:     strToFloat64(fields[10]),
		}

		itemMatches := runtimeItemsRe.FindAllStringSubmatch(fields[11], -1)
		for _, matches := range itemMatches {
			fields := matches[1:]
			if len(fields) != 7 {
				return nil, fmt.Errorf("invalid runtime data for thread item: %q", matches[0])
			}
			thread.Items = append(thread.Items, vppcalls.RuntimeItem{
				Name:           fields[0],
				State:          fields[1],
				Calls:          strToUint64(fields[2]),
				Vectors:        strToUint64(fields[3]),
				Suspends:       strToUint64(fields[4]),
				Clocks:         strToFloat64(fields[5]),
				VectorsPerCall: strToFloat64(fields[6]),
			})
		}

		threads = append(threads, thread)
	}

	info := &vppcalls.RuntimeInfo{
		Threads: threads,
	}

	return info, nil
}

var (
	// Regular expression to parse output from `show buffers`
	buffersRe = regexp.MustCompile(
		`^(\w+(?:[ \-]\w+)*)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+([\dkmg\.]+)\s+([\dkmg\.]+)\s+([\dkmg\.]+)\s+([\dkmg\.]+)(?:\s+)?$`,
	)
)

// GetBuffersInfo retrieves buffers info from VPP.
func (h *TelemetryHandler) GetBuffersInfo(ctx context.Context) (*vppcalls.BuffersInfo, error) {
	data, err := h.vpe.RunCli("show buffers")
	if err != nil {
		return nil, err
	}

	var items []vppcalls.BuffersItem

	for i, line := range strings.Split(string(data), "\n") {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Check first line
		if i == 0 {
			fields := strings.Fields(line)
			// Verify header
			if len(fields) != 11 || fields[0] != "Pool" {
				return nil, fmt.Errorf("invalid header for `show buffers` received: %q", line)
			}
			continue
		}

		// Parse lines using regexp
		matches := buffersRe.FindStringSubmatch(line)
		if len(matches)-1 != 9 {
			return nil, fmt.Errorf("parsing failed (%d matches) for `show buffers` line: %q", len(matches), line)
		}
		fields := matches[1:]

		items = append(items, vppcalls.BuffersItem{
			//ThreadID: uint(strToUint64(fields[0])),
			Name:  fields[0],
			Index: uint(strToUint64(fields[1])),
			Size:  strToUint64(fields[3]),
			Alloc: strToUint64(fields[7]),
			Free:  strToUint64(fields[5]),
			//NumAlloc: strToUint64(fields[6]),
			//NumFree:  strToUint64(fields[7]),
		})
	}

	info := &vppcalls.BuffersInfo{
		Items: items,
	}

	return info, nil
}

func strToFloat64(s string) float64 {
	// Replace 'k' (thousands) with 'e3' to make it parsable with strconv
	s = strings.Replace(s, "k", "e3", 1)
	s = strings.Replace(s, "K", "e3", 1)
	s = strings.Replace(s, "m", "e6", 1)
	s = strings.Replace(s, "M", "e6", 1)
	s = strings.Replace(s, "g", "e9", 1)
	s = strings.Replace(s, "G", "e9", 1)

	num, err := strconv.ParseFloat(s, 10)
	if err != nil {
		return 0
	}
	return num
}

func strToUint64(s string) uint64 {
	return uint64(strToFloat64(s))
}
