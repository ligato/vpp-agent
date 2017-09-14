# Release v1.0.5 (PLANNED)

## Cassandra
* connectivity status check

# Release v1.0.4 (NOT RELEASED)

## Profiling
* Logging duration (etcd connection establishment, kafka connection establishment, resync)

## Plugin Configuration
* new [examples/plugin_config](examples/plugin_config)
* new flag --config-dir=. (by default "." meaning current working directory)
* configuration files can but not need to have absolute paths anymore (e.g. --kafka-config=kafka.conf)
* if you put all configuration files (etcd.conf, kafka.conf etc.) in one directory agent will load them
* if you want to disable configuration file just put empty value for a particular flag (e.g. --kafka-config)

## Logging
* [logmanager plugin](logging/manager)
  * new optional flag --logs-config=logs.conf (showcase in [examples/logs_plugin](examples/logs_plugin))
  * this plugin is now part of LocalFlavor (see field Logs) & tries to load configuration
  * HTTP dependency is optional (if it is not set it just does not registers HTTP handlers)
* logger name added in logger fields (possible to grep only certain logger - effectively by plugin)

## Kafka
* kafka.Plugin.Disabled() returned if there is no kafka.conf present
* fixes inside Mux.NewSyncPublisher() & Mux.NewAsyncPublisher() related to previous partition changes
* Known Issues:
  * More than one network connection to Kafka (multiple instances of MUX)
  * TODO Minimalistic examples & documentation for Kafka API will be improved in a later release.

# Release v1.0.3 (2017-09-08)
* [FlavorAllConnectors](flavors/connectors)
    * Inlined plugins: ETCD, Kafka, Redis, Cassandra 
* [Kafka Partitions](messaging/kafka) 
    * Implemented new methods that allow to specify partitions & offset parameters:
      * publish: Mux.NewSyncPublisherToPartition() & Mux.NewAsyncPublisherToPartition()
      * watch: ProtoWatcher.WatchPartition()
    * Minimalistic examples & documentation for Kafka API will be improved in a later release. 

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
