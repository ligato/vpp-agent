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

package models_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	testmodel "go.ligato.io/vpp-agent/v3/pkg/models/testdata/proto"
)

func TestKey(t *testing.T) {
	tc := setupTest(t)
	defer tc.teardownTest()

	tests := []struct {
		name      string
		instance  *testmodel.Basic
		expectKey string
	}{
		{"named",
			&testmodel.Basic{Name: "basic0"},
			"config/module/v1/basic/basic0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tc.Expect(models.GetKey(test.instance)).To(Equal(test.expectKey))
		})
	}
}

type testContext struct {
	*GomegaWithT
}

func setupTest(t *testing.T) *testContext {
	g := NewGomegaWithT(t)

	ResetDefaultRegistry()

	basicModel := models.Register(&testmodel.Basic{}, models.Spec{
		Module:  "module",
		Version: "v1",
		Type:    "basic",
	})
	g.Expect(basicModel).ToNot(BeNil())

	return &testContext{GomegaWithT: g}
}
func (tc *testContext) teardownTest() {}
