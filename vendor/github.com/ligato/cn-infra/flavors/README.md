# Flavors

A flavor is a reusable collection of plugins with initialized 
[dependencies](../docs/guidelines/PLUGIN_DEPENDENCIES.md). CN-Infra provides
the following [flavors](../docs/guidelines/PLUGIN_FLAVORS.md):
* [local flavor](local) - a minimal set of plugins. It just initializes logging & statuchek.
  It is useful for embedding agent plugins to different projects that use their own infrasturcure.
* [RPC flavor](rpc) - a collection of plugins that exposes RPCs. It also register management API for:
  * status check (RPCs probed from systems such as K8s)
  * logging (for changing log level at runtime remotely)
* [etcd + Kafka flavor](etcdkafka) - adds etcd & Kafka client plugin instances to 
  the [RPC flavor](rpc)
  
The following diagram shows:
* plugins that are part of the flavor
* initialized (injected) [statuscheck](../health/statuscheck) dependency 
  inside [etcd client plugin](../db/keyval/etcdv3) and [Kafka client plugin](../messaging/kafka)
* [etcd + Kafka flavor](etcdkafka) extends [RPC flavor](rpc) 

![flavors](../docs/imgs/flavors.png)
