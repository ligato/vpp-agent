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

package main

import (
	"fmt"
	"os"
	"os/signal"

	"time"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"github.com/ligato/cn-infra/messaging/kafka/mux"
)

func main() {
	log.StandardLogger().SetLevel(logging.DebugLevel)
	mx, err := mux.InitMultiplexer("", "default", logroot.StandardLogger())
	if err != nil {
		os.Exit(1)
	}
	cn := mx.NewConnection("plugin")
	offset, err := cn.SendSyncString("test", mux.DefPartition, "key", "value")
	if err == nil {
		fmt.Println("Sync published ", offset)
	}

	succCh := make(chan *client.ProducerMessage)
	errCh := make(chan *client.ProducerError)
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)

	cn.SendAsyncString("test", mux.DefPartition, "key", "async!!", "meta", mux.ToBytesProducerChan(succCh), mux.ToBytesProducerErrChan(errCh))

	select {
	case success := <-succCh:
		fmt.Println("Successfully send async msg", success.Metadata)
	case err := <-errCh:
		fmt.Println("Error while sending async msg", err.Err, err.ProducerMessage.Metadata)
	}

	consumerChan := make(chan *client.ConsumerMessage)
	err = cn.ConsumeTopic(mux.ToBytesMsgChan(consumerChan), "test")
	mx.Start()
	if err == nil {
		fmt.Println("Consuming test partition")
	eventLoop:
		for {
			select {
			case msg := <-consumerChan:
				fmt.Println(string(msg.Key), string(msg.Value))
			case <-time.After(3 * time.Second):
				break eventLoop
			case <-signalChan:
				break eventLoop
			}
		}
	}

	mx.Close()
}
