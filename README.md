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

**Note:** The branch pantheon-dev has been **deprecated** (v1).

### Images

|Image|Image Size/Layers|Info|
|---|:---:|---|
|[![ligato/vpp-agent](https://img.shields.io/badge/image-ligato/vpp--agent-blue.svg?logo=docker&logoColor=white)](https://cloud.docker.com/u/ligato/repository/docker/ligato/vpp-agent)|![MicroBadger Size](https://img.shields.io/microbadger/image-size/ligato/vpp-agent.svg) ![MicroBadger Layers](https://img.shields.io/microbadger/layers/ligato/vpp-agent.svg)|minimal image for production|
|[![ligato/dev-vpp-agent](https://img.shields.io/badge/image-ligato/dev--vpp--agent-blue.svg?logo=docker&logoColor=white)](https://cloud.docker.com/u/ligato/repository/docker/ligato/dev-vpp-agent)|![MicroBadger Size](https://img.shields.io/microbadger/image-size/ligato/dev-vpp-agent.svg) ![MicroBadger Layers](https://img.shields.io/microbadger/layers/ligato/dev-vpp-agent.svg)|image prepared for developers|

The image tag `latest` is built from **master branch** and `dev` tag is built from **dev branch**.

## Quickstart

For a quick start with the VPP Agent, you can use the pre-built Docker images on DockerHub
that contain the VPP Agent and VPP: [ligato/vpp-agent][vpp-agent] (or for ARM64: [ligato/vpp-agent-arm64][vpp-agent-arm64]).

0. Start ETCD on your host (e.g. in Docker as described [here][etcd-local]).

   Note: for ARM64 see the information for [etcd][etcd].

1. Run VPP + VPP Agent in a Docker container:
```
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

Extensive documentation for the VPP Agent can be found at [docs.ligato.io](https://docs.ligato.io).

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

## Contributing

If you are interested in contributing, please see the [contribution guidelines][contribution].

[agentctl]: cmd/agentctl
[cn-infra]: https://github.com/ligato/cn-infra
[contiv-vpp]: https://github.com/contiv/vpp
[contribution]: CONTRIBUTING.md
[docker]: docker
[docker-image]: http://docs.ligato.io/en/latest/user-guide/get-agent/#build-local-image
[etcd]: docs/arm64/etcd.md
[etcd-local]: docker/dev/README.md#running-etcd-server-on-local-host
[govpp]: https://wiki.fd.io/view/GoVPP
[kafka]: docs/arm64/kafka.md
[ligato-docs]: http://docs.ligato.io/
[protobufs]: https://developers.google.com/protocol-buffers/
[vnf]: https://docs.ligato.io/en/latest/intro/glossary/#cnf
[vpp]: https://fd.io/technology/#vpp
[vpp-agent]: https://hub.docker.com/r/ligato/vpp-agent
[vpp-agent-arm64]: https://hub.docker.com/r/ligato/vpp-agent-arm64
[vpp-agent-ctl]: cmd/vpp-agent-ctl
