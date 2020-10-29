# Changelog

[![GitHub commits since latest release](https://img.shields.io/github/commits-since/ligato/vpp-agent/latest.svg?style=flat-square)](https://github.com/ligato/vpp-agent/compare/v3.2.0...master)

## Release Notes

- [v3.2.0](#v3.2.0)
- [v3.1.0](#v3.1.0)
- [v3.0.0](#v3.0.0)
  - [v3.0.1](#v3.0.1)
- [v2.5.0](#v2.5.0)
  - [v2.5.1](#v2.5.1)
- [v2.4.0](#v2.4.0)
- [v2.3.0](#v2.3.0)
- [v2.2.0](#v2.2.0)
- [v2.2.0-beta](#v2.2.0-beta)
- [v2.1.0](#v2.1.0)
  - [v2.1.1](#v2.1.1)
- [v2.0.0](#v2.0.0)
  - [v2.0.1](#v2.0.1)
  - [v2.0.2](#v2.0.2)
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
### COMPATIBILITY
### KNOWN ISSUES
### BREAKING CHANGES
### Bug Fixes
### Features
### Improvements
### Docker Images
### Documentation
-->

<a name="v3.2.0"></a>
# [3.2.0](https://github.com/ligato/vpp-agent/compare/v3.1.0...v3.2.0) (2020-10-XX)

### COMPATIBILITY
- VPP 20.09 (compatible)
- VPP 20.05 (default)
- VPP 20.01 (backwards compatible)
- VPP 19.08 (backwards compatible)
- ~~VPP 19.04~~ (no longer supported)

### Bug Fixes
- Fixes and improvements for agentctl and models [#1643](https://github.com/ligato/vpp-agent/pull/1643)
- Fix creation of multiple ipip tunnels [#1650](https://github.com/ligato/vpp-agent/pull/1650)
- Fix IPSec tun protect + add IPSec e2e test [#1654](https://github.com/ligato/vpp-agent/pull/1654)
- Fix bridge domain dump for VPP 20.05 [#1663](https://github.com/ligato/vpp-agent/pull/1663)
- Fix IPSec SA add/del in VPP 20.05 [#1664](https://github.com/ligato/vpp-agent/pull/1664)
- Update expected output of agentctl status command [#1673](https://github.com/ligato/vpp-agent/pull/1673)
- vpp/ifplugin: Recognize interface name prefix "tun" as TAP [#1674](https://github.com/ligato/vpp-agent/pull/1674)
- Fix IPv4 link-local IP address handling [#1715](https://github.com/ligato/vpp-agent/pull/1715)
- maps caching prometheus gauges weren't really used [#1741](https://github.com/ligato/vpp-agent/pull/1741)
- Permit agent to run even when VPP stats are unavailable [#1712](https://github.com/ligato/vpp-agent/pull/1712)
- Fix grpc context timeout for agentctl import command [#1718](https://github.com/ligato/vpp-agent/pull/1718)
- Changed nat44 pool key to prevent possible key collisions [#1725](https://github.com/ligato/vpp-agent/pull/1725)
- Remove forced delay for linux interface notifications [#1742](https://github.com/ligato/vpp-agent/pull/1742)

### Features
- agentctl: Add config get/update commands [#1709](https://github.com/ligato/vpp-agent/pull/1709)
- agentctl: Support specific history seq num and improve layout [#1717](https://github.com/ligato/vpp-agent/pull/1717)
- agentctl: Add config.resync subcommand (with resync) [#1642](https://github.com/ligato/vpp-agent/pull/1642)
- IP Flow Information eXport (IPFIX) plugin [#1649](https://github.com/ligato/vpp-agent/pull/1649)
- Add IPIP & IPSec point-to-multipoint support [#1669](https://github.com/ligato/vpp-agent/pull/1669)
- Wireguard plugin support [#1731](https://github.com/ligato/vpp-agent/pull/1731)
- Add tunnel mode support for VPP TAP interfaces [#1671](https://github.com/ligato/vpp-agent/pull/1671)
- New REST endpoint for retrieving version of Agent [#1670](https://github.com/ligato/vpp-agent/pull/1670)
- Add support for IPv6 ND address autoconfig [#1676](https://github.com/ligato/vpp-agent/pull/1676)
- Add VRF field to proxy ARP range [#1672](https://github.com/ligato/vpp-agent/pull/1672)
- Switch to new proto v2 (google.golang.org/protobuf) [#1691](https://github.com/ligato/vpp-agent/pull/1691)
- ipsec: allow configuring salt for encryption algorithm [#1698](https://github.com/ligato/vpp-agent/pull/1698)
- gtpu: Add RemoteTeid to GTPU interface [#1719](https://github.com/ligato/vpp-agent/pull/1719)
- Added support for NAT44 static mapping twice-NAT pool IP address reference [#1728](https://github.com/ligato/vpp-agent/pull/1728)
- add IP protocol number to ACL model [#1726](https://github.com/ligato/vpp-agent/pull/1726)
- gtpu: Add support for arbitrary DecapNextNode [#1721](https://github.com/ligato/vpp-agent/pull/1721)
- configurator: Add support for waiting until config update is done [#1734](https://github.com/ligato/vpp-agent/pull/1734)
- telemetry: Add reading VPP threads to the telemetry plugin [#1753](https://github.com/ligato/vpp-agent/pull/1753)
- linux: Add support for Linux VRFs [#1744](https://github.com/ligato/vpp-agent/pull/1744)
- VRRP support [#1744](https://github.com/ligato/vpp-agent/pull/1744)

### Improvements
- perf: Performance enhancement for adding many rules to Linux IP… [#1644](https://github.com/ligato/vpp-agent/pull/1644)
- Improve testing process for e2e/integration tests [#1757](https://github.com/ligato/vpp-agent/pull/1757)

### Documentation
- docs: Add example for developing agents with custom VPP plugins [#1665](https://github.com/ligato/vpp-agent/pull/1665)

### Other
- Delete unused REST handler for VPP commands [#1677](https://github.com/ligato/vpp-agent/pull/1677)
- separate model for IPSec Security Policies [#1679](https://github.com/ligato/vpp-agent/pull/1679)
- do not mark host_if_name for AF_PACKET as deprecated [#1745](https://github.com/ligato/vpp-agent/pull/1745)
- Store interface internal name & dev type as metadata [#1706](https://github.com/ligato/vpp-agent/pull/1706)
- Check if HostIfName contains non-printable characters [#1662](https://github.com/ligato/vpp-agent/pull/1662)
- Fix error message for duplicate keys [#1659](https://github.com/ligato/vpp-agent/pull/1659)


<a name="v3.1.0"></a>
# [3.1.0](https://github.com/ligato/vpp-agent/compare/v3.0.0...v3.1.0) (2020-03-13)

### BREAKING CHANGES
* Switch cn-infra dependency to using vanity import path [#1620](https://github.com/ligato/vpp-agent/pull/1620)

  To migrate, replace all cn-infra import paths (`github.com/ligato/cn-infra` -> `go.ligato.io/cn-infra/v2`)

  To update cn-infra dependency, run `go get -u go.ligato.io/cn-infra/v2@master`.

### Bug Fixes
* Add missing models to ConfigData [#1625](https://github.com/ligato/vpp-agent/pull/1625)
* Fix watching VPP events [#1640](https://github.com/ligato/vpp-agent/pull/1640)

### Features
* Allow customizing polling from stats poller [#1634](https://github.com/ligato/vpp-agent/pull/1634)
* IPIP tunnel + IPSec tunnel protection support [#1638](https://github.com/ligato/vpp-agent/pull/1638)
* Add prometheus metrics to govppmux [#1626](https://github.com/ligato/vpp-agent/pull/1626)
* Add prometheus metrics to kvscheduler [#1630](https://github.com/ligato/vpp-agent/pull/1630)

### Improvements
* Improve performance testing suite [#1630](https://github.com/ligato/vpp-agent/pull/1630)

<a name="v3.0.1"></a>
# [3.0.1](https://github.com/ligato/vpp-agent/compare/v3.0.0...v3.0.1) (2020-02-20)

### Bug Fixes
* Add missing models to ConfigData (https://github.com/ligato/vpp-agent/pull/1625)

<a name="v3.0.0"></a>
# [3.0.0](https://github.com/ligato/vpp-agent/compare/v2.5.0...master) (2020-02-10)

### COMPATIBILITY
- **VPP 20.01** (default)
- **VPP 19.08.1** (recommended)
- **VPP 19.04.4**

### KNOWN ISSUES
- VPP L3 plugin: `IPScanNeighbor` was disabled for VPP 20.01 due to VPP API changes (will be implemented later using new model)
- VPP NAT plugin: `VirtualReassembly` in `Nat44Global` was disabled for VPP 20.01 due to VPP API changes (will be implemented later in VPP L3 plugin using new model)

### BREAKING CHANGES
- migrate from dep to Go modules for dependency management and remove vendor directory [#1599](https://github.com/ligato/vpp-agent/pull/1599)
- use vanity import path  `go.ligato.io/vpp-agent/v3` in Go files [#1599](https://github.com/ligato/vpp-agent/pull/1599)
- move all _.proto_ files into `proto/ligato` directory and add check for breaking changes [#1599](https://github.com/ligato/vpp-agent/pull/1599)

### Bug Fixes
- check for duplicate Linux interface IP address [#1586](https://github.com/ligato/vpp-agent/pull/1586)

### New Features
- VPP interface plugin: Allow AF-PACKET to reference target Linux interface via logical name [#1616](https://github.com/ligato/vpp-agent/pull/1616)
- VPP L3 plugin: add support for L3 cross-connects [#1602](https://github.com/ligato/vpp-agent/pull/1602)
- VPP L3 plugin: IP flow hash settings support [#1610](https://github.com/ligato/vpp-agent/pull/1610)
- VPP NAT plugin: NAT interface and AddressPool API changes [#1595](https://github.com/ligato/vpp-agent/pull/1595)
- VPP plugins: support disabling VPP plugins [#1593](https://github.com/ligato/vpp-agent/pull/1593)
- VPP client: add support for govpp proxy [#1593](https://github.com/ligato/vpp-agent/pull/1593)

### Improvements
- optimize getting model keys, with up to 20% faster transactions [#1615](https://github.com/ligato/vpp-agent/pull/1615)
- agentctl output formatting improvements (#1581, #1582, #1589)
- generated VPP binary API now imports common types from `*_types` packages
- development docker images now have smaller size (~400MB less)
- start using Github Workflows for CI/CD pipeline
- add gRPC reflection service

<a name="v2.5.1"></a>
# [2.5.1](https://github.com/ligato/vpp-agent/compare/v2.5.0...v2.5.1) (2019-12-06)

### COMPATIBILITY
- **VPP 20.01-379** (`20.01-rc0~379-ga6b93eac5`)
- **VPP 20.01-324** (`20.01-rc0~324-g66a332cf1`)
- **VPP 19.08.1** (default)
- **VPP 19.04** (backward compatible)
- cn-infra v2.2

### Bug Fixes
* Fix linux interface dump ([#1577](https://github.com/ligato/vpp-agent/pull/1577))
* Fix VRF for SR policy ([#1578](https://github.com/ligato/vpp-agent/pull/1578))

<a name="v2.5.0"></a>
# [2.5.0](https://github.com/ligato/vpp-agent/compare/v2.4.0...v2.5.0) (2019-11-29)

### Compatibility
- **VPP 20.01-379** (`20.01-rc0~379-ga6b93eac5`)
- **VPP 20.01-324** (`20.01-rc0~324-g66a332cf1`)
- **VPP 19.08.1** (default)
- **VPP 19.04** (backward compatible)
- cn-infra v2.2

### New Features
* SRv6 global config (encap source address)
* Support for Linux configuration dumping

### Bug Fixes
* Update GoVPP with fix for stats conversion panic 

<a name="v2.4.0"></a>
# [2.4.0](https://github.com/ligato/vpp-agent/compare/v2.3.0...v2.4.0) (2019-10-21)

### Compatibility
- **VPP 20.01-379** (`20.01-rc0~379-ga6b93eac5`)
- **VPP 20.01-324** (`20.01-rc0~324-g66a332cf1`)
- **VPP 19.08.1** (default)
- **VPP 19.04** (backward compatible)
- cn-infra v2.2

### New Features
This release introduces compatibility with two different commits of the VPP 20.01. Previously compatible version was updated to commit `324-g66a332cf1`, and support for `379-ga6b93eac5` was added. Other previous versions remained.
* [Telemetry][vpp-telemetry]
  - Added `StatsPoller` service periodically retrieving VPP stats.

<a name="v2.3.0"></a>
# [2.3.0](https://github.com/ligato/vpp-agent/compare/v2.2.0...v2.3.0) (2019-10-04)

### Compatibility
- **VPP 20.01** (`20.01-rc0~161-ge5948fb49~b3570`)
- **VPP 19.08.1** (default)
- **VPP 19.04** (backward compatible)
- cn-infra v2.2

VPP support for version 19.08 was updated to 19.08.1. 
Support for 19.01 was dropped in this release. 

### Bug Fixes
* Linux interfaces with 'EXISTING' type should be resynced properly.
* Resolved issue with SRv6 removal.
* AgentCTL dump command fixed.
* ACL ICMP rule is now properly configured and data can be obtained using the ACL dump.
* Missing dependency for SRv6 L2 steering fixed.
* Fixed issue with possible division by zero and missing interface MTU.
* Namespace plugin uses a Docker event listener instead of periodical polling. This should prevent cases where quickly started microservice container was not detected.

### New Features
* [netalloc-plugin][netalloc-plugin]
  - A new plugin called netalloc which allows disassociating topology from addressing in the network configuration. Interfaces, routes and other network objects' addresses can be symbolic references into the pool of allocated addresses known to netalloc plugin. See [model][netalloc-plugin-model] for more information.
* [if-plugin][vpp-interface-plugin]
  - Added support for GRE tunnel interfaces. Choose the `GRE_TUNNEL` interface type with appropriate link data.
* [agentctl][agentctl]
  - Many new features and enhancements added to the AgentCTL:
    * version is defined as a parameter for root command instead of the separate command  
    * ETCD endpoints can be defined via the `ETCD_ENDPOINTS` environment variable
    * sub-command `config` supports `get/put/del` commands
    * `model` sub-commands improved
    * added VPP command to manage VPP instance
  Additionally, starting with this release the AgentCTL is a VPP-Agent main control tool and the vpp-agent-ctl was definitely removed.      

### Improvements
Many end-to-end tests introduced, gradually increasing VPP-Agent stability.
* [if-plugin][vpp-interface-plugin]
  - IP addresses assigned by the DHCP are excluded from the interface address descriptor.
  - VPP-Agent now processes status change notifications labeled by the VPP as UNKNOWN.
* [ns-plugin][linux-ns-plugin]
  - Dockerclient microservice polling replaced with an event listener.   
* [sr-plugin][sr-plugin]
  - SRv6 dynamic proxy routing now can be connected to a non-zero VRF table.   
  
<a name="v2.2.0"></a>
# [2.2.0](https://github.com/ligato/vpp-agent/compare/v2.2.0-beta...v2.2.0) (2019-08-26)

### Compatibility
- **VPP 19.08** (rc1)
- **VPP 19.04** (default)
- **VPP 19.01** (backward compatible)
- cn-infra v2.2

### Bug Fixes
- CN-infra version updated to 2.2 contains a supervisor fix which should prevent the issue where the supervisor logging occasionally caused the agent to crash during large outputs.

### New Features
* [if-plugin][vpp-interface-plugin]
  - Added option to configure SPAN records. Northbound data are formatted by the [SPAN model][span-model].

### Improvements
* [orchestrator][orchestrator-plugin]
  - Clientv2 is now recognized as separate data source by the orchestrator plugin. This feature allows to use the localclient together with other data sources.

### Documentation
- Updated documentation comments in the protobuf API.

<a name="v2.2.0-beta"></a>
# [2.2.0-beta](https://github.com/ligato/vpp-agent/compare/v2.1.1...v2.2.0-beta) (2019-08-09)

### Compatibility
- **VPP 19.08** (rc1)
- **VPP 19.04** (default)
- **VPP 19.01** (backward compatible)

### Bug Fixes
* Fixed SRv6 localsid delete case for non-zero VRF tables.
* Fixed interface IPv6 detection in the descriptor.
* Various bugs fixed in KV scheduler TXN post-processing.
* Interface plugin config names fixed, no stats publishers are now used by default. Instead, datasync is used (by default ETCD, Redis and Consul).
* Rx-placement and rx-mode is now correctly dependent on interface link state.
* Fixed crash for iptables rulechain with default microservice.
* Punt dump fixed in all supported VPP versions.
* Removal of registered punt sockets fixed after a resync.
* Punt socket paths should no longer be unintentionally recreated.
* IP redirect is now correctly dependent on RX interface.
* Fixed IPSec security association configuration for tunnel mode.
* Fixed URL for VPP metrics in telemetry plugin
* Routes are now properly dependent on VRF.

### New Features
* Defined new environment variable `DISABLE_INTERFACE_STATS` to generally disable interface plugin stats.
* Defined new environment variable `RESYNC_TIMEOU` to override default resync timeout. 
* Added [ETCD ansible python plugin][ansible] with example playbook. Consult [readme](ansible/README.md) for more information.

### Improvements
* [govppmux-plugin][govppmux-plugin]
  - GoVPPMux stats can be read with rest under path `/govppmux/stats`.
  - Added disabling of interface stats via the environment variable `DISABLE_INTERFACE_STATS`.
  - Added disabling of interface status publishing via environment variable `DISABLE_STATUS_PUBLISHING`.
* [kv-scheduler][kv-scheduler]
  - Added some more performance improvements.
  - The same key can be no more matched by multiple descriptors.
* [abf-plugin][vpp-abf-plugin]
  - ABF plugin was added to config data model and is now initialized in configurator.  
* [if-plugin][vpp-interface-plugin]
  - Interface rx-placement and rx-mode was enhanced and now allows per-queue configuration.
  - Added [examples](examples/kvscheduler/rxplacement) for rx-placement and rx-mode.
* [nat-plugin][vpp-nat-plugin]
  - NAT example updated for VPP 19.04
* [l3-plugin][vpp-l3-plugin]  
  - Route keys were changed to prevent collisions with some types of configuration. Route with outgoing interface now contains the interface name in the key.
  - Added support for DHCP proxy. A new descriptor allows calling CRUD operations to VPP DHCP proxy servers.
* [punt-plugin][vpp-punt-plugin]
  - Added support for Punt exceptions.  
  - IP redirect dump was implemented for VPP 19.08.
* [Telemetry][vpp-telemetry]
  - Interface metrics added to telemetry plugin. Note that the URL for prometheus export was changed to `/metrics/vpp`.
  - Plugin configuration file now has an option to skip certain metrics.
* [rest-plugin][rest-plugin]
  - Added support for IPSec plugin
  - Added support for punt plugin  
* [agentctl][agentctl]
  - We continuously update the new CTL tool. Various bugs were fixed some new features added.
  - Added new command `import` which can import configuration from file. 

### Docker Images
* The supervisor was replaced with VPP-Agent init plugin. 
* Images now use pre-built VPP images from [ligato/vpp-base](https://github.com/ligato/vpp-base)


<a name="v2.1.1"></a>
# [2.1.1](https://github.com/ligato/vpp-agent/compare/v2.1.0...v2.1.1) (2019-04-05)

### Compatibility
- **VPP 19.04** (`stable/1904`, recommended)
- **VPP 19.01** (backward compatible)

### Bug Fixes
* Fixed IPv6 detection for Linux interfaces [#1355](https://github.com/ligato/vpp-agent/pull/1355).
* Fixed config file names for ifplugin in VPP & Linux [#1341](https://github.com/ligato/vpp-agent/pull/1341).
* Fixed setting status publishers from env var: `VPP_STATUS_PUBLISHERS`.

### Improvements
* The start/stop timeouts for agent can be configured using env vars: `START_TIMEOUT=15s` and `STOP_TIMEOUT=5s`, with values parsed as duration.
* ABF was added to the `ConfigData` message for VPP [#1356](https://github.com/ligato/vpp-agent/pull/1356).
  
### Docker Images
* Images now install all compiled .deb packages from VPP (including `vpp-plugin-dpdk`).

<a name="v2.1.0"></a>
# [2.1.0](https://github.com/ligato/vpp-agent/compare/v2.0.2...v2.1.0) (2019-05-09)

### Compatibility
- **VPP 19.04** (`stable/1904`, recommended)
- **VPP 19.01** (backward compatible)
- cn-infra v2.1
- Go 1.11

The VPP 18.10 was deprecated and is no longer compatible.

### BREAKING CHANGES
* All non-zero VRF tables now must be explicitly created, providing a VRF proto-modeled data to the VPP-Agent. Otherwise, some configuration items will not be created as before (for example interface IP addresses). 

### Bug Fixes
* VPP ARP `retrieve` now also returns IPv6 entries. 

### New Features
* [govppmux-plugin][govppmux-plugin]
  - The GoVPPMux plugin configuration file contains a new option `ConnectViaShm`, which when set to `true` forces connecting to the VPP via shared memory prefix. This is an alternative to environment variable `GOVPPMUX_NOSOCK`.
* [configurator][configurator-plugin]
  - The configurator plugin now collects statistics which are available via the `GetStats()` function or via REST on URL `/stats/configurator`.  
* [kv-scheduler][kv-scheduler]
  - Added transaction statistics.  
* [abf-plugin][vpp-abf-plugin]
  - Added new plugin ABF - ACL-based forwarding, providing an option to configure routing based on matching ACL rules. An ABF entry configures interfaces which will be attached, list of forwarding paths and associated access control list.
* [if-plugin][vpp-interface-plugin]
  - Added support for Generic Segmentation Offload (GSO) for TAP interfaces.  
* [l3-plugin][vpp-l3-plugin]
  - A new model for VRF tables was introduced. Every VRF is defined by an index and an IP version, a new optional label was added. Configuration types using non-zero VRF now require it to be created, since the VRF is considered a dependency. VRFs with zero-index are present in the VPP by default and do not need to be configured (applies for both, IPv4 and IPv6).
* [agentctl][agentctl]
  - This tool becomes obsolete and was completely replaced with a new implementation. Please note that the development of this tool is in the early stages, and functionality is quite limited now. New and improved functionality is planned for the next couple of releases since our goal is to have a single vpp-agent control utility. Because of this, we have also deprecated the vpp-agent-ctl tool which will be most likely removed in the next release.  

### Improvements
* [kv-scheduler][kv-scheduler]  
  - The KV Scheduler received another performance improvements.
* [if-plugin][vpp-interface-plugin]
  - Attempt to configure a Bond interface with already existing ID returns a non-retriable error.
* [linux-if-plugin][linux-interface-plugin]
  - Before adding an IPv6 address to the Linux interface, the plugins will use `sysctl` to ensure the IPv6 is enabled in the target OS.  

### Docker Images
- Supervisord is started as a process with PID 1

### Documentation
- The ligato.io webpage is finally available, check out it [here][ligato.io]! We have also released a [new documentation site][ligato-docs] with a lot of new or updated articles, guides, tutorials and many more. Most of the README.md files scattered across the code were removed or updated and moved to the site.  

<a name="v2.0.2"></a>
# [2.0.2](https://github.com/ligato/vpp-agent/compare/v2.0.1...v2.0.2) (2019-04-19)

### Compatibility
- **VPP 19.01** (updated to `v19.01.1-14-g0f36ef60d`)
- **VPP 18.10** (backward compatible)
- cn-infra v2.0
- Go 1.11

This minor release brought compatibility with updated version of the VPP 19.01.

<a name="v2.0.1"></a>
# [2.0.1](https://github.com/ligato/vpp-agent/compare/v2.0.0...v2.0.1) (2019-04-05)

### Compatibility
- **VPP 19.01** (compatible by default, recommended)
- **VPP 18.10** (backward compatible)
- cn-infra v2.0
- Go 1.11

### Bug Fixes
* Fixed bug where Linux network namespace was not reverted in some cases.
* The VPP socketclient connection checks (and waits) for the socket file in the same manner as for the shared memory, giving 
  the GoVPPMux more time to connect in case the VPP startup is delayed. Also errors occurred during the shm/socket file watch 
  are now properly handled. 
* Fixed wrong dependency for SRv6 end functions referencing VRF tables (DT6,DT4,T).

### Improvements
* [GoVPPMux][govppmux-plugin]
  - Added option to adjust the number of connection attempts and time delay between them. Seek `retry-connect-count` and
  `retry-connect-timeout` fields in [govpp.conf][govppmux-conf]. Also keep in mind the total time in which 
  plugins can be initialized when using these fields. 
* [linux-if-plugin][linux-interface-plugin]
  - Default loopback MTU was set to 65536.
* [ns-plugin][linux-ns-plugin]
  - Plugin descriptor returns `ErrEscapedNetNs` if Linux namespace was changed but not reverted back before returned
  to scheduler. 
  
### Docker Images
* Supervisord process is now started with PID=1

<a name="v2.0.0"></a>
# [2.0.0](https://github.com/ligato/vpp-agent/compare/v1.8...v2.0.0) (2019-04-02)

### Compatibility
- **VPP 19.01** (compatible by default, recommended)
- **VPP 18.10** (backward compatible)
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
  - All vpp-agent models were reviewed and cleaned up. Various changes were done, like simple renaming (in order to have more meaningful fields, avoid duplicated names in types, etc.), improved model convenience (interface type-specific fields are now defined as `oneof`, preventing to set multiple or incorrect data) and other. All models were also moved to the common [api][models] folder.
* [KVScheduler][kv-scheduler]
  - Added new component called KVScheduler, as a reaction to various flaws and issues with race conditions between Vpp/Linux plugins, poor readability and poorly readable logging. Also the system of notifications between plugins was unreliable and hard to debug or even understand. Based on this experience, a new framework offers improved generic mechanisms to handle dependencies between configuration items and creates clean and readable transaction-based logging. Since this component significantly changed the way how plugins are defined, we recommend to learn more about it on the [VPP-Agent wiki page][wiki]. 
* [orchestrator][orchestrator-plugin]
  - The orchestrator is a new component which long-term added value will be a support for multiple northbound data sources (KVDB, GRPC, ...). The current implementation handles combination of GRPC + KVDB, which includes data changes and resync. In the future, any combination of sources will be supported.
* [GoVPPMux][govppmux-plugin]
  - Added `Ping()` method to the VPE vppcalls usable to test the VPP connection.  
* [if-plugin][vpp-interface-plugin]
  - UDP encapsulation can be configured to an IPSec tunnel interface
  - Support for new Bond-type interfaces.
  - Support for L2 tag rewrite (currently present in the interface plugin because of the inconsistent VPP API)
* [nat-plugin][vpp-nat-plugin]      
  - Added support for session affinity in NAT44 static mapping with load balancer.
* [sr-plugin][sr-plugin]
  - Support for Dynamic segment routing proxy with L2 segment routing unaware services.  
  - Added support for SRv6 end function End.DT4 and End.DT6.
* [linux-if-plugin][linux-interface-plugin]
  - Added support for new Linux interface type - loopback.
  - Attempt to assign already existing IP address to the interface does not cause an error.
* [linux-iptables][linux-iptables-plugin]
  - Added new linux IP tables plugin able to configure IP tables chain in the specified table, manage chain rules and set default chain policy.  

### Improvements
* [KVScheduler][kv-scheduler]
  - Performance improvements related to memory management.
* [GoVPPMux][govppmux-plugin]
  - Need for config file was removed, GoVPP is now set with default values if the startup config is not provided.
  - `DefaultReplyTimeout` is now configured globally, instead of set for every request separately.
  - Tolerated default health check timeout is now set to 250ms (up from 100ms). The old value had not provide enough time in some cases.
* [acl-plugin][vpp-acl-plugin]
  - Model moved to the [api/models][models]
* [if-plugin][vpp-interface-plugin]
  - Model reviewed, updated and moved to the [api/models][models].
  - Interface plugin now handles IPSec tunnel interfaces (previously done in IPSec plugin).
  - NAT related configuration was moved to its own plugin.
  - New interface stats (added in 1.8.1) use new GoVPP API, and publishing frequency was significantly decreased to handle creation of multiple 
  interfaces in short period of time.
* [IPSec-plugin][vpp-ipsec-plugin]
  - Model moved to the [api/models][models]
  - The IPSec interface is no longer processed by the IPSec plugin (moved to interface plugin).  
  - The ipsec link in interface model now uses the enum definitions from IPSec model. Also some missing crypto algorithms were added.
* [l2-plugin][vpp-l2-plugin]
   - Model moved to the [api/models][models] and split to three separate models for bridge domains, FIBs and cross connects.  
* [l3-plugin][vpp-l3-plugin]
   - Model moved to the [api/models][models] and split to three separate models for ARPs, Proxy ARPs including IP neighbor and Routes.
* [nat-plugin][vpp-nat-plugin]
  - Defined new plugin to handle NAT-related configuration and its own [model][nat-proto] (before a part of interface plugin).  
* [punt-plugin][vpp-punt-plugin]
  - Model moved to the [api/models][models].
  - Added retrieve support for punt socket. The current implementation is not final - plugin uses local cache (it will be enhanced when the appropriate VPP binary API call will be added).
* [stn-plugin][vpp-stn-plugin]
  - Model moved to the [api/models][models].
* [linux-if-plugin][linux-interface-plugin]
  - Model reviewed, updated and moved to the [api/models][models].  
* [linux-l3-plugin][linux-l3-plugin]
  - Model moved to the [api/models][models] and split to separate models for ARPs and Routes.
  - Linux routes and ARPs have a new dependency - the target interface is required to contain an IP address. 
* [ns-plugin][ns-plugin]
  - New auxiliary plugin to handle linux namespaces and microservices (evolved from ns-handler). Also defines [model][ns-proto] for generic linux namespace definition. 

### Docker Images
* Configuration file for GoVPP was removed, forcing to use default values (which are the same as they were in the file).
* Fixes for installing ARM64 debugger.
* Kafka is no longer required in order to run vpp-agent from the image.

### Documentation
* Added documentation for the punt plugin, describing main features and usage of the punt plugin.
* Added documentation for the [IPSec plugin][vpp-ipsec-plugin], describing main and usage of the IPSec plugin.
* Added documentation for the [interface plugin][vpp-interface-plugin]. The document is only available on [wiki page][wiki].
* Description improved in various proto files.
* Added a lot of new documentation for the KVScheduler (examples, troubleshooting, debugging guides, diagrams, ...)
* Added tutorial for KV Scheduler.
* Added many new documentation articles to the [wiki page][wiki]. However, most of is there only temporary
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
- [vpp-ifplugin][vpp-interface-plugin]
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
- [vpp-ifplugin][vpp-interface-plugin]
  * Rx-mode and Rx-placement now support dump via the respective binary API call
- vpp-rpc-plugin
  * GRPC now supports also IPSec configuration.
  * All currently supported configuration items can be also dumped/read via GRPC (similar to rest) 
  * GRPC now allows to automatically persist configuration to the data store. The desired DB has to be defined in the new GRPC config file (see [readme][readme] for additional information).
- [vpp-punt][punt-model]
  * Added simple new punt plugin. The plugin allows to register/unregister punt to host via Unix domain socket. The new [model][punt-model] was added for this configuration type. Since the VPP API is incomplete, the configuration does not support dump.  
  
### Improvements 
- [vpp-ifplugin][vpp-interface-plugin]
  * The VxLAN interface now support IPv4/IPv6 virtual routing and forwarding (VRF tables). 
  * Support for new interface type: VmxNet3. The VmxNet3 virtual network adapter has no physical counterpart since it is optimized for performance in a virtual machine. Because built-in drivers for this card are not provided by default in the OS, the user must install VMware Tools. The interface model was updated for the VmxNet3 specific configuration.
- [ipsec-plugin][vpp-ipsec-plugin]
  * IPSec resync processing for security policy databases (SPD) and security associations (SA) was improved. Data are properly read from northbound and southbound, compared and partially configured/removed, instead of complete cleanup and re-configuration. This does not appeal to IPSec tunnel interfaces.
  * IPSec tunnel can be now set as an unnumbered interface.
- [rest-plugin][rest-plugin]
  * In case of error, the output returns correct error code with cause (parsed from JSON) instead of an empty body  


<a name="v1.7.0"></a>
# [1.7.0](https://github.com/ligato/vpp-agent/compare/v1.6...v1.7) (2018-10-02)

### Compatibility
- VPP 18.10-rc0~505-ge23edac
- cn-infra v1.6
- Go 1.11
  
### Bug Fixes
  * Corrected several cases where various errors were silently ignored
  * GRPC registration is now done in Init() phase, ensuring that it finishes before GRPC server is started
  * Removed occasional cases where Linux tap interface was not configured correctly
  * Fixed FIB configuration failures caused by wrong updating of the metadata after several modifications
  * No additional characters are added to NAT tag and can be now configured with the full length without index out of range errors
  * Linux interface resync registers all VETH-type interfaces, despite the peer is not known
  * Status publishing to ETCD/Consul now should work properly
  * Fixed occasional failure caused by concurrent map access inside Linux plugin interface configurator
  * VPP route dump now correctly recognizes route type

### Features
- [vpp-ifplugin][vpp-interface-plugin]
  * It is now possible to dump unnumbered interface data
  * Rx-placement now uses specific binary API to configure instead of generic CLI API
- [vpp-l2plugin][vpp-l2-plugin]
  * Bridge domain ARP termination table can now be dumped  
- [linux-ifplugin][linux-interface-plugin]
  * Linux interface watcher was reintroduced.
  * Linux interfaces can be now dumped.
- [linux-l3plugin][linux-l3-plugin]
  * Linux ARP entries and routes can be dumped.   

### Improvements
- [vpp-plugins][vpp-plugins]
  * Improved error propagation in all the VPP plugins. Majority of errors now print the stack trace to the log output allowing better error tracing and debugging.
  * Stopwatch was removed from all vppcalls
- [linux-plugins][linux-plugins]
  * Improved error propagation in all Linux plugins (same way as for VPP)
  * Stopwatch was removed from all linuxcalls
- [govpp-plugn][govppmux-plugin]
  * Tracer (introduced in cn-infra 1.6) added to VPP message processing, replacing stopwatch. The measurement should be more precise and logged for all binary API calls. Also the rest plugin now allows showing traced entries. 
  
### Docker Images
  * The image can now be built on ARM64 platform   


<a name="v1.6.0"></a>
# [1.6.0](https://github.com/ligato/vpp-agent/compare/v1.5.2...v1.6) (2018-08-24)

### Compatibility
- VPP 18.10-rc0~169-gb11f903a
- cn-infra v1.5
  
### BREAKING CHANGES
- Flavors were replaced with new way of managing plugins.
- REST interface URLs were changed, see [readme][readme] for complete list.

### Bug Fixes
* if VPP routes are dumped, all paths are returned
* NAT load-balanced static mappings should be resynced correctly  
* telemetry plugin now correctly parses parentheses for `show node counters`
* telemetry plugin will not hide an error caused by value loading if the config file is not present
* Linux plugin namespace handler now correctly handles namespace switching for interfaces with IPv6 addresses. Default IPv6 address (link local) will not be moved to the new namespace if there are no more IPv6 addresses configured within the interface. This should prevent failures in some cases where IPv6 is not enabled in the destination namespace.
* VxLAN with non-zero VRF can be successfully removed
* Lint is now working again    
* VPP route resync works correctly if next hop IP address is not defined

### Features
* Deprecating flavors
  - CN-infra 1.5 brought new replacement for flavors and it would be a shame not to implement it in the vpp-agent. The old flavors package was removed and replaced with this new concept, visible in app package vpp-agent.
- [rest plugin][rest-plugin]
  * All VPP configuration types are now supported to be dumped using REST. The output consists of two parts; data formatted as NB proto model, and metadata with VPP specific configuration (interface indexes, different counters, etc.).
  * REST prefix was changed. The new URL now contains API version and purpose (dump, put). The list of all URLs can be found in the [readme][readme]
- [ifplugin][vpp-interface-plugin]
  * Added support for NAT virtual reassembly for both, IPv4 and IPv6. See change in 
    [nat proto file][nat-proto]
- [l3plugin][vpp-l3-plugin]
  * Vpp-agent now knows about DROP-type routes. They can be configured and also dumped. VPP default routes, which are DROP-type is recognized and registered. Currently, resync does not remove or correlate such a route type automatically, so no default routes are unintentionally removed.
  * New configurator for L3 IP scan neighbor was added, allowing to set/unset IP scan neigh parameters to the VPP.
  
### Improvements
- [vpp plugins][vpp-plugins]
  * all vppcalls were unified under API defined for every configuration type (e.g. interfaces, l2, l3, ...). Configurators now use special handler object to access vppcalls. This should prevent duplicates and make vppcalls cleaner and more understandable.
- [ifplugin][vpp-interface-plugin]
  * VPP interface DHCP configuration can now be dumped and added to resync processing
  * Interfaces and also L3 routes can be configured for non-zero VRF table if IPv6 is used. 
- [examples][examples]
  * All examples were reworked to use new flavors concept. The purpose was not changed.    

### Docker Images
- using Ubuntu 18.04 as the base image


<a name="v1.5.2"></a>
## [1.5.2](https://github.com/ligato/vpp-agent/compare/v1.5.1...v1.5.2) (2018-07-23)

### Compatibility
- VPP 18.07-rc0~358-ga5ee900
- cn-infra v1.4.1 (minor version fixes bug in Consul)

### Bug Fixes
- [Telemetry][vpp-telemetry]
  * Fixed bug where lack of config file could cause continuous polling. The interval now also cannot be changed to a value less than 5 seconds.
  * Telemetry plugin is now closed properly


<a name="v1.5.1"></a>
## 1.5.1 (2018-07-20)

### Compatibility
- VPP 18.07-rc0~358-ga5ee900
- cn-infra v1.4

### Features
- [Telemetry][vpp-telemetry]
  * Default polling interval was raised to 30s.
  * Added option to use telemetry config file to change polling interval, or turn the polling off, disabling the telemetry plugin. The change was added due to several reports where often polling is suspicious of interrupting VPP worker threads and causing packet drops and/or other negative impacts. More information how to use the config file can be found in the [readme][readme].


<a name="v1.5.0"></a>
# [1.5.0](https://github.com/ligato/vpp-agent/compare/v1.4.1...v1.5) (2018-07-16)

### Compatibility
- VPP 18.07-rc0~358-ga5ee900
- cn-infra v1.4

### BREAKING CHANGES
- The package `etcdv3` was renamed to `etcd`, along with its flag and configuration file.
- The package `defaultplugins` was renamed to `vpp` to make the purpose of the package clear

### Bug Fixes
- Fixed a few issues with parsing VPP metrics from CLI for [Telemetry][vpp-telemetry].
- Fixed bug in GoVPP occurring after some request timed out, causing the channel to receive replies from the previous request and always returning an error.
- Fixed issue which prevented setting interface to non-existing VRF.
- Fixed bug where removal of an af-packet interface caused attached Veth to go DOWN.
- Fixed NAT44 address pool resolution which was not correct in some cases.
- Fixed bug with adding SR policies causing incomplete configuration.

### Features
- [LinuxPlugin][linux-interface-plugin]
  * Is now optional and can be disabled via configuration file.
- [ifplugin][vpp-interface-plugin]
  * Added support for VxLAN multicast
  * Rx-placement can be configured on VPP interfaces
- [IPsec][vpp-ipsec-plugin]
  * IPsec UDP encapsulation can now be set (NAT traversal)  

### Docker Images
- Replace `START_AGENT` with `OMIT_AGENT` to match `RETAIN_SUPERVISOR` and keep both unset by default.
- Refactored and cleaned up execute scripts and remove unused scripts.
- Fixed some issues with `RETAIN_SUPERVISOR` option.
- Location of supervisord pid file is now explicitly set to
  `/run/supervisord.pid` in *supervisord.conf* file.
- The vpp-agent is now started  with single flag `--config-dir=/opt/vpp-agent/dev`, and will automatically load all configuration from that directory.


<a name="v1.4.1"></a>
## [1.4.1](https://github.com/ligato/vpp-agent/compare/v1.4.0...v1.4.1) (2018-06-11)

A minor release using newer VPP v18.04 version.

### Compatibility
- VPP v18.04 (2302d0d)
- cn-infra v1.3

### Bug Fixes
- VPP submodule was removed from the project. It should prevent various problems with dependency resolution.
- Fixed known bug present in the previous version of the VPP, issued as [VPP-1280][vpp-issue-1280]. Current version contains appropriate fix.  


<a name="v1.4.0"></a>
# [1.4.0](https://github.com/ligato/vpp-agent/compare/v1.3...v1.4.0) (2018-05-24)

### Compatibility
- VPP v18.04 (ac2b736)
- cn-infra v1.3

### Bug Fixes
  * Fixed case where the creation of the Linux route with unreachable gateway threw an error. The route is now appropriately cached and created when possible. 
  * Fixed issue with GoVPP channels returning errors after a timeout.
  * Fixed various issues related to caching and resync in L2 cross-connect
  * Split horizon group is now correctly assigned if an interface is created after bridge domain
  * Fixed issue where the creation of FIB while the interface was not a part of the bridge domain returned an error. 

### Known issues
  * VPP crash may occur if there is interface with non-default VRF (>0). There is an [VPP-1280][vpp-issue-1280] issue created with more details 

### Features
- [Consul][consul]
  * Consul is now supported as a key-value store alternative to ETCD. More information in the [readme][readme].
- [Telemetry][vpp-telemetry]
  * New plugin for collecting telemetry data about VPP metrics and serving them via HTTP server for Prometheus. More information in the [readme][readme].
- [Ipsecplugin][vpp-ipsec-plugin]
  * Now supports tunnel interface for encrypting all the data passing through that interface.
- GRPC 
  * Vpp-agent itself can act as a GRPC server (no need for external executable)
  * All configuration types are supported (incl. Linux interfaces, routes and ARP)
  * Client can read VPP notifications via vpp-agent.
- [SR plugin][sr-plugin]
  * New plugin with support for Segment Routing.
    More information in the [readme][readme].

### Improvements
- [ifplugin][vpp-interface-plugin]
  * Added support for self-twice-NAT
- __vpp-agent-grpc__ executable merged with [vpp-agent][vpp-agent] command.
- [govppmux][govppmux-plugin]
  * `configure reply timeout` can be configured.
  * Support for VPP started with custom shared memory prefix. SHM may be configured via the GoVPP plugin config file. More info in the [readme][readme]
  * Overall redundancy cleanup and corrected naming for all proto models.
  * Added more unit tests for increased coverage and code stability. 

### Documentation
- [localclient_linux][examples-vpp-local] now contains two examples, the old one demonstrating basic plugin functionality was moved to plugin package, and specialised example for [NAT][examples-nat] was added.
- [localclient_linux][examples-linux-local] now contains two examples, the old one demonstrating 
  [veth][examples-veth] interface usage was moved to package and new example for linux
  [tap][examples-tap] was added.


<a name="v1.3.0"></a>
# [1.3.0](https://github.com/ligato/vpp-agent/compare/v1.2...v1.3) (2018-03-22)

The vpp-agent is now using custom VPP branch [stable-1801-contiv][contiv-vpp1810].

### Compatibility
- VPP v18.01-rc0~605-g954d437
- cn-infra v1.2

### Bug Fixes
  * Resync of ifplugin in both, VPP and Linux, was improved. Interfaces with the same configuration data are not recreated during resync.
  * STN does not fail if IP address with a mask is provided.
  * Fixed ingress/egress interface resolution in ACL.
  * Linux routes now check network reachability for gateway address before configuration. It should prevent "network unreachable" errors during config.
  * Corrected bridge domain crash in case non-bvi interface was added to another non-bvi interface.
  * Fixed several bugs related to VETH and AF-PACKET configuration and resync.

### Features
- [ipsecplugin][vpp-ipsec-plugin]:
  * New plugin for IPSec added. The IPSec is supported for VPP only with Linux set manually for now. IKEv2 is not yet supported. More information in the [readme][readme].
- [nsplugin][linux-ns-plugin]
  * New namespace plugin added. The configurator handles common namespace and microservice processing and communication with other Linux plugins.
- [ifplugin][vpp-interface-plugin]
  * Added support for Network address translation. NAT plugin supports a configuration of NAT44 interfaces, address pools and DNAT. More information in the [readme][readme].
  * DHCP can now be configured for the interface  
- [l2plugin][vpp-l2-plugin]
  * Split-horizon group can be configured for bridge domain interface.
- [l3plugin][vpp-l3-plugin]
  * Added support for proxy ARP. For more information and configuration example, please see [readme][readme].
- [linux ifplugin][linux-interface-plugin]
  * Support for automatic interface configuration (currently only TAP).
        
### Improvements
- [aclplugin][agentctl]
  * Removed configuration order of interfaces. The access list can be now configured even if interfaces do not exist yet, and add them later.
- vpp-agent-ctl
  * The vpp-agent-ctl was refactored and command info was updated.

### Docker Images
  * VPP can be built and run in the release or debug mode. Read more information in the [readme][readme].
  * Production image is now smaller by roughly 40% (229MB).


<a name="v1.2.0"></a>
# [1.2.0](https://github.com/ligato/vpp-agent/compare/v1.1...v1.2) (2018-02-07)

### Compatibility
- VPP v18.04-rc0~90-gd95c39e
- cn-infra v1.1

### Bug Fixes
- Fixed interface assignment in ACLs
- Fixed bridge domain BVI modification resolution
- vpp-agent-grpc (removed in 1.4 release, since then it is a part of the vpp-agent) now compiles properly together with other commands.

### Known Issues
- VPP can occasionally cause a deadlock during checksum calculation (https://jira.fd.io/browse/VPP-1134)
- VPP-Agent might not properly handle initialization across plugins (this is not occurring currently, but needs to be tested more)

### Improvements
- [aclplugin][vpp-acl-plugin]
  * Improved resync of ACL entries. Every new ACL entry is correctly configured in the VPP and all obsolete entries are read and removed. 
- [ifplugin][vpp-interface-plugin]
  * Improved resync of interfaces, BFD sessions, authentication keys, echo functions and STN. Better resolution of persistence config for interfaces. 
- [l2plugin][vpp-l2-plugin]
  * Improved resync of bridge domains, FIB entries, and xConnect pairs. Resync now better correlates configuration present on the VPP with the NB setup.
- [linux-ifplugin][linux-interface-plugin]
  * ARP does not need the interface to be present on the VPP. Configuration is cached and put to the VPP if requirements are fulfilled. 
- Dependencies
  * Migrated from glide to dep  

### Docker Images
  * VPP compilation now skips building of Java/C++ APIs, this saves build time and final image size.
  * Development image now runs VPP in debug mode with various debug options added in [VPP config file][vpp-conf-file].


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
- [ifplugin][vpp-interface-plugin]
    - added support for un-numbered interfaces. The nterface can be marked as un-numbered with information about another interface containing required IP address. A un-numbered interface does not need to have IP address set.
    - added support for virtio-based TAPv2 interfaces.
    - interface status is no longer stored in the ETCD by default and it can be turned on using the appropriate setting in vpp-plugin.conf. See  [readme][readme] for more details.  
- [l2plugin][vpp-l2-plugin]
    - bridge domain status is no longer stored in the ETCD by default and it can be turned on using the appropriate setting in vpp-plugin.conf. See  [readme][readme] for more details.  

### Improvements
- [ifplugin][vpp-interface-plugin]
    - default MTU value was removed in order to be able to just pass empty MTU field. MTU now can be set only in interface configuration (preferred) or defined in vpp-plugin.conf. If none of them is set, MTU value will be empty.
    - interface state data are stored in statuscheck readiness probe
- [l3plugin][vpp-l3-plugin]
    - removed strict configuration order for VPP ARP entries and routes. Both ARP entry or route can be configured without interface already present.
- l4plugin (removed in v2.0)
   - removed strict configuration order for application namespaces. Application namespace can be configured without interface already present.
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
- [ifplugin][vpp-interface-plugin]
   - ability to configure STN rules.  See respective
   [readme][readme] in interface plugin for more details.
   - rx-mode settings can be set on interface. Ethernet-type interface can be set to POLLING mode, other types of interfaces supports also INTERRUPT and ADAPTIVE. Fields to set QueueID/QueueIDValid are also available
   - added possibility to add interface to any VRF table.
   - added defaultplugins API.
   - API contains new Method `DisableResync(keyPrefix ...string)`. One or more ETCD key prefixes can be used as a parameter to disable resync for that specific key(s).
- l4plugin (removed in v2.0)
   - added new l4 plugin to the VPP plugins. It can be used to enable/disable L4 features
   and configure application namespaces. See respective
    [readme][readme] in L4 plugin for more details.
   - support for VPP plugins/l3plugin ARP configuration. The configurator can perform the
   basic CRUD operation with ARP config.
- resync
  - resync error propagation improved. If any resynced configuration fails, rest of the resync completes and will not be interrupted. All errors which appear during resync are logged after. 
- [linux l3plugin][linux-l3-plugin]
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

- [Default VPP plugin][vpp-interface-plugin]
    - added resync strategies. Resync of VPP plugins can be set using defaultpluigns config file; Resync can be set to full (always resync everything) or dependent on VPP configuration (if there is none, skip resync). Resync can be also forced to skip using the parameter.
- [Linuxplugins L3Plugin][linux-l3-plugin]
    - added support for basic CRUD operations with the static Address resolution protocol entries and static Routes.


<a name="v1.0.6"></a>
## [1.0.6](https://github.com/ligato/vpp-agent/compare/v1.0.5...v1.0.6) (2017-10-17)

### Compatibility
- cn-infra v1.0.5

### Features

- [LinuxPlugin][linux-interface-plugin]
   - The configuration of vEth interfaces modified. Veth configuration defines two names: symbolic used internally and the one used in host OS. `HostIfName` field is optional. If it is not defined, the name in the host OS will be the same as the symbolic one - defined by `Name` field.


<a name="v1.0.5"></a>
## [1.0.5](https://github.com/ligato/vpp-agent/compare/v1.0.4...v1.0.5) (2017-09-26)

### Compatibility
- VPP version v17.10-rc0~334-gce41a5c
- cn-infra v1.0.4

### Features

- [GoVppMux][govppmux-plugin]
    - configuration file for govpp added
- Kafka Partitions
    - Changes in offset handling, only automatically partitioned messages (hash, random)
      have their offset marked. Manually partitioned messages are not marked.
    - Implemented post-init consumer (for manual partitioner only) which allows starting
      consuming after kafka-plugin Init()
    - Minimalistic examples & documentation for Kafka API will be improved in a later release.


<a name="v1.0.4"></a>
## [1.0.4](https://github.com/ligato/vpp-agent/compare/v1.0.3...v1.0.4) (2017-09-08)

### Features

- Kafka Partitions
    - Implemented new methods that allow to specify partitions & offset parameters:
      * publish: Mux.NewSyncPublisherToPartition() & Mux.NewAsyncPublisherToPartition()
      * watch: ProtoWatcher.WatchPartition()
    - Minimalistic examples & documentation for Kafka API will be improved in a later release.
- Flavors
    - reduced to only local.FlavorVppLocal & vpp.Flavor
- GoVPP
    - updated version waits until the VPP is ready to accept a new connection


<a name="v1.0.3"></a>
## [1.0.3](https://github.com/ligato/vpp-agent/compare/v1.0.2...v1.0.3) (2017-09-05)

### Compatibility
- VPP version v17.10-rc0~265-g809bc74 (upgraded because of VPP MEMIF fixes)

### Features

Enabled support for wathing data store `OfDifferentAgent()` - see:
* examples/idx_iface_cache (removed in v2.0)
* examples/examples/idx_bd_cache (removed in v2.0)
* examples/idx_veth_cache (removed in v2.0)

Preview of new Kafka client API methods that allows to fill also partition and offset argument. New methods implementation ignores these new parameters for now (fallback to existing implementation based on `github.com/bsm/sarama-cluster` and `github.com/Shopify/sarama`).


<a name="v1.0.2"></a>
## [1.0.2](https://github.com/ligato/vpp-agent/compare/v1.0.1...v1.0.2) (2017-08-28)

### Compatibility
- VPP version v17.10-rc0~203

### Known Issues
A rarely occurring problem during startup with binary API connectivity. VPP rejects binary API connectivity when VPP Agent tries to connect too early (plan fix this behavior in next release).

### Features

Algorithms for applying northbound configuration (stored in ETCD key-value data store)
to VPP in the proper order of VPP binary API calls implemented in [Default VPP plugin][vpp-interface-plugin]:
- network interfaces, especially:
  - MEMIFs (optimized data plane network interface tailored for a container to container network connectivity)
  - VETHs (standard Linux Virtual Ethernet network interface)
  - AF_Packets (for accessing VETHs and similar type of interface)
  - VXLANs, Physical Network Interfaces, loopbacks ...
- L2 BD & X-Connects
- L3 IP Routes & VRFs
- ACL (Access Control List)

Support for Linux VETH northbound configuration implemented in [Linux Plugin][linux-interface-plugin]
applied in proper order with VPP AF_Packet configuration.

Data Synchronization during startup for network interfaces & L2 BD
(support for the situation when ETCD contain configuration before VPP Agent starts).

Data replication and events:
- Updating operational data in ETCD (VPP indexes such as  sw_if_index) and statistics (port counters).
- Updating statistics in Redis (optional once redis.conf available - see flags).
- Publishing links up/down events to Kafka message bus.

- [Examples][examples]
- Tools:
  - [agentctl CLI tool][agentctl] that show state & configuration of VPP agents
  - [docker][docker]: container-based development environment for the VPP agent
- other features inherited from cn-infra:
  - health: status check & k8s HTTP/REST probes
  - logging: changing log level at runtime
- Ability to extend the behavior of the VPP Agent by creating new plugins on top of VPP Agent flavor (removed with CN-Infra v1.5).
  New plugins can access API for configured:
  - VPP Network interfaces,
  - Bridge domains and VETHs
    based on [idxvpp][idx-vpp] threadsafe map tailored for VPP data
    with advanced features (multiple watchers, secondary indexes).
- VPP Agent is embeddable in different software projects and with different systems by using Local Flavor (removed with CN-Infra v1.5) to reuse VPP Agent algorithms. For doing this there is VPP Agent client version 1 (removed in v2.0):
  - local client - for embedded VPP Agent (communication inside one operating system process, VPP Agent effectively used as a library)
  - remote client - for remote configuration of VPP Agent (while integrating for example with control plane)

[agentctl]: cmd/agentctl
[ansible]: ansible
[configurator-plugin]: plugins/configurator
[consul]: https://www.consul.io/
[contiv-vpp1810]: https://github.com/vpp-dev/vpp/tree/stable-1801-contiv
[docker]: docker
[examples]: examples
[examples-linux-local]: examples/localclient_linux
[examples-nat]: examples/localclient_vpp/nat
[examples-tap]: examples/localclient_linux/tap
[examples-veth]: examples/localclient_linux/veth
[examples-vpp-local]: examples/localclient_vpp
[govppmux-plugin]: plugins/govppmux
[govppmux-conf]: plugins/govppmux/govpp.conf
[idx-vpp]: pkg/idxvpp
[kv-scheduler]: plugins/kvscheduler
[ligato.io]: https://ligato.io/
[ligato-docs]: https://docs.ligato.io/en/latest/
[linux-interface-plugin]: plugins/linux/ifplugin
[linux-iptables-plugin]: plugins/linux/iptablesplugin
[linux-l3-plugin]: plugins/linux/l3plugin
[linux-ns-plugin]: plugins/linux/nsplugin
[linux-plugins]: plugins/linux
[nat-proto]: api/models/vpp/nat/nat.proto
[netalloc-plugin]: plugins/netalloc
[netalloc-plugin-model]: api/models/netalloc/netalloc.proto
[ns-plugin]: plugins/linux/nsplugin
[ns-proto]: api/models/linux/namespace/namespace.proto
[models]: api/models
[orchestrator-plugin]: plugins/orchestrator
[punt-model]: api/models/vpp/punt/punt.proto
[readme]: README.md
[rest-plugin]: plugins/restapi
[span-model]: api/models/vpp/interfaces/span.proto
[sr-plugin]: plugins/vpp/srplugin
[vpp-abf-plugin]: plugins/vpp/abfplugin
[vpp-acl-plugin]: plugins/vpp/aclplugin
[vpp-agent]: cmd/vpp-agent
[vpp-conf-file]: docker/dev/vpp.conf
[vpp-interface-plugin]: plugins/vpp/ifplugin
[vpp-issue-1280]: https://jira.fd.io/browse/VPP-1280
[vpp-ipsec-plugin]: plugins/vpp/ipsecplugin
[vpp-l2-plugin]: plugins/vpp/l2plugin
[vpp-l3-plugin]: plugins/vpp/l3plugin
[vpp-nat-plugin]: plugins/vpp/natplugin
[vpp-plugins]: plugins/vpp
[vpp-punt-plugin]: plugins/vpp/puntplugin
[vpp-stn-plugin]: plugins/vpp/stnplugin
[vpp-telemetry]: plugins/telemetry
[wiki]: https://github.com/ligato/vpp-agent/wiki
