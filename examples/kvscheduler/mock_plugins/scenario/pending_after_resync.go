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

// PendingAfterResync presents a scenario, in which some objects are pending
// after the first resync since their dependencies are not satisfied.
// Then, with the next update transaction, the missing items are set to be
// configured together with all the associated objects that were waiting for
// them.
func PendingAfterResync() {
	c := client.LocalClient
	listKnownModels(c)

	printMessage(
		"Resync config",
		"  - TAP interface tap1 and single loopback are configured",
		"  - BD is configured to contain tap2 and the loopback, but not tap1",
		"     -> binding between tap2 and BD will be PENDING",
		"  - FIB entry fib1 is defined to forward traffic for a certain",
		"    MAC address via tap1",
		"     -> tap1 will get configured, but not inside the bridge",
		"        domain, therefore the FIB will remain pending",
		"  - FIB entry fib2 is defined to forward traffic for a certain",
		"    MAC address via tap2",
		"     -> tap2 is not configured, therefore the FIB will remain",
		"        pending",
	)
	err := c.ResyncConfig(
		tap1, loopback1, bd1WithoutTap1, fib1, fib2,
	)
	if err != nil {
		log.Println(err)
	}
	informAboutGraphURL(0, true, true)

	printMessage(
		"Change config",
		"  - bridge domain is edited to also contain tap1",
		"     -> first, tap1 will be put into the BD",
		"     -> as a consequence, fib1 becomes ready and gets configured",
		"  - tap2 interface is requested to be configured",
		"     -> the interface is configured first",
		"     -> next, the interface is put into the bridge domain",
		"     -> finally, the pending fib2 becomes ready and gets configured",
	)

	req := c.ChangeRequest()
	req.Update(bd1, tap2)
	if err := req.Send(context.Background()); err != nil {
		log.Println(err)
	}
	informAboutGraphURL(1, false, true)
}
