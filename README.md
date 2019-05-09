# VPP Agent

![GitHub contributors](https://img.shields.io/github/contributors/ligato/vpp-agent.svg)
[![Build Status](https://travis-ci.org/ligato/vpp-agent.svg?branch=master)](https://travis-ci.org/ligato/vpp-agent)
[![Coverage Status](https://coveralls.io/repos/github/ligato/vpp-agent/badge.svg?branch=master)](https://coveralls.io/github/ligato/vpp-agent?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ligato/vpp-agent)](https://goreportcard.com/report/github.com/ligato/vpp-agent)
[![GoDoc](https://godoc.org/github.com/ligato/vpp-agent?status.svg)](https://godoc.org/github.com/ligato/vpp-agent)
[![GitHub license](https://img.shields.io/badge/license-Apache%20license%202.0-blue.svg)](https://github.com/ligato/vpp-agent/blob/master/LICENSE)

###### Please note that the content of this repository is currently **WORK IN PROGRESS**!

The VPP Agent is a Go implementation of a control/management plane for [VPP][vpp] based
cloud-native [Virtual Network Functions][vnf] (VNFs). The VPP Agent is built on top of 
[CN Infra][cn-infra], a framework for developing cloud-native VNFs (CNFs).

The VPP Agent can be used as-is as a management/control agent for VNFs  based on off-the-shelf
VPP (e.g. a VPP-based vswitch), or as a framework for developing management agents for VPP-based
CNFs. An example of a custom VPP-based CNF is the [Contiv-VPP][contiv-vpp] vswitch.

### Releases

|Release|Release Date|Info|
|---|:---:|---|
|[![stable](https://img.shields.io/github/release/ligato/vpp-agent.svg?label=release&logo=github)](https://github.com/ligato/vpp-agent/releases/latest)|![Release date](https://img.shields.io/github/release-date/ligato/vpp-agent.svg?label=)|latest release|
|[![latest](https://img.shields.io/github/release-pre/ligato/vpp-agent.svg?label=release&logo=github)](https://github.com/ligato/vpp-agent/releases)|![Release date](https://img.shields.io/github/release-date-pre/ligato/vpp-agent.svg?label=)|last release/pre-release|

Have a look at the [release notes](CHANGELOG.md) for a complete list of changes.

### Branches

|Branch|Last Commit|Info|
|---|:---:|---|
|[![master](https://img.shields.io/badge/branch-master-blue.svg?logo=git&logoColor=white)](https://github.com/ligato/vpp-agent/tree/master)|![GitHub last commit (branch)](https://img.shields.io/github/last-commit/ligato/vpp-agent/master.svg?label=)| has **moved to v2**, introducing several [breaking changes](https://github.com/ligato/vpp-agent/blob/master/CHANGELOG.md#v200) :warning:|
|[![dev](https://img.shields.io/badge/branch-dev-green.svg?logo=git&logoColor=white)](https://github.com/ligato/vpp-agent/tree/dev)|![GitHub last commit (branch)](https://img.shields.io/github/last-commit/ligato/vpp-agent/dev.svg?label=)|will be used for all the future **development**|
|[![pantheon-dev](https://img.shields.io/badge/branch-pantheon--dev-inactive.svg?logo=git&logoColor=white)](https://github.com/ligato/vpp-agent/tree/pantheon-dev)|![GitHub last commit (branch)](https://img.shields.io/github/last-commit/ligato/vpp-agent/pantheon-dev.svg?label=)|has been **deprecated** (v1) and will be removed in the following weeks|

### Images

|Image|Image Size/Layers|Info|
|---|:---:|---|
|[![ligato/vpp-agent](https://img.shields.io/badge/image-ligato/vpp--agent-blue.svg?logo=docker&logoColor=white)](https://cloud.docker.com/u/ligato/repository/docker/ligato/vpp-agent)|![MicroBadger Size](https://img.shields.io/microbadger/image-size/ligato/vpp-agent.svg) ![MicroBadger Layers](https://img.shields.io/microbadger/layers/ligato/vpp-agent.svg)|minimal image for production|
|[![ligato/dev-vpp-agent](https://img.shields.io/badge/image-ligato/dev--vpp--agent-blue.svg?logo=docker&logoColor=white)](https://cloud.docker.com/u/ligato/repository/docker/ligato/dev-vpp-agent)|![MicroBadger Size](https://img.shields.io/microbadger/image-size/ligato/dev-vpp-agent.svg) ![MicroBadger Layers](https://img.shields.io/microbadger/layers/ligato/dev-vpp-agent.svg)|image prepared for developers|

The image tag `latest` is built from **master branch** and `dev` tag is built from **dev branch**.

## Quickstart

For a quick start with the VPP Agent, you can use the pre-built Docker images on DockerHub
that contain the VPP Agent and VPP: [ligato/vpp-agent][vpp-agent] (or for ARM64: [ligato/vpp-agent-arm64][vpp-agent-arm64]).

0. Start ETCD (for image versions lower than 2.0, the Kafka is required as well) on your host (e.g. in Docker as described [here][etcd-local]).
   Note: **for ARM64 see the information for [kafka][kafka] and for [etcd][etcd]**.

1. Run VPP + VPP Agent in a Docker container:
```
docker pull ligato/vpp-agent
docker run -it --rm --name vpp --privileged ligato/vpp-agent
```

2. Manage VPP agent using agentctl:
```
docker exec -it vpp agentctl -h
```

3. Check the configuration (using agentctl or directly using VPP console):
```
docker exec -it vpp agentctl -e 172.17.0.1:2379 show
docker exec -it vpp vppctl -s localhost:5002
```

**Next Steps**

See [README][docker-image] of development docker image for more details.

## Documentation

Detailed documentation for the VPP Agent can be found at [ligato.io/vpp-agent][ligato-docs].

## Architecture

The VPP Agent is basically a set of VPP-specific plugins that use the 
CN-Infra framework to interact with other services/microservices in the
cloud (e.g. a KV data store, messaging, log warehouse, etc.). The VPP Agent
exposes VPP functionality to client apps via a higher-level model-driven 
API. Clients that consume this API may be either external (connecting to 
the VPP Agent via REST, gRPC API, Etcd or message bus transport), or local
Apps and/or Extension plugins running on the same CN-Infra framework in the 
same Linux process. 

The VNF Agent architecture is shown in the following figure: 

![vpp agent](docs/imgs/vpp_agent.png "VPP Agent & its Plugins on top of cn-infra")

Each (northbound) VPP API - L2, L3, ACL, ... - is implemented by a specific
VNF Agent plugin, which translates northbound API calls/operations into 
(southbound) low level VPP Binary API calls. Northbound APIs are defined 
using [protobufs][protobufs], which allow for the same functionality to be accessible
over multiple transport protocols (HTTP, gRPC, Etcd, ...). Plugins use the 
[GoVPP library][govpp] to interact with the VPP.

The following figure shows the VPP Agent in context of a cloud-native VNF, 
where the VNF's data plane is implemented using VPP/DPDK and 
its management/control planes are implemented using the VNF agent:

![context](docs/imgs/context.png "VPP Agent & its Plugins on top of cn-infra")

### Plugins
 
The set of plugins in the VPP Agent is as follows:
* [VPP plugins][docs-vpp-punt-plugin] - core plugins providing northbound APIs to _default_ VPP functionality:
  - [ACL][docs-vpp-acl-plugin]: - VPP Access Lists (VPP ACL plugin) 
  - [Interfaces][docs-vpp-interface-plugin] - VPP network interfaces (e.g. DPDK, MEMIF, AF_Packet, VXLAN, Loopback..)
  - [L2][docs-vpp-l2-plugin] - Bridge Domains, L2 cross-connects..
  - [L3][docs-vpp-l3-plugin] - IP Routes, ARPs, ProxyARPs, VRFs..
  - [IPSec][docs-vpp-ipsec-plugin] - Security policy databases and policy associations
  - [Punt][docs-vpp-punt-plugin] - punt to host (directly or via socket), IP redirect
  - [NAT][docs-vpp-nat-plugin] - network address translation configuration, DNAT44
  - [SR][docs-vpp-sr-plugin] - segment routing
* [Linux plugins][docs-linux-plugins] (VETH) - allows optional configuration of Linux virtual ethernet 
  interfaces
  - [Interfaces][docs-linux-interface-plugin] - Linux network interfaces (e.g. VETH, TAP..)
  - [L3][docs-linux-l3-plugin] - IP Routes, ARPs
  - [NS][docs-linux-ns-plugin] - Linux network namespaces
* [GoVPPmux][docs-govppmux-plugin] - plugin wrapper around GoVPP. Multiplexes plugins' access to
  VPP on a single connection.
* [RESTAPI][docs-rest-plugin] - provides API to retrieve actual state
* [KVScheduler][docs-kv-scheduler] - synchronizes the *desired state* described by northbound
  components with the *actual state* of the southbound. 

### Tools

The VPP agent repository also contains tools for building and troubleshooting 
of VNFs based on the VPP Agent:

* [agentctl][agentctl] - a CLI tool that shows the state of a set of 
   VPP agents can configure the agents
* [vpp-agent-ctl][vpp-agent-ctl] - a utility for testing VNF Agent 
  configuration. It contains a set of pre-defined configurations that can 
  be sent to the VPP Agent either interactively or in a script. 
* [docker][docker] - container-based development environment for the VPP
  agent and for app/extension plugins.

## Contributing

If you are interested in contributing, please see the [contribution guidelines][contribution].

[agentctl]: cmd/agentctl
[cn-infra]: https://github.com/ligato/cn-infra
[contiv-vpp]: https://github.com/contiv/vpp
[contribution]: CONTRIBUTING.md
[docker]: docker
[docker-image]: http://docs.ligato.io/en/latest/user-guide/get-agent/#build-local-image
[docs-govppmux-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#govppmux-plugin
[docs-kv-scheduler]: https://docs.ligato.io/en/latest/plugins/kvs-plugin/
[docs-linux-interface-plugin]: https://docs.ligato.io/en/latest/plugins/linux-plugins/#interface-plugin
[docs-linux-l3-plugin]: https://docs.ligato.io/en/latest/plugins/linux-plugins/#l3-plugin
[docs-linux-ns-plugin]: https://docs.ligato.io/en/latest/plugins/linux-plugins/#namespace-plugin
[docs-linux-plugins]: https://docs.ligato.io/en/latest/plugins/linux-plugins/
[docs-rest-plugin]: https://docs.ligato.io/en/latest/plugins/connection-plugins/#rest-plugin 
[docs-vpp-acl-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#access-control-lists-plugin
[docs-vpp-interface-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#interface-plugin
[docs-vpp-l2-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#l2-plugin
[docs-vpp-l3-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#l3-plugin
[docs-vpp-ipsec-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#ipsec-plugin
[docs-vpp-nat-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#nat-plugin
[docs-vpp-plugins]:https://docs.ligato.io/en/latest/plugins/vpp-plugins/
[docs-vpp-punt-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#punt-plugin
[docs-vpp-sr-plugin]: https://docs.ligato.io/en/latest/plugins/vpp-plugins/#sr-plugin
[etcd]: docs/arm64/etcd.md
[etcd-local]: docker/dev/README.md#running-etcd-server-on-local-host
[govpp]: https://wiki.fd.io/view/GoVPP
[kafka]: docs/arm64/kafka.md
[ligato-docs]: http://docs.ligato.io/
[protobufs]: https://developers.google.com/protocol-buffers/
[vnf]: https://github.com/ligato/cn-infra/blob/master/docs/readmes/cn_virtual_function.md
[vpp]: https://fd.io/technology/#vpp
[vpp-agent]: https://hub.docker.com/r/ligato/vpp-agent
[vpp-agent-arm64]: https://hub.docker.com/r/ligato/vpp-agent-arm64
[vpp-agent-ctl]: cmd/vpp-agent-ctl
