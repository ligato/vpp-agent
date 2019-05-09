# Release v2.1 (2019-05-09)

## Improvements
* [Datasync][datasync-plugin]
  - Added new option `WithClientLifetimeTTL`. This option defines `put` operation for the lifetime of a client (the TTL is not renewed if the client is closed).
* [ETCD][etcd-plugin]
  - Support for election mechanism, which determines leader instance for given prefix. Multiple ETCD instances can now compete for leadership. 
* [Bolt][bolt-plugin]
  - Some improvements to Bolt watchers were done in order to reduce a number of active go routines
  
## Fixed Bugs
* BoltDB watcher: keys sent to the close channel are now properly prefixed and delivered to the correct go routine so they will not be ignored     

## Documentation
* Majority of `README.md` files were removed. The content was updated, extended and moved to [Ligato documentation site][ligato-docs].

# Release v2.0 (2019-04-02)

## Breaking Changes
* The `ChangeEvent` interface was modified. Before, it returned single value of type `ProtoWatchResp`, now it contains a new method `GetChanges()` which return a list of values of that type. 

## New Features
* [Datasync][datasync-plugin]
  - The `ChangeEvent` interface provides a new method `GetContext` returning a context associated with the given event. 
  - The `ResyncEvent` interface also provides a new method called `GetContext` returning a context associated with the given resync event.
  - Resync time duration is shown in milliseconds
* [IdxMap][index-map]
  - New method `ListFields(string)` in the `NamedMapping` interface providing a map of fields associated with the item identified by the named parameter known as secondary indexes     
* [GRPC][grpc-plugin]
  - Added authentication support for the GRPC plugin
* [Probe][probe-plugin]
  - Support for non-fatal errors. The probe plugin keeps a lst of non-fatal plugins. Errors reported from the non-fatal plugin are effectively ignored in the vpp-agent overall status.
* [BoltDB][bolt-plugin]
  - The config file now has an option to filter duplicated notifications.  

## Improvements
* [Agent][agent-core]
  - `DumpStackTraceOnTimeout` is now disabled by default with possibility to enable it via the environment variable `DUMP_STACK_ON_TIMEOUT`.
  - The agent logs the last plugin in case it fails to start because of the timeout
  - Signal received during startup closes the agent instance
* [FileDB][filedb-plugin]
  - If the child process is not detached, the `PDeathSignal` is set preventing the child process to hang
  - The process watcher is started on process start, rather than on process creation
  - A new or an attached process can be created with custom I/O writer
  - Support for environment variables for processes  
* [Resync][resync-plugin]
  - The resync timeout was split into two values, for ACK timeout set to 10 second (up from 5) and for ACCEPT timeout set to 1 second.   
  
## Fixed Bugs
* The ETCD plugin watch is now properly closed using a context
* The plugin lookup works correctly with interface types where the inner type is slice or array of plugins
* `PropagateResync` now waits at the result of the resync event instead of returning immediately
* Used `jsonpb` to fix protobuf marshalling/unmarshalling
* Plugin config directory paths with ".." are now handled properly.
* If channel used by process manager is closed by the used (which should not be done), the plugin recovers

## Documentation
* Added "beginner" tutorials (hello world, plugin dependencies, REST handler, KV Store, plugin lookup) with examples. 

# Release v1.7 (2018-12-12)

## Major Topics

**FileDB**

  With the new release, the cn-infra introduces a new feature where the filesystem itself can be used as a key-value data store. Defined files or directories can be used to store configuration items the same way as any other database (ETCD, Redis, etc.). The data have to have prescribed format but otherwise, they respect proto models from particular plugins. Configuration files can be of JSON or YAML type. The fileDB plugin supports watcher, configuration status update or resync - features known from other key-value data stores.
  
**Process management plugin**

  A new plugin was added to extend cn-infra functionality - a process manager. The plugin can be used to create, start and monitor processes. Begin with creating new process instance like `NewProcess(<cmd>, <options>...)` and start it with `Start()`. 
  The process manager supports all the common commands like restart, termination, kill, wait for process completion, check process liveness or watch its status. More advanced features or options allow to detach the process from a parent, preserve it when an application is restarted or automatic cleanup of zombie processes.
  
