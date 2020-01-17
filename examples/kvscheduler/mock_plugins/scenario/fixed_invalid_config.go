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
	"context"
	"log"

	"go.ligato.io/vpp-agent/v3/client"
)

// InvalidConfigFixedWithUpdate presents a scenario, in which an interface is
// first defined with an invalid MAC address and therefore remains in the INVALID
// state. All the objects that depend on it must remain PENDING.
// The next update, however, will supply an updated interface configuration
// where the MAC address is fixed. The interface and all the associated pending
// objects are then successfully created.
func InvalidConfigFixedWithUpdate() {
	c := client.LocalClient
	listKnownModels(c)

	printMessage(
		"Resync config",
		"  - requested config contains:",
		"     -> 2 TAP interfaces (tap1, tap2) and one loopback inside",
		"        a bridge-domain",
		"     -> FIB entry forwarding traffic for a certain MAC address",
		"        via tap2",
		"  - the issue is that tap2 has invalid MAC address defined, which",
		"    will cause it to remain in the INVALID state after the transaction",
		"  - with tap2 not configured, the binding between the BD and tap2",
		"    will remain pending and so will the FIB entry",
	)
	err := c.ResyncConfig(
		tap1, tap2Invalid, loopback1, bd1, fib2,
	)
	if err != nil {
		log.Println(err)
	}
	informAboutGraphURL(0, true, true)

	printMessage(
		"Change config",
		"  - update tap2 to contain valid MAC address",
		"  - tap2 will now get successfully configured and added into",
		"    the bridge domain together with the associated FIB entry which",
		"    was pending",
	)

	req := c.ChangeRequest()
	req.Update(tap2)
	if err := req.Send(context.Background()); err != nil {
		log.Println(err)
	}
	informAboutGraphURL(1, false, true)
}
