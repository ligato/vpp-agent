# VPP Agent examples

There are several `main.go` files used as an illustration of the VPP-Agent functionality. Most of the examples show
a very simple use case using the real ETCD, GOVPP and/or Kafka plugins, so for specific examples, they have to be started at first.

Current examples:
* **[govpp_call](govpp_call/main.go)** is an example of a plugin with a configurator and a VPP channel. The example displays 
transformation between the model data and the VPP binary API which are then sent to the VPP
* **[idx_mapping_lookup](idx_mapping_lookup/main.go)** shows an usage of the name to index mapping (registration, read by name/index, 
un-registration)
* **[idx_mapping_watcher](idx_mapping_watcher/main.go)** shows how to watch on changes in the name to index mapping
* **[localclient_vpp](localclient_vpp/main.go)** demonstrates the use of the localclient package to transport example configuration into
    the VPP plugins running inside the same agent instance (i.e. the same OS process). Behind the scenes the configuration
    data are transported by means of go channels.
* **[localclient_linux](localclient_linux/main.go)** demonstrates the use of the localclient package to transport example configuration into
    the linux and VPP plugins running inside the same agent instance (i.e. the same OS process). Behind the scenes the configuration
    data are transported by means of go channels.

* **[other examples](https://github.com/ligato/cn-infra/tree/master/examples)**
 
## How to run example
 
 **1. Start ETCD server on localhost**
 
  ```
  sudo docker run -p 2379:2379 --name etcd --rm 
  quay.io/coreos/etcd:v3.0.16 /usr/local/bin/etcd \
  -advertise-client-urls http://0.0.0.0:2379 \
  -listen-client-urls http://0.0.0.0:2379
  ```
  
 **2. Start Kafka**

 ```
 sudo docker run -p 2181:2181 -p 9092:9092 --name kafka --rm \
  --env ADVERTISED_HOST=172.17.0.1 --env ADVERTISED_PORT=9092 spotify/kafka
 ```
 
 **3. Start VPP**
 ```
 vpp unix { interactive } plugins { plugin dpdk_plugin.so { disable } }
 ```
 
 **4. Start desired example**

 Example can be started now from particular directory.
 ```
 go run main.go  \
 --etcdv3-config=/opt/vpp-agent/dev/etcd.conf \
 --kafka-config=/opt/vpp-agent/dev/kafka.conf
 ```
