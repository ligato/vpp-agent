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

package app

import (
	"log"
	"strings"
	"testing"

	"go.ligato.io/vpp-agent/v2/plugins/vpp"
)

func TestHandlers(t *testing.T) {
	handlers := vpp.GetHandlers()

	log.Printf("listing %d handlers:", len(handlers))

	for h, handler := range handlers {
		versions := strings.Join(handler.Versions(), ", ")
		log.Printf(" - %s (%v)", h, versions)
	}
}
