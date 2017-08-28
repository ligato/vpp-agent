# kafka-console-producer

A simple command line tool to produce a single message to Kafka.

### Usage

Minimum invocation
```
    kafka-console-producer -topic=test -value=value -brokers=kafka1:9092
```

In order to configure kafka brokers environment variable can be used
```
    export KAFKA_PEERS=kafka1:9092,kafka2:9092,kafka3:9092
    kafka-console-producer -topic=test -value=value
```

The value can be passed from stdin by using pipes
```
    echo "hello world" | kafka-console-producer -topic=test
```

The key can be specified:
```
    echo "hello world" | kafka-console-producer -topic=test -key=key
```

Partitioning: by default, kafka-console-producer will partition as follows:
 - manual partitioning if a -partition is provided
 - hash partitioning by key if a -key is provided
 - random partioning otherwise.
 
You can override this using the -partitioner argument:
```
    echo "hello world" | kafka-console-producer -topic=test -key=key -partitioner=random
```

To display all command line options
```
    kafka-console-producer -help
```