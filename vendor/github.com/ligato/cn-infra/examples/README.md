# CN-infra examples

There are several `main.go` files used as an illustration of the cn-infra
functionality. Most of the examples show a very simple use case using the
real ETCD and/or Kafka plugins, so for specific examples, they have to be
started first.

Current examples:
* **[etcd](etcdv3_broker)** uses the ETCD data broker to write data into 
  ETCD, which are then caught by the watcher
* **[flags](flags/main.go)** example registers flags and shows their 
  runtime values
* **[kafka](kafka/main.go)** creates a simple plugin which registers a 
  Kafka consumer and sends a test notification
* **[logs](logs_logrus)** shows how to use the logger and wotk wiht 
  log levels

## How to run an example

 **1. Start ETCD server on localhost**

  ```
  sudo docker run -p 2379:2379 --name etcd --rm \
  quay.io/coreos/etcd:v3.0.16 /usr/local/bin/etcd \
  -advertise-client-urls http://0.0.0.0:2379 \
  -listen-client-urls http://0.0.0.0:2379
  ```

 **2. Start Kafka**

 ```
 sudo docker run -p 2181:2181 -p 9092:9092 --name kafka --rm \
  --env ADVERTISED_HOST=172.17.0.1 --env ADVERTISED_PORT=9092 spotify/kafka
 ```

 **3. Start the desired example**

 Each example can be started now from its directory.
 ```
 go run main.go  \
 --etcdv3-config=/opt/vnf-agent/dev/etcd.conf \
 --kafka-config=/opt/vnf-agent/dev/kafka.conf
 ```
