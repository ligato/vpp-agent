package mux

import (
	"fmt"
	"sync"

	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"github.com/ligato/cn-infra/utils/safeclose"
)

// Multiplexer encapsulates clients to kafka cluster (syncProducer, asyncProducer, consumer).
// It allows to create multiple Connections that use multiplexer's clients for communication
// with kafka cluster. The aim of Multiplexer is to decrease the number of connections needed.
// The set of topics to be consumed by Connections needs to be selected before the underlying
// consumer in Multiplexer is started. Once the Multiplexer's consumer has been
// started new topics can not be added.
type Multiplexer struct {
	logging.Logger
	// consumer used by the Multiplexer
	consumer *client.Consumer
	// syncProducer used by the Multiplexer
	syncProducer *client.SyncProducer
	// asyncProducer used by the Multiplexer
	asyncProducer *client.AsyncProducer

	// name is used for identification of stored last consumed offset in kafka. This allows
	// to follow up messages after restart.
	name string

	// guards access to mapping and started flag
	rwlock sync.RWMutex

	// started denotes whether the multiplexer is dispatching the messages or accepting subscriptions to
	// consume a topic. Once the multiplexer is started, new subscription can not be added.
	started bool

	// Mapping provides the mapping of subscribed consumers organized by topics(key of the first map)
	// name of the consumer(key of the second map)
	mapping map[string]*map[string]func(*client.ConsumerMessage)

	// factory that crates consumer used in the Multiplexer
	consumerFactory func(topics []string, groupId string) (*client.Consumer, error)
	closeCh         chan struct{}
}

// asyncMeta is auxiliary structure used by Multiplexer to distribute consumer messages
type asyncMeta struct {
	successClb func(*client.ProducerMessage)
	errorClb   func(error *client.ProducerError)
	usersMeta  interface{}
}

// NewMultiplexer creates new instance of Kafka Multiplexer
func NewMultiplexer(consumerFactory ConsumerFactory, syncP *client.SyncProducer, asyncP *client.AsyncProducer, name string, log logging.Logger) *Multiplexer {
	cl := &Multiplexer{consumerFactory: consumerFactory,
		Logger:        log,
		syncProducer:  syncP,
		asyncProducer: asyncP,
		name:          name,
		mapping:       map[string]*map[string]func(*client.ConsumerMessage){},
		closeCh:       make(chan struct{}),
	}

	go cl.watchAsyncProducerChannels()
	return cl
}

func (mux *Multiplexer) watchAsyncProducerChannels() {
	for {
		select {
		case err := <-mux.asyncProducer.Config.ErrorChan:
			mux.Println("Failed to produce message", err.Err)
			errMsg := err.ProducerMessage

			if errMeta, ok := errMsg.Metadata.(*asyncMeta); ok && errMeta.errorClb != nil {
				err.ProducerMessage.Metadata = errMeta.usersMeta
				errMeta.errorClb(err)
			}
		case success := <-mux.asyncProducer.Config.SuccessChan:

			if succMeta, ok := success.Metadata.(*asyncMeta); ok && succMeta.successClb != nil {
				success.Metadata = succMeta.usersMeta
				succMeta.successClb(success)
			}
		case <-mux.asyncProducer.GetCloseChannel():
			mux.Debug("Closing watch loop for async producer")
		}
	}
}

// Start should be called once all the Connections have been subscribed
// for topic consumption. An attempt to start consuming a topic after the multiplexer is started
// returns an error.
func (mux *Multiplexer) Start() error {
	mux.rwlock.Lock()
	defer mux.rwlock.Unlock()
	var err error

	if mux.started {
		return fmt.Errorf("Multiplexer has been started already")
	}

	// block further consumer consumers
	mux.started = true

	var topics []string

	for topic := range mux.mapping {
		topics = append(topics, topic)
	}

	if len(topics) == 0 {
		mux.Debug("No topics to be consumed")
		return nil
	}

	mux.WithFields(logging.Fields{"topics": topics}).Debug("Consuming started")

	mux.consumer, err = mux.consumerFactory(topics, mux.name)
	if err != nil {
		mux.Error(err)
		return err
	}

	go mux.genericConsumer()

	return nil
}

// Close cleans up the resources used by the Multiplexer
func (mux *Multiplexer) Close() {
	close(mux.closeCh)
	safeclose.Close(mux.consumer)
	safeclose.Close(mux.syncProducer)
	safeclose.Close(mux.asyncProducer)
}

// NewConnection creates instance of the Connection that will be provide access to shared Multiplexer's clients.
func (mux *Multiplexer) NewConnection(name string) *Connection {
	return &Connection{multiplexer: mux, name: name}
}

// NewProtoConnection creates instance of the ProtoConnection that will be provide access to shared Multiplexer's clients.
func (mux *Multiplexer) NewProtoConnection(name string, serializer keyval.Serializer) *ProtoConnection {
	return &ProtoConnection{multiplexer: mux, serializer: serializer, name: name}
}

func (mux *Multiplexer) propagateMessage(msg *client.ConsumerMessage) {
	mux.rwlock.RLock()
	defer mux.rwlock.RUnlock()

	if msg == nil {
		return
	}
	cons, found := mux.mapping[msg.Topic]

	// notify consumers
	if found {
		for _, clb := range *cons {
			// if we are not able to write into the channel we should skip the receiver
			// and report an error to avoid deadlock
			mux.Debug("offset ", msg.Offset, string(msg.Value), string(msg.Key), msg.Partition)

			clb(msg)
		}
	}
}

// genericConsumer handles incoming messages to the multiplexer and distributes them among the subscribers
func (mux *Multiplexer) genericConsumer() {
	mux.Debug("Generic consumer started")
	for {
		select {
		case <-mux.consumer.GetCloseChannel():
			mux.Debug("Closing consumer")
			return
		case msg := <-mux.consumer.Config.RecvMessageChan:
			mux.Debug("Kafka message received")
			mux.propagateMessage(msg)
			// Mark offset as read. If the Multiplexer is restarted it
			// continues to receive message after the last committed offset.
			mux.consumer.MarkOffset(msg, "")
		case err := <-mux.consumer.Config.RecvErrorChan:
			mux.Error("Received partitionConsumer error ", err)
		}
	}

}

func (mux *Multiplexer) stopConsuming(topic string, name string) error {
	mux.rwlock.Lock()
	defer mux.rwlock.Unlock()

	subs, found := mux.mapping[topic]
	if !found {
		return fmt.Errorf("Topic %s was not consumed by '%s'", topic, name)
	}
	_, found = (*subs)[name]
	if !found {
		return fmt.Errorf("Topic %s was not consumed by '%s'", topic, name)
	}
	delete(*subs, name)
	return nil
}
