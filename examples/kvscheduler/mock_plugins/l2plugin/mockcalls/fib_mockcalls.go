// Copyright (c) 2017 Cisco and/or its affiliates.
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

package mockcalls

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	l2 "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model"
)

// CreateL2FIB creates L2 FIB table entry in the mock SB.
func (h *MockFIBHandler) CreateL2FIB(fib *l2.FIBEntry) error {
	h.mockFIBs = append(h.mockFIBs, fib)
	h.log.Infof("Created L2 FIB entry: %v", fib)
	return nil
}

// DeleteL2FIB removes existing L2 FIB table entry from the mock SB.
func (h *MockFIBHandler) DeleteL2FIB(fib *l2.FIBEntry) error {
	var idx int
	for ; idx < len(h.mockFIBs); idx++ {
		if proto.Equal(h.mockFIBs[idx], fib) {
			break
		}
	}
	if idx == len(h.mockFIBs) {
		return fmt.Errorf("no such L2 FIB entry: %v", fib)
	}
	h.mockFIBs = append(h.mockFIBs[:idx], h.mockFIBs[idx+1:]...)
	h.log.Infof("Deleted L2 FIB entry: %v", fib)
	return nil
}

// DumpL2FIBs dumps L2 FIB table entries "configured" in the mock SB.
func (h *MockFIBHandler) DumpL2FIBs() ([]*l2.FIBEntry, error) {
	h.log.Infof("Dumped L2 FIB entries: %+v", h.mockFIBs)
	return h.mockFIBs, nil
}
