# VPP Agent

[![Build Status](https://travis-ci.org/ligato/vpp-agent.svg?branch=master)](https://travis-ci.org/ligato/vpp-agent)
[![Coverage Status](https://coveralls.io/repos/github/ligato/vpp-agent/badge.svg?branch=master)](https://coveralls.io/github/ligato/vpp-agent?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ligato/vpp-agent)](https://goreportcard.com/report/github.com/ligato/vpp-agent)
[![GitHub license](https://img.shields.io/badge/license-Apache%20license%202.0-blue.svg)](https://github.com/ligato/vpp-agent/blob/master/LICENSE)

Please note that the content of the repository is currently WORK IN PROGRESS.

The VPP Agent is a management tool for VPP ([Vector Packet Processing](https://fd.io/)) built on [cn-infra](github.com/ligato/cn-infra).

![vpp agent plugins](vpp_agent_plugins.png "VPP Agent Plugins on top of cn-infra")
 
* Default VPP Plugins - provides abstraction on top of VPP binary API for:
  * NET Interface - Network interfaces configuration (Gigi ETH, MEMIF, AF_Packet, VXLAN, Loopback...)
  * L2 - Bridge Domains, FIBs...
  * L3 - IP Routes, VRFs...
  * ACL - configures VPP ACL Plugin
* GOVPP - allows other plugins to access VPP independently on each other by means of connection multiplexing
* Linux (VETH) - configures Linux Virtual Ethernets
* Core - lifecycle management of plugins (loading, initialization, unloading) see [cn-infra](https://github.com/ligato/cn-infra)

# Quickstart(TBD)
1. Run VPP agent in Docker image
2. Configure the VPP agent using agentctl
3. Check the configurtion (using agentctl or directly using VPP console)

# Next Steps(TBD)
Deployment:
![K8s integration](k8s_deployment_thumb.png "VPP Agent - K8s integration")

Extensibility:
TBD

Design & architecture:
![VPP agent 10.000 feet](vpp_agent_10K_feet_thumb.png "VPP Agent - 10.000 feet view on the architecture")

Contribution:
If you are interested in contributing, please see the [contribution guidelines](CONTRIBUTING.md).

The tool used for managing third-party dependencies is [Glide](https://github.com/Masterminds/glide). After adding or updating
a dependency in `glide.yaml` run `make install-dep` to download specified dependencies into the vendor folder. 
