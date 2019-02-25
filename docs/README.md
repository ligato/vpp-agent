# Docs

Learn how to use, deploy and develop with VPP Agent.

### What is VPP Agent?

The VPP Agent is a Go implementation of a control/management plane for
[VPP][fdio-vpp] based cloud-native [Virtual Network Functions][vnf] (VNFs). The VPP
Agent is built on top of the [CN Infra][cn-infra], platform for developing 
cloud-native VNFs.

The VPP Agent can be used as-is as a management/control agent for VNFs 
based on off-the-shelf VPP (e.g. a VPP-based vswitch), or as a
platform for developing customized VNFs with customized VPP-based data.

### Introduction to VPP Agent
- [Architecture](/docs/Architecture.md) describes architecture of VPP Agent.
- [Design](/docs/Design.md) describes VPP Agent design.
- [Deployment](/docs/Deployment.md) provides information about deployment options.
  
### Getting started with VPP Agent
- [User Guide](https://github.com/ligato/vpp-agent/wiki/user-guide) is the best place to start for users.
- [Tutorials](/docs/tutorials/README.md) will show you step-by-step tutorials.
- [Troubleshooting](https://github.com/ligato/vpp-agent/wiki/FAQ) contains list of commonly encountered problems.
- [Release Changelog](/CHANGELOG.md) contains list of changes for released versions.

### Development with VPP Agent
- [Development Guide](https://github.com/ligato/vpp-agent/wiki/user-guide) is the best place to start for developers.
- [Examples](/examples/README.md) contains various examples.
- [KVScheduler](/docs/kvscheduler/README.md) provides documentation to KVScheduler.
- [GoDocs](https://godoc.org/github.com/ligato/vpp-agent) provides auto-generated code documentation.
- [Testing](https://github.com/ligato/vpp-agent/wiki/testing/Testing) contains information for testers about integration tests.

[fdio-vpp]: https://fd.io/technology/#vpp
[vnf]: https://github.com/ligato/cn-infra/blob/master/docs/readmes/cn_virtual_function.md
[cn-infra]: https://github.com/ligato/cn-infra
