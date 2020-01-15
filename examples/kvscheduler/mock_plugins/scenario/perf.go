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
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"encoding/binary"

	"go.ligato.io/vpp-agent/v3/client"

	interfaces "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/ifplugin/model"
	l2 "go.ligato.io/vpp-agent/v3/examples/kvscheduler/mock_plugins/l2plugin/model"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
)

const iterations = 1000
const groupSize = 100

func profile(c client.ConfigClient, runTxn func(iter int, withOrch bool), description ...string) {
	var (
		total time.Duration
		group time.Duration
	)

	for _, withOrch := range []bool{false, true} {
		emptyResync(c)
		total = 0

		// describe perf test
		withOrchStr := "With"
		if !withOrch {
			withOrchStr += "out"
		}
		msg := []string{fmt.Sprintf("%dx Change config (%s Orchestrator)", iterations, withOrchStr)}
		msg = append(msg, description...)
		printMessage(msg...)

		for i := 0; i < iterations; i++ {
			t := time.Now()
			runTxn(i, withOrch)
			elapsed := time.Since(t)
			total += elapsed
			group += elapsed

			if (i+1)%groupSize == 0 {
				fmt.Printf(DebugMsgColor, fmt.Sprintf("Txns %d-%d: %v",
					i-groupSize+1, i, group))
				group = 0
			}
		}
		fmt.Printf(DebugMsgColor, fmt.Sprintf("=> Total elapsed time: %v", total))
	}
}

func emptyResync(c client.ConfigClient) {
	printMessage("Empty Resync")
	err := c.ResyncConfig()
	if err != nil {
		log.Fatalln(err)
	}
}

func PerfTest() {
	c := client.LocalClient
	listKnownModels(c)

	// perf test #1
	addTap := func(iter int, withOrch bool) {
		tap := &interfaces.Interface{
			Name:    "tap-" + strconv.Itoa(iter),
			Type:    interfaces.Interface_TAP,
			Enabled: true,
		}
		if withOrch {
			req := c.ChangeRequest()
			req.Update(tap)
			if err := req.Send(context.Background()); err != nil {
				log.Fatalln(err)
			}
		} else {
			txn := kvscheduler.DefaultPlugin.StartNBTransaction()
			txn.SetValue(interfaces.InterfaceKey(tap.GetName()), tap)
			_, err := txn.Commit(context.Background())
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
	profile(c, addTap,
		"  - add new TAP interface")

	// perf test #3
	var bdIfaces []string
	addTapWithFIB := func(iter int, withOrch bool) {
		tap := &interfaces.Interface{
			Name:    "tap-" + strconv.Itoa(iter),
			Type:    interfaces.Interface_TAP,
			Enabled: true,
		}

		// bd
		if iter == 0 {
			bdIfaces = []string{}
		}
		bdIfaces = append(bdIfaces, tap.GetName())
		bd := &l2.BridgeDomain{
			Name:       "bd1",
			Interfaces: make([]*l2.BridgeDomain_Interface, 0, len(bdIfaces)),
		}
		for _, iface := range bdIfaces {
			bd.Interfaces = append(bd.Interfaces, &l2.BridgeDomain_Interface{
				Name: iface,
			})
		}

		// fib
		var hwAddr [6]byte
		binary.BigEndian.PutUint32(hwAddr[2:], uint32(iter))
		fib := &l2.FIBEntry{
			PhysAddress:       net.HardwareAddr(hwAddr[:]).String(),
			BridgeDomain:      bd.GetName(),
			Action:            l2.FIBEntry_FORWARD,
			OutgoingInterface: tap.GetName(),
		}

		if withOrch {
			req := c.ChangeRequest()
			req.Update(tap, bd, fib)
			if err := req.Send(context.Background()); err != nil {
				log.Fatalln(err)
			}
		} else {
			txn := kvscheduler.DefaultPlugin.StartNBTransaction()
			txn.SetValue(interfaces.InterfaceKey(tap.GetName()), tap)
			txn.SetValue(l2.BridgeDomainKey(bd.GetName()), bd)
			txn.SetValue(l2.FIBKey(fib.BridgeDomain, fib.PhysAddress), fib)
			_, err := txn.Commit(context.Background())
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
	profile(c, addTapWithFIB,
		"  - add new TAP interface",
		"  - insert the new TAP interface into the bridge domain",
		"  - add FIB routing traffic for a certain MAC via the new interface",
	)
}
