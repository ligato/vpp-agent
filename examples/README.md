# VPP Agent examples

This folder contains several examples that illustrate various aspects of
VPP Agent's functionality. Each example is structured as an individual 
executable with its own `main.go` file. Each example focuses on a very 
simple use case. All examples use ETCD, GOVPP and Kafka, so please make
sure there are running instances of ETCD, Kafka and VPP before starting
an example.

Current examples:
* **[govpp_call](govpp_call/main.go)** is an example of a plugin with a 
  configurator and a channel to send/receive data to VPP. The example 
  shows how to transform northbound model data to VPP binary API calls. 
* **[idx_mapping_lookup](idx_mapping_lookup/main.go)** shows the usage 
  of the name-to-index mapping (registration, read by name/index, 
  un-registration)
* **[idx_mapping_watcher](idx_mapping_watcher/main.go)** shows how to 
  watch on changes in a name-to-index mapping
* **[localclient_vpp](localclient_vpp/main.go)** demonstrates how to use
  the localclient package to push example configuration into VPP plugins 
  that run in the same agent instance (i.e. in the same OS process). 
  Behind the scenes, configuration data is transported via go channels.
* **[localclient_linux](localclient_linux/main.go)** demonstrates how to
  use the localclient package to push example configuration into linux and
  VPP plugins running within the same agent instance (i.e. within the same 
  OS process). Behind the scenes the configuration data is transported via
  go channels.
* **[remoteclient_grpc_vpp](remoteclient_grpc_vpp/main.go)** demonstrates how to
  use the remoteclient package to push example configuration into
  VPP default plugins running within different vpp-agent OS process.

* **[CN-Infra  examples][1]** demonstrate how to use the CN-Infra platform
  plugins.
  
## How to run an example
 
 **1. Start the ETCD server on localhost**
 
  ```
  sudo docker run -p 2379:2379 --name etcd --rm 
  quay.io/coreos/etcd:v3.1.0 /usr/local/bin/etcd \
  -advertise-client-urls http://0.0.0.0:2379 \
  -listen-client-urls http://0.0.0.0:2379
  ```
  
 **2. Start Kafka on localhost**

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
[1]: https://github.com/ligato/cn-infra/tree/master/examples 