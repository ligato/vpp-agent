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

package vppcalls

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSplitErrorName(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expNode, expReason string
	}{
		{"basic", "ipsec-input-ip4/IPSEC pkts received", "ipsec-input-ip4", "IPSEC pkts received"},
		{"ifname", "memif1/1001-output/interface is down", "memif1/1001-output", "interface is down"},
		{"reslash", "tcp6-input/inconsistent ip/tcp lengths", "tcp6-input", "inconsistent ip/tcp lengths"},
		{"toomany", "memif1/1001-output/Unrecognized / unknown chunk or chunk-state mismatch", "memif1/1001-output", "Unrecognized / unknown chunk or chunk-state mismatch"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RegisterTestingT(t)

			node, reason := SplitErrorName(test.input)
			Expect(node).To(Equal(test.expNode))
			Expect(reason).To(Equal(test.expReason))
		})
	}
}
