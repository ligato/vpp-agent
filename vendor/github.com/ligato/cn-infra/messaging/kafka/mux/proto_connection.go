package mux

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/messaging"
	"github.com/ligato/cn-infra/messaging/kafka/client"
)

// ProtoConnection is an entity that provides access to shared producers/consumers of multiplexer.
// The value of message are marshaled and unmarshaled to/from proto.message behind the scene.
type ProtoConnection struct {
	// multiplexer is used for access to kafka brokers
	multiplexer *Multiplexer

	// name identifies the connection
	name string

	// serializer marshals and unmarshals data to/from proto.Message
	serializer keyval.Serializer
}

type protoSyncPublisherKafka struct {
	conn  *ProtoConnection
	topic string
}

type protoAsyncPublisherKafka struct {
	conn        *ProtoConnection
	topic       string
	successChan chan *client.ProducerMessage
	errChan     chan *client.ProducerError
}

// SendSyncMessage sends a message using the sync API
func (conn *ProtoConnection) SendSyncMessage(topic string, key string, value proto.Message) (offset int64, err error) {
	data, err := conn.serializer.Marshal(value)
	if err != nil {
		return 0, err
	}
	msg, err := conn.multiplexer.syncProducer.SendMsg(topic, sarama.StringEncoder(key), sarama.ByteEncoder(data))
	if err != nil {
		return 0, err
	}
	return msg.Offset, err
}

// SendAsyncMessage sends a message using the async API
func (conn *ProtoConnection) SendAsyncMessage(topic string, key string, value proto.Message, meta interface{}, successChan chan *client.ProducerMessage, errChan chan *client.ProducerError) error {
	data, err := conn.serializer.Marshal(value)
	if err != nil {
		return err
	}
	auxMeta := &asyncMeta{successChan: successChan, errorChan: errChan, usersMeta: meta}
	conn.multiplexer.asyncProducer.SendMsg(topic, sarama.StringEncoder(key), sarama.ByteEncoder(data), auxMeta)
	return nil
}

// ConsumeTopic is called to start consuming given topics.
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *ProtoConnection) ConsumeTopic(msgChan chan *client.ProtoConsumerMessage, topics ...string) error {
	conn.multiplexer.rwlock.Lock()
	defer conn.multiplexer.rwlock.Unlock()

	if conn.multiplexer.started {
		return fmt.Errorf("ConsumeTopic can be called only if the multiplexer has not been started yet")
	}

	internalChannel := make(chan *client.ConsumerMessage)

	go func() {
	messageHandler:
		for {
			select {
			case msg := <-internalChannel:
				select {
				case msgChan <- client.NewProtoConsumerMessage(msg, conn.serializer):
				default:
					conn.multiplexer.Warn("Unable to deliver message to consumer")
				}
			case <-conn.multiplexer.closeCh:
				break messageHandler
			}
		}
		close(internalChannel)
	}()

	for _, topic := range topics {
		// check if we have already consumed the topic and partition
		subs, found := conn.multiplexer.mapping[topic]

		if !found {
			subs = &map[string]chan *client.ConsumerMessage{}
			conn.multiplexer.mapping[topic] = subs
		}
		// add subscription to consumerList
		(*subs)[conn.name] = internalChannel
		conn.multiplexer.mapping[topic] = subs
	}
	return nil
}

// StopConsuming cancels the previously created subscription for consuming the topic.
func (conn *ProtoConnection) StopConsuming(topic string) error {
	return conn.multiplexer.stopConsuming(topic, conn.name)
}

// NewSyncPublisher creates a new instance of protoSyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *ProtoConnection) NewSyncPublisher(topic string) messaging.ProtoPublisher {
	return &protoSyncPublisherKafka{conn, topic}
}

// Publish publishes a message into kafka
func (p *protoSyncPublisherKafka) Publish(key string, message proto.Message) error {
	_, err := p.conn.SendSyncMessage(p.topic, key, message)
	return err
}

// NewAsyncPublisher creates a new instance of protoAsyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *ProtoConnection) NewAsyncPublisher(topic string, successCh chan *client.ProducerMessage, errorCh chan *client.ProducerError) messaging.ProtoPublisher {
	return &protoAsyncPublisherKafka{conn, topic, successCh, errorCh}
}

// Publish publishes a message into kafka
func (p *protoAsyncPublisherKafka) Publish(key string, message proto.Message) error {
	return p.conn.SendAsyncMessage(p.topic, key, message, nil, p.successChan, p.errChan)
}
