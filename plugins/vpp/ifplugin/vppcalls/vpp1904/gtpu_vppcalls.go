//  Copyright (c) 2019 EMnify
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

package vpp1904

import (
	"errors"

	ifs "github.com/ligato/vpp-agent/api/models/vpp/interfaces"
)

// AddGtpuTunnel adds new GTPU interface.
func (h *InterfaceVppHandler) AddGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink, multicastIf uint32) (uint32, error) {
	err := errors.New("Not implemented")
	return 0, err
}

// DelGtpuTunnel removes GTPU interface.
func (h *InterfaceVppHandler) DelGtpuTunnel(ifName string, gtpuLink *ifs.GtpuLink) error {
	err := errors.New("Not implemented")
	return err
}
