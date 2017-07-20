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

// Binary stats-client is an example VPP management application that exercises the
// govpp API for interface counters together with asynchronous connection to VPP.
package main

// Generates Go bindings for all VPP APIs located in the json directory.
//go:generate binapi-generator --input-dir=bin_api --output-dir=bin_api

import (
	"fmt"
	"os"
	"os/signal"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/core/bin_api/vpe"
	"git.fd.io/govpp.git/examples/bin_api/interfaces"
)

func main() {
	fmt.Println("Starting stats VPP client...")

	// async connect to VPP
	conn, statCh, err := govpp.AsyncConnect()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer conn.Disconnect()

	// create an API channel that will be used in the examples
	ch, err := conn.NewAPIChannel()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer ch.Close()

	// create channel for Interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	var simpleCountersSubs *api.NotifSubscription
	var combinedCountersSubs *api.NotifSubscription
	var notifChan chan api.Message

	// loop until Interrupt signal is received
loop:
	for {
		select {

		case connEvent := <-statCh:
			// VPP connection state change
			switch connEvent.State {
			case core.Connected:
				fmt.Println("VPP connected.")
				if simpleCountersSubs == nil {
					simpleCountersSubs, combinedCountersSubs, notifChan = subscribeNotifications(ch)
				}
				requestStatistics(ch)

			case core.Disconnected:
				fmt.Println("VPP disconnected.")
			}

		case msg := <-notifChan:
			switch notif := msg.(type) {
			case *interfaces.VnetInterfaceSimpleCounters:
				// simple counter notification received
				processSimpleCounters(notif)
			case *interfaces.VnetInterfaceCombinedCounters:
				// combined counter notification received
				processCombinedCounters(notif)
			default:
				fmt.Println("Ignoring unknown VPP notification")
			}

		case <-sigChan:
			// interrupt received
			fmt.Println("Interrupt received, exiting.")
			break loop
		}
	}

	ch.UnsubscribeNotification(simpleCountersSubs)
	ch.UnsubscribeNotification(combinedCountersSubs)
}

// subscribeNotifications subscribes for interface counters notifications.
func subscribeNotifications(ch *api.Channel) (*api.NotifSubscription, *api.NotifSubscription, chan api.Message) {

	notifChan := make(chan api.Message, 100)
	simpleCountersSubs, _ := ch.SubscribeNotification(notifChan, interfaces.NewVnetInterfaceSimpleCounters)
	combinedCountersSubs, _ := ch.SubscribeNotification(notifChan, interfaces.NewVnetInterfaceCombinedCounters)

	return simpleCountersSubs, combinedCountersSubs, notifChan
}

// requestStatistics requests interface counters notifications from VPP.
func requestStatistics(ch *api.Channel) {
	ch.SendRequest(&vpe.WantStats{
		Pid:           uint32(os.Getpid()),
		EnableDisable: 1,
	}).ReceiveReply(&vpe.WantStatsReply{})
}

// processSimpleCounters processes simple counters received from VPP.
func processSimpleCounters(counters *interfaces.VnetInterfaceSimpleCounters) {
	fmt.Printf("%+v\n", counters)

	counterNames := []string{"Drop", "Punt", "IPv4", "IPv6", "RxNoBuf", "RxMiss", "RxError", "TxError", "MPLS"}

	for i := uint32(0); i < counters.Count; i++ {
		fmt.Printf("Interface '%d': %s = %d\n",
			counters.FirstSwIfIndex+i, counterNames[counters.VnetCounterType], counters.Data[i])
	}
}

// processCombinedCounters processes combined counters received from VPP.
func processCombinedCounters(counters *interfaces.VnetInterfaceCombinedCounters) {
	fmt.Printf("%+v\n", counters)

	counterNames := []string{"Rx", "Tx"}

	for i := uint32(0); i < counters.Count; i++ {
		fmt.Printf("Interface '%d': %s packets = %d, %s bytes = %d\n",
			counters.FirstSwIfIndex+i, counterNames[counters.VnetCounterType], counters.Data[i].Packets,
			counterNames[counters.VnetCounterType], counters.Data[i].Bytes)
	}
}
