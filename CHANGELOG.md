# Release v1.0.1 (2017-08-25)

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
- Updating statistics in Redis (optional once redis.conf available - see flag in FlavorRedis).
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
by using [Linux Local Flavor](flavors/linuxlocal) to reuse VPP Agent algorithms.
For doing this there is [VPP Agent client version 1](clientv1):
* local client - for embedded VPP Agent (communication inside one operating system process, VPP Agent effectively used as a library)
* remote client - for remote configuration of VPP Agent (while integrating for example with control plane)

## Known Issues
A rarely occurring problem during startup with binary API connectivity.
VPP rejects binary API connectivity when VPP Agent tries to connect 
too early (plan fix this behavior in next release).

## Compatibility
VPP version v17.10-rc0~203