## New features
  * Added new plugin allowing usage of a filesystem as the key-value data store. More details in the [readme][cn-infra-readme]
  * Added new plugin for external process management. More information in the [readme][cn-infra-readme]
  * StatusCheck plugin has a new option to define non-fatal plugins. A registered plugin marked as non-fatal is not propagated into overall status.
  * Added watcher for BoltDB database. The BoltDB now supports all the watcher options as other key-value data store plugins.
  
## Examples
  * Redis-lib simple example was updated and changed, now clearly demonstrating Redis plugin options and usage. The Redis "airport" was rewritten, the functionality remained unchanged, but it should be a lot easier to understand the code.
  
## Other
  * Agent start/stop error now prints stack trace  


# Release v1.6 (2018-10-04)

## Major topics

**Trace**

  We have introduced a new component which replaces stopwatch called tracer. The core functionality
  remained the same, it still allows to measure time duration between two parts of the code
  and store it internally. The motive was that the original implementation was somehow cumbersome,
  and the name did not reflect the main purpose.
  New tracer object is created via constructor `NewTracer(<name>, <logger>)`. It has very simple 
  API with method `LogTime(<name>, <start_time>)` which adds entry to the internal database 
  and `Get()`, which returns a proto-modelled list of entries. Database can be purged with `Clear()`.
  
**REST Security**

  Basic security support was added to RPC plugin. The caller can use pre-defined credentials to obtain
  authentication tokens required for access. Permission groups are also supported. To create
  a new permission group, use method from REST API `RegisterPermissionGroup(<groups>)`. Group
  is assigned to the user in `http.conf` file (where all users are defined).
  This feature is available only to REST and it is considered experimental in the current release, 
  and will be extended in the future.
  
## New Features
- [measure][tracer-plugin]
  * New component [tracer][tracer-plugin] was introduced. It serves the similar purpose
  as the stopwatch. More details in the [readme][cn-infra-readme]
  * Stopwatch was removed
- [statuscheck][statuscheck-plugin]
  * The liveness probe now shows also a state of all registered plugins (not only the overall state)
- [rest][rest-plugin]
  * New security functionality for REST plugin was added. To learn more about it, see the
  [readme][cn-infra-readme].  
- [logging][log-registry]
  * Logger API has two new methods, `SetOutput(<io.Writer>)` to set custom logging output 
  and `SetFormatter(<formatter>)` to set custom formatter before logged to output.   
  
## Other
  * Every proto file is generated with the gogo/proto package.  
  
## Bugfix
  * GRPC "listen and serve" call was moved to after init, which prevents calling services
  which were not yet registered. Also fixed a bug causing occasional panic if GRPC plugin 
  was disabled.
  * ETCD reconnect resync fixed

# Release v1.5 (2018-08-24)

## Major topics

**Flavors redesign**
  
  This version finally introduces new concept of flavors, system which handles plugins
  and dependencies between them. We believe that the new flavors are much more convenient
  and easier to understand and use. 
  New system widely uses principle of options. Application `NewAgent(options...)` takes
  a list of options as a parameter. The most important option is `AllPlugins(plugins...)`
  which allows to set a list of plugins to the agent. To add single plugin, use `Plugins...`.
  Another options can be used to set timeouts, exit signal in form of go channel or version.
  
  Every plugin defines default plugin instance and other useful methods to, for example, set
  custom plugin dependency. Plugins have their `options.go` file with implementation.
  
  Examples go hand in hand with new flavors, so all of them were updated. The change affects
  only example plugin initialization, the main purpose was left unchanged.
  
**Cryptodata plugin**

  New cryptodata plugin was added, providing support for encrypting and
  decrypting arbitrary data using configured private keys and support for wrapping 
  key/value bytes/proto broker to automatically decrypt any read value matching specified 
  pattern.

**BoldDB store**

  Our key/value databases have a new member, BoltDB. Plugin uses own configuration
  file to define path to BoltDB file, permission, etc. Bolt supports all standard
  cn-infra features, including resync.
  
