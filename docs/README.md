# Docs

Learn how to use, deploy and develop with VPP Agent.

### What is VPP Agent?

The VPP Agent is a Go implementation of a control/management plane for [VPP][1fdio-vpp] based
cloud-native [Virtual Network Functions][vnf] (VNFs). The VPP Agent is built on top of 
[CN Infra][cn-infra], a framework for developing cloud-native VNFs (CNFs).

The VPP Agent can be used as-is as a management/control agent for VNFs  based on off-the-shelf
VPP (e.g. a VPP-based vswitch), or as a framework for developing management agents for VPP-based
CNFs. An example of a custom VPP-based CNF is the [Contiv-VPP][contivvpp] vswitch.

### Introduction to VPP Agent
- [Architecture](Architecture.md) describes architecture of VPP Agent.
- [Design](Design.md) describes VPP Agent design.
- [Deployment](Deployment.md) provides information about deployment options.
  
### Getting started with VPP Agent
- [User Guide](https://github.com/ligato/vpp-agent/wiki/user-guide) is the best place to start for users.
- [Tutorials](tutorials/README.md) will show you step-by-step tutorials.
- [Troubleshooting](https://github.com/ligato/vpp-agent/wiki/FAQ) contains list of commonly encountered problems.
- [Release Changelog](https://github.com/ligato/vpp-agent/blob/master/CHANGELOG.md) contains list of changes for released versions.

### Development with VPP Agent
- [Development Guide](https://github.com/ligato/vpp-agent/wiki/development-guide) is the best place to start for developers.
- [Examples](https://github.com/ligato/vpp-agent/blob/master/examples/README.md) contains various examples.
- [KVScheduler](kvscheduler/README.md) contains developer guide for KVScheduler.
- [Testing](https://github.com/ligato/vpp-agent/wiki/testing/Testing) contains information about integration tests.
- [GoDocs](https://godoc.org/github.com/ligato/vpp-agent) provides auto-generated code documentation.

[fdio-vpp]: https://fd.io/technology/#vpp
[vnf]: https://github.com/ligato/cn-infra/blob/master/docs/readmes/cn_virtual_function.md
[cn-infra]: https://github.com/ligato/cn-infra
[contivvpp]: https://github.com/contiv/vpp
