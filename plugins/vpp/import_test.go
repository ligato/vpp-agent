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
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
	"text/tabwriter"

	"go.fd.io/govpp/api"

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

	type item struct {
		message string
		crc     string
		pkgPath string
	}
	items := []item{}
	for _, path := range msgTypes {
		for msgType := range path {
			typ := msgType.Elem()
			msg := reflect.New(typ).Interface().(api.Message)
			t := item{msg.GetMessageName(), msg.GetCrcString(), typ.PkgPath()}
			items = append(items, t)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].pkgPath == items[j].pkgPath {
			return items[i].message < items[j].message
		}
		return items[i].pkgPath < items[j].pkgPath
	})

	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "MESSAGE\tCRC\tPKG PATH\t\n")
	for _, item := range items {
		// typ := msgType.Elem()
		// msg := reflect.New(typ).Interface().(api.Message)
		// id := fmt.Sprintf("%s_%s", msg.GetMessageName(), msg.GetCrcString())
		// log.Printf("- msg: %s - %s (%v)", typ.String(), typ.PkgPath(), id)
		// fmt.Fprintf(w, "%s\t%s\t%s\t\n", msg.GetMessageName(), msg.GetCrcString(), typ.PkgPath())
		fmt.Fprintf(w, "%s\t%s\t%s\t\n", item.message, item.crc, item.pkgPath)
	}
	if err := w.Flush(); err != nil {
		return
	}
	fmt.Fprintf(os.Stdout, "%s", b)
}
