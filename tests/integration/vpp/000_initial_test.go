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

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
)

func TestPing(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := vppcalls.CompatibleVpeHandler(ctx.Chan)

	if err := h.Ping(); err != nil {
		t.Fatalf("control ping failed: %v", err)
	}
}

func TestVersion(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := vppcalls.CompatibleVpeHandler(ctx.Chan)

	info, err := h.GetVersionInfo()
	if err != nil {
		t.Fatalf("getting version info failed: %v", err)
	}
	t.Logf("version info: %+v", info)
	if info.Version == "" {
		t.Error("invalid version info")
	}
}
