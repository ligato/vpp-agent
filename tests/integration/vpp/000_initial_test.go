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

	"go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
)

func TestPing(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	vpp := vppcalls.CompatibleHandler(test.vppClient)

	if err := vpp.Ping(context.Background()); err != nil {
		t.Fatalf("control ping failed: %v", err)
	}
}

func TestGetVersion(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	vpp := vppcalls.CompatibleHandler(test.vppClient)

	versionInfo, err := vpp.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("getting version failed: %v", err)
	}

	t.Logf("version: %v", versionInfo.Version)
}

func TestGetPlugins(t *testing.T) {
	test := setupVPP(t)
	defer test.teardownVPP()

	vpp := vppcalls.CompatibleHandler(test.vppClient)

	plugins, err := vpp.GetPlugins(context.Background())
	if err != nil {
		t.Fatalf("getting pluggins failed: %v", err)
	}

	t.Logf("%d plugins: %v", len(plugins), plugins)
}
