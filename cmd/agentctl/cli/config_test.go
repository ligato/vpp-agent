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

package cli

import (
	"reflect"
	"testing"
)

func TestAdjustSecurity(t *testing.T) {
	// "want" will be compared with result of calling "adjustSecurity" function  with "insecureTLS" set to true.
	tests := map[string]struct {
		cfg  *TLSConfig
		want *TLSConfig
	}{
		"nil cfg": {
			cfg:  nil,
			want: &TLSConfig{SkipVerify: true},
		},

		"empty cfg": {
			cfg:  &TLSConfig{},
			want: &TLSConfig{SkipVerify: true},
		},

		"disabled + dont skip verify": {
			cfg: &TLSConfig{
				Disabled: true, CertFile: "/cert.pem", KeyFile: "/key.pem", CAFile: "/ca.pem", SkipVerify: false,
			},
			want: &TLSConfig{
				Disabled: false, SkipVerify: true,
			},
		},

		"disabled + skip verify": {
			cfg: &TLSConfig{
				Disabled: true, CertFile: "/cert.pem", KeyFile: "/key.pem", CAFile: "/ca.pem", SkipVerify: true,
			},
			want: &TLSConfig{
				Disabled: false, SkipVerify: true,
			},
		},

		"not disabled + dont skip verify": {
			cfg: &TLSConfig{
				Disabled: false, CertFile: "/cert.pem", KeyFile: "/key.pem", CAFile: "/ca.pem", SkipVerify: false,
			},
			want: &TLSConfig{
				Disabled: false, CertFile: "/cert.pem", KeyFile: "/key.pem", CAFile: "/ca.pem", SkipVerify: true,
			},
		},

		"not disabled + skip verify": {
			cfg: &TLSConfig{
				Disabled: false, CertFile: "/cert.pem", KeyFile: "/key.pem", CAFile: "/ca.pem", SkipVerify: true,
			},
			want: &TLSConfig{
				Disabled: false, CertFile: "/cert.pem", KeyFile: "/key.pem", CAFile: "/ca.pem", SkipVerify: true,
			},
		},
	}

	for name, tc := range tests {
		// Do not expect any changes for case when "insecureTLS" param is false.
		got := adjustSecurity("dummy", false, tc.cfg)
		if !reflect.DeepEqual(tc.cfg, got) {
			t.Fatalf("%s (insecureTLS = false): expected: %v, got: %v", name, tc.cfg, got)
		}

		got = adjustSecurity("dummy", true, tc.cfg)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("%s (insecureTLS = true): expected: %v, got: %v", name, tc.want, got)
		}
	}
}
