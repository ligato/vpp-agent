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

package metrics

import (
	"fmt"
	"reflect"
	"sync"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

var (
	mu                sync.RWMutex
	registeredMetrics = make(map[string]Retriever)
)

// Retriever defines function that returns metrics data
type Retriever func() interface{}

// Register registers given type with retriever to metrics.
func Register(metricType interface{}, retrieverFunc Retriever) {
	model, err := models.DefaultRegistry.GetModelFor(metricType)
	if err != nil {
		panic(fmt.Sprintf("type %T not registered as model", metricType))
	}
	if model.Spec().Class != "metrics" {
		panic(fmt.Sprintf("model %v not registered with class metrics", model.Name()))
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := registeredMetrics[model.Name()]; ok {
		panic(fmt.Sprintf("duplicate registration for metrics model %s", model.Name()))
	}
	registeredMetrics[model.Name()] = retrieverFunc
}

// Retrieve calls registered retriever for given metric and sets returned data to metric.
func Retrieve(metric interface{}) error {
	model, err := models.DefaultRegistry.GetModelFor(metric)
	if err != nil {
		return fmt.Errorf("type %T not registered as model", metric)
	}
	mu.RLock()
	retriever, ok := registeredMetrics[model.Name()]
	if !ok {
		mu.RUnlock()
		return fmt.Errorf("metric %v does not have registered retriever", model.Name())
	}
	mu.RUnlock()
	data := retriever()
	reflect.ValueOf(metric).Elem().Set(reflect.ValueOf(data).Elem())
	return nil
}
