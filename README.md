# VPP agent

[![Build Status](https://travis-ci.org/ligato/vpp-agent.svg?branch=master)](https://travis-ci.org/ligato/vpp-agent)
[![Coverage Status](https://coveralls.io/repos/github/ligato/vpp-agent/badge.svg?branch=master)](https://coveralls.io/github/ligato/vpp-agent?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ligato/vpp-agent)](https://goreportcard.com/report/github.com/ligato/vpp-agent)
[![GitHub license](https://img.shields.io/badge/license-Apache%20license%202.0-blue.svg)](https://github.com/ligato/vpp-agent/blob/master/LICENSE)

Please note that the content of the repository is currently work in progress.

The vpp agent is a management tool for vpp built on [cn-infra](github.com/ligato/cn-infra).

The tool used for managing third-party dependencies is [Glide](https://github.com/Masterminds/glide). After adding or updating
a dependency in `glide.yaml` run `make install-dep` to download specified dependencies into the vendor folder. 

If you are interested in contributing, please see the [contribution guidelines](CONTRIBUTING.md).

# Architecture

VPP Agent Plugins on top of cn-infra:

![vpp agent plugins](vpp_agent_plugins.png "VPP Agent Plugins on top of cn-infra")

10.000 feet architecture:

![VPP agent 10.000 feet](vpp_agent_10K_feet.png "VPP Agent - 10.000 feet view on the architecture")

* SFC Controller - renders desired network stitching configuration for multiple agents to the Data Store
* Control Plane APPs - renders specific network configuration for multiple agents to the Data Store
* Data Store - ETCD, Redis, Cassandra etc. to:
  * store the configuration
  * operational state (network counters & statistics, errors...)
* VPP vSwitch - Privileged container that cross connects multiple VNFs
* VPP VNF - Benefits of putting VPP to a container
 * supports failover
 * simplifies: upgrade, start/top, potentially also scaling
 * microservices: small & reusable apps
* Non VPP VNF - non VPP containers can interact together with VPP containers (see below MEMIFs, VETH)
* Messaging - AD-HOC events (e.g. link UP/Down)
 
K8s integration:

![K8s integration](k8s_deployment.png "VPP Agent - K8s integration")

Contiv deployment:
TBD - in memory calls (not remote calls)