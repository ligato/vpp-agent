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
	"log"

	"go.ligato.io/vpp-agent/v3/client"
	mocksb "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/mockcalls"
)

// FailureFixedWithRetry presents a scenario, in which interface fails to get
// created at a first attempt, but an automatic retry, triggered in the background
// after a delay, will succeed to create the interface and all the objects that
// were waiting for it.
func FailureFixedWithRetry() {
	c := client.LocalClient
	listKnownModels(c)

	mocksb.SimulateFailedTapCreation = true
	printMessage(
		"Resync config",
		"  - TAP interfaces tap1 to be configured inside a bridge-domain",
		"  - FIB entry defined to forward traffic for a certain",
		"    MAC address via tap1",
		"  - the TAP interface fails to get created (simulated in the mock SB)",
		"     -> the binding between the interface and the BD will remain pending",
		"     -> FIB will remain pending",
	)
	err := c.ResyncConfig(
		tap1, bd2, fib3,
	)
	if err != nil {
		log.Println(err)
	}
	informAboutGraphURL(0, true, false)

	printMessage(
		"Automatic Retry (triggered in the background)",
		"  - the repeated attempt to add tap1 will succeed",
		"     -> once configured, the interface is also added into the BD",
		"        and the pending FIB entry is created, both within the same",
		"        transaction",
	)
	// retry is scheduler to run 1 second after the failed transaction and
	// since printMessage() waits 2 seconds, the interface should be fixed by now
	informAboutGraphURL(1, false, true)
}
