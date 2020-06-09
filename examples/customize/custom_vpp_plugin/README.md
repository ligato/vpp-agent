# Custom VPP plugin

The Custom VPP plugin example contains a working example of custom agent which
adds support for a custom VPP plugin. This example can serve as a skeleton
code for developing custom agents adding new VPP functionality that is not
part of official VPP Agent.

## Structure

- [binapi](binapi/) - contains generated Go code of VPP binary API for a specific VPP version
- [proto](proto/) - contains Protobuf definition and model registration for the Northbound API
- [syslog](syslog/) - contains Ligato plugin Syslog that initializes vppcalls handler and registers KV descriptors that describe the Syslog models and their behaviour (CRUD operations, validation, dependencies, etc.)
- [main.go](main.go) - contains the agent example that wires all the components together and runs the agent
