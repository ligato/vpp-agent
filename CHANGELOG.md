# Release v1.1 (2018-01-22)

## Compatibility
VPP version v18.04-rc0~33-gb59bd65
cn-infra v1.0.8

# New Features
- [ifplugin](plugins/defaultplugins/ifplugin)
    - added support for un-numbered interfaces. Interface can be marked as un-numbered with information 
    about another interface containing required IP address. Un-numbered interface does not need to have 
    IP address set.
    - added support for virtio-based TAPv2 interfaces.
    - interface status is no longer stored in the ETCD by default and it can be turned on using appropriate
    setting in defaultplugins.conf. See  [readme](plugins/defaultplugins/README.md) for more details.  
- [l2plugin](plugins/defaultplugins/l2plugin)      
    - bridge domain status is no longer stored in the ETCD by default and it can be turned on using appropriate
    setting in defaultplugins.conf. See  [readme](plugins/defaultplugins/README.md) for more details.  

# Improvements
- [ifplugin](plugins/defaultplugins/ifplugin)
    - default MTU value was removed in order to be able to just pass empty MTU field. MTU now can be 
    set only in interface configuration (preffered) or defined in defaultplugins.conf. If none of them
    is set, MTU value will be empty.
    - interface state data are stored in statuscheck readiness probe
- [l3plugin](plugins/defaultplugins/l3plugin)
    - removed strict configuration order for VPP ARP entries and routes. Both ARP entry or route can 
    be configured without interface already present.
- [l4plugin](plugins/defaultplugins/l4plugin)    
   - removed strict configuration order for application namespaces. Application namespace can 
    be configured without interface already present.
    
#Localclient
- added API for ARP entries, L4 features, Application namespaces and STN rules.

# Logging
- consolidated and improved logging in defaultplugins and linuxplugins.    

# Bugfix
- fixed skip-resync parameter if defaultplugins.conf is not provided.
- corrected af_packet type interface behavior if veth interface is created/removed.
- several fixes related to the af_packet and veth interface type configuration.
- microservice and veth-interface related events are synchronized.

# Known Issues
- VPP can occasionally cause deadlock during checksum calculation (https://jira.fd.io/browse/VPP-1134)
- VPP-Agent might not properly handle initialization across plugins (this is not occuring currently, but needs to be tested more)

# Release v1.0.8 (2017-11-21)

## Compatibility
VPP version v18.01-rc0-309-g70bfcaf

## Major Themes

- [cn-infra]
    - updated to version 1.0.7
    
# New Features
- [ifplugin](plugins/defaultplugins/ifplugin)
   - ability to configure STN rules.  See respective
   [readme](plugins/defaultplugins/ifplugin/README.md) in interface plugin for more details.
   - rx-mode settings can be set on interface. Ethernet-type interface can be set to POLLING mode, 
   other types of interfaces supports also INTERRUPT and ADAPTIVE. Fields to set QueueID/QueueIDValid
   are also available
- [l4plugin](plugins/defaultplugins/l4plugin)
   - added new l4 plugin to the VPP plugins. It can be used to enable/disable L4 features 
   and configure application namespaces. See respective
    [readme](plugins/defaultplugins/l4plugin/README.md) in L4 plugin for more details.
   - support for VPP plugins/l3plugin ARP configuration. The configurator can perform the
   basic CRUD operation with ARP config.
   
# Defaultplugins
- [ifplugin](plugins/defaultplugins/ifplugin)
  - added possibility to add interface to any VRF table.
- [resync](plugins/defaultplugins/data_resync.go)
  - resync error propagation improved. If any resynced configuration fails, rest of the resync
  completes and will not be interrupted. All errors which appears during resync are logged after. 
- added defaultpligins API.
- API contains new Method `DisableResync(keyPrefix ...string)`. One or more ETCD key prefixes 
  can be used as a parameter to disable resync for that specific key(s).
     
# Linuxplugin
- [l3plugin](plugins/linuxplugin/l3plugin)
  - route configuration do not return error if required interface is missing. Instead, the 
  route data are internally stored and configured when the interface appears.      
     
## GOVPP
- delay flag removed from GoVPP plugin    

## Documentation
- improved in multiple vpp-agent packages

## Minor fixes/improvements
- removed deadlinks from README files

# Release v1.0.7 (2017-10-30)

## Compatibility
VPP version v18.01-rc0~154-gfc1c612

## Major Themes

- [cn-infra]
    - updated to version 1.0.6
    
- [Default VPP plugin](plugins/defaultplugins)
    - added resync strategies. Resync of VPP plugins (defaultplugins) can be set using 
    defaultpluigns config file; Resync can be set to full (always resync everything) or
    dependent on VPP configuration (if there is none, skip resync). Resync can be also 
    forced to skip using parameter. See appropriate changelog in 
    [Defaultplugins](plugins/defaultplugins) for details.
    
# New Features

- [Linuxplugins L3Plugin](plugins/linuxplugin/l3plugin)
    - added support for basic CRUD operations with static Address resolution protocol 
    entries and static Routes. 
        
# Release v1.0.6 (2017-10-17)

## Major Themes

- [cn-infra]
    - updated to version 1.0.5
- [LinuxPlugin](plugins/linuxplugin)
   - Configuration of vEth interfaces modified. Veth configuration defines
   two names: symbolic used internally and the one used in host OS.
   `HostIfName` field is optional. If it is not defined, the name in the host OS
   will be the same as the symbolic one - defined by `Name` field.

# Release v1.0.5 (2017-09-26)

## Compatibility
VPP version v17.10-rc0~334-gce41a5c

## Major Themes
- [cn-infra]
    - updated to version 1.0.4
- [GoVppMux](plugins/govppmux)
    - configuration file for govpp added
- [Kafka Partitions](vendor/github.com/ligato/cn-infra/messaging/kafka)
    - Changes in offset handling, only automatically partitioned messages (hash, random)
      have their offset marked. Manually partitioned messages are not marked.
    - Implemented post-init consumer (for manual partitioner only) which allows to start
      consuming after kafka-plugin Init()
    - Minimalistic examples & documentation for Kafka API will be improved in a later release.

# Release v1.0.4 (2017-09-08)

## Major Themes

- [Kafka Partitions](vendor/github.com/ligato/cn-infra/messaging/kafka)
    - Implemented new methods that allow to specificy partitions & offset parameters:
      * publish: Mux.NewSyncPublisherToPartition() & Mux.NewAsyncPublisherToPartition()
      * watch: ProtoWatcher.WatchPartition()
    - Minimalistic examples & documentation for Kafka API will be improved in a later release.
- [Flavors](flavors)
    - reduced to only [local.FlavorVppLocal](flavors/local/local_flavor.go) & [vpp.Flavor](flavors/vpp/vpp_flavor.go)
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
[Bridge domains](plugins/defaultplugins/l2plugin/bdidx) and [VETHs](plugins/linuxplugin/ifplugin/ifaceidx)
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
