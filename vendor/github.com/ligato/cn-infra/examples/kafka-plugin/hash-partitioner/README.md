# kafka-plugin Hash-Partitioner

A simple example demonstrating the usage of Kafka plugin API with
automatic (hash-based) partitioning.

### Requirements

To start the example you have to have Kafka broker running first.
if you don't have it installed locally you can use the following docker
image.
```
sudo docker run -p 2181:2181 -p 9092:9092 --name kafka --rm \
   --env ADVERTISED_HOST=172.17.0.1 --env ADVERTISED_PORT=9092 spotify/kafka
```

It will bring up Kafka broker listening on port 9092 for client
communication.

### Usage

To run the example, type:
```
go run main.go deps.go [-kafka-config <config-filepath>]
```

If `kafka-config` is unspecified, the application will try to search
for `kafka.conf` in the current working directory.
If the configuration file cannot be loaded, the initialization
of the kafka plugin will be skipped and the example scenario will thus
not execute in its entirety.
