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

// PendingAfterChange presents a scenario, in which originally configured objects
// must be deleted and set as PENDING, because some another object, which they
// depend on, was set to be un-configured by an update transaction.
func PendingAfterChange() {
	c := client.LocalClient
	listKnownModels(c)

	printMessage(
		"Resync config",
		"  - 2 TAP interfaces (tap1, tap2) and 1 loopback are configured",
		"    and added into a bridge-domain",
		"  - FIB entry is defined to forward traffic for a certain",
		"    MAC address via tap2",
	)
	err := c.ResyncConfig(
		tap1, tap2, loopback1, bd1, fib2,
	)
	if err != nil {
		log.Println(err)
	}
	informAboutGraphURL(0, true, true)

	printMessage(
		"Change config",
		"  - the loopback interface is put DOWN",
		"    (no further effect, remains in the bridge-domain)",
		"  - tap1 MAC address is changed",
		"     -> the changed is performed by Update, re-creation is not needed",
		"  - tap2 interface is removed",
		"     -> the associated FIB that depends on tap2 to be in the bridge",
		"        domain is un-configured *first*, and goes into the PENDING",
		"        state",
		"     -> the interface is correctly removed from BD *before* it is",
		"        un-configured",
		"     -> the binding between BD and the TAP becomes pending",
	)
	loopback1.Enabled = false
	tap1.PhysAddress = "11:22:33:44:55:66"

	req := c.ChangeRequest()
	req.Update(tap1, loopback1)
	req.Delete(tap2)
	if err := req.Send(context.Background()); err != nil {
		log.Println(err)
	}
	informAboutGraphURL(1, false, true)
}
