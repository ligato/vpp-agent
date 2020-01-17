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
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.ligato.io/vpp-agent/v3/client"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

const graphURL = "http://localhost:9191/scheduler/graph"

const (
	InfoMsgColor     = "\033[1;34m%s\033[0m\n"
	NoticeMsgColor   = "\033[1;35m%s\033[0m\n"
	NotifMsgColor    = "\033[1;33m%s\033[0m\n"
	ErrorMsgColor    = "\033[1;31m%s\033[0m\n"
	InputReqMsgColor = "\033[0;32m%s\033[0m\n"
	DebugMsgColor    = "\033[0;36m%s\033[0m\n"
)

// Run runs a scenario selected by the user.
func Run(kv kvs.KVScheduler, setLogging func(debugMode bool)) {
	time.Sleep(300 * time.Millisecond) // give agent logs time to get printed
	defer func() {
		fmt.Printf(InfoMsgColor, "The example scenario has finalized, the agent can be now terminated with CTRL-C.")
		fmt.Printf(InfoMsgColor, "But while the agent is still running, the REST API of KVScheduler can be explored.")
		fmt.Printf(InfoMsgColor, "Learn more from docs/kvscheduler/kvscheduler.md, section \"REST API\"")
	}()

	// let the user to select the scenario to run
	scenarioFnc := func() {}
	debugLog := true
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("")
	fmt.Printf(InputReqMsgColor, "Please select the example scenario to run: ")
	fmt.Printf(InputReqMsgColor, "1) BD-interface binding and FIB are pending after the first resync, but become ready after an update")
	fmt.Printf(InputReqMsgColor, "2) BD-interface binding and FIB are unconfigured and set as pending after an update")
	fmt.Printf(InputReqMsgColor, "3) Update transaction with invalid interface MAC address is reverted")
	fmt.Printf(InputReqMsgColor, "4) Invalid interface configuration fixed in the next Update transaction")
	fmt.Printf(InputReqMsgColor, "5) Change of the interface type performed via re-creation")
	fmt.Printf(InputReqMsgColor, "6) Failed interface creation fixed by subsequent Retry")
	fmt.Printf(InputReqMsgColor, "7) Run performance test")
inputLoop:
	for {
		fmt.Print("--> ")
		scanner.Scan()
		input := scanner.Text()
		switch input {
		case "1":
			scenarioFnc = PendingAfterResync
			break inputLoop
		case "2":
			scenarioFnc = PendingAfterChange
			break inputLoop
		case "3":
			scenarioFnc = RevertedChange
			break inputLoop
		case "4":
			scenarioFnc = InvalidConfigFixedWithUpdate
			break inputLoop
		case "5":
			scenarioFnc = InterfaceRecreation
			break inputLoop
		case "6":
			scenarioFnc = FailureFixedWithRetry
			break inputLoop
		case "7":
			debugLog = false
			scenarioFnc = PerfTest
			break inputLoop
		default:
			fmt.Printf(ErrorMsgColor, "Invalid option!")
			continue
		}
	}

	setLogging(debugLog)
	if debugLog {
		// watch and inform about value status updates
		ch := make(chan *kvscheduler.BaseValueStatus, 100)
		kv.WatchValueStatus(ch, nil)
		go watchValueStatus(ch)
	}

	// start the selected scenario
	scenarioFnc()
}

// watchValueStatus informs about value status updates.
func watchValueStatus(ch <-chan *kvscheduler.BaseValueStatus) {
	for {
		select {
		case status := <-ch:
			fmt.Printf(NotifMsgColor,
				fmt.Sprintf("Value status change: %v", status.String()))
		}
	}
}

// listKnownModels prints information about every registered model - in the case
// of this example, we have a model for interfaces, defined in ifplugin/model,
// and models for BDs and FIBs, defined under l2plugin/model.
func listKnownModels(c client.ConfigClient) {
	knownModels, err := c.KnownModels("config")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf(DebugMsgColor,
		fmt.Sprintf("Listing %d known models...", len(knownModels)))
	for _, model := range knownModels {
		fmt.Printf(DebugMsgColor, fmt.Sprintf(" - %v\n", model.String()))
	}
	time.Sleep(time.Second * 2)
}

func printMessage(lines ...string) {
	border := strings.Repeat("~", 80)
	fmt.Printf(InfoMsgColor, border)
	for _, line := range lines {
		fmt.Printf(InfoMsgColor, fmt.Sprintf("| %-76s |", line))
	}
	fmt.Printf(InfoMsgColor, border)
	time.Sleep(time.Second * 2)
}

func informAboutGraphURL(txnNum int, afterResync bool, sleep bool) {
	txnType := "data-update txn"
	if afterResync {
		txnType = "resync txn"
	}
	msg := fmt.Sprintf("Graph state after %s can be displayed at URL: %s?txn=%d\n",
		txnType, graphURL, txnNum)
	fmt.Printf(NoticeMsgColor, msg)
	if sleep {
		time.Sleep(time.Second * 5)
	}
}
