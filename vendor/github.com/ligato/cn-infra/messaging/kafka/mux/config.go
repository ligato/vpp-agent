package mux

import (
	"github.com/Shopify/sarama"
	"github.com/ghodss/yaml"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/messaging/kafka/client"
	"io/ioutil"
)

// Config holds the settings for kafka multiplexer.
type Config struct {
	Addrs []string `json:"addrs"`
}

// ConsumerFactory produces a consumer for the selected topics in a specified consumer group.
// The reason why a function(factory) is passed to Multiplexer instead of consumer instance is
// that list of topics to be consumed has to be known on consumer initialization.
// Multiplexer calls the function once the list of topics to be consumed is selected.
type ConsumerFactory func(topics []string, groupId string) (*client.Consumer, error)

// ConfigFromFile loads the Kafka multiplexer configuration from the
// specified file. If the specified file is valid and contains
// valid configuration, the parsed configuration is
// returned; otherwise, an error is returned.
func ConfigFromFile(fpath string) (*Config, error) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func getConsumerFactory(config *client.Config) ConsumerFactory {
	return func(topics []string, groupId string) (*client.Consumer, error) {
		config.SetRecvMessageChan(make(chan *client.ConsumerMessage))
		config.Topics = topics
		config.GroupID = groupId
		config.SetInitialOffset(sarama.OffsetOldest)

		return client.NewConsumer(config, nil)
	}
}

// InitMultiplexer initialize and returns new kafka multiplexer based on the supplied config file.
// Name is used as groupId identification of consumer. Kafka allows to store last read offset for
// a groupId. This is leveraged to deliver unread messages after restart.
func InitMultiplexer(configFile string, name string, log logging.Logger) (*Multiplexer, error) {

	var err error
	muxCfg := &Config{[]string{"127.0.0.1:9092"}}
	if configFile != "" {
		muxCfg, err = ConfigFromFile(configFile)
		if err != nil {
			return nil, err
		}
	}
	return InitMultiplexerWithConfig(muxCfg, name, log)
}

// InitMultiplexerWithConfig initialize and returns new kafka multiplexer
// based on the supplied configuration.
// Name is used as groupId identification of consumer. Kafka allows to store last read offset for
// a groupId. This is leveraged to deliver unread messages after restart.
func InitMultiplexerWithConfig(muxConfig *Config, name string, log logging.Logger) (*Multiplexer, error) {

	const errorFmt = "Failed to create Kafka %s, Configured broker(s) %v, Error: '%s'"

	log.WithField("addrs", muxConfig.Addrs).Debug("Kafka connecting")

	config := client.NewConfig(log)
	config.SetSendSuccess(true)
	config.SetSuccessChan(make(chan *client.ProducerMessage))
	config.SetSendError(true)
	config.SetErrorChan(make(chan *client.ProducerError))
	config.Brokers = muxConfig.Addrs

	syncProducer, err := client.NewSyncProducer(config, nil)
	if err != nil {
		log.Errorf(errorFmt, "SyncProducer", muxConfig.Addrs, err)
		return nil, err
	}

	asyncProducer, err := client.NewAsyncProducer(config, nil)
	if err != nil {
		log.Errorf(errorFmt, "AsyncProducer", muxConfig.Addrs, err)
		return nil, err
	}

	return NewMultiplexer(getConsumerFactory(config), syncProducer, asyncProducer, name, log), nil
}
