// Copyright (c) 2019 Bell Canada, Pantheon Technologies and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vpp_srv6_test

import (
	"testing"

	. "github.com/onsi/gomega"
	srv6 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/srv6"
)

// TestPolicyKey tests all cases for method PolicyKey
func TestPolicyKey(t *testing.T) {
	tests := []struct {
		name        string
		BSID        string
		expectedKey string
	}{
		{
			name:        "valid BD & iface names",
			BSID:        "a::",
			expectedKey: "config/vpp/srv6/v2/policy/a::",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RegisterTestingT(t)
			key := srv6.PolicyKey(test.BSID)
			Expect(key).To(Equal(test.expectedKey))
		})
	}
}
