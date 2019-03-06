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

package kvscheduler

import (
	"encoding/json"
	"expvar"
	"fmt"
	"sync"
	"time"
)

var (
	descriptorStatsMu sync.RWMutex
	descriptorStats   = map[string]*DescriptorStats{}
)

func GetDescriptorStats() map[string]*DescriptorStats {
	stats := map[string]*DescriptorStats{}
	descriptorStatsMu.RLock()
	for d, ds := range descriptorStats {
		dss := *ds
		stats[d] = &dss
	}
	descriptorStatsMu.RUnlock()
	return stats
}

type DescriptorStats struct {
	Methods map[string]*MethodStats
}

func (s *DescriptorStats) MarshalJSON() ([]byte, error) {
	d := map[string]string{}
	for m, ms := range s.Methods {
		d[m] = fmt.Sprintf(
			"calls: %d, duration: %s (avg: %s, max: %s)",
			ms.Calls, ms.TotalNs, ms.AvgNs, ms.MaxNs,
		)
	}
	return json.Marshal(d)
}

type MethodStats struct {
	Calls   uint64
	TotalNs time.Duration
	AvgNs   time.Duration
	MaxNs   time.Duration
}

func trackDescMethod(d, m string) func() {
	t := time.Now()
	descriptorStatsMu.RLock()
	desc := descriptorStats[d]
	method, ok := desc.Methods[m]
	descriptorStatsMu.RUnlock()
	if !ok {
		desc.Methods[m] = new(MethodStats)
		descriptorStatsMu.Lock()
		method = desc.Methods[m]
		descriptorStatsMu.Unlock()
	}
	return func() {
		took := time.Since(t)
		descriptorStatsMu.Lock()
		method.Calls++
		method.TotalNs += took
		method.AvgNs = method.TotalNs / time.Duration(method.Calls)
		if took > method.MaxNs {
			method.MaxNs = took
		}
		descriptorStatsMu.Unlock()
	}
}

func init() {
	expvar.Publish("kvdescriptors", expvar.Func(func() interface{} {
		return GetDescriptorStats()
	}))
}
