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
	"testing"

	"go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls"

	_ "go.ligato.io/vpp-agent/v3/plugins/telemetry"
)

func TestTelemetryNodeCounters(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	if test.versionInfo.Release() <= "19.04" {
		t.Skipf("SKIP for VPP %s<=19.04", test.versionInfo.Release())
	}

	h := vppcalls.CompatibleTelemetryHandler(test.vppClient)

	nodeCounters, err := h.GetNodeCounters(test.Context)
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

func TestTelemetryInterfaceStats(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	h := vppcalls.CompatibleTelemetryHandler(test.vppClient)

	ifStats, err := h.GetInterfaceStats(test.Context)
	if err != nil {
		t.Fatalf("getting interface stats failed: %v", err)
	} else {
		t.Logf("retrieved interface stats: %+v", ifStats)
	}
	if ifStats == nil || ifStats.Interfaces == nil {
		t.Fatal("expected interface stats, got nil")
	}
	if len(ifStats.Interfaces) == 0 {
		t.Fatalf("expected memory stats length > 0, got %v", len(ifStats.Interfaces))
	}
}
