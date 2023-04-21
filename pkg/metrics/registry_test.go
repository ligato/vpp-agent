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

package metrics_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/pkg/metrics"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	testmetrics "go.ligato.io/vpp-agent/v3/pkg/models/testdata/proto"
)

var n int32 = 0

func GetBasicMetrics() *testmetrics.Basic {
	return &testmetrics.Basic{
		ValueInt: n,
	}
}

func TestRetrieve(t *testing.T) {
	g := NewWithT(t)

	models.DefaultRegistry = models.NewRegistry()

	_, err := models.DefaultRegistry.Register(&testmetrics.Basic{}, models.Spec{
		Module: "testmodule",
		Type:   "testtype",
		Class:  "metrics",
	})
	g.Expect(err).ToNot(HaveOccurred())
	registered := models.RegisteredModels()
	g.Expect(registered).To(HaveLen(1))

	metrics.Register(&testmetrics.Basic{}, func() interface{} {
		return GetBasicMetrics()
	})

	n = 1

	var metricData testmetrics.Basic
	err = metrics.Retrieve(&metricData)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(metricData.ValueInt).To(Equal(int32(1)))
}
