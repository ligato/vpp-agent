package mux

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/ligato/cn-infra/messaging/kafka/client"
)

// Connection is an entity that provides access to shared producers/consumers of multiplexer.
type Connection struct {
	// multiplexer is used for access to kafka brokers
	multiplexer *Multiplexer

	// name identifies the connection
	name string
}

// BytesPublisher allows to publish a message of type []bytes into messaging system.
type BytesPublisher interface {
	Publish(key string, data []byte) error
}

type bytesSyncPublisherKafka struct {
	conn      *Connection
	topic     string
	partition int32
}

type bytesAsyncPublisherKafka struct {
	conn         *Connection
	topic        string
	partition    int32
	succCallback func(*client.ProducerMessage)
	errCallback  func(*client.ProducerError)
}

// ConsumeTopic is called to start consuming of a topic.
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *Connection) ConsumeTopic(msgClb func(message *client.ConsumerMessage), topics ...string) error {
	conn.multiplexer.rwlock.Lock()
	defer conn.multiplexer.rwlock.Unlock()

	if conn.multiplexer.started {
		return fmt.Errorf("ConsumeTopic can be called only if the multiplexer has not been started yet")
	}

	for _, topic := range topics {
		// check if we have already consumed the topic
		var found bool
		var subs *consumerSubscription
	LoopSubs:
		for _, subscription := range conn.multiplexer.mapping {
			if subscription.manual == true {
				// do not mix dynamic and manual mode
				continue
			}
			if subscription.topic == topic {
				found = true
				subs = subscription
				break LoopSubs
			}
		}

		if !found {
			subs = &consumerSubscription{
				manual:         false, // non-manual example
				topic:          topic,
				connectionName: conn.name,
				byteConsMsg:    msgClb,
			}
			// subscribe new topic
			conn.multiplexer.mapping = append(conn.multiplexer.mapping, subs)
		}

		// add subscription to consumerList
		subs.byteConsMsg = msgClb
	}

	return nil
}

// ConsumeTopicOnPartition is called to start consuming given topic on partition with offset
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *Connection) ConsumeTopicOnPartition(msgClb func(message *client.ConsumerMessage), topic string, partition int32, offset int64) error {
	conn.multiplexer.rwlock.Lock()
	defer conn.multiplexer.rwlock.Unlock()

	if conn.multiplexer.started {
		return fmt.Errorf("ConsumeTopicOnPartition can be called only if the multiplexer has not been started yet")
	}

	// check if we have already consumed the topic on partition and offset
	var found bool
	var subs *consumerSubscription

	for _, subscription := range conn.multiplexer.mapping {
		if subscription.manual == false {
			// do not mix dynamic and manual mode
			continue
		}
		if subscription.topic == topic && subscription.partition == partition && subscription.offset == offset {
			found = true
			subs = subscription
			break
		}
	}

	if !found {
		subs = &consumerSubscription{
			manual:         true, // manual example
			topic:          topic,
			partition:      partition,
			offset:         offset,
			connectionName: conn.name,
			byteConsMsg:    msgClb,
		}
		// subscribe new topic on partition
		conn.multiplexer.mapping = append(conn.multiplexer.mapping, subs)
	}

	// add subscription to consumerList
	subs.byteConsMsg = msgClb

	return nil
}

// StopConsuming cancels the previously created subscription for consuming the topic.
func (conn *Connection) StopConsuming(topic string) error {
	return conn.multiplexer.stopConsuming(topic, conn.name)
}

// StopConsumingPartition cancels the previously created subscription for consuming the topic, partition and offset
func (conn *Connection) StopConsumingPartition(topic string, partition int32, offset int64) error {
	return conn.multiplexer.stopConsumingPartition(topic, partition, offset, conn.name)
}

// SendSyncByte sends a message that uses byte encoder using the sync API
func (conn *Connection) SendSyncByte(topic string, partition int32, key []byte, value []byte) (offset int64, err error) {
	return conn.SendSyncMessage(topic, partition, sarama.ByteEncoder(key), sarama.ByteEncoder(value))
}

// SendSyncString sends a message that uses string encoder using the sync API
func (conn *Connection) SendSyncString(topic string, partition int32, key string, value string) (offset int64, err error) {
	return conn.SendSyncMessage(topic, partition, sarama.StringEncoder(key), sarama.StringEncoder(value))
}

