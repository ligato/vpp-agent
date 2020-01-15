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
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/pkg/models"
)

type TestMetrics struct {
	TestValue int
}

var n = 0

func GetTestMetrics() *TestMetrics {
	return &TestMetrics{
		TestValue: n,
	}
}

func TestRetrieve(t *testing.T) {
	RegisterTestingT(t)

	models.DefaultRegistry = models.NewRegistry()

	_, err := models.DefaultRegistry.Register(&TestMetrics{}, models.Spec{
		Module: "testmodule",
		Type:   "testtype",
		Class:  "metrics",
	})
	Expect(err).ToNot(HaveOccurred())

	Register(&TestMetrics{}, func() interface{} {
		return GetTestMetrics()
	})

	n = 1

	var metricData TestMetrics
	err = Retrieve(&metricData)
	Expect(err).ToNot(HaveOccurred())
	Expect(metricData.TestValue).To(Equal(1))
}
