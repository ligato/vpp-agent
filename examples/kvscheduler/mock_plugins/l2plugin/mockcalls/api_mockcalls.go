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
	"github.com/ligato/cn-infra/logging"

	l2 "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model"
	"go.ligato.io/vpp-agent/v3/pkg/idxvpp"
)

// map of bridge domains "configured" in the mock SB.
type mockBDs map[uint32]*l2.BridgeDomain

// MockBDAPI provides methods for managing bridge domains in the mock SB.
type MockBDAPI interface {
	MockBDWrite
	MockBDRead
}

// MockBDWrite provides write methods for bridge domains.
type MockBDWrite interface {
	// CreateBridgeDomain creates new bridge domain in the mock SB.
	CreateBridgeDomain(bdName string) (sbBDHandle uint32, err error)
	// DeleteBridgeDomain removes existing bridge domain from the mock SB.
	DeleteBridgeDomain(sbBDHandle uint32) error
	// AddInterfaceToBridgeDomain puts interface into bridge domain.
	AddInterfaceToBridgeDomain(sbBDHandle uint32, ifaceName string, isBVI bool) error
	// DeleteInterfaceFromBridgeDomain removes interface from bridge domain.
	DeleteInterfaceFromBridgeDomain(sbBDHandle uint32, ifaceName string) error
}

// MockBDRead provides read methods for bridge domains.
type MockBDRead interface {
	// DumpBridgeDomains dumps bridge domains "configured" in the mock SB.
	DumpBridgeDomains() (mockBDs, error)
}

// MockFIBAPI provides methods for managing FIBs.
type MockFIBAPI interface {
	MockFIBWrite
	MockFIBRead
}

// MockFIBWrite provides write methods for FIBs.
type MockFIBWrite interface {
	// CreateL2FIB creates L2 FIB table entry in the mock SB.
	CreateL2FIB(fib *l2.FIBEntry) error
	// DeleteL2FIB removes existing L2 FIB table entry from the mock SB.
	DeleteL2FIB(fib *l2.FIBEntry) error
}

// MockFIBRead provides read methods for FIBs.
type MockFIBRead interface {
	// DumpL2FIBs dumps L2 FIB table entries "configured" in the mock SB.
	DumpL2FIBs() ([]*l2.FIBEntry, error)
}

// MockBDHandler is accessor for bridge domain-related calls into mock SB.
type MockBDHandler struct {
	ifaceIndex idxvpp.NameToIndex // exposed by the ifplugin
	log        logging.Logger

	// mock SB
	nextBDHandle uint32
	mockBDs      mockBDs
}

// MockFIBHandler is accessor for FIB-related calls into mock SB.
type MockFIBHandler struct {
	ifaceIndex idxvpp.NameToIndex // exposed by the ifplugin
	bdIndex    idxvpp.NameToIndex // retrieved from kvscheduler by l2plugin itself
	log        logging.Logger

	// mock SB
	mockFIBs []*l2.FIBEntry
}

// NewMockBDHandler creates new instance of bridge domain handler for mock SB.
func NewMockBDHandler(ifaceIndex idxvpp.NameToIndex, log logging.Logger) MockBDAPI {
	return &MockBDHandler{
		ifaceIndex: ifaceIndex,
		log:        log,
		mockBDs:    make(mockBDs),
	}
}

// NewMockFIBHandler creates new instance of FIB handler for mock SB.
func NewMockFIBHandler(ifaceIndex idxvpp.NameToIndex, bdIndexes idxvpp.NameToIndex,
	log logging.Logger) MockFIBAPI {
	return &MockFIBHandler{
		ifaceIndex: ifaceIndex,
		bdIndex:    bdIndexes,
		log:        log,
	}
}
