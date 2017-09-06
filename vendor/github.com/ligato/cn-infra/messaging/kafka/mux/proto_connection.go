package mux

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
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
	conn      *ProtoConnection
	topic     string
	partition int32
}

type protoAsyncPublisherKafka struct {
	conn         *ProtoConnection
	topic        string
	partition    int32
	succCallback func(messaging.ProtoMessage)
	errCallback  func(messaging.ProtoMessageErr)
}

// SendSyncMessage sends a message using the sync API
func (conn *ProtoConnection) SendSyncMessage(topic string, partition int32, key string, value proto.Message) (offset int64, err error) {
	data, err := conn.serializer.Marshal(value)
	if err != nil {
		return 0, err
	}
	msg, err := conn.multiplexer.syncProducer.SendMsg(topic, partition, sarama.StringEncoder(key), sarama.ByteEncoder(data))
	if err != nil {
		return 0, err
	}
	return msg.Offset, err
}

// SendAsyncMessage sends a message using the async API
func (conn *ProtoConnection) SendAsyncMessage(topic string, partition int32, key string, value proto.Message, meta interface{}, successClb func(messaging.ProtoMessage), errClb func(messaging.ProtoMessageErr)) error {
	data, err := conn.serializer.Marshal(value)
	if err != nil {
		return err
	}
	succByteClb := func(msg *client.ProducerMessage) {
		protoMsg := &client.ProtoProducerMessage{
			ProducerMessage: msg,
			Serializer:      conn.serializer,
		}
		successClb(protoMsg)
	}

	errByteClb := func(msg *client.ProducerError) {
		protoMsg := &client.ProtoProducerMessageErr{
			ProtoProducerMessage: &client.ProtoProducerMessage{
				ProducerMessage: msg.ProducerMessage,
				Serializer:      conn.serializer,
			},
			Err: msg.Err,
		}
		errClb(protoMsg)
	}

	auxMeta := &asyncMeta{successClb: succByteClb, errorClb: errByteClb, usersMeta: meta}
	conn.multiplexer.asyncProducer.SendMsg(topic, partition, sarama.StringEncoder(key), sarama.ByteEncoder(data), auxMeta)
	return nil
}

// ConsumeTopic is called to start consuming given topics.
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *ProtoConnection) ConsumeTopic(msgClb func(messaging.ProtoMessage), topics ...string) error {
	conn.multiplexer.rwlock.Lock()
	defer conn.multiplexer.rwlock.Unlock()

	if conn.multiplexer.started {
		return fmt.Errorf("ConsumeTopic can be called only if the multiplexer has not been started yet")
	}

	byteClb := func(bm *client.ConsumerMessage) {
		pm := client.NewProtoConsumerMessage(bm, conn.serializer)
		msgClb(pm)
	}

	for _, topic := range topics {
		// check if we have already consumed the topic and partition
		subs, found := conn.multiplexer.mapping[topic]

		if !found {
			subs = &map[string]func(*client.ConsumerMessage){}
			conn.multiplexer.mapping[topic] = subs
		}
		// add subscription to consumerList
		(*subs)[conn.name] = byteClb
		conn.multiplexer.mapping[topic] = subs
	}
	return nil
}

// ConsumePartition is called to start consuming given topic on given partition and offset.
// Function can be called until the multiplexer is started, it returns an error otherwise.
// The provided channel should be buffered, otherwise messages might be lost.
func (conn *ProtoConnection) ConsumePartition(msgClb func(messaging.ProtoMessage), topic string,
	partition int32, offset int64) error {
	conn.multiplexer.Warn("Partition selection not supported yet")
	return conn.ConsumeTopic(msgClb, topic)
}

// StopConsuming cancels the previously created subscription for consuming the topic.
func (conn *ProtoConnection) StopConsuming(topic string) error {
	return conn.multiplexer.stopConsuming(topic, conn.name)
}

// Watch is an alias for ConsumeTopic method. The alias was added in order to conform to messaging.Mux interface.
func (conn *ProtoConnection) Watch(msgClb func(messaging.ProtoMessage), topics ...string) error {
	return conn.ConsumeTopic(msgClb, topics...)
}

// WatchPartition is an alias for ConsumePartition method. The alias was added in order
// to conform to messaging.Mux interface.
func (conn *ProtoConnection) WatchPartition(msgClb func(messaging.ProtoMessage), topic string,
	partition int32, offset int64) error {
	return conn.ConsumePartition(msgClb, topic, partition, offset)
}

// StopWatch is an alias for StopConsuming method. The alias was added in order to conform to messaging.Mux interface.
func (conn *ProtoConnection) StopWatch(topic string) error {
	return conn.StopConsuming(topic)
}

// NewSyncPublisher creates a new instance of protoSyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *ProtoConnection) NewSyncPublisher(topic string) messaging.ProtoPublisher {
	return &protoSyncPublisherKafka{conn, topic, DefPartition}
}

// NewSyncPublisherToPartition creates a new instance of protoSyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *ProtoConnection) NewSyncPublisherToPartition(topic string, partition int32) messaging.ProtoPublisher {
	return &protoSyncPublisherKafka{conn, topic, partition}
}

// Put publishes a message into kafka
func (p *protoSyncPublisherKafka) Put(key string, message proto.Message, opts ...datasync.PutOption) error {
	_, err := p.conn.SendSyncMessage(p.topic, p.partition, key, message)
	return err
}

// NewAsyncPublisher creates a new instance of protoAsyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *ProtoConnection) NewAsyncPublisher(topic string, successClb func(messaging.ProtoMessage), errorClb func(messaging.ProtoMessageErr)) messaging.ProtoPublisher {
	return &protoAsyncPublisherKafka{conn, topic, DefPartition, successClb, errorClb}
}

// NewAsyncPublisherToPartition creates a new instance of protoAsyncPublisherKafka that allows to publish sync kafka messages using common messaging API
func (conn *ProtoConnection) NewAsyncPublisherToPartition(topic string, partition int32, successClb func(messaging.ProtoMessage), errorClb func(messaging.ProtoMessageErr)) messaging.ProtoPublisher {
	return &protoAsyncPublisherKafka{conn, topic, partition, successClb, errorClb}
}

// Put publishes a message into kafka
func (p *protoAsyncPublisherKafka) Put(key string, message proto.Message, opts ...datasync.PutOption) error {
	return p.conn.SendAsyncMessage(p.topic, p.partition, key, message, nil, p.succCallback, p.errCallback)
}
