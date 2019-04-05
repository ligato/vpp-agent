# Release Changelog

- [v2.0.0](#v2.0.0)
  - [v2.0.1](#v2.0.1)
- [v1.8.0](#v1.8.0)
  - [v1.8.1](#v1.8.1)
- [v1.7.0](#v1.7.0)
- [v1.6.0](#v1.6.0)
- [v1.5.0](#v1.5.0)
  - [v1.5.1](#v1.5.1)
  - [v1.5.2](#v1.5.2)
- [v1.4.0](#v1.4.0)
  - [v1.4.1](#v1.4.1)
- [v1.3.0](#v1.3.0)
- [v1.2.0](#v1.2.0)
- [v1.1.0](#v1.1.0)
- [v1.0.8](#v1.0.8)
- [v1.0.7](#v1.0.7)
- [v1.0.6](#v1.0.6)
- [v1.0.5](#v1.0.5)
- [v1.0.4](#v1.0.4)
- [v1.0.3](#v1.0.3)
- [v1.0.2](#v1.0.2)

<!---
RELEASE CHANGELOG TEMPLATE:
<a name="vX.Y.Z"></a>
# [X.Y.Z](https://github.com/ligato/vpp-agent/compare/vX-1.Y-1.Z-1...vX.Y.Z) (YYYY-MM-DD)
### Compatibility
### BREAKING CHANGES
### Bug Fixes
### Known Issues
### Features
### Improvements
### Docker Images
### Documentation
-->

<a name="v2.0.1"></a>
# [2.0.1](https://github.com/ligato/vpp-agent/compare/v2.0.0...v2.0.1) (2019-04-05)

### Compatibility
- **VPP 19.01** (compatible by default, recommended)
- **VPP 18.10** (backwards compatible)
- cn-infra v2.0
- Go 1.11

### Bug Fixes
* Fixed bug where Linux network namespace was not reverted in some cases.
* The VPP socketclient connection checks (and waits) for the socket file in the same manner as for the shared memory, giving 
  the GoVPPMux more time to connect in case the VPP startup is delayed. Also errors occurred during the shm/socket file watch 
  are now properly handled. 
* Fixed wrong dependency for SRv6 end functions referencing VRF tables (DT6,DT4,T).

### Improvements
* [GoVPPMux](plugins/govppmux)
  - Added option to adjust the number of connection attempts and time delay between them. Seek `retry-connect-count` and
  `retry-connect-timeout` fields in [govpp.conf](plugins/govppmux/govpp.conf). Also keep in mind the total time in which 
  plugins can be initialized when using these fields. 
* [linux-if-plugin](plugins/linux/ifplugin)
  - Default loopback MTU was set to 65536.
* [ns-plugin](plugins/linux/nsplugin)
  - Plugin descriptor returns `ErrEscapedNetNs` if Linux namespace was changed but not reverted back before returned
  to scheduler. 
  
### Docker Images
* Supervisord process is now started with PID=1

<a name="v2.0.0"></a>
# [2.0.0](https://github.com/ligato/vpp-agent/compare/v1.8...v2.0.0) (2019-04-02)

### Compatibility
- **VPP 19.01** (compatible by default, recommended)
- **VPP 18.10** (backwards compatible)
- cn-infra v2.0
- Go 1.11

### BREAKING CHANGES
* All northbound models were re-written and simplified and most of them are no longer compatible with model data from v1.
* The `v1` label from all vpp-agent keys was updated to `v2`.
* Plugins using some kind of dependency on other VPP/Linux plugin (for example required interface) should be updated and handled by the KVScheduler.

### Bug Fixes
* We expect a lot of known and unknown race-condition and plugin dependency related issues to be solved by the KV Scheduler.
* MTU is omitted for the sub-interface type. 
* If linux plugin attempts to switch to non-existing namespace, it prints appropriate log message as warning, and continues with execution instead of interrupt it with error.
* Punt socket path string is cleaned from unwanted characters.
* Added VPE compatibility check for L3 plugin vppcalls.
* The MAC address assigned to an af-packet interface is used from the host only if not provided from the configuration.
* Fixed bug causing the agent to crash in an attempt to 'update' rx-placement with empty value.
* Switch interface from zero to non-zero VRF causes VPP issues - this limitation was now restricted only to unnumbered interfaces.
* IPSec tunnel dump now also retrieves integ/crypto keys.
* Errored operation should no more publish to the index mapping.
* Some obsolete Retval checks were removed.
* Error caused by missing DPDK interface is no longer retryable.
* Linux interface IP address without mask is now handled properly.
* Fixed bug causing agent to crash when some VPP plugin we support was not loaded.
* Fixed metrics retrieval in telemetry plugin.

### Known Issues
* The bidirectional forwarding detection (aka BFD plugin) was removed. We plan to add it in one of the future releases.
* The L4 plugin (application namespaces) was removed.
* We experienced problems with the VPP with some messages while using socket client connection. The issue kind was that the reply message
was not returned (GoVPP could not decode it). If you encounter similar error, please try to setup VPP connection using shared memory (see below). 

### Features
* Performance
  - The vpp-agent now supports connection via socket client (in addition to shared memory). The socket client connection provides higher performance and 
  message throughput, thus it was set as default connection type. The shared memory is still available via the environment variable `GOVPPMUX_NOSOCK`.
  - Many other changes, benchmarking and profiling was done to improve vpp-agent experience.
* Multi-VPP support
  - The VPP-agent can connect to multiple versions of the VPP with the same binary file without any additional building or code changes. See
  compatibility part to know which versions are supported. The list will be extended in the future.
* Models
  - All vpp-agent models were reviewed and cleaned up. Various changes were done, like simple renaming (in order to have more meaningful fields, avoid duplicated names in types, etc.), improved model convenience (interface type-specific fields are now defined as `oneof`, preventing to set multiple or incorrect data) and other. All models were also moved to the common [api](api) folder.
* [KVScheduler](plugins/kvscheduler)
  - Added new component called KVScheduler, as a reaction to various flaws and issues with race conditions between Vpp/Linux plugins, poor readability and poorly readable logging. Also the system of notifications between plugins was unreliable and hard to debug or even understand. Based on this experience, a new framework offers improved generic mechanisms to handle dependencies between configuration items and creates clean and readable transaction-based logging. Since this component significantly changed the way how plugins are defined, we recommend to learn more about it on the [VPP-Agent wiki page](https://github.com/ligato/vpp-agent/wiki/KVScheduler). 
* [orchestrator](plugins/orchestrator)
  - The orchestrator is a new component which long-term added value will be a support for multiple northbound data sources (KVDB, GRPC, ...). The current implementation handles combination of GRPC + KVDB, which includes data changes and resync. In the future, any combination of sources will be supported.
* [GoVPPMux](plugins/govppmux)
  - Added `Ping()` method to the VPE vppcalls usable to test the VPP connection.  
* [if-plugin](plugins/vpp/ifplugin)
  - UDP encapsulation can be configured to an IPSec tunnel interface
  - Support for new Bond-type interfaces.
  - Support for L2 tag rewrite (currently present in the interface plugin because of the inconsistent VPP API)
* [nat-plugin](plugins/vpp/natplugin)      
  - Added support for session affinity in NAT44 static mapping with load balancer.
* [sr-plugin](plugins/vpp/srplugin)
  - Support for Dynamic segment routing proxy with L2 segment routing unaware services.  
  - Added support for SRv6 end function End.DT4 and End.DT6.
* [linux-if-plugin](plugins/linux/ifplugin)
  - Added support for new Linux interface type - loopback.
  - Attempt to assign already existing IP address to the interface does not cause an error.
* [linux-iptables](plugins/linux/iptablesplugin)
  - Added new linux IP tables plugin able to configure IP tables chain in the specified table, manage chain rules and set default chain policy.  

### Improvements
* [KVScheduler](plugins/kvscheduler)
  - Performance improvements related to memory management.
* [GoVPPMux](plugins/govppmux)
  - Need for config file was removed, GoVPP is now set with default values if the startup config is not provided.
  - `DefaultReplyTimeout` is now configured globally, instead of set for every request separately.
  - Tolerated default health check timeout is now set to 250ms (up from 100ms). The old value had not provide enough time in some cases.
* [acl-plugin](plugins/vpp/aclplugin)
  - Model moved to the [api/models](api/models/vpp/acl)
* [if-plugin](plugins/vpp/ifplugin)
  - Model reviewed, updated and moved to the [api/models](api/models/vpp/interfaces/interface.proto).
  - Interface plugin now handles IPSec tunnel interfaces (previously done in IPSec plugin).
  - NAT related configuration was moved to its own plugin.
  - New interface stats (added in 1.8.1) use new GoVPP API, and publishing frequency was significantly decreased to handle creation of multiple 
  interfaces in short period of time.
* [IPSec-plugin](plugins/vpp/ipsecplugin)
  - Model moved to the [api/models](api/models/vpp/ipsec)
  - The IPSec interface is no longer processed by the IPSec plugin (moved to interface plugin).  
  - The ipsec link in interface model now uses the enum definitions from IPSec model. Also some missing crypto algorithms were added.
* [l2-plugin](plugins/vpp/l2plugin)
   - Model moved to the [api/models](api/models/vpp/l2) and split to three separate models for bridge domains, FIBs and cross connects.  
* [l3-plugin](plugins/vpp/l3plugin)
   - Model moved to the [api/models](api/models/vpp/l3) and split to three separate models for ARPs, Proxy ARPs including IP neighbor and Routes.
* [nat-plugin](plugins/vpp/ifplugin)
  - Defined new plugin to handle NAT-related configuration and its own [model](api/models/vpp/nat/nat.proto) (before a part of interface plugin).  
* [punt-plugin](plugins/vpp/puntplugin)
  - Model moved to the [api/models](api/models/vpp/punt/punt.proto).
  - Added retrieve support for punt socket. The current implementation is not final - plugin uses local cache (it will be enhanced when the appropriate VPP binary API call will be added).
* [stn-plugin](plugins/vpp/stnplugin)
  - Model moved to the [api/models](api/models/vpp/stn/stn.proto).
* [linux-if-plugin](plugins/linux/ifplugin)
  - Model reviewed, updated and moved to the [api/models](api/models/linux/interfaces/interface.proto).  
* [linux-l3-plugin](plugins/linux/l3plugin)
  - Model moved to the [api/models](api/models/linux/l3) and split to separate models for ARPs and Routes.
  - Linux routes and ARPs have a new dependency - the target interface is required to contain an IP address. 
* [ns-plugin](plugins/linux/nsplugin)
  - New auxiliary plugin to handle linux namespaces and microservices (evolved from ns-handler). Also defines [model](api/models/linux/namespace/namespace.proto) for generic linux namespace definition. 

### Docker Images
* Configuration file for GoVPP was removed, forcing to use default values (which are the same as they were in the file).
* Fixes for installing ARM64 debugger.
* Kafka is no longer required in order to run vpp-agent from the image.

### Documentation
* Added documentation for the [punt plugin](plugins/vpp/puntplugin/README.md), describing main features and usage of the punt plugin.
* Added documentation for the [IPSec plugin](plugins/vpp/ipsecplugin/README.md), describing main and usage of the IPSec plugin.
* Added documentation for the [interface plugin](plugins/vpp/ifplugin). The document is only available on [wiki page](https://github.com/ligato/vpp-agent/wiki/VPP-Interface-plugin).
* Description improved in various proto files.
* Added a lot of new documentation for the KVScheduler (examples, troubleshooting, debugging guides, diagrams, ...)
* Added tutorial for KV Scheduler.
* Added many new documentation articles to the [wiki page](https://github.com/ligato/vpp-agent/wiki). However, most of is there only temporary
since we are preparing new ligato.io website with all the documentation and other information about the Ligato project. Also majority of readme 
files from the vpp-agent repository will be removed in the future.

<a name="v1.8.1"></a>
# [1.8.1](https://github.com/ligato/vpp-agent/compare/v1.8..v1.8.1) (2019-03-04)

Motive for this minor release was updated VPP with several fixed bugs from the previous version. The VPP version also introduced new interface statistics mechanism, thus the stats processing was updated in the interface plugin.

### Compatibility
- v19.01-16~gd30202244
- cn-infra v1.7
- GO 1.11

### Bug Fixes
- VPP bug: fixed crash when attempting to run in kubernetes pod 
- VPP bug: fixed crash in barrier sync when vlib_worker_threads is zero

### Features
- [vpp-ifplugin](plugins/vpp/ifplugin)
  * Support for new VPP stats (the support for old ones were deprecated by the VPP, thus removed from the vpp-agent as well).
  
  
<a name="v1.8.0"></a>
# [1.8.0](https://github.com/ligato/vpp-agent/compare/v1.7...v1.8) (2018-12-12)

### Compatibility
- VPP v19.01-rc0~394-g6b4a32de
- cn-infra v1.7
- Go 1.11
    
### Bug Fixes
  * Pre-existing VETH-type interfaces are now read from the default OS namespace during resync if the Linux interfaces were dumped.  
  * The Linux interface dump method does not return an error if some interface namespace becomes suddenly unavailable at the read-time. Instead, this case is logged and all the other interfaces are returned as usual.
  * The Linux localclient's delete case for Linux interfaces now works properly.
  * The Linux interface dump now uses OS link name (instead of vpp-agent specific name) to read the interface attributes. This sometimes caused errors where an incorrect or even none interface was read.
  * Fixed bug where the unsuccessful namespace switch left the namespace file opened.
  * Fixed crash if the Linux plugin was disabled.
  * Fixed occasional crash in vpp-agent interface notifications.
  * Corrected interface counters for TX packets.
  * Access list with created TCP/UDP/ICMP rule, which remained as empty struct no longer causes vpp-agent to crash

### Features
- [vpp-ifplugin](plugins/vpp/ifplugin)
  * Rx-mode and Rx-placement now support dump via the respective binary API call
- [vpp-rpc-plugin](plugins/vpp/rpc)
  * GRPC now supports also IPSec configuration.
  * All currently supported configuration items can be also dumped/read via GRPC (similar to rest) 
  * GRPC now allows to automatically persist configuration to the data store. The desired DB has to be defined in the new GRPC config file (see [readme](plugins/vpp/rpc/README.md) for additional information).
- [vpp-punt](plugins/vpp/puntplugin)
  * Added simple new [punt plugin](plugins/vpp/puntplugin/README.md). The plugin allows to register/unregister punt to host via Unix domain socket. The new [model](plugins/vpp/model/punt/punt.proto) was added for this configuration type. Since the VPP API is incomplete, the configuration does not support dump.  
  
### Improvements 
- [vpp-ifplugin](plugins/vpp/ifplugin)
  * The VxLAN interface now support IPv4/IPv6 virtual routing and forwarding (VRF tables). 
  * Support for new interface type: VmxNet3. The VmxNet3 virtual network adapter has no physical counterpart since it is optimized for performance in a virtual machine. Because built-in drivers for this card are not provided by default in the OS, the user must install VMware Tools. The interface model was updated for the VmxNet3 specific configuration.
- [ipsec-plugin](plugins/vpp/ipsecplugin)
  * IPSec resync processing for security policy databases (SPD) and security associations (SA) was improved. Data are properly read from northbound and southbound, compared and partially configured/removed, instead of complete cleanup and re-configuration. This does not appeal to IPSec tunnel interfaces.
  * IPSec tunnel can be now set as an unnumbered interface.
- [rest-plugin](plugins/rest)
  * In case of error, the output returns correct error code with cause (parsed from JSON) instead of an empty body  


<a name="v1.7.0"></a>
# [1.7.0](https://github.com/ligato/vpp-agent/compare/v1.6...v1.7) (2018-10-02)

### Compatibility
- VPP 18.10-rc0~505-ge23edac
- cn-infra v1.6
- Go 1.11
  
### Bug Fixes
  * Corrected several cases where various errors were silently ignored
  * GRPC registration is now done in Init() phase, ensuring that it finishes before GRPC server
    is started
  * Removed occasional cases where Linux tap interface was not configured correctly
  * Fixed FIB configuration failures caused by wrong updating of the metadata after several 
    modifications
  * No additional characters are added to NAT tag and can be now configured with the full length 
    without index out of range errors
  * Linux interface resync registers all VETH-type interfaces, despite the peer is not known
  * Status publishing to ETCD/Consul now should work properly
  * Fixed occasional failure caused by concurrent map access inside Linux plugin interface 
    configurator
  * VPP route dump now correctly recognizes route type

### Features
- [vpp-ifplugin](plugins/vpp/ifplugin)
  * It is now possible to dump unnumbered interface data
  * Rx-placement now uses specific binary API to configure instead of generic CLI API
- [vpp-l2plugin](plugins/vpp/l2plugin)
  * Bridge domain ARP termination table can now be dumped  
- [vpp-ifplugin](plugins/linux/ifplugin)
  * Linux interface watcher was reintroduced.
  * Linux interfaces can be now dumped.
- [vpp-l3plugin](plugins/linux/l3plugin)
  * Linux ARP entries and routes can be dumped.   

### Improvements
- [vpp-plugins](plugins/vpp)
  * Improved error propagation in all the VPP plugins. Majority of errors now print the stack trace to
  the log output allowing better error tracing and debugging.
  * Stopwatch was removed from all vppcalls
- [linux-plugins](plugins/linux)
  * Improved error propagation in all Linux plugins (same way as for VPP)
  * Stopwatch was removed from all linuxcalls
- [govpp-plugn](plugins/govppmux)
  * Tracer (introduced in cn-infra 1.6) added to VPP message processing, replacing stopwatch. 
  The measurement should be more precise and logged for all binary API calls. Also the rest
  plugin now allows showing traced entries. 
  
### Docker Images
  * The image can now be built on ARM64 platform   


<a name="v1.6.0"></a>
# [1.6.0](https://github.com/ligato/vpp-agent/compare/v1.5.2...v1.6) (2018-08-24)

### Compatibility
- VPP 18.10-rc0~169-gb11f903a
- cn-infra v1.5
  
### BREAKING CHANGES
- Flavors were replaced with [new way](cmd/vpp-agent/app) of managing plugins.
- REST interface URLs were changed, see [readme](plugins/rest/README.md) for complete list.

### Bug Fixes
* if VPP routes are dumped, all paths are returned
* NAT load-balanced static mappings should be resynced correctly  
* telemetry plugin now correctly parses parentheses for `show node counters`
* telemetry plugin will not hide an error caused by value loading if the config file is not present
* Linux plugin namespace handler now correctly handles namespace switching for interfaces with IPv6 addresses.
  Default IPv6 address (link local) will not be moved to the new namespace if there are no more IPv6 addresses
  configured within the interface. This should prevent failures in some cases where IPv6 is not enabled in the
  destination namespace.
* VxLAN with non-zero VRF can be successfully removed
* Lint is now working again    
* VPP route resync works correctly if next hop IP address is not defined

### Features
* Deprecating flavors
  - CN-infra 1.5 brought new replacement for flavors and it would be a shame not to implement it
    in the vpp-agent. The old flavors package was removed and replaced with this new concept,
    visible in [app package vpp-agent](cmd/vpp-agent/app/vpp_agent.go).
- [rest plugin](plugins/rest)
  * All VPP configuration types are now supported to be dumped using REST. The output consists of two parts;
    data formatted as NB proto model, and metadata with VPP specific configuration (interface indexes,
    different counters, etc.).
  * REST prefix was changed. The new URL now contains API version and purpose (dump, put). The list of all 
    URLs can be found in the [readme](plugins/rest/README.md)
- [ifplugin](plugins/vpp/ifplugin)
  * Added support for NAT virtual reassembly for both, IPv4 and IPv6. See change in 
    [nat proto file](plugins/vpp/model/nat/nat.proto)
- [l3plugin](plugins/vpp/l3plugin)
  * Vpp-agent now knows about DROP-type routes. They can be configured and also dumped. VPP default routes, which are
    DROP-type is recognized and registered. Currently, resync does not remove or correlate such a route type
    automatically, so no default routes are unintentionally removed.
  * New configurator for L3 IP scan neighbor was added, allowing to set/unset IP scan neigh parameters to the VPP.
  
### Improvements
- [vpp plugins](plugins/vpp)
  * all vppcalls were unified under API defined for every configuration type (e.g. interfaces, l2, l3, ...).
  Configurators now use special handler object to access vppcalls. This should prevent duplicates and make
  vppcalls cleaner and more understandable.
- [ifplugin](plugins/vpp/ifplugin)
  * VPP interface DHCP configuration can now be dumped and added to resync processing
  * Interfaces and also L3 routes can be configured for non-zero VRF table if IPv6 is used. 
- [examples](examples)
  * All examples were reworked to use new flavors concept. The purpose was not changed.    

### Docker Images
- using Ubuntu 18.04 as the base image


<a name="v1.5.2"></a>
## [1.5.2](https://github.com/ligato/vpp-agent/compare/v1.5.1...v1.5.2) (2018-07-23)

### Compatibility
- VPP 18.07-rc0~358-ga5ee900
- cn-infra v1.4.1 (minor version fixes bug in Consul)

### Bug Fixes
- [Telemetry](plugins/telemetry)
  * Fixed bug where lack of config file could cause continuous polling. The interval now also
  cannot be changed to a value less than 5 seconds.
  * Telemetry plugin is now closed properly


<a name="v1.5.1"></a>
## 1.5.1 (2018-07-20)

### Compatibility
- VPP 18.07-rc0~358-ga5ee900
- cn-infra v1.4

### Features
- [Telemetry](plugins/telemetry)
  * Default polling interval was raised to 30s.
  * Added option to use telemetry config file to change polling interval, or turn the polling off,
  disabling the telemetry plugin. The change was added due to several reports where often polling
  is suspicious of interrupting VPP worker threads and causing packet drops and/or other
  negative impacts. More information how to use the config file can be found 
  in the [readme](plugins/telemetry/README.md)


<a name="v1.5.0"></a>
# [1.5.0](https://github.com/ligato/vpp-agent/compare/v1.4.1...v1.5) (2018-07-16)

### Compatibility
- VPP 18.07-rc0~358-ga5ee900
- cn-infra v1.4

### BREAKING CHANGES
- The package `etcdv3` was renamed to `etcd`, along with its flag and configuration file.
- The package `defaultplugins` was renamed to `vpp` to make the purpose of the package clear

### Bug Fixes
- Fixed a few issues with parsing VPP metrics from CLI for [Telemetry](plugins/telemetry).
- Fixed bug in GoVPP occurring after some request timed out, causing
  the channel to receive replies from the previous request and always returning an error.
- Fixed issue which prevented setting interface to non-existing VRF.
- Fixed bug where removal of an af-packet interface caused attached Veth to go DOWN.
- Fixed NAT44 address pool resolution which was not correct in some cases.
- Fixed bug with adding SR policies causing incomplete configuration.

### Features
- [LinuxPlugin](plugins/linux)
  * Is now optional and can be disabled via configuration file.
- [ifplugin](plugins/vpp/ifplugin)
  * Added support for VxLAN multicast
  * Rx-placement can be configured on VPP interfaces
- [IPsec](plugins/vpp/ipsecplugin)
  * IPsec UDP encapsulation can now be set (NAT traversal)  

### Docker Images
- Replace `START_AGENT` with `OMIT_AGENT` to match `RETAIN_SUPERVISOR`
  and keep both unset by default.
- Refactored and cleaned up execute scripts and remove unused scripts.
- Fixed some issues with `RETAIN_SUPERVISOR` option.
- Location of supervisord pid file is now explicitly set to
  `/run/supervisord.pid` in *supervisord.conf* file.
- The vpp-agent is now started  with single flag `--config-dir=/opt/vpp-agent/dev`,
  and will automatically load all configuration from that directory.


<a name="v1.4.1"></a>
## [1.4.1](https://github.com/ligato/vpp-agent/compare/v1.4.0...v1.4.1) (2018-06-11)

A minor release using newer VPP v18.04 version.

### Compatibility
- VPP v18.04 (2302d0d)
- cn-infra v1.3

### Bug Fixes
- VPP submodule was removed from the project. It should prevent various problems with dependency
  resolution.
- Fixed known bug present in the previous version of the VPP, issued as
  [VPP-1280](https://jira.fd.io/browse/VPP-1280). Current version contains appropriate fix.  


<a name="v1.4.0"></a>
# [1.4.0](https://github.com/ligato/vpp-agent/compare/v1.3...v1.4) (2018-05-24)

### Compatibility
- VPP v18.04 (ac2b736)
- cn-infra v1.3

### Bug Fixes
  * Fixed case where the creation of the Linux route with unreachable gateway threw an error. 
    The route is now appropriately cached and created when possible. 
  * Fixed issue with GoVPP channels returning errors after a timeout.
  * Fixed various issues related to caching and resync in L2 cross-connect
  * Split horizon group is now correctly assigned if an interface is created after bridge domain
  * Fixed issue where the creation of FIB while the interface was not a part of the bridge domain returned an error. 

### Known issues
  * VPP crash may occur if there is interface with non-default VRF (>0). There is an 
    [VPP-1280](https://jira.fd.io/browse/VPP-1280) issue created with more details 

### Features
- [Consul](https://www.consul.io/)
  * Consul is now supported as a key-value store alternative to ETCD.
    More information in the [readme](https://github.com/ligato/cn-infra/blob/master/db/keyval/consul/README.md).
- [Telemetry](plugins/telemetry)
  * New plugin for collecting telemetry data about VPP metrics
    and serving them via HTTP server for Prometheus.
    More information in the [readme](plugins/telemetry/README.md).
- [Ipsecplugin](plugins/vpp/ipsecplugin)
  * Now supports tunnel interface for encrypting all the data
    passing through that interface.
- [GRPC](plugins/vpp/rpc) 
  * Vpp-agent itself can act as a GRPC server (no need for external executable)
  * All configuration types are supported (incl. Linux interfaces, routes and ARP)
  * Client can read VPP notifications via vpp-agent.
- [SR plugin](plugins/vpp/srplugin)
  * New plugin with support for Segment Routing.
    More information in the [readme](plugins/vpp/srplugin/README.md).

### Improvements
- [ifplugin](plugins/vpp/ifplugin) 
  * Added support for self-twice-NAT
- __vpp-agent-grpc__ executable merged with [vpp-agent](cmd/vpp-agent) command.
- [govppmux](plugins/govppmux) 
  * [configure reply timeout](plugins/govppmux/README.md) can be configured.
  * Support for VPP started with custom shared memory prefix. SHM may be configured via the GoVPP
  plugin config file. More info in the [readme](plugins/govppmux/README.md)
  * Overall redundancy cleanup and corrected naming for all proto models.
  * Added more unit tests for increased coverage and code stability. 

### Documentation
- [localclient_linux](examples/localclient_vpp) now contains two examples, the old one demonstrating
  basic plugin functionality was moved to [plugin](examples/localclient_vpp/plugins) package, and specialised example for [NAT](examples/localclient_vpp/nat) was added.
- [localclient_linux](examples/localclient_linux) now contains two examples, the old one demonstrating 
  [veth](examples/localclient_linux/veth) interface usage was moved to package and new example for linux
  [tap](examples/localclient_linux/tap) was added.


<a name="v1.3.0"></a>
# [1.3.0](https://github.com/ligato/vpp-agent/compare/v1.2...v1.3) (2018-03-22)

The vpp-agent is now using custom VPP branch [stable-1801-contiv](https://github.com/vpp-dev/vpp/tree/stable-1801-contiv).

### Compatibility
- VPP v18.01-rc0~605-g954d437
- cn-infra v1.2

### Bug Fixes
  * Resync of ifplugin in both, VPP and Linux, was improved. Interfaces
    with the same configuration data are not recreated during resync.
  * STN does not fail if IP address with a mask is provided.
  * Fixed ingress/egress interface resolution in ACL.
  * Linux routes now check network reachability for gateway address b
    before configuration. It should prevent "network unreachable" errors
    during config.
  * Corrected bridge domain crash in case non-bvi interface was added to
    another non-bvi interface.
  * Fixed several bugs related to VETH and AF-PACKET configuration and resync.

### Features
- [ipsecplugin](plugins/vpp/ipsecplugin)
  * New plugin for IPSec added. The IPSec is supported for VPP only
    with Linux set manually for now. IKEv2 is not yet supported.
    More information in the [readme](plugins/vpp/ipsecplugin/README.md).
- [nsplugin](plugins/linux/nsplugin)
  * New namespace plugin added. The configurator handles common namespace
    and microservice processing and communication with other Linux plugins.
- [ifplugin](plugins/vpp/ifplugin)
  * Added support for Network address translation. NAT plugin supports
    a configuration of NAT44 interfaces, address pools and DNAT.
    More information in the [readme](plugins/vpp/ifplugin/README.md).
  * DHCP can now be configured for the interface  
- [l2plugin](plugins/vpp/l2plugin)
  * Split-horizon group can be configured for bridge domain interface.
- [l3plugin](plugins/vpp/l3plugin)
  * Added support for proxy ARP. For more information and configuration 
    example, please see [readme](plugins/vpp/l3plugin/README.md).
- [linux ifplugin](plugins/linux/ifplugin)
  * Support for automatic interface configuration (currently only TAP).
        
### Improvements
- [aclplugin](plugins/vpp/aclplugin)
  * Removed configuration order of interfaces. The access list can be now 
    configured even if interfaces do not exist yet, and add them later.
- [vpp-agent-ctl](cmd/vpp-agent-ctl) 
  * The vpp-agent-ctl was refactored and command info was updated.

### Docker Images
  * VPP can be built and run in the release or debug mode.
  Read more information in the [readme](https://github.com/ligato/vpp-agent/blob/pantheon-dev/docker/dev/README.md).
  * Production image is now smaller by roughly 40% (229MB).


<a name="v1.2.0"></a>
# [1.2.0](https://github.com/ligato/vpp-agent/compare/v1.1...v1.2) (2018-02-07)

### Compatibility
- VPP v18.04-rc0~90-gd95c39e
- cn-infra v1.1

### Bug Fixes
- Fixed interface assignment in ACLs
- Fixed bridge domain BVI modification resolution
- [vpp-agent-grpc](cmd/vpp-agent) (removed in 1.4 release, since then it is a part of the vpp-agent) now compiles properly together with other commands.

### Known Issues
- VPP can occasionally cause a deadlock during checksum calculation (https://jira.fd.io/browse/VPP-1134)
- VPP-Agent might not properly handle initialization across plugins (this is not occurring currently, but needs to be tested more)

### Improvements
- [aclplugin](plugins/vpp/aclplugin) 
  * Improved resync of ACL entries. Every new ACL entry is correctly configured in the VPP and all obsolete entries are read and removed. 
- [ifplugin](plugins/vpp/ifplugin) 
  * Improved resync of interfaces, BFD sessions, authentication keys, echo functions and STN. Better resolution of persistence config for interfaces. 
- [l2plugin](plugins/vpp/l2plugin) 
  * Improved resync of bridge domains, FIB entries, and xConnect pairs. Resync now better correlates configuration present on the VPP with the NB setup.
- [linux-ifplugin](plugins/linux/l3plugin) 
  * ARP does not need the interface to be present on the VPP. Configuration is cached and put to the VPP if requirements are fulfilled. 
- Dependencies
  * Migrated from glide to dep  

### Docker Images
  * VPP compilation now skips building of Java/C++ APIs,
    this saves build time and final image size.
  * Development image now runs VPP in debug mode with
    various debug options added in [VPP config file](docker/dev/vpp.conf).


<a name="v1.1.0"></a>
# [1.1.0](https://github.com/ligato/vpp-agent/compare/v1.0.8...v1.1) (2018-01-22)

### Compatibility
- VPP version v18.04-rc0~33-gb59bd65
- cn-infra v1.0.8

### Bug Fixes
- fixed skip-resync parameter if vpp-plugin.conf is not provided.
- corrected af_packet type interface behavior if veth interface is created/removed.
- several fixes related to the af_packet and veth interface type configuration.
- microservice and veth-interface related events are synchronized.

### Known Issues
- VPP can occasionally cause a deadlock during checksum calculation (https://jira.fd.io/browse/VPP-1134)
- VPP-Agent might not properly handle initialization across plugins (this is not occurring currently, but needs to be tested more)

### Features
- [ifplugin](plugins/vpp/ifplugin)
    - added support for un-numbered interfaces. The nterface can be marked as un-numbered with information
    about another interface containing required IP address. A un-numbered interface does not need to have 
    IP address set.
    - added support for virtio-based TAPv2 interfaces.
    - interface status is no longer stored in the ETCD by default and it can be turned on using the appropriate
    setting in vpp-plugin.conf. See  [readme](plugins/vpp/README.md) for more details.  
- [l2plugin](plugins/vpp/l2plugin)
    - bridge domain status is no longer stored in the ETCD by default and it can be turned on using the appropriate
    setting in vpp-plugin.conf. See  [readme](plugins/vpp/README.md) for more details.  

### Improvements
- [ifplugin](plugins/vpp/ifplugin)
    - default MTU value was removed in order to be able to just pass empty MTU field. MTU now can be
    set only in interface configuration (preferred) or defined in vpp-plugin.conf. If none of them
    is set, MTU value will be empty.
    - interface state data are stored in statuscheck readiness probe
- [l3plugin](plugins/vpp/l3plugin)
    - removed strict configuration order for VPP ARP entries and routes. Both ARP entry or route can 
    be configured without interface already present.
- [l4plugin](plugins/vpp/l4plugin)
   - removed strict configuration order for application namespaces. Application namespace can 
    be configured without interface already present.
- localclient
   - added API for ARP entries, L4 features, Application namespaces, and STN rules.
- logging
   - consolidated and improved logging in vpp and Linux plugins.     


<a name="v1.0.8"></a>
## [1.0.8](https://github.com/ligato/vpp-agent/compare/v1.0.7...v1.0.8) (2017-11-21)

### Compatibility
- VPP v18.01-rc0-309-g70bfcaf
- cn-infra v1.0.7

### Features
- [ifplugin](plugins/vpp/ifplugin)
   - ability to configure STN rules.  See respective
   [readme](plugins/vpp/ifplugin/README.md) in interface plugin for more details.
   - rx-mode settings can be set on interface. Ethernet-type interface can be set to POLLING mode, other types of interfaces supports also INTERRUPT and ADAPTIVE. Fields to set QueueID/QueueIDValid are also available
   - added possibility to add interface to any VRF table.
   - added defaultplugins API.
   - API contains new Method `DisableResync(keyPrefix ...string)`. One or more ETCD key prefixes can be used as a parameter to disable resync for that specific key(s).
- [l4plugin](plugins/vpp/l4plugin)
   - added new l4 plugin to the VPP plugins. It can be used to enable/disable L4 features
   and configure application namespaces. See respective
    [readme](plugins/vpp/l4plugin/README.md) in L4 plugin for more details.
   - support for VPP plugins/l3plugin ARP configuration. The configurator can perform the
   basic CRUD operation with ARP config.
- [resync](plugins/vpp/data_resync.go)
  - resync error propagation improved. If any resynced configuration fails, rest of the resync completes and will not be interrupted. All errors which appear during resync are logged after. 
- [linux l3plugin](plugins/linux/l3plugin)
  - route configuration does not return an error if the required interface is missing. Instead, the route data are internally stored and configured when the interface appears.  
- GoVPP
  - delay flag removed from GoVPP plugin 

### Improvements
- removed dead links from README files

### Documentation
- improved in multiple vpp-agent packages


<a name="v1.0.7"></a>
## [1.0.7](https://github.com/ligato/vpp-agent/compare/v1.0.6...v1.0.7) (2017-10-30)

### Compatibility
- VPP version v18.01-rc0~154-gfc1c612
- cn-infra v1.0.6

### Features

- [Default VPP plugin](plugins/vpp)
    - added resync strategies. Resync of VPP plugins can be set using
    defaultpluigns config file; Resync can be set to full (always resync everything) or
    dependent on VPP configuration (if there is none, skip resync). Resync can be also
    forced to skip using the parameter. See appropriate changelog in 
    [VPP plugins](plugins/vpp) for details.
- [Linuxplugins L3Plugin](plugins/linux/l3plugin)
    - added support for basic CRUD operations with the static Address resolution protocol 
    entries and static Routes.


<a name="v1.0.6"></a>
## [1.0.6](https://github.com/ligato/vpp-agent/compare/v1.0.5...v1.0.6) (2017-10-17)

### Compatibility
- cn-infra v1.0.5

### Features

- [LinuxPlugin](plugins/linux)
   - The configuration of vEth interfaces modified. Veth configuration defines
   two names: symbolic used internally and the one used in host OS.
   `HostIfName` field is optional. If it is not defined, the name in the host OS
   will be the same as the symbolic one - defined by `Name` field.


<a name="v1.0.5"></a>
## [1.0.5](https://github.com/ligato/vpp-agent/compare/v1.0.4...v1.0.5) (2017-09-26)

### Compatibility
- VPP version v17.10-rc0~334-gce41a5c
- cn-infra v1.0.4

### Features

- [GoVppMux](plugins/govppmux)
    - configuration file for govpp added
- [Kafka Partitions](vendor/github.com/ligato/cn-infra/messaging/kafka)
    - Changes in offset handling, only automatically partitioned messages (hash, random)
      have their offset marked. Manually partitioned messages are not marked.
    - Implemented post-init consumer (for manual partitioner only) which allows starting
      consuming after kafka-plugin Init()
    - Minimalistic examples & documentation for Kafka API will be improved in a later release.


<a name="v1.0.4"></a>
## [1.0.4](https://github.com/ligato/vpp-agent/compare/v1.0.3...v1.0.4) (2017-09-08)

### Features

- [Kafka Partitions](vendor/github.com/ligato/cn-infra/messaging/kafka)
    - Implemented new methods that allow to specify partitions & offset parameters:
      * publish: Mux.NewSyncPublisherToPartition() & Mux.NewAsyncPublisherToPartition()
      * watch: ProtoWatcher.WatchPartition()
    - Minimalistic examples & documentation for Kafka API will be improved in a later release.
- [Flavors](flavors)
    - reduced to only [local.FlavorVppLocal](flavors/local/local_flavor.go) & [vpp.Flavor](flavors/vpp/vpp_flavor.go)
- [GoVVPP]
    - updated version waits until the VPP is ready to accept a new connection


<a name="v1.0.3"></a>
## [1.0.3](https://github.com/ligato/vpp-agent/compare/v1.0.2...v1.0.3) (2017-09-05)

### Compatibility
- VPP version v17.10-rc0~265-g809bc74 (upgraded because of VPP MEMIF fixes)

### Features

Enabled support for wathing data store `OfDifferentAgent()` - see:
* [examples/idx_iface_cache](examples/idx_iface_cache/main.go)
* [examples/examples/idx_bd_cache](examples/idx_bd_cache/main.go)
* [examples/idx_veth_cache](examples/idx_veth_cache/main.go)

Preview of new Kafka client API methods that allows to fill also partition and offset argument. New methods implementation ignores these new parameters for now (fallback to existing implementation based on `github.com/bsm/sarama-cluster` and `github.com/Shopify/sarama`).


<a name="v1.0.2"></a>
## [1.0.2](https://github.com/ligato/vpp-agent/compare/v1.0.1...v1.0.2) (2017-08-28)

### Compatibility
- VPP version v17.10-rc0~203

### Known Issues
A rarely occurring problem during startup with binary API connectivity.
VPP rejects binary API connectivity when VPP Agent tries to connect
too early (plan fix this behavior in next release).

### Features

Algorithms for applying northbound configuration (stored in ETCD key-value data store)
to VPP in the proper order of VPP binary API calls implemented in [Default VPP plugin](plugins/vpp):
- network interfaces, especially:
  - MEMIFs (optimized data plane network interface tailored for a container to container network connectivity)
  - VETHs (standard Linux Virtual Ethernet network interface)
  - AF_Packets (for accessing VETHs and similar type of interface)
  - VXLANs, Physical Network Interfaces, loopbacks ...
- L2 BD & X-Connects
- L3 IP Routes & VRFs
- ACL (Access Control List)

Support for Linux VETH northbound configuration implemented in [Linux Plugin](plugins/linux)
applied in proper order with VPP AF_Packet configuration.

Data Synchronization during startup for network interfaces & L2 BD
(support for the situation when ETCD contain configuration before VPP Agent starts).

Data replication and events:
- Updating operational data in ETCD (VPP indexes such as  sw_if_index) and statistics (port counters).
- Updating statistics in Redis (optional once redis.conf available - see flags).
- Publishing links up/down events to Kafka message bus.

- [Examples](examples)
- Tools:
  - [agentctl CLI tool](cmd/agentctl) that show state & configuration of VPP agents
  - [docker](docker): container-based development environment for the VPP agent
- other features inherited from cn-infra:
  - [health](https://github.com/ligato/cn-infra/tree/master/health): status check & k8s HTTP/REST probes
  - [logging](https://github.com/ligato/cn-infra/tree/master/logging): changing log level at runtime
- Ability to extend the behavior of the VPP Agent by creating new plugins on top of [VPP Agent flavor](flavors/vpp).
  New plugins can access API for configured:
  - [VPP Network interfaces](plugins/vpp/ifplugin/ifaceidx),
  - [Bridge domains](plugins/vpp/l2plugin/l2idx) and [VETHs](plugins/linux/ifplugin/ifaceidx)
    based on [idxvpp](pkg/idxvpp) threadsafe map tailored for VPP data
    with advanced features (multiple watchers, secondary indexes).
- VPP Agent is embeddable in different software projects and with different systems by using [Local Flavor](flavors/local) to reuse VPP Agent algorithms. For doing this there is [VPP Agent client version 1](clientv1):
  - local client - for embedded VPP Agent (communication inside one operating system process, VPP Agent effectively used as a library)
  - remote client - for remote configuration of VPP Agent (while integrating for example with control plane)
