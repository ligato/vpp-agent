# Release v1.0.5 (2017-10-17)

## Profiling
* new [logging/measure](logging/measure) - time measurement utility to measure duration of binary api calls 
  or linuxplugin netlink during resync. The feature is disabled by default and it can be enabled in 
  defaultplugins.conf and linuxplugin.conf file (see plugin's readme)

## Kafka
* proto_connection.go and bytes_connection.go consolidated, bytes_connection.go now mirrors all 
  the functionality from proto_connection.go. 
  * mux can create two types of connection, standard bytes connection and bytes manual connection.
    This enables to call only respective methods on them (to use manual partitioning, it is needed to 
    create manual connection, etc.)
  * method ConsumeTopicOnPartition renamed to ConsumeTopic (similar naming as in the proto_connection.go).
    The rest of the signature is not changed.     
* post-init watcher enabled in bytes_connection.go api
* added methods MarkOffset and CommitOffsets to both, proto and bytes connection. Automatic offset marking
  was removed
* one instance of mux in kafka plugin 
* new field `group-id` can be added to kafka.conf. This value is used as a Group ID in order to set it 
  manually. In case the value is not provided, the service label is used instead (just like before). 
      
# Release v1.0.4 (2017-9-25)

## Documentation
* Improved documentation of public APIs (comments)
* Improved documentation of examples (comments, doc.go, README.md)
* Underscore in example suffixes "_lib" and "_plugin" replaced with a dash

## Health, status check & probes
* status check is now registered also for Cassandra & Redis
* new prometheus format probe support (in rpcflavor)

## Profiling
* Logging duration (etcd connection establishment, kafka connection establishment, resync)

## Plugin Configuration
* new [examples/configs-plugin](examples/configs-plugin)
* new flag --config-dir=. (by default "." meaning current working directory)
* configuration files can but not need to have absolute paths anymore (e.g. --kafka-config=kafka.conf)
* if you put all configuration files (etcd.conf, kafka.conf etc.) in one directory agent will load them
* if you want to disable configuration file just put empty value for a particular flag (e.g. --kafka-config)

## Logging
* [logmanager plugin](logging/logmanager)
  * new optional flag --logs-config=logs.conf (showcase in [examples/logs-plugin](examples/logs-plugin))
  * this plugin is now part of LocalFlavor (see field Logs) & tries to load configuration
  * HTTP dependency is optional (if it is not set it just does not registers HTTP handlers)
* logger name added in logger fields (possible to grep only certain logger - effectively by plugin)

## Kafka
* kafka.Plugin.Disabled() returned if there is no kafka.conf present
* Connection in bytes_connection.go renamed to BytesConnection
* kafka plugin initializes two multiplexers for dynamic mode (automatic partitions) and manual mode.
  Every multiplexer can create its own connection and provides access to different set of methods 
  (publishing to partition, watching on partition/offset)
* ProtoWatcher from API was changed - methods WatchPartition and StopWatchPartition were removed 
  from the ProtoWatcher interface and added to newly created ProtoPartitionWatcher. There is also a new 
  method under Mux interface - NewPartitionWatcher(subscriber) which returns ProtoPartitionWatcher
  instance that allows to call partition-related methods
* Offset mark is done for hash/default-partitioned messages only. Manually partitioned message's offset 
  is not marked.
* It is possible to start kafka consumer on partition after kafka plugin initialization procedure. New
  example [post-init-consumer](examples/kafka-plugin/post-init-consumer) was created to show the 
  functionality     
* fixes inside Mux.NewSyncPublisher() & Mux.NewAsyncPublisher() related to previous partition changes
* Known Issues:
  * More than one network connection to Kafka (multiple instances of MUX)
  * TODO Minimalistic examples & documentation for Kafka API will be improved in a later release.

## Flavors
* optionally GPRC server can be enabled in [rpc flavor](flavors/rpc) using --grpc-port=9111 (or using config gprc.conf)
* [Flavor interface](core/list_flavor_plugin.go) now contains three methods: Plugins(), Inject(), LogRegistry() to standardize these methods over all flavors. Note, LogRegistry() is usually embedded using local flavor.

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
