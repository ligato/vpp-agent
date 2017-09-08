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

// Binary simple-client is an example VPP management application that exercises the
// govpp API on real-world use-cases.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/profile"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/core/bin_api/vpe"
)

const (
	defaultSyncRequestCount  = 1000
	defaultAsyncRequestCount = 1000000
)

func main() {
	// parse optional flags
	var sync, prof bool
	var cnt int
	flag.BoolVar(&sync, "sync", false, "run synchronous perf test")
	flag.IntVar(&cnt, "cnt", 0, "count of requests to be sent to VPP")
	flag.BoolVar(&prof, "prof", false, "generate profile data")
	flag.Parse()

	if cnt == 0 {
		// no specific count defined - use defaults
		if sync {
			cnt = defaultSyncRequestCount
		} else {
			cnt = defaultAsyncRequestCount
		}
	}

	if prof {
		defer profile.Start().Stop()
	}

	// log only errors
	core.SetLogger(&logrus.Logger{Level: logrus.ErrorLevel})

	// connect to VPP
	conn, err := govpp.Connect()
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
	defer conn.Disconnect()

	// create an API channel
	ch, err := conn.NewAPIChannelBuffered(cnt, cnt)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}
	defer ch.Close()

	// run the test & measure the time
	start := time.Now()

	if sync {
		// run synchronous test
		syncTest(ch, cnt)
	} else {
		// run asynchronous test
		asyncTest(ch, cnt)
	}

	elapsed := time.Since(start)
	fmt.Println("Test took:", elapsed)
	fmt.Printf("Requests per second: %.0f\n", float64(cnt)/elapsed.Seconds())
}

func syncTest(ch *api.Channel, cnt int) {
	fmt.Printf("Running synchronous perf test with %d requests...\n", cnt)

	for i := 0; i < cnt; i++ {
		req := &vpe.ControlPing{}
		reply := &vpe.ControlPingReply{}

		err := ch.SendRequest(req).ReceiveReply(reply)
		if err != nil {
			log.Println("Error in reply:", err)
			os.Exit(1)
		}
	}
}

func asyncTest(ch *api.Channel, cnt int) {
	fmt.Printf("Running asynchronous perf test with %d requests...\n", cnt)

	// start a new go routine that reads the replies
	var wg sync.WaitGroup
	wg.Add(1)
	go readAsyncReplies(ch, cnt, &wg)

	// send asynchronous requests
	sendAsyncRequests(ch, cnt)

	// wait until all replies are recieved
	wg.Wait()
}

func sendAsyncRequests(ch *api.Channel, cnt int) {
	for i := 0; i < cnt; i++ {
		ch.ReqChan <- &api.VppRequest{
			Message: &vpe.ControlPing{},
		}
	}
}

func readAsyncReplies(ch *api.Channel, expectedCnt int, wg *sync.WaitGroup) {
	cnt := 0

	for {
		// receive a reply
		reply := <-ch.ReplyChan
		if reply.Error != nil {
			log.Println("Error in reply:", reply.Error)
			os.Exit(1)
		}

		// decode the message
		msg := &vpe.ControlPingReply{}
		err := ch.MsgDecoder.DecodeMsg(reply.Data, msg)
		if reply.Error != nil {
			log.Println("Error by decoding:", err)
			os.Exit(1)
		}

		// count and return if done
		cnt++
		if cnt >= expectedCnt {
			wg.Done()
			return
		}
	}
}
