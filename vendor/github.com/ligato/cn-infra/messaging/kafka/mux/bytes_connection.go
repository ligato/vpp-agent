package mux

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/messaging/kafka/client"
)

// Connection is an entity that provides access to shared producers/consumers of multiplexer.
type Connection struct {
	// multiplexer is used for access to kafka brokers
	multiplexer *Multiplexer

	// name identifies the connection
	name string
}

type bytesSyncPublisherKafka struct {
	conn  *Connection
	topic string
}

type bytesAsyncPublisherKafka struct {
	conn        *Connection
	topic       string
	successChan chan *client.ProducerMessage
	errChan     chan *client.ProducerError
}

// ConsumeTopic is called to start consuming of a topic.
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *Connection) ConsumeTopic(msgChan chan *client.ConsumerMessage, topics ...string) error {
	conn.multiplexer.rwlock.Lock()
	defer conn.multiplexer.rwlock.Unlock()

	if conn.multiplexer.started {
		return fmt.Errorf("ConsumeTopic can be called only if the multiplexer has not been started yet")
	}

	for _, topic := range topics {
		// check if we have already consumed the topic and partition
		subs, found := conn.multiplexer.mapping[topic]

		if !found {
			subs = &map[string]chan *client.ConsumerMessage{}
			conn.multiplexer.mapping[topic] = subs
		}
		// add subscription to consumerList
		(*subs)[conn.name] = msgChan
		conn.multiplexer.mapping[topic] = subs
	}
	return nil
}

// StopConsuming cancels the previously created subscription for consuming the topic.
func (conn *Connection) StopConsuming(topic string) error {
	return conn.multiplexer.stopConsuming(topic, conn.name)
}

// SendSyncByte sends a message that uses byte encoder using the sync API
func (conn *Connection) SendSyncByte(topic string, key []byte, value []byte) (offset int64, err error) {
	return conn.SendSyncMessage(topic, sarama.ByteEncoder(key), sarama.ByteEncoder(value))
}

// SendSyncString sends a message that uses string encoder using the sync API
func (conn *Connection) SendSyncString(topic string, key string, value string) (offset int64, err error) {
	return conn.SendSyncMessage(topic, sarama.StringEncoder(key), sarama.StringEncoder(value))
}

//SendSyncMessage sends a message using the sync API
func (conn *Connection) SendSyncMessage(topic string, key client.Encoder, value client.Encoder) (offset int64, err error) {
	msg, err := conn.multiplexer.syncProducer.SendMsg(topic, key, value)
	if err != nil {
		return 0, err
	}
	return msg.Offset, err
}

// SendAsyncByte sends a message that uses byte encoder using the async API
func (conn *Connection) SendAsyncByte(topic string, key []byte, value []byte, meta interface{}, successChan chan *client.ProducerMessage, errChan chan *client.ProducerError) {
	conn.SendAsyncMessage(topic, sarama.ByteEncoder(key), sarama.ByteEncoder(value), meta, successChan, errChan)
}

// SendAsyncString sends a message that uses string encoder using the async API
func (conn *Connection) SendAsyncString(topic string, key string, value string, meta interface{}, successChan chan *client.ProducerMessage, errChan chan *client.ProducerError) {
	conn.SendAsyncMessage(topic, sarama.StringEncoder(key), sarama.StringEncoder(value), meta, successChan, errChan)
}

// SendAsyncMessage sends a message using the async API
func (conn *Connection) SendAsyncMessage(topic string, key client.Encoder, value client.Encoder, meta interface{}, successChan chan *client.ProducerMessage, errChan chan *client.ProducerError) {
	auxMeta := &asyncMeta{successChan: successChan, errorChan: errChan, usersMeta: meta}
	conn.multiplexer.asyncProducer.SendMsg(topic, key, value, auxMeta)
}

// NewSyncPublisher creates a new instance of bytesSyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *Connection) NewSyncPublisher(topic string) messaging.BytesPublisher {
	return &bytesSyncPublisherKafka{conn, topic}
}

// Publish publishes a message into kafka
func (p *bytesSyncPublisherKafka) Publish(key string, data []byte) error {
	_, err := p.conn.SendSyncByte(p.topic, []byte(key), data)
	return err
}

// NewAsyncPublisher creates a new instance of bytesAsyncPublisherKafka that allows to publish async kafka messages using common messaging API
func (conn *Connection) NewAsyncPublisher(topic string, successCh chan *client.ProducerMessage, errorCh chan *client.ProducerError) messaging.BytesPublisher {
	return &bytesAsyncPublisherKafka{conn, topic, successCh, errorCh}
}

// Publish publishes a message into kafka
func (p *bytesAsyncPublisherKafka) Publish(key string, data []byte) error {
	p.conn.SendAsyncMessage(p.topic, sarama.StringEncoder(key), sarama.ByteEncoder(data), nil, p.successChan, p.errChan)
	return nil
}
