Welcome to the vpp-agent user guide.

## The vpp-agent basics
- Get the general [overview](https://github.com/ligato/vpp-agent/wiki/Overview) of to learn what the vpp-agent is, what it can do and what components is it made of.
- Begin with the [quickstart guide](articles/Quickstart.md) how to download, install and run the vpp-agent. 
- [Learn more about how to](https://github.com/ligato/vpp-agent/wiki/Learn-how-to) prepare the environment for the vpp-agent and how to use it for the configuration.
- A [list of all keys](https://github.com/ligato/vpp-agent/wiki/KeyOverview) supported

## The vpp-agent concepts
- The plugin lifecycle.
- [Key-value datastore](https://github.com/ligato/vpp-agent/wiki/KV-Store).

## Plugins and components
- The vpp-agent plugins managing the VPP configuration:
  * [Access Lists](https://github.com/ligato/vpp-agent/wiki/ACL-plugin)
  * [Interfaces](https://github.com/ligato/vpp-agent/wiki/VPP-Interface-plugin)
  * [IPSec](https://github.com/ligato/vpp-agent/wiki/IPSec-plugin)
  * [L2 plugin](https://github.com/ligato/vpp-agent/wiki/L2-plugin)
  * [L3 plugin](https://github.com/ligato/vpp-agent/wiki/L3-plugin)
  * [NAT plugin](https://github.com/ligato/vpp-agent/wiki/NAT-plugin)
  * [Punt](https://github.com/ligato/vpp-agent/wiki/Punt-plugin)
  * STN plugin
- Plugins managing the Linux configuration:
  * [Linux Interfaces](https://github.com/ligato/vpp-agent/wiki/Linux-Interface-plugin)
  * [Linux L3 plugin](https://github.com/ligato/vpp-agent/wiki/Linux-L3)
  * [Namespaces](https://github.com/ligato/vpp-agent/wiki/Namespace-plugin)
- Other vpp-agent plugins:
  * Providing northbound access:
    - [Clientv2](https://github.com/ligato/vpp-agent/wiki/Clientv2)
    - [REST API](https://github.com/ligato/vpp-agent/wiki/REST)
    - [GRPC](https://github.com/ligato/vpp-agent/wiki/GRPC)
  * Managing the data flow and synchronization:
    - Configurator  
    - Orchestrator
    - [KV Scheduler](https://github.com/ligato/vpp-agent/wiki/KVScheduler)
      * [Implement your own KV Descriptor](https://github.com/ligato/vpp-agent/wiki/Implementing-your-own-KVDescriptor) // more fitting to the development guide
  * Providing the VPP connection:
    - [GoVPP mux](https://github.com/ligato/vpp-agent/wiki/Govppmux)
  * Collecting and exporting the VPP statistics:
    - [Telemetry](https://github.com/ligato/vpp-agent/wiki/Telemetry)

## Utilities and examples    
- The [VPP-Agent-ctl](https://github.com/ligato/vpp-agent/blob/master/cmd/vpp-agent-ctl/README.md) to test the vpp-agent with pre-prepared configuration data.
- [Examples](https://github.com/ligato/vpp-agent/blob/master/examples/README.md) for various vpp-agent features.        