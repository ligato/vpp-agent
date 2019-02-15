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

	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
	. "github.com/onsi/gomega"
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

// TestParsePolicySegmentList tests all cases for method ParsePolicySegmentList
func TestParsePolicySegmentList(t *testing.T) {
	tests := []struct {
		name                          string
		key                           string
		expectedPolicyBSID            string
		expectedPolicySegmentListName string
		expectedIsPSLKey              bool
	}{
		{
			name: "valid policy segment list key",
			key:  "config/vpp/srv6/v2/policysegmentlist/slname/policy/a::",
			expectedPolicySegmentListName: "slname",
			expectedPolicyBSID:            "a::",
			expectedIsPSLKey:              true,
		},
		{
			name:             "invalid policy segment list key due to missing policy part",
			key:              "config/vpp/srv6/v2/policysegmentlist/slname",
			expectedIsPSLKey: false,
		},
		{
			name:             "invalid policy segment list key due to policy part misspeling",
			key:              "config/vpp/srv6/v2/policysegmentlist/slname/poXlicy/a::",
			expectedIsPSLKey: false,
		},
		{
			name:             "invalid policy segment list key due to policy segment list part misspeling",
			key:              "config/vpp/srv6/v2/policyXsegmentlist/slname/policy/a::",
			expectedIsPSLKey: false,
		},
		{
			name:             "invalid policy segment list key due to module part misspeling",
			key:              "config/vpp/srXv6/v2/policysegmentlist/slname/policy/a::",
			expectedIsPSLKey: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RegisterTestingT(t)
			bsid, slName, isKey := srv6.ParsePolicySegmentList(test.key)
			Expect(isKey).To(Equal(test.expectedIsPSLKey))
			if test.expectedPolicyBSID != "" {
				Expect(bsid).To(Equal(test.expectedPolicyBSID))
			}
			if test.expectedPolicySegmentListName != "" {
				Expect(slName).To(Equal(test.expectedPolicySegmentListName))
			}
		})
	}
}
