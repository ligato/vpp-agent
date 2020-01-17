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
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

// RevertedChange presents a scenario, in which an update transaction fails
// and gets reverted.
func RevertedChange() {
	c := client.LocalClient
	listKnownModels(c)

	printMessage(
		"Resync config",
		"  - interface tap1 is configured inside bridge-domain named bd2",
	)
	err := c.ResyncConfig(
		tap1, bd2,
	)
	if err != nil {
		log.Println(err)
	}
	informAboutGraphURL(0, true, true)

	printMessage(
		"Change config (revert-on-failure enabled)",
		"  - new FIB entry is configured to forward traffic for a certain",
		"    MAC address via tap1",
		"  - bridge domain bd2 is edited to also include tap2 interface",
		"     -> the binding is planned to be temporarily pending until",
		"        the interface gets configured in the next step",
		"  - attempt to configure tap2 fails",
		"     -> the MAC address is invalid as told by the Validate method()",
		"        of the descriptor for interfaces",
		"     -> transaction is reverted",
		"         -> BD is changed back to contain only tap1",
		"         -> the FIB is removed",
	)
	req := c.ChangeRequest()
	req.Update(bd2WithTap2, fib3, tap2Invalid)

	// IMPORTANT: by default, transactions are best effort - applying the maximum
	// possible subset of requested changes - i.e. allowing partial effect,
	// which is suitable for KVDB NB in combination with automatic Retry
	// of failed operations (see WithRetry option from KVScheduler API).
	// To get the standard transaction behaviour, where either everything or
	// nothing from the transaction is applied, WithRevert(context) must
	// be used to customize the context supplied to the Send() method of the
	// local client.
	ctx := context.Background()
	ctx = kvs.WithRevert(ctx)

	if err := req.Send(ctx); err != nil {
		log.Println(err)
	}
	informAboutGraphURL(1, false, true)
}
