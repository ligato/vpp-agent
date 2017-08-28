# Flavors

A [flavors](../docs/guidelines/PLUGIN_FLAVORS.md) is a reusable collection of plugins 
with initialized [dependencies](../docs/guidelines/PLUGIN_DEPENDENCIES.md). 

Most importatnt CN-Infra flavors:
* [local flavor](local) - a minimal set of plugins. It just initializes logging & statuchek.
  It is useful for embedding agent plugins to different projects that use their own infrasturcure.
* [RPC flavor](rpc) - a collection of plugins that exposes RPCs. It also register management API for:
  * status check (RPCs probed from systems such as K8s)
  * logging (for changing log level at runtime remotely)
* [etcd + Kafka flavor](etcdkafka) - is combination of Kafka & etcd flavor (reused in many examples)
* [all connectors flavor](allcon) - is combination of all following flavors with local and RPC flavor.

There are following individual connector flavors:
* [Cassandra flavor](cassandra) - adds Cassadnra client plugin related instances to the [local flavor](local)
* [etcd flavor](etcd) - adds etcd client plugin related instances to the [local flavor](local) 
  the [local flavor](local)
* [Redis flavor](redis) - adds Redis client plugin related instances to the [local flavor](local)
* [Kafka flavor](kafka) - adds Kafka client plugin related instances to the [local flavor](local)

  
The following diagram shows:
* plugins that are part of the flavor
* initialized (injected) [statuscheck](../health/statuscheck) dependency 
  inside [etcd client plugin](../db/keyval/etcdv3) and [Kafka client plugin](../messaging/kafka)
* [etcd + Kafka flavor](etcdkafka) extends [local flavor](local) 

![flavors](../docs/imgs/flavors.png)
