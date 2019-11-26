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
	"log"
	"strings"
	"testing"

	"go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	"go.ligato.io/vpp-agent/v2/plugins/vpp"
)

func TestPing(t *testing.T) {
	ctx := setupVPP(t)
	defer ctx.teardownVPP()

	h := vppcalls.CompatibleHandler(ctx.vppClient)

	if err := h.Ping(context.TODO()); err != nil {
		t.Fatalf("control ping failed: %v", err)
	}

	handlers := vpp.GetHandlers()
	log.Printf("listing %d handlers:", len(handlers))
	for h, handler := range handlers {
		log.Printf(" - %s (%v)",
			h, strings.Join(handler.Versions(), ", "))
	}
}
