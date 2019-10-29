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

package vpp

import (
	"context"
	"testing"

	_ "go.ligato.io/vpp-agent/v2/plugins/telemetry"
	"go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls"
)

func TestTelemetryNodeCounters(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	if ctx.versionInfo.Release() <= "19.04" {
		t.Skipf("SKIP for VPP %s", ctx.versionInfo.Release())
	}

	h := vppcalls.CompatibleTelemetryHandler(ctx.vppBinapi, ctx.vppStats)

	nodeCounters, err := h.GetNodeCounters(context.Background())
	if err != nil {
		t.Fatalf("getting node counters failed: %v", err)
	}
	t.Logf("retrieved %d node counters", len(nodeCounters.Counters))
	if nodeCounters.Counters == nil {
		t.Fatal("expected node counters, got nil")
	}
	if len(nodeCounters.Counters) == 0 {
		t.Fatalf("expected node counters length > 0, got %v", len(nodeCounters.Counters))
	}
}

func TestTelemetryInterfacStats(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := vppcalls.CompatibleTelemetryHandler(ctx.vppBinapi, ctx.vppStats)

	ifStats, err := h.GetInterfaceStats(context.Background())
	if err != nil {
		t.Fatalf("getting interface stats failed: %v", err)
	}
	t.Logf("retrieved interface stats: %+v", ifStats)
	if ifStats.Interfaces == nil {
		t.Fatal("expected interface stats, got nil")
	}
	if len(ifStats.Interfaces) == 0 {
		t.Fatalf("expected memory stats length > 0, got %v", len(ifStats.Interfaces))
	}
}