## New Features
- [agent][agent-core]
  * Agent uses new concept of flavors. Since all the examples moved to this new plugin
  management, package flavors was removed. Learn more about new flavors 
  [here][agent-core]
- [cryptodata][cryptodata-plugin]
  * New plugin cryptodata contains implementation files for encryption/decryption
  support. In order to try the functionality, examples for [library][examples-cryptodata-lib] ,
  [plugin][examples-cryptodata] and [proto-plugin][examples-cryptodata-proto] 
  were added.
- [boltDB][bolt-plugin]
  * Added support for the BoltDB keyval database. There is also a new example 
  for [BoltDB plugin][examples-bolt].
- [logging][logrus] 
  * Added support for external hooks   

# Release v1.4.1 (2018-07-23)

## Bugfix
  * Fixed issue in Consul client that caused brokers to incorrectly
  trim prefixes and thus storing invalid revisions for resync.

# Release v1.4 (2018-07-16)

## Breaking Changes
  * Package **etcdv3** was renamed to **etcd**. This change affects imports. Also 
  pay attention to the configuration flag which can also be influenced by the change.
  Based on the change, flag for ETCD configuration file `--etcdv3-config` is now defined 
  as `--etcd-config`.

## New Features
  * Support for GRPC unix domain socket type. Socket types tcp, tcp4, tcp6, unix 
  and unixpacket can be used with GRPC. Desired socket type and address/file can be 
  specified via grpc configuration file ([example here][grpc-conf]). More
  information [here][grpc-plugin]
  * Rest plugin security improved. Security features are the usage of client/server certificates
  (HTTPS) and basic HTTP authentication with username and password. Those features 
  are disabled by default. Information about how to use it and example can be found [here][rest-plugin]
  
## Other
  * Example configuration files with description were added to every plugin which 
  supports/uses them. 
  
## Bugfix
  * Fixed occasional failure of method deriving config files
  * Fixed multiple issues in logs-lib example (logger, HTTP usage)  
  * To prevent incorrect values in subsequent changes, previous value of key should 
  be correctly cleaned up if the resync was called outside of initialization phase.
  * Fixed the logger configuration file. All created loggers are correctly set with a log
  level according to the map in file. The default log level can be also set, but keep in mind
  that the environmental variable `INITIAL_LOGLVL` replaces the value from the config.

# Release v1.3 (2018-05-24)

## New Features
  * Automatic resync if ETCD was disconnected and connected again. The feature is
  disabled by default. See [plugin directory][etcd-plugin] to learn how to 
  enable the feature.
  * Watch registration [API][datasync-api] now contains new method 
  __Registration()__ allowing to register new key to all adapters.
  * New plugin for Consul. See [plugin directory][consul-plugin]
  for more information.
  * In-memory mapping method __UpdateMetadata()__ now triggers events. Use __IsUpdate()__
  so see if the event comes from the update notification.
  
## Bugfix
  * Transport for statuscheck plugin fixed
  * Fixed bug where watcher was closed after server restart if database was compacted  

# Release v1.2 (2018-03-22)

## New Features
  * Added support for ETCD compacting. Information about how to use 
    it can be found in the [readme][etcd-plugin]
  * Name-to-index mapping API was extended with new method 'Update'.
    The purpose of the method is to update metadata value under specific
    key without triggering events, so mapping entry can be kept up to date. 

## Bugfix
  * Fixed syncbase issue where delete request for a non-existing item
    used to trigger a change notification
  * Getting of previous value for ProtoWatchResp 'delete' event now
    returns correct data  

# Release v1.1 (2018-02-07)

## Dependencies
  * Migrated from glide to dep

## Prometheus
  * Introduced Prometheus plugin with examples

# Release v1.0.8 (2018-01-22)

## Kafka
  * Added support for Kafka TLS.

## Logging
  * Logger config file now enables to set every logger to desired level or use default level for all loggers
    within plugins. For this purpose it is also possible to use environment variable INITITAL_LOGLVL.

## Statuscheck
  * Readiness probe now allows to report interfaces' state to the proble output. 

## Dependencies
  * Sirupsen package is now lower-cased according to recommandations.   

# Release v1.0.7 (2017-11-14)

