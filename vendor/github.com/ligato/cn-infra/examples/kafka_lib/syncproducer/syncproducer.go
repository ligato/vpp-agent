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
	"strings"

	"github.com/namsral/flag"

	"github.com/ligato/cn-infra/examples/kafka_lib/utils"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/messaging/kafka/client"
)

var (
	brokerList  = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster. You can also set the KAFKA_PEERS environment variable")
	partitioner = flag.String("partitioner", "hash", "The partitioning scheme to use. Can be `hash`, `manual`, or `random`")
	partition   = flag.Int("partition", -1, "The partition to produce to.")
	debug       = flag.Bool("debug", false, "turn on debug logging")
	silent      = flag.Bool("silent", false, "Turn off printing the message's topic, partition, and offset to stdout")
)

func main() {
	log.StandardLogger().SetLevel(logging.DebugLevel)
	flag.Parse()

	if *brokerList == "" {
		printUsageErrorAndExit("no -brokers specified. Alternatively, set the KAFKA_PEERS environment variable")
	}

	// init config
	config := client.NewConfig(logroot.StandardLogger())
	config.SetDebug(*debug)
	config.SetPartition(int32(*partition))
	config.SetPartitioner(*partitioner)
	config.SetBrokers(strings.Split(*brokerList, ",")...)
	// init producer
	producer, err := client.NewSyncProducer(config, nil)
	if err != nil {
		fmt.Printf("NewSyncProducer errored: %v\n", err)
		os.Exit(1)
	}

	// get command
	for {
		command := utils.GetCommand()
		switch command.Cmd {
		case "quit":
			err := closeProducer(producer)
			if err != nil {
				fmt.Println("terminated abnormally")
				os.Exit(1)
			}
			fmt.Println("ended successfully")
			os.Exit(0)
		case "message":
			err := sendMessage(producer, command.Message)
			if err != nil {
				fmt.Printf("send message error: %v\n", err)
			}

		default:
			fmt.Println("invalid command")
		}
	}
}

// send message
func sendMessage(producer *client.SyncProducer, msg utils.Message) error {
	var (
		msgKey   []byte
		msgValue []byte
	)

	// init message
	if msg.Key != "" {
		msgKey = []byte(msg.Key)
	}
	msgValue = []byte(msg.Text)

	// send message
	_, err := producer.SendMsgByte(msg.Topic, msgKey, msgValue)
	if err != nil {
		log.StandardLogger().Errorf("SendMsg Error: %v", err)
		return err
	}
	fmt.Println("message sent")
	return nil
}

func closeProducer(producer *client.SyncProducer) error {
	// close producer
	fmt.Println("Closing producer ...")
	err := producer.Close()
	if err != nil {
		fmt.Printf("SyncProducer close errored: %v\n", err)
		log.StandardLogger().Errorf("SyncProducer close errored: %v", err)
		return err
	}
	return nil
}

func printUsageErrorAndExit(message string) {
	fmt.Fprintln(os.Stderr, "ERROR:", message)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}

func stdinAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
