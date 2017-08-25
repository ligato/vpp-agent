# kafka-cli-asyncproducer

A simple command line tool to send messages to Kafka using an asynchronous Kafka producer

### Usage

- Minimum invocation
  `kafka-cli-asyncproducer -brokers=localhost:9092`

- If environment variable KAFKA_PEERS is defined and `-brokers` is not specified then the KAFKA_PEERS environment variable will be used to determine the brokers.
  `export KAFKA_PEERS=kafka1:9092,kafka2:9092,kafka3:9092`

- A prompt will be displayed:
    Enter command [quit|message]:
        - enter quit to end
        - enter message, or `<enter>` to send a message

- If message is entered the the following prompts will be displayed:
        - enter message (enter the message text)
        - enter key (enter the message key or `<enter>`)
        - enter meta (enter the message meta data or `<enter>`)

- To terminate this producer enter `ctrl-c`. The message `closing producer ...` will be displayed.

- When a message is successfully sent then `message sent successfully - <msg>` will be displayed.

- If a message errors then `message errored - <error>` will be displayed.
- If `quit` is entered, the consumer will be closed and `ended successfully` will be displayed.

### Options

-brokers
: The comma separated list of brokers in the Kafka cluster. You can also set the KAFKA_PEERS environment variable in-place of setting this option.

-partitioner
: default: **hash**. The partitioning scheme to use. Can be `hash`, `manual`, or `random`

-partition
: The partition to produce to. Only used if `partitioner=manual`. If the partition is > -1 then the partitioner will automatically be set to **manual**.
 
debug
: Turns-on debug logging.

silent
: Turns-off printing the message's topic, partition, and offset to stdout.

