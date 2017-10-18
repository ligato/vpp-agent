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

	"github.com/namsral/flag"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/messaging/kafka/client"
)

var (
	// Flags used to read the input arguments.
	brokerList = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster")
	topicList  = flag.String("topics", "", "REQUIRED: the topics to consume")
	groupID    = flag.String("groupid", "", "REQUIRED: the group name")
	offset     = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest`")
	debug      = flag.Bool("debug", false, "turns on debug logging")
	commit     = flag.Bool("commit", false, "Commit offsets (default: true)")
)

func main() {
	log.StandardLogger().SetLevel(logging.DebugLevel)

	flag.Parse()

	if *brokerList == "" {
		printUsageErrorAndExit("You have to provide -brokers as a comma-separated list, or set the KAFKA_PEERS environment variable.")
	}

	if *topicList == "" {
		printUsageErrorAndExit("-topic is required")
	}

	if *groupID == "" {
		printUsageErrorAndExit("-groupID is required")
	}

	// Determine the initial offset type.
	initialOffset := sarama.OffsetNewest
	_ = initialOffset
	switch *offset {
	case "oldest":
		initialOffset = sarama.OffsetOldest
	case "newest":
		initialOffset = sarama.OffsetNewest
	default:
		printUsageErrorAndExit("-offset should be `oldest` or `newest`")
	}

	// init config
	config := client.NewConfig(logroot.StandardLogger())
	config.SetDebug(*debug)
	config.SetInitialOffset(initialOffset)
	config.SetRecvNotification(true)
	config.SetRecvNotificationChan(make(chan *cluster.Notification)) // channel for notification delivery
	config.SetRecvError(true)
	config.SetRecvErrorChan(make(chan error))                     // channel for error delivery
	config.SetRecvMessageChan(make(chan *client.ConsumerMessage)) // channel for message delivery
	config.SetBrokers(*brokerList)
	config.SetTopics(*topicList)
	config.SetGroup(*groupID)

	// Demonstrate NewConsumer() API to create a new message consumer.
	consumer, err := client.NewConsumer(config, nil)
	if err != nil {
		fmt.Printf("Failed to create a new Kafka consumer: %v", err)
		os.Exit(1)
	}

	// Consume messages in a separate go routine.
	go watchChannels(consumer, config)

	// Wait for the interrupt signal.
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		select {
		case <-signalChan:
			consumer.Close()
			log.StandardLogger().Debug("exiting")
		}
	}()

	fmt.Println("waiting for consumer to close ...")
	consumer.WaitForClose()
	fmt.Println("consumer closed")

}

// watchChannels watches channels configured for delivery of Kafka messages,
// notifications and errors.
func watchChannels(consumer *client.Consumer, cfg *client.Config) {

	for {
		select {
		case notification, more := <-cfg.RecvNotificationChan:
			if more {
				handleNotifcation(consumer, notification)
			}
		case err, more := <-cfg.RecvErrorChan:
			if more {
				fmt.Printf("Message Recv Errored: %v\n", err)
			}
		case msg, more := <-cfg.RecvMessageChan:
			if more {
				messageCallback(consumer, msg, *commit)
			}
		case <-consumer.GetCloseChannel():
			return
		}
	}
}

func handleNotifcation(consumer *client.Consumer, note *cluster.Notification) {
	if note == nil {
		return
	}
	fmt.Println("Rebalanced Consumer at ", time.Now())
	fmt.Println("Claimed: ")
	consumer.PrintNotification(note.Claimed)
	fmt.Println("Released: ")
	consumer.PrintNotification(note.Released)
	fmt.Println("Current: ")
	consumer.PrintNotification(note.Current)

	subs := consumer.Subscriptions()
	fmt.Printf("\n\nCurrent Subscriptions: \n")
	consumer.PrintNotification(subs)
}

func messageCallback(consumer *client.Consumer, msg *client.ConsumerMessage, commitOffset bool) {
	if msg == nil {
		return
	}
	fmt.Printf("Consumer Message - Topic: msg.Topic, Key: %s, Value: %s, Partition: %d Offset: %d\n", string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)

	if commitOffset {
		consumer.MarkOffset(msg, "")
		err := consumer.CommitOffsets()
		if err != nil {
			log.StandardLogger().Errorf("CommitOffset Errored: %v", err)
		}
		log.StandardLogger().Info("Message Offset committed")
	}
}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}
