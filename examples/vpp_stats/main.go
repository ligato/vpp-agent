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

// The VPP stats example shows how to use telemetry API to access
// VPP stats via the GoVPP stats socket API and the telemetry vpp calls.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/agent"
	"go.ligato.io/cn-infra/v2/config"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/telemetry"
	"go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls"
)

const PluginName = "stats-example"

func main() {
	ep := &StatsExamplePlugin{
		Log:       logging.DefaultLogger,
		Telemetry: &telemetry.DefaultPlugin,
	}
	stopExample := make(chan struct{})

	a := agent.NewAgent(
		agent.AllPlugins(ep),
		agent.QuitOnClose(stopExample),
	)
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}

	go closeExample("Stats example finished", stopExample)
}

// StatsExamplePlugin displays VPP stats using telemetry plugin
type StatsExamplePlugin struct {
	handler vppcalls.TelemetryVppAPI

	config.PluginConfig
	Log       logging.Logger
	Telemetry *telemetry.Plugin
}

func (p *StatsExamplePlugin) Init() error {
	var err error
	p.handler, err = vppcalls.NewHandler(&govppmux.DefaultPlugin)
	if err != nil {
		panic(err)
	}

	go p.processStats()
	return nil
}

func (p *StatsExamplePlugin) Close() error {
	p.Log.Info("Stats example closed")
	return nil
}

func (p *StatsExamplePlugin) String() string {
	return PluginName
}

func closeExample(message string, stopExample chan struct{}) {
	time.Sleep(10 * time.Second)
	logrus.DefaultLogger().Info(message)
	close(stopExample)
}

func (p *StatsExamplePlugin) processStats() {
	// give the Agent some time to initialize
	// so the output is not mixed
	time.Sleep(1 * time.Second)
	p.Log.Infoln("Processing stats")

	var errors []error

	// collect stats
	ifStats, err := p.handler.GetInterfaceStats(context.Background())
	if err != nil {
		errors = append(errors, fmt.Errorf("eroror retireving interface stats: %v", err))
	}
	nodeCounters, err := p.handler.GetNodeCounters(context.Background())
	if err != nil {
		errors = append(errors, fmt.Errorf("eroror retireving node counters: %v", err))
	}

	systemStats, err := p.handler.GetSystemStats(context.Background())
	if err != nil {
		errors = append(errors, fmt.Errorf("eroror retireving system stats: %v", err))
	}

	runtimeInfo, err := p.handler.GetRuntimeInfo(context.Background())
	if err != nil {
		errors = append(errors, fmt.Errorf("eroror retireving runtime info: %v", err))
	}

	bufferInfo, err := p.handler.GetBuffersInfo(context.Background())
	if err != nil {
		errors = append(errors, fmt.Errorf("eroror retireving buffers info: %v", err))
	}

	threadsInfo, err := p.handler.GetThreads(context.Background())
	if err != nil {
		errors = append(errors, fmt.Errorf("eroror retireving threads: %v", err))
	}

	memoryInfo, err := p.handler.GetMemory(context.Background())
	fmt.Printf("mem %v, err %v", memoryInfo, err)
	if err != nil {
		errors = append(errors, fmt.Errorf("eroror retireving memory info: %v", err))
	}

	// print all errors and return if there is any
	if len(errors) != 0 {
		for _, err := range errors {
			p.Log.Error(err)
		}
		return
	}

	// print collected stats
	printIfStats(ifStats)
	printNodeCounters(nodeCounters)
	printSystemStats(systemStats)
	printRuntimeInfo(runtimeInfo)
	printBufferInfo(bufferInfo)
	printThreadsInfo(threadsInfo)
	printMemoryInfo(memoryInfo)
}

func printIfStats(ifStats *api.InterfaceStats) {
	for _, ifStat := range ifStats.Interfaces {
		fmt.Printf(`
Interface name: %s (sw_if_idx %d)
	Received: %d (rx errors %d)
	Transmitted: %d (tx errors %d)
	Drops: %d
`, ifStat.InterfaceName, ifStat.InterfaceIndex, ifStat.Rx, ifStat.RxErrors,
			ifStat.Tx, ifStat.TxErrors, ifStat.Drops)
	}
}

func printNodeCounters(nodeCountersInfo *vppcalls.NodeCounterInfo) {
	maxLen := 5
	for i, nodeCounters := range nodeCountersInfo.GetCounters() {
		if i >= maxLen {
			// do not print everything, it is not necessary
			break
		}
		fmt.Printf(`
Node name: %s 
Node: %s 

`, nodeCounters.Name, nodeCounters.Node)
	}
	if len(nodeCountersInfo.GetCounters()) >= maxLen {
		fmt.Printf("... and another %d nodes\n", len(nodeCountersInfo.GetCounters())-maxLen)
	}
}

func printSystemStats(systemStats *api.SystemStats) {
	fmt.Printf(`
Last update: %d
Last stats clear: %d
Input rate: %d
Num. Worker Threads: %d
Vector rate: %d (per worker: %+v)
Heartbeat: %d
`, systemStats.LastUpdate, systemStats.LastStatsClear, systemStats.InputRate, systemStats.NumWorkerThreads,
		systemStats.VectorRate, systemStats.VectorRatePerWorker, systemStats.Heartbeat)
}

func printRuntimeInfo(runtimeInfo *vppcalls.RuntimeInfo) {
	for _, thread := range runtimeInfo.GetThreads() {
		fmt.Printf("\nThread: %s (ID %d)", thread.Name, thread.ID)
	}
}

func printBufferInfo(bufferInfo *vppcalls.BuffersInfo) {
	for _, buffer := range bufferInfo.GetItems() {
		fmt.Printf(`

Buffer name: %s (index %d)
	Alloc: %d (num %d)
	Free: %d (num %d)
	Size: %d
	Thread ID: %d
`, buffer.Name, buffer.Index, buffer.Alloc, buffer.NumAlloc, buffer.Free, buffer.NumFree, buffer.Size, buffer.ThreadID)
	}
}

func printThreadsInfo(threadsInfo *vppcalls.ThreadsInfo) {
	for _, thread := range threadsInfo.GetItems() {
		fmt.Printf(`
Thread name: %s (ID %d)
	Type: %s
	PID: %d 
	Core: %d (CPU ID %d, CPU socket %d)
`, thread.Name, thread.ID, thread.Type, thread.PID, thread.Core, thread.CPUID, thread.CPUSocket)
	}
}

func printMemoryInfo(memoryInfo *vppcalls.MemoryInfo) {
	for _, thread := range memoryInfo.GetThreads() {
		fmt.Printf(`
Thread %d %s
  size %d, %d pages, page size %d
  total: %d, used: %d, free: %d, trimmable: %d
    free chunks %d free fastbin blks %d
    max total allocated %d
`, thread.ID, thread.Name, thread.Size, thread.Pages, thread.PageSize, thread.Total, thread.Used,
			thread.Free, thread.Trimmable, thread.FreeChunks, thread.FreeFastbinBlks, thread.MaxTotalAlloc)
	}
}
