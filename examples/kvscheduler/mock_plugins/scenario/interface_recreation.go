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
	interfaces "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/model"
)

// InterfaceRecreation presents a scenario, in which the interface type is changed,
// which cannot be applied incrementally via Update, but requires the interface
// to be fully re-created, together with all the objects that depend on it.
func InterfaceRecreation() {
	c := client.LocalClient
	listKnownModels(c)

	printMessage(
		"Resync config",
		"  - TAP interfaces tap1 configured inside a bridge-domain",
		"  - FIB entry defined to forward traffic for a certain",
		"    MAC address via tap1",
	)
	err := c.ResyncConfig(
		tap1, bd2, fib3,
	)
	if err != nil {
		log.Println(err)
	}
	informAboutGraphURL(0, true, true)

	printMessage(
		"Change config",
		"  - the TAP interface tap1 is requested to be turned into loopback",
		"     -> the change requires the interface to be fully re-created",
		"     -> before the obsolete tap1 is un-configured, the associated FIB",
		"        entry and the binding between the interface and BD must be deleted",
		"        first",
		"     -> once tap1 is re-created as loopback (the name becomes confusing),",
		"        it is then re-added back into the bridge domain and at last",
		"        the FIB entry is restored",
	)
	tap1.Type = interfaces.Interface_LOOPBACK

	req := c.ChangeRequest()
	req.Update(tap1)
	if err := req.Send(context.Background()); err != nil {
		log.Println(err)
	}
	informAboutGraphURL(1, false, true)
}
