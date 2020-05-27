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

package vppagent

import (
	"os"
	"os/signal"
	"reflect"

	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"
)

func waitForSignal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	logging.Infof("waiting for signal..")
	return c
}

func runAfterInit(x interface{}) {
	p, ok := x.(infra.PostInit)
	if !ok {
		return
	}
	if err := p.AfterInit(); err != nil {
		panic(err)
	}
}

func forEachField(v interface{}, cb func(field interface{})) {
	val := reflect.ValueOf(v)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.IsNil() {
			continue
		}
		cb(field.Interface())
	}
}