## Agent, Flavors
Input arguments of `core.NewAgent()` were changed:
  * it is possible to call NewAgent without options: `core.NewAgent(flavor)`
  * you can pass options like this: `core.NewAgent(flavor, core.WithTimeout(1* time.Second))`
  * there is `core.NewAgentDeprecated()` for backward compatibility

This release contains utilities/options to avoid writing the new flavor go structures 
(Inject, Plugins methods) for simple customizations:
* if you just expose the RPCs you can write
  ```
  rpc.NewAgent(rpc.WithPlugins(func(flavor *rpc.FlavorRPC) []*core.NamedPlugin {
    return []*core.NamedPlugin{{"myplugin1", &MyPlugin{&flavor.GRPC},
                               {"myplugin2", &MyPlugin{&flavor.GRPC},}
  }))
  ```
* if you want to use one simple plugin (without any client or a server) you can write:
  ```
  flavor := &local.FlavorLocal{}
  core.NewAgent(core.Inject(flavor), core.WithPlugin("myplugin1", &MyPlugin{Deps: flavor.PluginInfraDeps("myplugin1")}))
  ```
* if you want to combine multiple flavors to inject their plugins to new MyPlugin
  ```
  loc := &local.FlavorLocal{}
  rpcs := &rpc.FlavorRPC{FlavorLocal: loc}
  cons := &connectors.AllConnectorsFlavor{FlavorLocal: loc}
  core.NewAgent(core.Inject(rpcs, cons), core.WithPlugin("myplugin", &MyPlugin{Deps: Deps{&rpcs.GRPC, &cons.ETCD}}))
  ```

## ETCD/Datasync
* GetPrevValue enabled in the proto watcher API. Etcd-lib/watcher example was updated to show 
  the added functionality.
* Fixed datasync internal state after resync causing that the resynced data were handled as created
  after first modification.
* Fixed issue where datasync plugin was stuck on close

## Cassandra
Added TLS support

# Release v1.0.6 (2017-10-30)

## ETCD/Datasync
* etcd new feature PutIfNotExists adds key-value pair if the key doesn't exist.
* feature GetPrevValue() used to obtain previous value from key-value database was returned to API
* watcher registration object has a new method to close single subscribed key. Key can be un-subscribed in runtime.
  See example usage in [examples/datasync-plugin][examples-datasync] for more details  

## Documentation
* improved documentation/code comments in datasync, config and core packages 

# Release v1.0.5 (2017-10-17)

## Profiling
* new [logging/measure][tracer-plugin] - time measurement utility to measure duration of function or 
  any part of the code. Use `NewStopwatch(name string, log logging.Logger)` to create an 
  instance of stopwatch with name, desired logger and table with measured time entries. See [plugin folder][tracer-plugin] for more information.

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

# Release v1.0.4 (2017-09-25)

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
* new [examples/configs-plugin][examples-config]
* new flag --config-dir=. (by default "." meaning current working directory)
* configuration files can but not need to have absolute paths anymore (e.g. --kafka-config=kafka.conf)
* if you put all configuration files (etcd.conf, kafka.conf etc.) in one directory agent will load them
* if you want to disable configuration file just put empty value for a particular flag (e.g. --kafka-config)

## Logging
* [logmanager plugin][logmanager-plugin]
  * new optional flag --logs-config=logs.conf (showcase in [examples/logs-plugin][examples-logs])
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
  example [post-init-consumer][examples-kafka-post-init] was created to show the
  functionality
* fixes inside Mux.NewSyncPublisher() & Mux.NewAsyncPublisher() related to previous partition changes
* Known Issues:
  * More than one network connection to Kafka (multiple instances of MUX)
  * TODO Minimalistic examples & documentation for Kafka API will be improved in a later release.

## Flavors
* optionally GPRC server can be enabled in rpc flavor (removed in v1.5) using --grpc-port=9111 (or using config gprc.conf)
* Flavor interface (removed in v1.5) now contains three methods: Plugins(), Inject(), LogRegistry() to standardize these methods over all flavors. Note, LogRegistry() is usually embedded using local flavor.

