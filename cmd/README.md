# Executables

This package groups executables that can be built from sources in this 
repository:

- [vpp-agent](vpp-agent/main.go) - the VPP Agent executable
- [agentctl](agentctl/agentctl.go) - CLI tool that allows to show the state 
  and to configure agents connected to etcd
- [vpp-agent-ctl](vpp-agent-ctl/main.go) - a utility for testing VNF Agent 
  configuration. It contains a set of pre-defined configurations that can 
  be sent to the VPP Agent either interactively or in a script.