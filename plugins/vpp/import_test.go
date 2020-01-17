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

package vpp_test

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"git.fd.io/govpp.git/api"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"

	// force import of all vpp plugins
	_ "go.ligato.io/vpp-agent/v3/cmd/vpp-agent/app"
)

func TestVppHandlers(t *testing.T) {
	handlers := vpp.GetHandlers()
	log.Printf("%d handlers:", len(handlers))

	for h, handler := range handlers {
		log.Printf("- handler: %-10s (%v)", h, handler.Versions())
	}
}

func TestBinapiMessage(t *testing.T) {
	msgTypes := api.GetRegisteredMessageTypes()
	log.Printf("%d binapi messages:", len(msgTypes))

	for msgType := range msgTypes {
		typ := msgType.Elem()
		msg := reflect.New(typ).Interface().(api.Message)
		id := fmt.Sprintf("%s_%s", msg.GetMessageName(), msg.GetCrcString())
		log.Printf("- msg: %s - %s (%v)", typ.String(), typ.PkgPath(), id)
	}
}
