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

package vpp2001

import (
	"testing"

	vpp_ifs "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2001/interfaces"
	ifs "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

func TestIsAdminStateUp(t *testing.T) {
	tests := []struct {
		input vpp_ifs.IfStatusFlags
		want  bool
	}{
		{input: 0, want: false},
		{input: 1, want: true},
		{input: 2, want: false},
		{input: 3, want: true},
		{input: 4, want: false},
		{input: 5, want: true},
		{input: 6, want: false},
	}

	for _, tc := range tests {
		got := isAdminStateUp(tc.input)
		if tc.want != got {
			t.Fatalf("for input: %d, want: %v, got: %v", tc.input, tc.want, got)
		}
	}
}

func TestIsLinkStateUp(t *testing.T) {
	tests := []struct {
		input vpp_ifs.IfStatusFlags
		want  bool
	}{
		{input: 0, want: false},
		{input: 1, want: false},
		{input: 2, want: true},
		{input: 3, want: true},
		{input: 4, want: false},
		{input: 5, want: false},
		{input: 6, want: true},
	}

	for _, tc := range tests {
		got := isLinkStateUp(tc.input)
		if tc.want != got {
			t.Fatalf("for input: %d, want: %v, got: %v", tc.input, tc.want, got)
		}
	}
}

func TestAdminStateToInterfaceStatus(t *testing.T) {
	tests := []struct {
		input vpp_ifs.IfStatusFlags
		want  ifs.InterfaceState_Status
	}{
		{input: 0, want: ifs.InterfaceState_DOWN},
		{input: 1, want: ifs.InterfaceState_UP},
		{input: 2, want: ifs.InterfaceState_DOWN},
		{input: 3, want: ifs.InterfaceState_UP},
		{input: 4, want: ifs.InterfaceState_DOWN},
	}

	for _, tc := range tests {
		got := adminStateToInterfaceStatus(tc.input)
		if tc.want != got {
			t.Fatalf("for input: %d, want: %v, got: %v", tc.input, tc.want, got)
		}
	}
}

func TestLinkStateToInterfaceStatus(t *testing.T) {
	tests := []struct {
		input vpp_ifs.IfStatusFlags
		want  ifs.InterfaceState_Status
	}{
		{input: 0, want: ifs.InterfaceState_DOWN},
		{input: 1, want: ifs.InterfaceState_DOWN},
		{input: 2, want: ifs.InterfaceState_UP},
		{input: 3, want: ifs.InterfaceState_UP},
		{input: 4, want: ifs.InterfaceState_DOWN},
	}

	for _, tc := range tests {
		got := linkStateToInterfaceStatus(tc.input)
		if tc.want != got {
			t.Fatalf("for input: %d, want: %v, got: %v", tc.input, tc.want, got)
		}
	}
}