# Release v1.0.3 (2017-09-08)
* FlavorAllConnectors (removed in v1.5)
    * Inlined plugins: ETCD, Kafka, Redis, Cassandra
* [Kafka Partitions][kafka-plugin]
    * Implemented new methods that allow to specify partitions & offset parameters:
      * publish: Mux.NewSyncPublisherToPartition() & Mux.NewAsyncPublisherToPartition()
      * watch: ProtoWatcher.WatchPartition()
    * Minimalistic examples & documentation for Kafka API will be improved in a later release.

# Release v1.0.2 (2017-08-28)

## Major Themes

The major themes for Release v1.0.2 are as follows:
* Libraries (GO Lang packages) for connecting to Data Bases and Message Bus.
  Set of these libraries provides unified client API and configuration for:
    * [Cassandra][cassandra-plugin]
    * [etcd][etcd-plugin]
    * [Redis][redis-plugin]
    * [Kafka][kafka-plugin]
* [Data Synchronization][datasync-plugin] plugin for watching and writing data asynchronously; it is currently implemented only for the [db/keyval API][keyval-api] API. It facilitates reading of data during startup or after reconnection to a data store and then watching incremental changes.
* Agent [Core][agent-core] that provides plugin lifecycle management
(initialization and graceful shutdown of plugins) is able to run
different flavors (removed in v1.5) (reusable collection of plugins):
    * local flavor - a minimal collection of plugins:
      * [statuscheck][statuscheck-plugin]
      * [servicelabel][service-label-plugin]
      * [resync orch]([datasync-restsync]
      * [log registry][log-registry]
    * RPC flavor - exposes REST API for all plugins, especially for:
      * [statuscheck][statuscheck-plugin] (RPCs probed from systems such as K8s)
      * [logging][logmanager-plugin] (for changing log level at runtime remotely)
    * connector flavors:
      * Cassandra flavor
      * etcd flavor
      * Redis flavor
      * Kafka flavor
* [Examples][examples]:
* [Docker][docker] container-based development environment
* Helpers:
  * [IDX Map][index-map] is a reusable thread-safe in memory data structure.
      This map is designed for sharing key value based data
      (lookup by primary & secondary indexes, plus watching individual changes).
      It is useful for:
      - implementing backend for plugin specific API shared by multiple plugins;
      - caching of a remote data store.
  * [Config]([config]: helpers for loading plugin specific configuration.
  * [Service Label][service-label-plugin]: retrieval of a unique identifier for a CN-Infra based app.
  
[agent-core]: agent 
[bolt-plugin]: db/keyval/bolt
[cassandra-plugin]: db/sql/cassandra
[cn-infra-readme]: README.md
[config]: config
[consul-plugin]: db/keyval/consul
[cryptodata-plugin]: db/cryptodata
[datasync-api]: datasync/datasync_api.go
[datasync-plugin]: datasync
[datasync-restsync]: datasync/restsync
[docker]: docker
[etcd-plugin]: db/keyval/etcd
[examples]: examples
[examples-bolt]: examples/bolt-plugin
[examples-config]: examples/configs-plugin
[examples-cryptodata]: examples/cryptodata-plugin
[examples-cryptodata-lib]: examples/cryptodata-lib
[examples-cryptodata-proto]: examples/cryptodata-proto-plugin
[examples-datasync]: examples/datasync-plugin
[examples-kafka-post-init]: examples/kafka-plugin/post-init-consumer
[examples-logs]: examples/logs-plugin
[filedb-plugin]: db/keyval/filedb
[grpc-conf]: rpc/grpc/grpc.conf
[grpc-plugin]: rpc/grpc
[index-map]: idxmap
[kafka-plugin]: messaging/kafka
[keyval-api]: db/keyval
[ligato-docs]: https://docs.ligato.io/en/latest/
[log-registry]: logging
[logmanager-plugin]: logging/logmanager
[logrus]: logging/logrus
[probe-plugin]: health/probe
[redis-plugin]: db/keyval/redis
[rest-plugin]: rpc/rest
[resync-plugin]: datasync/resync
[service-label-plugin]: servicelabel
[statuscheck-plugin]: health/statuscheck
[tracer-plugin]: logging/measure
