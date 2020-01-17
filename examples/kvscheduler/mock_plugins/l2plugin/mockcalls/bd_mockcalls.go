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

	l2 "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model"
)

// CreateBridgeDomain creates new bridge domain in the mock SB.
func (h *MockBDHandler) CreateBridgeDomain(bdName string) (sbBDHandle uint32, err error) {
	sbBDHandle = h.nextBDHandle
	h.nextBDHandle++
	h.mockBDs[sbBDHandle] = &l2.BridgeDomain{
		Name: bdName,
	}
	h.log.Debugf("Created bridge domain: %s", bdName)
	return sbBDHandle, nil
}

// DeleteBridgeDomain removes existing bridge domain from the mock SB.
func (h *MockBDHandler) DeleteBridgeDomain(sbBDHandle uint32) error {
	bd, err := h.getBridgeDomain(sbBDHandle)
	if err != nil {
		return err
	}
	delete(h.mockBDs, sbBDHandle)
	h.log.Debugf("Deleted bridge domain: %s", bd.Name)
	return nil
}

// AddInterfaceToBridgeDomain puts interface into bridge domain.
func (h *MockBDHandler) AddInterfaceToBridgeDomain(sbBDHandle uint32, ifaceName string, isBVI bool) error {
	bd, err := h.getBridgeDomain(sbBDHandle)
	if err != nil {
		return err
	}
	for _, iface := range bd.Interfaces {
		if iface.Name == ifaceName {
			return fmt.Errorf("interface '%s' is already in the bridge domain '%s'",
				ifaceName, bd.Name)
		}
	}
	bd.Interfaces = append(bd.Interfaces, &l2.BridgeDomain_Interface{
		Name:                    ifaceName,
		BridgedVirtualInterface: isBVI,
	})
	h.log.Debugf("Added interface '%s' into the bridge domain '%s'",
		ifaceName, bd.Name)
	return nil
}

// DeleteInterfaceFromBridgeDomain removes interface from bridge domain.
func (h *MockBDHandler) DeleteInterfaceFromBridgeDomain(sbBDHandle uint32, ifaceName string) error {
	bd, err := h.getBridgeDomain(sbBDHandle)
	if err != nil {
		return err
	}
	var idx int
	for ; idx < len(bd.Interfaces); idx++ {
		if bd.Interfaces[idx].Name == ifaceName {
			break
		}
	}
	if idx == len(bd.Interfaces) {
		return fmt.Errorf("interface '%s' is not in the bridge domain '%s'",
			ifaceName, bd.Name)
	}
	bd.Interfaces = append(bd.Interfaces[:idx], bd.Interfaces[idx+1:]...)
	h.log.Debugf("Removed interface '%s' from the bridge domain '%s'",
		ifaceName, bd.Name)
	return nil
}

// DumpBridgeDomains dumps bridge domains "configured" in the mock SB.
func (h *MockBDHandler) DumpBridgeDomains() (mockBDs, error) {
	h.log.Debugf("Dumped mock bridge domains: %+v", h.mockBDs)
	return h.mockBDs, nil
}

// getInterface returns configuration of bridge domain represented in the mock SB
// with the given integer handle.
func (h *MockBDHandler) getBridgeDomain(sbBDHandle uint32) (*l2.BridgeDomain, error) {
	bd, exists := h.mockBDs[sbBDHandle]
	if !exists {
		return nil, fmt.Errorf("cannot find bridge domain with index: %d", sbBDHandle)
	}
	return bd, nil
}
