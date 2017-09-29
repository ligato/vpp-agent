# Release v1.0.5 (2017-9-26)

## Compatibility
VPP version v17.10-rc0~334-gce41a5c

## Major Themes
- [cn-infra]
    - updated to version 1.0.4
- [GoVppMux](plugins/govppmux)
    - configuration file for govpp added
- [Kafka Partitions](messaging/kafka)
    - Changes in offset handling, only automatically partitioned messages (hash, random)
      have their offset marked. Manually partitioned messages are not marked.
    - Implemented post-init consumer (for manual partitioner only) which allows to start
      consuming after kafka-plugin Init()   
    - Minimalistic examples & documentation for Kafka API will be improved in a later release.
    
# Release v1.0.4 (2017-09-08)

## Major Themes

- [Kafka Partitions](messaging/kafka)
    - Implemented new methods that allow to specificy partitions & offset parameters:
      * publish: Mux.NewSyncPublisherToPartition() & Mux.NewAsyncPublisherToPartition()
      * watch: ProtoWatcher.WatchPartition()
    - Minimalistic examples & documentation for Kafka API will be improved in a later release.
- [Flavors](flavors)
    - reduced to only [local.FlavorVppLocal](flavors/linuxlocal/local_flavor.go) & [vpp.Flavor](flavors/vpp/vpp_flavor.go)
- [goVpp]
    - updated version waits until vpp is ready to accept a new connection

# Release v1.0.3 (2017-09-05)

## Major Themes

Enabled support for wathing data store `OfDifferentAgent()` - see:
* [examples/idx_iface_cache](examples/idx_iface_cache/main.go)
* [examples/examples/idx_bd_cache](examples/idx_bd_cache/main.go)
* [examples/idx_veth_cache](examples/idx_veth_cache/main.go)

Preview of new Kafka client API methods that allows to fill also partition and offset argument. New methods implementation ignores these new parameters for now (fallbacking to existing implementation based on `github.com/bsm/sarama-cluster` and `github.com/Shopify/sarama`).

## Compatibility
VPP version v17.10-rc0~265-g809bc74 (upgraded because of VPP MEMIF fixes).


# Release v1.0.2 (2017-08-28)

## Major Themes

Algorithms for applying northbound configuration (stored in ETCD key value data store)
to VPP in the proper order of VPP binary API calls implemented in [Default VPP plugin](plugins/defaultplugins):
- network interfaces, especially:
  - MEMIFs (optimized dataplane network interface tailored for a container to container network connectivity)
  - VETHs (standard Linux Virtual Ethernet network interface)
  - AF_Packets (for accessing VETHs and similar type of interface)
  - VXLANs, Physical Network Interfaces, loopbacks ...
- L2 BD & X-Connects
- L3 IP Routes & VRFs
- ACL (Access Control List)

Support for Linux VETH northbound configuration implemented in [Linux Plugin](plugins/linuxplugin)
applied in proper order with VPP AF_Packet configuration.

Data Synchronization during startup for network interfaces & L2 BD
(support for situation when ETCD contain configuration before VPP Agent starts).

Data replication and events:
- Updating operational data in ETCD (VPP indexes such as  sw_if_index) and statistics (port counters).
- Updating statistics in Redis (optional once redis.conf available - see flags).
- Publishing link up/down events to Kafka message bus.

Miscellaneous:
- [Examples](examples)
- Tools:
  - [agentctl CLI tool](cmd/agentctl) that show state & configuration of VPP agents
  - [docker](docker): container-based development environment for the VPP agent
- other features inherited from cn-infra:
  - [health](https://github.com/ligato/cn-infra/tree/master/health): status check & k8s HTTP/REST probes
  - [logging](https://github.com/ligato/cn-infra/tree/master/logging): changing log level at runtime

### Extensibility & integration
Ability to extend the behavior of the VPP Agent by creating new plugins on top of [VPP Agent flavor](flavors/vpp).
New plugins can access API for configured:
[VPP Network interfaces](plugins/defaultplugins/ifplugin/ifaceidx),
[Bridge domains](plugins/defaultplugins/l2plugin/bdidx) and [VETHs](plugins/linuxplugin/ifaceidx)
based on [idxvpp](idxvpp) threadsafe map tailored for VPP data
with advanced features (multiple watchers, secondary indexes).

VPP Agent is embeddable in different software projects and with different systems
by using [Local Flavor](flavors/local) to reuse VPP Agent algorithms.
For doing this there is [VPP Agent client version 1](clientv1):
* local client - for embedded VPP Agent (communication inside one operating system process, VPP Agent effectively used as a library)
* remote client - for remote configuration of VPP Agent (while integrating for example with control plane)

## Known Issues
A rarely occurring problem during startup with binary API connectivity.
VPP rejects binary API connectivity when VPP Agent tries to connect
too early (plan fix this behavior in next release).

## Compatibility
VPP version v17.10-rc0~203
