# Release v1.0.3 (NOT RELEASED)
- [FlavorAllConnectors](flavors/connectors) - Inlined plugins: ETCD, Kafka, Redis, Cassandra 

# Release v1.0.2 (2017-08-28)

## Major Themes

The major themes for Release v1.0.2 are as follows:
* Libraries (GO Lang packages) for connecting to Data Bases and Message Bus.
  Set of these libraries provides unified client API and configuration for:
    * [Cassandra](db/sql/cassandra)
    * [etcd](db/keyval/etcdv3)
    * [Redis](db/keyval/redis)
    * [Kafka](db/)
* [Data Synchronization](datasync) plugin for watching and writing data asynchronously; it is currently implemented only for the [db/keyval API](db/keyval) API. It facilitates reading of data during startup or after reconnection to a data store and then watching incremental changes.
* Agent [Core](core) that provides plugin lifecycle management 
(initialization and graceful shutdown of plugins) is able to run
different [flavors](flavors) (reusable collection of plugins):
    * [local flavor](flavors/local) - a minimal collection of plugins: 
      * [statuscheck](health/statuscheck) 
      * [servicelabel](servicelabel) 
      * [resync orch](datasync/restsync) 
      * [log registry](logging)
    * [RPC flavor](flavors/rpc) - exposes REST API for all plugins, especially for:
      * [statuscheck](health/statuscheck) (RPCs probed from systems such as K8s)
      * [logging](logging/logmanager) (for changing log level at runtime remotely)
    * connector flavors: 
      * Cassandra flavor
      * etcdv3 flavor
      * Redis flavor
      * Kafka flavor
* [Examples](examples)
* [Docker](docker) container-based development environment 
* Helpers:
  * [IDX Map](idxmap) is a reusable thread-safe in memory data structure.
      This map is designed for sharing key value based data
      (lookup by primary & secondary indexes, plus watching individual changes).
      It is useful for:
      - implementing backend for plugin specific API shared by multiple plugins;
      - caching of a remote data store.
  * [Config](config): helpers for loading plugin specific configuration.
  * [Service Label](servicelabel): retrieval of a unique identifier for a CN-Infra based app.
