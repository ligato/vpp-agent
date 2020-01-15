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

package scenario

import (
	interfaces "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/model"
	l2 "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model"
)

// Items used across all scenarios:

var (
	loopback1 = &interfaces.Interface{
		Name:    "loopback1",
		Type:    interfaces.Interface_TAP,
		Enabled: true,
	}

	tap1 = &interfaces.Interface{
		Name:    "tap1",
		Type:    interfaces.Interface_TAP,
		Enabled: true,
	}

	tap2 = &interfaces.Interface{
		Name:        "tap2",
		Type:        interfaces.Interface_TAP,
		Enabled:     true,
		PhysAddress: "11:22:33:44:55:66",
	}

	tap2Invalid = &interfaces.Interface{
		Name:        "tap2",
		Type:        interfaces.Interface_TAP,
		Enabled:     true,
		PhysAddress: "invalid-hw-address",
	}

	bd1 = &l2.BridgeDomain{
		Name: "bd1",
		Interfaces: []*l2.BridgeDomain_Interface{
			{
				Name: tap1.GetName(),
			},
			{
				Name: tap2.GetName(),
			},
			{
				Name:                    loopback1.GetName(),
				BridgedVirtualInterface: true,
			},
		},
	}

	bd1WithoutTap1 = &l2.BridgeDomain{ // bd1 edited
		Name: "bd1",
		Interfaces: []*l2.BridgeDomain_Interface{
			{
				Name: tap2.GetName(),
			},
			{
				Name:                    loopback1.GetName(),
				BridgedVirtualInterface: true,
			},
		},
	}

	bd2 = &l2.BridgeDomain{
		Name: "bd2",
		Interfaces: []*l2.BridgeDomain_Interface{
			{
				Name: tap1.GetName(),
			},
		},
	}

	bd2WithTap2 = &l2.BridgeDomain{
		Name: "bd2",
		Interfaces: []*l2.BridgeDomain_Interface{
			{
				Name: tap1.GetName(),
			},
			{
				Name: tap2.GetName(),
			},
		},
	}

	fib1 = &l2.FIBEntry{
		PhysAddress:       "cc:cc:cc:dd:dd:dd",
		BridgeDomain:      bd1.GetName(),
		Action:            l2.FIBEntry_FORWARD,
		OutgoingInterface: tap1.GetName(),
	}

	fib2 = &l2.FIBEntry{
		PhysAddress:       "aa:aa:aa:bb:bb:bb",
		BridgeDomain:      bd1.GetName(),
		Action:            l2.FIBEntry_FORWARD,
		OutgoingInterface: tap2.GetName(),
	}

	fib3 = &l2.FIBEntry{
		PhysAddress:       "ee:ee:ee:ff:ff:ff",
		BridgeDomain:      bd2.GetName(),
		Action:            l2.FIBEntry_FORWARD,
		OutgoingInterface: tap1.GetName(),
	}
)