//SendSyncMessage sends a message using the sync API
func (conn *Connection) SendSyncMessage(topic string, partition int32, key client.Encoder, value client.Encoder) (offset int64, err error) {
	msg, err := conn.multiplexer.syncProducer.SendMsg(topic, partition, key, value)
	if err != nil {
		return 0, err
	}
	return msg.Offset, err
}

// SendAsyncByte sends a message that uses byte encoder using the async API
func (conn *Connection) SendAsyncByte(topic string, partition int32, key []byte, value []byte, meta interface{}, successClb func(*client.ProducerMessage), errClb func(*client.ProducerError)) {
	conn.SendAsyncMessage(topic, partition, sarama.ByteEncoder(key), sarama.ByteEncoder(value), meta, successClb, errClb)
}

// SendAsyncString sends a message that uses string encoder using the async API
func (conn *Connection) SendAsyncString(topic string, partition int32, key string, value string, meta interface{}, successClb func(*client.ProducerMessage), errClb func(*client.ProducerError)) {
	conn.SendAsyncMessage(topic, partition, sarama.StringEncoder(key), sarama.StringEncoder(value), meta, successClb, errClb)
}

// SendAsyncMessage sends a message using the async API
func (conn *Connection) SendAsyncMessage(topic string, partition int32, key client.Encoder, value client.Encoder, meta interface{}, successClb func(*client.ProducerMessage), errClb func(*client.ProducerError)) {
	auxMeta := &asyncMeta{successClb: successClb, errorClb: errClb, usersMeta: meta}
	conn.multiplexer.asyncProducer.SendMsg(topic, partition, key, value, auxMeta)
}

// NewSyncPublisher creates a new instance of bytesSyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *Connection) NewSyncPublisher(topic string) (BytesPublisher, error) {
	if conn.multiplexer.partitioner == client.Manual {
		return nil, fmt.Errorf("unable to use default sync publisher with 'manual' partitioner")
	}
	return &bytesSyncPublisherKafka{conn, topic, DefPartition}, nil
}

// NewSyncPublisherToPartition creates a new instance of bytesSyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *Connection) NewSyncPublisherToPartition(topic string, partition int32) (BytesPublisher, error) {
	if conn.multiplexer.partitioner != client.Manual {
		return nil, fmt.Errorf("sync publisher to partition can be used only with 'manual' partitioner")
	}
	return &bytesSyncPublisherKafka{conn, topic, partition}, nil
}

// Put publishes a message into kafka
func (p *bytesSyncPublisherKafka) Publish(key string, data []byte) error {
	_, err := p.conn.SendSyncByte(p.topic, p.partition, []byte(key), data)
	return err
}

// NewAsyncPublisher creates a new instance of bytesAsyncPublisherKafka that allows to publish async kafka messages using common messaging API
func (conn *Connection) NewAsyncPublisher(topic string, successClb func(*client.ProducerMessage), errorClb func(err *client.ProducerError)) (BytesPublisher, error) {
	if conn.multiplexer.partitioner == client.Manual {
		return nil, fmt.Errorf("unable to use default async publisher with 'manual' partitioner")
	}
	return &bytesAsyncPublisherKafka{conn, topic, DefPartition, successClb, errorClb}, nil
}

// NewAsyncPublisherToPartition creates a new instance of bytesAsyncPublisherKafka that allows to publish async kafka messages using common messaging API
func (conn *Connection) NewAsyncPublisherToPartition(topic string, partition int32, successClb func(*client.ProducerMessage), errorClb func(err *client.ProducerError)) (BytesPublisher, error) {
	if conn.multiplexer.partitioner != client.Manual {
		return nil, fmt.Errorf("async publisher to partition can be used only with 'manual' partitioner")
	}
	return &bytesAsyncPublisherKafka{conn, topic, partition, successClb, errorClb}, nil
}

// Put publishes a message into kafka
func (p *bytesAsyncPublisherKafka) Publish(key string, data []byte) error {
	p.conn.SendAsyncMessage(p.topic, p.partition, sarama.StringEncoder(key), sarama.ByteEncoder(data), nil, p.succCallback, p.errCallback)
	return nil
}
